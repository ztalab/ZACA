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

package example

import (
	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/ZACA/pkg/logger/redis_hook"
	"go.uber.org/zap/zapcore"
	"log"
)

var (
	EnvEnableRedisOutput bool // Simulated environment variables
	EnvDebug             bool
)

func init() {
	EnvEnableRedisOutput = true
	EnvDebug = true
	initLogger()
}

func initLogger() {
	conf := &logger.Conf{
		Level:  zapcore.DebugLevel, // Output log level
		Caller: true,               //Whether to open record calling folder + number of lines + function name
		Debug:  true,               // Enable debug
		// All logs output to redis are above info level
		AppInfo: &logger.ConfigAppData{
			AppVersion: "1.0",
			Language:   "zh-cn",
		},
	}
	if !EnvDebug || EnvEnableRedisOutput {
		// In case of production environment
		conf.Level = zapcore.InfoLevel
		conf.HookConfig = &redis_hook.HookConfig{
			Key:  "log_key",
			Host: "redis.msp",
			Port: 6380,
		}
	}
	err := logger.GlobalConfig(*conf)
	if err != nil {
		log.Print("[ERR] Logger init error: ", err)
	}
	logger.Infof("info test: %v", "data")
}
