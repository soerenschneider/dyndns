package nats

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// stubConnect replaces connect, failing the first failures calls, and shortens
// the reconnect delays.
func stubConnect(t *testing.T, failures int32) *atomic.Int32 {
	t.Helper()

	calls := &atomic.Int32{}
	origConnect, origMin, origMax := connect, reconnectMinDelay, reconnectMaxDelay
	connect = func(hybrid.NatsConfig) (jetstream.JetStream, error) {
		if calls.Add(1) <= failures {
			return nil, errors.New("connection refused")
		}
		return connectedJetStream{}, nil
	}
	reconnectMinDelay, reconnectMaxDelay = time.Millisecond, 4*time.Millisecond
	t.Cleanup(func() {
		connect, reconnectMinDelay, reconnectMaxDelay = origConnect, origMin, origMax
	})

	return calls
}

func TestConnectInBackground_RetriesUntilConnected(t *testing.T) {
	calls := stubConnect(t, 3)

	var connected atomic.Bool
	done := connectInBackground(context.Background(), hybrid.NatsConfig{}, func(jetstream.JetStream) {
		connected.Store(true)
	})

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("did not connect")
	}
	if !connected.Load() || calls.Load() != 4 {
		t.Errorf("connected = %v after %d attempts, want true after 4", connected.Load(), calls.Load())
	}
}

func TestConnectInBackground_StopsWhenCancelled(t *testing.T) {
	stubConnect(t, 1<<30)

	ctx, cancel := context.WithCancel(context.Background())
	done := connectInBackground(ctx, hybrid.NatsConfig{}, func(jetstream.JetStream) {
		t.Error("connected despite failing connect")
	})
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("still reconnecting after ctx was cancelled")
	}
}

type failingConsumer struct {
	jetstream.Consumer
	fetches atomic.Int32
}

func (f *failingConsumer) Fetch(int, ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	f.fetches.Add(1)
	return nil, errors.New("nats: no responders available")
}

func TestConsume_BacksOffOnFetchErrors(t *testing.T) {
	orig := fetchErrorBackoff
	fetchErrorBackoff = 50 * time.Millisecond
	t.Cleanup(func() { fetchErrorBackoff = orig })

	n := &NatsDyndnsServer{config: &hybrid.NatsConfig{}}
	cons := &failingConsumer{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	n.consume(ctx, cons)

	// Without a backoff, a failing fetch is retried right away, thousands of
	// times in 300ms.
	if got := cons.fetches.Load(); got > 10 {
		t.Errorf("fetched %d times in 300ms, want a pause between failed fetches", got)
	}
}

// Listen used to go on with a nil connection and panic when dyndns shut down
// before nats was reachable.
func TestListen_CancelledBeforeConnected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	n := &NatsDyndnsServer{config: &hybrid.NatsConfig{}, reqChan: make(chan update.UpdateRecordRequest)}

	errChan := make(chan error, 1)
	go func() {
		errChan <- n.Listen(ctx, &sync.WaitGroup{})
	}()
	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Listen() error = %v, want nil on shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Listen() did not return after ctx was cancelled")
	}
}
