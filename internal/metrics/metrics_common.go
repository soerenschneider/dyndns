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

	MqttConnectionsLostTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "mqtt_connections_lost_total",
	})
)

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Msgf("can not start metrics server at %s: %v", addr, err)
	}
}
