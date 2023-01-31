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
	server          = "server"
	DefaultListener = ":9191"
)

var (
	Version = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "version",
	}, []string{"version", "hash"})

	ProcessStartTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "start_time_seconds",
	})
	MqttBrokersConfiguredTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "mqtt",
		Name:      "brokers_configured_total",
	})

	MqttBrokersConnectedTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "mqtt",
		Name:      "brokers_connected_total",
	})

	MqttConnectionsLostTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "mqtt_connections_lost_total",
	})

	NotificationErrors = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "notification_errors",
	})
)

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Msgf("can not start metrics server at %s: %v", addr, err)
	}
}
