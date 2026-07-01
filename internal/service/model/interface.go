package model

import (
	"context"

	"tool-agent/internal/common"
)

// ServiceIFace 模型服务接口
type ServiceIFace interface {
	// List 查询模型列表
	List(ctx context.Context, keyword string, offset, limit int) (common.ServiceResult, error)
}
