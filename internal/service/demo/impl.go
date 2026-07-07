package demo

import (
	"tool-agent/internal/dao"
	"tool-agent/utils"

	"gorm.io/gorm"
)

type DemoService struct {
	db *gorm.DB
}

func NewDemoService() *DemoService {
	s := &DemoService{db: utils.DB()}
	dao.SetDefault(utils.DB())
	return s
}
