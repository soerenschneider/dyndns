package states

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/util"
)

// ipConfirmedState is set after the dns record has been verified successfully
type ipConfirmedState struct {
	previouslyResolvedIp *common.DnsRecord
}

func NewIpConfirmedState(prev *common.DnsRecord) State {
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

func (state *ipConfirmedState) EvaluateState(context Client, resolved *common.DnsRecord) bool {
	hasIpChanged := !state.previouslyResolvedIp.Equals(resolved)
	state.previouslyResolvedIp = resolved

	if hasIpChanged {
		log.Info().Msgf("New IP detected: %s", resolved)
		context.SetState(NewIpNotConfirmedState())

		/*
			if context.notificationImpl != nil {
				if err := context.notificationImpl.NotifyUpdatedIpDetected(resolved); err != nil {
					metrics.NotificationErrors.Inc()
				}
			}
		*/
	}

	ips, err := util.LookupDns(resolved.Host)
	if err == nil {
		found := false
		for _, hostIp := range ips {
			if hostIp == resolved.IpV4 || hostIp == resolved.IpV6 {
				found = true
				break
			}
		}
		if !found {
			log.Info().Msgf("Detected changed DNS record: %s", resolved)
			context.SetState(NewIpNotConfirmedState())
		}
	}

	return hasIpChanged
}

func (state *ipConfirmedState) WaitInterval() time.Duration {
	return 45 * time.Second
}
