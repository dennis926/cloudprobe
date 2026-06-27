package collector

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Metrics 采集的系统指标
type Metrics struct {
	Timestamp     int64                  `json:"timestamp"`
	Hostname      string                 `json:"hostname"`
	OS            string                 `json:"os"`
	Platform      string                 `json:"platform"`
	CPU           CPUInfo                `json:"cpu"`
	Memory        MemoryInfo             `json:"memory"`
	Disk          []DiskInfo             `json:"disk"`
	Network       NetworkInfo            `json:"network"`
	Load          LoadInfo               `json:"load"`
	Uptime        uint64                 `json:"uptime"`
	BootTime      uint64                 `json:"boot_time"`
	ProcessCount  int                    `json:"process_count"`
}

// CPUInfo CPU信息
type CPUInfo struct {
	Percent     []float64 `json:"percent"`
	LogicalCnt  int       `json:"logical_count"`
	PhysicalCnt int       `json:"physical_count"`
	ModelName   string    `json:"model_name"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// LoadInfo 负载信息
type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// Collector 指标采集器
type Collector struct {
	hostname     string
	os           string
	platform     string
	cpuModel     string
	cpuLogical   int
	cpuPhysical  int
	lastNetIO    *net.IOCountersStat
	lastNetTime  time.Time
}

// NewCollector 创建采集器
func NewCollector() (*Collector, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("get host info failed: %w", err)
	}

	cpuInfos, err := cpu.Info()
	if err != nil && len(cpuInfos) > 0 {
		cpuInfos = []cpu.InfoStat{{ModelName: "unknown"}}
	}

	logicalCnt, _ := cpu.Counts(true)
	physicalCnt, _ := cpu.Counts(false)

	modelName := "unknown"
	if len(cpuInfos) > 0 {
		modelName = cpuInfos[0].ModelName
	}

	return &Collector{
		hostname:    hostInfo.Hostname,
		os:          runtime.GOOS,
		platform:    hostInfo.Platform,
		cpuModel:    modelName,
		cpuLogical:  logicalCnt,
		cpuPhysical: physicalCnt,
	}, nil
}

// Collect 采集所有指标
func (c *Collector) Collect() (*Metrics, error) {
	metrics := &Metrics{
		Timestamp: time.Now().Unix(),
		Hostname:  c.hostname,
		OS:        c.os,
		Platform:  c.platform,
	}

	// CPU
	cpuPercents, err := cpu.Percent(0, true)
	if err == nil {
		metrics.CPU = CPUInfo{
			Percent:     cpuPercents,
			LogicalCnt:  c.cpuLogical,
			PhysicalCnt: c.cpuPhysical,
			ModelName:   c.cpuModel,
		}
	}

	// Memory
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.Memory = MemoryInfo{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
		}
	}

	// Disk
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, part := range partitions {
			usage, err := disk.Usage(part.Mountpoint)
			if err != nil {
				continue
			}
			metrics.Disk = append(metrics.Disk, DiskInfo{
				Path:        part.Mountpoint,
				Total:       usage.Total,
				Used:        usage.Used,
				Free:        usage.Free,
				UsedPercent: usage.UsedPercent,
			})
		}
	}

	// Load
	loadInfo, err := load.Avg()
	if err == nil {
		metrics.Load = LoadInfo{
			Load1:  loadInfo.Load1,
			Load5:  loadInfo.Load5,
			Load15: loadInfo.Load15,
		}
	}

	// Network
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		now := time.Now()
		if c.lastNetIO != nil {
			duration := now.Sub(c.lastNetTime).Seconds()
			if duration > 0 {
				metrics.Network = NetworkInfo{
					BytesSent:   netIO[0].BytesSent - c.lastNetIO.BytesSent,
					BytesRecv:   netIO[0].BytesRecv - c.lastNetIO.BytesRecv,
					PacketsSent: netIO[0].PacketsSent - c.lastNetIO.PacketsSent,
					PacketsRecv: netIO[0].PacketsRecv - c.lastNetIO.PacketsRecv,
				}
			}
		}
		c.lastNetIO = &netIO[0]
		c.lastNetTime = now
	}

	// Host
	hostInfo, err := host.Info()
	if err == nil {
		metrics.Uptime = hostInfo.Uptime
		metrics.BootTime = hostInfo.BootTime
		metrics.ProcessCount = int(hostInfo.Procs)
	}

	return metrics, nil
}
