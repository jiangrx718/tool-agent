package commands

import (
	"tool-agent/commands/demo"
	"tool-agent/utils/graceful"

	"github.com/urfave/cli/v2"
)

// Worker go run main.go worker start
func Worker() *cli.Command {
	return &cli.Command{
		Name: "worker",
		Subcommands: []*cli.Command{
			{
				Name: "start",
				Action: func(ctx *cli.Context) error {
					graceful.Start(demo.InitWorkerDemo()) // 示例定时任务
					graceful.Wait()
					return nil
				},
			},
		},
	}
}
