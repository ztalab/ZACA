package certmanager

import (
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/mesh"
	"gitlab.oneitfarm.com/bifrost/capitalizone/util"
)

// CertCleaner ...
type CertCleaner struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

// NewCertCleaner ...
func NewCertCleaner() *CertCleaner {
	return &CertCleaner{
		logger: v2log.Named("cleaner").SugaredLogger,
		db:     core.Is.Db,
	}
}

// AutoGC ...
func (cc *CertCleaner) AutoGC() {
	cc.logger.Info("开启自动 GC")
	t := time.NewTicker(time.Hour)
	for {
		cc.logger.Info("执行自动清理")
		cc.GarbageCollect()
		<-t.C
	}
}

// GarbageCollect ...
func (cc *CertCleaner) GarbageCollect() {
	cc.logger.Debug("获取 Mesh 元数据")
	body, err := mesh.GetAllDynamicServiceMetadataRaw()
	if err != nil {
		cc.logger.Errorf("获取 Mesh 元数据错误: %s", err)
		return
	}
	cc.logger.Debugf("Mesh 元数据: %s", string(body))
	runtimeUniqueIDs := make([]string, 0)
	for _, uid := range gjson.GetBytes(body, "data.list.#.unique_id").Array() {
		runtimeUniqueIDs = append(runtimeUniqueIDs, uid.String())
	}
	cc.logger.With("unique_ids", runtimeUniqueIDs).Info("清理下线 Sidecar 证书")
	if err := cc.cleanDownSidecarCerts(runtimeUniqueIDs); err != nil {
		cc.logger.Errorf("清理下线 Sidecar 证书: %s", err)
	}
	if err := cc.cleanRevokedSidecarCerts(); err != nil {
		cc.logger.Errorf("清理 Sidecar 主动吊销证书: %s", err)
	}
}

// 若模式为 Vault 储存, 删除 vault 对应 KV
func (cc *CertCleaner) cleanDownSidecarCerts(runtimeUniqueIDs []string) error {
	query := cc.db.Model(&model.Certificates{}).
		Where("ca_label = ?", caclient.RoleSidecar).
		Select("common_name").
		Group("common_name")
	var certs []model.Certificates
	if err := query.Find(&certs).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "证书查询错误")
	}
	dbUniqueIDs := make([]string, 0, len(certs))
	for _, cert := range certs {
		dbUniqueIDs = append(dbUniqueIDs, cert.CommonName.String)
	}
	// cc.logger.With("unique_ids", dbUniqueIDs).Debug("db unique_ids")

	deleteIDs := cast.ToStringSlice(
		mapset.NewSetFromSlice(util.StringSliceToInterfaceSlice(dbUniqueIDs)).
			Difference(mapset.NewSetFromSlice(util.StringSliceToInterfaceSlice(runtimeUniqueIDs))).
			ToSlice())

	if len(deleteIDs) == 0 {
		return nil
	}

	deleteIDMap := make(map[string]bool, len(deleteIDs))
	for _, deleteID := range deleteIDs {
		deleteIDMap[deleteID] = true
	}

	if hook.EnableVaultStorage {
		for _, certRow := range certs {
			if _, ok := deleteIDMap[certRow.CaLabel.String]; ok {
				if err := core.Is.VaultSecret.DeleteCertPEM(certRow.SerialNumber); err != nil {
					core.Is.Logger.Warnf("vault 删除错误: %s, sn: %s", err, certRow.SerialNumber)
				}
			}
		}
	}

	cc.logger.With("unique_ids", deleteIDs).Info("清理下线 Sidecar 证书")
	if err := cc.db.Where("common_name IN (?)", deleteIDs).Delete(&model.Certificates{}).Error; err != nil {
		return errors.Wrap(err, "证书批量删除出错")
	}

	return nil
}

// 若模式为 Vault 储存, 删除 vault 对应 KV
func (cc *CertCleaner) cleanRevokedSidecarCerts() error {
	query := cc.db.Where("ca_label = ?", caclient.RoleSidecar).
		Where("status = ?", "revoked").
		Where("reason = ?", 1)
	var certs []model.Certificates
	if err := query.Find(&certs).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "证书查询错误")
	}

	if hook.EnableVaultStorage {
		for _, certRow := range certs {
			if err := core.Is.VaultSecret.DeleteCertPEM(certRow.SerialNumber); err != nil {
				core.Is.Logger.Warnf("vault 删除错误: %s, sn: %s", err, certRow.SerialNumber)
			}
		}
	}

	err := query.
		Delete(&model.Certificates{}).Error
	if err != nil {
		return errors.Wrap(err, "证书批量删除出错")
	}
	cc.logger.Info("清理 Sidecar 主动吊销证书")
	return nil
}
