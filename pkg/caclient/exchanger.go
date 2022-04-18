package caclient

import (
	"github.com/cloudflare/backoff"
	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/keyprovider"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	"gitlab.oneitfarm.com/bifrost/cfssl/transport"
	"gitlab.oneitfarm.com/bifrost/cfssl/transport/roots"
	"go.uber.org/zap"
	"net/url"
	"reflect"
)

const (
	// CertRefreshDurationRate 证书轮回时间率
	CertRefreshDurationRate int = 2
)

// Exchanger ...
type Exchanger struct {
	Transport   *Transport
	IDGIdentity *spiffe.IDGIdentity
	OcspFetcher OcspClient

	caAddr string
	logger *zap.SugaredLogger

	caiConf *Conf
}

func init() {
	// CFSSL API Client 连接 API Server 不进行证书验证 (单向 TLS)
	hook.ClientInsecureSkipVerify = true
}

// NewExchangerWithKeypair ...
func (cai *CAInstance) NewExchangerWithKeypair(id *spiffe.IDGIdentity, keyPEM []byte, certPEM []byte) (*Exchanger, error) {
	tr, err := cai.NewTransport(id, keyPEM, certPEM)
	if err != nil {
		return nil, err
	}
	of, err := NewOcspMemCache(cai.Logger.Sugar().Named("ocsp"), cai.Conf.OcspAddr)
	if err != nil {
		return nil, err
	}
	return &Exchanger{
		Transport:   tr,
		IDGIdentity: id,
		OcspFetcher: of,
		logger:      cai.Logger.Sugar().Named("ca"),
		caAddr:      cai.CaAddr,

		caiConf: &cai.Conf,
	}, nil
}

// NewExchanger ...
func (cai *CAInstance) NewExchanger(id *spiffe.IDGIdentity) (*Exchanger, error) {
	tr, err := cai.NewTransport(id, nil, nil)
	if err != nil {
		return nil, err
	}
	of, err := NewOcspMemCache(cai.Logger.Sugar().Named("ocsp"), cai.Conf.OcspAddr)
	if err != nil {
		return nil, err
	}
	return &Exchanger{
		Transport:   tr,
		IDGIdentity: id,
		OcspFetcher: of,
		logger:      cai.Logger.Sugar().Named("ca"),
		caAddr:      cai.CaAddr,

		caiConf: &cai.Conf,
	}, nil
}

// NewTransport ...
func (cai *CAInstance) NewTransport(id *spiffe.IDGIdentity, keyPEM []byte, certPEM []byte) (*Transport, error) {
	l := cai.Logger.Sugar()

	l.Debug("NewTransport 开始")

	if _, err := url.Parse(cai.CaAddr); err != nil {
		return nil, errors.Wrap(err, "CA ADDR 错误")
	}

	var tr = &Transport{
		CertRefreshDurationRate: CertRefreshDurationRate,
		Identity:                cai.CFIdentity,
		Backoff:                 &backoff.Backoff{},
		logger:                  l.Named("ca"),
	}

	l.Debugf("[NEW]: 证书轮换率: %v", tr.CertRefreshDurationRate)

	l.Debug("roots 初始化")
	store, err := roots.New(cai.CFIdentity.Roots)
	if err != nil {
		return nil, err
	}
	tr.TrustStore = store

	l.Debug("client roots 初始化")
	if len(cai.CFIdentity.ClientRoots) > 0 {
		// 如果 cai.CFIdentity.Roots 与cai.CFIdentity.ClientRoots 相同，则不重复请求
		if !reflect.DeepEqual(cai.CFIdentity.Roots, cai.CFIdentity.ClientRoots) {
			store, err = roots.New(cai.CFIdentity.ClientRoots)
			if err != nil {
				return nil, err
			}
		}

		tr.ClientTrustStore = store
	}

	l.Debug("xkeyProvider 初始化")
	xkey, err := keyprovider.NewXKeyProvider(id)
	if err != nil {
		return nil, err
	}

	xkey.CSRConf = cai.CSRConf
	if keyPEM != nil && certPEM != nil {
		l.Debug("xkeyProvider 设置 keyPEM")
		if err := xkey.SetPrivateKeyPEM(keyPEM); err != nil {
			return nil, err
		}
		l.Debug("xkeyProvider 设置 certPEM")
		if err := xkey.SetCertificatePEM(certPEM); err != nil {
			return nil, err
		}
	}
	tr.Provider = xkey

	l.Debug("CA 初始化")
	tr.CA, err = transport.NewCA(cai.CFIdentity)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// RotateController ...
func (ex *Exchanger) RotateController() *RotateController {
	return &RotateController{
		transport:   ex.Transport,
		rotateAfter: ex.caiConf.RotateAfter,
		logger:      ex.logger.Named("rotator"),
	}
}
