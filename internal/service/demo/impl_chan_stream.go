package demo

import (
	"context"
	"errors"
	"io"
	"os"
	"tool-agent/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// ChatStream 流式响应的单条消息
type ChatStream struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    DataItem `json:"data"`
}

// DataItem 流式消息数据体
type DataItem struct {
	Results any `json:"results,omitempty"`
}

const (
	chatStreamCodeOK    = 0
	chatStreamCodeError = 500
)

// ChatStream 以 channel 的方式流式返回模型生成内容。
//
// 协议：
//   - 每个内容片段：ChatStream{Code: 0, Message: "chunk", Data.Results: "<文本片段>"}
//   - 全部结束：ChatStream{Code: 0, Message: "done"}
//   - 出现错误：ChatStream{Code: 500, Message: "error: ..."}
//
// channel 在流结束、出错或 ctx 取消时会被关闭，调用方读取到 ok=false 即视为结束。
func (s *DemoService) ChatStream(ctx context.Context, question string) (chan ChatStream, error) {
	// 所有工作都在 goroutine 内完成：创建模型、发起流式请求、逐块读取。
	// channel 立即返回，调用方可以马上开始消费，实现真正的一块一块到达。
	result := make(chan ChatStream, 1000)

	go func() {
		logger := utils.SugarContext(ctx)
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("[demo] ChatStream panic: %v", err)
			}
			close(result)
		}()

		if question == "" {
			question = "你好，请介绍一下 Eino 框架"
		}

		// 1. 创建 ChatModel（DeepSeek）
		// 使用 openai 适配器，兼容所有 OpenAI 兼容接口（DeepSeek / GPT / GLM / Kimi / 千问 等）
		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  os.Getenv("CHAT_MODEL_KEY"),
			Model:   os.Getenv("CHAT_MODEL_NAME"),
			BaseURL: os.Getenv("CHAT_MODEL_BASE_URL"),
		})
		if err != nil {
			logger.Errorf("[demo] ChatStream 创建 ChatModel 失败: %v", err)
			return
		}

		// 2. 准备消息
		messages := []*schema.Message{
			schema.SystemMessage("你是一个友好的 AI 助手"),
			schema.UserMessage(question),
		}

		// 3. 发起流式请求
		reader, err := chatModel.Stream(ctx, messages)
		if err != nil {
			logger.Errorf("[demo] 启动流式生成失败: %v", err)
			return
		}
		defer reader.Close()

		// 4. 逐块消费 StreamReader，把每个 chunk 推入 channel
		for {
			// 客户端断开 / 上游取消时退出
			if ctx.Err() != nil {
				return
			}

			msg, err := reader.Recv()
			switch {
			case errors.Is(err, io.EOF):
				// 正常结束，推送完成标记
				select {
				case <-ctx.Done():
				case result <- ChatStream{Code: chatStreamCodeOK, Message: "done"}:
				}
				return
			case err != nil:
				logger.Errorf("[demo] 流式接收错误: %v", err)
				select {
				case <-ctx.Done():
				case result <- ChatStream{Code: chatStreamCodeError, Message: "error: " + err.Error()}:
				}
				return
			}

			if msg != nil && msg.Content != "" {
				select {
				case <-ctx.Done():
					return
				case result <- ChatStream{
					Code:    chatStreamCodeOK,
					Message: "chunk",
					Data:    DataItem{Results: msg.Content},
				}:
				}
			}
		}
	}()

	return result, nil
}
