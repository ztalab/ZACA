package keymanager

import (
	"gitlab.oneitfarm.com/bifrost/cfssl/initca"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

// SelfSigner ...
type SelfSigner struct {
	logger *v2log.Logger
}

// NewSelfSigner ...
func NewSelfSigner() *SelfSigner {
	return &SelfSigner{
		logger: v2log.Named("self-signer"),
	}
}

// Run 自签名证书并储存
func (ss *SelfSigner) Run() error {
	key, cert, _ := GetKeeper().GetCachedSelfKeyPairPEM()
	if key != nil && cert != nil {
		ss.logger.Info("证书已存在, 跳过自签名过程")
		return nil
	}
	ss.logger.Warn("没有证书, 即将自签名证书")
	cert, _, key, err := initca.New(getRootCSRTemplate())
	if err != nil {
		ss.logger.Errorf("initca 创建错误: %v", err)
		return err
	}
	ss.logger.With("key", string(key), "cert", string(cert)).Debugf("自签证书完成")
	if err = GetKeeper().SetKeyPairPEM(key, cert); err != nil {
		ss.logger.Errorf("储存证书错误: %v", err)
		return err
	}

	// TODO 开启协程自动轮换证书

	return nil
}
