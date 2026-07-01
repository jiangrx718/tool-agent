package utils

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

// InitLogger 初始化日志
func InitLogger() error {
	cfg := zap.NewProductionConfig()
	if viper.GetString("log.level") == "debug" || Debug() {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	sugar = logger.Sugar()
	return nil
}

// Sugar 获取 SugaredLogger
func Sugar() *zap.SugaredLogger {
	if sugar == nil {
		logger, _ := zap.NewProduction()
		sugar = logger.Sugar()
	}
	return sugar
}

// SugarContext 获取带 context 的 SugaredLogger
func SugarContext(ctx context.Context) *zap.SugaredLogger {
	return Sugar()
}
