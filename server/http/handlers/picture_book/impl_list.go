package picture_book

import (
	pictureBookService "tool-agent/internal/service/picture_book"
	"tool-agent/server/http/httputil"
	"tool-agent/server/http/response"
	"tool-agent/utils"

	"github.com/gin-gonic/gin"
)

// ListQuery 绘本列表查询参数
type ListQuery struct {
	httputil.Pagination
	Title  string `form:"title"`
	Type   int    `form:"type"`
	Status string `form:"status"`
}

// List 绘本列表接口
func (h *PictureBookHandler) List(ctx *gin.Context) {
	var query ListQuery
	var logger = utils.SugarContext(ctx)
	if err := ctx.ShouldBindQuery(&query); err != nil {
		logger.Infow("Handler PictureBook List ctx.ShouldBindQuery err", "error", err)
		response.ParameterError(ctx)
		return
	}

	offset := int(query.Offset)
	limit := int(query.Limit)
	if limit <= 0 {
		limit = 10
	}

	result, err := h.service.List(ctx, query.Title, query.Type, query.Status, offset, limit)
	if err != nil {
		logger.Errorw("Handler PictureBook List service.List error", "error", err)
		response.InternalError(ctx)
		return
	}

	if result.Code != 0 {
		response.Failed(ctx, result.Code, result.Msg, result.Data)
		return
	}

	data, ok := result.Data.(pictureBookService.ListResponseData)
	if !ok {
		response.InternalError(ctx)
		return
	}

	response.SuccessfulWithPagination(
		ctx,
		data.List,
		&data.Offset,
		&data.Limit,
		func() *int { v := int(data.Count); return &v }(),
	)
}
