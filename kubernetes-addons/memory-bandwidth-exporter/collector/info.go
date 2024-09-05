package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// container
	sumTotalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "sum_total_memory_bandwidth"),
		"The sum of total memory bandwidth for all sockets in MBps.",
		[]string{"containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	sumLocalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "sum_local_memory_bandwidth"),
		"The sum of local memory bandwidth for all sockets in MBps.",
		[]string{"containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	totalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "total_memory_bandwidth"),
		"One socket total memory bandwidth in MBps.",
		[]string{"socketId", "containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	localMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "local_memory_bandwidth"),
		"One socket local memory bandwidth in MBps.",
		[]string{"socketId", "containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	sumLLCacheDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "sum_llc_occupancy"),
		"The sum of llc occupancy for all sockets in MiB.",
		[]string{"containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	llcacheDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "llc_occupancy"),
		"One socket llc occupancy in MiB.",
		[]string{"socketId", "containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	cpuUtilizationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "cpu_utilization"),
		"The CPU utilization of the container refers to the number of CPUs it uses.",
		[]string{"containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	memoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, containerCollectorSubsystem, "memory"),
		"The memory usage of the container in MiB.",
		[]string{"containerId", "containerName", "podName", "nameSpace"}, nil,
	)
	//node
	nodeTotalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "total_memory_bandwidth"),
		"The sum of total memory bandwidth for all sockets in MBps.",
		[]string{"nodeName"}, nil,
	)
	nodeLocalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "local_memory_bandwidth"),
		"The sum of local memory bandwidth for all sockets in MBps.",
		[]string{"nodeName"}, nil,
	)
	socketTotalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, socketCollectorSubsystem, "total_memory_bandwidth"),
		"One socket total memory bandwidth in MBps.",
		[]string{"socketId", "nodeName"}, nil,
	)
	socketLocalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, socketCollectorSubsystem, "local_memory_bandwidth"),
		"One socket local memory bandwidth in MBps.",
		[]string{"socketId", "nodeName"}, nil,
	)
	nodeLLCacheDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "llc_occupancy"),
		"The sum of llc occupancy for all sockets in MiB.",
		[]string{"nodeName"}, nil,
	)
	socketLLCacheDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, socketCollectorSubsystem, "llc_occupancy"),
		"One socket llc occupancy in MiB.",
		[]string{"socketId", "nodeName"}, nil,
	)
	// class
	classTotalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, classCollectorSubsystem, "total_memory_bandwidth"),
		"The sum of total memory bandwidth for all sockets in MBps.",
		[]string{"className", "nodeName"}, nil,
	)
	classLocalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, classCollectorSubsystem, "local_memory_bandwidth"),
		"The sum of local memory bandwidth for all sockets in MBps.",
		[]string{"className", "nodeName"}, nil,
	)
	classLLCacheDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, classCollectorSubsystem, "llc_occupancy"),
		"The sum of llc occupancy for all sockets in MiB.",
		[]string{"className", "nodeName"}, nil,
	)
	socketClassTotalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, classCollectorSubsystem, "socket_total_memory_bandwidth"),
		"One socket total memory bandwidth in MBps.",
		[]string{"socketId", "className", "nodeName"}, nil,
	)
	socketClassLocalMemoryBandwidthDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, classCollectorSubsystem, "socket_local_memory_bandwidth"),
		"One socket local memory bandwidth in MBps.",
		[]string{"socketId", "className", "nodeName"}, nil,
	)
	socketClassLLCacheDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, classCollectorSubsystem, "socket_llc_occupancy"),
		"One socket llc occupancy in MiB.",
		[]string{"socketId", "className", "nodeName"}, nil,
	)
)

type stats struct {
	oldStats       RawStats
	processedStats ProcessedStats
}

type ProcessedStats struct {
	socketNum          int
	SumMemoryBandwidth ProcessedMemoryBandwidthStats
	SumCache           ProcessedCacheStats
	MemoryBandwidth    map[string]ProcessedMemoryBandwidthStats
	Cache              map[string]ProcessedCacheStats
	CPUUtilization     float64 // cpu nums, not %
	Memory             float64 // MiB
}

type RawStats struct {
	SocketNum       int
	MemoryBandwidth map[string]RawMemoryBandwidthStats
	Cache           map[string]RawCacheStats
	CPUUtilization  *RawCPUStats
	Memory          int64 // bytes
}

type RawCPUStats struct {
	CPU       int64 // microseconds
	TimeStamp string
}

type ProcessedMemoryBandwidthStats struct {
	// The 'mbm_total_bytes' to MBps
	TotalMBps float64
	// The 'mbm_local_bytes'. to MBps
	LocalMBps float64
}

// MemoryBandwidthStats corresponds to MBM (Memory Bandwidth Monitoring).
type RawMemoryBandwidthStats struct {
	// The 'mbm_total_bytes'
	TotalBytes          uint64
	TotalBytesTimeStamp string
	// The 'mbm_local_bytes'.
	LocalBytes          uint64
	LocalBytesTimeStamp string
}

type RawCacheStats struct {
	// The 'llc_occupancy'
	LLCOccupancy uint64
}

type ProcessedCacheStats struct {
	// The 'llc_occupancy' to MiB
	LLCOccupancy float64
}
