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
	"flag"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strings"

	"github.com/spf13/viper"
	cfssl_config "github.com/ztalab/cfssl/config"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/core/config"
	"github.com/ztalab/ZACA/pkg/influxdb"
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

func parseConfigs(c *cli.Context) (core.Config, error) {
	// Cmdline flags
	flag.Parse()
	// Default config
	viper.SetConfigName(fmt.Sprintf("%v", ConfName))
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return core.Config{}, err
	}

	// Merge config frm ENV
	viper.SetEnvPrefix(G_)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	hs, err := os.Hostname()
	if err != nil {
		return core.Config{}, err
	}
	hostname := hs
	if v := os.Getenv("HOSTNAME"); v != "" {
		hostname = v
	}

	conf := core.Config{IConfig: config.IConfig{
		Log: config.Log{
			LogProxy: config.LogProxy{
				Host: viper.GetString("log.log-proxy.host"),
				Port: viper.GetInt("log.log-proxy.port"),
				Key:  viper.GetString("log.log-proxy.key"),
			},
		},
		Keymanager: config.Keymanager{
			UpperCa:  viper.GetStringSlice("keymanager.upper-ca"),
			SelfSign: viper.GetBool("keymanager.self-sign"),
			CsrTemplates: config.CsrTemplates{
				RootCa: config.RootCa{
					O:      viper.GetString("keymanager.csr-templates.root-ca.o"),
					Expiry: viper.GetString("keymanager.csr-templates.root-ca.expiry"),
				},
				IntermediateCa: config.IntermediateCa{
					O:      viper.GetString("keymanager.csr-templates.intermediate-ca.o"),
					Ou:     viper.GetString("keymanager.csr-templates.intermediate-ca.ou"),
					Expiry: viper.GetString("keymanager.csr-templates.intermediate-ca.expiry"),
				},
			},
		},
		Singleca: config.Singleca{
			ConfigPath: viper.GetString("singleca.config-path"),
		},
		OCSPHost: viper.GetString("ocsp-host"),
		HTTP: config.HTTP{
			OcspListen: viper.GetString("http.ocsp-listen"),
			CaListen:   viper.GetString("http.ca-listen"),
			Listen:     viper.GetString("http.listen"),
		},
		Mysql: config.Mysql{
			Dsn: viper.GetString("mysql.dsn"),
		},
		Vault: config.Vault{
			Enabled: viper.GetBool("vault.enabled"),
			Addr:    viper.GetString("vault.addr"),
			Token:   viper.GetString("vault.token"),
			Prefix:  viper.GetString("vault.prefix"),
		},
		Influxdb: influxdb.CustomConfig{
			Enabled:             viper.GetBool("influxdb.enabled"),
			Address:             viper.GetString("influxdb.address"),
			Port:                viper.GetInt("influxdb.port"),
			UDPAddress:          viper.GetString("influxdb.udp_address"),
			Database:            viper.GetString("influxdb.database"),
			Precision:           viper.GetString("influxdb.precision"),
			UserName:            viper.GetString("influxdb.username"),
			Password:            viper.GetString("influxdb.password"),
			ReadUserName:        viper.GetString("influxdb.read-username"),
			ReadPassword:        viper.GetString("influxdb.read-password"),
			MaxIdleConns:        viper.GetInt("influxdb.max-idle-conns"),
			MaxIdleConnsPerHost: viper.GetInt("influxdb.max-idle-conns-per-host"),
			IdleConnTimeout:     viper.GetInt("influxdb.idle-conn-timeout"),
			FlushSize:           viper.GetInt("influxdb.flush-size"),
			FlushTime:           viper.GetInt("influxdb.flush-time"),
		},
		SwaggerEnabled: viper.GetBool("swagger-enabled"),
		Debug:          viper.GetBool("debug"),
		Version:        viper.GetString("version"),
		Hostname:       hostname,
		Ocsp: config.Ocsp{
			CacheTime: viper.GetInt("ocsp.cache-time"),
		},
	}}

	// ref: https://github.com/golang-migrate/migrate/issues/313
	if !strings.Contains(conf.Mysql.Dsn, "multiStatements") {
		conf.Mysql.Dsn += "&multiStatements=true"
	}

	cfg, err := cfssl_config.LoadFile(conf.Singleca.ConfigPath)
	if err != nil {
		return conf, fmt.Errorf("cfssl configuration file %s Error: %s", conf.Singleca.ConfigPath, err)
	}

	cfg.Signing.Default.OCSP = conf.OCSPHost

	conf.Singleca.CfsslConfig = cfg

	return conf, nil
}
