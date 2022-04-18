package model

import (
	"database/sql"
	"time"

	"github.com/guregu/null"
	"github.com/satori/go.uuid"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
	_ = uuid.UUID{}
)

/*
DB Table Details
-------------------------------------


CREATE TABLE `schema_migrations` (
  `version` bigint(20) NOT NULL,
  `dirty` tinyint(1) NOT NULL,
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4

JSON Sample
-------------------------------------
{    "version": 14,    "dirty": 61}



*/

// SchemaMigrations struct is a row record of the schema_migrations table in the cap database
type SchemaMigrations struct {
	//[ 0] version                                        bigint               null: false  primary: true   isArray: false  auto: false  col: bigint          len: -1      default: []
	Version int64 `gorm:"primary_key;column:version;type:bigint;" json:"version" db:"version"`
	//[ 1] dirty                                          tinyint              null: false  primary: false  isArray: false  auto: false  col: tinyint         len: -1      default: []
	Dirty int32 `gorm:"column:dirty;type:tinyint;" json:"dirty" db:"dirty"`
}

var schema_migrationsTableInfo = &TableInfo{
	Name: "schema_migrations",
	Columns: []*ColumnInfo{

		&ColumnInfo{
			Index:              0,
			Name:               "version",
			Comment:            ``,
			Notes:              ``,
			Nullable:           false,
			DatabaseTypeName:   "bigint",
			DatabaseTypePretty: "bigint",
			IsPrimaryKey:       true,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "bigint",
			ColumnLength:       -1,
			GoFieldName:        "Version",
			GoFieldType:        "int64",
			JSONFieldName:      "version",
			ProtobufFieldName:  "version",
			ProtobufType:       "int64",
			ProtobufPos:        1,
		},

		&ColumnInfo{
			Index:              1,
			Name:               "dirty",
			Comment:            ``,
			Notes:              ``,
			Nullable:           false,
			DatabaseTypeName:   "tinyint",
			DatabaseTypePretty: "tinyint",
			IsPrimaryKey:       false,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "tinyint",
			ColumnLength:       -1,
			GoFieldName:        "Dirty",
			GoFieldType:        "int32",
			JSONFieldName:      "dirty",
			ProtobufFieldName:  "dirty",
			ProtobufType:       "int32",
			ProtobufPos:        2,
		},
	},
}

// TableName sets the insert table name for this struct type
func (s *SchemaMigrations) TableName() string {
	return "schema_migrations"
}

// BeforeSave invoked before saving, return an error if field is not populated.
func (s *SchemaMigrations) BeforeSave() error {
	return nil
}

// Prepare invoked before saving, can be used to populate fields etc.
func (s *SchemaMigrations) Prepare() {
}

// Validate invoked before performing action, return an error if field is not populated.
func (s *SchemaMigrations) Validate(action Action) error {
	return nil
}

// TableInfo return table meta data
func (s *SchemaMigrations) TableInfo() *TableInfo {
	return schema_migrationsTableInfo
}
