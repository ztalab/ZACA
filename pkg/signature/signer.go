package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"

	"github.com/pkg/errors"
)

// Signer ...
type Signer struct {
	priv crypto.PrivateKey
}

// NewSigner ...
func NewSigner(priv crypto.PrivateKey) *Signer {
	return &Signer{priv: priv}
}

// Sign 签名
func (s *Signer) Sign(text []byte) (sign string, err error) {
	switch priv := s.priv.(type) {
	case *ecdsa.PrivateKey:
		sign, err = EcdsaSign(priv, text)
		return
	case *rsa.PrivateKey:
		// TODO 支持 RSA
		return "", errors.New("algo not supported")
	default:
		return "", errors.New("algo not supported")
	}
}

// Verifier ...
type Verifier struct {
	pub crypto.PublicKey
}

// NewVerifier ...
func NewVerifier(pub crypto.PublicKey) *Verifier {
	return &Verifier{pub: pub}
}

// Verify 验证签名
func (v *Verifier) Verify(text []byte, sign string) (bool, error) {
	switch pub := v.pub.(type) {
	case *ecdsa.PublicKey:
		return EcdsaVerify(text, sign, pub)
	case *rsa.PublicKey:
		// TODO 支持 RSA
	default:
	}
	return false, errors.New("algo not supported")
}
