package middleware

import "github.com/gin-gonic/gin"

func EventStreamHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		// 禁止 nginx / 反向代理缓冲响应，确保每个 chunk 立即推送到客户端
		c.Writer.Header().Set("X-Accel-Buffering", "no")
		c.Next()
	}
}
