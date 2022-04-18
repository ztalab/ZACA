package vaultinit

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-discover"
	vaultAPI "github.com/hashicorp/vault/api"
	jsoniter "github.com/json-iterator/go"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/discover/k8s"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gorm.io/gorm"
)

// keyname ...
const (
	StoreKeyName = "vault"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		IdleConnTimeout: 60 * time.Second,
	},
	Timeout: 60 * time.Second,
}

// Init ...
func Init() {
	if os.Getenv("IS_VAULT_INIT") == "false" {
		return
	}

	d := discover.Discover{
		Providers: map[string]discover.Provider{
			"k8s": &k8s.Provider{},
		},
	}
	l := log.New(os.Stderr, "", log.LstdFlags)

retry:
	addrs, err := d.Addrs(core.Is.Config.Vault.Discover, l)
	if err != nil {
		logger.Errorf("Vault addr discover err: %s", err)
		time.Sleep(5 * time.Second)
		goto retry
	}

	if len(addrs) == 0 {
		logger.Error("Vault node = 0")
		time.Sleep(5 * time.Second)
		goto retry
	}

	var inited bool
	// 获取整体 Inited 状态
	{
		for _, addr := range addrs {
			conf := &vaultAPI.Config{
				Address:    "http://" + addr + ":8200",
				HttpClient: httpClient,
			}
			cli, _ := vaultAPI.NewClient(conf)
			status, err := cli.Sys().Health()
			if err != nil {
				logger.With("addr", addr).Errorf("Get init status err: %s", err)
				time.Sleep(5 * time.Second)
				goto retry
			}

			if status.Initialized {
				inited = true
			}
		}
	}

	// 初始化
	if !inited {
		if err := vaultInit(addrs[0]); err != nil {
			time.Sleep(5 * time.Second)
			goto retry
		}
	}

	// 解密
	for _, addr := range addrs {
		if err := vaultUnseal(addr); err != nil {
			logger.With("addr", addr).Errorf("Vault Unseal err: %s", err)
		}
	}

	go func() {
		time.Sleep(1 * time.Minute)
		Init()
	}()
}

func vaultInit(addr string) error {
	conf := &vaultAPI.Config{
		Address:    "http://" + addr + ":8200",
		HttpClient: httpClient,
	}
	cli, _ := vaultAPI.NewClient(conf)
	status, err := cli.Sys().Health()
	if err != nil {
		logger.With("addr", addr).Errorf("Get init status err: %s", err)
		return err
	}

	if status.Initialized {
		return nil
	}

	logger.With("addr", addr).Info("Vault init...")
	resp, err := cli.Sys().Init(&vaultAPI.InitRequest{
		SecretShares:    5,
		SecretThreshold: 3,
	})
	if err != nil {
		logger.With("addr", addr).Errorf("Vault init err: %s", err)
		return err
	}
	logger.With("addr", addr).Infof("Vault inited success") // 敏感信息不能流入日志
	data, _ := jsoniter.MarshalToString(resp)
	// 临时储存 DB
	keyPair := &model.SelfKeypair{
		Name:        StoreKeyName,
		PrivateKey:  sql.NullString{String: data, Valid: true},
		Certificate: sql.NullString{String: "", Valid: true},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := core.Is.Db.Create(keyPair).Error; err != nil {
		logger.With("key", resp).Errorf("Store Vaule key err: %s", err)
	}

	return nil
}

func vaultUnseal(addr string) error {
	conf := &vaultAPI.Config{
		Address:    "http://" + addr + ":8200",
		HttpClient: httpClient,
	}
	cli, _ := vaultAPI.NewClient(conf)
	status, err := cli.Sys().Health()
	if err != nil {
		logger.With("addr", addr).Errorf("Get init status err: %s", err)
		return err
	}

	if !status.Sealed {
		return nil
	}

	keyPair := &model.SelfKeypair{}
	err = core.Is.Db.Where("name = ?", StoreKeyName).Order("id desc").First(keyPair).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("Vault key not found")
			return err
		}
		logger.Errorf("DB query err: %s", err)
		return err
	}
	key := keyPair.PrivateKey.String
	keys := new(vaultAPI.InitResponse)
	if err := jsoniter.UnmarshalFromString(key, keys); err != nil {
		logger.Errorf("Unmarshal keys err: %s", err)
		return err
	}
	logger.With("addr", addr, "keys", keys).Info("Vault unseal...")
	for _, unsealKey := range keys.Keys {
		resp, err := cli.Sys().Unseal(unsealKey)
		if err != nil {
			logger.With("addr", addr).Errorf("Vault Unseal err: %s", err)
			continue
		}
		if !resp.Sealed {
			logger.With("addr", addr).Info("Vault unsealed.")
			break
		}
	}

	status, err = cli.Sys().Health()
	if err != nil {
		return err
	}

	if status.Sealed {
		logger.With("addr", addr).Error("Vault unseal failed")
	}

	return nil
}
