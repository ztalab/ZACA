// Package ca 配置类展示
package ca

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/config"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

type RoleProfile struct {
	Name           string        `json:"name"`
	Usages         []string      `json:"usages"`
	ExpiryString   string        `json:"expiry_string"`
	ExpiryDuration time.Duration `json:"expiry_duration" swaggertype:"string"`
	AuthKey        string        `json:"auth_key"`
	IsCa           bool          `json:"is_ca"`
}

// RoleProfiles 展示环境隔离状态
//  不需要参数
func (l *Logic) RoleProfiles() ([]RoleProfile, error) {
	cfg := core.Is.Config.Singleca.CfsslConfig
	if cfg == nil {
		l.logger.Error("cfssl config 为空")
		return nil, errors.New("cfssl config 为空")
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
