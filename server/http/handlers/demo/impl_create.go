package demo

import (
	"io"
	"tool-agent/server/http/httputil"

	"github.com/gin-gonic/gin"
)

func (h *DemoHandler) DemoCreate(ctx *gin.Context) {

	ch, err := h.service.ChatStream(ctx, "")
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
