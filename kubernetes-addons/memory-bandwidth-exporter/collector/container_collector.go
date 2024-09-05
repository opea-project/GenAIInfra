package collector

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/opea-project/GenAIInfra/kubernetes-addons/memory-bandwidth-exporter/info"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/clock"
)

const (
	containerCollectorSubsystem = "container"
)

type containerCollector struct {
	statsCache         map[string]stats
	containerInfos     map[string]info.ContainerInfo
	interval           time.Duration
	logger             log.Logger
	namespaceWhiteList []string
	monTimes           int
	metrics            map[string]struct{}
}

func init() {
	registerCollector(containerCollectorSubsystem, defaultEnabled, NewContainerCollector)
}

// NewContainerCollector returns a new Collector exposing container level memory bandwidth metrics.
func NewContainerCollector(logger log.Logger, interval time.Duration) (Collector, error) {
	var ns []string
	if *namespaceWhiteList != "" {
		ns = strings.Split(*namespaceWhiteList, ",")
	}
	c := &containerCollector{
		statsCache:         make(map[string]stats),
		containerInfos:     make(map[string]info.ContainerInfo),
		interval:           interval,
		logger:             logger,
		namespaceWhiteList: ns,
		monTimes:           *monTimes,
		metrics:            make(map[string]struct{}),
	}
	logger.Log("info", "new container collector", "metrics", *containerCollectorMetrics)
	if *containerCollectorMetrics == allMetrics {
		for _, m := range allContainerMetrics {
			c.metrics[m] = struct{}{}
		}
	} else if *containerCollectorMetrics != noMetrics {
		for _, m := range strings.Split(*containerCollectorMetrics, ",") {
			c.metrics[m] = struct{}{}
		}
	}
	c.Start()
	return c, nil
}

