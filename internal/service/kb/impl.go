package kb

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"

	"tool-agent/internal/kbagent"
	"tool-agent/model"
	"tool-agent/utils"
)

// ServiceIFace 知识库服务接口
type ServiceIFace interface {
	CreateDocument(ctx context.Context, title, content string) (uint64, error)
	ListDocuments(ctx context.Context) ([]model.KbDocument, error)
	DeleteDocument(ctx context.Context, id uint64) error
	AgentChat(ctx context.Context, question string) (string, []*kbagent.Source, error)
	AgentStream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error)
	RagAsk(ctx context.Context, question string) (string, []*kbagent.Source, error)
	RagStream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error)
}

// Service 知识库服务
type Service struct {
	agent *kbagent.Agent
}

// NewKBService 创建知识库服务，初始化 Eino 智能体
func NewKBService() *Service {
	cfg := loadConfig()
	agent, err := kbagent.NewAgent(context.Background(), cfg, utils.DB())
	if err != nil {
		utils.Sugar().Errorf("[kb] init agent failed (chat/rag will unavailable): %v", err)
	}
	return &Service{agent: agent}
}

func loadConfig() *kbagent.Config {
	return &kbagent.Config{
		Chat: kbagent.ChatConfig{
			APIKey:      viper.GetString("llm.chat.api_key"),
			BaseURL:     viper.GetString("llm.chat.base_url"),
			Model:       viper.GetString("llm.chat.model"),
			Temperature: float32(viper.GetFloat64("llm.chat.temperature")),
		},
		Embedding: kbagent.EmbeddingConfig{
			APIKey:  viper.GetString("llm.embedding.api_key"),
			BaseURL: viper.GetString("llm.embedding.base_url"),
			Model:   viper.GetString("llm.embedding.model"),
		},
		Agent: kbagent.AgentConfig{
			MaxStep:        viper.GetInt("llm.agent.max_step"),
			TopK:           viper.GetInt("llm.agent.top_k"),
			ScoreThreshold: viper.GetFloat64("llm.agent.score_threshold"),
		},
	}
}

func (s *Service) CreateDocument(ctx context.Context, title, content string) (uint64, error) {
	return s.agent.AddDocument(ctx, title, content)
}

func (s *Service) ListDocuments(ctx context.Context) ([]model.KbDocument, error) {
	return s.agent.ListDocuments(ctx)
}

func (s *Service) DeleteDocument(ctx context.Context, id uint64) error {
	return s.agent.DeleteDocument(ctx, id)
}

func (s *Service) AgentChat(ctx context.Context, question string) (string, []*kbagent.Source, error) {
	return s.agent.AgentChat(ctx, question)
}

func (s *Service) AgentStream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error) {
	return s.agent.AgentStream(ctx, question)
}

func (s *Service) RagAsk(ctx context.Context, question string) (string, []*kbagent.Source, error) {
	return s.agent.RagAsk(ctx, question)
}

func (s *Service) RagStream(ctx context.Context, question string) (*schema.StreamReader[*schema.Message], error) {
	return s.agent.RagStream(ctx, question)
}