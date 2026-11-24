package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	conf2 "github.com/soerenschneider/dyndns/v2/internal/conf"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/internal/notification"
	"github.com/soerenschneider/dyndns/v2/internal/util"
	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

const (
	// maxMessageAge bounds how long a captured request can be replayed, e.g.
	// right after a restart wiped the cache. Clients re-send unconfirmed updates
	// every few minutes with a fresh timestamp, so legitimate requests are young.
	maxMessageAge = 1 * time.Hour
	// maxClockSkew tolerates clients whose clock runs slightly ahead.
	maxClockSkew = 5 * time.Minute
)

var (
	ErrorMessageTooOld     = errors.New("message timestamp is too old")
	ErrorMessageFromFuture = errors.New("message timestamp is in the future")
	// ErrorMessageSuperseded is returned for requests older than the last one
	// accepted for the host: a replay, or a duplicate delivered late via
	// another transport. Applying it would roll the record back.
	ErrorMessageSuperseded = errors.New("message is superseded by a newer one")
)

type propagator interface {
	PropagateChange(ctx context.Context, ip update.DnsRecord) error
}

type DyndnsServer struct {
	knownHosts map[string][]verification.VerificationKey
	requests   chan update.UpdateRecordRequest
	propagator propagator
	// cache holds the last accepted record per host. Its timestamp orders
	// incoming requests, so it is only written after signature verification.
	cache            map[string]update.DnsRecord
	notificationImpl notification.Notification
	// dnsMatches reports whether DNS already reflects a record.
	dnsMatches func(*update.DnsRecord) bool

	// lock serializes UpdateRecord so the ordering check and the cache update
	// cannot interleave.
	lock sync.Mutex
}

func NewServer(config conf2.Conf, propagator propagator, requests chan update.UpdateRecordRequest, notifyImpl notification.Notification) (*DyndnsServer, error) {
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

	var knownHosts map[string][]verification.VerificationKey
	switch config.Mode {
	case internal.EdgeMode:
		keypair, err := verification.NewEdge(config.Client.Host)
		if err != nil {
			return nil, fmt.Errorf("could not build 'edge' encryption key: %w", err)
		}
		knownHosts = map[string][]verification.VerificationKey{
			config.Client.Host: {keypair},
		}
	case internal.ServerMode:
		knownHosts, err = config.Server.DecodePublicKeys()
		if err != nil {
			return nil, err
		}
	}

	server := DyndnsServer{
		knownHosts:       knownHosts,
		requests:         requests,
		propagator:       propagator,
		cache:            make(map[string]update.DnsRecord, len(config.Server.KnownHosts)),
		notificationImpl: notifyImpl,
		dnsMatches:       util.HostnameMatchesRecord,
	}

	return &server, nil
}

// checkTimestamp rejects requests outside [now-maxMessageAge, now+maxClockSkew].
func checkTimestamp(record update.DnsRecord, now time.Time) error {
	if record.Timestamp.Before(now.Add(-maxMessageAge)) {
		metrics.IgnoredMessage.WithLabelValues(record.Host, "message_too_old").Inc()
		return ErrorMessageTooOld
	}

	if record.Timestamp.After(now.Add(maxClockSkew)) {
		metrics.IgnoredMessage.WithLabelValues(record.Host, "message_from_future").Inc()
		return ErrorMessageFromFuture
	}

	return nil
}

// checkOrder compares a verified request with the last accepted one for its
// host. It returns isDuplicate for a repeat of that request and
// ErrorMessageSuperseded for anything older. The signature only covers whole
// seconds, so that is the resolution requests are ordered at.
func (server *DyndnsServer) checkOrder(record update.DnsRecord) (isDuplicate bool, err error) {
	last, ok := server.cache[record.Host]
	if !ok {
		return false, nil
	}

	reqSecond, lastSecond := record.Timestamp.Unix(), last.Timestamp.Unix()
	switch {
	case reqSecond < lastSecond:
		metrics.IgnoredMessage.WithLabelValues(record.Host, "message_superseded").Inc()
		return false, ErrorMessageSuperseded
	case last.Equals(&record):
		return true, nil
	case reqSecond == lastSecond:
		metrics.IgnoredMessage.WithLabelValues(record.Host, "message_superseded").Inc()
		return false, ErrorMessageSuperseded
	}

	return false, nil
}

func (server *DyndnsServer) verifyMessage(env update.UpdateRecordRequest) error {
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

func (server *DyndnsServer) UpdateRecord(ctx context.Context, req update.UpdateRecordRequest) error {
	if err := req.Validate(); err != nil {
		metrics.MessageValidationsFailed.WithLabelValues(req.PublicIp.Host, "invalid_fields").Inc()
		return fmt.Errorf("invalid envelope received: %v", err)
	}

	if err := server.verifyMessage(req); err != nil {
		return err
	}

	if err := checkTimestamp(req.PublicIp, time.Now()); err != nil {
		return err
	}

	server.lock.Lock()
	defer server.lock.Unlock()

	isDuplicate, err := server.checkOrder(req.PublicIp)
	if err != nil {
		return err
	}

	if isDuplicate {
		server.cache[req.PublicIp.Host] = req.PublicIp // advance the timestamp
		log.Info().Str("component", "server").Str("host", req.PublicIp.Host).Msg("Request for host is cached, not performing changes")
		return nil
	}

	if server.dnsMatches(&req.PublicIp) {
		server.cache[req.PublicIp.Host] = req.PublicIp
		log.Info().Str("component", "server").Str("host", req.PublicIp.Host).Str("ipv4", req.PublicIp.IpV4).Str("ipv6", req.PublicIp.IpV6).Msg("host already has desired address, not updating")
		return nil
	}

	log.Info().Str("component", "server").Str("host", req.PublicIp.Host).Str("ipv4", req.PublicIp.IpV4).Str("ipv6", req.PublicIp.IpV6).Strs("withdrawn_families", req.PublicIp.WithdrawnFamilies).Msg("Verifying signature succeeded, updating host")
	if err := server.propagator.PropagateChange(ctx, req.PublicIp); err != nil {
		metrics.DnsPropagationErrors.WithLabelValues(req.PublicIp.Host).Inc()
		return fmt.Errorf("could not propagate dns change for domain '%s': %v", req.PublicIp.Host, err)
	}

	if server.notificationImpl != nil {
		pubIp := req.PublicIp
		err := server.notificationImpl.NotifyUpdatedIpApplied(&pubIp)
		if err != nil {
			metrics.NotificationErrors.Inc()
		}
	}
	log.Info().Str("component", "server").Str("host", req.PublicIp.Host).Str("ipv4", req.PublicIp.IpV4).Str("ipv6", req.PublicIp.IpV6).Msg("Successfully propagated change")
	metrics.SuccessfulDnsPropagationsTotal.WithLabelValues(req.PublicIp.Host).Inc()

	// Add to cache
	server.cache[req.PublicIp.Host] = req.PublicIp
	return nil
}

func (server *DyndnsServer) Listen(ctx context.Context) {
	for {
		select {
		case request := <-server.requests:
			metrics.MessagesReceivedTotal.Inc()
			metrics.LatestMessageTimestamp.SetToCurrentTime()

			log.Info().Str("component", "server").Msg("Picked up a new change request")
			err := server.UpdateRecord(ctx, request)
			if err != nil && !errors.Is(err, ErrorMessageTooOld) && !errors.Is(err, ErrorMessageSuperseded) {
				log.Error().Err(err).Str("component", "server").Msg("Change has not been propagated")
			}
			time.Sleep(1 * time.Second)
		case <-ctx.Done():
			return
		}
	}
}
