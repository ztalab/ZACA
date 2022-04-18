package initer

import (
	"github.com/urfave/cli"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/certmanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/datastore"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/event"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"log"
	"os"
	"runtime"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/vaultinit"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/vaultsecret"

	// ...
	_ "gitlab.oneitfarm.com/bifrost/capitalizone/util"
)

// Init 初始化
func Init(c *cli.Context) error {
	conf, err := parseConfigs(c)
	if err != nil {
		return err
	}
	initLogger(&conf)
	log.Printf("started with conf: %+v", conf)

	hook.EnableVaultStorage = conf.Vault.Enabled

	l := &core.Logger{Logger: logger.S()}

	db, err := mysqlDialer(&conf, l)
	if err != nil {
		logger.Fatal(err)
	}
	// Redis Connect
	redisClient, err := redisDialer(&conf, l)
	if err != nil {
		logger.Fatal(err)
	}

	i := &core.I{
		Config:             &conf,
		Logger:             l,
		Db:                 db,
		RedisClusterClient: redisClient,
	}

	if hook.EnableVaultStorage {
		logger.Info("启用 Vault 加密储存引擎")
		vaultClient, err := vaultDialer(&conf, l)
		if err != nil {
			logger.Fatal(err)
			return err
		}
		i.VaultClient = vaultClient
		i.VaultSecret = vaultsecret.NewVaultSecret(vaultClient, conf.Vault.Prefix)
	}

	core.Is = i
	// 初始化influxdb
	go influxdbDialer(&conf, l)

	if core.Is.Config.Vault.Enabled {
		vaultinit.Init()
	} else {
		go vaultinit.Init()
	}

	// 资源监控
	if runtime.GOOS == "linux" && redisClient != nil {
		// 初始化推送事件客户端
		event.InitEventClient(redisClient)
		InitMonitor()
	}

	// TODO 迁移
	if os.Getenv("IS_MIGRATION") == "true" {
		datastore.RunMigration()
		os.Exit(1)
	}
	// CA Start
	if err := keymanager.InitKeeper(); err != nil {
		return err
	}
	// Certs
	_ = certmanager.NewCertCleaner()
	// go cc.AutoGC()

	logger.Info("success started.")
	return nil
}
