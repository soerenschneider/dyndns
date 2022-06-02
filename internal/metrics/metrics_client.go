//go:build client

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
)

const (
	namespace       = "dyndns"
	client          = "client"
	DefaultListener = ":9191"
)

var (
	Version = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "version",
	}, []string{"version", "hash"})

	MqttConnectionsLostTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "mqtt_connections_lost_total",
	})

	IpResolveErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolves_errors_total",
	}, []string{"host", "resolver"})

	InvalidResolvedIps = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolves_invalid_total",
	}, []string{"host", "resolver"})

	ResolvedIps = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolves_success_total",
	}, []string{"host", "resolver"})

	LastCheck = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolves_last_check_timestamp_seconds",
	}, []string{"host", "resolver"})

	UpdateDispatchErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "updates_dispatch_errors_total",
	}, []string{"host"})

	UpdatesDispatched = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "updates_dispatched_total",
	})

	StatusChangeTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "state_changed_timestamp",
	}, []string{"host", "from", "to"})
)

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Msgf("can not start metrics server at %s: %v", addr, err)
	}
}
