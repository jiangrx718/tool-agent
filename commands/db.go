package commands

import (
	"tool-agent/model"
	"tool-agent/utils"

	"github.com/urfave/cli/v2"
)

// go  run main.go db migrate
var dbCommand = &cli.Command{
	Name:  "db",
	Usage: "数据库操作",
	Subcommands: []*cli.Command{
		{
			Name:  "migrate",
			Usage: "自动创建/更新数据表",
			Action: func(ctx *cli.Context) error {
				db := utils.DB()
				if db == nil {
					return cli.Exit("database not initialized", 1)
				}

				if err := db.AutoMigrate(
					&model.SPictureBook{},
					&model.KbDocument{},
				); err != nil {
					utils.Sugar().Errorf("AutoMigrate error %v", err)
					return err
				}

				utils.Sugar().Info("AutoMigrate success")
				return nil
			},
		},
	},
}
