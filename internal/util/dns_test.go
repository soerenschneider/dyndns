package util

import (
	"testing"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

func TestRecordMatchesIps(t *testing.T) {
	const v4, v6 = "8.8.8.8", "2a00:1450:4001:830::200e"
	tests := []struct {
		name   string
		record update.DnsRecord
		ips    []string
		want   bool
	}{
		{"ipv4 matches", update.DnsRecord{IpV4: v4}, []string{v4}, true},
		{"ipv4 differs", update.DnsRecord{IpV4: v4}, []string{"1.1.1.1"}, false},
		{"both match", update.DnsRecord{IpV4: v4, IpV6: v6}, []string{v6, v4}, true},
		{"ipv6 missing", update.DnsRecord{IpV4: v4, IpV6: v6}, []string{v4}, false},
		{"ipv6 matches in non-canonical form", update.DnsRecord{IpV6: "2a00:1450:4001:0830:0000:0000:0000:200e"}, []string{v6}, true},
		{"undeclared ipv4 is ignored without withdrawal", update.DnsRecord{IpV6: v6}, []string{"1.1.1.1", v6}, true},
		{"withdrawn ipv4 still present", update.DnsRecord{IpV6: v6, WithdrawnFamilies: []string{update.AddrFamilyIpv4}}, []string{"1.1.1.1", v6}, false},
		{"withdrawn ipv4 gone", update.DnsRecord{IpV6: v6, WithdrawnFamilies: []string{update.AddrFamilyIpv4}}, []string{v6}, true},
		{"withdrawn ipv6 still present", update.DnsRecord{IpV4: v4, WithdrawnFamilies: []string{update.AddrFamilyIpv6}}, []string{v4, "2001:4860::1"}, false},
		{"empty record", update.DnsRecord{}, []string{v4}, false},
		{"withdraw only, host has no records", update.DnsRecord{WithdrawnFamilies: []string{update.AddrFamilyIpv4}}, nil, true},
		{"withdraw only, A record left", update.DnsRecord{WithdrawnFamilies: []string{update.AddrFamilyIpv4}}, []string{v4}, false},
		{"withdraw only ipv4, unmanaged AAAA is ignored", update.DnsRecord{WithdrawnFamilies: []string{update.AddrFamilyIpv4}}, []string{v6}, true},
		{"no dns answers", update.DnsRecord{IpV4: v4}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RecordMatchesIps(&tt.record, tt.ips); got != tt.want {
				t.Errorf("RecordMatchesIps() = %v, want %v", got, tt.want)
			}
		})
	}
}
