package collector

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	classCollectorSubsystem = "rdtClass"
)

type classCollector struct {
	statsCache map[string]*stats
	interval   time.Duration
	logger     log.Logger
	nodeName   string
	metrics    map[string]struct{}
}

func init() {
	registerCollector(classCollectorSubsystem, defaultDisabled, NewClassCollector)
}

// NewClassCollector returns a new Collector exposing class level memory bandwidth metrics.
func NewClassCollector(logger log.Logger, interval time.Duration) (Collector, error) {
	c := &classCollector{
		statsCache: make(map[string]*stats),
		interval:   interval,
		logger:     logger,
		nodeName:   *nodeName,
		metrics:    make(map[string]struct{}),
	}
	logger.Log("info", "new class collector", "metrics", *classCollectorMetrics)
	if *classCollectorMetrics == allMetrics {
		for _, m := range allClassMetrics {
			c.metrics[m] = struct{}{}
		}
	} else if *classCollectorMetrics != noMetrics {
		for _, m := range strings.Split(*classCollectorMetrics, ",") {
			c.metrics[m] = struct{}{}
		}
	}
	c.Start()
	return c, nil
}

func (c *classCollector) Start() {
	c.logger.Log("info", "start class collector", "metrics", getMetricsKeys(c.metrics))
	if isNeedCollectMbLLc(c.metrics) {
		go func() {
			for {
				err := c.updateClasses()
				if err != nil {
					c.logger.Log("error", "class collector update classes failed", "err", err)
				}
				time.Sleep(jitter(c.interval))
			}
		}()
	}
	go func() {
		for {
			err := c.updateStats()
			if err != nil {
				c.logger.Log("error", "class collector update stats failed", "err", err)
			}
			time.Sleep(jitter(c.interval))
		}
	}()
}

func (c *classCollector) updateClasses() error {
	excludeDirs := map[string]bool{
		"info":       true,
		"mon_data":   true,
		"mon_groups": true,
	}
	files, err := os.ReadDir(rootResctrlPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			dirName := file.Name()
			_, ok := c.statsCache[dirName]
			if !excludeDirs[dirName] && !ok {
				c.statsCache[dirName] = nil
			}
		}
	}
	return nil
}

func (c *classCollector) updateStats() error {
	for class := range c.statsCache {
		newStats := RawStats{}
		var err error
		if isNeedCollectMbLLc(c.metrics) {
			newStats.SocketNum, newStats.MemoryBandwidth, newStats.Cache, err =
				getIntelRDTStatsFrom(filepath.Join(rootResctrlPath, class))
			if err != nil {
				return err
			}
		}
		if c.statsCache[class] != nil {
			pStats, err := processStats(c.statsCache[class].oldStats, newStats)
			if err != nil {
				return err
			}
			c.statsCache[class] = &stats{
				oldStats:       newStats,
				processedStats: pStats,
			}
		} else {
			c.statsCache[class] = &stats{
				oldStats:       newStats,
				processedStats: ProcessedStats{},
			}
		}
	}
	return nil
}

func (c *classCollector) Update(ch chan<- prometheus.Metric) error {
	if len(c.statsCache) == 0 {
		c.logger.Log("info", "class collector stats have no cache")
		return nil
	}
	if !isNeedCollectMbLLc(c.metrics) {
		return nil
	}
	for cid, stats := range c.statsCache {
		ch <- prometheus.MustNewConstMetric(
			classTotalMemoryBandwidthDesc,
			prometheus.GaugeValue,
			stats.processedStats.SumMemoryBandwidth.TotalMBps,
			cid,
			c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			classLocalMemoryBandwidthDesc,
			prometheus.GaugeValue,
			stats.processedStats.SumMemoryBandwidth.LocalMBps,
			cid,
			c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			classLLCacheDesc,
			prometheus.GaugeValue,
			stats.processedStats.SumCache.LLCOccupancy,
			cid,
			c.nodeName,
		)
		for sid, s := range stats.processedStats.MemoryBandwidth {
			ch <- prometheus.MustNewConstMetric(
				socketClassTotalMemoryBandwidthDesc,
				prometheus.GaugeValue,
				s.TotalMBps,
				sid,
				cid,
				c.nodeName,
			)
			ch <- prometheus.MustNewConstMetric(
				socketClassLocalMemoryBandwidthDesc,
				prometheus.GaugeValue,
				s.LocalMBps,
				sid,
				cid,
				c.nodeName,
			)
		}
		for sid, s := range stats.processedStats.Cache {
			ch <- prometheus.MustNewConstMetric(
				socketClassLLCacheDesc,
				prometheus.GaugeValue,
				s.LLCOccupancy,
				sid,
				cid,
				c.nodeName,
			)
		}
	}
	return nil
}
