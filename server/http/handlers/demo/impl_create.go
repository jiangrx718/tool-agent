package demo

import (
	"io"
	"tool-agent/server/http/httputil"

	"github.com/gin-gonic/gin"
)

// DemoCreate 流式创建接口
//
// 通过 SSE（Server-Sent Events）将服务层返回的 channel 逐条推送给客户端。
// 客户端断开连接时，request context 会被取消，服务层 goroutine 随之退出。
func (h *DemoHandler) DemoCreate(ctx *gin.Context) {
	ch, err := h.service.ChatStream(ctx.Request.Context(), "")
	if err != nil {
		httputil.ServerError(ctx, err)
		return
	}
	
	ctx.Stream(func(w io.Writer) bool {
		if msg, ok := <-ch; ok {
			ctx.SSEvent("message", msg)
			return true
		}
		return false
	})

	return
}
