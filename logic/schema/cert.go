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

package schema

import (
	"crypto/x509"
	"time"

	"github.com/ztalab/cfssl/certinfo"

	"github.com/ztalab/ZACA/pkg/spiffe"
)

type sdkRole string

// SampleCert Certificate list cert
type SampleCert struct {
	SN        string    `mapstructure:"sn,omitempty" json:"sn"`
	AKI       string    `mapstructure:"aki,omitempty" json:"aki"`
	CN        string    `mapstructure:"cn,omitempty" json:"cn"`
	O         string    `mapstructure:"o,omitempty" json:"o"`
	OU        string    `mapstructure:"ou,omitempty" json:"ou"`
	Role      string   `mapstructure:"role,omitempty" json:"role,omitempty"`
	UniqueId  string    `mapstructure:"unique_id,omitempty" json:"unique_id"`
	Status    string    `mapstructure:"status,omitempty" json:"status" example:"good"`
	IssuedAt  time.Time `mapstructure:"issued_at,omitempty" json:"issued_at,omitempty"`
	NotBefore time.Time `mapstructure:"not_before,omitempty" json:"not_before,omitempty"`
	Expiry    time.Time `mapstructure:"expiry,omitempty" json:"expiry,omitempty"`
	RevokedAt time.Time `mapstructure:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}

// Certificate represents a JSON description of an X.509 certificate.
type Certificate struct {
	Subject            certinfo.Name `mapstructure:"subject,omitempty" json:"subject,omitempty" swaggertype:"object"`
	Issuer             certinfo.Name `mapstructure:"issuer,omitempty" json:"issuer,omitempty" swaggertype:"object"`
	SerialNumber       string        `mapstructure:"serial_number,omitempty" json:"serial_number,omitempty"`
	SANs               []string      `mapstructure:"sans,omitempty" json:"sans,omitempty"`
	NotBefore          time.Time     `mapstructure:"not_before,omitempty" json:"not_before"`
	NotAfter           time.Time     `mapstructure:"not_after,omitempty" json:"not_after"`
	SignatureAlgorithm string        `mapstructure:"sigalg,omitempty" json:"sigalg"`
	AKI                string        `mapstructure:"authority_key_id,omitempty" json:"authority_key_id"`
	SKI                string        `mapstructure:"subject_key_id,omitempty" json:"subject_key_id"`
	RawPEM             string        `mapstructure:"-" json:"-"`
}

// Certificate details cert
type FullCert struct {
	SampleCert
	CertStr  string            `mapstructure:"cert_str,omitempty" json:"cert_str"` // Show certificate details
	CertInfo *Certificate      `mapstructure:"cert_info,omitempty" json:"cert_info,omitempty"`
	RawCert  *x509.Certificate `mapstructure:"-" json:"-"`
}

type CaMetadata struct {
	O        string              `mapstructure:"o,omitempty" json:"o"`
	OU       string              `mapstructure:"ou,omitempty" json:"ou"`
	SpiffeId *spiffe.IDGIdentity `mapstructure:"spiffe_id" json:"spiffe_id,omitempty"`
}
