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

package keymanager

import (
	"crypto/tls"
	"net/url"

	"github.com/pkg/errors"
	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/cfssl/api/client"
	"github.com/ztalab/cfssl/auth"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/ztalab/ZACA/core"
)

type UpperClients interface {
	DoWithRetry(f func(*client.AuthRemote) error) error
	AllClients() map[string]*client.AuthRemote
}

type upperClients struct {
	// ip to client
	clients map[string]*client.AuthRemote
	logger  *zap.SugaredLogger
}

func (uc *upperClients) DoWithRetry(f func(*client.AuthRemote) error) error {
	if len(uc.clients) == 0 {
		return errors.New("No clients available")
	}
	var errGroup error
	for _, upperClient := range uc.clients {
		err := f(upperClient)
		if err == nil {
			// success
			return nil
		}
		uc.logger.With("upper", upperClient.Hosts()).Warnf("upper ca Execution error: %s", err)
		multierr.AppendInto(&errGroup, err)
	}
	return errGroup
}

func (uc *upperClients) AllClients() map[string]*client.AuthRemote {
	return uc.clients
}

func NewUpperClients(adds []string) (UpperClients, error) {
	if len(adds) == 0 {
		return nil, errors.New("Upper CA Address configuration error")
	}
	ap, err := auth.New(core.Is.Config.Singleca.CfsslConfig.AuthKeys["intermediate"].Key, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Auth key Configuration error")
	}
	clients := make(map[string]*client.AuthRemote)
	for _, addr := range adds {
		upperAddr, err := url.Parse(addr)
		if err != nil {
			return nil, errors.Wrap(err, "Upper CA Address resolution error")
		}
		upperClient := client.NewAuthServer(addr, &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}, ap)
		clients[upperAddr.Host] = upperClient
	}
	logger.Infof("Upper CA Client Quantity: %v", len(clients))
	return &upperClients{
		clients: clients,
		logger:  logger.Named("upperca").SugaredLogger,
	}, nil
}
