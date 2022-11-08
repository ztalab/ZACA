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

package datastore

import (
	"errors"
	"github.com/ztalab/ZACA/pkg/logger"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/database/mysql/cfssl-model/model"
	"github.com/ztalab/ZACA/pkg/vaultsecret"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrNotFound ...
var ErrNotFound = errors.New("not found")

// policy
const (
	PolicyDB    = "db"
	PolicyVault = "vault"
	PolicyMixed = "mix"
)

// DataStorer Data storage
type DataStorer struct {
	logger      *zap.SugaredLogger
	db          *gorm.DB
	vaultSecret *vaultsecret.VaultSecret
	policy      string
}

// DefaultDataStorer ...
func DefaultDataStorer() *DataStorer {
	return &DataStorer{
		logger:      logger.S().Named("datastore"),
		db:          core.Is.Db,
		vaultSecret: core.Is.VaultSecret,
		policy:      PolicyMixed,
	}
}

// GetWorkloadCertPEM Obtain workload certificate according to SN
func (ds *DataStorer) GetWorkloadCertPEM(sn string) ([]byte, error) {
	getFromDB := func() ([]byte, error) {
		certModel := new(model.Certificates)
		if err := ds.db.Model(&model.Certificates{}).Where("sn = ?", sn).First(certModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		return []byte(certModel.Pem), nil
	}

	getFromVault := func() ([]byte, error) {
		certStr, err := ds.vaultSecret.GetCertPEM(sn)
		if err != nil {
			return nil, err
		}
		return []byte(*certStr), nil
	}

	switch ds.policy {
	case PolicyDB:
		return getFromDB()
	case PolicyVault:
		return getFromVault()
	case PolicyMixed:
		certBytes, _ := getFromVault()
		if len(certBytes) > 0 {
			return certBytes, nil
		}
		return getFromDB()
	default:
		return nil, errors.New("unsupported")
	}
}
