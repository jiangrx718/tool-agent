package demo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"tool-agent/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

const (
	weatherSystemPrompt = `你是一个天气助手。当用户询问某个城市的天气时，请调用 get_weather 工具获取实时天气信息，并基于返回的数据用简洁的中文向用户汇报当前气温、天气状况、湿度和风速。
如果用户没有说明城市，请礼貌地追问要查询哪个城市的天气。`

	// weatherMaxStep 手动 Agent 循环最大轮数（模型调用 + 工具执行 各算一轮）
	weatherMaxStep = 10
)

// weatherDeps 缓存天气智能体的共享依赖，只初始化一次
var weatherDeps struct {
	once  sync.Once
	model model.ToolCallingChatModel
	tool  tool.InvokableTool
	err   error
}

// WeatherStream 以 channel 的方式流式返回天气智能体的回答，协议与 ChatStream 一致：
//   - 内容片段：ChatStream{Code: 0, Message: "chunk", Data.Results: "<文本片段>"}
//   - 全部结束：ChatStream{Code: 0, Message: "done"}
//   - 出现错误：ChatStream{Code: 500, Message: "error: ..."}
//
// 智能体由 DeepSeek 模型 + get_weather 工具组成，模型自主决定是否调用工具。
// 采用手动 Agent 循环实现，避免 eino ReAct 默认 StreamToolCallChecker
// 在模型先输出文本后发出工具调用时提前判定为"无工具调用"的问题。
func (s *DemoService) WeatherStream(ctx context.Context, question string) (chan ChatStream, error) {
	chatModel, weatherTool, err := getWeatherDeps(ctx)
	if err != nil {
		return nil, err
	}

	result := make(chan ChatStream, 1000)

	go func() {
		logger := utils.SugarContext(ctx)
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("[demo] WeatherStream panic: %v", err)
				select {
				case <-ctx.Done():
				case result <- ChatStream{Code: chatStreamCodeError, Message: "error: " + fmt.Sprintf("%v", err)}:
				}
			}
			close(result)
		}()

		if question == "" {
			question = "今天北京天气怎么样？"
		}

		messages := []*schema.Message{
			schema.SystemMessage(weatherSystemPrompt),
			schema.UserMessage(question),
		}

		// 手动 Agent 循环：模型流式生成 -> 判断是否调用工具 -> 执行工具 -> 回传结果 -> 继续生成
		for step := 0; step < weatherMaxStep; step++ {
			if ctx.Err() != nil {
				return
			}

			// 1. 获取工具信息，绑定到模型
			toolInfo, err := weatherTool.Info(ctx)
			if err != nil {
				logger.Errorf("[demo] 获取天气工具信息失败: %v", err)
				return
			}
			boundModel, err := chatModel.WithTools([]*schema.ToolInfo{toolInfo})
			if err != nil {
				logger.Errorf("[demo] 绑定天气工具失败: %v", err)
				return
			}

			// 2. 流式生成：实时推送文本片段，同时累积完整消息
			reader, err := boundModel.Stream(ctx, messages)
			if err != nil {
				logger.Errorf("[demo] 启动天气模型流式生成失败: %v", err)
				return
			}

			fullMsg, streamErr := pumpStream(ctx, result, reader)
			reader.Close()
			if streamErr != nil {
				logger.Errorf("[demo] 天气模型流式接收错误: %v", streamErr)
				select {
				case <-ctx.Done():
				case result <- ChatStream{Code: chatStreamCodeError, Message: "error: " + streamErr.Error()}:
				}
				return
			}
			if fullMsg == nil {
				return
			}

			// 3. 判断是否有工具调用：流结束后再判定，避免提前终止
			if len(fullMsg.ToolCalls) == 0 {
				select {
				case <-ctx.Done():
				case result <- ChatStream{Code: chatStreamCodeOK, Message: "done"}:
				}
				return
			}

			// 4. 执行工具调用，把 assistant 消息和工具结果都加入上下文
			messages = append(messages, fullMsg)
			for _, tc := range fullMsg.ToolCalls {
				if ctx.Err() != nil {
					return
				}
				args := tc.Function.Arguments
				if args == "" {
					args = "{}"
				}
				toolResult, err := weatherTool.InvokableRun(ctx, args)
				if err != nil {
					toolResult = fmt.Sprintf(`{"error": "%s"}`, err.Error())
				}
				messages = append(messages, schema.ToolMessage(toolResult, tc.ID, schema.WithToolName(tc.Function.Name)))
			}
			// 继续循环：让模型基于工具结果再次生成
		}

		// 达到最大轮数仍未结束
		logger.Warnf("[demo] 天气智能体达到最大轮数 %d", weatherMaxStep)
		select {
		case <-ctx.Done():
		case result <- ChatStream{Code: chatStreamCodeOK, Message: "done"}:
		}
	}()

	return result, nil
}

