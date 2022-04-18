package ca

import (
	"strings"

	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	logic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/ca"
)

func init() {
	// 载入类型...
	logic.DoNothing()
}

// RoleProfiles 环境隔离类型
// @Tags CA
// @Summary (p1)环境隔离类型
// @Description 环境隔离类型
// @Produce json
// @Param short query bool false "只返回类型列表, 供搜索条件"
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
