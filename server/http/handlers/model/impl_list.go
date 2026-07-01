package model

import (
	modelService "tool-agent/internal/service/model"
	"tool-agent/server/http/httputil"
	"tool-agent/server/http/response"
	"tool-agent/utils"

	"github.com/gin-gonic/gin"
)

// ListQuery 模型列表查询参数
type ListQuery struct {
	httputil.Pagination
	Keyword string `form:"keyword"`
}

// List 模型列表接口
func (mh *ModelHandler) List(ctx *gin.Context) {
	query := ListQuery{}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.SugarContext(ctx).Infow("Handler Model List ctx.ShouldBindQuery err", "error", err)
		response.ParameterError(ctx)
		return
	}

	offset := int(query.Offset)
	limit := int(query.Limit)
	if limit <= 0 {
		limit = 10
	}

	result, err := mh.service.List(ctx, query.Keyword, offset, limit)
	if err != nil {
		utils.SugarContext(ctx).Errorw("Handler Model List service.List error", "error", err)
		response.InternalError(ctx)
		return
	}

	if result.Code != 0 {
		response.Failed(ctx, result.Code, result.Msg, result.Data)
		return
	}

	data, ok := result.Data.(modelService.ListResponseData)
	if !ok {
		response.InternalError(ctx)
		return
	}

	response.SuccessfulWithPagination(
		ctx,
		data.List,
		&offset,
		&limit,
		func() *int { v := int(data.Count); return &v }(),
	)
}
