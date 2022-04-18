package file_source

import (
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"gitlab.oneitfarm.com/bifrost/go-netstat/netstat"
)

func (*FileSource) GetNetstat() (tcp map[string]int, err error) {
	socks, err := netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		logger.Warnf("获取服务TCP连接失败 ", err)
		return
	}
	tcp = make(map[string]int)
	if len(socks) > 0 {
		for _, value := range socks {
			state := value.State.String()
			tcp[state] += 1
		}
	}
	return
}
