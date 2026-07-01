package dao

import (
	"context"
	"tool-agent/model"

	"gorm.io/gorm"
)

// ModelDAO 模型表数据访问层
type ModelDAO struct {
	db *gorm.DB
}

// NewModelDAO 创建 ModelDAO
func NewModelDAO(db *gorm.DB) *ModelDAO {
	return &ModelDAO{db: db}
}

// ListQuery 模型列表查询条件
type ListQuery struct {
	Keyword string
	Offset  int
	Limit   int
}

// List 查询模型列表
func (d *ModelDAO) List(ctx context.Context, query ListQuery) ([]*model.Model, int64, error) {
	db := d.db.WithContext(ctx).Model(&model.Model{})
	if query.Keyword != "" {
		db = db.Where("model_name LIKE ?", "%"+query.Keyword+"%")
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.Model
	if err := db.Order("id DESC").Offset(query.Offset).Limit(query.Limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, count, nil
}
