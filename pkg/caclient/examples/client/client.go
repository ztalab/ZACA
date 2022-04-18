package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	caAddr     = flag.String("ca", "https://127.0.0.1:8081", "CA Server")
	ocspAddr   = flag.String("ocsp", "http://127.0.0.1:8082", "Ocsp Server")
	serverAddr = flag.String("server", "https://127.0.0.1:6066", "")
	authKey    = "0739a645a7d6601d9d45f6b237c4edeadad904f2fce53625dfdd541ec4fc8134"
)

func init() {
	logger.GlobalConfig(logger.Conf{
		Debug: true,
		Level: zapcore.DebugLevel,
	})
}

func main() {
	flag.Parse()
	client, err := NewSidecarMTLSClient()
	if err != nil {
		logger.Fatalf("Client init error: %v", err)
	}
	ticker := time.Tick(time.Second)
	for i := 0; i < 1000; i++ {
		<-ticker

		resp, err := client.Get(*serverAddr)
		if err != nil {
			logger.With("resp", resp).Error(err)
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Infof("请求结果: %v, %s", resp.StatusCode, body)
	}
}

// mTLS Client 使用示例
func NewSidecarMTLSClient() (*http.Client, error) {
	l, _ := logger.NewZapLogger(&logger.Conf{
		// Level: 2,
		Level: -1,
	})
	c := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleSidecar, *caAddr),
		caclient.WithAuthKey(authKey),
		caclient.WithOcspAddr(*ocspAddr),
		caclient.WithLogger(l),
	)
	ex, err := c.NewExchanger(&spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "cluster_test",
		UniqueID:  "client1",
	})
	if err != nil {
		return nil, errors.Wrap(err, "Exchanger 初始化失败")
	}
	cfger, err := ex.ClientTLSConfig("supreme")
	if err != nil {
		panic(err)
	}
	cfger.BindExtraValidator(func(identity *spiffe.IDGIdentity) error {
		fmt.Println("id: ", identity.String())
		return nil
	})
	tlsCfg := cfger.TLSConfig()
	//tlsCfg.VerifyConnection = func(state tls.ConnectionState) error {
	//	cert := state.PeerCertificates[0]
	//	fmt.Println("服务器证书生成时间: ", cert.NotBefore.String())
	//	return nil
	//}
	client := httpClient(tlsCfg)
	// 启动证书轮换
	go ex.RotateController().Run()
	// util.ExtractCertFromExchanger(ex)

	resp, err := client.Get("http://www.baidu.com")
	if err != nil {
		panic(err)
	}

	fmt.Println("baidu 测试: ", resp.StatusCode)

	return client, nil
}

func httpClient(cfg *tls.Config) *http.Client {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   cfg,
			DisableKeepAlives: true,
		},
	}
	return &client
}
