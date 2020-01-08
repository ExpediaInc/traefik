package metrics

import (
	"github.com/go-kit/kit/metrics"
	"os"
	"strings"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/statsd"
	"github.com/patrickmn/go-cache"
)

var statsdClient = statsd.New(getMetricsPrefix(), kitlog.LoggerFunc(func(keyvals ...interface{}) error {
	log.Info(keyvals)
	return nil
}))

var statsdTicker *time.Ticker
var statsdDynamicCache = cache.New(cache.NoExpiration, cache.NoExpiration)

const (
	statsdMetricsBackendReqsName      = "backend.request.total"
	statsdMetricsBackendLatencyName   = "backend.request.duration"
	statsdRetriesTotalName            = "backend.retries.total"
	statsdConfigReloadsName           = "config.reload.total"
	statsdConfigReloadsFailureName    = statsdConfigReloadsName + ".failure"
	statsdLastConfigReloadSuccessName = "config.reload.lastSuccessTimestamp"
	statsdLastConfigReloadFailureName = "config.reload.lastFailureTimestamp"
	statsdEntrypointReqsName          = "entrypoint.request.total"
	statsdEntrypointReqDurationName   = "entrypoint.request.duration"
	statsdEntrypointOpenConnsName     = "entrypoint.connections.open"
	statsdOpenConnsName               = "backend.connections.open"
	statsdServerUpName                = "backend.server.up"
)

// RegisterStatsd registers the metrics pusher if this didn't happen yet and creates a statsd Registry instance.
func RegisterStatsd(config *types.Statsd) Registry {
	if statsdTicker == nil {
		statsdTicker = initStatsdTicker(config)
	}

	return &standardRegistry{
		enabled:                        true,
		configReloadsCounter:           statsdClient.NewCounter(statsdConfigReloadsName, 1.0),
		configReloadsFailureCounter:    statsdClient.NewCounter(statsdConfigReloadsFailureName, 1.0),
		lastConfigReloadSuccessGauge:   statsdClient.NewGauge(statsdLastConfigReloadSuccessName),
		lastConfigReloadFailureGauge:   statsdClient.NewGauge(statsdLastConfigReloadFailureName),
		entrypointReqsCounter:          statsdClient.NewCounter(statsdEntrypointReqsName, 1.0),
		entrypointReqDurationHistogram: statsdClient.NewTiming(statsdEntrypointReqDurationName, 1.0),
		entrypointOpenConnsGauge:       statsdClient.NewGauge(statsdEntrypointOpenConnsName),
		backendReqsCounter:             statsdClient.NewCounter(statsdMetricsBackendReqsName, 1.0),
		backendReqDurationHistogram:    statsdClient.NewTiming(statsdMetricsBackendLatencyName, 1.0),
		backendRetriesCounter:          statsdClient.NewCounter(statsdRetriesTotalName, 1.0),
		backendOpenConnsGauge:          statsdClient.NewGauge(statsdOpenConnsName),
		backendServerUpGauge:           statsdClient.NewGauge(statsdServerUpName),
	}
}

// initStatsdTicker initializes metrics pusher and creates a statsdClient if not created already
func initStatsdTicker(config *types.Statsd) *time.Ticker {
	address := config.Address
	if len(address) == 0 {
		address = "localhost:8125"
	}
	pushInterval, err := time.ParseDuration(config.PushInterval)
	if err != nil {
		log.Warnf("Unable to parse %s into pushInterval, using 10s as default value", config.PushInterval)
		pushInterval = 10 * time.Second
	}

	report := time.NewTicker(pushInterval)

	safe.Go(func() {
		statsdClient.SendLoop(report.C, "udp", address)
	})

	return report
}

// StopStatsd stops internal statsdTicker which controls the pushing of metrics to StatsD Agent and resets it to `nil`
func StopStatsd() {
	if statsdTicker != nil {
		statsdTicker.Stop()
	}
	statsdTicker = nil
}

func GetAppName() string {
	appName := os.Getenv("APP_NAME")
	if appName == ""  {
		os.Setenv("APP_NAME",  "test-proxy")
		return os.Getenv("APP_NAME")
	}
	return appName
}

