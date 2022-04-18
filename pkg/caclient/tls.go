package caclient

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/pkg/errors"

	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/cfssl/transport/core"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

// TLSGenerator ...
type TLSGenerator struct {
	Cfg *tls.Config
}

// NewTLSGenerator ...
func NewTLSGenerator(cfg *tls.Config) *TLSGenerator {
	return &TLSGenerator{Cfg: cfg}
}

// ExtraValidator 自定义验证函数, 在验证证书成功后执行
type ExtraValidator func(identity *spiffe.IDGIdentity) error

// BindExtraValidator 注册自定义验证函数
func (tg *TLSGenerator) BindExtraValidator(validator ExtraValidator) {
	vc := func(state tls.ConnectionState) error {
		// 若没有证书, 会在上一阶段被阻断
		if len(state.PeerCertificates) == 0 {
			return nil
		}
		cert := state.PeerCertificates[0]
		var id *spiffe.IDGIdentity
		if len(cert.URIs) > 0 {
			id, _ = spiffe.ParseIDGIdentity(cert.URIs[0].String())
		}
		return validator(id)
	}
	getServerTls := tg.Cfg.GetConfigForClient
	if getServerTls != nil {
		// 服务端动态获取
		tg.Cfg.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			tlsCfg, err := getServerTls(info)
			if err != nil {
				return nil, err
			}
			tlsCfg.VerifyConnection = vc
			return tlsCfg, nil
		}
	} else {
		tg.Cfg.VerifyConnection = vc
	}
}

// TLSConfig 获取 Golang 原生 TLS Config
func (tg *TLSGenerator) TLSConfig() *tls.Config {
	return tg.Cfg
}

// ClientTLSConfig ...
func (ex *Exchanger) ClientTLSConfig(host string) (*TLSGenerator, error) {
	lo := ex.logger
	lo.Debug("client tls started.")
	if _, err := ex.Transport.GetCertificate(); err != nil {
		return nil, errors.Wrap(err, "客户端证书获取错误")
	}
	c, err := ex.Transport.TLSClientAuthClientConfig(host)
	if err != nil {
		return nil, err
	}
	c.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) > 0 && len(verifiedChains) > 0 {
			leaf, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				lo.Errorf("leaf 证书解析错误: %v", err)
				return err
			}
			if ok, err := ex.OcspFetcher.Validate(leaf, verifiedChains[0][1]); !ok {
				return err
			}
		}
		return nil
	}
	return NewTLSGenerator(c), nil
}

// ServerHTTPSConfig ...
func (ex *Exchanger) ServerHTTPSConfig() (*TLSGenerator, error) {
	lo := ex.logger
	lo.Debug("server tls started.")
	if _, err := ex.Transport.GetCertificate(); err != nil {
		return nil, errors.Wrap(err, "服务器证书获取错误")
	}
	c, err := ex.Transport.TLSClientAuthServerConfig()
	if err != nil {
		return nil, err
	}
	c.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		tlsConfig := &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := ex.Transport.GetCertificate()
				if err != nil {
					logger.Named("transport").Errorf("服务器证书获取错误: %v", err)
					return nil, err
				}
				return cert, nil
			},
			ClientAuth:   tls.NoClientCert,
			CipherSuites: core.CipherSuites,
			MinVersion:   tls.VersionTLS12,
		}
		return tlsConfig, nil
	}
	return NewTLSGenerator(c), nil
}

// ServerTLSConfig ...
func (ex *Exchanger) ServerTLSConfig() (*TLSGenerator, error) {
	lo := ex.logger
	lo.Debug("server tls started.")
	if _, err := ex.Transport.GetCertificate(); err != nil {
		return nil, errors.Wrap(err, "服务器证书获取错误")
	}
	c, err := ex.Transport.TLSClientAuthServerConfig()
	if err != nil {
		return nil, err
	}
	c.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		tlsConfig := &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := ex.Transport.GetCertificate()
				if err != nil {
					logger.Named("transport").Errorf("服务器证书获取错误: %v", err)
					return nil, err
				}
				return cert, nil
			},
			RootCAs:   ex.Transport.TrustStore.Pool(),
			ClientCAs: ex.Transport.ClientTrustStore.Pool(),
			VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				if len(rawCerts) > 0 && len(verifiedChains) > 0 {
					leaf, err := x509.ParseCertificate(rawCerts[0])
					if err != nil {
						lo.Errorf("leaf 证书解析错误: %v", err)
						return err
					}
					if ok, err := ex.OcspFetcher.Validate(leaf, verifiedChains[0][1]); !ok {
						return err
					}
				}
				return nil
			},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			CipherSuites: core.CipherSuites,
			MinVersion:   tls.VersionTLS12,
		}
		return tlsConfig, nil
	}
	return NewTLSGenerator(c), nil
}
