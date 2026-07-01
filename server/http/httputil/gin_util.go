package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HttpError HTTP 错误响应
type HttpError struct {
	Code int    `json:"code" example:"400"`
	Msg  string `json:"msg"  example:"参数错误"`
}

// BadRequest 400 错误
func BadRequest(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusOK, HttpError{
		Code: 400,
		Msg:  err.Error(),
	})
}

// ServerError 500 错误
func ServerError(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusOK, HttpError{
		Code: 500,
		Msg:  err.Error(),
	})
}

// Pagination 分页参数
type Pagination struct {
	Limit  int64 `json:"limit" form:"limit,default=10"`
	Offset int64 `json:"offset" form:"offset,default=0"`
}
