package ca

import (
	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	logic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/ca"
)

func init() {
	// 载入类型...
	logic.DoNothing()
}

// IntermediateTopology 子CA拓扑
// @Tags CA
// @Summary 子CA拓扑
// @Description 子CA拓扑
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=[]logic.IntermediateObject} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/intermediate_topology [get]
func (a *API) IntermediateTopology(c *helper.HTTPWrapContext) (interface{}, error) {
	return a.logic.IntermediateTopology()
}

// UpperCaIntermediateTopology 上层CA拓扑
// @Tags CA
// @Summary 上层CA拓扑
// @Description 上层CA拓扑
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=[]logic.IntermediateObject} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/upper_ca_intermediate_topology [get]
func (a *API) UpperCaIntermediateTopology(c *helper.HTTPWrapContext) (interface{}, error) {
	body, err := a.logic.UpperCaIntermediateTopology()
	if err != nil {
		return nil, err
	}

	return body, nil
}
