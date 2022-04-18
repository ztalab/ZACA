package initer

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/capitalizone/logic/schema"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/event"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/influxdb"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource/file_source"
	utils "gitlab.oneitfarm.com/bifrost/capitalizone/util"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"sync"
	"time"
)

const (
	alertTypeService   = "service"
	resourceTypeMemory = "memory"
	resourceTypeCPU    = "cpu"
	resourceTypeBoth   = "both"

	RESOURCE_CPU_TYPE  = 1
	RESOURCE_MEM_TYPE  = 2
	RESOURCE_BOTH_TYPE = 3
)

// 属性资源结构体池
var resourceDataPool sync.Pool

type MonitorAttributes struct {
	PodIp       string  `json:"pod_ip"`    // 当前pod ip
	UniqueId    string  `json:"unique_id"` // 服务识别号
	Hostname    string  `json:"hostname"`
	ServiceName string  `json:"service_name"` // 服务可读名称
	CpuLimit    float64 `json:"cpu_limit"`    // Cpu阈值
	MemLimit    float64 `json:"mem_limit"`    // 内存阈值

	DiskRead  int64 `json:"disk_read"`  // 磁盘读取速率，bytes/sec
	DiskWrite int64 `json:"disk_write"` // 磁盘写入速率，bytes/sec

	Cpu          float64 `json:"cpu"`            // 使用的cpu，单位：%
	TotalMemory  int64   `json:"total_memory"`   // 总内存，单位：mb
	Memory       float64 `json:"memory"`         // 使用的内存，单位：%
	MemoryMB     int64   `json:"memory_mb"`      // 使用的内存，单位：mb
	MemoryMBAV   int64   `json:"memory_mb_av"`   // 剩余的内存，单位：mb
	CpuCoreCount float64 `json:"cpu_core_count"` // cpu核心
}

// 资源监控
func InitMonitor() {
	utils.GoWithRecover(func() {
		InitMonitorHandle()
	}, func(r interface{}) {
		InitMonitor()
	})
}

// InitMonitorHandle 监控
func InitMonitorHandle() {
	// 初始化池
	attr := new(MonitorAttributes)
	cpu := make(chan float64)
	disk := make(chan resource.DiskStat)
	attr.PodIp = schema.GetLocalIpLabel()
	attr.UniqueId = "capitalizone_" + core.Is.Config.Registry.Command
	attr.Hostname = core.Is.Config.Hostname
	attr.ServiceName = "CA中心"

	attr.CpuLimit = core.Is.Config.Metrics.CpuLimit
	attr.MemLimit = core.Is.Config.Metrics.MemLimit

	cpuTimes := time.Millisecond * 250

	resourceHd := file_source.NewFileSource()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	// 初始化池
	resourceDataPool.New = func() interface{} {
		return &resourceData{}
	}
	// cpu阈值计算
	for range ticker.C {

		if resourceHd.InitData() {
			attr.CpuCoreCount = resourceHd.GetCpuCount()
		}

		// 自身容器资源情况
		scRsd := resourceDataPool.Get().(*resourceData)
		scResource := scRsd.getResource(resourceHd, cpu, cpuTimes, disk)

		// 容器资源使用情况
		attr.TotalMemory = scResource.ContainerTotalMemoryMB
		attr.MemoryMB = scResource.ContainerUsedMemoryMB
		attr.Memory = scResource.UsedMemoryRatio
		attr.MemoryMBAV = scResource.ContainerTotalMemoryMB - scResource.ContainerUsedMemoryMB // 业务剩余内存
		attr.Cpu = scResource.ContainerCpuUsedRatio
		attr.DiskRead = int64(scResource.ContainerDiskRead)
		attr.DiskWrite = int64(scResource.ContainerDiskWrite)

		// report to influxdb
		if core.Is.Config.Influxdb.Enabled {
			AddMetricsPoint(attr)
		}
		// 告警
		AlertError(attr)

		// 回收
		scRsd.reset()
		resourceDataPool.Put(scRsd)
	}
}

