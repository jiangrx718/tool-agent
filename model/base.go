package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModelFieldId struct {
	Id uint64 `gorm:"column:id;type:bigint unsigned;primary_key;auto_increment;comment:主键" json:"id"`
}

type BaseModelFieldTime struct {
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:添加时间;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间;index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"deleted_at"`
}
