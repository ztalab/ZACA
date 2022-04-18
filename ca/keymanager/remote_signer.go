package keymanager

import (
	jsoniter "github.com/json-iterator/go"
	cfssl_client "gitlab.oneitfarm.com/bifrost/cfssl/api/client"
	"gitlab.oneitfarm.com/bifrost/cfssl/cli/genkey"
	"gitlab.oneitfarm.com/bifrost/cfssl/csr"
	"gitlab.oneitfarm.com/bifrost/cfssl/signer"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

// RemoteSigner ...
type RemoteSigner struct {
	logger *v2log.Logger
}

// NewRemoteSigner ...
func NewRemoteSigner() *RemoteSigner {
	return &RemoteSigner{
		logger: v2log.Named("remote-signer"),
	}
}

// Run 调用远程 CA 签名证书并持久化储存
func (ss *RemoteSigner) Run() error {
	if core.Is.Config.Keymanager.SelfSign {
		return nil
	}
	key, cert, _ := GetKeeper().GetCachedSelfKeyPairPEM()
	if key != nil && cert != nil {
		ss.logger.Info("证书已存在, 跳过远程签名过程")
		return nil
	}
	ss.logger.Warn("没有证书, 即将远程签名证书")
	g := &csr.Generator{Validator: genkey.Validator}
	csrBytes, key, err := g.ProcessRequest(getIntermediateCSRTemplate())
	if err != nil {
		ss.logger.Errorf("key, csr 生产错误: %v", err)
		return err
	}

	signReq := signer.SignRequest{
		Request: string(csrBytes),
		Profile: "intermediate",
	}

	signReqBytes, _ := jsoniter.Marshal(&signReq)
	err = GetKeeper().RootClient.DoWithRetry(func(remote *cfssl_client.AuthRemote) error {
		certResp, err := remote.Sign(signReqBytes)
		if err != nil {
			return err
		}
		cert = certResp
		return nil
	})
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