// 告警
func AlertError(attr *MonitorAttributes) {
	var serviceCpuLimit float64
	if attr.CpuCoreCount > 0 {
		serviceCpuLimit = attr.CpuCoreCount * attr.CpuLimit
	}
	// 服务资源预警
	if serviceCpuLimit > 0 {
		if attr.Memory > attr.MemLimit || attr.Cpu > serviceCpuLimit {
			resourceType := 0
			if attr.Memory > attr.MemLimit && attr.Cpu > serviceCpuLimit {
				// cpu + 内存 预警
				resourceType = RESOURCE_BOTH_TYPE
			} else if attr.Memory > attr.MemLimit {
				// 内存预警
				resourceType = RESOURCE_MEM_TYPE
			} else if attr.Cpu > serviceCpuLimit {
				// cpu预警
				resourceType = RESOURCE_CPU_TYPE
			}
			data := MSPReportResource{
				UniqueId:     attr.UniqueId,
				AlertType:    4,
				ResourceType: resourceType,
				Cpu:          attr.Cpu,
				CpuCoreCount: attr.CpuCoreCount,
				CpuThreshold: serviceCpuLimit,
				Mem:          attr.Memory,
				MemoryMB:     attr.MemoryMB,
				MemThreshold: attr.MemLimit,
				HostName:     attr.Hostname,
				ServiceName:  attr.ServiceName,
				// cpu内存 request
				//ServiceMemRequests: attr.ServiceMemRequests,
				//ServiceMemLimits:   attr.ServiceMemLimits,
				//ServiceCPURequests: attr.ServiceCPURequests,
				//ServiceCPULimits:   attr.ServiceCPULimits,
				EventTime: time.Now().UnixNano() / 1e6,
			}
			event.Client().Report(event.EVENT_TYPE_RESOURCE, data)
			recordResourceWarning(data)
		}
	}
}

var (
	_metricsFields = make(map[string]interface{})
	_tagFields     = make(map[string]string)
)

// 时序日志
func AddMetricsPoint(attr *MonitorAttributes) {
	if !core.Is.Config.Influxdb.Enabled {
		return
	}

	_metricsFields["cpu"] = attr.Cpu
	_metricsFields["cpu_core_count"] = attr.CpuCoreCount
	_metricsFields["disk_read"] = attr.DiskRead
	_metricsFields["disk_write"] = attr.DiskWrite
	_metricsFields["total_memory"] = attr.TotalMemory
	_metricsFields["memory"] = attr.Memory
	_metricsFields["memory_mb"] = attr.MemoryMB
	_metricsFields["memory_mb_av"] = attr.MemoryMBAV

	_tagFields["pod_ip"] = attr.PodIp
	_tagFields["unique_id"] = attr.UniqueId
	_tagFields["hostname"] = attr.Hostname
	_tagFields["service_name"] = attr.ServiceName

	core.Is.Metrics.AddPoint(&influxdb.MetricsData{
		Measurement: schema.MetricsCaCpuMem,
		Fields:      _metricsFields,
		Tags:        _tagFields,
	})
}

type resourceData struct {
	ContainerTotalMemoryMB int64          // 容器总内存，单位：mb
	ContainerUsedMemoryMB  int64          // 容器使用的内存，单位：mb
	ProcessUsedMemoryMB    int64          // 当前进程所使用的内存
	UsedMemoryRatio        float64        // 已经使用内存的比例
	ContainerCpuUsedRatio  float64        // 容器内cpu使用率，单位：%
	ContainerDiskRead      uint64         // 容器内磁盘读取，单位bytes/sec
	ContainerDiskWrite     uint64         // 容器内磁盘写入，单位bytes/sec
	ContainerNetFlowRX     uint64         // pod内网卡接收流量,单位bytes/sec
	ContainerNetFlowTX     uint64         // pod内网卡输出流量,单位bytes/sec
	Netstat                map[string]int // TCP 连接详情
}

// isRemote 则不获取进程内存
func (res *resourceData) getResource(rs resource.Resource, cpu chan float64, cpuTimes time.Duration, disk chan resource.DiskStat) *resourceData {
	if !rs.InitSuccess() {
		return res
	}
	var err error
	res.ContainerTotalMemoryMB, res.ContainerUsedMemoryMB = utils.GetContainerMemory(rs)
	if res.ContainerTotalMemoryMB == 0 || res.ContainerUsedMemoryMB == 0 {
		return res
	}

	res.ProcessUsedMemoryMB, err = rs.GetRss()
	if err != nil {
		logger.Warnf("获取服务进程内存失败", err)
	}

	used := decimal.NewFromInt(res.ContainerUsedMemoryMB)
	total := decimal.NewFromInt(res.ContainerTotalMemoryMB)
	res.UsedMemoryRatio, _ = used.DivRound(total, 2).Mul(decimal.NewFromInt(100)).Float64()
	// disk
	go utils.GetContainerDisk(rs, disk, time.Second)
	diskChan := <-disk
	res.ContainerDiskRead = diskChan.Read
	res.ContainerDiskWrite = diskChan.Write
	// cpu
	go utils.GetContainerCpu(rs, cpu, cpuTimes)
	res.ContainerCpuUsedRatio, _ = decimal.NewFromFloat(<-cpu).Round(2).Float64()
	return res
}

