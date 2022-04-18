package http_source

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"strconv"
	"strings"
)

// Doc: https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt
// Reference: http://linuxperf.com/?p=142

func (hs *HttpSource) CurrentMemStat() (stat *resource.MemStat, err error) {
	var m map[string]uint64
	m, err = resource.ReadMapFromRemote(hs.remoteUrl, "/sys/fs/cgroup/memory/memory.stat")
	if err != nil {
		return nil, err
	}

	stat = &resource.MemStat{}
	stat.Total, err = hs.totalMemory(m)
	if err != nil {
		return nil, err
	}

	stat.SwapTotal, stat.SwapUsed = hs.swapState(m)

	stat.Cached = m["total_cache"]
	stat.MappedFile = m["total_mapped_file"]
	// RSS计算规则修改
	memoryUsageInBytes, err := resource.ReadIntFromRemote(hs.remoteUrl, "/sys/fs/cgroup/memory/memory.usage_in_bytes")
	if err != nil {
		stat.RSS = m["total_rss"] + stat.MappedFile
	} else {
		if v, ok := m["total_inactive_file"]; ok {
			if uint64(memoryUsageInBytes) < v {
				memoryUsageInBytes = 0
			} else {
				memoryUsageInBytes -= int64(v)
			}
		}
		stat.RSS = uint64(memoryUsageInBytes)
	}
	return
}

func (hs *HttpSource) getHostMemTotal() (n uint64, err error) {
	var scanner *bufio.Scanner

	ret, err := resource.RemoteGetSourceByte(hs.remoteUrl, "/proc/meminfo")
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(ret)
	scanner = bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] != "MemTotal:" {
			continue
		}
		parts[1] = strings.TrimSpace(parts[1])
		value := strings.TrimSuffix(parts[1], "kB")
		value = strings.TrimSpace(value)
		n, err = strconv.ParseUint(value, 10, 64)
		n *= 1024
		if err != nil {
			return 0, err
		}
		break
	}
	return
}

func (hs *HttpSource) totalMemory(m map[string]uint64) (uint64, error) {
	hostTotal, err := hs.getHostMemTotal()
	if err != nil {
		return 0, err
	}
	limit, ok := m["hierarchical_memory_limit"]
	if !ok {
		return 0, fmt.Errorf("missing hierarchical_memory_limit")
	}
	if hostTotal > limit {
		return limit, nil
	}
	return hostTotal, nil
}

func (hs *HttpSource) swapState(m map[string]uint64) (total uint64, used uint64) {
	memSwap, ok := m["hierarchical_memsw_limit"]
	if !ok {
		return 0, 0
	}

	mem := m["hierarchical_memory_limit"]
	if memSwap == mem {
		return 0, 0
	}

	total = memSwap - mem
	used = m["total_swap"]
	return total, used
}

// 获取当前进程所占用的内存
func (hs *HttpSource) GetRss() (int64, error) {
	return 0, nil
}
