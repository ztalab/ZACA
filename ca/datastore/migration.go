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

package datastore

import (
	"github.com/ztalab/ZACA/pkg/logger"
	"time"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/database/mysql/cfssl-model/model"
	"github.com/ztalab/ZACA/pkg/vaultsecret"
)

// RunMigration Migrate MySQL data to vault
func RunMigration() {
	logger.Debug("MySQL -> Vault Database migration")
	certRows := make([]*model.Certificates, 0)
	result := core.Is.Db.Model(&model.Certificates{}).Where("expiry > ? AND revoked_at is NULL", time.Now()).Limit(10000).
		Find(&certRows)
	for _, row := range certRows {
		if row.Pem == "" {
			continue
		}
		if pemStr, err := core.Is.VaultSecret.GetCertPEM(row.SerialNumber); err != nil || *pemStr == "" {
			core.Is.Logger.Debugf("Vault Transfer %s", row.SerialNumber)
			if err := core.Is.VaultSecret.StoreCertPEM(row.SerialNumber, row.Pem); err != nil {
				core.Is.Logger.Errorf("Vault store cert %s Error: %s", row.SerialNumber, err)
			}
		}
	}
	if result.Error != nil {
		core.Is.Logger.Errorf("Error migrating Mysql to vault: %s", result.Error)
	}

	caKeyPair := new(model.SelfKeypair)
	if err := core.Is.Db.Model(&model.SelfKeypair{}).Where("name = ?", "ca").First(caKeyPair).Error; err == nil {
		if err := core.Is.VaultSecret.StoreCertPEMKey(vaultsecret.CALocalStoreKey,
			caKeyPair.Certificate.String, caKeyPair.PrivateKey.String); err != nil {
			core.Is.Logger.Errorf("Vault ca cert Storage error: %s", err)
		}
	}

	trustKeyPair := new(model.SelfKeypair)
	if err := core.Is.Db.Model(&model.SelfKeypair{}).Where("name = ?", "trust").First(trustKeyPair).Error; err == nil {
		if err := core.Is.VaultSecret.StoreCertPEM(vaultsecret.CATructCertsKey, trustKeyPair.Certificate.String); err != nil {
			core.Is.Logger.Errorf("Vault trust cert Storage error: %s", err)
		}
	}
}
