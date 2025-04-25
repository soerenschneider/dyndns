//go:build server

package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ServerHeartbeat = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "heartbeat_timestamp_seconds",
	})

	KnownHostsHash = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "known_hosts_configuration_hash",
	})

	DnsPropagationRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "dns_propagation_requests_total",
	})

	LatestMessageTimestamp = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "dns_propagation_request_timestamp_seconds",
	})

	SuccessfulDnsPropagationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "dns_propagations_total",
	}, []string{"host"})

	DnsPropagationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "dns_propagations_errors_total",
	}, []string{"host"})

	MessagesReceivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "messages_received_total",
	})

	SignatureVerificationsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "signature_verifications_errors_total",
	}, []string{"host"})

	PublicKeyMissing = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "public_keys_missing_total",
	}, []string{"host"})

	IgnoredMessage = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "messages_ignored_total",
	}, []string{"host", "reason"})

	MessageValidationsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "message_validations_failed_total",
	}, []string{"host", "reason"})

	VaultTokenLifetime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "vault_token_expiry_time_seconds",
	})

	MessageParsingFailed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: server,
		Name:      "message_parsing_failed_total",
	})
)

func StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			ServerHeartbeat.SetToCurrentTime()
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
