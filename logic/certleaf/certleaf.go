package certleaf

import (
	"crypto/x509"
	"encoding/hex"

	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/schema"
)

type LeafCert struct {
	IssuerCert *LeafCert `json:"issuer_cert"`
	*schema.FullCert
}

type CertChainParams struct {
	SelfCert bool   `form:"self_cert"` // 展示自己的证书
	SN       string `form:"sn"`
	AKI      string `form:"aki"`
}

// CertChain 证书链
func (l *Logic) CertChain(params *CertChainParams) (*LeafCert, error) {
	var cert *x509.Certificate
	var err error
	if params.SelfCert {
		// Ca 自身证书
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
			return nil, errors.Wrap(err, "数据库查询错误")
		}
		parsedCert, err := helpers.ParseCertificatePEM([]byte(row.Pem))
		if err != nil {
			l.logger.Errorf("证书解析错误: %s", err)
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
