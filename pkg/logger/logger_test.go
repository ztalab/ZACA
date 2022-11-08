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

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestNewLogger(t *testing.T) {
	defer Sync()
	GlobalConfig(Conf{
		Debug:  true,
		Caller: true,
		AppInfo: &ConfigAppData{
			AppName:    "test",
			AppID:      "test",
			AppVersion: "1.0",
			AppKey:     "test",
			Channel:    "1",
			SubOrgKey:  "key",
			Language:   "zh",
		},
	})
	S().Info("test")
}

func TestColorLogger(t *testing.T) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()

	logger.Info("Now logs should be colored")
}
