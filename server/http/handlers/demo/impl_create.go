package demo

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// 先写入响应头并 flush，让客户端立即收到 SSE 连接已建立
	ctx.Writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-ctx.Request.Context().Done():
			// 客户端断开连接
			return
		case msg, ok := <-ch:
			if !ok {
				// channel 关闭，流结束
				return
			}
			data, _ := json.Marshal(msg)
			fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
