package kbagent

// Config 知识库智能体的整体配置
type Config struct {
	Chat      ChatConfig
	Embedding EmbeddingConfig
	Agent     AgentConfig
}

// ChatConfig 对话模型配置（DeepSeek / 千问3 均兼容 OpenAI 接口）
type ChatConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	Temperature float32
}

// EmbeddingConfig 向量化模型配置（千问 DashScope text-embedding-v3）
type EmbeddingConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// AgentConfig 智能体行为参数
type AgentConfig struct {
	MaxStep        int     // ReAct 最大循环步数
	TopK           int     // 检索返回的文档数量
	ScoreThreshold float64 // 相似度阈值，低于此值的结果被过滤
}
