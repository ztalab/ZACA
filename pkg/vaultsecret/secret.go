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

package vaultsecret

import (
	"github.com/ztalab/ZACA/pkg/logger"
	"strings"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/spf13/cast"
)

const (
	StorePEMPath = "pem"
	StoreKeyPath = "key"

	CALocalStoreKey = "local_store"
	CATructCertsKey = "trust_certs"
)

// VaultSecret ...
type VaultSecret struct {
	cli    *vaultAPI.Client
	prefix string
}

// NewVaultSecret ...
func NewVaultSecret(cli *vaultAPI.Client, prefix string) *VaultSecret {
	return &VaultSecret{cli: cli, prefix: strings.TrimSuffix(prefix, "/") + "/"}
}

// StoreCertPEM ...
func (v *VaultSecret) StoreCertPEM(sn string, pem string) error {
	_, err := v.cli.Logical().Write(v.prefix+"data/"+StorePEMPath+"/"+sn, map[string]interface{}{
		"data": map[string]interface{}{
			"pem": pem,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// StoreCertPEMKey ...
func (v *VaultSecret) StoreCertPEMKey(sn string, pem string, key string) error {
	_, err := v.cli.Logical().Write(v.prefix+"data/"+StorePEMPath+"/"+sn, map[string]interface{}{
		"data": map[string]interface{}{
			"pem": pem,
			"key": key,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// GetCertPEM ...
func (v *VaultSecret) GetCertPEM(sn string) (*string, error) {
	data, err := v.cli.Logical().Read(v.prefix + "data/" + StorePEMPath + "/" + sn)
	if err != nil {
		return nil, err
	}
	var pem string
	if data != nil {
		pem = cast.ToString(cast.ToStringMap(data.Data["data"])["pem"])
	}
	return &pem, nil
}

// GetCertPEMKey ...
func (v *VaultSecret) GetCertPEMKey(sn string) (*string, *string, error) {
	data, err := v.cli.Logical().Read(v.prefix + "data/" + StorePEMPath + "/" + sn)
	if err != nil {
		return nil, nil, err
	}
	logger.S().With("data", data.Data).Debugf("Vault Obtain CERT KEY")
	var pem string
	var key string
	if data != nil {
		pem = cast.ToString(cast.ToStringMap(data.Data["data"])["pem"])
		key = cast.ToString(cast.ToStringMap(data.Data["data"])["key"])
	}
	return &pem, &key, nil
}

// DeleteCertPEM ...
func (v *VaultSecret) DeleteCertPEM(sn string) error {
	_, err := v.cli.Logical().Delete(v.prefix + "data/" + StorePEMPath + "/" + sn)
	if err != nil {
		return err
	}
	return nil
}
