package main

import (
	"crypto/x509/pkix"
	"fmt"

	"github.com/spf13/pflag"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/attrmgr"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/keygen"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
)

var (
	caURL = pflag.String("ca-url", "https://127.0.0.1:8081", "CA URL")
	token = pflag.String("token", "1fb4d8144367a1cdc59500a2e81f7902a4cd5da4a1f1b2211eff42202b5b70e8", "Auth Token")
)

// PEM
var (
	caPEM   string
	certPEM string
	keyPEM  string
)

func init() {
	pflag.Parse()
}

func main() {
	sign()
}

func sign() error {
	cai := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleGatekeeper /*哨兵*/, *caURL),
		caclient.WithAuthKey(*token),
	)

	mgr, err := cai.NewCertManager()
	if err != nil {
		return err
	}

	// CA PEM
	caPEMBytes, err := mgr.CACertsPEM()
	if err != nil {
		return err
	}
	caPEM = string(caPEMBytes)

	fmt.Println("certs pem: \n", caPEM)

	// KEY PEM
	_, keyPEMBytes, _ := keygen.GenKey(keygen.EcdsaSigAlg)

	// 证书扩展字段
	attr := attrmgr.New()
	ext, _ := attr.ToPkixExtension(&attrmgr.Attributes{
		// 注入参数 Map[string]interface{}
		Attrs: map[string]interface{}{
			"k1": "v1",
			"k2": "v2",
		},
	})

	// gen csr
	csrPEM, _ := keygen.GenCustomExtendCSR(keyPEMBytes, &spiffe.IDGIdentity{
		SiteID:    "site", /* Site 标识 */
		ClusterID: "cluster",
		UniqueID:  "gatekeeper",
	}, &keygen.CertOptions{ /* 通常为固定值 */
		CN:   "msp.sentry",
		Host: "msp.sentry,127.0.0.1",
	}, []pkix.Extension{ext} /* 注入扩展字段 */)

	fmt.Println("CSR PEM:\n", string(csrPEM))

	// get cert
	certPEMBytes, err := mgr.SignPEM(csrPEM, "sentry")
	if err != nil {
		panic(err)
	}
	certPEM = string(certPEMBytes)
	fmt.Println("CERT:\n", certPEM)

	return nil
}