func GetHostName() string {
	hostName, err := os.Hostname()
	if err != nil  {
		return "no-host-available"
	}
	return hostName
}

func getMetricsPrefix() string {
	var sb strings.Builder
	sb.WriteString("traefik.")
	sb.WriteString(GetAppName())
	sb.WriteString(".")
	sb.WriteString(GetHostName())
	sb.WriteString(".")
	return sb.String()
}

func (r *standardRegistry) BackendReqsCounterWithLabel(labelValues []string) metrics.Counter {
	counterDynamicName := statsdMetricsBackendReqsName + strings.Join(labelValues, "." )
	if counter, found := statsdDynamicCache.Get(counterDynamicName); found {
		return counter.(metrics.Counter)
	}
	dynamicCounter := statsdClient.NewCounter(counterDynamicName , 1.0)
	statsdDynamicCache.Set(counterDynamicName, dynamicCounter, cache.NoExpiration)
	return dynamicCounter
}

func (r *standardRegistry) BackendReqDurationHistogramWithLabel(labelValues []string) metrics.Histogram {
	histogramDynamicName := statsdMetricsBackendLatencyName + strings.Join(labelValues, "." )
	if histogram, found := statsdDynamicCache.Get(histogramDynamicName); found {
		return histogram.(metrics.Histogram)
	}
	dynamicHistogram := statsdClient.NewTiming(histogramDynamicName , 1.0)
	statsdDynamicCache.Set(histogramDynamicName, dynamicHistogram, cache.NoExpiration)
	return dynamicHistogram
}

func (r *standardRegistry) BackendOpenConnsGaugeWithLabel(labelValues []string) metrics.Gauge {
	openConnsGaugeDynamicName := statsdOpenConnsName + strings.Join(labelValues, "." )
	if gauge, found := statsdDynamicCache.Get(openConnsGaugeDynamicName); found {
		return gauge.(metrics.Gauge)
	}
	dynamicGauge := statsdClient.NewGauge(openConnsGaugeDynamicName)
	statsdDynamicCache.Set(openConnsGaugeDynamicName, dynamicGauge, cache.NoExpiration)
	return dynamicGauge
}

func (r *standardRegistry) EntrypointReqsCounterWithLabel(labelValues []string) metrics.Counter {
	counterDynamicName := statsdEntrypointReqsName + strings.Join(labelValues, "." )
	if counter, found := statsdDynamicCache.Get(counterDynamicName); found {
		return counter.(metrics.Counter)
	}
	dynamicCounter := statsdClient.NewCounter(counterDynamicName , 1.0)
	statsdDynamicCache.Set(counterDynamicName, dynamicCounter, cache.NoExpiration)
	return dynamicCounter
}

func (r *standardRegistry) EntrypointReqDurationHistogramWithLabel(labelValues []string) metrics.Histogram {
	histogramDynamicName := statsdEntrypointReqDurationName + strings.Join(labelValues, "." )
	if histogram, found := statsdDynamicCache.Get(histogramDynamicName); found {
		return histogram.(metrics.Histogram)
	}
	dynamicHistogram := statsdClient.NewTiming(histogramDynamicName , 1.0)
	statsdDynamicCache.Set(histogramDynamicName, dynamicHistogram, cache.NoExpiration)
	return dynamicHistogram
}

func (r *standardRegistry) EntrypointOpenConnsGaugeWithLabel(labelValues []string) metrics.Gauge {
	openConnsGaugeDynamicName := statsdEntrypointOpenConnsName + strings.Join(labelValues, "." )
	if gauge, found := statsdDynamicCache.Get(openConnsGaugeDynamicName); found {
		return gauge.(metrics.Gauge)
	}
	dynamicGauge := statsdClient.NewGauge(openConnsGaugeDynamicName)
	statsdDynamicCache.Set(openConnsGaugeDynamicName, dynamicGauge, cache.NoExpiration)
	return dynamicGauge
}

func (r *standardRegistry) IsStatsd() bool {
	return true
}