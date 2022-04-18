package test

import (
	"crypto/tls"
	"fmt"
	"github.com/valyala/fasthttp"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	cflog "gitlab.oneitfarm.com/bifrost/cfssl/log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMTls(t *testing.T) {
	cflog.Level = cflog.LevelDebug
	c := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleSidecar, "https://127.0.0.1:8081"),
		caclient.WithOcspAddr("http://127.0.0.1:8082"))
	serverEx, err := c.NewExchanger(&spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "cluster_test",
		UniqueID:  "server1",
	})
	clientEx, err := c.NewExchanger(&spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "cluster_test",
		UniqueID:  "client1",
	})
	if err != nil {
		t.Error("transport 错误: ", err)
	}

	serverTls, err := serverEx.ServerTLSConfig()
	if err != nil {
		t.Error("服务器 tls 获取错误: ", err)
	}
	fmt.Println("------------- 服务器信任证书 --------------")
	fmt.Println(string(helpers.EncodeCertificatesPEM(serverEx.Transport.ClientTrustStore.Certificates())))
	fmt.Println("------------- END 服务器信任证书 --------------")

	clientTls, err := clientEx.ClientTLSConfig("")
	if err != nil {
		t.Error("client tls config get error: ", err)
	}
	fmt.Println("------------- 客户端信任证书 --------------")
	fmt.Println(string(helpers.EncodeCertificatesPEM(clientEx.Transport.TrustStore.Certificates())))
	fmt.Println("------------- END 客户端信任证书 --------------")

	go func() {
		httpsServer(serverTls.TLSConfig())
	}()
	client := httpClient(clientTls.TLSConfig())
	time.Sleep(2 * time.Second)

	var messages = []string{"hello world", "hello", "world"}
	for range messages {
		resp, err := client.Get("https://127.0.0.1:8082/test111111")
		if err != nil {
			fmt.Fprint(os.Stderr, "请求失败: ", err)
		}

		fmt.Println("请求成功: ", resp.Status)
	}
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

func httpsServer(cfg *tls.Config) {
	ln, err := net.Listen("tcp4", "0.0.0.0:8082")
	if err != nil {
		panic(err)
	}

	defer ln.Close()

	lnTls := tls.NewListener(ln, cfg)

	if err := fasthttp.Serve(lnTls, func(ctx *fasthttp.RequestCtx) {
		str := ctx.Request.String()
		fmt.Println("服务器接收: ", str)
		ctx.SetStatusCode(200)
		ctx.SetBody([]byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))
	}); err != nil {
		panic(err)
	}
}