func (r *resourceData) reset() {
	r.ContainerCpuUsedRatio = 0
	r.ContainerTotalMemoryMB = 0 // 容器总内存，单位：mb
	r.ContainerUsedMemoryMB = 0  // 容器使用的内存，单位：mb
	r.ProcessUsedMemoryMB = 0    // 当前进程所使用的内存
	r.UsedMemoryRatio = 0
	r.ContainerCpuUsedRatio = 0 // 容器内cpu使用率，单位：%
	r.ContainerDiskRead = 0     // 容器内磁盘读取，单位bytes/sec
	r.ContainerDiskWrite = 0    // 容器内磁盘写入，单位bytes/sec
}

// 上报到redis msp
//
type MSPReportResource struct {
	UniqueId           string  `json:"unique_id"`
	AlertType          int     `json:"alert_type"`     // 1注册中心兜底2注册中心自定义3sidecar资源兜底4服务资源兜底
	ResourceType       int     `json:"resource_type"`  // 资源类型1cpu2内存 3所有
	Cpu                float64 `json:"cpu"`            // cpu百分比
	CpuCoreCount       float64 `json:"cpu_core_count"` // cpu核数
	CpuThreshold       float64 `json:"cpu_threshold"`  // cpu阈值
	Mem                float64 `json:"mem"`            // 百分比
	MemoryMB           int64   `json:"memory_mb"`
	MemThreshold       float64 `json:"mem_threshold"` // 内存阈值
	HostName           string  `json:"hostname"`
	ServiceName        string  `json:"service_name"`         // 服务可读名称
	ServiceMemRequests int     `json:"service_mem_requests"` // 服务内存request
	ServiceMemLimits   int     `json:"service_mem_limits"`   // 服务内存limit
	ServiceCPURequests int     `json:"service_cpu_requests"` // 服务cpu request
	ServiceCPULimits   int     `json:"service_cpu_limits"`   // 服务cpu limit
	EventTime          int64   `json:"event_time"`
}

// 写入资源预警日志
func recordResourceWarning(alert MSPReportResource) {
	cpu := utils.FloatToString(alert.Cpu) + "%"
	cpuThreshold := utils.FloatToString(alert.CpuThreshold) + "%"
	mem := utils.FloatToString(alert.Mem) + "%"
	memThreshold := utils.FloatToString(alert.MemThreshold) + "%"

	// 发送告警
	var eventMsg, rule, alertType, resourceType string
	eventMsg = "微服务资源兜底预警"
	alertType = alertTypeService

	if alert.Cpu >= alert.CpuThreshold {
		alert.ResourceType = RESOURCE_CPU_TYPE
		resourceType = resourceTypeCPU
		rule = fmt.Sprintf("当前CPU: %s(核), 达到阈值: %s(核)", cpu, cpuThreshold)
	}
	if alert.Mem >= alert.MemThreshold {
		alert.ResourceType = RESOURCE_MEM_TYPE
		resourceType = resourceTypeMemory
		rule = fmt.Sprintf("当前内存: %s, 达到阈值: %s", mem, memThreshold)
	}
	if alert.Cpu >= alert.CpuThreshold && alert.Mem >= alert.MemThreshold {
		alert.ResourceType = RESOURCE_BOTH_TYPE
		resourceType = resourceTypeBoth
		rule = fmt.Sprintf("当前CPU: %s(核), 达到阈值: %s(核); 当前内存: %s, 达到阈值: %s",
			cpu,
			cpuThreshold,
			mem,
			memThreshold,
		)
	}
	// 日志写入
	textLog := fmt.Sprintf(`发生时间:%s;事件:%s;服务名称:%s;服务识别号:%s;HOSTNAME:%s;触发规则:%s`,
		time.Unix(0, alert.EventTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
		eventMsg,
		alert.ServiceName,
		alert.UniqueId,
		alert.HostName,
		rule,
	)
	logger.With("customLog1", "resource", "customLog2", alertType, "customLog3", resourceType).Error(textLog)
}
