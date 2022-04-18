package influxdb

import (
	cilog "gitlab.oneitfarm.com/bifrost/cilog/v2"
	_ "gitlab.oneitfarm.com/bifrost/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "gitlab.oneitfarm.com/bifrost/influxdata/influxdb1-client/v2"
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
		cilog.Errorf("InfluxDBUDPClient err: %v", err)
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
