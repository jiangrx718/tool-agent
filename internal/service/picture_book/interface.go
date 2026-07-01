package picture_book

import (
	"context"

	"tool-agent/internal/common"
)

// ServiceIFace 绘本服务接口
type ServiceIFace interface {
	// Create 创建绘本
	Create(ctx context.Context, title, icon, categoryId string, bookType int, status string, position int) (*common.ServiceResult, error)
	// Update 更新绘本
	Update(ctx context.Context, bookId, title, icon, categoryId string, bookType int, status string, position int) (*common.ServiceResult, error)
	// Delete 删除绘本
	Delete(ctx context.Context, bookId string) (*common.ServiceResult, error)
	// List 查询绘本列表
	List(ctx context.Context, title string, bookType int, status string, offset, limit int) (*common.ServiceResult, error)
}
