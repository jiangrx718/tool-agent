package utils

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitDB 从配置初始化数据库连接
func InitDB() error {
	dialect := viper.GetString("tools.dialect")
	dsn := viper.GetString("tools.dsn")

	if dialect == "" || dsn == "" {
		return fmt.Errorf("database config is empty, dialect: %s, dsn: %s", dialect, dsn)
	}

	var dialector gorm.Dialector
	switch dialect {
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		return fmt.Errorf("unsupported database dialect: %s", dialect)
	}

	gormLogger := logger.Default
	if Debug() {
		gormLogger = gormLogger.LogMode(logger.Info)
	} else {
		gormLogger = gormLogger.LogMode(logger.Silent)
	}

	instance, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("open database error: %w", err)
	}

	db = instance
	return nil
}

// DB 获取数据库实例
func DB() *gorm.DB {
	return db
}

// DBWithContext 获取带 context 的数据库实例
func DBWithContext(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}
