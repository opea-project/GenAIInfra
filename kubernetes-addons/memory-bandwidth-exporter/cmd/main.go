package main

import (
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"sort"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/opea-project/GenAIInfra/kubernetes-addons/memory-bandwidth-exporter/collector"
	"github.com/opea-project/GenAIInfra/kubernetes-addons/memory-bandwidth-exporter/plugin"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	disableExporterMetrics = kingpin.Flag(
		"web.disable-exporter-metrics",
		"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
	).Default("true").Bool()
	maxRequests = kingpin.Flag(
		"web.max-requests",
		"Maximum number of parallel scrape requests. Use 0 to disable.",
	).Default("40").Int()
	maxProcs = kingpin.Flag(
		"runtime.gomaxprocs",
		"The target number of CPUs Go will run on (GOMAXPROCS)",
	).Envar("GOMAXPROCS").Default("1").Int()
	toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9100")
	interval     = kingpin.Flag(
		"collector.interval",
		"memory bandwidth exporter collect metrics interval",
	).Default("3s").Duration()
	pluginName = kingpin.Flag(
		"NRIplugin.name",
		"Plugin name to register to NRI",
	).Default("mb-nri-plugin").String()
	pluginIdx = kingpin.Flag(
		"NRIplugin.idx",
		"Plugin index to register to NRI",
	).Default("11").String()
	disableWatch = kingpin.Flag(
		"disableWatch",
		"Disable watching hook directories for new hooks",
	).Default("false").Bool()
)

type handler struct {
	unfilteredHandler http.Handler
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	maxRequests             int
	logger                  log.Logger
}

func newHandler(includeExporterMetrics bool, maxRequests int, logger log.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		maxRequests:             maxRequests,
		logger:                  logger,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}
	if innerHandler, err := h.innerHandler(); err != nil {
		panic(fmt.Sprintf("Couldn't create metrics handler: %s", err))
	} else {
		h.unfilteredHandler = innerHandler
	}
	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	level.Debug(h.logger).Log("msg", "collect query:", "filters", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}
	// To serve filtered metrics, we create a filtering handler on the fly.
	filteredHandler, err := h.innerHandler(filters...)
	if err != nil {
		level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}
	filteredHandler.ServeHTTP(w, r)
}

// innerHandler is used to create both the one unfiltered http.Handler to be
// wrapped by the outer handler and also the filtered handlers created on the
// fly. The former is accomplished by calling innerHandler without any arguments
// (in which case it will log all the collectors enabled via command-line
// flags).
func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	level.Info(h.logger).Log("msg", "filters", filters)
	nc, err := collector.NewCollector(h.logger, *interval, filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %s", err)
	}

	// Only log the creation of an unfiltered handler, which should happen
	// only once upon startup.
	if len(filters) == 0 {
		level.Info(h.logger).Log("msg", "Enabled collectors")
		collectors := []string{}
		for n := range nc.Collectors {
			collectors = append(collectors, n)
		}
		sort.Strings(collectors)
		for _, c := range collectors {
			level.Info(h.logger).Log("collector", c)
		}
	}

	r := prometheus.NewRegistry()
	r.MustRegister(versioncollector.NewCollector("memory_bandwidth_exporter"))
	if err := r.Register(nc); err != nil {
		return nil, fmt.Errorf("couldn't register collector: %s", err)
	}

	var handler http.Handler
	if h.includeExporterMetrics {
		handler = promhttp.HandlerFor(
			prometheus.Gatherers{h.exporterMetricsRegistry, r},
			promhttp.HandlerOpts{
				ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
				ErrorHandling:       promhttp.ContinueOnError,
				MaxRequestsInFlight: h.maxRequests,
				Registry:            h.exporterMetricsRegistry,
			},
		)
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	} else {
		handler = promhttp.HandlerFor(
			r,
			promhttp.HandlerOpts{
				ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
				ErrorHandling:       promhttp.ContinueOnError,
				MaxRequestsInFlight: h.maxRequests,
			},
		)
	}

	return handler, nil
}

// The memory bandwidth exporter currently only supports Intel platforms with RDT, and will fail on other platforms.
// We will add node feature discovery in the future.
func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	version.Version = "v0.1.0"
	kingpin.Version(version.Print("memory bandwidth exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting memory bandwidth exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	if user, err := user.Current(); err == nil && user.Uid == "0" {
		level.Warn(logger).Log("msg", `Memory bandwidth exporter is running as root user. This 
		exporter is designed to run as unprivileged user, root is not required.`)
	}
	runtime.GOMAXPROCS(*maxProcs)
	level.Info(logger).Log("msg", "Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))
	isNeedMakeMonitorGroup := collector.ParseCollectorMetrics()
	nriPlugin := &plugin.Plugin{
		PluginName:   *pluginName,
		PluginIdx:    *pluginIdx,
		DisableWatch: *disableWatch,
		Logger:       logger,
	}
	level.Info(logger).Log("msg", "Starting NRI plugin")
	errChan := make(chan error)
	go func() {
		errChan <- nriPlugin.Run(isNeedMakeMonitorGroup)
	}()
	go func() {
		if err := <-errChan; err != nil {
			level.Error(logger).Log("Failed to run nriPlugin", "error", err)
		}
	}()

	http.Handle(*metricsPath, newHandler(!*disableExporterMetrics, *maxRequests, logger))
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        "Memory Bandwidth Exporter",
			Description: "Prometheus Memory Bandwidth Exporter",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
