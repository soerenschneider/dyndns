package client

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/util"
	"math/rand"
	"time"
)

const jitterSeconds = 15

type initialState struct{}

func (state *initialState) PerformIpLookup() bool {
	return true
}

func (state *initialState) String() string {
	return "initialState"
}

func (state *initialState) Name() string {
	return state.String()
}

func (state *initialState) EvaluateState(context *Client, resolved *common.ResolvedIp) bool {
	// This is just a dummy state, we'll immediately set the next state and invoke it
	context.setState(NewIpNotConfirmedState())
	return context.state.EvaluateState(context, resolved)
}

func (state *initialState) WaitInterval() time.Duration {
	return defaultResolveInterval
}

// ipNotConfirmedState is the state after we detect an ip update. we stay in this state until the dns record has been
// verified to contain our resolved ips
type ipNotConfirmedState struct {
	checks       int64
	waitInterval time.Duration
}

func NewIpNotConfirmedState() State {
	return &ipNotConfirmedState{checks: 0, waitInterval: 30 * time.Second}
}

func (state *ipNotConfirmedState) PerformIpLookup() bool {
	// after detecting an ip update, only perform a new lookup after being called for the 10th time (300s) or the state
	// has changed
	state.checks++
	return state.checks%10 == 0
}

func (state *ipNotConfirmedState) String() string {
	return fmt.Sprintf("ipNotConfirmedState (%d checks)", state.checks)
}

func (state *ipNotConfirmedState) Name() string {
	return "ipNotConfirmedState"
}

func (state *ipNotConfirmedState) EvaluateState(context *Client, resolved *common.ResolvedIp) bool {
	ips, err := util.LookupDns(resolved.Host)
	if err != nil {
		log.Info().Msgf("Error looking up dns record %s: %v", resolved.Host, err)
		return true
	}

	for _, hostIp := range ips {
		if hostIp == resolved.IpV4 || hostIp == resolved.IpV6 {
			log.Info().Msgf("DNS record %s verified", resolved.Host)
			context.setState(NewIpConfirmedState(resolved))
			return false
		}
	}

	log.Info().Msgf("DNS entry for host %s differs to new ip: %v", resolved.Host, resolved)
	if state.checks%120 == 0 {
		log.Info().Msgf("Verifying for %d minutes already, re-sending message..", int64(state.waitInterval.Seconds())*state.checks/60)
		return true
	}

	return false
}

func (state *ipNotConfirmedState) WaitInterval() time.Duration {
	return state.waitInterval
}

// ipConfirmedState is set after the dns record has been verified successfully
type ipConfirmedState struct {
	previouslyResolvedIp *common.ResolvedIp
	checks               int64
	since                time.Time
}

func NewIpConfirmedState(prev *common.ResolvedIp) State {
	return &ipConfirmedState{
		previouslyResolvedIp: prev,
		checks:               0,
		since:                time.Now(),
	}
}

func (state *ipConfirmedState) PerformIpLookup() bool {
	return true
}

func (state *ipConfirmedState) String() string {
	return fmt.Sprintf("ipConfirmedState (%s)", state.previouslyResolvedIp)
}

func (state *ipConfirmedState) Name() string {
	return "ipConfirmedState"
}

func (state *ipConfirmedState) EvaluateState(context *Client, resolved *common.ResolvedIp) bool {
	hasIpChanged := !state.previouslyResolvedIp.Equals(resolved)
	state.previouslyResolvedIp = resolved

	state.checks++
	if hasIpChanged {
		log.Info().Msgf("New IP detected: %s", resolved)
		context.setState(NewIpNotConfirmedState())
	} else if state.checks%240 == 0 {
		lastChange := int64(time.Now().Sub(state.since).Minutes())
		log.Info().Msgf("Performed %d checks since %d minutes without a new IP", state.checks, lastChange)
	}

	return hasIpChanged
}

func (state *ipConfirmedState) WaitInterval() time.Duration {
	return defaultResolveInterval + jitter()
}

func jitter() time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(rand.Intn(jitterSeconds*2)-jitterSeconds) * time.Second
}
