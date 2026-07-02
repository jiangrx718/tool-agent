package commands

import (
	"os"
	"os/signal"
	"syscall"

	"tool-agent/server/http/handlers"
	"tool-agent/utils"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

// Serve 启动 HTTP 服务
func Serve(ctx *cli.Context) error {
	srv := utils.NewHttpServer(viper.GetString("server.addr"))
	srv.RegisterHandler(handlers.NewHandler)

	listenCtx, stop := signal.NotifyContext(ctx.Context, os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv.GracefulStart(listenCtx)
	return nil
}