func (c *containerCollector) Start() {
	c.logger.Log("info", "start container collector", "metrics", getMetricsKeys(c.metrics))
	go func() {
		for {
			select {
			case cdata, ok := <-info.ContainerInfoChan:
				if !ok {
					c.logger.Log("err", "Channel closed, stopping data processing.")
				}
				err := c.processContainerData(cdata)
				if err != nil {
					c.logger.Log("err", fmt.Sprintf("Cannot process container data: %v", err))
				}
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func (c *containerCollector) processContainerData(data map[string]info.ContainerInfo) error {
	for containerId, containerInfo := range data {
		if len(c.namespaceWhiteList) > 0 && stringInSlice(containerInfo.NameSpace, c.namespaceWhiteList) {
			continue
		}
		if data[containerId].Operation == 0 {
			delete(c.statsCache, containerId)
			delete(c.containerInfos, containerId)
		}
		level.Info(c.logger).Log("msg", "ContainerInfoChan received", "operation", data[containerId].Operation,
			"pod name", data[containerId].PodName, "container id", containerId, "container name",
			data[containerId].ContainerName, "namespace", data[containerId].NameSpace)
		if data[containerId].Operation == 1 || data[containerId].Operation == 2 {
			c.containerInfos[containerId] = containerInfo
			go c.housekeeping(containerId)
		}

		if data[containerId].Operation == 2 && isNeedCollectMbLLc(c.metrics) {
			err := makeMonitorGroup(c.containerInfos[containerId].MonGroupPath)
			if err != nil {
				return fmt.Errorf("failed to create monitor group: %v", err)
			}
			go c.updatePids(containerId)
		}
	}
	return nil
}

func (c *containerCollector) updatePids(containerId string) {
	for i := 0; i < c.monTimes; i++ {
		c.logger.Log("debug", fmt.Sprintf(`Scan for the %v time and update the pids of the container 
		created before the exporter started.`, i+1))
		if _, ok := c.containerInfos[containerId]; !ok {
			return
		}
		if err := writePidsToTasks(c.containerInfos[containerId].MonGroupPath,
			c.containerInfos[containerId].CgroupPath); err != nil {
			c.logger.Log("err", fmt.Sprintf("failed to update container %v stats: %v", containerId, err))
		}
		time.Sleep(jitter(c.interval))
	}
}

func (c *containerCollector) housekeeping(containerId string) {
	clock := clock.RealClock{}
	houseKeepingTimer := clock.NewTimer(c.interval)
	defer houseKeepingTimer.Stop()
	for range houseKeepingTimer.C() {
		_, err := os.Stat(c.containerInfos[containerId].CgroupPath)
		if os.IsNotExist(err) {
			c.logger.Log("info", fmt.Sprintf("container %v cgroup path %v does not exist, deleting cache",
				containerId, c.containerInfos[containerId].CgroupPath))
			delete(c.statsCache, containerId)
			delete(c.containerInfos, containerId)
			return
		}
		if err != nil {
			c.logger.Log("err", fmt.Sprintf("failed to stat cgroup path %v: %v", c.containerInfos[containerId].CgroupPath, err))
			return
		}
		if err := c.updateStats(containerId); err != nil {
			c.logger.Log("err", fmt.Sprintf("failed to update container %v stats: %v", containerId, err))
			return
		}
		houseKeepingTimer.Reset(jitter(c.interval))
	}
}

func (c *containerCollector) updateStats(containerId string) error {
	newStats := RawStats{}
	var err error
	if isNeedCollectMbLLc(c.metrics) {
		newStats.SocketNum, newStats.MemoryBandwidth, newStats.Cache, err =
			getIntelRDTStatsFrom(c.containerInfos[containerId].MonGroupPath)
		if err != nil {
			return err
		}
	}
	if isNeedCollectCpu(c.metrics) {
		cpuUtilization, err := getCPUUtilizationFrom(c.containerInfos[containerId].CgroupPath)
		if err != nil {
			return err
		}
		newStats.CPUUtilization = &cpuUtilization
	}
	if isNeedCollectMemory(c.metrics) {
		newStats.Memory, err = getMemorySizeFrom(c.containerInfos[containerId].CgroupPath)
		if err != nil {
			return err
		}
	}
	if oldStats, ok := c.statsCache[containerId]; ok {
		pStats, err := processStats(oldStats.oldStats, newStats)
		if err != nil {
			return err
		}
		c.statsCache[containerId] = stats{
			oldStats:       newStats,
			processedStats: pStats,
		}
	} else {
		c.statsCache[containerId] = stats{
			oldStats:       newStats,
			processedStats: ProcessedStats{},
		}
	}
	return nil
}

func (c *containerCollector) Update(ch chan<- prometheus.Metric) error {
	if len(c.statsCache) == 0 {
		c.logger.Log("info", "container collector stats have no cache")
		return nil
	}
	for cid, stats := range c.statsCache {
		if isNeedCollectMbLLc(c.metrics) && stats.processedStats.MemoryBandwidth != nil {
			ch <- prometheus.MustNewConstMetric(
				sumTotalMemoryBandwidthDesc,
				prometheus.GaugeValue,
				stats.processedStats.SumMemoryBandwidth.TotalMBps,
				cid,
				c.containerInfos[cid].ContainerName,
				c.containerInfos[cid].PodName,
				c.containerInfos[cid].NameSpace,
			)
			ch <- prometheus.MustNewConstMetric(
				sumLocalMemoryBandwidthDesc,
				prometheus.GaugeValue,
				stats.processedStats.SumMemoryBandwidth.LocalMBps,
				cid,
				c.containerInfos[cid].ContainerName,
				c.containerInfos[cid].PodName,
				c.containerInfos[cid].NameSpace,
			)
			for sid, s := range stats.processedStats.MemoryBandwidth {
				ch <- prometheus.MustNewConstMetric(
					totalMemoryBandwidthDesc,
					prometheus.GaugeValue,
					s.TotalMBps,
					sid,
					cid,
					c.containerInfos[cid].ContainerName,
					c.containerInfos[cid].PodName,
					c.containerInfos[cid].NameSpace,
				)
				ch <- prometheus.MustNewConstMetric(
					localMemoryBandwidthDesc,
					prometheus.GaugeValue,
					s.LocalMBps,
					sid,
					cid,
					c.containerInfos[cid].ContainerName,
					c.containerInfos[cid].PodName,
					c.containerInfos[cid].NameSpace,
				)
			}
		}
		if isNeedCollectMbLLc(c.metrics) && stats.processedStats.Cache != nil {
			ch <- prometheus.MustNewConstMetric(
				sumLLCacheDesc,
				prometheus.GaugeValue,
				stats.processedStats.SumCache.LLCOccupancy,
				cid,
				c.containerInfos[cid].ContainerName,
				c.containerInfos[cid].PodName,
				c.containerInfos[cid].NameSpace,
			)
			for sid, s := range stats.processedStats.Cache {
				ch <- prometheus.MustNewConstMetric(
					llcacheDesc,
					prometheus.GaugeValue,
					s.LLCOccupancy,
					sid,
					cid,
					c.containerInfos[cid].ContainerName,
					c.containerInfos[cid].PodName,
					c.containerInfos[cid].NameSpace,
				)
			}
		}
		if isNeedCollectCpu(c.metrics) {
			ch <- prometheus.MustNewConstMetric(
				cpuUtilizationDesc,
				prometheus.GaugeValue,
				stats.processedStats.CPUUtilization,
				cid,
				c.containerInfos[cid].ContainerName,
				c.containerInfos[cid].PodName,
				c.containerInfos[cid].NameSpace,
			)
		}
		if isNeedCollectMemory(c.metrics) {
			ch <- prometheus.MustNewConstMetric(
				memoryDesc,
				prometheus.GaugeValue,
				stats.processedStats.Memory,
				cid,
				c.containerInfos[cid].ContainerName,
				c.containerInfos[cid].PodName,
				c.containerInfos[cid].NameSpace,
			)
		}
	}
	return nil
}
