package file_source

import (
	"bufio"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Doc: https://www.kernel.org/doc/Documentation/cgroup-v1/cpusets.txt
// Reference: https://segmentfault.com/a/1190000008323952
// Reference: https://my.oschina.net/jxcdwangtao/blog/828648

var (
	coreCount        uint64
	limitedCoreCount float64
	cpuTick          int
)

// Errors
var (
	ErrCantGetCoreCount        = fmt.Errorf("can't get core count")
	ErrCantGetLimitedCoreCount = fmt.Errorf("can't get limited core count")
	ErrNoCPUTick               = fmt.Errorf("no cpu tick")
)

func (cf *FileSource) GetCPUStat(interval time.Duration, callback resource.CPUStatCallback) {
	if cpuTick == 0 {
		callback(nil, ErrNoCPUTick)
		return
	}
	if coreCount == 0 {
		callback(nil, ErrCantGetCoreCount)
		return
	}
	if limitedCoreCount < 0.01 {
		callback(nil, ErrCantGetLimitedCoreCount)
		return
	}

	prevSystem, err := getSystemCPUUsage()
	if err != nil {
		callback(nil, err)
		return
	}

	prevTotal, err := getTotalCPUUsage()
	if err != nil {
		callback(nil, err)
		return
	}

	go func() {
		time.Sleep(interval)

		system, err := getSystemCPUUsage()
		if err != nil {
			callback(nil, err)
			return
		}
		total, err := getTotalCPUUsage()
		if err != nil {
			callback(nil, err)
			return
		}

		throttled, err := getCPUThrottled()
		if err != nil {
			callback(nil, err)
			return
		}

		stat := &resource.CPUStat{}
		stat.LimitedCores = limitedCoreCount
		stat.Throttled = throttled
		cpuDelta := float64(total - prevTotal)
		systemDelta := float64(system-prevSystem) * tickToNano()
		if systemDelta > 1.0 {
			stat.Usage = (cpuDelta / systemDelta) * float64(coreCount) * 100.0
		}
		callback(stat, nil)
	}()
}

// units are difference between /proc/stat and cpuacct.usage
// cpuacct.usage's unit is nano second
// /proc/stat's unit is (1 / CLK_TCK)
func tickToNano() float64 {
	if cpuTick == 0 {
		return 0.0
	}

	return 1000.0 * 1000.0 * 1000.0 / float64(cpuTick)
}

func (fs *FileSource) InitData() bool {
	var err error
	coreCount, err = fs.GetCoreCount()
	if err != nil {
		logger.Warnf("file cpu init error1", err)
	}

	limitedCoreCount, err = fs.GetLimitedCoreCount()
	if err != nil {
		logger.Warnf("file cpu init error2", err)
	}

	out, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		logger.Warnf("file cpu init error3", err)
	}
	cpuTick, err = strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		logger.Warnf("file cpu init error4", err)
	}
	return limitedCoreCount > 0
}

func getSystemCPUUsage() (uint64, error) {
	// $ cat /proc/stat
	// cpu  42812 0 17335 3256641 333 9 1748 0 0 0

	var scanner *bufio.Scanner

	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner = bufio.NewScanner(file)

	prefix := "cpu "
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		line = strings.TrimSpace(strings.TrimLeft(line, prefix))
		parts := strings.Split(line, " ")
		var total uint64
		for _, part := range parts {
			if part == "" {
				continue
			}
			tmp, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parsing uint64 in /proc/stat, err: %v", err)
			}
			total += tmp
		}
		return total, nil
	}
	return 0, fmt.Errorf("cpu line not found in /proc/stat")
}

func getTotalCPUUsage() (uint64, error) {
	return resource.ReadNumberFromFile("/sys/fs/cgroup/cpuacct/cpuacct.usage")
}

func (fs *FileSource) GetCoreCount() (uint64, error) {
	var data []byte
	var err error

	file, err := os.Open("/sys/fs/cgroup/cpuacct/cpuacct.usage_percpu")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	data, err = ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}

	line := strings.TrimSpace(string(data))
	parts := strings.Split(line, " ")
	l := len(parts)
	return uint64(l), nil
}

func getCPUThrottled() (uint64, error) {
	var m map[string]uint64
	var err error
	m, err = resource.ReadMapFromFile("/sys/fs/cgroup/cpu/cpu.stat")
	if err != nil {
		return 0, err
	}
	return m["nr_throttled"], nil
}

func (fs *FileSource) GetLimitedCoreCount() (float64, error) {
	if limitedCoreCount > 0 {
		return limitedCoreCount, nil
	}
	var quota, period int64
	var err error
	quota, err = resource.ReadIntFromFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	if err != nil {
		return 0, err
	}

	if quota == -1 {
		return getLimitedCoreCountFromCPUSet()
	}
	period, err = resource.ReadIntFromFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if err != nil {
		return 0, err
	}

	if period <= 0 {
		return 0, fmt.Errorf("cfs_period_us is zero")
	}

	return float64(quota) / float64(period), nil
}

func getLimitedCoreCountFromCPUSet() (float64, error) {
	var data []byte
	file, err := os.Open("/sys/fs/cgroup/cpuset/cpuset.cpus")
	if err != nil {
		return 0.0, err
	}
	defer file.Close()

	data, err = ioutil.ReadAll(file)
	if err != nil {
		return 0.0, err
	}

	var cores int

	line := strings.TrimSpace(string(data))
	parts := strings.Split(line, ",")
	for _, part := range parts {
		r := strings.Split(part, "-")
		if len(r) == 1 {
			cores++
			continue
		}
		if len(r) > 2 {
			return 0.0, fmt.Errorf("Invalid list format of cpuset.cpus: %s", line)
		}

		f, e1 := strconv.Atoi(r[0])
		t, e2 := strconv.Atoi(r[1])
		if e1 != nil || e2 != nil {
			return 0.0, fmt.Errorf("Invalid list format of cpuset.cpus: %s", line)
		}
		cores += t - f + 1
	}
	return float64(cores), nil
}

func (*FileSource) GetCpuCount() float64 {
	return limitedCoreCount
}
