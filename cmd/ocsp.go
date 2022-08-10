/*
Copyright 2022-present The Ztalab Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ocsp_responder "github.com/ztalab/ZACA/ca/ocsp"
	"github.com/ztalab/ZACA/ca/singleca"
	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/cfssl/ocsp"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// InitOcspServer Initialize OCSP service
func InitOcspServer(ctx context.Context, ocspSigner ocsp.Signer) func() {
	src, err := ocsp_responder.NewSharedSources(ocspSigner)
	if err != nil {
		logger.Errorf("OCSP Sources Create error: %v", err)
		panic(err)
	}
	ocsp_responder.CountAll()
	mux := http.NewServeMux()
	mux.Handle("/", ocsp.NewResponder(src, nil))

	addr := core.Is.Config.HTTP.OcspListen
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	go func() {
		logger.Infof("OCSP server is running at %s.", addr)
		err := srv.ListenAndServe()
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
			err = metricsSrv.ListenAndServe()
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

// RunOcsp Running services
func RunOcsp(ctx context.Context) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	app := singleca.OcspServer()
	cleanFunc := InitOcspServer(ctx, app)

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
	logger.Infof("Exit OCSP service")
	time.Sleep(time.Second)
	os.Exit(state)
	return nil
}
