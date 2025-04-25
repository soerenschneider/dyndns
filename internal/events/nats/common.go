package nats

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/metrics"
)

var (
	connections       = map[string]jetstream.JetStream{}
	mutex             sync.Mutex
	ErrNotInitialized = errors.New("nats not initialized")
)

func Close(ctx context.Context, js jetstream.JetStream) error {
	mutex.Lock()
	defer mutex.Unlock()

	c := js.Conn()
	if c != nil {
		err := c.FlushWithContext(ctx)
		c.Close()
		delete(connections, js.Conn().ConnectedUrl())
		return err
	}

	return nil
}

func Connect(config conf.NatsConfig) (jetstream.JetStream, error) {
	mutex.Lock()
	defer mutex.Unlock()

	js, found := connections[config.Url]
	if found {
		return js, nil
	}

	nc, err := nats.Connect(config.Url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(5*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warn().Str("component", "nats").Str("url", config.Url).Msg("Disconnected")
			metrics.NatsConnectionStatus.WithLabelValues(config.Url, "connected").Set(0)
			metrics.NatsConnectionStatus.WithLabelValues(config.Url, "disconnected").Set(1)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info().Str("components", "nats").Str("url", config.Url).Msg("Reconnected")
			metrics.NatsConnectionStatus.WithLabelValues(config.Url, "connected").Set(1)
			metrics.NatsConnectionStatus.WithLabelValues(config.Url, "disconnected").Set(0)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Warn().Str("component", "nats").Str("url", config.Url).Msg("Connection closed")
		}))
	if err != nil {
		return nil, err
	}

	metrics.NatsConnectionConfigured.Set(1)
	metrics.NatsConnectionStatus.WithLabelValues(config.Url, "connected").Set(1)
	metrics.NatsConnectionStatus.WithLabelValues(config.Url, "disconnected").Set(0)

	js, err = jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	connections[config.Url] = js
	return js, nil
}
