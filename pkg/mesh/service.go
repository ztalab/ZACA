package mesh

import (
	"github.com/pkg/errors"

	"github.com/go-resty/resty/v2"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

var httpClient = resty.New()

const (
	MSPAPIPrefix         = "/api/v1"
	MSPAPIServiceDynamic = MSPAPIPrefix + "/service_unit/dynamic?page=1&limit_num=999999"
)

func GetAllDynamicServiceMetadataRaw() ([]byte, error) {
	resp, err := httpClient.R().Get(core.Is.Config.Mesh.MSPPortalAPI + MSPAPIServiceDynamic)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() > 200 {
		v2log.With("header", resp.Header()).Warn("请求 MSP 错误")
		return nil, errors.New("response error")
	}
	return resp.Body(), nil
}
