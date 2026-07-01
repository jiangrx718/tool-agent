package model

import (
	modelService "tool-agent/internal/service/model"

	"github.com/gin-gonic/gin"
)

// NewModelHandler 创建模型处理器
func NewModelHandler(engine *gin.Engine) *ModelHandler {
	return &ModelHandler{
		engine:  engine,
		service: modelService.NewModelService(),
	}
}

// ModelHandler 模型处理器
type ModelHandler struct {
	engine  *gin.Engine
	service modelService.ServiceIFace
}

// RegisterRoutes 注册模型路由
func (mh *ModelHandler) RegisterRoutes(routerGroup *gin.RouterGroup) {
	g := routerGroup.Group("/models")

	// 模型列表
	g.GET("/list", mh.List)
}
