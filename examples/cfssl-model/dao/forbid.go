package dao

import (
	"context"
	"time"

	"gitlab.oneitfarm.com/bifrost/capitalizone/examples/cfssl-model/model"

	"github.com/guregu/null"
	"github.com/satori/go.uuid"
)

var (
	_ = time.Second
	_ = null.Bool{}
	_ = uuid.UUID{}
)

// GetAllForbid is a function to get a slice of record(s) from forbid table in the cap database
// params - page     - page requested (defaults to 0)
// params - pagesize - number of records in a page  (defaults to 20)
// params - order    - db sort order column
// error - ErrNotFound, db Find error
func GetAllForbid(ctx context.Context, page, pagesize int64, order string) (results []*model.Forbid, totalRows int, err error) {

	resultOrm := DB.Model(&model.Forbid{})
	resultOrm.Count(&totalRows)

	if page > 0 {
		offset := (page - 1) * pagesize
		resultOrm = resultOrm.Offset(offset).Limit(pagesize)
	} else {
		resultOrm = resultOrm.Limit(pagesize)
	}

	if order != "" {
		resultOrm = resultOrm.Order(order)
	}

	if err = resultOrm.Find(&results).Error; err != nil {
		err = ErrNotFound
		return nil, -1, err
	}

	return results, totalRows, nil
}

// GetForbid is a function to get a single record from the forbid table in the cap database
// error - ErrNotFound, db Find error
func GetForbid(ctx context.Context, argId uint32) (record *model.Forbid, err error) {
	record = &model.Forbid{}
	if err = DB.First(record, argId).Error; err != nil {
		err = ErrNotFound
		return record, err
	}

	return record, nil
}

// AddForbid is a function to add a single record to forbid table in the cap database
// error - ErrInsertFailed, db save call failed
func AddForbid(ctx context.Context, record *model.Forbid) (result *model.Forbid, RowsAffected int64, err error) {
	db := DB.Save(record)
	if err = db.Error; err != nil {
		return nil, -1, ErrInsertFailed
	}

	return record, db.RowsAffected, nil
}

// UpdateForbid is a function to update a single record from forbid table in the cap database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateForbid(ctx context.Context, argId uint32, updated *model.Forbid) (result *model.Forbid, RowsAffected int64, err error) {

	result = &model.Forbid{}
	db := DB.First(result, argId)
	if err = db.Error; err != nil {
		return nil, -1, ErrNotFound
	}

	if err = Copy(result, updated); err != nil {
		return nil, -1, ErrUpdateFailed
	}

	db = db.Save(result)
	if err = db.Error; err != nil {
		return nil, -1, ErrUpdateFailed
	}

	return result, db.RowsAffected, nil
}

// DeleteForbid is a function to delete a single record from forbid table in the cap database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteForbid(ctx context.Context, argId uint32) (rowsAffected int64, err error) {

	record := &model.Forbid{}
	db := DB.First(record, argId)
	if db.Error != nil {
		return -1, ErrNotFound
	}

	db = db.Delete(record)
	if err = db.Error; err != nil {
		return -1, ErrDeleteFailed
	}

	return db.RowsAffected, nil
}
