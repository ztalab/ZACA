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

package spiffe

import (
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"strings"
)

// IDG Identity
// be like "spiffe://siteid/clusterid/unique_id"
type IDGIdentity struct {
	SiteID    string `json:"site_id"`
	ClusterID string `json:"cluster_id"`
	UniqueID  string `json:"unique_id"`
}

func ParseIDGIdentity(s string) (*IDGIdentity, error) {
	id, err := spiffeid.FromString(s)
	if err != nil {
		return nil, err
	}
	split := strings.Split(strings.Trim(id.Path(), "/"), "/")
	var idi IDGIdentity
	idi.SiteID = id.TrustDomain().String()
	if len(split) > 0 {
		idi.ClusterID = split[0]
	}
	if len(split) > 1 {
		idi.UniqueID = split[1]
	}
	return &idi, nil
}

func (i IDGIdentity) SpiffeID() spiffeid.ID {
	id, _ := spiffeid.New(i.SiteID, i.ClusterID, i.UniqueID)
	return id
}

func (i IDGIdentity) String() string {
	return i.SpiffeID().String()
}
