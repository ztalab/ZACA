package file_source

import (
	"bufio"
	"errors"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type FileSource struct {
	IsInit bool
}

func NewFileSource() resource.Resource {
	res := &FileSource{}
	res.InitData()
	return res
}

func (res *FileSource) InitSuccess() bool {
	return limitedCoreCount > 0
}

func (mf *FileSource) CurrentMemStat() (stat *resource.MemStat, err error) {
	var m map[string]uint64
	m, err = resource.ReadMapFromFile("/sys/fs/cgroup/memory/memory.stat")
	if err != nil {
		return nil, err
	}

	stat = &resource.MemStat{}
	stat.Total, err = totalMemory(m)
	if err != nil {
		return nil, err
	}

	stat.SwapTotal, stat.SwapUsed = swapState(m)

	stat.Cached = m["total_cache"]
	stat.MappedFile = m["total_mapped_file"]
	// RSS计算规则修改
	memoryUsageInBytes, err := resource.ReadIntFromFile("/sys/fs/cgroup/memory/memory.usage_in_bytes")
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

func getHostMemTotal() (n uint64, err error) {
	var scanner *bufio.Scanner

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner = bufio.NewScanner(file)

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

func totalMemory(m map[string]uint64) (uint64, error) {
	hostTotal, err := getHostMemTotal()
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

func swapState(m map[string]uint64) (total uint64, used uint64) {
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

func (*FileSource) GetRss() (int64, error) {
	buf, err := ioutil.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, err
	}
	fields := strings.Split(string(buf), " ")
	if len(fields) < 2 {
		return 0, errors.New("cannot parse statm")
	}

	rss, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, err
	}
	r := decimal.NewFromInt(rss * int64(os.Getpagesize()))
	return r.Div(decimal.NewFromInt32(1024 * 1024)).IntPart(), err
}
