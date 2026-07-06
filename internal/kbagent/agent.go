package kbagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	openaimodel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"

	toolAgentModel "tool-agent/model"
	"tool-agent/utils"
)

// Source 检索来源信息
type Source struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

// Agent 知识库智能体，封装 ReAct Agent + RAG Graph 两种问答模式
type Agent struct {
	chatModel  model.ToolCallingChatModel
	embedder   *OpenAICompatibleEmbedder
	store      *VectorStore
	retriever  *InMemoryRetriever
	reactAgent *react.Agent
	ragGraph   compose.Runnable[string, *schema.Message]
	db         *gorm.DB
}

// NewAgent 创建知识库智能体，初始化所有 Eino 组件
func NewAgent(ctx context.Context, cfg *Config, db *gorm.DB) (*Agent, error) {
	// 1. 对话模型（eino-ext openai 兼容 DeepSeek / 千问3）
	temp := cfg.Chat.Temperature
	chatModel, err := openaimodel.NewChatModel(ctx, &openaimodel.ChatModelConfig{
		APIKey:      cfg.Chat.APIKey,
		BaseURL:     cfg.Chat.BaseURL,
		Model:       cfg.Chat.Model,
		Temperature: &temp,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	// 2. 向量化器 + 向量存储 + 检索器
	embedder := NewOpenAICompatibleEmbedder(cfg.Embedding)
	store := NewVectorStore()
	topK := cfg.Agent.TopK
	if topK <= 0 {
		topK = 3
	}
	retriever := NewInMemoryRetriever(embedder, store, topK, cfg.Agent.ScoreThreshold)

	// 3. 知识库搜索工具（供 ReAct Agent 调用）
	searchTool, err := createKBSearchTool(retriever)
	if err != nil {
		return nil, fmt.Errorf("create kb search tool: %w", err)
	}

	// 4. ReAct 智能体
	maxStep := cfg.Agent.MaxStep
	if maxStep <= 0 {
		maxStep = 10
	}
	reactAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{searchTool},
		},
		MaxStep: maxStep,
		MessageModifier: func(ctx context.Context, msgs []*schema.Message) []*schema.Message {
			out := make([]*schema.Message, 0, len(msgs)+1)
			out = append(out, schema.SystemMessage(agentSystemPrompt))
			out = append(out, msgs...)
			return out
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create react agent: %w", err)
	}

	// 5. RAG 检索增强生成图（直接检索 -> 生成，无 Agent 循环）
	ragGraph, err := buildRAGGraph(ctx, chatModel, retriever)
	if err != nil {
		return nil, fmt.Errorf("build rag graph: %w", err)
	}

	a := &Agent{
		chatModel:  chatModel,
		embedder:   embedder,
		store:      store,
		retriever:  retriever,
		reactAgent: reactAgent,
		ragGraph:   ragGraph,
		db:         db,
	}

	// 6. 从数据库加载已有文档到内存索引
	if db != nil {
		if err := a.loadDocuments(ctx); err != nil {
			utils.Sugar().Warnf("[kbagent] load documents from db: %v", err)
		} else {
			utils.Sugar().Infof("[kbagent] loaded %d documents into vector store", store.Count())
		}
	}

	return a, nil
}

// ---- 文档管理 ----

// AddDocument 将文档向量化后存入数据库和内存索引
func (a *Agent) AddDocument(ctx context.Context, title, content string) (uint64, error) {
	vectors, err := a.embedder.EmbedStrings(ctx, []string{content})
	if err != nil {
		return 0, fmt.Errorf("embed document: %w", err)
	}

	embJSON, err := json.Marshal(vectors[0])
	if err != nil {
		return 0, fmt.Errorf("marshal embedding: %w", err)
	}

	doc := toolAgentModel.KbDocument{
		Title:     title,
		Content:   content,
		Embedding: string(embJSON),
	}
	if err := a.db.WithContext(ctx).Create(&doc).Error; err != nil {
		return 0, fmt.Errorf("create document: %w", err)
	}

	a.store.Add(strconv.FormatUint(doc.Id, 10), title, content, vectors[0])
	return doc.Id, nil
}

// ListDocuments 列出所有文档（不含向量字段）
func (a *Agent) ListDocuments(ctx context.Context) ([]toolAgentModel.KbDocument, error) {
	var docs []toolAgentModel.KbDocument
	err := a.db.WithContext(ctx).
		Select("id, title, content, created_at, updated_at").
		Order("id DESC").
		Find(&docs).Error
	return docs, err
}

// DeleteDocument 删除文档（数据库 + 内存索引）
func (a *Agent) DeleteDocument(ctx context.Context, id uint64) error {
	if err := a.db.WithContext(ctx).Delete(&toolAgentModel.KbDocument{}, id).Error; err != nil {
		return err
	}
	a.store.Remove(strconv.FormatUint(id, 10))
	return nil
}

func (a *Agent) loadDocuments(ctx context.Context) error {
	var docs []toolAgentModel.KbDocument
	if err := a.db.WithContext(ctx).Find(&docs).Error; err != nil {
		return err
	}
	for _, doc := range docs {
		var vec []float64
		if err := json.Unmarshal([]byte(doc.Embedding), &vec); err != nil {
			utils.Sugar().Warnf("[kbagent] skip document %d, bad embedding: %v", doc.Id, err)
			continue
		}
		a.store.Add(strconv.FormatUint(doc.Id, 10), doc.Title, doc.Content, vec)
	}
	return nil
}

// ---- ReAct 智能体问答 ----

// AgentChat 使用 ReAct 智能体回答问题，模型自主决定是否搜索知识库
func (a *Agent) AgentChat(ctx context.Context, question string) (answer string, sources []*Source, err error) {
	sources = a.retrieveSources(ctx, question)

	msg, err := a.reactAgent.Generate(ctx, []*schema.Message{
		schema.UserMessage(question),
	})
	if err != nil {
		return "", sources, fmt.Errorf("react agent generate: %w", err)
	}
	return msg.Content, sources, nil
}

// AgentStream 使用 ReAct 智能体流式回答，返回消息流
func (a *Agent) AgentStream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error) {
	return a.reactAgent.Stream(ctx, []*schema.Message{
		schema.UserMessage(question),
	})
}

