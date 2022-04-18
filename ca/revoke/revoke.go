// Package revoke implements the HTTP handler for the revoke command
package revoke

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"gitlab.oneitfarm.com/bifrost/cfssl/api"
	"gitlab.oneitfarm.com/bifrost/cfssl/certdb"
	cf_err "gitlab.oneitfarm.com/bifrost/cfssl/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	"gitlab.oneitfarm.com/bifrost/cfssl/ocsp"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/events"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/signature"
	"gitlab.oneitfarm.com/bifrost/capitalizone/util"
)

// A Handler accepts requests with a serial number parameter
// and revokes
type Handler struct {
	dbAccessor certdb.Accessor
	logger     *v2log.Logger
}

// NewHandler returns a new http.Handler that handles a revoke request.
func NewHandler(dbAccessor certdb.Accessor) http.Handler {
	return &api.HTTPHandler{
		Handler: &Handler{
			dbAccessor: dbAccessor,
			logger:     v2log.Named("revoke"),
		},
		Methods: []string{"POST"},
	}
}

// This type is meant to be unmarshalled from JSON
type JsonRevokeRequest struct {
	Serial  string `json:"serial"`
	AKI     string `json:"authority_key_id"`
	Reason  string `json:"reason"`
	Nonce   string `json:"nonce"`
	Sign    string `json:"sign"`
	AuthKey string `json:"auth_key"`
	Profile string `json:"profile"`
}

// Handle responds to revocation requests. It attempts to revoke
// a certificate with a given serial number
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()

	// Default the status to good so it matches the cli
	var req JsonRevokeRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		return cf_err.NewBadRequestString("Unable to parse revocation request")
	}

	if len(req.Serial) == 0 {
		return cf_err.NewBadRequestString("serial number is required but not provided")
	}

	certRecord := &model.Certificates{}
	if err := core.Is.Db.Where("serial_number = ? AND authority_key_identifier = ?", req.Serial, req.AKI).First(certRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			h.logger.With("sn", req.Serial, "aki", req.AKI).Warn("证书不存在")
		} else {
			h.logger.With("sn", req.Serial, "aki", req.AKI).Errorf("证书获取错误: %v", err)
		}
		return cf_err.NewBadRequest(err)
	}

	// 从 vault 获取证书 PEM
	if hook.EnableVaultStorage {
		pem, err := core.Is.VaultSecret.GetCertPEM(req.Serial)
		if err != nil {
			h.logger.With("sn", req.Serial, "aki", req.AKI).Warnf("Vault 获取错误: %v", err)
		} else {
			certRecord.Pem = *pem
		}
	}

	cert, err := helpers.ParseCertificatePEM([]byte(certRecord.Pem))
	if err != nil {
		h.logger.With("sn", req.Serial, "aki", req.AKI).Errorf("证书 PEM 解析错误: %v", err)
		return cf_err.NewBadRequest(err)
	}

	// TODO 兼容标准 CFSSL 认证方式
	var valid bool
	if req.AuthKey == "" {
		v := signature.NewVerifier(cert.PublicKey)
		valid, err = v.Verify([]byte(req.Nonce), req.Sign)
		if err != nil {
			h.logger.With("sn", req.Serial, "aki", req.AKI).Warnf("验证错误: %v", err)
			return cf_err.NewBadRequest(err)
		}
	} else {
		if req.Profile == "" {
			return cf_err.NewBadRequest(errors.New("profile 未指定"))
		}
		if req.Profile != string(caclient.RoleIDGRegistry) {
			return cf_err.NewBadRequest(errors.New("profile 不被允许进行吊销操作"))
		}
		if authKey, ok := core.Is.Config.Singleca.CfsslConfig.AuthKeys[req.Profile]; ok {
			if authKey.Key == req.AuthKey {
				valid = true
			}
		}
	}

	if !valid {
		h.logger.With("sn", req.Serial, "aki", req.AKI).Warnf("证书无法对应: %v", err)
		return cf_err.NewBadRequest(err)
	}

	var reasonCode int
	reasonCode, err = ocsp.ReasonStringToCode("keycompromise")
	if err != nil {
		return cf_err.NewBadRequestString("Invalid reason code")
	}

	// 删除 vault 对应的证书 KV
	if hook.EnableVaultStorage {
		if err := core.Is.VaultSecret.DeleteCertPEM(req.Serial); err != nil {
			h.logger.With("sn", req.Serial, "aki", req.AKI).Warnf("Vault 删除错误: %v", err)
		}
	}

	err = h.dbAccessor.RevokeCertificate(req.Serial, req.AKI, reasonCode)
	if err != nil {
		h.logger.With("sn", req.Serial, "aki", req.AKI).Warnf("数据库操作错误: %v", err)
		return err
	}

	AddMetricsPoint(cert)

	events.NewWorkloadLifeCycle("self-revoke", events.OperatorSDK, events.CertOp{
		UniqueId: cert.Subject.CommonName,
		SN:       req.Serial,
		AKI:      req.AKI,
	}).Log()

	h.logger.With("sn", req.Serial, "aki", req.AKI, "uri", util.GetSanURI(cert)).Info("Workload 主动吊销证书")

	result := map[string]string{}
	return api.SendResponse(w, result)
}
