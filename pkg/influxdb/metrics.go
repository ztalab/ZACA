package influxdb

import (
	"errors"
	"fmt"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	client "gitlab.oneitfarm.com/bifrost/influxdata/influxdb1-client/v2"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics ...
type Metrics struct {
	mu                 sync.Mutex
	conf               *CustomConfig
	batchPoints        client.BatchPoints
	point              chan *client.Point
	flushTimer         *time.Ticker
	InfluxDBHttpClient *HTTPClient
	counter            uint64
}

// MetricsData ...
type MetricsData struct {
	Measurement string                 `json:"measurement"`
	Fields      map[string]interface{} `json:"fields"`
	Tags        map[string]string      `json:"tags"`
}

// Response ...
type Response struct {
	State int      `json:"state"`
	Data  struct{} `json:"data"`
	Msg   string   `json:"msg"`
}

// NewMetrics ...
func NewMetrics(influxDBHttpClient *HTTPClient, conf *CustomConfig) (metrics *Metrics) {
	bp, err := client.NewBatchPoints(influxDBHttpClient.BatchPointsConfig)
	if err != nil {
		v2log.Named("metrics").Errorf("custom-influxdb client.NewBatchPoints err: %v", err)
		return
	}
	metrics = &Metrics{
		conf:               conf,
		batchPoints:        bp,
		point:              make(chan *client.Point, 16),
		flushTimer:         time.NewTicker(time.Duration(conf.FlushTime) * time.Second), // 默认定时 30s 发送一次数据
		InfluxDBHttpClient: influxDBHttpClient,
	}
	go metrics.worker()
	return
}

func (mt *Metrics) AddPoint(metricsData *MetricsData) {
	if mt == nil {
		return
	}
	//atomic.AddUint64(&mt.counter, 1)
	pt, err := client.NewPoint(metricsData.Measurement, metricsData.Tags, metricsData.Fields, time.Now())
	if err != nil {
		v2log.Named("metrics").Errorf("custom-influxdb client.NewPoint err: %s", err)
		return
	}
	mt.point <- pt
}

func (mt *Metrics) worker() {
	for {
		select {
		case p, ok := <-mt.point:
			if !ok {
				mt.flush()
				return
			}
			mt.batchPoints.AddPoint(p)
			// 当点数量达到50的时候，发送数据
			//fmt.Println("当前缓存的点的个数: ", mt.batchPoints.GetPointsNum())
			//fmt.Println("当前缓存的点FlushSize: ", mt.conf.FlushSize)
			if mt.batchPoints.GetPointsNum() >= mt.conf.FlushSize {
				mt.flush()
			}
		case <-mt.flushTimer.C:
			//fmt.Println("定时器到，flush数据---------------------")
			mt.flush()
		}
	}
}

func (mt *Metrics) flush() {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	if mt.batchPoints.GetPointsNum() == 0 {
		return
	}
	err := mt.Write()
	if err != nil {
		if strings.Contains(err.Error(), io.EOF.Error()) {
			err = nil
		} else {
			v2log.Named("metric").Errorf("custom-influxdb client.Write err: %s", err)
		}
	}
	defer mt.InfluxDBHttpClient.FluxDBHttpClose()
	// 清空所有的点
	mt.batchPoints.ClearPoints()
}

// 写入数据超时处理
func (mt *Metrics) Write() error {
	ch := make(chan error, 1)
	go func() {
		ch <- mt.InfluxDBHttpClient.FluxDBHttpWrite(mt.batchPoints)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(800 * time.Millisecond): // 800豪秒超时
		return errors.New("write timeout")
	}
}

func (mt *Metrics) count() {
	for {
		time.Sleep(time.Second)
		fmt.Println("计数器：", atomic.LoadUint64(&mt.counter))
	}
}
