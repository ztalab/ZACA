package singleca

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

func getTrustCerts(path string) ([]*x509.Certificate, error) {
	pemCerts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("信任证书文件错误: %v", err)
	}
	certs, err := helpers.ParseCertificatesPEM(pemCerts)
	if err != nil {
		return nil, fmt.Errorf("获取信任证书失败: %v", err)
	}
	logger.Named("trust-certs").Infof("获取到信任证书数量: %v", len(certs))
	return certs, nil
}
