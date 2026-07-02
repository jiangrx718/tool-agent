package demo

import (
	"tool-agent/utils"
)

// Demo 示例定时任务,实现 cron.Job 接口
type Demo struct{}

// NewDemo 创建示例任务实例
func NewDemo() *Demo {
	return &Demo{}
}

// Run 定时任务执行入口
func (d *Demo) Run() {
	utils.Sugar().Infow("示例定时任务执行")
}
