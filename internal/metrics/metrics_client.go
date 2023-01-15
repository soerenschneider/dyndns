//go:build client

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

	CurrentStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: client,
		Name:      "current_state_bool",
	}, []string{"host", "state"})
)
