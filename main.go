/*
Copyright 2022-present The Ztalab Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"github.com/urfave/cli"
	"github.com/ztalab/ZACA/cmd"
	"github.com/ztalab/ZACA/initer"
	"github.com/ztalab/ZACA/pkg/logger"
	"os"
)

func main() {
	ctx := context.Background()

	app := cli.NewApp()
	app.Name = "capitalizone"
	app.Version = "1.0.0"
	app.Usage = "capitalizone"
	app.Commands = []cli.Command{
		newApiCmd(ctx),
		newTlsCmd(ctx),
		newOcspCmd(ctx),
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "conf",
			Usage: "配置文件(.yaml)",
			Value: "conf.yml",
		},
	}
	app.Before = initer.Init
	err := app.Run(os.Args)
	if err != nil {
		logger.Named("Init").Errorf(err.Error())
	}
}

// newApiCmd Running API services
func newApiCmd(ctx context.Context) cli.Command {
	return cli.Command{
		Name:  "api",
		Usage: "Running API service",
		Action: func(c *cli.Context) error {
			return cmd.RunHttp(ctx)
		},
	}
}

// newTlsCmd Running TLS service
func newTlsCmd(ctx context.Context) cli.Command {
	return cli.Command{
		Name:  "tls",
		Usage: "Running TLS service",
		Action: func(c *cli.Context) error {
			return cmd.RunTls(ctx)
		},
	}
}

// newOcspCmd Running OCSP service
func newOcspCmd(ctx context.Context) cli.Command {
	return cli.Command{
		Name:  "ocsp",
		Usage: "Run OCSP service",
		Action: func(c *cli.Context) error {
			return cmd.RunOcsp(ctx)
		},
	}
}
