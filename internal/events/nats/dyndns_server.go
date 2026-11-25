package nats

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// fetchErrorBackoff is the pause after a failed fetch; replaced in tests.
var fetchErrorBackoff = 5 * time.Second

type NatsDyndnsServer struct {
	isInitialized  atomic.Bool
	config         *hybrid.NatsConfig
	js             jetstream.JetStream
	isOnlyListener bool
	reqChan        chan update.UpdateRecordRequest
}

func NewNatsDyndnsServer(ctx context.Context, config *hybrid.NatsConfig, js jetstream.JetStream, reqChan chan update.UpdateRecordRequest) (*NatsDyndnsServer, error) {
	if config == nil {
		return nil, errors.New("nil config supplied")
	}

	ret := &NatsDyndnsServer{
		config:        config,
		js:            js,
		isInitialized: atomic.Bool{},
		reqChan:       reqChan,
	}

	ret.isInitialized.Store(js != nil)

	if js == nil {
		connectInBackground(ctx, *config, func(js jetstream.JetStream) {
			ret.js = js
			ret.isInitialized.Store(true)
		})
	}

	return ret, nil
}

func (n *NatsDyndnsServer) Close(ctx context.Context) error {
	if !n.isInitialized.Load() {
		return nil
	}

	return Close(ctx, n.js)
}

// waitUntilConnected blocks until the connection is established. It returns
// ctx's error if ctx is done first.
func (n *NatsDyndnsServer) waitUntilConnected(ctx context.Context) error {
	if !n.isInitialized.Load() {
		if n.isOnlyListener {
			return ErrNotInitialized
		}

		for !n.isInitialized.Load() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
	}

	return nil
}

func (n *NatsDyndnsServer) Listen(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	if err := n.waitUntilConnected(ctx); err != nil {
		if ctx.Err() != nil {
			// Shut down before the connection was established.
			return nil
		}
		return err
	}

	cons, err := n.buildConsumer(ctx)
	if err != nil {
		return err
	}

	n.consume(ctx, cons)

	if err := n.js.Conn().FlushTimeout(3 * time.Second); err != nil {
		log.Error().Err(err).Str("component", "nats").Msg("Could not flush nats connection")
	}
	n.js.Conn().Close()
	return nil
}

// consume fetches messages until ctx is done. A failing fetch, e.g. while nats
// is unreachable, is retried after fetchErrorBackoff instead of right away.
func (n *NatsDyndnsServer) consume(ctx context.Context, cons jetstream.Consumer) {
	for ctx.Err() == nil {
		msgs, err := cons.Fetch(3, jetstream.FetchMaxWait(1*time.Second))
		if err != nil {
			log.Error().Err(err).Str("component", "nats").Msg("Failed to fetch messages from nats stream")
			metrics.NatsErrors.WithLabelValues(n.config.Url, "fetching").Inc()
			select {
			case <-ctx.Done():
			case <-time.After(fetchErrorBackoff):
			}
			continue
		}

		for msg := range msgs.Messages() {
			n.handleMessage(ctx, msg)
		}

		if err := msgs.Error(); err != nil {
			log.Error().Err(err).Str("component", "nats").Msg("Error while consuming")
			metrics.NatsErrors.WithLabelValues(n.config.Url, "consuming").Inc()
		}
	}
}

func (n *NatsDyndnsServer) handleMessage(ctx context.Context, msg jetstream.Msg) {
	var env update.UpdateRecordRequest
	if err := json.Unmarshal(msg.Data(), &env); err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Warn().Msgf("Can't parse message: %v", err)
		return
	}

	// The server stops reading requests on shutdown.
	select {
	case n.reqChan <- env:
	case <-ctx.Done():
	}
}

func (n *NatsDyndnsServer) buildConsumer(ctx context.Context) (jetstream.Consumer, error) {
	stream, err := n.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     n.config.StreamName,
		Subjects: n.config.ListenUpdatesSubjects,
	})
	if err != nil {
		return nil, err
	}

	var cons jetstream.Consumer
	cons, err = stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      n.config.ConsumerName,
		Durable:   n.config.ConsumerName,
		AckPolicy: jetstream.AckNonePolicy,
	})
	if err != nil {
		return nil, err
	}

	return cons, nil
}
