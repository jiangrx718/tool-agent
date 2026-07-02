package picture_book

import (
	"tool-agent/server/http/response"
	"tool-agent/utils"

	"github.com/gin-gonic/gin"
)

// CreateReq 创建绘本请求参数
type CreateReq struct {
	Title      string `json:"title" binding:"required"`
	Icon       string `json:"icon"`
	CategoryId string `json:"category_id"`
	Type       int    `json:"type"`
	Status     string `json:"status"`
	Position   int    `json:"position"`
}

// Create 创建绘本接口
func (h *PictureBookHandler) Create(ctx *gin.Context) {
	var reqBody CreateReq
	var logger = utils.SugarContext(ctx)
	if err := ctx.Bind(&reqBody); err != nil {
		logger.Infow("Handler PictureBook Create ctx.Bind err", "error", err)
		response.ParameterError(ctx, err)
		return
	}

	result, err := h.service.Create(ctx, reqBody.Title, reqBody.Icon, reqBody.CategoryId, reqBody.Type, reqBody.Status, reqBody.Position)
	if err != nil {
		logger.Errorw("Handler PictureBook Create service.Create error", "error", err)
		response.InternalError(ctx)
		return
	}

	if result.GetCode() != 0 {
		response.Failed(ctx, result.GetCode(), result.GetMessage(), result.GetData())
		return
	}

	ctx.JSON(200, result)
	return
}
