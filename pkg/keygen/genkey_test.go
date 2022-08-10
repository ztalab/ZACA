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

package keygen

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"testing"

	"github.com/ztalab/ZACA/pkg/attrmgr"
	"github.com/ztalab/ZACA/pkg/spiffe"
)

var (
	testID = &spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "test_cluster",
		UniqueID:  "unique_id",
	}
)

func TestGenKey(t *testing.T) {
	cases := []struct {
		SigAlg SupportedSignatureAlgorithms
	}{
		{
			SigAlg: RsaSigAlg,
		},
		{
			SigAlg: EcdsaSigAlg,
		},
	}
	for _, a := range cases {
		_, key, err := GenKey(a.SigAlg)
		if err != nil {
			t.Error(err)
		}
		t.Log(a.SigAlg, " key: ", string(key))
	}
}

func TestGenCSR(t *testing.T) {
	cases := []struct {
		SigAlg SupportedSignatureAlgorithms
	}{
		{
			SigAlg: RsaSigAlg,
		},
		{
			SigAlg: EcdsaSigAlg,
		},
	}
	for _, a := range cases {
		_, key, _ := GenKey(a.SigAlg)
		csr, err := GenCSR(key, CertOptions{
			Host: "test.com,192.168.2.80,domain.com",
			Org:  "test",
		})
		if err != nil {
			t.Error(err)
		}
		t.Log("csr: ", string(csr))
	}
}

func TestGenWorkloadCSR(t *testing.T) {
	id := testID
	cases := []struct {
		SigAlg SupportedSignatureAlgorithms
	}{
		{
			SigAlg: RsaSigAlg,
		},
		{
			SigAlg: EcdsaSigAlg,
		},
	}
	for _, a := range cases {
		_, key, _ := GenKey(a.SigAlg)
		csr, err := GenWorkloadCSR(key, id)
		if err != nil {
			t.Error(err)
		}
		t.Log("csr: ", string(csr))
	}
}

func TestGenCustomExtendCSR(t *testing.T) {
	id := testID
	_, keyPEM, _ := GenKey(RsaSigAlg)
	cases := []struct {
		CertOptions *CertOptions
		Exts        []pkix.Extension
	}{
		{
			CertOptions: &CertOptions{
				CN: "test",
			},
			Exts: []pkix.Extension{
				{
					Id:       asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7, 8, 1},
					Critical: true,
					Value:    []byte("fake data"),
				},
				{
					Id:       asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7, 8, 2},
					Critical: true,
					Value:    []byte("fake data"),
				},
			},
		},
	}
	for _, a := range cases {
		csrBytes, err := GenCustomExtendCSR(keyPEM, id, a.CertOptions, a.Exts)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("csr: ", string(csrBytes))
	}
}

func TestGenAttrExtensionCSR(t *testing.T) {
	id := testID
	_, keyPEM, _ := GenKey(EcdsaSigAlg)

	opts := &CertOptions{
		CN: "test",
	}
	mgr := attrmgr.New()
	ext, _ := mgr.ToPkixExtension(&attrmgr.Attributes{
		Attrs: map[string]interface{}{
			"k1": "v1",
			"k2": "v2",
		},
	})

	csrBytes, err := GenCustomExtendCSR(keyPEM, id, opts, []pkix.Extension{ext})
	if err != nil {
		t.Error(err)
	}
	fmt.Println("csr: ", string(csrBytes))
}
