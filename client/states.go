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
	return DefaultResolveInterval
}

// ipNotConfirmedState is the state after we detect an ip update. we stay in this state until the dns record has been
// verified to contain our resolved ips
type ipNotConfirmedState struct {
	checks       int64
	waitInterval time.Duration
}

func NewIpNotConfirmedState() State {
	return &ipNotConfirmedState{
		checks:       0,
		waitInterval: 30 * time.Second,
	}
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
	if state.checks%10 == 0 {
		since := time.Now().Sub(context.lastStateChange)
		log.Info().Msgf("Re-sending update as no propagation has happened since %v", since)
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
}

func NewIpConfirmedState(prev *common.ResolvedIp) State {
	return &ipConfirmedState{
		previouslyResolvedIp: prev,
	}
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

	if hasIpChanged {
		log.Info().Msgf("New IP detected: %s", resolved)
		context.setState(NewIpNotConfirmedState())

		if context.notificationImpl != nil {
			context.notificationImpl.NotifyUpdatedIpDetected(resolved)
		}
	}

	return hasIpChanged
}

func (state *ipConfirmedState) WaitInterval() time.Duration {
	return DefaultResolveInterval
}

func jitter() time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(rand.Intn(jitterSeconds*2)-jitterSeconds) * time.Second
}
