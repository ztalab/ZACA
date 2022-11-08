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

package ca

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/ztalab/cfssl/config"

	"github.com/ztalab/ZACA/core"
)

type RoleProfile struct {
	Name           string        `json:"name"`
	Usages         []string      `json:"usages"`
	ExpiryString   string        `json:"expiry_string"`
	ExpiryDuration time.Duration `json:"expiry_duration" swaggertype:"string"`
	AuthKey        string        `json:"auth_key"`
	IsCa           bool          `json:"is_ca"`
}

// RoleProfiles Show environmental isolation status
//  No parameters are required
func (l *Logic) RoleProfiles() ([]RoleProfile, error) {
	cfg := core.Is.Config.Singleca.CfsslConfig
	if cfg == nil {
		l.logger.Error("cfssl config Empty")
		return nil, errors.New("cfssl config Empty")
	}
	roles := make([]RoleProfile, 0, len(cfg.Signing.Profiles)+1)

	parseRoleProfile := func(name string, profile *config.SigningProfile) RoleProfile {
		role := RoleProfile{
			Name:           strings.Title(name),
			Usages:         profile.Usage,
			ExpiryString:   profile.ExpiryString,
			ExpiryDuration: profile.Expiry,
			AuthKey:        profile.AuthKeyName,
			IsCa:           profile.CAConstraint.IsCA,
		}
		return role
	}

	for name, profile := range cfg.Signing.Profiles {
		roles = append(roles, parseRoleProfile(name, profile))
	}

	roles = append(roles, parseRoleProfile("default", cfg.Signing.Default))

	return roles, nil
}
