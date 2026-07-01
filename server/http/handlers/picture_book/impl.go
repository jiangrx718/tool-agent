package picture_book

import (
	pictureBookService "tool-agent/internal/service/picture_book"

	"github.com/gin-gonic/gin"
)

// NewPictureBookHandler 创建绘本处理器
func NewPictureBookHandler(engine *gin.Engine) *PictureBookHandler {
	return &PictureBookHandler{
		engine:  engine,
		service: pictureBookService.NewPictureBookService(),
	}
}

// PictureBookHandler 绘本处理器
type PictureBookHandler struct {
	engine  *gin.Engine
	service pictureBookService.ServiceIFace
}

// RegisterRoutes 注册绘本路由
func (h *PictureBookHandler) RegisterRoutes(routerGroup *gin.RouterGroup) {
	g := routerGroup.Group("/picture_book")

	// 绘本相关接口
	g.POST("/create", h.Create)
	g.POST("/update", h.Update)
	g.POST("/delete", h.Delete)
	g.GET("/list", h.List)
}
