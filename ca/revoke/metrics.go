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

package revoke

import (
	"crypto/x509"
	"sync/atomic"
	"time"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/logic/schema"
	"github.com/ztalab/ZACA/pkg/influxdb"
)

var overallRevokeCounter uint64

func CountAll() {
	if !core.Is.Config.Influxdb.Enabled {
		return
	}
	go func() {
		for {
			<-time.After(5 * time.Second)
			core.Is.Metrics.AddPoint(&influxdb.MetricsData{
				Measurement: schema.MetricsOverall,
				Fields: map[string]interface{}{
					"revoke_count": atomic.LoadUint64(&overallRevokeCounter),
				},
				Tags: map[string]string{
					"ip": schema.GetLocalIpLabel(),
				},
			})
		}
	}()
}

func AddMetricsPoint(cert *x509.Certificate) {
	if !core.Is.Config.Influxdb.Enabled {
		return
	}
	if cert == nil {
		return
	}
	atomic.AddUint64(&overallRevokeCounter, 1)
	core.Is.Metrics.AddPoint(&influxdb.MetricsData{
		Measurement: schema.MetricsCaRevoke,
		Fields: map[string]interface{}{
			"certs_num": 1,
		},
		Tags: map[string]string{
			"unique_id": cert.Subject.CommonName,
			"ip":        schema.GetLocalIpLabel(),
		},
	})
}
