package main

import (
	"os"

	_ "github.com/cloudslit/cloudslit/fullnode/docs"
	"github.com/cloudslit/cloudslit/fullnode/pkg/logger"
	"github.com/cloudslit/cloudslit/fullnode/server"
	"github.com/urfave/cli"
	_ "go.uber.org/automaxprocs"
)

// @title ZAManager API
// @version 1.0.0
// @description This is ZAManager api list.
// @host 127.0.0.1:80
// @BasePath /api/v1
func main() {
	app := cli.NewApp()
	app.Name = "ZAManager"
	app.Author = "TS"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Value: "http",
			Usage: "run server type: http",
		},
		cli.StringFlag{
			Name:  "c",
			Value: "config.yaml",
			Usage: "config file url",
		},
	}
	app.Before = server.InitService
	app.Action = func(c *cli.Context) error {
		println("RunHttp Server.")
		serverType := c.String("server")
		switch serverType {
		case "http":
			server.RunHTTP()
		default:
			server.RunHTTP()
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		logger.Errorf(nil, "app run error : %v", err)
		panic("app run error:" + err.Error())
	}
}