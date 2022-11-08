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

package keyprovider

import (
	"fmt"
	"github.com/ztalab/ZACA/pkg/keygen"
	"github.com/ztalab/ZACA/pkg/spiffe"
	"testing"
)

var (
	testID = &spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "test_cluster",
		UniqueID:  "unique_id",
	}
)

func TestXKeyProvider_Generate(t *testing.T) {
	kp, _ := NewXKeyProvider(&spiffe.IDGIdentity{})
	err := kp.Generate(string(keygen.EcdsaSigAlg), 0)
	if err != nil {
		t.Error(err)
	}
	err = kp.Generate(string(keygen.RsaSigAlg), 0)
	if err != nil {
		t.Error(err)
	}
}

func TestXKeyProvider_CertificateRequest(t *testing.T) {
	kp, _ := NewXKeyProvider(testID)
	err := kp.Generate(string(keygen.EcdsaSigAlg), 0)
	if err != nil {
		t.Error(err)
	}
	csr, err := kp.CertificateRequest(nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----------------------------\n",
		string(csr))
}
