package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	rand2 "math/rand"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/conf"
)

type NatsDyndnsClient struct {
	config *conf.NatsConfig

	js            jetstream.JetStream
	isInitialized atomic.Bool
}

func NewNatsDyndnsClient(config *conf.NatsConfig, js jetstream.JetStream) (*NatsDyndnsClient, error) {
	if config == nil {
		return nil, errors.New("nil config supplied")
	}

	ret := &NatsDyndnsClient{
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

func (n *NatsDyndnsClient) Close(ctx context.Context) error {
	if !n.isInitialized.Load() {
		return nil
	}

	return Close(ctx, n.js)
}

func (n *NatsDyndnsClient) Notify(msg *common.UpdateRecordRequest) error {
	if !n.isInitialized.Load() {
		return ErrNotInitialized
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal envelope: %w", err)
	}

	ctx := context.Background()
	ack, err := n.js.PublishMsg(ctx, &nats.Msg{
		Data:    data,
		Subject: n.config.DispatchUpdatesSubject,
	})
	if err != nil {
		return err
	}

	log.Debug().Uint64("sequence number", ack.Sequence).Str("stream", ack.Stream).Msg("Published msg")
	return nil
}
