// Package workload 证书生命周期管理
package workload

import (
	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	logic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/workload"
)

// RevokeCerts 吊销证书
// @Tags Workload
// @Summary (p3)Revoke
// @Description 吊销证书
// @Produce json
// @Param body body logic.RevokeCertsParams true "sn+aki / unique_id 二选一"
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /workload/lifecycle/revoke [post]
func (a *API) RevokeCerts(c *helper.HTTPWrapContext) (interface{}, error) {
	var req logic.RevokeCertsParams
	c.BindG(&req)

	err := a.logic.RevokeCerts(&req)
	if err != nil {
		return nil, err
	}

	return "revoked", nil
}

// RecoverCerts 恢复证书
// @Tags Workload
// @Summary (p3)Recover
// @Description 恢复证书
// @Produce json
// @Param body body logic.RecoverCertsParams true "sn+aki / unique_id 二选一"
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /workload/lifecycle/recover [post]
func (a *API) RecoverCerts(c *helper.HTTPWrapContext) (interface{}, error) {
	var req logic.RecoverCertsParams
	c.BindG(&req)

	err := a.logic.RecoverCerts(&req)
	if err != nil {
		return nil, err
	}

	return "recovered", nil
}

// ForbidNewCerts 禁止某个 UniqueID 申请证书
// @Tags Workload
// @Summary 禁止申请证书
// @Description 禁止某个 UniqueID 申请证书
// @Produce json
// @Param body body logic.ForbidNewCertsParams true " "
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /workload/lifecycle/forbid_new_certs [post]
func (a *API) ForbidNewCerts(c *helper.HTTPWrapContext) (interface{}, error) {
	var req logic.ForbidNewCertsParams
	c.BindG(&req)

	err := a.logic.ForbidNewCerts(&req)
	if err != nil {
		return nil, err
	}

	return "success", nil
}

// RecoverForbidNewCerts 恢复允许某个 UniqueID 申请证书
// @Tags Workload
// @Summary 恢复申请证书
// @Description 恢复允许某个 UniqueID 申请证书
// @Produce json
// @Param body body logic.ForbidNewCertsParams true " "
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /workload/lifecycle/recover_forbid_new_certs [post]
func (a *API) RecoverForbidNewCerts(c *helper.HTTPWrapContext) (interface{}, error) {
	var req logic.ForbidNewCertsParams
	c.BindG(&req)

	err := a.logic.RecoverForbidNewCerts(&req)
	if err != nil {
		return nil, err
	}

	return "success", nil
}

type ForbidUnitParams struct {
	UniqueID string `json:"unique_id" binding:"required"`
}

// ForbidUnit 吊销并禁止服务证书
// @Tags Workload
// @Summary (p1)吊销并禁止服务证书
// @Description 吊销并禁止服务证书
// @Produce json
// @Param json body ForbidUnitParams true " "
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /workload/lifecycle/forbid_unit [post]
func (a *API) ForbidUnit(c *helper.HTTPWrapContext) (interface{}, error) {
	var req ForbidUnitParams
	c.BindG(&req)

	err := a.logic.ForbidNewCerts(&logic.ForbidNewCertsParams{
		UniqueIds: []string{req.UniqueID},
	})
	if err != nil {
		a.logger.With("req", req).Errorf("禁止申请证书失败: %s", err)
		return nil, err
	}

	// 2021.04.15 (功能性调整) 证书启用/禁用影响证书通信 OCSP 认证、Sidecar mTLS 使用，不会吊销证书
	// err = a.logic.RevokeCerts(&logic.RevokeCertsParams{
	//    UniqueId: req.UniqueID,
	// })
	// if err != nil {
	//    a.logger.With("req", req).Errorf("吊销服务证书失败: %s", err)
	//    return nil, err
	// }

	return "success", nil
}

// RecoverUnit 恢复并允许服务证书
// @Tags Workload
// @Summary (p1)恢复并允许服务证书
// @Description 恢复并允许服务证书
// @Produce json
// @Param json body ForbidUnitParams true " "
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /workload/lifecycle/recover_unit [post]
func (a *API) RecoverUnit(c *helper.HTTPWrapContext) (interface{}, error) {
	var req ForbidUnitParams
	c.BindG(&req)

	err := a.logic.RecoverForbidNewCerts(&logic.ForbidNewCertsParams{
		UniqueIds: []string{req.UniqueID},
	})
	if err != nil {
		a.logger.With("req", req).Errorf("恢复申请证书失败: %s", err)
		return nil, err
	}

	// err = a.logic.RecoverCerts(&logic.RecoverCertsParams{
	//    UniqueId: req.UniqueID,
	// })
	// if err != nil {
	//    a.logger.With("req", req).Errorf("恢复服务证书失败: %s", err)
	//    return nil, err
	// }

	return "success", nil
}
