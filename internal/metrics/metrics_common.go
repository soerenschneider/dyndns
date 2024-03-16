package metrics

import (
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	namespace       = "dyndns"
	client          = "client"
	server          = "server"
	DefaultListener = "0.0.0.0:9191"
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

	MqttReconnectionsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "mqtt",
		Name:      "reconnections_total",
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

	SqsApiCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "sqs",
		Help:      "The total amount of SQS API calls",
		Name:      "api_calls_total",
	}, []string{"operation"})
)

func StartMetricsServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:              addr,
		ReadTimeout:       3 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           mux,
	}

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("can not start metrics server")
	}
}
