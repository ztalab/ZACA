package ca

import (
	"crypto/tls"
	"net/http"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/upperca"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/dao"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/schema"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
)

const (
	UpperCaApiIntermediateTopology = "/api/v1/cap/ca/intermediate_topology"
)

var httpClient = resty.NewWithClient(&http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
})

type IntermediateObject struct {
	Certs    []*schema.FullCert    `mapstructure:"certs" json:"certs"`
	Metadata schema.CaMetadata     `mapstructure:"metadata" json:"metadata"`
	Children []*IntermediateObject `json:"children"`
	Current  bool                  `json:"current"`
}

// IntermediateTopology 获取自身签发的子集群证书
func (l *Logic) IntermediateTopology() ([]*IntermediateObject, error) {
	db := l.db.Session(&gorm.Session{})
	db = db.Where("ca_label = ?", caclient.RoleIntermediate)
	db = db.Select(
		"ca_label",
		"issued_at",
		"serial_number",
		"authority_key_identifier",
		"status",
		"not_before",
		"expiry",
		"revoked_at",
		"pem",
	)
	list, _, err := dao.GetAllCertificates(db, 1, 100, "issued_at desc")
	if err != nil {
		return nil, errors.Wrap(err, "数据库查询错误")
	}
	l.logger.Debugf("查询结果数量: %v", len(list))
	intermediateMap := make(map[string]*IntermediateObject, 0)
	for _, row := range list {
		rawCert, err := helpers.ParseCertificatePEM([]byte(row.Pem))
		if err != nil {
			l.logger.With("row", row).Errorf("CA 证书解析错误: %s", err)
			continue
		}
		if len(rawCert.Subject.OrganizationalUnit) == 0 || len(rawCert.Subject.Organization) == 0 {
			l.logger.With("row", row).Warn("CA 证书缺少 O/OU 字段")
			continue
		}
		ou := rawCert.Subject.OrganizationalUnit[0]
		if _, ok := intermediateMap[ou]; !ok {
			intermediateMap[ou] = &IntermediateObject{
				Metadata: schema.GetCaMetadataFromX509Cert(rawCert),
			}
		}
		intermediateMap[ou].Certs = append(intermediateMap[ou].Certs, schema.GetFullCertByX509Cert(rawCert))
	}

	result := make([]*IntermediateObject, 0, len(intermediateMap))
	for _, v := range intermediateMap {
		result = append(result, v)
	}

	return result, nil
}

// UpperCaIntermediateTopology 获取上级 CA 的
func (l *Logic) UpperCaIntermediateTopology() ([]*IntermediateObject, error) {
	if core.Is.Config.Keymanager.SelfSign {
		return l.IntermediateTopology()
	}

	var resp *resty.Response
	err := upperca.ProxyRequest(func(host string) error {
		res, err := httpClient.R().Get(host + UpperCaApiIntermediateTopology)
		if err != nil {
			l.logger.With("upperca", host).Errorf("UpperCA 请求错误: %s", err)
			return err
		}
		resp = res
		return nil
	})
	if err != nil {
		l.logger.Errorf("UpperCA 子CA拓扑获取失败: %s", err)
		return nil, err
	}

	body := resp.Body()
	var response struct {
		Data []*IntermediateObject `json:"data"`
	}
	if err := jsoniter.Unmarshal(body, &response); err != nil {
		l.logger.With("body", string(body)).Errorf("json 解析错误: %s", err)
		return nil, errors.Wrap(err, "json 解析错误")
	}

	return response.Data, nil
}
