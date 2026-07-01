package picture_book

import (
	"tool-agent/server/http/response"
	"tool-agent/utils"

	"github.com/gin-gonic/gin"
)

// DeleteReq 删除绘本请求参数
type DeleteReq struct {
	BookId string `json:"book_id" binding:"required"`
}

// Delete 删除绘本接口
func (h *PictureBookHandler) Delete(ctx *gin.Context) {
	var reqBody DeleteReq
	var logger = utils.SugarContext(ctx)
	if err := ctx.Bind(&reqBody); err != nil {
		logger.Infow("Handler PictureBook Delete ctx.Bind err", "error", err)
		response.ParameterError(ctx)
		return
	}

	result, err := h.service.Delete(ctx, reqBody.BookId)
	if err != nil {
		logger.Errorw("Handler PictureBook Delete service.Delete error", "error", err)
		response.InternalError(ctx)
		return
	}

	if result.Code != 0 {
		response.Failed(ctx, result.Code, result.Msg, result.Data)
		return
	}

	response.Successful(ctx, result.Data)
}
