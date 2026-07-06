package kbagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/embedding"
)

// OpenAICompatibleEmbedder 调用任意 OpenAI 兼容的 /embeddings 接口。
// 支持 DashScope text-embedding-v3、OpenAI text-embedding-3-small 等。
type OpenAICompatibleEmbedder struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAICompatibleEmbedder 创建一个基于 OpenAI 兼容接口的向量化器
func NewOpenAICompatibleEmbedder(cfg EmbeddingConfig) *OpenAICompatibleEmbedder {
	return &OpenAICompatibleEmbedder{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type embeddingReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResp struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// EmbedStrings 实现 embedding.Embedder 接口，批量将文本转为向量
func (e *OpenAICompatibleEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	body, err := json.Marshal(embeddingReq{Model: e.model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	url := strings.TrimRight(e.baseURL, "/") + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var data embeddingResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if data.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", data.Error.Message)
	}
	if len(data.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	result := make([][]float64, len(data.Data))
	for i, d := range data.Data {
		result[i] = d.Embedding
	}
	return result, nil
}
