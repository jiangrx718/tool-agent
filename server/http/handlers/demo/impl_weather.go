package demo

import (
	"io"
	"tool-agent/server/http/httputil"

	"github.com/gin-gonic/gin"
)

type ChatReq struct {
	Question string `json:"question" binding:"required"`
}

// DemoWeather 天气查询流式接口
//
// 通过 SSE（Server-Sent Events）将天气智能体返回的 channel 逐条推送给客户端。
// 客户端断开连接时，request context 会被取消，服务层 goroutine 随之退出。
func (h *DemoHandler) DemoWeather(ctx *gin.Context) {
	var reqBody ChatReq
	if err := ctx.Bind(&reqBody); err != nil {
		httputil.BadRequest(ctx, err)
		return
	}
	ch, err := h.service.WeatherStream(ctx.Request.Context(), reqBody.Question)
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
