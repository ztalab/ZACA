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
	"strings"

	"github.com/ztalab/ZACA/api/helper"
	logic "github.com/ztalab/ZACA/logic/ca"
)

func init() {
	// Load type...
	logic.DoNothing()
}

// RoleProfiles Environmental isolation type
// @Tags CA
// @Summary (p1)Environmental isolation type
// @Description Environmental isolation type
// @Produce json
// @Param short query bool false "Only a list of types is returned for search criteria"
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=logic.RoleProfile} " "
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=[]string} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/role_profiles [get]
func (a *API) RoleProfiles(c *helper.HTTPWrapContext) (interface{}, error) {
	profiles, err := a.logic.RoleProfiles()
	if err != nil {
		return nil, err
	}
	if c.G.Query("short") == "true" {
		roles := make([]string, 0, len(profiles)-1)
		for _, profile := range profiles {
			if strings.ToLower(profile.Name) == "default" {
				continue
			}
			roles = append(roles, strings.ToLower(profile.Name))
		}
		return roles, nil
	}
	return profiles, nil
}
