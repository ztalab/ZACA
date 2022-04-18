package main

import (
	"crypto"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/attrmgr"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/keygen"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ocsp"
)

var (
	caAddr   = flag.String("ca", "https://127.0.0.1:8081", "CA Server")
	ocspAddr = flag.String("ocsp", "http://capitalizone-tls.msp:8082", "Ocsp Server")
)

var (
	keyPEM  []byte
	certPEM []byte
)

func init() {
	logger.GlobalConfig(logger.Conf{
		Debug: true,
		Level: zapcore.DebugLevel,
	})
}

func main() {

	flag.Parse()
	generateCert()
	getCertAttr()
	//httpServer()
}

func generateCert() {
	cai := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleGatekeeper, *caAddr),
		caclient.WithOcspAddr(*ocspAddr),
		caclient.WithAuthKey("1fb4d8144367a1cdc59500a2e81f7902a4cd5da4a1f1b2211eff42202b5b70e8"),
		caclient.WithLogger(logger.N()))
	cm, err := cai.NewCertManager()
	if err != nil {
		panic(err)
	}
	_, keyPEM, _ = keygen.GenKey(keygen.EcdsaSigAlg)
	fmt.Println("keyPEM:\n", string(keyPEM))
	// attr
	mgr := attrmgr.New()
	ext, _ := mgr.ToPkixExtension(&attrmgr.Attributes{
		Attrs: map[string]interface{}{
			"allow_site":     []string{"*"},
			"inbound_port":   5092,
			"socks5":         []string{"*"},
			"tunnel":         map[string]string{},
			"websocket":      []string{"*"},
			"websocket_port": 5091,
			"type":           "server",

			//"type": "client",
			//"site@test.zsnb.xyz:443": map[string]map[int][]string{
			//	"websocket": map[int][]string{
			//		48080: []string{"*"},
			//	},
			//},
		},
	})

	// gen csr
	csrPEM, _ := keygen.GenCustomExtendCSR(keyPEM, &spiffe.IDGIdentity{
		SiteID:    "site",
		ClusterID: "cluster",
		UniqueID:  "gatekeeper",
	}, &keygen.CertOptions{
		CN: "gatekeeper",
		//Host: "test.zsnb.xyz",
		Host: "site@test.zsnb.xyz:443",
	}, []pkix.Extension{ext})

	fmt.Println("CSR:\n", string(csrPEM))

	// get cert
	certPEM, err = cm.SignPEM(csrPEM, "gatekeeper")
	if err != nil {
		panic(err)
	}
	fmt.Println("CERT:\n", string(certPEM))

	caCert, err := cm.CACertsPEM()
	if err != nil {
		panic(err)
	}
	fmt.Println("caCert:\n", string(caCert))
}

func getCertAttr() {
	mgr := attrmgr.New()
	cert, err := helpers.ParseCertificatePEM(certPEM)
	if err != nil {
		panic(err)
	}
	attr, err := mgr.GetAttributesFromCert(cert)
	if err != nil {
		panic(err)
	}

	spew.Dump(attr)
}

func httpServer() {
	fmt.Println("get ca client...")
	cai := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleGatekeeper, *caAddr),
		caclient.WithOcspAddr(*ocspAddr),
		caclient.WithAuthKey("1fb4d8144367a1cdc59500a2e81f7902a4cd5da4a1f1b2211eff42202b5b70e8"),
		caclient.WithLogger(logger.N()))
	fmt.Println("get keypair ...")
	ex, err := cai.NewExchangerWithKeypair(&spiffe.IDGIdentity{}, keyPEM, certPEM)
	if err != nil {
		panic(err)
	}
	fmt.Println("get cert...")
	tlsCert, err := ex.Transport.GetCertificate()
	if err != nil {
		panic(err)
	}
	cm, err := cai.NewCertManager()
	if err != nil {
		panic(err)
	}
	caCert, err := cm.CACert()
	if err != nil {
		panic(err)
	}
	fmt.Println("ocsp validate...")
	ocspReq, _ := ocsp.CreateRequest(tlsCert.Leaf, caCert, &ocsp.RequestOptions{
		Hash: crypto.SHA1,
	})
	ocspResp, err := caclient.SendOcspRequest("http://ocspv2.gw108.oneitfarm.com", ocspReq, tlsCert.Leaf, caCert)
	if err != nil {
		panic(err)
	}
	fmt.Println("OCSP Status: ", ocspResp.Status)
	fmt.Println("get tls config...")
	tlsCfger, err := ex.ServerTLSConfig()
	if err != nil {
		panic(err)
	}
	tlsCfg := tlsCfger.TLSConfig()
	fmt.Printf("tls config: %#v", tlsCfg)
}
