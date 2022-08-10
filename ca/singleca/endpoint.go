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

package singleca

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"

	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/cfssl/api"
	"github.com/ztalab/cfssl/api/bundle"
	"github.com/ztalab/cfssl/api/certinfo"
	"github.com/ztalab/cfssl/api/crl"
	"github.com/ztalab/cfssl/api/gencrl"
	"github.com/ztalab/cfssl/api/generator"
	"github.com/ztalab/cfssl/api/health"
	"github.com/ztalab/cfssl/api/info"
	"github.com/ztalab/cfssl/api/initca"
	apiocsp "github.com/ztalab/cfssl/api/ocsp"
	"github.com/ztalab/cfssl/api/scan"
	"github.com/ztalab/cfssl/api/signhandler"
	certsql "github.com/ztalab/cfssl/certdb/sql"

	"github.com/ztalab/ZACA/ca/keymanager"
	"github.com/ztalab/ZACA/ca/revoke"
	"github.com/ztalab/ZACA/ca/signer"
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
		// Prefetch, Run during initialization to ensure that the certificate is loaded at startup
		if _, err := keymanager.GetKeeper().GetL3CachedTrustCerts(); err != nil {
			logger.Fatal("Certificate acquisition error: %v", err)
		}
		return info.NewTrustCertsHandler(s, func() []*x509.Certificate {
			certs, err := keymanager.GetKeeper().GetL3CachedTrustCerts()
			if err != nil {
				logger.Errorf("Trust Certificate acquisition error: %v", err)
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
