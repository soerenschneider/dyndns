package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
	"github.com/soerenschneider/dyndns/v2/internal/server"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// fakeApp stops shutdownDelay after ctx is done, or never if ignoreCtx is set.
type fakeApp struct {
	shutdownDelay time.Duration
	ignoreCtx     bool
	stopped       atomic.Bool
}

func (f *fakeApp) StartApp(ctx context.Context) error {
	if f.ignoreCtx {
		select {}
	}
	<-ctx.Done()
	time.Sleep(f.shutdownDelay)
	f.stopped.Store(true)
	return nil
}

func TestRun_WaitsForShutdown(t *testing.T) {
	app := &fakeApp{shutdownDelay: 50 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := run(ctx, app); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !app.stopped.Load() {
		t.Error("run() returned before the app had shut down")
	}
}

func TestRun_ShutdownTimeout(t *testing.T) {
	orig := shutdownTimeout
	shutdownTimeout = 50 * time.Millisecond
	t.Cleanup(func() { shutdownTimeout = orig })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := run(ctx, &fakeApp{ignoreCtx: true}); err == nil {
		t.Fatal("run() error = nil, want a timeout for an app that does not stop")
	}
}

func TestEdgeMode_WaitsForServerAndClient(t *testing.T) {
	server := &fakeApp{shutdownDelay: 10 * time.Millisecond}
	client := &fakeApp{shutdownDelay: 50 * time.Millisecond}
	app := appEdgeMode{serverApp: server, clientApp: client}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := app.StartApp(ctx); err != nil {
		t.Fatal(err)
	}
	if !server.stopped.Load() || !client.stopped.Load() {
		t.Errorf("returned with server stopped=%v, client stopped=%v; want both", server.stopped.Load(), client.stopped.Load())
	}
}

// fakeListener fails right away if err is set, otherwise it listens until ctx is done.
type fakeListener struct {
	err error
}

func (f *fakeListener) Listen(ctx context.Context, _ *sync.WaitGroup) error {
	if f.err != nil {
		return f.err
	}
	<-ctx.Done()
	return nil
}

type noopPropagator struct{}

func (noopPropagator) PropagateChange(context.Context, update.DnsRecord) error { return nil }

func newTestServerApp(t *testing.T, listeners ...listener) appServerMode {
	t.Helper()

	config := conf.GetDefaultConfig()
	config.Mode = internal.ServerMode
	requests := make(chan update.UpdateRecordRequest)
	dyndnsServer, err := server.NewServer(*config, noopPropagator{}, requests, nil)
	if err != nil {
		t.Fatal(err)
	}

	return appServerMode{server: dyndnsServer, listeners: listeners, requests: requests}
}

func TestServerMode_SurvivesFailingListener(t *testing.T) {
	app := newTestServerApp(t, &fakeListener{err: errors.New("broker unreachable")}, &fakeListener{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.StartApp(ctx)
	}()

	select {
	case err := <-done:
		t.Fatalf("StartApp() returned %v while a listener was still running", err)
	case <-time.After(200 * time.Millisecond):
	}

	cancel()
	if err := <-done; err != nil {
		t.Errorf("StartApp() error = %v, want nil on shutdown", err)
	}
}

func TestServerMode_FailsWhenAllListenersFail(t *testing.T) {
	app := newTestServerApp(t, &fakeListener{err: errors.New("broker unreachable")}, &fakeListener{err: errors.New("port in use")})

	done := make(chan error, 1)
	go func() {
		done <- app.StartApp(context.Background())
	}()

	select {
	case err := <-done:
		if !errors.Is(err, errAllListenersFailed) {
			t.Errorf("StartApp() error = %v, want %v", err, errAllListenersFailed)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("StartApp() kept running after all listeners failed")
	}
}
