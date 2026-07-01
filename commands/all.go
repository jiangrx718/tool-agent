package commands

import "github.com/urfave/cli/v2"

// AllCommands 返回所有命令
func AllCommands() []*cli.Command {
	return []*cli.Command{
		generate,
		dbCommand,
	}
}
