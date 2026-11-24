package resolvers

import (
	"errors"
	"testing"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type fakeResolver struct {
	record *update.DnsRecord
	err    error
}

func (f *fakeResolver) Resolve() (*update.DnsRecord, error) { return f.record, f.err }
func (f *fakeResolver) Name() string                        { return "fake" }
func (f *fakeResolver) Host() string                        { return "my.host.tld" }

func TestFamilyResolver_Resolve(t *testing.T) {
	const v4, v6 = "8.8.8.8", "2a00:1450:4001:830::200e"
	both := &update.DnsRecord{IpV4: v4, IpV6: v6}
	failing := &fakeResolver{err: errors.New("boom")}

	tests := []struct {
		name    string
		ipv4    IpResolver
		ipv6    IpResolver
		wantV4  string
		wantV6  string
		wantErr bool
	}{
		{
			name:   "each family taken from its own resolver only",
			ipv4:   &fakeResolver{record: both},
			ipv6:   &fakeResolver{record: &update.DnsRecord{IpV4: "1.1.1.1", IpV6: v6}},
			wantV4: v4,
			wantV6: v6,
		},
		{
			name:   "ipv6 resolver fails, ipv4 kept",
			ipv4:   &fakeResolver{record: both},
			ipv6:   failing,
			wantV4: v4,
		},
		{
			name:   "ipv4 resolver fails, ipv6 kept",
			ipv4:   failing,
			ipv6:   &fakeResolver{record: both},
			wantV6: v6,
		},
		{
			name:    "both fail",
			ipv4:    failing,
			ipv6:    failing,
			wantErr: true,
		},
		{
			name:    "both succeed without addresses",
			ipv4:    &fakeResolver{record: &update.DnsRecord{}},
			ipv6:    &fakeResolver{record: &update.DnsRecord{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := NewFamilyResolver("my.host.tld", tt.ipv4, tt.ipv6)
			if err != nil {
				t.Fatalf("NewFamilyResolver() error = %v", err)
			}

			got, err := resolver.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.IpV4 != tt.wantV4 || got.IpV6 != tt.wantV6 {
				t.Errorf("Resolve() = (%q, %q), want (%q, %q)", got.IpV4, got.IpV6, tt.wantV4, tt.wantV6)
			}
			if got.Host != "my.host.tld" || got.Timestamp.IsZero() {
				t.Errorf("Resolve() host/timestamp not set: %+v", got)
			}
		})
	}
}

func TestNewFamilyResolver_RequiresBoth(t *testing.T) {
	if _, err := NewFamilyResolver("my.host.tld", &fakeResolver{}, nil); err == nil {
		t.Error("expected error for missing ipv6 resolver")
	}
	if _, err := NewFamilyResolver("my.host.tld", nil, &fakeResolver{}); err == nil {
		t.Error("expected error for missing ipv4 resolver")
	}
}
