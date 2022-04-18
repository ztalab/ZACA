package signature

import (
	"crypto/ecdsa"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/keygen"
	"testing"
)

func TestEcdsaSign(t *testing.T) {
	priv, _, _ := keygen.GenKey(keygen.EcdsaSigAlg)
	s := NewSigner(priv)
	sign, err := s.Sign([]byte("测试"))
	if err != nil {
		panic(err)
	}
	fmt.Println(sign)
}

func TestEcdsaVerify(t *testing.T) {
	text := []byte("测试")
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
