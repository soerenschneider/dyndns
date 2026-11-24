package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/v2/internal/notification"
	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

const orderTestHost = "my-host.tld"

type fakePropagator struct {
	calls []update.DnsRecord
	err   error
}

func (f *fakePropagator) PropagateChange(_ context.Context, ip update.DnsRecord) error {
	f.calls = append(f.calls, ip)
	return f.err
}

// newOrderTestServer builds a server whose DNS never matches, so every
// accepted change is propagated.
func newOrderTestServer(p propagator) *DyndnsServer {
	return &DyndnsServer{
		knownHosts:       map[string][]verification.VerificationKey{orderTestHost: {&SimpleVerifier{true}}},
		propagator:       p,
		cache:            map[string]update.DnsRecord{},
		notificationImpl: &notification.DummyNotification{},
		dnsMatches:       func(*update.DnsRecord) bool { return false },
	}
}

func request(ipv4 string, ts time.Time) update.UpdateRecordRequest {
	return update.UpdateRecordRequest{
		PublicIp:  update.DnsRecord{Host: orderTestHost, IpV4: ipv4, Timestamp: ts},
		Signature: "sig",
	}
}

func TestUpdateRecord_Ordering(t *testing.T) {
	now := time.Now()
	type step struct {
		req           update.UpdateRecordRequest
		wantErr       error
		wantPropagate bool
	}
	tests := []struct {
		name  string
		steps []step
	}{
		{
			name: "newer change is applied",
			steps: []step{
				{req: request("8.8.8.8", now.Add(-time.Minute)), wantPropagate: true},
				{req: request("8.8.4.4", now), wantPropagate: true},
			},
		},
		{
			name: "duplicate is not propagated again",
			steps: []step{
				{req: request("8.8.8.8", now.Add(-time.Minute)), wantPropagate: true},
				{req: request("8.8.8.8", now)},
			},
		},
		{
			name: "replayed older request does not roll back",
			steps: []step{
				{req: request("8.8.8.8", now.Add(-10*time.Minute)), wantPropagate: true},
				{req: request("8.8.4.4", now), wantPropagate: true},
				{req: request("8.8.8.8", now.Add(-10*time.Minute)), wantErr: ErrorMessageSuperseded},
			},
		},
		{
			name: "late duplicate of the current record is still rejected as older",
			steps: []step{
				{req: request("8.8.8.8", now), wantPropagate: true},
				{req: request("8.8.8.8", now.Add(-time.Minute)), wantErr: ErrorMessageSuperseded},
			},
		},
		{
			name: "different content within the same second is rejected",
			steps: []step{
				{req: request("8.8.8.8", now.Truncate(time.Second)), wantPropagate: true},
				{req: request("8.8.4.4", now.Truncate(time.Second).Add(500*time.Millisecond)), wantErr: ErrorMessageSuperseded},
			},
		},
		{
			name: "duplicate advances the timestamp used for ordering",
			steps: []step{
				{req: request("8.8.8.8", now.Add(-10*time.Minute)), wantPropagate: true},
				{req: request("8.8.8.8", now.Add(-time.Minute))},
				{req: request("8.8.4.4", now.Add(-5*time.Minute)), wantErr: ErrorMessageSuperseded},
			},
		},
		{
			name: "too old",
			steps: []step{
				{req: request("8.8.8.8", now.Add(-maxMessageAge-time.Minute)), wantErr: ErrorMessageTooOld},
			},
		},
		{
			name: "from the future",
			steps: []step{
				{req: request("8.8.8.8", now.Add(maxClockSkew+time.Minute)), wantErr: ErrorMessageFromFuture},
			},
		},
		{
			name: "slightly ahead clock is tolerated",
			steps: []step{
				{req: request("8.8.8.8", now.Add(maxClockSkew-time.Minute)), wantPropagate: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &fakePropagator{}
			server := newOrderTestServer(p)
			for i, s := range tt.steps {
				before := len(p.calls)
				err := server.UpdateRecord(context.Background(), s.req)
				if !errors.Is(err, s.wantErr) {
					t.Fatalf("step %d: UpdateRecord() error = %v, want %v", i, err, s.wantErr)
				}
				if propagated := len(p.calls) > before; propagated != s.wantPropagate {
					t.Fatalf("step %d: propagated = %v, want %v", i, propagated, s.wantPropagate)
				}
			}
		})
	}
}

func TestUpdateRecord_FailedPropagationIsRetried(t *testing.T) {
	p := &fakePropagator{err: errors.New("route53 down")}
	server := newOrderTestServer(p)

	if err := server.UpdateRecord(context.Background(), request("8.8.8.8", time.Now().Add(-time.Minute))); err == nil {
		t.Fatal("expected propagation error")
	}

	p.err = nil
	if err := server.UpdateRecord(context.Background(), request("8.8.8.8", time.Now())); err != nil {
		t.Fatalf("retry: UpdateRecord() error = %v", err)
	}
	if len(p.calls) != 2 {
		t.Errorf("propagations = %d, want 2: a failed propagation must not be cached", len(p.calls))
	}
}

func TestUpdateRecord_WithdrawOnly(t *testing.T) {
	withdrawOnly := func(ts time.Time) update.UpdateRecordRequest {
		return update.UpdateRecordRequest{
			PublicIp:  update.DnsRecord{Host: orderTestHost, WithdrawnFamilies: []string{update.AddrFamilyIpv4}, Timestamp: ts},
			Signature: "sig",
		}
	}

	t.Run("stale A record is withdrawn", func(t *testing.T) {
		p := &fakePropagator{}
		server := newOrderTestServer(p)

		if err := server.UpdateRecord(context.Background(), withdrawOnly(time.Now())); err != nil {
			t.Fatalf("UpdateRecord() error = %v", err)
		}
		if len(p.calls) != 1 || !p.calls[0].IsWithdrawn(update.AddrFamilyIpv4) {
			t.Fatalf("propagated %+v, want one withdrawal of ip4", p.calls)
		}
	})

	t.Run("host already has no records", func(t *testing.T) {
		p := &fakePropagator{}
		server := newOrderTestServer(p)
		server.dnsMatches = func(*update.DnsRecord) bool { return true }

		if err := server.UpdateRecord(context.Background(), withdrawOnly(time.Now())); err != nil {
			t.Fatalf("UpdateRecord() error = %v", err)
		}
		if len(p.calls) != 0 {
			t.Errorf("propagations = %d, want 0", len(p.calls))
		}
		if _, ok := server.cache[orderTestHost]; !ok {
			t.Error("withdraw-only request must be recorded for ordering")
		}
	})

	t.Run("replayed older address does not undo the withdrawal", func(t *testing.T) {
		p := &fakePropagator{}
		server := newOrderTestServer(p)

		if err := server.UpdateRecord(context.Background(), withdrawOnly(time.Now())); err != nil {
			t.Fatalf("UpdateRecord() error = %v", err)
		}
		if err := server.UpdateRecord(context.Background(), request("8.8.8.8", time.Now().Add(-time.Minute))); !errors.Is(err, ErrorMessageSuperseded) {
			t.Errorf("UpdateRecord() error = %v, want %v", err, ErrorMessageSuperseded)
		}
	})
}
