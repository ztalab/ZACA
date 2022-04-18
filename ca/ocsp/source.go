package ocsp

import (
	"encoding/hex"
	"math"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	"gitlab.oneitfarm.com/bifrost/cfssl/ocsp"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gitlab.oneitfarm.com/bifrost/go-toolbox/memorycacher"
	"go.uber.org/zap"
	stdocsp "golang.org/x/crypto/ocsp"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/events"
)

const (
	CertStatusGood           = "good"
	CertStatusUnknown        = "unknown"
	CertStatusNotFound       = "notfound"
	CertStatusServerError    = "servererror"
	CertStatusCertParseError = "certparseerror"
	CertStatusOCSPSignError  = "ocspsignerror"
)

var CertStatusIntMap = map[string]int{
	CertStatusGood:           200,
	CertStatusUnknown:        599,
	CertStatusNotFound:       404,
	CertStatusServerError:    500,
	CertStatusCertParseError: 400,
	CertStatusOCSPSignError:  502,
}

// SharedSources 进程 Cache 保障高效访问, 后续若访问量大可以考虑 Redis
type SharedSources struct {
	DB         *gorm.DB
	Cache      *memorycacher.Cache
	Logger     *zap.SugaredLogger
	OcspSigner ocsp.Signer
}

// NewSharedSources ...
func NewSharedSources(signer ocsp.Signer) (*SharedSources, error) {
	if core.Is.Db == nil {
		return nil, errors.New("database instance not found")
	}
	cacheTime := time.Duration(core.Is.Config.Ocsp.CacheTime)
	return &SharedSources{
		DB:         core.Is.Db,
		Logger:     v2log.Named("ocsp-ss").SugaredLogger,
		Cache:      memorycacher.New(cacheTime*time.Minute, memorycacher.NoExpiration, math.MaxInt64),
		OcspSigner: signer,
	}, nil
}

// Response 查询 DB 返回 OCSP 数据结构
func (ss *SharedSources) Response(req *stdocsp.Request) ([]byte, http.Header, error) {
	if req == nil {
		return nil, nil, errors.New("called with nil request")
	}

	aki := hex.EncodeToString(req.IssuerKeyHash)
	sn := req.SerialNumber

	if sn == nil {
		return nil, nil, errors.New("request contains no serial")
	}
	strSN := sn.String()

	if cachedResp, ok := ss.Cache.Get(strSN + aki); ok {
		if resp, ok := cachedResp.([]byte); ok {
			ss.Logger.With("sn", strSN, "aki", aki).Debugf("ocspResp cache 击中")
			// TODO 获取 UniqueID
			AddMetricsPoint("", true, CertStatusUnknown)
			return resp, nil, nil
		}
		ss.Logger.With("sn", strSN, "aki", aki).Errorf("cache 值解析错误")
	}

	// 数据库查询
	certRecord := &model.Certificates{}
	if err := ss.DB.Where("serial_number = ? AND authority_key_identifier = ?", strSN, aki).First(certRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ss.Logger.With("sn", strSN, "aki", aki).Warnw("证书不存在")
			AddMetricsPoint("", false, CertStatusNotFound)
			return nil, nil, ocsp.ErrNotFound
		}
		ss.Logger.With("sn", strSN, "aki", aki).Errorf("证书获取错误: %v", err)
		AddMetricsPoint("", false, CertStatusServerError)
		return nil, nil, errors.Wrap(err, "server error")
	}

	// 从 vault 获取证书 PEM
	if hook.EnableVaultStorage {
		pem, err := core.Is.VaultSecret.GetCertPEM(strSN)
		if err != nil {
			ss.Logger.With("sn", strSN, "aki", aki).Warnf("Vault 获取错误: %v", err)
		} else {
			certRecord.Pem = *pem
		}
	}

	cert, err := helpers.ParseCertificatePEM([]byte(certRecord.Pem))
	if err != nil {
		ss.Logger.With("sn", strSN, "aki", aki).Errorf("证书 PEM 解析错误: %v", err)
		AddMetricsPoint("", false, CertStatusCertParseError)
		return nil, nil, errors.Wrap(err, "cert err")
	}

	signReq := &ocsp.SignRequest{
		Certificate: cert,
		Status:      certRecord.Status,
		Reason:      int(certRecord.Reason.Int64),
		RevokedAt:   certRecord.RevokedAt,
	}

	ocspResp, err := ss.OcspSigner.Sign(*signReq)
	if err != nil {
		ss.Logger.With("sn", strSN, "aki", aki).Errorf("OCSP Sign 错误: %v", err)
		AddMetricsPoint(cert.Subject.CommonName, false, CertStatusOCSPSignError)
		return nil, nil, errors.Wrap(err, "internal err")
	}

	events.NewWorkloadLifeCycle("oscp-sign", events.OperatorSDK, events.CertOp{
		UniqueId: cert.Subject.CommonName,
		SN:       strSN,
		AKI:      aki,
	}).Log()

	ss.Cache.SetDefault(strSN+aki, ocspResp)

	ss.Logger.With("sn", strSN, "aki", aki).Infof("OCSP 签名完成")

	AddMetricsPoint(cert.Subject.CommonName, false, CertStatusGood)
	return ocspResp, nil, nil
}
