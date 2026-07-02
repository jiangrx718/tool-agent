package demo

import (
	"context"

	"tool-agent/utils"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

// WorkerDemo 示例定时任务 worker
type WorkerDemo struct {
	client *cron.Cron
}

// InitWorkerDemo 创建示例定时任务 worker
func InitWorkerDemo() *WorkerDemo {
	return &WorkerDemo{
		client: cron.New(
			cron.WithChain(
				cron.DelayIfStillRunning(cron.DefaultLogger),
			),
		),
	}
}

// GracefulStart 启动定时任务,context 取消时优雅停止
func (w *WorkerDemo) GracefulStart(ctx context.Context) {
	utils.Sugar().Debug("开启示例定时任务")

	cronExpr := viper.GetString("demo.cron")
	if cronExpr == "" {
		cronExpr = "*/1 * * * *"
	}

	id, err := w.client.AddJob(cronExpr, NewDemo())
	if err != nil {
		utils.Sugar().Infow("create demo crontab", "error", err, "entry_id", id)
		return
	}

	w.client.Start()
	<-ctx.Done()
	w.client.Stop()
	w.client.Remove(id)
	utils.Sugar().Debug("关闭示例定时任务")
}
