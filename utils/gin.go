package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HttpServerHandler HTTP 服务处理器接口
type HttpServerHandler interface {
	RegisterRoutes()
}

// HttpServer HTTP 服务
type HttpServer struct {
	http.Server
	router   *gin.Engine
	handlers []HttpServerHandler
}

// NewHttpServer 创建 HTTP 服务
func NewHttpServer(listen string) *HttpServer {
	if Debug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	srv := &HttpServer{
		router: r,
		Server: http.Server{
			Addr:    listen,
			Handler: r,
		},
		handlers: []HttpServerHandler{},
	}

	return srv
}

// RegisterHandler 注册路由处理器
func (s *HttpServer) RegisterHandler(funcs ...func(*gin.Engine) HttpServerHandler) {
	for _, fun := range funcs {
		s.handlers = append(s.handlers, fun(s.router))
	}
}

// GracefulStart 优雅启动
func (s *HttpServer) GracefulStart(ctx context.Context) error {
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Sugar().Errorf("Server listen error: %s", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.Shutdown(shutdownCtx)
}

// RequestID 请求 ID 中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("x-request-id")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("x-request-id", id)
		c.Next()
		c.Header("x-request-id", id)
	}
}
