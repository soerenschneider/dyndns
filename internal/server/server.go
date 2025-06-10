package server

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	conf2 "github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/server/dns"
	"github.com/soerenschneider/dyndns/internal/util"
	"github.com/soerenschneider/dyndns/internal/verification"
)

// timestampGracePeriod must be a negative number
const timestampGracePeriod = -24 * time.Hour

var ErrorMessageTooOld = errors.New("message timestamp is too old")

type DyndnsServer struct {
	knownHosts       map[string][]verification.VerificationKey
	requests         chan common.UpdateRecordRequest
	propagator       dns.Propagator
	cache            map[string]common.DnsRecord
	notificationImpl notification.Notification

	lock sync.RWMutex
}

func NewServer(config conf2.ServerConf, propagator dns.Propagator, requests chan common.UpdateRecordRequest, notifyImpl notification.Notification) (*DyndnsServer, error) {
	err := conf2.ValidateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid conf passed: %v", err)
	}

	if nil == propagator {
		return nil, errors.New("no dns propagator provided")
	}

	if nil == requests {
		return nil, errors.New("empty/closed channel provided")
	}

	if notifyImpl == nil {
		notifyImpl = &notification.DummyNotification{}
	}

	decoded, err := config.DecodePublicKeys()
	if err != nil {
		return nil, err
	}

	server := DyndnsServer{
		knownHosts:       decoded,
		requests:         requests,
		propagator:       propagator,
		cache:            make(map[string]common.DnsRecord, len(config.KnownHosts)),
		notificationImpl: notifyImpl,
	}

	return &server, nil
}

func (server *DyndnsServer) isCached(env common.UpdateRecordRequest) bool {
	server.lock.RLock()
	defer server.lock.RUnlock()
	entry, ok := server.cache[env.PublicIp.Host]
	if !ok {
		return false
	}

	return entry.Equals(&env.PublicIp)
}

func (server *DyndnsServer) verifyMessage(env common.UpdateRecordRequest) error {
	hostPublicKeys, ok := server.knownHosts[env.PublicIp.Host]
	if !ok {
		metrics.PublicKeyMissing.WithLabelValues(env.PublicIp.Host).Inc()
		return fmt.Errorf("message for unknown host '%s' received", env.PublicIp.Host)
	}

	for _, hostPublicKey := range hostPublicKeys {
		verified := hostPublicKey.Verify(env.Signature, env.PublicIp)
		if verified {
			return nil
		}
	}

	metrics.SignatureVerificationsFailed.WithLabelValues(env.PublicIp.Host).Inc()
	return fmt.Errorf("verifying signature FAILED for host '%s'", env.PublicIp.Host)
}

func (server *DyndnsServer) HandlePropagateRequest(env common.UpdateRecordRequest) error {
	if err := env.Validate(); err != nil {
		metrics.MessageValidationsFailed.WithLabelValues(env.PublicIp.Host, "invalid_fields").Inc()
		return fmt.Errorf("invalid envelope received: %v", err)
	}

	if err := server.verifyMessage(env); err != nil {
		return err
	}

	if env.PublicIp.Timestamp.Before(time.Now().Add(timestampGracePeriod)) {
		metrics.IgnoredMessage.WithLabelValues(env.PublicIp.Host, "message_too_old").Inc()
		return ErrorMessageTooOld
	}

	if server.isCached(env) {
		log.Info().Str("component", "server").Str("host", env.PublicIp.Host).Msg("Request for host is cached, not performing changes")
		return nil
	}

	if util.HostnameMatchesIp(env.PublicIp.Host, env.PublicIp.IpV4, env.PublicIp.IpV6) {
		log.Info().Str("component", "server").Str("host", env.PublicIp.Host).Str("ipv4", env.PublicIp.IpV4).Str("ipv6", env.PublicIp.IpV6).Msg("host already has desired address, not updating")
		return nil
	}

	log.Info().Str("component", "server").Str("host", env.PublicIp.Host).Str("ipv4", env.PublicIp.IpV4).Str("ipv6", env.PublicIp.IpV6).Msg("Verifying signature succeeded, updating host")
	if err := server.propagator.PropagateChange(env.PublicIp); err != nil {
		metrics.DnsPropagationErrors.WithLabelValues(env.PublicIp.Host).Inc()
		return fmt.Errorf("could not propagate dns change for domain '%s': %v", env.PublicIp.Host, err)
	}

	if server.notificationImpl != nil {
		pubIp := env.PublicIp
		err := server.notificationImpl.NotifyUpdatedIpApplied(&pubIp)
		if err != nil {
			metrics.NotificationErrors.Inc()
		}
	}
	log.Info().Str("component", "server").Str("host", env.PublicIp.Host).Str("ipv4", env.PublicIp.IpV4).Str("ipv6", env.PublicIp.IpV6).Msg("Successfully propagated change")
	metrics.SuccessfulDnsPropagationsTotal.WithLabelValues(env.PublicIp.Host).Inc()

	// Add to cache
	server.cache[env.PublicIp.Host] = env.PublicIp
	return nil
}

func (server *DyndnsServer) Listen() {
	for request := range server.requests {
		metrics.MessagesReceivedTotal.Inc()
		metrics.LatestMessageTimestamp.SetToCurrentTime()

		log.Info().Str("component", "server").Msg("Picked up a new change request")
		err := server.HandlePropagateRequest(request)
		if err != nil && !errors.Is(err, ErrorMessageTooOld) {
			log.Error().Err(err).Str("component", "server").Msg("Change has not been propagated")
		}
	}
}
