//go:build server

package nats

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	rand2 "math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/metrics"
)

type NatsDyndnsServer struct {
	isInitialized  atomic.Bool
	config         *conf.NatsConfig
	js             jetstream.JetStream
	isOnlyListener bool
	reqChan        chan common.UpdateRecordRequest
}

func NewNatsDyndnsServer(config *conf.NatsConfig, js jetstream.JetStream, reqChan chan common.UpdateRecordRequest) (*NatsDyndnsServer, error) {
	if config == nil {
		return nil, errors.New("nil config supplied")
	}

	ret := &NatsDyndnsServer{
		config:        config,
		js:            js,
		isInitialized: atomic.Bool{},
	}

	ret.isInitialized.Store(js == nil)

	if js == nil {
		go func() {
			for {
				js, err := Connect(*config)
				if err != nil {
					rand := rand2.Intn(15) //nolint G404
					time.Sleep(time.Duration(rand) * time.Second)
				} else {
					ret.js = js
					ret.isInitialized.Store(true)
					return
				}
			}
		}()
	}

	return ret, nil
}

func (n *NatsDyndnsServer) Close(ctx context.Context) error {
	if !n.isInitialized.Load() {
		return nil
	}

	return Close(ctx, n.js)
}

func (n *NatsDyndnsServer) waitUntilConnected(ctx context.Context) error {
	if !n.isInitialized.Load() {
		if n.isOnlyListener {
			return ErrNotInitialized
		}

		for !n.isInitialized.Load() {
			select {
			case <-ctx.Done():
				return nil
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}

	return nil
}

// nolint cyclop
func (n *NatsDyndnsServer) Listen(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	if err := n.waitUntilConnected(ctx); err != nil {
		return err
	}

	cons, err := n.buildConsumer(ctx)
	if err != nil {
		return err
	}

	for {
		msgs, err := cons.Fetch(3, jetstream.FetchMaxWait(1*time.Second))
		if err != nil {
			slog.Error("failed to fetch messages from nats stream", "err", err)
			continue
		}

		select {
		case <-ctx.Done():
			if err := n.js.Conn().FlushTimeout(3 * time.Second); err != nil {
				slog.Error("could not flush nats connection", "err", err)
			}
			n.js.Conn().Close()
			return nil
		default:
			for msg := range msgs.Messages() {
				var env common.UpdateRecordRequest
				if err := json.Unmarshal(msg.Data(), &env); err != nil {
					metrics.MessageParsingFailed.Inc()
					log.Warn().Msgf("Can't parse message: %v", err)
					continue
				}

				n.reqChan <- env
			}
		}

		if err := msgs.Error(); err != nil {
			slog.Error("error while consuming", "err", err)
			metrics.NatsErrors.WithLabelValues(n.config.Url, "consuming").Inc()
		}
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
