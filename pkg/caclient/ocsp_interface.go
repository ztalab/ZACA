package caclient

import "crypto/x509"

// OcspClient Ocsp 客户端
type OcspClient interface {
	Validate(leaf, issuer *x509.Certificate) (bool, error)
	Reset()
}
