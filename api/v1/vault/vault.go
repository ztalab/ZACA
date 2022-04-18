package vault

import (
	"errors"
	"os"

	vaultAPI "github.com/hashicorp/vault/api"
	jsoniter "github.com/json-iterator/go"
	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/database/mysql/cfssl-model/model"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/vaultinit"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gorm.io/gorm"
)

// RootToken ...
func RootToken(c *helper.HTTPWrapContext) (interface{}, error) {
	verifyKey := os.Getenv("IS_VAULT_VERIFY_KEY")
	if verifyKey == "" {
		return nil, errors.New("verify key not found")
	}

	if c.G.Query("verify_key") != verifyKey {
		return nil, errors.New("verify key error")
	}

	envRootToken := core.Is.Config.Vault.Token
	if envRootToken != "" {
		return envRootToken, nil
	}

	keyPair := &model.SelfKeypair{}
	if err := core.Is.Db.Where("name = ?", vaultinit.StoreKeyName).Order("id desc").First(keyPair).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("Vault key not found")
			return nil, errors.New("vault key not found")
		}
		logger.Errorf("DB query err: %s", err)
		return nil, err
	}
	key := keyPair.PrivateKey.String
	keys := new(vaultAPI.InitResponse)
	if err := jsoniter.UnmarshalFromString(key, keys); err != nil {
		logger.Errorf("Unmarshal keys err: %s", err)
		return nil, err
	}

	return keys.RootToken, nil
}
