package util

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

func ExtractCertFromExchanger(ex *caclient.Exchanger) {
	logger := v2log.Named("keypair-exporter")
	tlsCert, err := ex.Transport.GetCertificate()
	if err != nil {
		logger.Errorf("TLS 证书获取失败: %v", err)
		return
	}
	cert := helpers.EncodeCertificatePEM(tlsCert.Leaf)
	keyBytes, err := x509.MarshalPKCS8PrivateKey(tlsCert.PrivateKey)
	if err != nil {
		logger.Errorf("TLS 证书 Private Key 获取失败: %v", err)
		return
	}

	key := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	trustCerts := ex.Transport.TrustStore.Certificates()
	caCerts := make([][]byte, 0, len(trustCerts))

	fmt.Println("--- CA 证书 Stared ---")
	for _, caCert := range trustCerts {
		caCertBytes := helpers.EncodeCertificatePEM(caCert)
		caCerts = append(caCerts, caCertBytes)
		fmt.Println("---\n", string(caCertBytes), "\n---")
	}
	fmt.Println("--- CA 证书 End ---")
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("--- 私钥 Stared ---\n", string(key), "\n--- 私钥 End ---")
	fmt.Println("--- 证书 Stared ---\n", string(cert), "\n--- 证书 End ---")
}
