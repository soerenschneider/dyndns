package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/client/resolvers"
	"github.com/soerenschneider/dyndns/internal/client/states"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/verification"
	"go.uber.org/multierr"
)

const DefaultResolveInterval = 45 * time.Second

type EventDispatch interface {
	Notify(msg *common.UpdateRecordRequest) error
}

type Client struct {
	signature        verification.SignatureKeypair
	resolver         resolvers.IpResolver
	reconciler       *Reconciler
	state            states.State
	lastStateChange  time.Time
	notificationImpl notification.Notification
	resolveInterval  time.Duration
	forceSendUpdate  bool
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
		resolver:         resolver,
		resolveInterval:  DefaultResolveInterval,
		reconciler:       reconciler,
		signature:        signature,
		lastStateChange:  time.Now(),
		notificationImpl: notifyImpl,
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

func (client *Client) Run() {
	ticker := time.NewTicker(client.resolveInterval)
	var resolvedIp *common.DnsRecord
	tick := func() {
		var err error
		resolvedIp, err = client.Resolve(resolvedIp)
		if err != nil {
			log.Info().Err(err).Str("component", "client").Msg("error while iterating")
		}

		if client.resolveInterval != client.state.WaitInterval() {
			ticker.Reset(client.state.WaitInterval())
		}
	}

	tick()
	for range ticker.C {
		tick()
	}
}

func (client *Client) resolveIp() (*common.DnsRecord, error) {
	resolvedIp, err := client.resolver.Resolve()
	metrics.LastCheck.WithLabelValues(client.resolver.Host(), client.resolver.Name()).SetToCurrentTime()

	if err != nil {
		metrics.IpResolveErrors.WithLabelValues(client.resolver.Host(), client.resolver.Name(), "").Inc()
		return nil, fmt.Errorf("error while resolving ip: %v", err)
	}
	if !resolvedIp.IsValid() {
		metrics.InvalidResolvedIps.WithLabelValues(client.resolver.Host(), client.resolver.Name(), "").Inc()
		return nil, fmt.Errorf("resolvedip is invalid")
	}

	return resolvedIp, err
}

func (client *Client) Resolve(prev *common.DnsRecord) (*common.DnsRecord, error) {
	resolvedIp, err := client.resolveIp()
	if err != nil {
		return prev, err
	}

	var errs error
	if client.state.EvaluateState(client, resolvedIp) {
		signature := client.signature.Sign(*resolvedIp)
		req := &common.UpdateRecordRequest{
			PublicIp:  *resolvedIp,
			Signature: signature,
		}
		errs = client.reconciler.RegisterUpdate(req)
	}

	return resolvedIp, errs
}

func (client *Client) NotifyUpdatedIpDetected(resolved *common.DnsRecord) error {
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
	log.Info().Str("component", "client").Str("old_state", oldState.Name()).Str("new_state", state.Name()).Msgf("State changed from after %s", stateChangeTime.Sub(client.lastStateChange))
	metrics.StatusChangeTimestamp.WithLabelValues(client.resolver.Host(), oldState.Name(), state.Name()).Set(float64(stateChangeTime.Unix()))
	metrics.CurrentStatus.WithLabelValues(client.resolver.Host(), client.state.Name()).Set(0)
	metrics.CurrentStatus.WithLabelValues(client.resolver.Host(), state.Name()).Set(1)

	client.state = state
	client.lastStateChange = stateChangeTime
}
