package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	flag.Parse()
	Start(func() {
		client, err := NewFakeMTLSClient()
		if err != nil {
			logger.Fatalf("Client init error: %v", err)
		}
		ticker := time.Tick(time.Second)
		for i := 0; i < 100; i++ {
			<-ticker

			resp, err := client.Get("http://127.0.0.1:8082")
			if err != nil {
				logger.With("resp", resp).Error(err)
				continue
			}
			body, _ := ioutil.ReadAll(resp.Body)
			logger.Infof("请求结果: %v, %s", resp.StatusCode, body)
		}
	})
}

// mTLS Client 使用示例
func NewFakeMTLSClient() (*http.Client, error) {
	c := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleSidecar, "https://127.0.0.1:8081"),
		caclient.WithOcspAddr("http://127.0.0.1:8082"))
	ex, err := c.NewExchanger(&spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "cluster_test",
		UniqueID:  "client1",
	})
	if err != nil {
		return nil, errors.Wrap(err, "Exchanger 初始化失败")
	}
	cfgg, err := ex.ClientTLSConfig("")
	tlsCfg := cfgg.TLSConfig()
	tlsCfg.InsecureSkipVerify = true
	tlsCfg.VerifyPeerCertificate = nil
	tlsCfg.VerifyConnection = func(state tls.ConnectionState) error {
		cert := state.PeerCertificates[0]
		fmt.Println("服务器证书生成时间: ", cert.NotBefore.String())
		return nil
	}
	client := httpClient(tlsCfg)
	// 启动证书轮换
	go ex.RotateController().Run()
	// util.ExtractCertFromExchanger(ex)
	return client, nil
}

func httpClient(cfg *tls.Config) *http.Client {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     cfg,
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 50,
		},
	}
	return &client
}

func Start(f func()) {
	hook.ClientInsecureSkipVerify = true
	os.Chdir("./../../../../")
	os.Setenv("IS_ENV", "test")
	//cli.Start(func(i *core.I) error {
	//	// CA Start
	//	go func() {
	//		err := singleca.Server()
	//		if err != nil {
	//			i.Logger.Fatal(err)
	//		}
	//	}()
	//	return nil
	//}, func(i *core.I) error {
	//	time.Sleep(2 * time.Second)
	//
	//	f()
	//
	//	os.Exit(0)
	//	return nil
	//})

}
