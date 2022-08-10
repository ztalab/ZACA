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

package redis_hook

import (
	"go.uber.org/zap/zapcore"
	"strings"
)

// zap need extra data for fields
func CreateZapOriginLogMessage(entry *zapcore.Entry, data map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{}, len(data))
	if data != nil {
		for k, v := range data {
			fields[k] = v
		}
	}
	var level = strings.ToUpper(entry.Level.String())
	if level == "ERROR" {
		level = "ERR"
	}
	if level == "WARN" {
		level = "WARNING"
	}
	if level == "FATAL" {
		level = "CRIT"
	}
	fields["level"] = level
	fields["message"] = entry.Message
	return fields
}
