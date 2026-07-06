package handlers

import (
	"tool-agent/server/http/handlers/demo"
	"tool-agent/server/http/handlers/kb"
	"tool-agent/server/http/handlers/picture_book"
	"tool-agent/utils"

	"github.com/gin-gonic/gin"
)

// Handler 根路由处理器
type Handler struct {
	router *gin.Engine
}

// NewHandler 创建根路由处理器
func NewHandler(router *gin.Engine) utils.HttpServerHandler {
	h := &Handler{router: router}
	h.RegisterRoutes()
	return h
}

// RegisterRoutes 注册所有路由
func (h *Handler) RegisterRoutes() {
	g := h.router.Group("/api")

	// 绘本相关接口
	picture_book.NewPictureBookHandler(h.router).RegisterRoutes(g)

	// 知识库智能体相关接口
	kb.NewKBHandler(h.router).RegisterRoutes(g)

	// 智能体Demo相关接口
	demo.NewDemoHandler(h.router).RegisterRoutes(g)

}
