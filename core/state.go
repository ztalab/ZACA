package core

import (
	"context"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/ztalab/ZACA/core/config"
	"github.com/ztalab/ZACA/pkg/influxdb"
	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/ZACA/pkg/vaultsecret"
	"gorm.io/gorm"
)

// Config ...
type Config struct {
	config.IConfig
}

// Is ...
var Is *I

// Elector ...
type Elector interface {
	IsLeader() bool
}

// Logger ...
type Logger struct {
	*logger.Logger
}

// I ...
type I struct {
	Ctx         context.Context
	Config      *Config
	Logger      *Logger
	Db          *gorm.DB
	Elector     Elector
	Metrics     *influxdb.Metrics
	VaultClient *vaultAPI.Client
	VaultSecret *vaultsecret.VaultSecret
}
