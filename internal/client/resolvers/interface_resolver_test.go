package resolvers

import (
	"net"
	"testing"

	"github.com/soerenschneider/dyndns/v2/internal"
)

func ipNet(s string) net.Addr {
	ip, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	n.IP = ip
	return n
}

func TestPickIpV4(t *testing.T) {
	tests := []struct {
		name  string
		addrs []net.Addr
		want  string
	}{
		{
			name: "no addresses",
			want: "",
		},
		{
			name:  "only ipv6",
			addrs: []net.Addr{ipNet("2a00:1450:4001:830::200e/64"), ipNet("fe80::1/64")},
			want:  "",
		},
		{
			name:  "single public ipv4",
			addrs: []net.Addr{ipNet("8.8.8.8/24")},
			want:  "8.8.8.8",
		},
		{
			name:  "public ipv4 preferred over earlier private one",
			addrs: []net.Addr{ipNet("192.168.1.10/24"), ipNet("fe80::1/64"), ipNet("8.8.8.8/24")},
			want:  "8.8.8.8",
		},
		{
			name:  "only cgnat, returned so the client can report it",
			addrs: []net.Addr{ipNet("fe80::1/64"), ipNet("100.64.12.34/10")},
			want:  "100.64.12.34",
		},
		{
			name:  "first non-public ipv4 is the fallback",
			addrs: []net.Addr{ipNet("100.64.12.34/10"), ipNet("192.168.1.10/24")},
			want:  "100.64.12.34",
		},
		{
			name:  "non IPNet addresses are skipped",
			addrs: []net.Addr{&net.IPAddr{IP: net.ParseIP("1.1.1.1")}, ipNet("8.8.8.8/24")},
			want:  "8.8.8.8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickIpV4(tt.addrs); got != tt.want {
				t.Errorf("pickIpV4() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterfaceWithIp(t *testing.T) {
	loopback, err := net.InterfaceByName(loopbackName(t))
	if err != nil {
		t.Skipf("no loopback interface: %v", err)
	}
	ifaces := []net.Interface{*loopback}

	got, err := interfaceWithIp(ifaces, net.ParseIP("127.0.0.1"))
	if err != nil || got != loopback.Name {
		t.Errorf("interfaceWithIp(127.0.0.1) = %q, %v, want %q", got, err, loopback.Name)
	}

	if got, err := interfaceWithIp(ifaces, net.ParseIP("192.0.2.1")); err == nil {
		t.Errorf("interfaceWithIp(192.0.2.1) = %q, want an error", got)
	}
}

func TestDefaultRouteInterface(t *testing.T) {
	if _, err := defaultRouteInterface("ip5"); err == nil {
		t.Error("defaultRouteInterface(ip5) wants an error for an unknown address family")
	}

	// Depends on the host's routes, so only check consistency when there is one.
	got, err := defaultRouteInterface(internal.AddrFamilyIpv4)
	if err != nil {
		t.Skipf("no ipv4 default route: %v", err)
	}
	if _, err := net.InterfaceByName(got); err != nil {
		t.Errorf("defaultRouteInterface() = %q, not an interface: %v", got, err)
	}
}

func loopbackName(t *testing.T) string {
	t.Helper()
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skipf("could not list interfaces: %v", err)
	}
	for _, ief := range ifaces {
		if ief.Flags&net.FlagLoopback != 0 {
			return ief.Name
		}
	}
	t.Skip("no loopback interface")
	return ""
}
