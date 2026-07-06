package utils

import (
	"context"
	"log"
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

	r := gin.Default()
	r.Use(gin.Logger())
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
func (s *HttpServer) GracefulStart(ctx context.Context) {
	go func() {
		// service connections
		log.Printf("Server listen on %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(c); err != nil {
		log.Printf("Server Shutdown: %s\n", err)
	}
	log.Println("Server exiting")
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
