package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Graceful 优雅启动接口,实现该接口的对象可由 Start 托管生命周期
type Graceful interface {
	GracefulStart(ctx context.Context)
}

var (
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
)

// initContext 惰性创建一个由 SIGINT/SIGTERM 信号驱动的共享 context
func initContext() {
	once.Do(func() {
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	})
}

// Start 非阻塞地启动一个 Graceful 服务,在独立 goroutine 中运行
func Start(srv Graceful) {
	initContext()
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.GracefulStart(ctx)
	}()
}

// Wait 阻塞等待信号,收到信号后取消共享 context,并等待所有已启动的服务退出
func Wait() {
	initContext()
	<-ctx.Done()
	wg.Wait()
	if cancel != nil {
		cancel()
	}
}
