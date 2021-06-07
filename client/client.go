package client

import (
	"dyndns/client/resolvers"
	"dyndns/internal/common"
	"dyndns/internal/events"
	"dyndns/internal/metrics"
	"dyndns/internal/verification"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

const defaultResolveInterval = 2 * time.Minute

type State interface {
	// PerformIpLookup returns true when a lookup for a potential new ip should be performed.
	PerformIpLookup() bool
	// EvaluateState evaluates the current state and returns true if the client should proceed sending a change request
	// using the currently detected ip
	EvaluateState(client *Client, ip *common.ResolvedIp) bool
	// WaitInterval returns the amount of time to sleep after a tick.
	WaitInterval() time.Duration
	Name() string
}

type Client struct {
	signature       verification.SignatureKeypair
	resolver        resolvers.IpResolver
	dispatcher      events.EventDispatch
	state           State
	lastStateChange time.Time
}

func NewClient(resolver resolvers.IpResolver, signature verification.SignatureKeypair, dispatcher events.EventDispatch) (*Client, error) {
	if resolver == nil {
		return nil, errors.New("no resolver provided")
	}
	if signature == nil {
		return nil, errors.New("no signature provider given")
	}
	if dispatcher == nil {
		return nil, errors.New("no dispatcher provided")
	}

	c := Client{
		resolver:        resolver,
		dispatcher:      dispatcher,
		signature:       signature,
		state:           &initialState{},
		lastStateChange: time.Now(),
	}

	return &c, nil
}

func (client *Client) Run() {
	var resolvedIp *common.ResolvedIp
	for {
		var err error
		resolvedIp, err = client.Resolve(resolvedIp)
		if err != nil {
			log.Info().Msgf("Error while iteration: %v", err)
		}
		time.Sleep(client.state.WaitInterval())
	}
}

func (client *Client) resolveIp() (*common.ResolvedIp, error) {
	resolvedIp, err := client.resolver.Resolve()
	metrics.LastCheck.WithLabelValues(client.resolver.Host(), client.resolver.Name()).SetToCurrentTime()

	if err != nil {
		metrics.IpResolveErrors.WithLabelValues(client.resolver.Host(), client.resolver.Name()).Inc()
		return nil, fmt.Errorf("error while resolving ip: %v", err)
	}
	if !resolvedIp.IsValid() {
		metrics.InvalidResolvedIps.WithLabelValues(client.resolver.Host(), client.resolver.Name()).Inc()
		return nil, fmt.Errorf("resolvedip is invalid")
	}

	metrics.ResolvedIps.WithLabelValues(client.resolver.Host(), client.resolver.Name()).Inc()
	return resolvedIp, err
}

func (client *Client) ResolveSingle() (*common.ResolvedIp, error) {
	return client.Resolve(nil)
}

func (client *Client) Resolve(prev *common.ResolvedIp) (*common.ResolvedIp, error) {
	var resolvedIp = prev

	if prev == nil || client.state.PerformIpLookup() {
		var err error
		resolvedIp, err = client.resolveIp()
		if err != nil {
			return prev, err
		}
	}

	if client.state.EvaluateState(client, resolvedIp) {
		signature := client.signature.Sign(*resolvedIp)
		env := common.Envelope{
			PublicIp:  *resolvedIp,
			Signature: signature,
		}

		err := client.dispatcher.Notify(env)
		if err != nil {
			metrics.UpdateDispatchErrors.WithLabelValues(client.resolver.Host()).Inc()
			return resolvedIp, fmt.Errorf("could not dispatch ip update notification: %v", err)
		}
		metrics.UpdatesDispatched.Inc()
	}

	return resolvedIp, nil
}

func (client *Client) setState(state State) {
	stateChangeTime := time.Now()
	delta := stateChangeTime.Sub(client.lastStateChange)
	oldState := client.state
	log.Info().Msgf("State changed from %s -> %s after %s", oldState, state, delta)
	metrics.StatusChangeTimestamp.WithLabelValues(client.resolver.Host(), oldState.Name(), state.Name()).Set(float64(stateChangeTime.Unix()))

	client.state = state
	client.lastStateChange = stateChangeTime
}
