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

package events

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/ztalab/ZACA/pkg/logger"
)

const (
	LoggerName                = "events"
	CategoryWorkloadLifecycle = "workload_lifecycle"
)

var CategoriesStrings = map[string]string{
	CategoryWorkloadLifecycle: "Workload life cycle",
}

const (
	OperatorMSP = "MSP platform"
	OperatorSDK = "SDK"
)

type CertOp struct {
	UniqueId string `json:"unique_id"`
	SN       string `json:"sn"`
	AKI      string `json:"aki"`
}

// Op Operation record
type Op struct {
	Operator string      `json:"operator"` // Operator
	Category string      `json:"category"` // Classification
	Type     string      `json:"type"`     // Operation type
	Obj      interface{} `json:"obj"`      // Operation object
}

func (o *Op) Log() {
	objStr, _ := jsoniter.MarshalToString(o.Obj)
	logger.Named(LoggerName).
		With("flag", fmt.Sprintf("%s.%s", o.Category, o.Type)).
		With("data", o.Obj).
		Infof("Classification: %s, Operation: %s, Operator: %s, Operation object: %v", CategoriesStrings[o.Category], o.Type, o.Operator, objStr)
}
