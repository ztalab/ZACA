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
	"github.com/pkg/errors"
	"github.com/ztalab/ZACA/pkg/logger/redis_hook"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ciCore zapcore.Core

var (
	std          *Logger
	stdCallerFix *Logger

	n *zap.Logger
)

// Logger
type Logger struct {
	*zap.SugaredLogger
	conf *Conf
}

// Conf to configure
type Conf struct {
	Caller     bool
	Debug      bool
	Level      zapcore.Level
	Encoding   string                 // json, console
	AppInfo    *ConfigAppData         // fixed fields
	HookConfig *redis_hook.HookConfig // set to nil if disabled
	ZapConfig  *zap.Config            // for custom
}

type ConfigAppData struct {
	AppName    string
	AppID      string
	AppVersion string
	AppKey     string
	Channel    string
	SubOrgKey  string
	Language   string
}

// Clone ...
func Clone(l *Logger) *Logger {
	c := *l.conf
	return &Logger{
		SugaredLogger: l.SugaredLogger,
		conf:          &c,
	}
}

// S Get singleton
func S() *Logger {
	return std
}

// N Zap Logger
func N() *zap.Logger {
	return n
}

// GlobalConfig init
func GlobalConfig(conf Conf) error {
	c := conf
	l, err := newLogger(&c)
	if err != nil {
		return err
	}
	std = &Logger{
		SugaredLogger: l.Sugar(),
		conf:          &c,
	}
	stdCallerFix = &Logger{
		SugaredLogger: l.WithOptions(zap.AddCallerSkip(1)).Sugar(),
		conf:          &c,
	}
	n = std.Desugar()
	return nil
}

func init() {
	l, _ := newLogger(&Conf{
		Level: zapcore.InfoLevel,
	})
	std = &Logger{
		SugaredLogger: l.Sugar(),
		conf:          &Conf{},
	}
	stdCallerFix = &Logger{
		SugaredLogger: l.WithOptions(zap.AddCallerSkip(1)).Sugar(),
		conf:          &Conf{},
	}
	n = std.Desugar()
}

// NewZapLogger Create custom Logger
func NewZapLogger(c *Conf) (l *zap.Logger, err error) {
	return newLogger(c)
}

func newLogger(c *Conf) (l *zap.Logger, err error) {
	var conf zap.Config
	if c.ZapConfig != nil {
		conf = *c.ZapConfig
	} else {
		conf = zap.NewProductionConfig()
		conf.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		if c.Debug {
			conf = zap.NewDevelopmentConfig()
			conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		if c.Encoding != "" {
			conf.Encoding = c.Encoding
		} else {
			conf.Encoding = "console"
		}
	}
	conf.Level = zap.NewAtomicLevelAt(c.Level)
	if c.HookConfig != nil {
		hook, err := redis_hook.NewHook(*c.HookConfig)
		if err != nil {
			return nil, err
		}
		_ciCore = NewCiCore(hook)
		l, err = conf.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, _ciCore)
		}))
		if err != nil {
			return nil, errors.Wrap(err, "zap core init error")
		}
	} else {
		l, err = conf.Build()
	}
	if err != nil {
		return nil, errors.Wrap(err, "zap core init error")
	}
	l = l.WithOptions(zap.WithCaller(c.Caller), zap.AddStacktrace(zapcore.ErrorLevel))
	return
}
