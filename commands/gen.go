package commands

import (
	"tool-agent/model"
	"tool-agent/utils"

	"github.com/urfave/cli/v2"
	"gorm.io/gen"
)

var generate = &cli.Command{
	Name: "generate",
	Action: func(ctx *cli.Context) error {
		conn := utils.DB()
		if conn == nil {
			return cli.Exit("database not initialized", 1)
		}

		g := gen.NewGenerator(gen.Config{
			OutPath: "internal/dao",
			Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		})

		g.UseDB(conn)

		g.ApplyBasic(
			model.SPictureBook{},
		)

		g.Execute()
		return nil
	},
}
