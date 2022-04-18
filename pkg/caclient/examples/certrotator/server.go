package main

import (
	"crypto/tls"
	"flag"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient/examples/util"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap/zapcore"
	"net"
	"time"
)

var (
	caAddr = flag.String("ca", "", "CA Server")
)

func main() {
	logger.GlobalConfig(logger.Conf{
		Debug: true,
		Level: zapcore.DebugLevel,
	})

	flag.Parse()
	err := NewSidecarMTLSServer()
	if err != nil {
		logger.Fatal(err)
	}
	select {}
}

// mTLS Server 使用示例
func NewSidecarMTLSServer() error {
	c := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleSidecar, *caAddr),
		caclient.WithRotateAfter(10*time.Second),
		caclient.WithAuthKey("0739a645a7d6601d9d45f6b237c4edeadad904f2fce53625dfdd541ec4fc8134"),
	)
	ex, err := c.NewExchanger(&spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "cluster_test",
		UniqueID:  "server1",
	})
	if err != nil {
		return errors.Wrap(err, "Exchanger 初始化失败")
	}
	tlsCfg, err := ex.ServerTLSConfig()
	if err != nil {
		panic(err)
	}
	go func() {
		httpsServer(tlsCfg.TLSConfig())
	}()
	// 启动证书轮换
	go ex.RotateController().Run()
	go func() {
		<-time.After(15 * time.Second)
		logger.Infof("手动触发证书失效")
		ex.Transport.ManualRevoke()
		ex.Transport.RefreshKeys()
		logger.Infof("证书轮换完成")
		// ex.RevokeItSelf()
	}()
	util.ExtractCertFromExchanger(ex)
	return nil
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
		logger.Info("Recv: ", str)
		ctx.SetStatusCode(200)
		ctx.SetBody([]byte("Hello"))
	}); err != nil {
		panic(err)
	}
}

func revokeCert(ex *caclient.Exchanger) {
	ex.Transport.RefreshKeys()
}
