package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess      = 0
	CodeParameterErr = 400
	CodeInternalErr  = 500
)

// Response 统一响应结构
type Response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Data   any    `json:"data,omitempty"`
	Offset *int   `json:"offset,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
	Count  *int   `json:"count,omitempty"`
}

// Successful 返回成功响应
func Successful(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: "操作成功", Data: data})
}

// SuccessfulWithPagination 返回带分页的成功响应
func SuccessfulWithPagination(ctx *gin.Context, data any, offset *int, limit *int, count *int) {
	ctx.JSON(http.StatusOK, Response{
		Code:   CodeSuccess,
		Msg:    "操作成功",
		Data:   data,
		Count:  count,
		Limit:  limit,
		Offset: offset,
	})
}

// ParameterError 参数错误
func ParameterError(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeParameterErr, Msg: "参数错误"})
}

// InternalError 内部错误
func InternalError(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeInternalErr, Msg: "服务器错误"})
}

// Failed 返回失败响应
func Failed(ctx *gin.Context, code int, message string, data any) {
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: code, Msg: message, Data: data})
}
