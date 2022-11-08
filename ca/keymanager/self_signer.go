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
	"github.com/ztalab/ZACA/pkg/logger"
	"github.com/ztalab/cfssl/initca"
)

// SelfSigner ...
type SelfSigner struct {
	logger *logger.Logger
}

// NewSelfSigner ...
func NewSelfSigner() *SelfSigner {
	return &SelfSigner{
		logger: logger.Named("self-signer"),
	}
}

// Run Self signed certificate and saved
func (ss *SelfSigner) Run() error {
	key, cert, _ := GetKeeper().GetCachedSelfKeyPairPEM()
	if key != nil && cert != nil {
		ss.logger.Info("The certificate already exists. Skip the self signing process")
		return nil
	}
	ss.logger.Warn("No certificate, self signed certificate")
	cert, _, key, err := initca.New(getRootCSRTemplate())
	if err != nil {
		ss.logger.Errorf("initca Create error: %v", err)
		return err
	}
	ss.logger.With("key", string(key), "cert", string(cert)).Debugf("Self signed certificate completed")
	if err = GetKeeper().SetKeyPairPEM(key, cert); err != nil {
		ss.logger.Errorf("Error saving certificate: %v", err)
		return err
	}

	return nil
}
