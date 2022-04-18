package workload

import (
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/dao"
)

type UnitsForbidQueryParams struct {
	UniqueIds []string `json:"unique_ids"`
}

type UnitForbidQueryItem struct {
	UniqueId string `json:"unique_id"`
	Forbid   bool   `json:"forbid"`
}

type UnitsForbidQueryResult struct {
	Status map[string]UnitForbidQueryItem `json:"status"`
}

// UnitsForbidQuery 查询 unique_id 是否被禁止申请证书
func (l *Logic) UnitsForbidQuery(params *UnitsForbidQueryParams) (*UnitsForbidQueryResult, error) {
	db := l.db.Where("unique_id IN ?", params.UniqueIds).
		Where("deleted_at IS NULL")
	list, _, err := dao.GetAllForbid(db, 1, 1000, "id desc")
	if err != nil {
		l.logger.Errorf("数据库查询错误: %s", err)
		return nil, err
	}
	result := UnitsForbidQueryResult{
		Status: make(map[string]UnitForbidQueryItem),
	}

	l.logger.Debugf("查询结果: %v", list)

	for _, uid := range params.UniqueIds {
		result.Status[uid] = UnitForbidQueryItem{
			UniqueId: uid,
			Forbid:   false,
		}
	}

	for _, row := range list {
		result.Status[row.UniqueID] = UnitForbidQueryItem{
			UniqueId: row.UniqueID,
			Forbid:   true,
		}
	}

	return &result, nil
}
