package events

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

const (
	LoggerName                = "events"
	CategoryWorkloadLifecycle = "workload_lifecycle"
)

var CategoriesStrings = map[string]string{
	CategoryWorkloadLifecycle: "Workload生命周期",
}

const (
	OperatorMSP = "MSP平台"
	OperatorSDK = "SDK"
)

type CertOp struct {
	UniqueId string `json:"unique_id"`
	SN       string `json:"sn"`
	AKI      string `json:"aki"`
}

// Op 操作记录
type Op struct {
	Operator string      `json:"operator"` // 操作人
	Category string      `json:"category"` // 分类
	Type     string      `json:"type"`     // 操作类型
	Obj      interface{} `json:"obj"`      // 操作对象
}

func (o *Op) Log() {
	objStr, _ := jsoniter.MarshalToString(o.Obj)
	v2log.Named(LoggerName).
		With(v2log.DynFieldCustomLog1, fmt.Sprintf("%s.%s", o.Category, o.Type)).
		With(v2log.DynFieldCustomLog3, o.Obj).
		Infof("分类: %s, 操作: %s, 操作者: %s, 操作对象: %v", CategoriesStrings[o.Category], o.Type, o.Operator, objStr)
}