// ---- RAG 检索增强问答 ----

// RagAsk 使用 RAG 图回答问题（先检索后生成，固定流程）
func (a *Agent) RagAsk(ctx context.Context, question string) (answer string, sources []*Source, err error) {
	sources = a.retrieveSources(ctx, question)

	msg, err := a.ragGraph.Invoke(ctx, question)
	if err != nil {
		return "", sources, fmt.Errorf("rag graph invoke: %w", err)
	}
	return msg.Content, sources, nil
}

// RagStream 使用 RAG 图流式回答
func (a *Agent) RagStream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error) {
	return a.ragGraph.Stream(ctx, question)
}

// ---- 内部辅助 ----

func (a *Agent) retrieveSources(ctx context.Context, question string) []*Source {
	docs, err := a.retriever.Retrieve(ctx, question)
	if err != nil {
		utils.Sugar().Warnf("[kbagent] retrieve sources: %v", err)
		return nil
	}
	sources := make([]*Source, len(docs))
	for i, doc := range docs {
		title, _ := doc.MetaData["title"].(string)
		score, _ := doc.MetaData["score"].(float64)
		sources[i] = &Source{ID: doc.ID, Title: title, Score: score}
	}
	return sources
}

// createKBSearchTool 创建知识库检索工具，供 ReAct Agent 调用
func createKBSearchTool(r *InMemoryRetriever) (tool.InvokableTool, error) {
	return toolutils.InferTool(
		"search_knowledge_base",
		"搜索知识库文档。当用户询问知识库中可能有的内容时，调用此工具检索相关文档。",
		func(ctx context.Context, input kbSearchInput) (*kbSearchOutput, error) {
			docs, err := r.Retrieve(ctx, input.Query)
			if err != nil {
				return nil, err
			}
			out := &kbSearchOutput{Total: len(docs)}
			for _, doc := range docs {
				title, _ := doc.MetaData["title"].(string)
				score, _ := doc.MetaData["score"].(float64)
				out.Documents = append(out.Documents, kbDoc{
					Title: title, Content: doc.Content, Score: score,
				})
			}
			return out, nil
		},
	)
}

// buildRAGGraph 构建 RAG 检索增强生成图：
//
//	START -> retrieve_and_format(lambda) -> chat(model) -> END
//
// lambda 节点先检索文档，再拼装 system+user 消息，交给对话模型生成。
func buildRAGGraph(ctx context.Context, chatModel model.BaseChatModel, r *InMemoryRetriever) (compose.Runnable[string, *schema.Message], error) {
	graph := compose.NewGraph[string, *schema.Message]()

	if err := graph.AddLambdaNode("retrieve_and_format",
		compose.InvokableLambda(func(ctx context.Context, query string) ([]*schema.Message, error) {
			docs, err := r.Retrieve(ctx, query)
			if err != nil {
				return nil, fmt.Errorf("retrieve in rag: %w", err)
			}
			return []*schema.Message{
				schema.SystemMessage(fmt.Sprintf(ragSystemPrompt, formatDocContext(docs))),
				schema.UserMessage(query),
			}, nil
		}),
	); err != nil {
		return nil, fmt.Errorf("add lambda node: %w", err)
	}

	if err := graph.AddChatModelNode("chat", chatModel); err != nil {
		return nil, fmt.Errorf("add chat model node: %w", err)
	}

	if err := graph.AddEdge(compose.START, "retrieve_and_format"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("retrieve_and_format", "chat"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("chat", compose.END); err != nil {
		return nil, err
	}

	return graph.Compile(ctx, compose.WithGraphName("KnowledgeBaseRAG"))
}

func formatDocContext(docs []*schema.Document) string {
	if len(docs) == 0 {
		return "（未检索到相关文档）"
	}
	var sb strings.Builder
	for i, doc := range docs {
		title, _ := doc.MetaData["title"].(string)
		sb.WriteString(fmt.Sprintf("[%d] 标题: %s\n内容: %s\n\n", i+1, title, doc.Content))
	}
	return sb.String()
}

// ---- 工具 I/O 类型 ----

type kbSearchInput struct {
	Query string `json:"query" jsonschema_description:"要搜索的问题或关键词"`
}

type kbDoc struct {
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

type kbSearchOutput struct {
	Documents []kbDoc `json:"documents"`
	Total     int     `json:"total"`
}

// ---- Prompt ----

const agentSystemPrompt = `你是一个专业的文档知识库助手。请遵循以下规则：
1. 当用户提问时，优先使用 search_knowledge_base 工具检索知识库中的相关文档
2. 基于检索到的文档内容回答问题，不要编造文档中不存在的信息
3. 如果知识库中没有相关信息，请如实告知用户
4. 回答时注明信息来源（引用文档标题）
5. 使用中文回答`

const ragSystemPrompt = `你是一个专业的文档知识库助手。请根据以下检索到的文档内容回答用户的问题。

检索到的文档：
%s
请遵循以下规则：
1. 基于上述文档内容回答问题，不要编造信息
2. 如果文档中没有相关信息，请如实告知用户
3. 回答时注明信息来源（引用文档标题）
4. 使用中文回答`
