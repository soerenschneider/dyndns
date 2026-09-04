package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IpResolveErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolves_errors_total",
	}, []string{"host", "resolver", "name"})

	IpsResolved = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolved_successful_total",
	}, []string{"host", "resolver", "name"})

	ReconcilersActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "reconcilers_pending_changes",
	}, []string{"host"})

	ReconcilerTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "reconciler_timestamp_seconds",
	}, []string{"host"})

	InvalidResolvedIps = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "ip_resolves_invalid_total",
	}, []string{"host", "resolver", "url"})

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

	UpdatesDispatched = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "updates_dispatched_total",
	}, []string{"host"})

	StatusChangeTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "state_changed_timestamp_seconds",
	}, []string{"host", "from", "to"})

	CurrentStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "current_state",
	}, []string{"host", "state"})

	ResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "resolver_response_time_seconds",
	}, []string{"resolver"})
)
