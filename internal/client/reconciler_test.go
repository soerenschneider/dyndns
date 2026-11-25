package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type fakeUpdater struct {
	calls   atomic.Int32
	err     error
	barrier *sync.WaitGroup
}

func (f *fakeUpdater) UpdateRecord(context.Context, update.UpdateRecordRequest) error {
	f.calls.Add(1)
	if f.barrier != nil {
		// Wait for all dispatchers, so they finish at the same time.
		f.barrier.Done()
		f.barrier.Wait()
	}
	return f.err
}

func request() *update.UpdateRecordRequest {
	return &update.UpdateRecordRequest{PublicIp: update.DnsRecord{Host: "home.example.com"}}
}

// Run with -race: concurrent successful dispatchers used to write to the
// pending map at the same time, which crashed the client.
func TestReconciler_ConcurrentSuccessfulDispatchers(t *testing.T) {
	const numDispatchers = 8

	for i := 0; i < 20; i++ {
		barrier := &sync.WaitGroup{}
		barrier.Add(numDispatchers)

		fakes := map[string]*fakeUpdater{}
		updaters := map[string]RecordUpdater{}
		for d := 0; d < numDispatchers; d++ {
			fake := &fakeUpdater{barrier: barrier}
			key := fmt.Sprintf("dispatcher-%d", d)
			fakes[key] = fake
			updaters[key] = fake
		}

		r, err := NewReconciler(updaters, false)
		if err != nil {
			t.Fatal(err)
		}

		if err := r.RegisterUpdate(context.Background(), request()); err != nil {
			t.Fatalf("RegisterUpdate() error = %v", err)
		}
		if len(r.pendingChanges) != 0 {
			t.Fatalf("pending after all dispatchers succeeded: %v", r.pendingChanges)
		}

		// Nothing is pending, so nothing is sent again.
		if err := r.dispatch(context.Background()); err != nil {
			t.Fatal(err)
		}
		for key, fake := range fakes {
			if got := fake.calls.Load(); got != 1 {
				t.Errorf("%s called %d times, want 1", key, got)
			}
		}
	}
}

func TestReconciler_FailedDispatcherStaysPending(t *testing.T) {
	ok := &fakeUpdater{}
	failing := &fakeUpdater{err: errors.New("broker down")}
	r, err := NewReconciler(map[string]RecordUpdater{"ok": ok, "failing": failing}, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := r.RegisterUpdate(context.Background(), request()); err == nil {
		t.Fatal("RegisterUpdate() error = nil, want the failing dispatcher's error")
	}
	if _, pending := r.pendingChanges["failing"]; !pending || len(r.pendingChanges) != 1 {
		t.Fatalf("pending = %v, want only the failing dispatcher", r.pendingChanges)
	}

	// The next run only retries the failing dispatcher.
	failing.err = nil
	if err := r.dispatch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(r.pendingChanges) != 0 {
		t.Errorf("pending after retry = %v, want none", r.pendingChanges)
	}
	if ok.calls.Load() != 1 || failing.calls.Load() != 2 {
		t.Errorf("calls ok=%d failing=%d, want 1 and 2", ok.calls.Load(), failing.calls.Load())
	}
}

func TestReconciler_StopAfterFirstSuccess(t *testing.T) {
	failing := &fakeUpdater{err: errors.New("broker down")}
	r, err := NewReconciler(map[string]RecordUpdater{"ok": &fakeUpdater{}, "failing": failing}, true)
	if err != nil {
		t.Fatal(err)
	}

	_ = r.RegisterUpdate(context.Background(), request())
	if len(r.pendingChanges) != 0 {
		t.Errorf("pending = %v, want none after a successful dispatch", r.pendingChanges)
	}
}

func TestReconciler_RunStopsWhenCancelled(t *testing.T) {
	r, err := NewReconciler(map[string]RecordUpdater{"ok": &fakeUpdater{}}, false)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after ctx was cancelled")
	}
}
