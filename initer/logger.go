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

package initer

import (
	"github.com/ztalab/ZACA/pkg/logger/redis_hook"
	"log"

	"github.com/ztalab/ZACA/pkg/logger"
	"go.uber.org/zap/zapcore"

	"github.com/ztalab/ZACA/core"
)

func initLogger(config *core.Config) {
	conf := &logger.Conf{
		AppInfo: &logger.ConfigAppData{
			AppVersion: config.Version,
			Language:   "zh-cn",
		},
		Debug:  config.Debug,
		Caller: true,
	}
	if config.Debug {
		conf.Level = zapcore.DebugLevel
	} else {
		conf.Level = zapcore.InfoLevel
		if config.Log.LogProxy.Host != "" {
			conf.HookConfig = &redis_hook.HookConfig{
				Key:  config.Log.LogProxy.Key,
				Host: config.Log.LogProxy.Host,
				Port: config.Log.LogProxy.Port,
			}
		}
	}
	if warn := logger.GlobalConfig(*conf); warn != nil {
		log.Print("[WARN] logger init error:", warn)
	}

	log.Print("[INIT] logger init success.")
}
