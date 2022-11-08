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

package ca

import (
	"github.com/ztalab/ZACA/api/helper"
	logic "github.com/ztalab/ZACA/logic/ca"
)

func init() {
	// load type...
	logic.DoNothing()
}

// IntermediateTopology Sub-CA topology
// @Tags CA
// @Summary Sub-CA topology
// @Description Sub-CA topology
// @Produce json
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=[]logic.IntermediateObject} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/intermediate_topology [get]
func (a *API) IntermediateTopology(c *helper.HTTPWrapContext) (interface{}, error) {
	return a.logic.IntermediateTopology()
}

// UpperCaIntermediateTopology Upper CA topology
// @Tags CA
// @Summary Upper CA topology
// @Description Upper CA topology
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
