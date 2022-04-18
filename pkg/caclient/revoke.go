package caclient

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"net/http"

	"github.com/pkg/errors"

	jsoniter "github.com/json-iterator/go"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/signature"
)

var revokePath = "/api/v1/cfssl/revoke"

// This type is meant to be unmarshalled from JSON
type RevokeRequest struct {
	Serial  string `json:"serial"`
	AKI     string `json:"authority_key_id"`
	Reason  string `json:"reason"`
	Nonce   string `json:"nonce"`
	Sign    string `json:"sign"`
	AuthKey string `json:"auth_key"`
	Profile string `json:"profile"`
}

// RevokeItSelf 吊销自身证书
func (ex *Exchanger) RevokeItSelf() error {
	tlsCert, err := ex.Transport.GetCertificate()
	if err != nil {
		return err
	}
	cert := tlsCert.Leaf
	priv := tlsCert.PrivateKey

	if err := revokeCert(ex.caAddr, priv, cert); err != nil {
		return err
	}
	ex.logger.With("sn", cert.SerialNumber.String()).Info("服务下线吊销自身证书")

	return nil
}

func (cai *CAInstance) RevokeCert(priv crypto.PublicKey, cert *x509.Certificate) error {
	return revokeCert(cai.CaAddr, priv, cert)
}

func revokeCert(caAddr string, priv crypto.PublicKey, cert *x509.Certificate) error {
	s := signature.NewSigner(priv)

	nonce := cert.SerialNumber.String()

	sign, err := s.Sign([]byte(nonce))
	if err != nil {
		return err
	}

	req := &RevokeRequest{
		Serial: cert.SerialNumber.String(),
		AKI:    hex.EncodeToString(cert.AuthorityKeyId),
		Reason: "", // 默认为 0
		Nonce:  nonce,
		Sign:   sign,
	}

	reqBytes, _ := jsoniter.Marshal(req)

	buf := bytes.NewBuffer(reqBytes)

	resp, err := httpClient.Post(caAddr+revokePath, "application/json", buf)
	if err != nil {
		return errors.Wrap(err, "请求错误")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("请求错误")
	}

	return nil
}
