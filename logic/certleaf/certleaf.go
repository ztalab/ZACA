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

package certleaf

import (
	"crypto/x509"
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/ztalab/cfssl/helpers"
	"gorm.io/gorm"

	"github.com/ztalab/ZACA/ca/keymanager"
	"github.com/ztalab/ZACA/database/mysql/cfssl-model/model"
	"github.com/ztalab/ZACA/logic/schema"
)

type LeafCert struct {
	IssuerCert *LeafCert `json:"issuer_cert"`
	*schema.FullCert
}

type CertChainParams struct {
	SelfCert bool   `form:"self_cert"` // Show your certificate
	SN       string `form:"sn"`
	AKI      string `form:"aki"`
}

// CertChain
func (l *Logic) CertChain(params *CertChainParams) (*LeafCert, error) {
	var cert *x509.Certificate
	var err error
	if params.SelfCert {
		_, cert, err = keymanager.GetKeeper().GetCachedSelfKeyPair()
		if err != nil {
			return nil, err
		}
	} else if params.AKI != "" && params.SN != "" {
		db := l.db.Session(&gorm.Session{})
		row := &model.Certificates{}
		if err := db.Where(&model.Certificates{
			SerialNumber:           params.SN,
			AuthorityKeyIdentifier: params.AKI,
		}).First(&row).Error; err != nil {
			return nil, errors.Wrap(err, "Database query error")
		}
		parsedCert, err := helpers.ParseCertificatePEM([]byte(row.Pem))
		if err != nil {
			l.logger.Errorf("Certificate parsing error: %s", err)
			return nil, err
		}
		cert = parsedCert
	} else {
		return nil, errors.New("params not valid")
	}

	return GetLeafCert(cert)
}

func GetLeafCert(cert *x509.Certificate) (*LeafCert, error) {
	if cert == nil {
		return nil, errors.New("cert not valid")
	}

	trustCerts, err := keymanager.GetKeeper().GetL3CachedTrustCerts()
	if err != nil {
		return nil, err
	}

	trustStoreCerts := make(map[string]*x509.Certificate, len(trustCerts)+1)
	for _, trustCert := range trustCerts {
		ski := hex.EncodeToString(trustCert.SubjectKeyId)
		trustStoreCerts[ski] = trustCert
	}

	_, caCert, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
	if err != nil {
		return nil, err
	}

	caCertSki := hex.EncodeToString(caCert.SubjectKeyId)
	trustStoreCerts[caCertSki] = caCert

	leaf := &LeafCert{
		IssuerCert: nil,
		FullCert:   schema.GetFullCertByX509Cert(cert),
	}
	current := leaf
	for {
		aki := hex.EncodeToString(current.RawCert.AuthorityKeyId)
		trustCert, ok := trustStoreCerts[aki]
		if !ok {
			break
		}
		current.IssuerCert = &LeafCert{FullCert: schema.GetFullCertByX509Cert(trustCert)}
		current = current.IssuerCert
	}

	return leaf, nil
}
