package resource

import "time"

type MemStat struct {
	Total      uint64 // min(hierarchical_memory_limit, host memory total)
	RSS        uint64 // rss in memory.stat + mapped_file
	Cached     uint64 // mapped_file + unmapped_file + tmpfs
	MappedFile uint64 // mapped_file

	SwapTotal uint64
	SwapUsed  uint64
}

type CPUStat struct {
	LimitedCores float64
	Usage        float64
	Throttled    uint64 // cpu.stat: nr_throttled
}

type DiskStat struct {
	Read  uint64
	Write uint64
}

type NetworkStat struct {
	RxBytes uint64
	TxBytes uint64
}

type CPUStatCallback func(stat *CPUStat, err error)

type DiskStatCallback func(stat *DiskStat, err error)

type NetStatCallback func(stat *NetworkStat, err error)

type Resource interface {
	CurrentMemStat() (stat *MemStat, err error)
	GetCPUStat(interval time.Duration, callback CPUStatCallback)
	CurrentDiskStat(interval time.Duration, callback DiskStatCallback)
	CurrentNetworkStat(interval time.Duration, callback NetStatCallback)
	GetRss() (int64, error)
	GetNetstat() (map[string]int, error)
	GetLimitedCoreCount() (float64, error)
	GetCpuCount() float64
	InitSuccess() bool
	InitData() bool
}
