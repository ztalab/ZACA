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

package dao

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

// BuildInfo is used to define the application build info, and inject values into via the build process.
type BuildInfo struct {

	// BuildDate date string of when build was performed filled in by -X compile flag
	BuildDate string

	// LatestCommit date string of when build was performed filled in by -X compile flag
	LatestCommit string

	// BuildNumber date string of when build was performed filled in by -X compile flag
	BuildNumber string

	// BuiltOnIP date string of when build was performed filled in by -X compile flag
	BuiltOnIP string

	// BuiltOnOs date string of when build was performed filled in by -X compile flag
	BuiltOnOs string

	// RuntimeVer date string of when build was performed filled in by -X compile flag
	RuntimeVer string
}

type LogSql func(ctx context.Context, sql string)

var (
	// ErrNotFound error when record not found
	ErrNotFound = fmt.Errorf("record Not Found")

	// ErrUnableToMarshalJSON error when json payload corrupt
	ErrUnableToMarshalJSON = fmt.Errorf("json payload corrupt")

	// ErrUpdateFailed error when update fails
	ErrUpdateFailed = fmt.Errorf("db update error")

	// ErrInsertFailed error when insert fails
	ErrInsertFailed = fmt.Errorf("db insert error")

	// ErrDeleteFailed error when delete fails
	ErrDeleteFailed = fmt.Errorf("db delete error")

	// ErrBadParams error when bad params passed in
	ErrBadParams = fmt.Errorf("bad params error")

	// DB reference to database
	DB *gorm.DB

	// AppBuildInfo reference to build info
	AppBuildInfo *BuildInfo

	// Logger function that will be invoked before executing sql
	Logger LogSql
)

// Copy a src struct into a destination struct
func Copy(dst interface{}, src interface{}) error {
	dstV := reflect.Indirect(reflect.ValueOf(dst))
	srcV := reflect.Indirect(reflect.ValueOf(src))

	if !dstV.CanAddr() {
		return errors.New("copy to value is unaddressable")
	}

	if srcV.Type() != dstV.Type() {
		return errors.New("different types can be copied")
	}

	for i := 0; i < dstV.NumField(); i++ {
		f := srcV.Field(i)
		if !isZeroOfUnderlyingType(f.Interface()) {
			dstV.Field(i).Set(f)
		}
	}

	return nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
