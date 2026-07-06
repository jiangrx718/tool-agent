package kbagent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// InMemoryRetriever 实现 retriever.Retriever 接口。
// 先用 embedder 将 query 转为向量，再在 VectorStore 中做余弦相似度搜索。
type InMemoryRetriever struct {
	embedder  *OpenAICompatibleEmbedder
	store     *VectorStore
	topK      int
	threshold float64
}

// NewInMemoryRetriever 创建内存检索器
func NewInMemoryRetriever(emb *OpenAICompatibleEmbedder, store *VectorStore, topK int, threshold float64) *InMemoryRetriever {
	return &InMemoryRetriever{
		embedder:  emb,
		store:     store,
		topK:      topK,
		threshold: threshold,
	}
}

// Retrieve 实现 retriever.Retriever 接口
func (r *InMemoryRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	opt := retriever.GetCommonOptions(&retriever.Options{}, opts...)

	topK := r.topK
	if opt.TopK != nil {
		topK = *opt.TopK
	}
	threshold := r.threshold
	if opt.ScoreThreshold != nil {
		threshold = *opt.ScoreThreshold
	}

	if r.store.Count() == 0 {
		return []*schema.Document{}, nil
	}

	vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}

	results := r.store.Search(vectors[0], topK, threshold)
	docs := make([]*schema.Document, len(results))
	for i, res := range results {
		docs[i] = &schema.Document{
			ID:      res.ID,
			Content: res.Content,
			MetaData: map[string]any{
				"title": res.Title,
				"score": res.Score,
			},
		}
	}
	return docs, nil
}
