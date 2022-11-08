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

// Sign
func (s *Signer) Sign(text []byte) (sign string, err error) {
	switch priv := s.priv.(type) {
	case *ecdsa.PrivateKey:
		sign, err = EcdsaSign(priv, text)
		return
	case *rsa.PrivateKey:
		// Todo supports RSA
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

// Verify Verify signature
func (v *Verifier) Verify(text []byte, sign string) (bool, error) {
	switch pub := v.pub.(type) {
	case *ecdsa.PublicKey:
		return EcdsaVerify(text, sign, pub)
	case *rsa.PublicKey:
		// Todo supports RSA
	default:
	}
	return false, errors.New("algo not supported")
}
