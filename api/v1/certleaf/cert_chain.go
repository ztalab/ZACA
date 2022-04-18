package certleaf

import (
	"errors"

	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	caLogic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/ca"
	logic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/certleaf"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/schema"
)

// CertChain 证书链
// @Tags certleaf
// @Summary (p1)CertChain
// @Description 获取证书链信息
// @Produce json
// @Param self_cert query bool false "展示 CA 自身证书链"
// @Param sn query string false "SN+AKI 查询指定证书"
// @Param aki query string false "SN+AKI 查询指定证书"
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=logic.LeafCert} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /certleaf/cert_chain [get]
func (a *API) CertChain(c *helper.HTTPWrapContext) (interface{}, error) {
	var req logic.CertChainParams
	c.BindG(&req)

	leaf, err := a.logic.CertChain(&req)
	if err != nil {
		return nil, err
	}

	return leaf, nil
}

type RootCertChains struct {
	Root *caLogic.IntermediateObject `json:"root"`
}

// CertChainFromRoot Root视角下所有证书链
// @Tags certleaf
// @Summary (p1)根视角证书链
// @Description Root视角下所有证书链
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=RootCertChains} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /certleaf/cert_chain_from_root [get]
func (a *API) CertChainFromRoot(c *helper.HTTPWrapContext) (interface{}, error) {
	leaf, err := a.logic.CertChain(&logic.CertChainParams{
		SelfCert: true,
	})

	if err != nil {
		return nil, err
	}

	if leaf.IssuerCert == nil {
		return nil, errors.New("issuer cert not valid")
	}

	rootCert := leaf.IssuerCert
	chain := RootCertChains{
		Root: &caLogic.IntermediateObject{},
	}
	chain.Root.Metadata = schema.GetCaMetadataFromX509Cert(rootCert.RawCert)
	chain.Root.Certs = append(chain.Root.Certs, rootCert.FullCert)

	children, err := caLogic.NewLogic().UpperCaIntermediateTopology()
	if err != nil {
		a.logger.Errorf("获取上层 CA 拓扑结构错误: %s", err)
	}
	chain.Root.Children = children

	for _, child := range chain.Root.Children {
		if child.Metadata.O == leaf.CertInfo.Subject.Organization {
			child.Current = true
		}
	}

	return chain, nil
}
