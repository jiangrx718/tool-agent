package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// InitViper 初始化配置
func InitViper(appName, configFile string) error {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("app", appName)

	if err := LoadConfigInLocal(configFile); err != nil {
		return err
	}

	if Debug() {
		fmt.Printf("Loaded config from: %s\n", viper.ConfigFileUsed())
	}

	return nil
}

// LoadConfigInLocal 加载本地配置文件
func LoadConfigInLocal(filename string) error {
	if filename == "" {
		viper.SetConfigFile("config/app.yml")
	} else {
		viper.SetConfigFile(filename)
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// Env 当前环境
func Env() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

// Debug 是否调试模式
func Debug() bool {
	viper.SetDefault("debug", os.Getenv("DEBUG"))
	return viper.GetBool("debug")
}

// AppName 应用名称
func AppName() string {
	return viper.GetString("app")
}

// CtxValue 从 context 取值
func CtxValue(ctx context.Context, key string) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
