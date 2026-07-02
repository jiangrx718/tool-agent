package picture_book

import (
	"tool-agent/server/http/response"
	"tool-agent/utils"

	"github.com/gin-gonic/gin"
)

// UpdateReq 更新绘本请求参数
type UpdateReq struct {
	BookId     string `json:"book_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Icon       string `json:"icon"`
	CategoryId string `json:"category_id" binding:"required"`
	Type       int    `json:"type"`
	Status     string `json:"status"`
	Position   int    `json:"position"`
}

// Update 更新绘本接口
func (h *PictureBookHandler) Update(ctx *gin.Context) {
	var reqBody UpdateReq
	var logger = utils.SugarContext(ctx)
	if err := ctx.Bind(&reqBody); err != nil {
		logger.Infow("Handler PictureBook Update ctx.Bind err", "error", err)
		response.ParameterError(ctx, err)
		return
	}

	result, err := h.service.Update(ctx, reqBody.BookId, reqBody.Title, reqBody.Icon, reqBody.CategoryId, reqBody.Type, reqBody.Status, reqBody.Position)
	if err != nil {
		logger.Errorw("Handler PictureBook Update service.Update error", "error", err)
		response.InternalError(ctx)
		return
	}

	if result.GetCode() != 0 {
		response.Failed(ctx, result.GetCode(), result.GetMessage(), result.GetData())
		return
	}

	response.Successful(ctx, result.GetData())
}
