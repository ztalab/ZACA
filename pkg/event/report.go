package event

import (
	"encoding/json"
	"gitlab.oneitfarm.com/bifrost/go-toolbox/rediscluster"
	"time"

	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

const (
	MSP_EVENT             = "msp:event_msg"
	EVENT_TYPE_FUSE       = "fuse"
	EVENT_TYPE_RATE_LIMIT = "ratelimit"
	EVENT_TYPE_HEARTBEAT  = "heartbeat"
	EVENT_TYPE_RESOURCE   = "resource"
)

type EventReport struct {
	redisCluster *rediscluster.Cluster
}

type MSPEvent struct {
	EventType string      `json:"event_type"` // fuse, ratelimit, heartbeat，resource
	EventTime int64       `json:"event_time"`
	EventBody interface{} `json:"event_body"`
}

var _eventReport *EventReport

func InitEventClient(rc *rediscluster.Cluster) *EventReport {
	if _eventReport == nil {
		_eventReport = &EventReport{redisCluster: rc}
	}
	return _eventReport
}

func Client() *EventReport {
	return _eventReport
}

func (ev *EventReport) Report(evType string, eventBody interface{}) {
	msg := MSPEvent{
		EventType: evType,
		EventTime: time.Now().UnixNano() / 1e6,
		EventBody: eventBody,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("event report json.Marshal err", err)
		return
	}
	// 兼容logproxy
	_, err = ev.redisCluster.Do("LPUSH", MSP_EVENT, b)

	if err != nil {
		logger.Errorf("heartbeat report redis rpush", err)
		// return
	}
	//logger.Warnf(string(b))
	return
}
