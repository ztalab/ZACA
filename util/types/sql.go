package types

import (
	"database/sql"
	jsoniter "github.com/json-iterator/go"
)

type NullString sql.NullString
type NullInt64 sql.NullInt64
type NullInt32 sql.NullInt32
type NullBool sql.NullBool
type NullTime sql.NullTime

//MarshalJSON method is called by json.Marshal,
//whenever it is of type NullString
func (x *NullString) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return jsoniter.Marshal(x.String)
}

func (x *NullInt64) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return jsoniter.Marshal(x.Int64)
}

func (x *NullInt32) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return jsoniter.Marshal(x.Int32)
}

func (x *NullBool) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return jsoniter.Marshal(x.Bool)
}

func (x *NullTime) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		return []byte("null"), nil
	}
	return jsoniter.Marshal(x.Time)
}