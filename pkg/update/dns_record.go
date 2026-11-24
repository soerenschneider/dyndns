package update

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strings"
	"time"
)

type UpdateRecordRequest struct {
	PublicIp  DnsRecord `json:"public_ip"`
	Signature string    `json:"signature"`
}

func (r *UpdateRecordRequest) Validate() error {
	if strings.TrimSpace(r.Signature) == "" {
		return errors.New("signature is missing")
	}

	return r.PublicIp.Validate()
}

const (
	AddrFamilyIpv4 = "ip4"
	AddrFamilyIpv6 = "ip6"
)

type DnsRecord struct {
	IpV4 string `json:"ipv4,omitempty"`
	IpV6 string `json:"ipv6,omitempty"`
	// WithdrawnFamilies lists the address families ("ip4", "ip6") the client
	// determined to have no publicly routable address, e.g. an IPv4 behind
	// CGNAT. Their records are removed instead of being left pointing at a
	// stale address.
	WithdrawnFamilies []string  `json:"withdrawn_families,omitempty"`
	Host              string    `json:"host"`
	Timestamp         time.Time `json:"timestamp"`
}

func NewResolvedIp(host string) *DnsRecord {
	return &DnsRecord{
		Host:      host,
		Timestamp: time.Now(),
	}
}

// Equals checks for equality and ignores timestamps
func (resolved *DnsRecord) Equals(ip *DnsRecord) bool {
	if ip == nil || resolved == nil {
		return false
	}

	if resolved.IpV4 != ip.IpV4 {
		return false
	}

	if resolved.IpV6 != ip.IpV6 {
		return false
	}

	if resolved.Host != ip.Host {
		return false
	}

	return slices.Equal(resolved.sortedWithdrawnFamilies(), ip.sortedWithdrawnFamilies())
}

// Withdraw marks addrFamily as having no publicly routable address.
func (resolved *DnsRecord) Withdraw(addrFamily string) {
	if !resolved.IsWithdrawn(addrFamily) {
		resolved.WithdrawnFamilies = append(resolved.WithdrawnFamilies, addrFamily)
	}
}

func (resolved *DnsRecord) IsWithdrawn(addrFamily string) bool {
	return slices.Contains(resolved.WithdrawnFamilies, addrFamily)
}

func (resolved *DnsRecord) sortedWithdrawnFamilies() []string {
	sorted := slices.Clone(resolved.WithdrawnFamilies)
	slices.Sort(sorted)
	return sorted
}

func (resolved *DnsRecord) HasIpV6() bool {
	return len(resolved.IpV6) > 0
}

func (resolved *DnsRecord) HasIpV4() bool {
	return len(resolved.IpV4) > 0
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

// nonPublicV4 lists special-purpose IPv4 ranges that netip's IsPrivate and
// IsGlobalUnicast do not already exclude.
var nonPublicV4 = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // "this network"
	cgnat,                                    // CGNAT
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // documentation
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"), // documentation
	netip.MustParsePrefix("203.0.113.0/24"),  // documentation
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved
}

// globalUnicastV6 is the range IANA has delegated for global use; ULA,
// link-local, multicast and v4-mapped addresses all fall outside it.
var globalUnicastV6 = netip.MustParsePrefix("2000::/3")

var nonPublicV6 = []netip.Prefix{
	netip.MustParsePrefix("2001::/32"),     // Teredo
	netip.MustParsePrefix("2001:2::/48"),   // benchmarking
	netip.MustParsePrefix("2001:10::/28"),  // ORCHID
	netip.MustParsePrefix("2001:db8::/32"), // documentation
	netip.MustParsePrefix("2002::/16"),     // 6to4
	netip.MustParsePrefix("3fff::/20"),     // documentation
}

// IsPublicIpV4 reports whether ip is an IPv4 address that is routable on the
// public internet.
func IsPublicIpV4(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is4() {
		return false
	}

	if !addr.IsGlobalUnicast() || addr.IsPrivate() {
		return false
	}

	return !containsAddr(nonPublicV4, addr)
}

