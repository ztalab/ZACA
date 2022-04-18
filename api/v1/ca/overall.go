package ca

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
)

type OverallCertsCountItem struct {
	Role       string `json:"role"`        // 类别
	Total      int64  `json:"total"`       // 证书总数
	UnitsCount int64  `json:"units_count"` // 服务数量
}

type OverallCertsCountResponse struct {
	Total int64                   `json:"total"`
	Certs []OverallCertsCountItem `json:"certs"`
}

// OverallCertsCount 证书分类
// @Tags CA
// @Summary (p2)证书分类
// @Description 证书总数、根据分类划分的数量、对应服务数量
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=OverallCertsCountResponse} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/overall_certs_count [get]
func (a *API) OverallCertsCount(c *helper.HTTPWrapContext) (interface{}, error) {
	query := func() *gorm.DB {
		return core.Is.Db.Session(&gorm.Session{}).Model(&model.Certificates{}).
			Where("expiry > ?", time.Now()).
			Where("reason IN ?", []int{0, 2})
	}

	var total int64
	if err := query().Count(&total).Error; err != nil {
		a.logger.Errorf("mysql query err: %s", err)
		return nil, err
	}

	roleProfiles, err := a.logic.RoleProfiles()
	if err != nil {
		a.logger.Errorf("获取 role profiles 错误: %s", err)
		return nil, errors.New("获取 role profiles 错误")
	}

	res := &OverallCertsCountResponse{
		Total: total,
		Certs: make([]OverallCertsCountItem, 0),
	}

	for _, roleProfile := range roleProfiles {
		role := roleProfile.Name
		item := OverallCertsCountItem{Role: role}
		if err := query().Where("ca_label = ?", strings.ToLower(role)).Count(&item.Total).Error; err != nil {
			a.logger.Errorf("mysql query err: %s", err)
			return nil, err
		}

		if err := query().Where("ca_label = ?", strings.ToLower(role)).
			Where(`common_name != ""`).Group("common_name").Count(&item.UnitsCount).Error; err != nil {
			a.logger.Errorf("mysql query err: %s", err)
			return nil, err
		}
		res.Certs = append(res.Certs, item)
	}

	return res, nil
}

type OverallExpiryGroup struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type OverallExpiryCertsResponse struct {
	ExpiryTotal int64                `json:"expiry_total"`
	ExpiryCerts []OverallExpiryGroup `json:"expiry_certs"`
}

// OverallExpiryCerts 证书有效期
// @Tags CA
// @Summary (p2)证书有效期
// @Description 证书已过期数量, 一周内过期数量, 1/3个月内过期数量, 3个月后过期数量
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=OverallExpiryCertsResponse} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/overall_expiry_certs [get]
func (a *API) OverallExpiryCerts(c *helper.HTTPWrapContext) (interface{}, error) {
	query := func() *gorm.DB {
		return core.Is.Db.Model(&model.Certificates{}).
			Where("expiry < ?", time.Now()).
			Where("revoked_at IS NULL")
	}

	var total int64
	if err := query().Count(&total).Error; err != nil {
		a.logger.Errorf("mysql query err: %s", err)
		return nil, err
	}

	res := &OverallExpiryCertsResponse{
		ExpiryTotal: total,
		ExpiryCerts: make([]OverallExpiryGroup, 0),
	}

	// 一周内
	{
		item := OverallExpiryGroup{Name: "1w"}
		count, err := getExpiryCountByDuration(7*24*time.Hour, time.Now())
		if err != nil {
			return nil, err
		}
		item.Count = count
		res.ExpiryCerts = append(res.ExpiryCerts, item)
	}

	// 一个月内
	{
		item := OverallExpiryGroup{Name: "1m"}
		count, err := getExpiryCountByDuration(30*24*time.Hour, time.Now())
		if err != nil {
			return nil, err
		}
		item.Count = count
		res.ExpiryCerts = append(res.ExpiryCerts, item)
	}

	// 三个月内
	{
		item := OverallExpiryGroup{Name: "3m"}
		count, err := getExpiryCountByDuration(3*30*24*time.Hour, time.Now())
		if err != nil {
			return nil, err
		}
		item.Count = count
		res.ExpiryCerts = append(res.ExpiryCerts, item)
	}

	// 三个月后
	{
		item := OverallExpiryGroup{Name: "3m+"}
		count, err := getExpiryCountByDuration(999*30*24*time.Hour, time.Now().AddDate(0, 3, 0))
		if err != nil {
			return nil, err
		}
		item.Count = count
		res.ExpiryCerts = append(res.ExpiryCerts, item)
	}

	return res, nil
}

func getExpiryCountByDuration(period time.Duration, before time.Time) (int64, error) {
	// 一周内
	// 过期时间 - 当前时间 <= 一周
	// 过期时间 <= 当前时间 + 一周
	expiryDate := time.Now().Add(period)
	query := core.Is.Db.Session(&gorm.Session{}).Model(&model.Certificates{}).
		Where("expiry > ?", before).
		Where("expiry < ?", expiryDate).
		Where("reason = 0").
		Where(`common_name != ""`)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		v2log.Errorf("mysql query err: %s", err)
		return 0, err
	}

	return count, nil
}

type OverallUnitsEnableItem struct {
	CertsCount int64 `json:"certs_count"`
	UnitsCount int64 `json:"units_count"`
}

type OverallUnitsEnableStatus struct {
	Enable  OverallUnitsEnableItem `json:"enable"`
	Disable OverallUnitsEnableItem `json:"disable"`
}

// OverallUnitsEnableStatus 启用情况
// @Tags CA
// @Summary (p2)启用情况
// @Description 已启用总数, 禁用总数, 对应服务数
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=OverallUnitsEnableStatus} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/overall_units_enable_status [get]
func (a *API) OverallUnitsEnableStatus(c *helper.HTTPWrapContext) (interface{}, error) {
	query := func() *gorm.DB {
		return core.Is.Db.Session(&gorm.Session{}).Model(&model.Certificates{}).
			Where("expiry > ?", time.Now())
	}

	res := &OverallUnitsEnableStatus{}

	{
		if err := query().Where("reason = ?", 0).Count(&res.Enable.CertsCount).Error; err != nil {
			a.logger.Errorf("mysql query err: %s", err)
			return nil, err
		}
		if err := query().Where("reason = ?", 2).Count(&res.Disable.CertsCount).Error; err != nil {
			a.logger.Errorf("mysql query err: %s", err)
			return nil, err
		}
	}

	{
		if err := query().Where("reason = ?", 0).Group("common_name").Count(&res.Enable.UnitsCount).Error; err != nil {
			a.logger.Errorf("mysql query err: %s", err)
			return nil, err
		}
		if err := query().Where("reason = ?", 2).Group("common_name").Count(&res.Disable.UnitsCount).Error; err != nil {
			a.logger.Errorf("mysql query err: %s", err)
			return nil, err
		}
	}

	return res, nil
}
