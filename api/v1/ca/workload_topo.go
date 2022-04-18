package ca

import (
	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	logic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/ca"
)

func init() {
	// 载入类型...
	logic.DoNothing()
}

// WorkloadUnits CA 下 Units
// 	以 UniqueID 为单元
// @Tags CA
// @Summary (p1)服务单元
// @Description CA 下 Units
// @Produce json
// @Param page query int false "页数, 默认1"
// @Param limit_num query int false "页数限制, 默认20"
// @Param unique_id query string false "UniqueID 查询"
// @Success 200 {object} helper.MSPNormalizeHTTPResponseBody{data=helper.MSPNormalizeList{list=[]logic.WorkloadUnit}} " "
// @Failure 400 {object} helper.HTTPWrapErrorResponse
// @Failure 500 {object} helper.HTTPWrapErrorResponse
// @Router /ca/workload_units [get]
func (a *API) WorkloadUnits(c *helper.HTTPWrapContext) (interface{}, error) {
	var req = struct {
		helper.MSPNormalizeListPaginateParams
		UniqueId string `form:"unique_id"`
	}{
		MSPNormalizeListPaginateParams: helper.DefaultMSPNormalizeListPaginateParams,
	}
	c.BindG(&req)

	data, total, err := a.logic.WorkloadUnits(&logic.WorkloadUnitsParams{
		Page:     req.Page,
		PageSize: req.LimitNum,
		UniqueId: req.UniqueId,
	})
	if err != nil {
		return nil, err
	}

	list := helper.MSPNormalizeList{
		List: data,
		Paginate: helper.MSPNormalizePaginate{
			Total:    total,
			Current:  req.Page,
			PageSize: req.LimitNum,
		},
	}
	return list, nil
}
