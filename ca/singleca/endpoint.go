package singleca

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"

	"gitlab.oneitfarm.com/bifrost/cfssl/api"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/bundle"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/certinfo"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/crl"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/gencrl"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/generator"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/health"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/info"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/initca"
	apiocsp "gitlab.oneitfarm.com/bifrost/cfssl/api/ocsp"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/scan"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/signhandler"
	certsql "gitlab.oneitfarm.com/bifrost/cfssl/certdb/sql"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"

	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/revoke"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/signer"
)

// V1APIPrefix is the prefix of all CFSSL V1 API Endpoints.
var V1APIPrefix = "/api/v1/cfssl/"

// v1APIPath prepends the V1 API prefix to endpoints not beginning with "/"
func v1APIPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = V1APIPrefix + path
	}
	return (&url.URL{Path: path}).String()
}

var wrapHandler = defaultWrapHandler

// The default wrapper simply returns the normal handler and prefixes the path appropriately
func defaultWrapHandler(path string, handler http.Handler, err error) (string, http.Handler, error) {
	return v1APIPath(path), handler, err
}

var endpoints = map[string]func() (http.Handler, error){
	"sign": func() (http.Handler, error) {
		if s == nil {
			return nil, errBadSigner
		}

		h, err := signer.NewHandlerFromSigner(s)
		if err != nil {
			return nil, err
		}

		if conf.CABundleFile != "" && conf.IntBundleFile != "" {
			sh := h.Handler.(*signhandler.Handler)
			if err := sh.SetBundler(conf.CABundleFile, conf.IntBundleFile); err != nil {
				return nil, err
			}
		}

		return h, nil
	},

	"authsign": func() (http.Handler, error) {
		if s == nil {
			return nil, errBadSigner
		}

		h, err := signer.NewAuthHandlerFromSigner(s)
		if err != nil {
			return nil, err
		}

		if conf.CABundleFile != "" && conf.IntBundleFile != "" {
			sh := h.(*api.HTTPHandler).Handler.(*signhandler.AuthHandler)
			if err := sh.SetBundler(conf.CABundleFile, conf.IntBundleFile); err != nil {
				return nil, err
			}
		}

		signer.CountAll()
		return h, nil
	},

	"info": func() (http.Handler, error) {
		if s == nil {
			return nil, errBadSigner
		}
		// Prefetch, 在初始化时运行, 保证证书在启动时被加载完成
		if _, err := keymanager.GetKeeper().GetL3CachedTrustCerts(); err != nil {
			logger.Fatal("证书获取错误: %v", err)
		}
		return info.NewTrustCertsHandler(s, func() []*x509.Certificate {
			certs, err := keymanager.GetKeeper().GetL3CachedTrustCerts()
			if err != nil {
				logger.Errorf("Trust 证书获取错误: %v", err)
			}
			return certs
		})
	},

	"crl": func() (http.Handler, error) {
		if s == nil {
			return nil, errBadSigner
		}

		if db == nil {
			return nil, errNoCertDBConfigured
		}

		return crl.NewHandler(certsql.NewAccessor(db), conf.CAFile, conf.CAKeyFile)
	},

	"gencrl": func() (http.Handler, error) {
		if s == nil {
			return nil, errBadSigner
		}
		return gencrl.NewHandler(), nil
	},

	"newcert": func() (http.Handler, error) {
		if s == nil {
			return nil, errBadSigner
		}
		h := generator.NewCertGeneratorHandlerFromSigner(generator.CSRValidate, s)
		if conf.CABundleFile != "" && conf.IntBundleFile != "" {
			cg := h.(api.HTTPHandler).Handler.(*generator.CertGeneratorHandler)
			if err := cg.SetBundler(conf.CABundleFile, conf.IntBundleFile); err != nil {
				return nil, err
			}
		}
		return h, nil
	},

	"bundle": func() (http.Handler, error) {
		return bundle.NewHandler(conf.CABundleFile, conf.IntBundleFile)
	},

	"newkey": func() (http.Handler, error) {
		return generator.NewHandler(generator.CSRValidate)
	},

	"init_ca": func() (http.Handler, error) {
		return initca.NewHandler(), nil
	},

	"scan": func() (http.Handler, error) {
		return scan.NewHandler(conf.CABundleFile)
	},

	"scaninfo": func() (http.Handler, error) {
		return scan.NewInfoHandler(), nil
	},

	"certinfo": func() (http.Handler, error) {
		if db != nil {
			return certinfo.NewAccessorHandler(certsql.NewAccessor(db)), nil
		}

		return certinfo.NewHandler(), nil
	},

	"ocspsign": func() (http.Handler, error) {
		if ocspSigner == nil {
			return nil, errBadSigner
		}
		return apiocsp.NewHandler(ocspSigner), nil
	},

	"revoke": func() (http.Handler, error) {
		if db == nil {
			return nil, errNoCertDBConfigured
		}
		revoke.CountAll()
		return revoke.NewHandler(certsql.NewAccessor(db)), nil
	},

	"health": func() (http.Handler, error) {
		return health.NewHealthCheck(), nil
	},
}
