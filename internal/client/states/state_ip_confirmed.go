package states

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/internal/util"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// ipConfirmedState is set after the dns record has been verified successfully
type ipConfirmedState struct {
	previouslyResolvedIp *update.DnsRecord
}

func NewIpConfirmedState(prev *update.DnsRecord) State {
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

func (state *ipConfirmedState) EvaluateState(client Client, resolved *update.DnsRecord) bool {
	hasIpChanged := !state.previouslyResolvedIp.Equals(resolved)
	state.previouslyResolvedIp = resolved

	if hasIpChanged {
		log.Info().Str("component", "state_machine").Str("state", state.Name()).Str("ipv4", resolved.IpV4).Str("ipv6", resolved.IpV6).Str("host", resolved.Host).Msg("New IP detected")
		client.SetState(NewIpNotConfirmedState())

		if err := client.NotifyUpdatedIpDetected(resolved); err != nil {
			log.Error().Err(err).Msg("could not send notification")
			metrics.NotificationErrors.Inc()
		}
	}

	ips, err := util.LookupDns(resolved.Host)
	if err == nil {
		if !util.RecordMatchesIps(resolved, ips) {
			log.Info().Str("component", "state_machine").Str("state", state.Name()).Str("ipv4", resolved.IpV4).Str("ipv6", resolved.IpV6).Str("host", resolved.Host).Msg("Detected changed DNS record")
			client.SetState(NewIpNotConfirmedState())
		}
	}

	return hasIpChanged
}

func (state *ipConfirmedState) WaitInterval() time.Duration {
	return 45 * time.Second
}
