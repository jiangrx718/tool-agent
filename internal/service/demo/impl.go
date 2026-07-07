package demo

import (
	"tool-agent/internal/dao"
	"tool-agent/utils"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"gorm.io/gorm"
)

type DemoService struct {
	db           *gorm.DB
	weatherModel model.ToolCallingChatModel
	weatherTool  tool.InvokableTool
}

func NewDemoService() *DemoService {
	s := &DemoService{db: utils.DB()}
	dao.SetDefault(utils.DB())
	return s
}
