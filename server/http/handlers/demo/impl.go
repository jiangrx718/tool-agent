package demo

import (
	"tool-agent/internal/service/demo"
	demoService "tool-agent/internal/service/demo"
	"tool-agent/server/http/middleware"

	"github.com/gin-gonic/gin"
)

func NewDemoHandler(engine *gin.Engine) *DemoHandler {
	return &DemoHandler{
		engine:  engine,
		service: demo.NewDemoService(),
	}
}

type DemoHandler struct {
	engine  *gin.Engine
	service demoService.ServiceIFace
}

func (h *DemoHandler) RegisterRoutes(routerGroup *gin.RouterGroup) {
	g := routerGroup.Group("/demo")

	// 绘本相关接口
	g.GET("/create", middleware.EventStreamHeadersMiddleware(), h.DemoCreate)
}
