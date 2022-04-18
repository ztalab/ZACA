package caclient

import (
	"crypto/x509"
	"encoding/hex"
	"math"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ocsp"

	"gitlab.oneitfarm.com/bifrost/go-toolbox/memorycacher"
	"go.uber.org/zap"
)

var _ OcspClient = &ocspMemCache{}

// ocspMemCache ...
type ocspMemCache struct {
	cache   *memorycacher.Cache
	logger  *zap.SugaredLogger
	ocspURL string // ca server + /ocsp
}

// NewOcspMemCache ...
func NewOcspMemCache(logger *zap.SugaredLogger, ocspAddr string) (OcspClient, error) {
	return &ocspMemCache{
		cache:   memorycacher.New(30*time.Minute, memorycacher.NoExpiration, math.MaxInt64),
		logger:  logger,
		ocspURL: ocspAddr,
	}, nil
}

// Validate ...
func (of *ocspMemCache) Validate(leaf, issuer *x509.Certificate) (bool, error) {
	if atomic.LoadInt64(&ocspBlockSign) == 1 {
		return false, errors.New("ocsp 请求被禁用")
	}
	if leaf == nil || issuer == nil {
		return false, errors.New("leaf/issuer 参数缺失")
	}
	lo := of.logger.With("sn", leaf.SerialNumber.String(), "aki", hex.EncodeToString(leaf.AuthorityKeyId), "id", leaf.URIs[0])
	// 缓存获取
	if _, ok := of.cache.Get(leaf.SerialNumber.String()); ok {
		return true, nil
	}
	ocspRequest, err := ocsp.CreateRequest(leaf, issuer, &ocspOpts)
	if err != nil {
		lo.Errorf("ocsp req create err: %s", err)
		return false, errors.Wrap(err, "ocsp req 创建失败")
	}
	getOcspFunc := func() (interface{}, error) {
		return SendOcspRequest(of.ocspURL, ocspRequest, leaf, issuer)
	}
	sgValue, err, _ := sg.Do("ocsp"+leaf.SerialNumber.String(), getOcspFunc)
	if err != nil {
		lo.Errorf("ocsp 请求错误: %v", err)
		// 这里因为 Ca Server 原因导致验证失败, 允许请求, 下一次再重试
		return true, errors.Wrap(err, "ocsp 请求错误")
	}
	ocspResp, ok := sgValue.(*ocsp.Response)
	if !ok {
		lo.Error("single flight 解析错误")
		return false, errors.New("single flight 解析错误")
	}
	lo.Debugf("验证 OCSP, 结果: %v", ocspResp.Status)
	if ocspResp.Status == int(ocsp.Success) {
		of.cache.SetDefault(leaf.SerialNumber.String(), true)
		return true, nil
	}
	lo.Warnf("证书 OCSP 验证失效")
	return false, errors.New("ocsp 验证失败, 证书被吊销")
}

func (of *ocspMemCache) Reset() {
	of.cache.Flush()
}
