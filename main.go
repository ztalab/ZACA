package main

import (
	"context"
	"github.com/urfave/cli"
	"gitlab.oneitfarm.com/bifrost/capitalizone/cmd"
	"gitlab.oneitfarm.com/bifrost/capitalizone/initer"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"os"
)

func main() {
	ctx := context.Background()

	app := cli.NewApp()
	app.Name = "capitalizone"
	app.Version = "1.0.0"
	app.Usage = "capitalizone"
	app.Commands = []cli.Command{
		newHttpCmd(ctx),
		newTlsCmd(ctx),
		newOcspCmd(ctx),
	}
	app.Before = initer.Init
	err := app.Run(os.Args)
	if err != nil {
		logger.Named("Init").Errorf(err.Error())
	}
}

// newHttpCmd 运行http服务
func newHttpCmd(ctx context.Context) cli.Command {
	return cli.Command{
		Name:  "http",
		Usage: "运行http服务",
		Action: func(c *cli.Context) error {
			return cmd.RunHttp(ctx)
		},
	}
}

// newTlsCmd 运行tls服务
func newTlsCmd(ctx context.Context) cli.Command {
	return cli.Command{
		Name:  "tls",
		Usage: "运行tls服务",
		Action: func(c *cli.Context) error {
			return cmd.RunTls(ctx)
		},
	}
}

// newOcspCmd 运行tls服务
func newOcspCmd(ctx context.Context) cli.Command {
	return cli.Command{
		Name:  "ocsp",
		Usage: "运行ocsp服务",
		Action: func(c *cli.Context) error {
			return cmd.RunOcsp(ctx)
		},
	}
}
