package signature

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/pkg/errors"
)

// EcdsaSign 对 text 签名 返回加密结果, 结果为数字证书 r,s 的序列化后拼接, 然后用 hex 转换为 string
func EcdsaSign(priv *ecdsa.PrivateKey, text []byte) (string, error) {
	hash := sha256.Sum256(text)
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		return "", err
	}
	return EcdsaSignEncode(r, s)
}

// EcdsaSignEncode r, s 转换成字符串
func EcdsaSignEncode(r, s *big.Int) (string, error) {
	rt, err := r.MarshalText()
	if err != nil {
		return "", err
	}
	st, err := s.MarshalText()
	if err != nil {
		return "", err
	}
	b := string(rt) + "," + string(st)
	return hex.EncodeToString([]byte(b)), nil
}

// EcdsaSignDecode r, s 字符串解析
func EcdsaSignDecode(sign string) (rint, sint big.Int, err error) {
	b, err := hex.DecodeString(sign)
	if err != nil {
		err = errors.New("decrypt error," + err.Error())
		return
	}
	rs := strings.Split(string(b), ",")
	if len(rs) != 2 {
		err = errors.New("decode fail")
		return
	}
	err = rint.UnmarshalText([]byte(rs[0]))
	if err != nil {
		err = errors.New("decrypt rint fail, " + err.Error())
		return
	}
	err = sint.UnmarshalText([]byte(rs[1]))
	if err != nil {
		err = errors.New("decrypt sint fail, " + err.Error())
		return
	}
	return
}

// EcdsaVerify 校验文本内容是否与签名一致 使用公钥校验签名和文本内容
func EcdsaVerify(text []byte, sign string, pubKey *ecdsa.PublicKey) (bool, error) {
	hash := sha256.Sum256(text)
	rint, sint, err := EcdsaSignDecode(sign)
	if err != nil {
		return false, err
	}
	result := ecdsa.Verify(pubKey, hash[:], &rint, &sint)
	return result, nil
}
