// Package ca workload 相关
package ca

import (
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/tal-tech/go-zero/core/fx"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/dao"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/util"
)

const AllCertsCacheKey = "all_certs_cache"

// WorkloadUnit 以 UniqueID 划分的 Workload 单元
type WorkloadUnit struct {
	Role          caclient.Role `json:"role"`
	ValidNum      int           `json:"valid_num"`       // 有效证书数量
	FirstIssuedAt time.Time     `json:"first_issued_at"` // 首次签发证书日期
	UniqueId      string        `json:"unique_id"`
	Forbidden     bool          `json:"forbidden"` // 是否被禁止
}

type WorkloadUnitsParams struct {
	Page, PageSize int
	UniqueId       string
}

// WorkloadUnits CA 下 Units
// 	返回目前活跃的 Units 及概要
func (l *Logic) WorkloadUnits(params *WorkloadUnitsParams) ([]*WorkloadUnit, int64, error) {
	db := l.db.Session(&gorm.Session{})
	// 默认筛选没有过期的
	db = db.Where("expiry > ?", time.Now()).
		Where("status", "good")
	db = db.Select(
		"ca_label",
		"common_name",
		"issued_at",
		"serial_number",
		"authority_key_identifier",
		"status",
		"not_before",
		"expiry",
		"revoked_at",
	)

	certs, err := getCerts(db)
	if err != nil {
		return make([]*WorkloadUnit, 0), 0, errors.Wrap(err, "数据库查询错误")
	}

	var i int
	var total int64
	units := make([]*WorkloadUnit, 0)
	fx.From(func(source chan<- interface{}) {
		for _, cert := range certs {
			source <- cert
		}
	}).Group(func(item interface{}) interface{} {
		cert := item.(*model.Certificates)
		return cert.CommonName
	}).Walk(func(item interface{}, pipe chan<- interface{}) {
		certs := item.([]interface{})
		firstCert := certs[0].(*model.Certificates)
		for _, certObj := range certs {
			cert := certObj.(*model.Certificates)
			if cert.IssuedAt.Before(firstCert.IssuedAt) {
				firstCert = cert
			}
		}
		unit := &WorkloadUnit{
			Role:          caclient.Role(firstCert.CaLabel.String),
			ValidNum:      len(certs),
			FirstIssuedAt: firstCert.IssuedAt,
			UniqueId:      firstCert.CommonName.String,
		}
		pipe <- unit
		atomic.AddInt64(&total, 1)
	}).Filter(func(item interface{}) bool {
		unit := item.(*WorkloadUnit)
		if params.UniqueId != "" {
			return unit.UniqueId == params.UniqueId
		}
		return true
	}).Sort(func(a, b interface{}) bool {
		aObj := a.(*WorkloadUnit)
		bObj := b.(*WorkloadUnit)
		return aObj.FirstIssuedAt.Before(bObj.FirstIssuedAt)
	}).Split(params.PageSize).ForEach(func(item interface{}) {
		i++
		if i == params.Page {
			group := item.([]interface{})
			for _, obj := range group {
				unit := obj.(*WorkloadUnit)
				units = append(units, unit)
			}
		}
	})

	return units, total, nil
}

func getCerts(db *gorm.DB) ([]*model.Certificates, error) {
	var certs []*model.Certificates
	var err error
	allCerts, ok := util.MapCache.Get(AllCertsCacheKey)
	if !ok {
		certs, _, err = dao.GetAllCertificates(db, 1, 10000, "issued_at desc")
		if err != nil {
			return nil, errors.Wrap(err, "数据库查询错误")
		}
		util.MapCache.SetDefault(AllCertsCacheKey, certs)
	}
	if allCerts != nil {
		certs = allCerts.([]*model.Certificates)
	}
	return certs, nil
}