// pumpStream 消费模型的流式输出：把每个非空文本片段推入 channel（实时），
// 同时累积合并成一个完整的 Message 用于后续工具调用判断。
// 返回累积后的完整消息；streamErr 非 nil 表示读取中途出错。
func pumpStream(ctx context.Context, result chan ChatStream, reader *schema.StreamReader[*schema.Message]) (*schema.Message, error) {
	var chunks []*schema.Message
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		msg, err := reader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if msg != nil {
			chunks = append(chunks, msg)
			if msg.Content != "" {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case result <- ChatStream{
					Code:    chatStreamCodeOK,
					Message: "chunk",
					Data:    DataItem{Results: msg.Content},
				}:
				}
			}
		}
	}
	if len(chunks) == 0 {
		return nil, nil
	}
	merged, err := schema.ConcatMessages(chunks)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// getWeatherDeps 懒加载天气智能体的共享依赖（模型 + 工具），只初始化一次。
// 复用 demo 的 DeepSeek 环境变量配置。
func getWeatherDeps(ctx context.Context) (model.ToolCallingChatModel, tool.InvokableTool, error) {
	weatherDeps.once.Do(func() {
		// 使用 openai 适配器，兼容所有 OpenAI 兼容接口（DeepSeek / GPT / GLM / Kimi / 千问 等）
		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  os.Getenv("CHAT_MODEL_KEY"),
			Model:   os.Getenv("CHAT_MODEL_NAME"),
			BaseURL: os.Getenv("CHAT_MODEL_BASE_URL"),
		})
		if err != nil {
			weatherDeps.err = fmt.Errorf("WeatherStream 创建 ChatModel 失败: %w", err)
			return
		}

		weatherTool, err := createWeatherTool()
		if err != nil {
			weatherDeps.err = fmt.Errorf("创建天气工具失败: %w", err)
			return
		}

		weatherDeps.model = chatModel
		weatherDeps.tool = weatherTool
	})

	if weatherDeps.err != nil {
		return nil, nil, weatherDeps.err
	}
	return weatherDeps.model, weatherDeps.tool, nil
}

// ---- 天气工具 ----

type weatherToolInput struct {
	City string `json:"city" jsonschema_description:"要查询天气的城市名称，例如：北京、上海、Shanghai" jsonschema:"required"`
}

type weatherToolOutput struct {
	City        string `json:"city"`
	Temperature string `json:"temperature"`
	Description string `json:"description"`
	Humidity    string `json:"humidity"`
	WindSpeed   string `json:"wind_speed"`
}

// createWeatherTool 创建 get_weather 工具，供智能体调用
func createWeatherTool() (tool.InvokableTool, error) {
	return toolutils.InferTool(
		"get_weather",
		"查询指定城市的实时天气信息。当用户询问某个城市的天气时调用此工具。",
		func(ctx context.Context, input weatherToolInput) (*weatherToolOutput, error) {
			return queryWeather(ctx, input.City)
		},
	)
}

// wttrResponse 只解析 wttr.in JSON 响应中需要的字段
type wttrResponse struct {
	CurrentCondition []struct {
		TempC       string `json:"temp_C"`
		Humidity    string `json:"humidity"`
		WindSpeed   string `json:"windspeedKmph"`
		WindDir     string `json:"winddir16Point"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"current_condition"`
	NearestArea []struct {
		AreaName []struct {
			Value string `json:"value"`
		} `json:"areaName"`
	} `json:"nearest_area"`
}

// queryWeather 调用 wttr.in 免费天气 API 获取实时天气，无需 API Key
func queryWeather(ctx context.Context, city string) (*weatherToolOutput, error) {
	if city == "" {
		return nil, errors.New("城市名称不能为空")
	}
	city = strings.TrimSpace(city)

	reqURL := fmt.Sprintf("https://wttr.in/%s?format=j1&lang=zh", url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建天气请求失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求天气服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("天气服务返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取天气响应失败: %w", err)
	}

	var wttr wttrResponse
	if err := json.Unmarshal(body, &wttr); err != nil {
		return nil, fmt.Errorf("解析天气响应失败: %w", err)
	}
	if len(wttr.CurrentCondition) == 0 {
		return nil, errors.New("未获取到天气数据")
	}

	cur := wttr.CurrentCondition[0]
	desc := ""
	if len(cur.WeatherDesc) > 0 {
		desc = cur.WeatherDesc[0].Value
	}
	areaName := city
	if len(wttr.NearestArea) > 0 && len(wttr.NearestArea[0].AreaName) > 0 {
		areaName = wttr.NearestArea[0].AreaName[0].Value
	}

	return &weatherToolOutput{
		City:        areaName,
		Temperature: cur.TempC + "°C",
		Description: desc,
		Humidity:    cur.Humidity + "%",
		WindSpeed:   cur.WindSpeed + " km/h " + cur.WindDir,
	}, nil
}
