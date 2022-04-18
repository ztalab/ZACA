package datastore

import (
	"time"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/vaultsecret"
	v2 "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

// RunMigration 迁移 MySQL 数据到 Vault
func RunMigration() {
	v2.Debug("MySQL -> Vault 数据库迁移")
	certRows := make([]*model.Certificates, 0)
	result := core.Is.Db.Model(&model.Certificates{}).Where("expiry > ? AND revoked_at is NULL", time.Now()).Limit(10000).
		Find(&certRows)
	for _, row := range certRows {
		if row.Pem == "" {
			continue
		}
		if pemStr, err := core.Is.VaultSecret.GetCertPEM(row.SerialNumber); err != nil || *pemStr == "" {
			core.Is.Logger.Debugf("Vault 迁移 %s", row.SerialNumber)
			if err := core.Is.VaultSecret.StoreCertPEM(row.SerialNumber, row.Pem); err != nil {
				core.Is.Logger.Errorf("Vault store cert %s 错误: %s", row.SerialNumber, err)
			}
		}
	}
	if result.Error != nil {
		core.Is.Logger.Errorf("迁移 MySQL 到 Vault 错误: %s", result.Error)
	}

	caKeyPair := new(model.SelfKeypair)
	if err := core.Is.Db.Model(&model.SelfKeypair{}).Where("name = ?", "ca").First(caKeyPair).Error; err == nil {
		if err := core.Is.VaultSecret.StoreCertPEMKey(vaultsecret.CALocalStoreKey,
			caKeyPair.Certificate.String, caKeyPair.PrivateKey.String); err != nil {
			core.Is.Logger.Errorf("Vault ca cert 储存错误: %s", err)
		}
	}

	trustKeyPair := new(model.SelfKeypair)
	if err := core.Is.Db.Model(&model.SelfKeypair{}).Where("name = ?", "trust").First(trustKeyPair).Error; err == nil {
		if err := core.Is.VaultSecret.StoreCertPEM(vaultsecret.CATructCertsKey, trustKeyPair.Certificate.String); err != nil {
			core.Is.Logger.Errorf("Vault trust cert 储存错误: %s", err)
		}
	}
}
