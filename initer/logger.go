package initer

import (
	"log"

	"gitlab.oneitfarm.com/bifrost/cilog"
	"gitlab.oneitfarm.com/bifrost/cilog/redis_hook"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap/zapcore"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

func initLogger(config *core.Config) {
	conf := &logger.Conf{
		AppInfo: &cilog.ConfigAppData{
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
		conf.HookConfig = &redis_hook.HookConfig{
			Key:  config.Log.LogProxy.Key,
			Host: config.Log.LogProxy.Host,
			Port: config.Log.LogProxy.Port,
		}
	}
	if warn := logger.GlobalConfig(*conf); warn != nil {
		log.Print("[WARN] logger init error:", warn)
	}

	log.Print("[INIT] logger init success.")
}
