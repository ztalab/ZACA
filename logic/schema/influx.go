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
	"net"
	"net/url"

	"github.com/ztalab/ZACA/core/config"
)

const (
	MetricsOcspResponses = config.MetricsTablePrefix + "ocsp_responses"
	MetricsUpperCaInfo   = config.MetricsTablePrefix + "upper_ca_info"
	MetricsOverall       = config.MetricsTablePrefix + "ca_overall"
	MetricsCaSign        = config.MetricsTablePrefix + "ca_sign"
	MetricsCaRevoke      = config.MetricsTablePrefix + "ca_revoke"
	MetricsCaCpuMem      = config.MetricsTablePrefix + "cpu_mem"

	MetricsUpperCaTypeInfo = "ca_info"

	MetricsLabelIp = "ip"
)

func GetHostFromUrl(addr string) string {
	host, err := url.Parse(addr)
	if err != nil {
		return ""
	}
	return host.Host
}

func GetLocalIpLabel() string {
	return internetIP
}

var internetIP = getInternetIP()

func getInternetIP() (IP string) {
	// Find native IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				if ip4[0] == 10 {
					// Assign new IP
					IP = ip4.String()
				}
			}
		}
	}
	return
}
