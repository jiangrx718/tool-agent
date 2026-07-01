package model

import (
	"time"
)

type Model struct {
	Id            uint64    `gorm:"column:id;type:bigint(20) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	DataType      int       `gorm:"column:data_type;type:tinyint(4);default:0;NOT NULL" json:"data_type"`
	ModelId       string    `gorm:"column:model_id;type:varchar(256);NOT NULL" json:"model_id"`
	ModelName     string    `gorm:"column:model_name;type:varchar(256);NOT NULL" json:"model_name"`
	ModelPath     string    `gorm:"column:model_path;type:varchar(512);NOT NULL" json:"model_path"`
	RunPath       string    `gorm:"column:run_path;type:varchar(512);NOT NULL" json:"run_path"`
	ParamsPath    string    `gorm:"column:params_path;type:varchar(512);NOT NULL" json:"params_path"`
	TrainCallback string    `gorm:"column:train_callback;type:varchar(1024);NOT NULL" json:"train_callback"`
	InferCallback string    `gorm:"column:infer_callback;type:varchar(1024);NOT NULL" json:"infer_callback"`
	InferPath     string    `gorm:"column:infer_path;type:varchar(1024);NOT NULL" json:"infer_path"`
	Status        int       `gorm:"column:status;type:tinyint(4);default:1;NOT NULL" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
}

func (m *Model) TableName() string {
	return "model"
}
