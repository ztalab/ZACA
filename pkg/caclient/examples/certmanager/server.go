package main

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/keygen"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap/zapcore"

	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
)

var (
	caAddr   = pflag.String("ca", "https://127.0.0.1:8081", "CA Server")
	ocspAddr = pflag.String("ocsp", "http://127.0.0.1:8082", "Ocsp Server")
	authKey  = pflag.String("auth-key", "ea62fa7c27307017694689f0adff09f63186cadfe92fb802133f980b75858fc6", "Auth Key")
)

func init() {
	_ = logger.GlobalConfig(logger.Conf{
		Debug: true,
		Level: zapcore.DebugLevel,
	})
}

func main() {
	pflag.Parse()
	err := NewIDGRegistry()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// NewIDGRegistry 注册中心测试示例
func NewIDGRegistry() error {
	cai := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleIDGRegistry, *caAddr),
		caclient.WithAuthKey(*authKey),
		caclient.WithOcspAddr(*ocspAddr),
	)
	cm, err := cai.NewCertManager()
	if err != nil {
		logger.Errorf("cert manager 创建错误: %s", err)
		return err
	}

	_, keyPEM, _ := keygen.GenKey(keygen.RsaSigAlg)
	logger.Info("生成 RSA 私钥")

	csrBytes, err := keygen.GenCustomExtendCSR(keyPEM, &spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "test_cluster",
		UniqueID:  "idg_registy_0001",
	}, &keygen.CertOptions{
		CN: "test",
	}, []pkix.Extension{
		{
			Id:       asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7, 8, 1},
			Critical: true,
			Value:    []byte("fake data"),
		},
		{
			Id:       asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7, 8, 2},
			Critical: true,
			Value:    []byte("fake data"),
		},
	})
	if err != nil {
		return err
	}
	logger.Infof("生成自定义 CSR: \n%s", string(csrBytes))

	// 申请证书
	certBytes, err := cm.SignPEM(csrBytes, "test")
	if err != nil {
		logger.Errorf("申请证书失败: %s", err)
		return err
	}

	logger.Infof("从 CA 申请证书: \n%s", string(certBytes))

	// 验证证书
	if err := cm.VerifyCertDefaultIssuer(certBytes); err != nil {
		logger.Errorf("验证证书失败: %s", err)
		return err
	}
	logger.Infof("验证证书成功, 证书有效")

	// 吊销证书
	if err := cm.RevokeIDGRegistryCert(certBytes); err != nil {
		logger.Errorf("吊销证书失败: %s", err)
		return err
	}
	logger.Infof("吊销证书成功")

	return nil
}
