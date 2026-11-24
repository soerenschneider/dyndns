//go:build linux

package util

import (
	"fmt"
	"net/netip"
	"sort"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// globalUnicast is the range IANA has actually delegated for global use.
// ULA (fc00::/7) and link-local (fe80::/10) fall outside it.
var globalUnicast = netip.MustParsePrefix("2000::/3")

var notRoutable = []netip.Prefix{
	netip.MustParsePrefix("2001::/32"),     // Teredo
	netip.MustParsePrefix("2001:db8::/32"), // documentation
	netip.MustParsePrefix("2002::/16"),     // 6to4
}

// LocalGlobalIpV6Addr returns the address on iface that is suitable for an AAAA
// record: globally routable, preferred, and stable across privacy-address
// rotation.
func LocalGlobalIpV6Addr(iface string) (netip.Addr, error) {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("looking up interface %q: %w", iface, err)
	}

	addrs, err := netlink.AddrList(link, unix.AF_INET6)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("listing addresses on %q: %w", iface, err)
	}

	const reject = unix.IFA_F_TEMPORARY | // rotates; do not publish
		unix.IFA_F_DEPRECATED | // valid but not for new connections
		unix.IFA_F_TENTATIVE | // DAD not finished
		unix.IFA_F_DADFAILED

	var candidates []netlink.Addr
	for _, a := range addrs {
		if a.Flags&reject != 0 {
			continue
		}
		if a.Scope != unix.RT_SCOPE_UNIVERSE {
			continue
		}

		ip, ok := netip.AddrFromSlice(a.IP)
		if !ok {
			continue
		}
		ip = ip.Unmap()

		if !globalUnicast.Contains(ip) {
			continue
		}
		if slicesContainsPrefix(notRoutable, ip) {
			continue
		}

		candidates = append(candidates, a)
	}

	if len(candidates) == 0 {
		return netip.Addr{}, fmt.Errorf("no global IPv6 address on %q", iface)
	}

	// Prefer a statically configured address over a SLAAC one, then sort by
	// address so the choice is deterministic when a host has several prefixes.
	sort.SliceStable(candidates, func(i, j int) bool {
		pi := candidates[i].Flags&unix.IFA_F_PERMANENT != 0
		pj := candidates[j].Flags&unix.IFA_F_PERMANENT != 0
		if pi != pj {
			return pi
		}
		return candidates[i].IP.String() < candidates[j].IP.String()
	})

	ip, _ := netip.AddrFromSlice(candidates[0].IP)
	return ip.Unmap(), nil
}

func slicesContainsPrefix(prefixes []netip.Prefix, ip netip.Addr) bool {
	for _, p := range prefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}
