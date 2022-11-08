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
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	influx_client "github.com/ztalab/ZACA/pkg/influxdb/influxdb-client/v2"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/database/mysql"
	"github.com/ztalab/ZACA/pkg/influxdb"
)

func mysqlDialer(config *core.Config, logger *core.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysqlDriver.Open(config.Mysql.Dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}
	d, _ := db.DB()
	if err = d.Ping(); err != nil {
		return nil, err
	}
	if err = mysql.Migrate(db); err != nil {
		logger.Errorf("MySQL Schema migrate error: %v", err)
	}
	return db, nil
}

func influxdbDialer(config *core.Config, logger *core.Logger) {
	if !config.Influxdb.Enabled {
		logger.Warn("Influxdb Function disabled")
		return
	}
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			metrics, err := func() (*influxdb.Metrics, error) {
				client, err := influx_client.NewHTTPClient(influx_client.HTTPConfig{
					Addr:                fmt.Sprintf("http://%v:%v", config.Influxdb.Address, config.Influxdb.Port),
					Username:            config.Influxdb.UserName,
					Password:            config.Influxdb.Password,
					MaxIdleConns:        config.Influxdb.MaxIdleConns,
					MaxIdleConnsPerHost: config.Influxdb.MaxIdleConns,
				})
				if err != nil {
					return nil, err
				}
				if _, _, err := client.Ping(1 * time.Second); err != nil {
					return nil, err
				}
				metrics := influxdb.NewMetrics(&influxdb.HTTPClient{
					Client: client,
					BatchPointsConfig: influx_client.BatchPointsConfig{
						Precision: config.Influxdb.Precision,
						Database:  config.Influxdb.Database,
					},
				}, &config.Influxdb)
				return metrics, nil
			}()
			if err != nil {
				logger.Error("Influxdb Init Fail:", err)
			} else {
				core.Is.Metrics = metrics
				return
			}
		}
	}
}

func vaultDialer(config *core.Config, logger *core.Logger) (*vaultAPI.Client, error) {
	conf := &vaultAPI.Config{
		Address: config.Vault.Addr,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 5 * time.Second,
		},
	}
	cli, err := vaultAPI.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "create Vault client")
	}
	cli.SetToken(config.Vault.Token)
	status, err := cli.Sys().SealStatus()
	if err != nil {
		logger.Errorf("vaule seal status err: %s", err)
		return nil, err
	}
	logger.Infof("sealed: %v, process: %v", status.Sealed, status.Progress)

	return cli, nil
}
