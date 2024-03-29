package states

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/util"
)

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

func (state *ipNotConfirmedState) EvaluateState(context Client, resolved *common.DnsRecord) bool {
	ips, err := util.LookupDns(resolved.Host)
	if err != nil {
		log.Info().Msgf("Error looking up dns record %s: %v", resolved.Host, err)
		return true
	}

	for _, hostIp := range ips {
		if hostIp == resolved.IpV4 || hostIp == resolved.IpV6 {
			log.Info().Msgf("DNS record %s verified", resolved.Host)
			context.SetState(NewIpConfirmedState(resolved))
			return false
		}
	}

	log.Info().Msgf("DNS entry for host %s differs to new ip: %v", resolved.Host, resolved)
	if state.checks%10 == 0 {
		since := time.Since(context.GetLastStateChange())
		log.Info().Msgf("Re-sending update as no propagation has happened since %v", since)
		return true
	}

	return false
}

func (state *ipNotConfirmedState) WaitInterval() time.Duration {
	return state.waitInterval
}
