package collector

import (
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	nodeCollectorSubsystem   = "node"
	socketCollectorSubsystem = "socket"
)

type nodeCollctor struct {
	interval     time.Duration
	logger       log.Logger
	nodeName     string
	statsCache   *stats
	monGroupPath string
	metrics      map[string]struct{}
}

func init() {
	registerCollector(nodeCollectorSubsystem, defaultDisabled, NewNodeCollector)
}

// NewNodeCollector returns a new Collector exposing node level memory bandwidth metrics.
func NewNodeCollector(logger log.Logger, interval time.Duration) (Collector, error) {
	c := &nodeCollctor{
		interval:     interval,
		logger:       logger,
		monGroupPath: rootResctrlPath,
		nodeName:     *nodeName,
		metrics:      make(map[string]struct{}),
	}
	logger.Log("info", "new node collector", "metrics:", *nodeCollectorMetrics)
	if *nodeCollectorMetrics == allMetrics {
		for _, m := range allNodeMetrics {
			c.metrics[m] = struct{}{}
		}
	} else if *nodeCollectorMetrics != noMetrics {
		for _, m := range strings.Split(*nodeCollectorMetrics, ",") {
			c.metrics[m] = struct{}{}
		}
	}
	c.Start()
	return c, nil
}

func (c *nodeCollctor) Start() {
	c.logger.Log("info", "start node collector", "metrics", getMetricsKeys(c.metrics))
	go func() {
		for {
			err := c.updateStats()
			if err != nil {
				c.logger.Log("error", "node collector update stats failed", "err", err)
			}
			time.Sleep(jitter(c.interval))
		}
	}()
}

func (c *nodeCollctor) updateStats() error {
	newStats := RawStats{}
	var err error
	if isNeedCollectMbLLc(c.metrics) {
		newStats.SocketNum, newStats.MemoryBandwidth, newStats.Cache, err = getIntelRDTStatsFrom(c.monGroupPath)
		if err != nil {
			return err
		}
	}
	if c.statsCache != nil {
		pStats, err := processStats(c.statsCache.oldStats, newStats)
		if err != nil {
			return err
		}
		c.statsCache = &stats{
			oldStats:       newStats,
			processedStats: pStats,
		}
	} else {
		c.statsCache = &stats{
			oldStats:       newStats,
			processedStats: ProcessedStats{},
		}
	}
	return nil
}

func (c *nodeCollctor) Update(ch chan<- prometheus.Metric) error {
	if c.statsCache == nil {
		c.logger.Log("info", "node collector stats have no cache")
		return nil
	}
	if !isNeedCollectMbLLc(c.metrics) {
		return nil
	}
	ch <- prometheus.MustNewConstMetric(
		nodeTotalMemoryBandwidthDesc,
		prometheus.GaugeValue,
		c.statsCache.processedStats.SumMemoryBandwidth.TotalMBps,
		c.nodeName,
	)
	ch <- prometheus.MustNewConstMetric(
		nodeLocalMemoryBandwidthDesc,
		prometheus.GaugeValue,
		c.statsCache.processedStats.SumMemoryBandwidth.LocalMBps,
		c.nodeName,
	)
	ch <- prometheus.MustNewConstMetric(
		nodeLLCacheDesc,
		prometheus.GaugeValue,
		c.statsCache.processedStats.SumCache.LLCOccupancy,
		c.nodeName,
	)
	for socket, stats := range c.statsCache.processedStats.Cache {
		ch <- prometheus.MustNewConstMetric(
			socketLLCacheDesc,
			prometheus.GaugeValue,
			stats.LLCOccupancy,
			socket,
			c.nodeName,
		)
	}
	for socket, stats := range c.statsCache.processedStats.MemoryBandwidth {
		ch <- prometheus.MustNewConstMetric(
			socketTotalMemoryBandwidthDesc,
			prometheus.GaugeValue,
			stats.TotalMBps,
			socket,
			c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			socketLocalMemoryBandwidthDesc,
			prometheus.GaugeValue,
			stats.LocalMBps,
			socket,
			c.nodeName,
		)
	}

	return nil
}
