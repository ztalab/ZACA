package singleca

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	// ...
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/certdb/sql"
	"gitlab.oneitfarm.com/bifrost/cfssl/cli"
	// ...
	_ "gitlab.oneitfarm.com/bifrost/cfssl/cli/ocspsign"
	"gitlab.oneitfarm.com/bifrost/cfssl/ocsp"
	"gitlab.oneitfarm.com/bifrost/cfssl/signer"
	"gitlab.oneitfarm.com/bifrost/cfssl/signer/local"

	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	ocsp_responder "gitlab.oneitfarm.com/bifrost/capitalizone/ca/ocsp"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/upperca"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

var (
	conf        cli.Config
	s           signer.Signer
	ocspSigner  ocsp.Signer
	db          *sqlx.DB
	router      = mux.NewRouter()
	proxyClient = resty.NewWithClient(&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
)

// registerHandlers instantiates various handlers and associate them to corresponding endpoints.
func registerHandlers() {
	logger := core.Is.Logger.Named("cfssl-handlers")

	disabled := make(map[string]bool)
	if conf.Disable != "" {
		for _, endpoint := range strings.Split(conf.Disable, ",") {
			disabled[endpoint] = true
		}
	}

	for path, getHandler := range endpoints {
		logger.Debugf("getHandler for %s", path)

		if _, ok := disabled[path]; ok {
			logger.Infof("endpoint '%s' is explicitly disabled", path)
		} else if handler, err := getHandler(); err != nil {
			logger.Warnf("endpoint '%s' is disabled: %v", path, err)
		} else {
			if path, handler, err = wrapHandler(path, handler, err); err != nil {
				logger.Warnf("endpoint '%s' is disabled by wrapper: %v", path, err)
			} else {
				logger.Infof("endpoint '%s' is enabled", path)
				router.Handle(path, handler)
			}
		}
	}
	logger.Info("Handler set up complete.")
}

func Server() (*mux.Router, error) {
	var err error
	logger := core.Is.Logger.Named("singleca")

	// 证书签名
	if core.Is.Config.Keymanager.SelfSign {
		conf = cli.Config{
			Disable: "sign,crl,gencrl,newcert,bundle,newkey,init_ca,scan,scaninfo,certinfo,ocspsign,/",
		}
		if err := keymanager.NewSelfSigner().Run(); err != nil {
			logger.Fatalf("自签名证书错误: %v", err)
		}
		router.PathPrefix("/api/v1/cap/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			localPort := core.Is.Config.HTTP.Listen
			splits := strings.Split(localPort, ":")
			localPort = splits[len(splits)-1]
			localUrl := "http://127.0.0.1:" + localPort + "/api/v1/"
			localUrl += strings.TrimPrefix(strings.Replace(request.URL.Path, "/api/v1/cap/", "", 1), "/")
			var resp *resty.Response
			var err error
			switch request.Method {
			case http.MethodGet:
				resp, err = proxyClient.R().
					Get(localUrl)
			case http.MethodPost:
				bodyBytes, _ := ioutil.ReadAll(request.Body)
				resp, err = proxyClient.R().
					SetBody(bodyBytes).
					Post(localUrl)
			default:
				writer.WriteHeader(404)
				writer.Write([]byte("404 not found"))
			}

			if err != nil {
				logger.Errorf("请求错误: %s", err)
				writer.WriteHeader(500)
				writer.Write([]byte("server error"))
			}

			writer.WriteHeader(200)
			writer.Write(resp.Body())
		})
	} else {
		conf = cli.Config{
			Disable: "crl,gencrl,newcert,bundle,newkey,init_ca,scan,scaninfo,certinfo,ocspsign,/",
		}
		if err := keymanager.NewRemoteSigner().Run(); err != nil {
			logger.Fatalf("远程签名证书错误: %v", err)
		}
		// 上级 CA 健康检查
		go upperca.NewChecker().Run()
	}

	logger.Info("Initializing signer")

	// signer 赋值给全局变量 s
	if s, err = local.NewDynamicSigner(
		func() crypto.Signer {
			priv, _, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Priv Key 错误: %v", err)
			}
			return priv
		}, func() *x509.Certificate {
			_, cert, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Cert 错误: %v", err)
			}
			return cert
		}, func() x509.SignatureAlgorithm {
			priv, _, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Priv Key 错误: %v", err)
			}
			return signer.DefaultSigAlgo(priv)
		}, core.Is.Config.Singleca.CfsslConfig.Signing); err != nil {
		logger.Errorf("couldn't initialize signer: %v", err)
		return nil, err
	}
	// 替换 DB SQL
	db, err = sqlx.Open("mysql", core.Is.Config.Mysql.Dsn)
	if err != nil {
		logger.Errorf("Sqlx 初始化出错: %v", err)
		return nil, err
	}
	s.SetDBAccessor(sql.NewAccessor(db))

	if ocspSigner, err = ocsp.NewDynamicSigner(
		func() *x509.Certificate {
			_, cert, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Cert 错误: %v", err)
			}
			return cert
		}, func() crypto.Signer {
			priv, _, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Priv Key 错误: %v", err)
			}
			return priv
			// cfssl 默认 96h
		}, 4*24*time.Hour); err != nil {
		logger.Warnf("couldn't initialize ocsp signer: %v", err)
	}

	endpoints["ocsp"] = func() (http.Handler, error) {
		src, err := ocsp_responder.NewSharedSources(ocspSigner)
		if err != nil {
			logger.Errorf("OCSP Sources 创建错误: %v", err)
			return nil, errors.Wrap(err, "sources 创建错误")
		}
		ocsp_responder.CountAll()
		return ocsp.NewResponder(src, nil), nil
	}

	registerHandlers()

	//tlsCfg := tls.Config{
	//	GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	//		return keymanager.GetKeeper().GetCachedTLSKeyPair()
	//	},
	//	InsecureSkipVerify: true,
	//	ClientAuth:         tls.NoClientCert, // 运行集群内客户端单向 TLS 获取
	//}
	//
	//// 启动 OCSP 服务器
	//go func() {
	//	src, err := ocsp_responder.NewSharedSources(ocspSigner)
	//	if err != nil {
	//		logger.Errorf("OCSP Sources 创建错误: %v", err)
	//		return
	//	}
	//	ocsp_responder.CountAll()
	//	mux := http.NewServeMux()
	//	mux.Handle("/", ocsp.NewResponder(src, nil))
	//
	//	srv := &http.Server{
	//		Addr: core.Is.Config.HTTP.OcspListen,
	//		Handler: mux,
	//	}
	//	logger.Infof("Start OCSP Responser at %s, host: %s", srv.Addr, core.Is.Config.OCSPHost)
	//	if err := srv.ListenAndServe(); err != nil {
	//		logger.Errorf("OCSP Server 启动失败: %s", err)
	//	}
	//}()
	//
	//go func() {
	//	if err := tlsServe(core.Is.Config.HTTP.CaListen, &tlsCfg); err != nil {
	//		logger.Fatalf("CA TLS Server 启动失败: %s", err)
	//	}
	//}()

	return router, nil
}

func tlsServe(addr string, tlsConfig *tls.Config) error {
	server := http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
		Handler:   router,
	}
	return server.ListenAndServeTLS("", "")
}

// OcspServer ocsp服务
func OcspServer() ocsp.Signer {
	logger := core.Is.Logger.Named("singleca")
	ocspSigner, err := ocsp.NewDynamicSigner(
		func() *x509.Certificate {
			_, cert, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Cert 错误: %v", err)
			}
			return cert
		}, func() crypto.Signer {
			priv, _, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
			if err != nil {
				logger.Errorf("获取 Priv Key 错误: %v", err)
			}
			return priv
			// cfssl 默认 96h
		}, 4*24*time.Hour)
	if err != nil {
		logger.Warnf("couldn't initialize ocsp signer: %v", err)
	}
	return ocspSigner
}
