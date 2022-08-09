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

package ocsp

import (
	"sync/atomic"
	"time"

	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/logic/schema"
	"github.com/ztalab/ZACA/pkg/influxdb"
)

var (
	overallOcspSuccessCounter uint64
	overallOcspFailedCounter  uint64
	overallOcspCachedCounter  uint64
)

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
					"ocsp_success_count": atomic.LoadUint64(&overallOcspSuccessCounter),
				},
				Tags: map[string]string{
					"ip": schema.GetLocalIpLabel(),
				},
			})
			core.Is.Metrics.AddPoint(&influxdb.MetricsData{
				Measurement: schema.MetricsOverall,
				Fields: map[string]interface{}{
					"ocsp_failed_count": atomic.LoadUint64(&overallOcspFailedCounter),
				},
				Tags: map[string]string{
					"ip": schema.GetLocalIpLabel(),
				},
			})
			core.Is.Metrics.AddPoint(&influxdb.MetricsData{
				Measurement: schema.MetricsOverall,
				Fields: map[string]interface{}{
					"ocsp_cached_count": atomic.LoadUint64(&overallOcspCachedCounter),
				},
				Tags: map[string]string{
					"ip": schema.GetLocalIpLabel(),
				},
			})
		}
	}()
}

func AddMetricsPoint(uniqueID string, hitCache bool, certStatus string) {
	if !core.Is.Config.Influxdb.Enabled {
		return
	}
	cacheStatus := "miss"
	if hitCache {
		cacheStatus = "hit"
		atomic.AddUint64(&overallOcspCachedCounter, 1)
	}

	var fieldType string

	if certStatus == CertStatusGood {
		atomic.AddUint64(&overallOcspSuccessCounter, 1)
		fieldType = "success"
	} else {
		atomic.AddUint64(&overallOcspFailedCounter, 1)
		fieldType = "failed"
	}

	core.Is.Metrics.AddPoint(&influxdb.MetricsData{
		Measurement: schema.MetricsOcspResponses,
		Fields: map[string]interface{}{
			"times": 1,
		},
		Tags: map[string]string{
			"unique_id": uniqueID,
			"cache":     cacheStatus,
			"status":    certStatus,
			"type":      fieldType,
			"ip":        schema.GetLocalIpLabel(),
		},
	})
}
