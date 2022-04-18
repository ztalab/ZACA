package core

import (
	"context"

	vaultAPI "github.com/hashicorp/vault/api"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gitlab.oneitfarm.com/bifrost/go-toolbox/rediscluster"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core/config"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/influxdb"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/vaultsecret"
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
	Ctx                context.Context
	Config             *Config
	RedisClusterClient *rediscluster.Cluster
	Logger             *Logger
	Db                 *gorm.DB
	KubeClient         *kubernetes.Clientset
	Elector            Elector
	Metrics            *influxdb.Metrics
	VaultClient        *vaultAPI.Client
	VaultSecret        *vaultsecret.VaultSecret
}
