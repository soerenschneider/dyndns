package conf

import (
	"testing"

	"github.com/soerenschneider/dyndns/v2/internal"
)

func TestClientConf_ResolverFor(t *testing.T) {
	tests := []struct {
		name   string
		conf   ClientConf
		family string
		want   string
	}{
		{"default is http", ClientConf{}, internal.AddrFamilyIpv4, internal.ResolverHttp},
		{"interface set, defaults to interface", ClientConf{NetworkInterface: "eth0"}, internal.AddrFamilyIpv6, internal.ResolverInterface},
		{"explicit ipv4 wins over interface", ClientConf{NetworkInterface: "eth0", Ipv4Resolver: "http"}, internal.AddrFamilyIpv4, internal.ResolverHttp},
		{"explicit ipv4 does not affect ipv6", ClientConf{NetworkInterface: "eth0", Ipv4Resolver: "http"}, internal.AddrFamilyIpv6, internal.ResolverInterface},
		{"explicit ipv6", ClientConf{NetworkInterface: "eth0", Ipv6Resolver: "interface"}, internal.AddrFamilyIpv6, internal.ResolverInterface},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conf.ResolverFor(tt.family); got != tt.want {
				t.Errorf("ResolverFor(%q) = %q, want %q", tt.family, got, tt.want)
			}
		})
	}
}

func TestClientConf_ValidateResolvers(t *testing.T) {
	both := []string{internal.AddrFamilyIpv4, internal.AddrFamilyIpv6}
	tests := []struct {
		name    string
		conf    ClientConf
		wantErr bool
	}{
		{
			name: "http for both",
			conf: ClientConf{Host: "my.host.tld", AddrFamilies: both},
		},
		{
			name: "ipv4 via http, ipv6 via interface",
			conf: ClientConf{Host: "my.host.tld", AddrFamilies: both, NetworkInterface: "eth0", Ipv4Resolver: "http", Ipv6Resolver: "interface"},
		},
		{
			name: "ipv4 via http, ipv6 via interface by default",
			conf: ClientConf{Host: "my.host.tld", AddrFamilies: both, NetworkInterface: "eth0", Ipv4Resolver: "http"},
		},
		{
			name: "interface resolver without interface uses the default route",
			conf: ClientConf{Host: "my.host.tld", AddrFamilies: both, Ipv6Resolver: "interface"},
		},
		{
			name:    "unknown resolver",
			conf:    ClientConf{Host: "my.host.tld", AddrFamilies: both, Ipv4Resolver: "dns"},
			wantErr: true,
		},
		{
			name:    "resolver for disabled family",
			conf:    ClientConf{Host: "my.host.tld", AddrFamilies: []string{internal.AddrFamilyIpv4}, NetworkInterface: "eth0", Ipv6Resolver: "interface"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateConfig(tt.conf); (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
