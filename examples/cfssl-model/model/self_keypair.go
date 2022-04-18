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


CREATE TABLE `self_keypair` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(40) NOT NULL,
  `private_key` text,
  `certificate` text,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  UNIQUE KEY `id` (`id`),
  KEY `name` (`name`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4

JSON Sample
-------------------------------------
{    "name": "LrZSWvuAbMHHLxiDYKhSGWqZF",    "private_key": "wUAOGjvBXIDoCrsYqDQpetUdT",    "certificate": "QtxLIwyVDgjqgXkVmfmGWhyiC",    "created_at": "2022-09-04T23:03:17.873061067+08:00",    "updated_at": "2188-10-06T19:24:13.423938207+08:00",    "id": 42}


Comments
-------------------------------------
[ 0] column is set for unsignedWarning table: self_keypair does not have a primary key defined, setting col position 1 id as primary key




*/

// SelfKeypair struct is a row record of the self_keypair table in the cap database
type SelfKeypair struct {
	//[ 0] id                                             uint                 null: false  primary: true   isArray: false  auto: true   col: uint            len: -1      default: []
	ID uint32 `gorm:"primary_key;AUTO_INCREMENT;column:id;type:uint;" json:"id" db:"id"`
	//[ 1] name                                           varchar(40)          null: false  primary: false  isArray: false  auto: false  col: varchar         len: 40      default: []
	Name string `gorm:"column:name;type:varchar;size:40;" json:"name" db:"name"`
	//[ 2] private_key                                    text(65535)          null: true   primary: false  isArray: false  auto: false  col: text            len: 65535   default: []
	PrivateKey sql.NullString `gorm:"column:private_key;type:text;size:65535;" json:"private_key" db:"private_key"`
	//[ 3] certificate                                    text(65535)          null: true   primary: false  isArray: false  auto: false  col: text            len: 65535   default: []
	Certificate sql.NullString `gorm:"column:certificate;type:text;size:65535;" json:"certificate" db:"certificate"`
	//[ 4] created_at                                     timestamp            null: true   primary: false  isArray: false  auto: false  col: timestamp       len: -1      default: []
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;" json:"created_at" db:"created_at"`
	//[ 5] updated_at                                     timestamp            null: true   primary: false  isArray: false  auto: false  col: timestamp       len: -1      default: []
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;" json:"updated_at" db:"updated_at"`
}

var self_keypairTableInfo = &TableInfo{
	Name: "self_keypair",
	Columns: []*ColumnInfo{

		&ColumnInfo{
			Index:   0,
			Name:    "id",
			Comment: ``,
			Notes: `column is set for unsignedWarning table: self_keypair does not have a primary key defined, setting col position 1 id as primary key
`,
			Nullable:           false,
			DatabaseTypeName:   "uint",
			DatabaseTypePretty: "uint",
			IsPrimaryKey:       true,
			IsAutoIncrement:    true,
			IsArray:            false,
			ColumnType:         "uint",
			ColumnLength:       -1,
			GoFieldName:        "ID",
			GoFieldType:        "uint32",
			JSONFieldName:      "id",
			ProtobufFieldName:  "id",
			ProtobufType:       "uint32",
			ProtobufPos:        1,
		},

		&ColumnInfo{
			Index:              1,
			Name:               "name",
			Comment:            ``,
			Notes:              ``,
			Nullable:           false,
			DatabaseTypeName:   "varchar",
			DatabaseTypePretty: "varchar(40)",
			IsPrimaryKey:       false,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "varchar",
			ColumnLength:       40,
			GoFieldName:        "Name",
			GoFieldType:        "string",
			JSONFieldName:      "name",
			ProtobufFieldName:  "name",
			ProtobufType:       "string",
			ProtobufPos:        2,
		},

		&ColumnInfo{
			Index:              2,
			Name:               "private_key",
			Comment:            ``,
			Notes:              ``,
			Nullable:           true,
			DatabaseTypeName:   "text",
			DatabaseTypePretty: "text(65535)",
			IsPrimaryKey:       false,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "text",
			ColumnLength:       65535,
			GoFieldName:        "PrivateKey",
			GoFieldType:        "sql.NullString",
			JSONFieldName:      "private_key",
			ProtobufFieldName:  "private_key",
			ProtobufType:       "string",
			ProtobufPos:        3,
		},

		&ColumnInfo{
			Index:              3,
			Name:               "certificate",
			Comment:            ``,
			Notes:              ``,
			Nullable:           true,
			DatabaseTypeName:   "text",
			DatabaseTypePretty: "text(65535)",
			IsPrimaryKey:       false,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "text",
			ColumnLength:       65535,
			GoFieldName:        "Certificate",
			GoFieldType:        "sql.NullString",
			JSONFieldName:      "certificate",
			ProtobufFieldName:  "certificate",
			ProtobufType:       "string",
			ProtobufPos:        4,
		},

		&ColumnInfo{
			Index:              4,
			Name:               "created_at",
			Comment:            ``,
			Notes:              ``,
			Nullable:           true,
			DatabaseTypeName:   "timestamp",
			DatabaseTypePretty: "timestamp",
			IsPrimaryKey:       false,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "timestamp",
			ColumnLength:       -1,
			GoFieldName:        "CreatedAt",
			GoFieldType:        "time.Time",
			JSONFieldName:      "created_at",
			ProtobufFieldName:  "created_at",
			ProtobufType:       "uint64",
			ProtobufPos:        5,
		},

		&ColumnInfo{
			Index:              5,
			Name:               "updated_at",
			Comment:            ``,
			Notes:              ``,
			Nullable:           true,
			DatabaseTypeName:   "timestamp",
			DatabaseTypePretty: "timestamp",
			IsPrimaryKey:       false,
			IsAutoIncrement:    false,
			IsArray:            false,
			ColumnType:         "timestamp",
			ColumnLength:       -1,
			GoFieldName:        "UpdatedAt",
			GoFieldType:        "time.Time",
			JSONFieldName:      "updated_at",
			ProtobufFieldName:  "updated_at",
			ProtobufType:       "uint64",
			ProtobufPos:        6,
		},
	},
}

// TableName sets the insert table name for this struct type
func (s *SelfKeypair) TableName() string {
	return "self_keypair"
}

// BeforeSave invoked before saving, return an error if field is not populated.
func (s *SelfKeypair) BeforeSave() error {
	return nil
}

// Prepare invoked before saving, can be used to populate fields etc.
func (s *SelfKeypair) Prepare() {
}

// Validate invoked before performing action, return an error if field is not populated.
func (s *SelfKeypair) Validate(action Action) error {
	return nil
}

// TableInfo return table meta data
func (s *SelfKeypair) TableInfo() *TableInfo {
	return self_keypairTableInfo
}
