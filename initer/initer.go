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
	"github.com/urfave/cli"
	"github.com/ztalab/ZACA/ca/datastore"
	"github.com/ztalab/ZACA/ca/keymanager"
	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/ZACA/pkg/vaultsecret"
	"github.com/ztalab/cfssl/hook"
	"log"
	"os"

	// ...
	_ "github.com/ztalab/ZACA/util"
)

// Init Initialization
func Init(c *cli.Context) error {
	conf, err := InitConfigs(c, "conf.yml")
	if err != nil {
		return err
	}
	initLogger(&conf)
	log.Printf("started with conf: %+v", conf)

	hook.EnableVaultStorage = conf.Vault.Enabled

	l := &core.Logger{Logger: logger.S()}

	db, err := mysqlDialer(&conf, l)
	if err != nil {
		logger.Fatal(err)
	}
	i := &core.I{
		Config: &conf,
		Logger: l,
		Db:     db,
	}

	if hook.EnableVaultStorage {
		logger.Info("Enable vault encrypted storage engine")
		vaultClient, err := vaultDialer(&conf, l)
		if err != nil {
			logger.Fatal(err)
			return err
		}
		i.VaultClient = vaultClient
		i.VaultSecret = vaultsecret.NewVaultSecret(vaultClient, conf.Vault.Prefix)
	}

	core.Is = i
	// Initialize incluxdb
	go influxdbDialer(&conf, l)

	if os.Getenv("IS_MIGRATION") == "true" {
		datastore.RunMigration()
		os.Exit(1)
	}
	// CA Start
	if err := keymanager.InitKeeper(); err != nil {
		return err
	}

	logger.Info("success started.")
	return nil
}
