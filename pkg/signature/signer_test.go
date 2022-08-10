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
	"crypto/ecdsa"
	"fmt"
	"github.com/ztalab/ZACA/pkg/keygen"
	"testing"
)

func TestEcdsaSign(t *testing.T) {
	priv, _, _ := keygen.GenKey(keygen.EcdsaSigAlg)
	s := NewSigner(priv)
	sign, err := s.Sign([]byte("Test"))
	if err != nil {
		panic(err)
	}
	fmt.Println(sign)
}

func TestEcdsaVerify(t *testing.T) {
	text := []byte("Test")
	priv, _, _ := keygen.GenKey(keygen.EcdsaSigAlg)
	s := NewSigner(priv)
	sign, err := s.Sign(text)
	if err != nil {
		panic(err)
	}
	key := priv.(*ecdsa.PrivateKey)
	v := NewVerifier(key.Public())
	result, err := v.Verify(text, sign)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
