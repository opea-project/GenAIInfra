package collector

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "rdt"

var (
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"memory_bandwidth_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

const (
	defaultEnabled  = true
	defaultDisabled = false
)

var (
	factories              = make(map[string]func(logger log.Logger, interval time.Duration) (Collector, error))
	initiatedCollectorsMtx = sync.Mutex{}
	initiatedCollectors    = make(map[string]Collector)
	collectorState         = make(map[string]*bool)
	collectorMetrics       = make(map[string]string)
)

func registerCollector(collector string, isDefaultEnabled bool, factory func(
	logger log.Logger, interval time.Duration) (Collector, error)) {
	collectorState[collector] = &isDefaultEnabled
	factories[collector] = factory
}

func ParseCollectorMetrics() bool {
	isNeedNRIPlugin := false
	for collector := range collectorState {
		if collector == containerCollectorSubsystem {
			var isDefaultEnabled bool
			if containerCollectorMetrics == nil || *containerCollectorMetrics == noMetrics {
				isDefaultEnabled = false
				collectorMetrics[collector] = noMetrics
			} else {
				isDefaultEnabled = true
				collectorMetrics[collector] = *containerCollectorMetrics
			}
			collectorState[collector] = &isDefaultEnabled
			if containerCollectorMetrics != nil && (*containerCollectorMetrics == allMetrics ||
				strings.Contains(*containerCollectorMetrics, "mb") ||
				strings.Contains(*containerCollectorMetrics, "llc")) {
				isNeedNRIPlugin = true
			}
		}
		if collector == classCollectorSubsystem {
			var isDefaultEnabled bool
			if classCollectorMetrics == nil || *classCollectorMetrics == noMetrics {
				isDefaultEnabled = false
				collectorMetrics[collector] = noMetrics
			} else {
				isDefaultEnabled = true
				collectorMetrics[collector] = *classCollectorMetrics
			}
			collectorState[collector] = &isDefaultEnabled
		}
		if collector == nodeCollectorSubsystem {
			var isDefaultEnabled bool
			if nodeCollectorMetrics == nil || *nodeCollectorMetrics == noMetrics {
				isDefaultEnabled = false
				collectorMetrics[collector] = noMetrics
			} else {
				isDefaultEnabled = true
				collectorMetrics[collector] = *nodeCollectorMetrics
			}
			collectorState[collector] = &isDefaultEnabled
		}
	}
	return isNeedNRIPlugin
}

// Collector implements the prometheus.Collector interface.
type MBCollector struct {
	Collectors map[string]Collector
	logger     log.Logger
}

// NewCollector creates a new Collector.
func NewCollector(logger log.Logger, interval time.Duration, filters ...string) (*MBCollector, error) {
	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()
	for key, enabled := range collectorState {
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := factories[key](log.With(logger, "collector", key), interval)
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
	}
	return &MBCollector{Collectors: collectors, logger: logger}, nil
}

// Describe implements the prometheus.Collector interface.
func (n MBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (n MBCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch, n.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger log.Logger) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			level.Debug(logger).Log("msg", "collector returned no data", "name", name, "duration_seconds", duration.Seconds(),
				"err", err)
		} else {
			level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		level.Debug(logger).Log("msg", "collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}
