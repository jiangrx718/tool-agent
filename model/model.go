package model

// Model 模型表
type Model struct {
	BaseModelID
	DataType      int8   `gorm:"column:data_type;not null;default:0" json:"data_type"`
	ModelID       string `gorm:"column:model_id;size:256;not null;default:'';index:idx_model_id" json:"model_id"`
	ModelName     string `gorm:"column:model_name;size:256;not null;default:''" json:"model_name"`
	ModelPath     string `gorm:"column:model_path;size:512;not null;default:''" json:"model_path"`
	RunPath       string `gorm:"column:run_path;size:512;not null;default:''" json:"run_path"`
	ParamsPath    string `gorm:"column:params_path;size:512;not null;default:''" json:"params_path"`
	TrainCallback string `gorm:"column:train_callback;size:1024;not null;default:''" json:"train_callback"`
	InferCallback string `gorm:"column:infer_callback;size:1024;not null;default:''" json:"infer_callback"`
	InferPath     string `gorm:"column:infer_path;size:1024;not null;default:''" json:"infer_path"`
	Status        int8   `gorm:"column:status;not null;default:1" json:"status"`
	BaseModelTime
}

func (Model) TableName() string {
	return "model"
}
