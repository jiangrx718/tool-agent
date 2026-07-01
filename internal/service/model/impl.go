package model

import (
	"tool-agent/internal/dao"
	"tool-agent/utils"

	"gorm.io/gorm"
)

// Service 模型服务
type Service struct {
	db       *gorm.DB
	modelDao *dao.ModelDAO
}

// NewModelService 创建模型服务
func NewModelService() *Service {
	return &Service{
		db:       utils.DB(),
		modelDao: dao.NewModelDAO(utils.DB()),
	}
}
