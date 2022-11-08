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

package influxdb

import (
	_ "github.com/ztalab/ZACA/pkg/influxdb/influxdb-client" // this is important because of the bug in go mod
	client "github.com/ztalab/ZACA/pkg/influxdb/influxdb-client/v2"
	"github.com/ztalab/ZACA/pkg/logger"
)

// UDPClient UDP Client
type UDPClient struct {
	Conf              *Config
	BatchPointsConfig client.BatchPointsConfig
	client            client.Client
}

func (p *UDPClient) newUDPV1Client() *UDPClient {
	udpClient, err := client.NewUDPClient(client.UDPConfig{
		Addr: p.Conf.UDPAddress,
	})
	if err != nil {
		logger.Errorf("InfluxDBUDPClient err: %v", err)
	}
	p.client = udpClient
	return p
}

// FluxDBUDPWrite ...
func (p *UDPClient) FluxDBUDPWrite(bp client.BatchPoints) (err error) {
	err = p.newUDPV1Client().client.Write(bp)
	return
}

// HTTPClient HTTP Client
type HTTPClient struct {
	Client            client.Client
	BatchPointsConfig client.BatchPointsConfig
}

// FluxDBHttpWrite ...
func (p *HTTPClient) FluxDBHttpWrite(bp client.BatchPoints) (err error) {
	return p.Client.Write(bp)
}

// FluxDBHttpClose ...
func (p *HTTPClient) FluxDBHttpClose() (err error) {
	return p.Client.Close()
}
