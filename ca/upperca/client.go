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

package upperca

import (
	"strings"

	cfssl_client "github.com/ztalab/cfssl/api/client"

	"github.com/ztalab/ZACA/ca/keymanager"
)

func ProxyRequest(f func(host string) error) error {
	return keymanager.GetKeeper().RootClient.DoWithRetry(func(remote *cfssl_client.AuthRemote) error {
		host := strings.TrimSuffix(remote.Hosts()[0], "/")
		return f(host)
	})
}
