package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/events"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/util"
	"github.com/soerenschneider/dyndns/internal/verification"
	"github.com/soerenschneider/dyndns/server/dns"
)

// timestampGracePeriod must be a negative number
const timestampGracePeriod = -24 * time.Hour

var ErrorMessageTooOld = errors.New("message timestamp is too old")

type Server struct {
	knownHosts       map[string][]verification.VerificationKey
	listener         events.EventListener
	requests         chan common.Envelope
	propagator       dns.Propagator
	cache            map[string]common.ResolvedIp
	notificationImpl notification.Notification
}

func NewServer(config conf.ServerConf, propagator dns.Propagator, requests chan common.Envelope, notifyImpl notification.Notification) (*Server, error) {
	err := conf.ValidateConfig(config)
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

	server := Server{
		knownHosts:       config.DecodePublicKeys(),
		requests:         requests,
		propagator:       propagator,
		cache:            make(map[string]common.ResolvedIp, len(config.KnownHosts)),
		notificationImpl: notifyImpl,
	}

	return &server, nil
}

func (server *Server) isCached(env common.Envelope) bool {
	entry, ok := server.cache[env.PublicIp.Host]
	if !ok {
		return false
	}

	return entry.Equals(&env.PublicIp)
}

func (server *Server) verifyMessage(env common.Envelope) error {
	hostPublicKeys, ok := server.knownHosts[env.PublicIp.Host]
	if !ok {
		metrics.PublicKeyMissing.WithLabelValues(env.PublicIp.Host).Inc()
		return fmt.Errorf("message for unknown domain '%s' received", env.PublicIp.Host)
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

func (server *Server) handlePropagateRequest(env common.Envelope) error {
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
		log.Info().Msgf("Request for host %s is cached, not performing changes", env.PublicIp.Host)
		return nil
	}

	if util.HostnameMatchesIp(env.PublicIp.Host, env.PublicIp.IpV4, env.PublicIp.IpV6) {
		log.Info().Msgf("Host %s already resolved to IPv4 %s / Ipv6 %s, not performing changes", env.PublicIp.Host, env.PublicIp.IpV4, env.PublicIp.IpV6)
		return nil
	}

	log.Info().Msgf("Verifying signature succeeded for domain '%v', performing DNS change", env.PublicIp)
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
	log.Info().Msgf("Successfully propagated change '%s'", env.PublicIp.String())
	metrics.SuccessfulDnsPropagationsTotal.WithLabelValues(env.PublicIp.Host).Inc()

	// Add to cache
	server.cache[env.PublicIp.Host] = env.PublicIp
	return nil
}

func (server *Server) Listen() {
	for request := range server.requests {
		metrics.MessagesReceivedTotal.Inc()
		metrics.LatestMessageTimestamp.SetToCurrentTime()

		log.Info().Msg("Picked up a new change request")
		err := server.handlePropagateRequest(request)
		if err != nil && !errors.Is(err, ErrorMessageTooOld) {
			log.Error().Msgf("Change has not been propagated: %v", err)
		}
	}
}
