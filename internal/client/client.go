package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/client/resolvers"
	"github.com/soerenschneider/dyndns/v2/internal/client/states"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/internal/notification"
	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"go.uber.org/multierr"
)

const DefaultResolveInterval = 45 * time.Second

type Client struct {
	signature                verification.SignatureKeypair
	resolver                 resolvers.IpResolver
	reconciler               *Reconciler
	state                    states.State
	lastStateChange          time.Time
	notificationImpl         notification.Notification
	resolveInterval          time.Duration
	forceSendUpdate          bool
	seenNonRoutableAddresses map[string]struct{}
}

type Opts func(c *Client) error

func NewClient(resolver resolvers.IpResolver, signature verification.SignatureKeypair, reconciler *Reconciler, notifyImpl notification.Notification, opts ...Opts) (*Client, error) {
	if resolver == nil {
		return nil, errors.New("no resolver provided")
	}
	if signature == nil {
		return nil, errors.New("no signature provider given")
	}
	if reconciler == nil {
		return nil, errors.New("no reconciler provided")
	}

	c := &Client{
		resolver:                 resolver,
		resolveInterval:          DefaultResolveInterval,
		reconciler:               reconciler,
		signature:                signature,
		lastStateChange:          time.Now(),
		notificationImpl:         notifyImpl,
		seenNonRoutableAddresses: make(map[string]struct{}),
	}

	var errs error
	for _, opt := range opts {
		if err := opt(c); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	c.state = states.NewInitialState(c.forceSendUpdate)

	return c, errs
}

func (client *Client) Run(ctx context.Context) {
	ticker := time.NewTicker(client.resolveInterval)
	var resolvedIp *update.DnsRecord
	tick := func() {
		var err error
		resolvedIp, err = client.Resolve(ctx, resolvedIp)
		if err != nil {
			log.Info().Err(err).Str("component", "client").Msg("error while iterating")
		}

		if client.resolveInterval != client.state.WaitInterval() {
			ticker.Reset(client.state.WaitInterval())
		}
	}

	tick()
	for {
		select {
		case <-ticker.C:
			tick()
		case <-ctx.Done():
			return
		}
	}
}

func (client *Client) resolveIp() (*update.DnsRecord, error) {
	resolvedIp, err := client.resolver.Resolve()
	metrics.LastCheck.WithLabelValues(client.resolver.Host(), client.resolver.Name()).SetToCurrentTime()

	if err != nil {
		metrics.IpResolveErrors.WithLabelValues(client.resolver.Host(), client.resolver.Name(), "").Inc()
		return nil, fmt.Errorf("error while resolving ip: %v", err)
	}

	// Publish what is routable rather than nothing at all, e.g. the IPv6 of a
	// host whose IPv4 sits behind CGNAT.
	droppedV4, droppedV6 := resolvedIp.DropNonPublic()
	client.reportDroppedIp("ipv4", droppedV4)
	client.reportDroppedIp("ipv6", droppedV6)

	// A dropped address is a definite answer that the family has no public
	// address, unlike a failed resolve, so ask the server to remove its record
	// instead of leaving it pointing at a stale address.
	if droppedV4 != "" {
		resolvedIp.Withdraw(update.AddrFamilyIpv4)
	}
	if droppedV6 != "" {
		resolvedIp.Withdraw(update.AddrFamilyIpv6)
	}

	if !resolvedIp.IsValid() {
		metrics.InvalidResolvedIps.WithLabelValues(client.resolver.Host(), client.resolver.Name(), "").Inc()
		return nil, fmt.Errorf("no publicly routable ip resolved")
	}

	return resolvedIp, err
}

func (client *Client) reportDroppedIp(family, ip string) {
	if ip == "" {
		return
	}

	reason := "not_public"
	if update.IsCGNAT(ip) {
		reason = "cgnat"
	}

	metrics.DroppedResolvedIps.WithLabelValues(client.resolver.Host(), client.resolver.Name(), family, reason).Inc()

	_, found := client.seenNonRoutableAddresses[ip]
	if !found {
		log.Warn().Str("component", "client").Str("host", client.resolver.Host()).Str("resolver", client.resolver.Name()).Str("family", family).Str("ip", ip).Str("reason", reason).Msg("dropping resolved address that is not publicly routable")
		client.seenNonRoutableAddresses[ip] = struct{}{}
	}
}

func (client *Client) Resolve(ctx context.Context, prev *update.DnsRecord) (*update.DnsRecord, error) {
	resolvedIp, err := client.resolveIp()
	if err != nil {
		return prev, err
	}

	var errs error
	if client.state.EvaluateState(client, resolvedIp) {
		signature := client.signature.Sign(*resolvedIp)
		req := &update.UpdateRecordRequest{
			PublicIp:  *resolvedIp,
			Signature: signature,
		}
		errs = client.reconciler.RegisterUpdate(ctx, req)
	}

	return resolvedIp, errs
}

func (client *Client) NotifyUpdatedIpDetected(resolved *update.DnsRecord) error {
	if client.notificationImpl == nil {
		return nil
	}
	return client.notificationImpl.NotifyUpdatedIpDetected(resolved)
}

func (client *Client) GetState() states.State {
	return client.state
}

func (client *Client) GetLastStateChange() time.Time {
	return client.lastStateChange
}

func (client *Client) SetState(state states.State) {
	stateChangeTime := time.Now()
	oldState := client.state
	log.Info().Str("component", "client").Str("old_state", oldState.Name()).Str("new_state", state.Name()).Float64("duration_s", stateChangeTime.Sub(client.lastStateChange).Seconds()).Msgf("State changed")
	metrics.StatusChangeTimestamp.WithLabelValues(client.resolver.Host(), oldState.Name(), state.Name()).Set(float64(stateChangeTime.Unix()))
	metrics.CurrentStatus.WithLabelValues(client.resolver.Host(), client.state.Name()).Set(0)
	metrics.CurrentStatus.WithLabelValues(client.resolver.Host(), state.Name()).Set(1)

	client.state = state
	client.lastStateChange = stateChangeTime
}
