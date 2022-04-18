package test

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	cflog "gitlab.oneitfarm.com/bifrost/cfssl/log"
	"net"
	"testing"
	"time"
)

func BenchmarkNormalHTTP(b *testing.B) {
	go func() {
		httpServer()
	}()
	time.Sleep(2 * time.Second)
	client := httpClient(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get("http://127.0.0.1:8083/test")
	}
}

func BenchmarkHTTPS(b *testing.B) {
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
		b.Error("transport 错误: ", err)
	}

	serverTls, err := serverEx.ServerTLSConfig()
	if err != nil {
		b.Error("服务器 tls 获取错误: ", err)
	}
	clientTls, err := clientEx.ClientTLSConfig("127.0.0.1")
	if err != nil {
		b.Error("client tls config get error: ", err)
	}

	go func() {
		httpsServer(serverTls.TLSConfig())
	}()
	client := httpClient(clientTls.TLSConfig())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get("https://127.0.0.1:8082/test")
	}
}

func httpServer() {
	ln, err := net.Listen("tcp4", "0.0.0.0:8083")
	if err != nil {
		panic(err)
	}

	defer ln.Close()

	if err := fasthttp.Serve(ln, func(ctx *fasthttp.RequestCtx) {
		str := ctx.Request.String()
		fmt.Println("服务器接收: ", str)
		ctx.SetStatusCode(200)
		ctx.SetBody([]byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))
	}); err != nil {
		panic(err)
	}
}
