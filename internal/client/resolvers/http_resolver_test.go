package resolvers

import (
	"net"
	"testing"

	"github.com/soerenschneider/dyndns/v2/internal"
)

func Test_repair(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no need",
			args: args{
				body: "nonewline",
			},
			want: "nonewline",
		},
		{
			name: "replace newline",
			args: args{
				body: "nonewline\n",
			},
			want: "nonewline",
		},
		{
			name: "replace newline start",
			args: args{
				body: "\nnonewline",
			},
			want: "nonewline",
		},
		{
			name: "replace newlines",
			args: args{
				body: "\nnonewline\n",
			},
			want: "nonewline",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repair(tt.args.body); got != tt.want {
				t.Errorf("repair() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCgnatLocalIpV4(t *testing.T) {
	tests := []struct {
		name      string
		family    string
		localAddr net.Addr
		wantIp    string
		wantOk    bool
	}{
		{"cgnat ipv4", internal.AddrFamilyIpv4, &net.TCPAddr{IP: net.ParseIP("100.64.12.34")}, "100.64.12.34", true},
		{"private ipv4 behind regular nat", internal.AddrFamilyIpv4, &net.TCPAddr{IP: net.ParseIP("192.168.1.10")}, "192.168.1.10", false},
		{"public ipv4", internal.AddrFamilyIpv4, &net.TCPAddr{IP: net.ParseIP("8.8.8.8")}, "8.8.8.8", false},
		{"ipv6 family is never cgnat", internal.AddrFamilyIpv6, &net.TCPAddr{IP: net.ParseIP("100.64.12.34")}, "", false},
		{"not a tcp address", internal.AddrFamilyIpv4, &net.UDPAddr{IP: net.ParseIP("100.64.12.34")}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIp, gotOk := cgnatLocalIpV4(tt.family, tt.localAddr)
			if gotOk != tt.wantOk || (tt.wantOk && gotIp != tt.wantIp) {
				t.Errorf("cgnatLocalIpV4() = (%q, %v), want (%q, %v)", gotIp, gotOk, tt.wantIp, tt.wantOk)
			}
		})
	}
}
