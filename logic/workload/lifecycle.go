// Package workload 证书生命周期管理
package workload

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/ocsp"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/dao"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/events"
)

type RevokeCertsParams struct {
	SN       string `json:"sn"`
	AKI      string `json:"aki"`
	UniqueId string `json:"unique_id"`
}

// RevokeCerts 吊销证书
// 	1. 通过 SN/AKI 吊销证书
//  2. 通过 UniqueId 统一吊销证书
func (l *Logic) RevokeCerts(params *RevokeCertsParams) error {
	// 1. 通过标识找到证书
	db := l.db.Session(&gorm.Session{})

	db = db.Where("status = ?", "good").
		Where("expiry > ?", time.Now())

	if params.UniqueId != "" {
		db = db.Where("common_name = ?", params.UniqueId)
	} else if params.AKI != "" && params.SN != "" {
		db = db.Where("serial_number = ? AND authority_key_identifier = ?", params.SN, params.AKI)
	} else {
		return errors.New("参数错误")
	}

	certs, _, err := dao.GetAllCertificates(db, 1, 1000, "issued_at desc")
	if err != nil {
		l.logger.With("params", params).Errorf("数据库查询错误: %s", err)
		return errors.Wrap(err, "数据库查询错误")
	}

	if len(certs) == 0 {
		return errors.New("未找到证书")
	}

	// 2. 批量吊销证书
	reason, _ := ocsp.ReasonStringToCode("cacompromise")
	err = l.db.Transaction(func(tx *gorm.DB) error {
		for _, cert := range certs {
			err := tx.Model(&model.Certificates{}).Where(&model.Certificates{
				SerialNumber:           cert.SerialNumber,
				AuthorityKeyIdentifier: cert.AuthorityKeyIdentifier,
			}).Update("status", "revoked").
				Update("reason", reason).
				Update("revoked_at", time.Now()).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.logger.Errorf("批量吊销证书错误: %s", err)
		return errors.Wrap(err, "批量吊销证书错误")
	}

	// 3. 记录操作日志
	for _, cert := range certs {
		events.NewWorkloadLifeCycle("revoke", events.OperatorMSP, events.CertOp{
			UniqueId: cert.CommonName.String,
			SN:       cert.SerialNumber,
			AKI:      cert.AuthorityKeyIdentifier,
		}).Log()
	}

	return nil
}

type RecoverCertsParams struct {
	SN       string `json:"sn"`
	AKI      string `json:"aki"`
	UniqueId string `json:"unique_id"`
}

// RecoverCerts 恢复证书
// 	1. 通过 SN/AKI 恢复证书
//  2. 通过 UniqueId 统一恢复证书
func (l *Logic) RecoverCerts(params *RecoverCertsParams) error {
	// 1. 通过标识找到证书
	db := l.db.Session(&gorm.Session{})

	db = db.Where("status = ?", "revoked").
		Where("expiry > ?", time.Now())

	switch {
	case params.UniqueId != "":
		db = db.Where("common_name = ?", params.UniqueId)
	case params.AKI != "" && params.SN != "":
		db = db.Where("serial_number = ? AND authority_key_identifier = ?", params.SN, params.AKI)
	default:
		return errors.New("参数错误")
	}

	certs, _, err := dao.GetAllCertificates(db, 1, 1000, "issued_at desc")
	if err != nil {
		l.logger.With("params", params).Errorf("数据库查询错误: %s", err)
		return errors.Wrap(err, "数据库查询错误")
	}

	if len(certs) == 0 {
		return errors.New("未找到证书")
	}

	// 2. 批量恢复证书
	err = l.db.Transaction(func(tx *gorm.DB) error {
		for _, cert := range certs {
			err := tx.Model(&model.Certificates{}).Where(&model.Certificates{
				SerialNumber:           cert.SerialNumber,
				AuthorityKeyIdentifier: cert.AuthorityKeyIdentifier,
			}).Update("status", "good").
				Update("reason", 0).
				Update("revoked_at", nil).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 3. 记录操作日志
	for _, cert := range certs {
		events.NewWorkloadLifeCycle("recover", events.OperatorMSP, events.CertOp{
			UniqueId: cert.CommonName.String,
			SN:       cert.SerialNumber,
			AKI:      cert.AuthorityKeyIdentifier,
		}).Log()
	}

	return nil
}

type ForbidNewCertsParams struct {
	UniqueIds []string `json:"unique_ids"`
}

// ForbidNewCerts 禁止某个 UniqueID 申请证书
//	1. 禁止 UniqueId 申请新证书
//  2. 日志记录
func (l *Logic) ForbidNewCerts(params *ForbidNewCertsParams) error {
	err := l.db.Transaction(func(tx *gorm.DB) error {
		for _, uid := range params.UniqueIds {
			record := model.Forbid{
				UniqueID:  uid,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_, _, err := dao.AddForbid(tx, &record)
			if err != nil {
				l.logger.With("record", record).Errorf("数据库插入错误: %s", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.logger.Errorf("数据库插入错误: %s", err)
		return err
	}

	// 日志记录
	for _, uid := range params.UniqueIds {
		events.NewWorkloadLifeCycle("forbid", events.OperatorMSP, events.CertOp{
			UniqueId: uid,
		}).Log()
	}

	return nil
}

// RecoverForbidNewCerts 恢复允许某个 UniqueID 申请证书
//	1. 允许 UniqueId 申请新证书
func (l *Logic) RecoverForbidNewCerts(params *ForbidNewCertsParams) error {
	err := l.db.Transaction(func(tx *gorm.DB) error {
		for _, uid := range params.UniqueIds {
			err := tx.Model(&model.Forbid{}).Where("unique_id = ?", uid).
				Where("deleted_at IS NULL").
				Update("deleted_at", time.Now()).Error
			if err != nil {
				l.logger.With("unique_id", uid).Errorf("数据库更新错误: %s", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.logger.Errorf("数据库更新错误: %s", err)
		return err
	}

	// 日志记录
	for _, uid := range params.UniqueIds {
		events.NewWorkloadLifeCycle("recover-forbid", events.OperatorMSP, events.CertOp{
			UniqueId: uid,
		}).Log()
	}

	return nil
}
