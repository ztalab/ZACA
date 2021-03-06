package cmd

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ztalab/ZACA/ca/keymanager"
	"github.com/ztalab/ZACA/ca/singleca"
	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/pkg/logger"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// InitTlsServer Initialize TLS service
func InitTlsServer(ctx context.Context, handler *mux.Router) func() {
	addr := core.Is.Config.HTTP.CaListen
	tlsCfg := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return keymanager.GetKeeper().GetCachedTLSKeyPair()
		},
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
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
		// Timing monitoring
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
		logger.Infof("Received signal[%s]", sig.String())
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
	logger.Infof("TLS service exit")
	time.Sleep(time.Second)
	os.Exit(state)
	return nil
}
