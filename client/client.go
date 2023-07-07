package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/client/resolvers"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/verification"
)

const DefaultResolveInterval = 45 * time.Second

type State interface {
	// EvaluateState evaluates the current state and returns true if the client should proceed sending a change request
	// using the currently detected ip
	EvaluateState(client *Client, ip *common.ResolvedIp) bool
	// WaitInterval returns the amount of time to sleep after a tick.
	WaitInterval() time.Duration
	Name() string
}

type Client struct {
	signature        verification.SignatureKeypair
	resolver         resolvers.IpResolver
	reconciler       *Reconciler
	state            State
	lastStateChange  time.Time
	notificationImpl notification.Notification
}

func NewClient(resolver resolvers.IpResolver, signature verification.SignatureKeypair, reconciler *Reconciler, notifyImpl notification.Notification) (*Client, error) {
	if resolver == nil {
		return nil, errors.New("no resolver provided")
	}
	if signature == nil {
		return nil, errors.New("no signature provider given")
	}
	if reconciler == nil {
		return nil, errors.New("no reconciler provided")
	}

	c := Client{
		resolver:         resolver,
		reconciler:       reconciler,
		signature:        signature,
		state:            &initialState{},
		lastStateChange:  time.Now(),
		notificationImpl: notifyImpl,
	}

	return &c, nil
}

func (client *Client) Run() {
	ticker := time.NewTicker(DefaultResolveInterval)
	var resolvedIp *common.ResolvedIp
	tick := func() {
		var err error
		resolvedIp, err = client.Resolve(resolvedIp)
		if err != nil {
			log.Info().Msgf("Error while iteration: %v", err)
		}

		if DefaultResolveInterval != client.state.WaitInterval() {
			ticker.Reset(client.state.WaitInterval())
		}
	}

	tick()
	for range ticker.C {
		tick()
	}
}

func (client *Client) resolveIp() (*common.ResolvedIp, error) {
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

func (client *Client) Resolve(prev *common.ResolvedIp) (*common.ResolvedIp, error) {
	resolvedIp, err := client.resolveIp()
	if err != nil {
		return prev, err
	}

	var errs error
	if client.state.EvaluateState(client, resolvedIp) {
		signature := client.signature.Sign(*resolvedIp)
		env := &common.Envelope{
			PublicIp:  *resolvedIp,
			Signature: signature,
		}
		errs = client.reconciler.RegisterUpdate(env)
	}

	return resolvedIp, errs
}

func (client *Client) setState(state State) {
	stateChangeTime := time.Now()
	oldState := client.state
	log.Info().Msgf("State changed from %s -> %s after %s", oldState, state, stateChangeTime.Sub(client.lastStateChange))
	metrics.StatusChangeTimestamp.WithLabelValues(client.resolver.Host(), oldState.Name(), state.Name()).Set(float64(stateChangeTime.Unix()))
	metrics.CurrentStatus.WithLabelValues(client.resolver.Host(), client.state.Name()).Set(0)
	metrics.CurrentStatus.WithLabelValues(client.resolver.Host(), state.Name()).Set(1)

	client.state = state
	client.lastStateChange = stateChangeTime
}
