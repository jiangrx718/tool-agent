package main

import (
	"os"

	"tool-agent/commands"
	"tool-agent/utils"

	"github.com/urfave/cli/v2"
)

var configFile string

var Version = "local"

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Action = commands.Serve
	app.Before = initConfig
	app.After = flush
	app.Commands = commands.AllCommands()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Value:       "", // 默认从 config 目录读取
			Usage:       "specify the location of the configuration file",
			Required:    false,
			Destination: &configFile,
		},
	}

	if err := app.Run(os.Args); err != nil {
		utils.Sugar().Fatal(err)
	}
}

// initConfig 初始化配置
func initConfig(*cli.Context) error {
	if err := utils.InitViper("tool-agent", configFile); err != nil {
		return err
	}

	utils.InitFromViper()

	if err := utils.InitDB(); err != nil {
		return err
	}

	return nil
}

// flush 退出前刷新日志缓冲
func flush(*cli.Context) error {
	utils.Flush()
	return nil
}