// IsPublicIpV6 reports whether ip is an IPv6 address that is routable on the
// public internet.
func IsPublicIpV6(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is6() || addr.Is4In6() || addr.Zone() != "" {
		return false
	}

	if !globalUnicastV6.Contains(addr) {
		return false
	}

	return !containsAddr(nonPublicV6, addr)
}

// IsCGNAT reports whether ip lies in the carrier-grade NAT range 100.64.0.0/10.
func IsCGNAT(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	return err == nil && cgnat.Contains(addr)
}

func containsAddr(prefixes []netip.Prefix, addr netip.Addr) bool {
	for _, p := range prefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// IsValid reports whether the record holds at least one address or withdrawn
// family, and every address it holds is publicly routable for its family.
func (resolved *DnsRecord) IsValid() bool {
	if !resolved.HasIpV4() && !resolved.HasIpV6() && len(resolved.WithdrawnFamilies) == 0 {
		return false
	}

	if resolved.HasIpV4() && !IsPublicIpV4(resolved.IpV4) {
		return false
	}

	if resolved.HasIpV6() && !IsPublicIpV6(resolved.IpV6) {
		return false
	}

	return true
}

// DropNonPublic clears every address on the record that is not publicly
// routable, e.g. a CGNAT IPv4 next to a usable IPv6, and returns the addresses
// it removed.
func (resolved *DnsRecord) DropNonPublic() (droppedV4, droppedV6 string) {
	if resolved.HasIpV4() && !IsPublicIpV4(resolved.IpV4) {
		droppedV4, resolved.IpV4 = resolved.IpV4, ""
	}

	if resolved.HasIpV6() && !IsPublicIpV6(resolved.IpV6) {
		droppedV6, resolved.IpV6 = resolved.IpV6, ""
	}

	return droppedV4, droppedV6
}

func (resolved *DnsRecord) String() string {
	var parts []string
	if resolved.HasIpV4() {
		parts = append(parts, resolved.IpV4+" (v4)")
	}
	if resolved.HasIpV6() {
		parts = append(parts, resolved.IpV6+" (v6)")
	}
	for _, family := range resolved.sortedWithdrawnFamilies() {
		parts = append(parts, family+" withdrawn")
	}

	return fmt.Sprintf("%s: %s", resolved.Host, strings.Join(parts, ", "))
}

func (resolved *DnsRecord) Hash() string {
	value := fmt.Sprintf("%d%s%s%s", resolved.Timestamp.Unix(), resolved.Host, resolved.IpV4, resolved.IpV6)
	// Only appended when set, so records without withdrawn families hash (and
	// verify) exactly as before.
	if len(resolved.WithdrawnFamilies) > 0 {
		value += "withdrawn:" + strings.Join(resolved.sortedWithdrawnFamilies(), ",")
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func (resolved *DnsRecord) Validate() error {
	if len(resolved.Host) == 0 {
		return errors.New("domain is missing")
	}

	if len(resolved.IpV4) == 0 && len(resolved.IpV6) == 0 && len(resolved.WithdrawnFamilies) == 0 {
		return errors.New("both ipv4 and ipv6 are empty and nothing is withdrawn")
	}

	if resolved.HasIpV4() && !IsPublicIpV4(resolved.IpV4) {
		return fmt.Errorf("ipv4 %q is not a publicly routable ipv4 address", resolved.IpV4)
	}

	if resolved.HasIpV6() && !IsPublicIpV6(resolved.IpV6) {
		return fmt.Errorf("ipv6 %q is not a publicly routable ipv6 address", resolved.IpV6)
	}

	for _, family := range resolved.WithdrawnFamilies {
		switch {
		case family != AddrFamilyIpv4 && family != AddrFamilyIpv6:
			return fmt.Errorf("unknown withdrawn address family %q", family)
		case family == AddrFamilyIpv4 && resolved.HasIpV4():
			return errors.New("ipv4 is both set and withdrawn")
		case family == AddrFamilyIpv6 && resolved.HasIpV6():
			return errors.New("ipv6 is both set and withdrawn")
		}
	}

	if resolved.Timestamp.IsZero() {
		return fmt.Errorf("timestamp empty")
	}

	return nil
}
