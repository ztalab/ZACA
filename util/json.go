package util

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	// RegisterFuzzyDecoders decode input from PHP with tolerance.
	//  It will handle string/number auto conversation, and treat empty [] as empty struct.
	extra.RegisterFuzzyDecoders()
}