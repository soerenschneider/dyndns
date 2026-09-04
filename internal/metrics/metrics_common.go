package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	}, []string{"version", "hash", "go"})

	Heartbeat = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "heartbeat_timestamp_seconds",
	})

	ProcessStartTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "start_time_seconds",
	})

	NatsConnectionConfigured = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "nats_connection_configured_bool",
	})

	NatsConnectionStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "nats_connection_status",
	}, []string{"url", "status"})

	NatsErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "nats_connection_errors_total",
	}, []string{"url", "error"})

	MqttBrokersConfiguredTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "mqtt_brokers_configured",
	})

	MqttReconnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "mqtt_reconnections_total",
	})

	MqttBrokersConnectedTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "mqtt_brokers_connected",
	})

	MqttConnectionsLostTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "events",
		Name:      "mqtt_connections_lost_total",
	})

	NotificationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "notification_errors_total",
	})

	SqsApiCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "sqs",
		Name:      "api_calls_total",
	}, []string{"operation"})
)

func StartMetricsServer(ctx context.Context, addr string) error {
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

	errChan := make(chan error)
	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}

	return nil
}
