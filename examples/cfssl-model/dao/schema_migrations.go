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

// GetAllSchemaMigrations is a function to get a slice of record(s) from schema_migrations table in the cap database
// params - page     - page requested (defaults to 0)
// params - pagesize - number of records in a page  (defaults to 20)
// params - order    - db sort order column
// error - ErrNotFound, db Find error
func GetAllSchemaMigrations(ctx context.Context, page, pagesize int64, order string) (results []*model.SchemaMigrations, totalRows int, err error) {

	resultOrm := DB.Model(&model.SchemaMigrations{})
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

// GetSchemaMigrations is a function to get a single record from the schema_migrations table in the cap database
// error - ErrNotFound, db Find error
func GetSchemaMigrations(ctx context.Context, argVersion int64) (record *model.SchemaMigrations, err error) {
	record = &model.SchemaMigrations{}
	if err = DB.First(record, argVersion).Error; err != nil {
		err = ErrNotFound
		return record, err
	}

	return record, nil
}

// AddSchemaMigrations is a function to add a single record to schema_migrations table in the cap database
// error - ErrInsertFailed, db save call failed
func AddSchemaMigrations(ctx context.Context, record *model.SchemaMigrations) (result *model.SchemaMigrations, RowsAffected int64, err error) {
	db := DB.Save(record)
	if err = db.Error; err != nil {
		return nil, -1, ErrInsertFailed
	}

	return record, db.RowsAffected, nil
}

// UpdateSchemaMigrations is a function to update a single record from schema_migrations table in the cap database
// error - ErrNotFound, db record for id not found
// error - ErrUpdateFailed, db meta data copy failed or db.Save call failed
func UpdateSchemaMigrations(ctx context.Context, argVersion int64, updated *model.SchemaMigrations) (result *model.SchemaMigrations, RowsAffected int64, err error) {

	result = &model.SchemaMigrations{}
	db := DB.First(result, argVersion)
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

// DeleteSchemaMigrations is a function to delete a single record from schema_migrations table in the cap database
// error - ErrNotFound, db Find error
// error - ErrDeleteFailed, db Delete failed error
func DeleteSchemaMigrations(ctx context.Context, argVersion int64) (rowsAffected int64, err error) {

	record := &model.SchemaMigrations{}
	db := DB.First(record, argVersion)
	if db.Error != nil {
		return -1, ErrNotFound
	}

	db = db.Delete(record)
	if err = db.Error; err != nil {
		return -1, ErrDeleteFailed
	}

	return db.RowsAffected, nil
}
