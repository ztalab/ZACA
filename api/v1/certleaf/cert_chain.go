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

package certleaf

import (
	"errors"

	"github.com/ztalab/ZACA/api/helper"
	caLogic "github.com/ztalab/ZACA/logic/ca"
	logic "github.com/ztalab/ZACA/logic/certleaf"
	"github.com/ztalab/ZACA/logic/schema"
)

// CertChain Certificate chain
// @Tags certleaf
// @Summary (p1)CertChain
// @Description Get certificate chain information
// @Produce json
// @Param self_cert query bool false "Show CA's own certificate chain"
// @Param sn query string false "SN+AKI Query the specified certificate"
// @Param aki query string false "SN+AKI Query the specified certificate"
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

// CertChainFromRoot All certificate chains from the root Perspective
// @Tags certleaf
// @Summary (p1)Root view certificate chain
// @Description All certificate chains from the root Perspective
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
		a.logger.Errorf("Error getting upper CA topology: %s", err)
	}
	chain.Root.Children = children

	for _, child := range chain.Root.Children {
		if child.Metadata.O == leaf.CertInfo.Subject.Organization {
			child.Current = true
		}
	}

	return chain, nil
}
