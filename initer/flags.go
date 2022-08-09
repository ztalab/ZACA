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

package initer

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strings"

	cfssl_config "github.com/ztalab/cfssl/config"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/core/config"
	"gopkg.in/yaml.v3"
)

const (
	G_       = "IS"
	ConfName = "conf"
)

func InitConfigs(c *cli.Context, configURL string) (core.Config, error) {
	var conf config.IConfig
	f, err := os.Open(configURL)
	if err != nil {
		return core.Config{}, err
	}
	if err = yaml.NewDecoder(f).Decode(&conf); err != nil {
		return core.Config{}, err
	}
	if v := os.Getenv("IS_ENV"); v != "" {
		conf.Debug = true
	}
	if v := os.Getenv("IS_SINGLECA_CONFIG_PATH"); v != "" {
		conf.Singleca.ConfigPath = v
	}
	if v := os.Getenv("IS_MYSQL_DSN"); v != "" {
		conf.Mysql.Dsn = v
	}
	if v := os.Getenv("IS_KEYMANAGER_UPPER_CA"); v != "" {
		conf.Keymanager.UpperCa = strings.Split(v, " ")
	}
	if v := os.Getenv("IS_HTTP_CA_LISTEN"); v != "" {
		conf.HTTP.CaListen = v
	}
	if v := os.Getenv("IS_KEYMANAGER_SELF_SIGN"); v != "true" {
		conf.Keymanager.SelfSign = true
	}
	// ref: https://github.com/golang-migrate/migrate/issues/313
	if !strings.Contains(conf.Mysql.Dsn, "multiStatements") {
		conf.Mysql.Dsn += "&multiStatements=true"
	}

	cfg, err := cfssl_config.LoadFile(conf.Singleca.ConfigPath)
	if err != nil {
		return core.Config{}, fmt.Errorf("cfssl configuration file %s Error: %s", conf.Singleca.ConfigPath, err)
	}
	cfg.Signing.Default.OCSP = conf.OCSPHost
	conf.Singleca.CfsslConfig = cfg

	return core.Config{
		IConfig: conf,
	}, nil
}
