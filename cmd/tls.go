package cmd

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/singleca"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// InitTlsServer 初始化Tls服务
func InitTlsServer(ctx context.Context, handler *mux.Router) func() {
	addr := core.Is.Config.HTTP.CaListen
	tlsCfg := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return keymanager.GetKeeper().GetCachedTLSKeyPair()
		},
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert, // 运行集群内客户端单向 TLS 获取
	}
	srv := &http.Server{
		Addr:         addr,
		TLSConfig:    tlsCfg,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		logger.Infof("TLS server is running at %s.", addr)
		err := srv.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	if !core.Is.Config.Debug {
		// 时序监控
		metrics := http.NewServeMux()
		metrics.Handle("/metrics", promhttp.Handler())
		metrics.HandleFunc("/debug/pprof/", pprof.Index)
		metrics.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		metrics.HandleFunc("/debug/pprof/profile", pprof.Profile)
		metrics.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		metrics.HandleFunc("/debug/pprof/trace", pprof.Trace)
		metricsAddr := core.Is.Config.HTTP.Listen
		metricsSrv := &http.Server{
			Addr:        metricsAddr,
			Handler:     metrics,
			ReadTimeout: 5 * time.Second,
			//WriteTimeout: 10 * time.Second,
			IdleTimeout: 15 * time.Second,
		}

		go func() {
			logger.Infof("Metrics server is running at %s.", metricsAddr)
			err := metricsSrv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()
	}

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(30))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf(err.Error())
		}
	}
}

// Run 运行服务
func RunTls(ctx context.Context) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	app, err := singleca.Server()
	if err != nil {
		return err
	}
	cleanFunc := InitTlsServer(ctx, app)

EXIT:
	for {
		sig := <-sc
		logger.Infof("接收到信号[%s]", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			state = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}

	cleanFunc()
	logger.Infof("Tls服务退出")
	time.Sleep(time.Second)
	os.Exit(state)
	return nil
}
