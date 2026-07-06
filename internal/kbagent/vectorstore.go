package kbagent

import (
	"math"
	"sort"
	"sync"
)

type vectorEntry struct {
	id      string
	title   string
	content string
	vector  []float64
}

// SearchResult 单条检索结果
type SearchResult struct {
	ID      string
	Title   string
	Content string
	Score   float64
}

// VectorStore 进程内向量存储，使用余弦相似度做最近邻搜索。
// 数据持久化在 MySQL（通过 Agent 层），内存索引在启动时从数据库加载。
type VectorStore struct {
	mu      sync.RWMutex
	entries []*vectorEntry
}

// NewVectorStore 创建空的向量存储
func NewVectorStore() *VectorStore {
	return &VectorStore{}
}

// Add 添加或更新一条文档向量
func (s *VectorStore) Add(id, title, content string, vector []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.entries {
		if e.id == id {
			s.entries[i] = &vectorEntry{id, title, content, vector}
			return
		}
	}
	s.entries = append(s.entries, &vectorEntry{id, title, content, vector})
}

// Remove 按 ID 删除一条文档
func (s *VectorStore) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.entries {
		if e.id == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			return
		}
	}
}

// Search 执行余弦相似度搜索，返回 TopK 结果（按分数降序）
func (s *VectorStore) Search(queryVec []float64, topK int, threshold float64) []*SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*SearchResult, 0, len(s.entries))
	for _, e := range s.entries {
		score := cosineSimilarity(queryVec, e.vector)
		if score >= threshold {
			results = append(results, &SearchResult{
				ID: e.id, Title: e.title, Content: e.content, Score: score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results
}

// Count 返回当前存储的文档数量
func (s *VectorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

func cosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
