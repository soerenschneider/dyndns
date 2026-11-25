package nats

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
)

var (
	connections       = map[string]jetstream.JetStream{}
	mutex             sync.Mutex
	ErrNotInitialized = errors.New("nats not initialized")

	// connect and the reconnect delays are replaced in tests.
	connect           = Connect
	reconnectMinDelay = 1 * time.Second
	reconnectMaxDelay = 1 * time.Minute
)

// connectInBackground connects to nats until it succeeds or ctx is done,
// waiting twice as long after every failed attempt, and hands the connection to
// onConnect. The returned channel is closed once it has connected or given up.
func connectInBackground(ctx context.Context, config hybrid.NatsConfig, onConnect func(jetstream.JetStream)) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		delay := reconnectMinDelay
		for {
			js, err := connect(config)
			if err == nil {
				onConnect(js)
				return
			}

			log.Warn().Err(err).Str("component", "nats").Str("url", config.Url).Dur("retry_in", delay).Msg("Could not connect, retrying")
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			delay = min(2*delay, reconnectMaxDelay)
		}
	}()

	return done
}

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

func Connect(config hybrid.NatsConfig) (jetstream.JetStream, error) {
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
