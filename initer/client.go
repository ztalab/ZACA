package initer

import (
	"crypto/tls"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/go-toolbox/rediscluster"
	"net/http"
	"time"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	influx_client "gitlab.oneitfarm.com/bifrost/influxdata/influxdb1-client/v2"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/influxdb"
	"gitlab.oneitfarm.com/bifrost/capitalizone/util"
)

func kubeDialer(config *core.Config, logger *core.Logger) (kubeCli *kubernetes.Clientset, err error) {
	err = util.RetryWithTimeout(func() error {
		kubeConfig, err := rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("failed to create in-cluster kubernetes client configuration: %v", err)
		}
		logger.Infof("kubeconfig: %v", kubeConfig)

		kubeCli, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %v", err)
		}
		return nil
	}, time.Second, 1*time.Minute, logger.SugaredLogger)

	logger.Info("kubernetes client inited.")
	return
}

func mysqlDialer(config *core.Config, logger *core.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysqlDriver.Open(config.Mysql.Dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}
	d, _ := db.DB()
	if err = d.Ping(); err != nil {
		return nil, err
	}
	if err = mysql.Migrate(db); err != nil {
		logger.Errorf("MySQL Schema migrate error: %v", err)
	}
	return db, nil
}

func influxdbDialer(config *core.Config, logger *core.Logger) {
	if !config.Influxdb.Enabled {
		logger.Warn("Influxdb 功能禁用")
		return
	}
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			metrics, err := func() (*influxdb.Metrics, error) {
				client, err := influx_client.NewHTTPClient(influx_client.HTTPConfig{
					Addr:                fmt.Sprintf("http://%v:%v", config.Influxdb.Address, config.Influxdb.Port),
					Username:            config.Influxdb.UserName,
					Password:            config.Influxdb.Password,
					MaxIdleConns:        config.Influxdb.MaxIdleConns,
					MaxIdleConnsPerHost: config.Influxdb.MaxIdleConns,
				})
				if err != nil {
					return nil, err
				}
				if _, _, err := client.Ping(1 * time.Second); err != nil {
					return nil, err
				}
				metrics := influxdb.NewMetrics(&influxdb.HTTPClient{
					Client: client,
					BatchPointsConfig: influx_client.BatchPointsConfig{
						Precision: config.Influxdb.Precision,
						Database:  config.Influxdb.Database,
					},
				}, &config.Influxdb)
				return metrics, nil
			}()
			if err != nil {
				logger.Error("Influxdb Init Fail:", err)
			} else {
				core.Is.Metrics = metrics
				return
			}
		}
	}
}

func vaultDialer(config *core.Config, logger *core.Logger) (*vaultAPI.Client, error) {
	conf := &vaultAPI.Config{
		Address: config.Vault.Addr,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 5 * time.Second,
		},
	}
	cli, err := vaultAPI.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "create Vault client")
	}
	cli.SetToken(config.Vault.Token)
	status, err := cli.Sys().SealStatus()
	if err != nil {
		logger.Errorf("vaule seal status err: %s", err)
		return nil, err
	}
	logger.Infof("sealed: %v, process: %v", status.Sealed, status.Progress)

	return cli, nil
}

func redisDialer(config *core.Config, logger *core.Logger) (cluster *rediscluster.Cluster, err error) {
	fmt.Printf("redis nodes: %s", config.Redis.Nodes)
	if len(config.Redis.Nodes) == 0 {
		logger.Warn("Redis Nodes未配置")
		return nil, nil
	}
	cluster, err = rediscluster.NewCluster(
		&rediscluster.Options{
			StartNodes:   config.Redis.Nodes,
			ConnTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
			KeepAlive:    16,
			AliveTime:    60 * time.Second,
		})

	if err != nil {
		return nil, err
	}

	resp, err := cluster.Do("ping", "")
	if err != nil {
		return nil, err
	}
	logger.Infof("redis ping: %v", resp)

	return
}
