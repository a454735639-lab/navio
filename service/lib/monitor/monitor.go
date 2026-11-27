package monitor

import (
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type CPUInfo struct {
	CoreCount int32     `json:"coreCount"`
	CPUNum    int       `json:"cpuNum"`
	Model     string    `json:"model"`
	Usages    []float64 `json:"usages"`
}

type DiskInfo struct {
	Mountpoint  string  `json:"mountpoint"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type NetIOCountersInfo struct {
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
	Name      string `json:"name"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

// 获取CPU信息
func GetCPUInfo() (CPUInfo, error) {
	cpuInfoRes := CPUInfo{}
	cpuInfo, err := cpu.Info()

	if err == nil && len(cpuInfo) > 0 {
		cpuInfoRes.CoreCount = cpuInfo[0].Cores
		cpuInfoRes.Model = cpuInfo[0].ModelName
	}
	numCPU, _ := cpu.Counts(true)
	cpuInfoRes.CPUNum = numCPU
	cpuPercentages, err := cpu.Percent(time.Second, true)
	cpuInfoRes.Usages = cpuPercentages

	return cpuInfoRes, err
}

// 获取内存信息 单位：MB
func GetMemoryInfo() (MemoryInfo, error) {
	memoryInfo := MemoryInfo{}
	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		memoryInfo.Free = memInfo.Free
		memoryInfo.Total = memInfo.Total
		memoryInfo.Used = memInfo.Used
		memoryInfo.UsedPercent = memInfo.UsedPercent
	}

	return memoryInfo, err
}

// 获取每个磁盘分区使用情况
func GetDiskInfo() ([]DiskInfo, error) {
	disks := []DiskInfo{}
	// 获取所有磁盘分区的信息
	partitions, err := disk.Partitions(true)
	if err != nil {
		return disks, err
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// fmt.Printf("Error getting disk usage for %s: %v\n", partition.Mountpoint, err)
			continue
		}

		disks = append(disks, DiskInfo{
			Mountpoint:  partition.Mountpoint,
			Total:       usage.Total / 1024 / 1024,
			Used:        usage.Used / 1024 / 1024,
			Free:        usage.Free / 1024 / 1024,
			UsedPercent: usage.UsedPercent,
		})
	}

	return disks, nil
}

func GetDiskMountpoints() ([]disk.PartitionStat, error) {
	return disk.Partitions(true)
}

func GetDiskInfoByPath(path string) (*DiskInfo, error) {
	diskInfo := DiskInfo{}
	usage, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}
	diskInfo.Free = usage.Free
	diskInfo.Mountpoint = usage.Path
	diskInfo.Total = usage.Total
	diskInfo.Used = usage.Used
	diskInfo.UsedPercent = usage.UsedPercent
	return &diskInfo, nil
}

// 获取网络统计信息
func GetNetIOCountersInfo() ([]NetIOCountersInfo, error) {
	netInfo := []NetIOCountersInfo{}
	netStats, err := net.IOCounters(true)
	if err == nil {
		for _, netStat := range netStats {
			netInfo = append(netInfo, NetIOCountersInfo{
				BytesRecv: netStat.BytesRecv,
				BytesSent: netStat.BytesSent,
				Name:      netStat.Name,
			})

		}
	}
	return netInfo, err
}

// NetworkInfo 全局网络流量（聚合所有物理网卡，排除 lo 回环）
type NetworkInfo struct {
	RxBytes uint64 `json:"rxBytes"` // 累计接收字节
	TxBytes uint64 `json:"txBytes"` // 累计发送字节
}

// HostNetDevPath 容器内读取宿主机网络计数的路径。
// compose 里把宿主机 /proc/net/dev 只读挂载到这里（- /proc/net/dev:/host-net/dev:ro）。
// 文件不存在（本地裸跑/未挂载）时自动回退到容器自身 netns，功能不退化。
const HostNetDevPath = "/host-net/dev"

// isVirtualNetIF 判断是否为不应计入主机总流量的虚拟/回环接口。
// Docker 网桥(bridge/veth/docker0/br-*)会把同一份流量在多张虚拟网卡上重复计数，
// 必须排除，否则主机汇总被成倍虚高。
func isVirtualNetIF(name string) bool {
	lower := strings.ToLower(name)
	if lower == "lo" || lower == "lo0" {
		return true
	}
	virtualPrefixes := []string{"veth", "br-", "docker", "virbr", "vif", "tap", "tun", "dummy", "kube", "cali", "flannel", "nodelocaldns"}
	for _, p := range virtualPrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// GetNetworkInfo 获取全局网络流量累计值。
// 优先读宿主机 /proc/net/dev（真实物理网卡，排除回环与 Docker 虚拟接口），
// 避免容器 netns 只能看到自己几 MB 的流量、或网桥重复计数。
func GetNetworkInfo() (NetworkInfo, error) {
	var (
		netStats []net.IOCountersStat
		err      error
	)
	if _, statErr := os.Stat(HostNetDevPath); statErr == nil {
		netStats, err = net.IOCountersByFile(true, HostNetDevPath)
	} else {
		netStats, err = net.IOCounters(true)
	}
	if err != nil {
		return NetworkInfo{}, err
	}
	var info NetworkInfo
	for _, v := range netStats {
		if isVirtualNetIF(v.Name) {
			continue
		}
		info.RxBytes += v.BytesRecv
		info.TxBytes += v.BytesSent
	}
	return info, nil
}

// func GetCountDiskInfo() {
// 	// 获取所有磁盘的总使用情况
// 	allUsage, err := disk.Usage("/")
// 	if err != nil {
// 		fmt.Printf("Error getting total disk usage: %v\n", err)
// 		return
// 	}

// 	// 打印所有磁盘的总使用情况
// 	fmt.Println("Total Disk Usage:")
// 	fmt.Printf("Total: %d MB\n", allUsage.Total/1024/1024)
// 	fmt.Printf("Used: %d MB\n", allUsage.Used/1024/1024)
// 	fmt.Printf("Free: %d MB\n", allUsage.Free/1024/1024)
// 	fmt.Printf("Usage: %.2f%%\n", allUsage.UsedPercent)
// }
