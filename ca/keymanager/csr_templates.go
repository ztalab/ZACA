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
	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/cfssl/csr"
)

// getRootCSRTemplate Root CA
var getRootCSRTemplate = func() *csr.CertificateRequest {
	return &csr.CertificateRequest{
		Names: []csr.Name{
			{O: core.Is.Config.Keymanager.CsrTemplates.RootCa.O},
		},
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 4096,
		},
		CA: &csr.CAConfig{
			Expiry: core.Is.Config.Keymanager.CsrTemplates.RootCa.Expiry,
		},
	}
}

// getIntermediateCSRTemplate
var getIntermediateCSRTemplate = func() *csr.CertificateRequest {
	return &csr.CertificateRequest{
		Names: []csr.Name{
			{
				O:  core.Is.Config.Keymanager.CsrTemplates.IntermediateCa.O,
				OU: core.Is.Config.Keymanager.CsrTemplates.IntermediateCa.Ou,
			},
		},
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 4096,
		},
		CA: &csr.CAConfig{
			Expiry: core.Is.Config.Keymanager.CsrTemplates.IntermediateCa.Expiry,
		},
	}
}
