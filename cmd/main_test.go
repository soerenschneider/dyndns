package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
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
