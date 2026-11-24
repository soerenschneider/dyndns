package util

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// HostnameMatchesRecord reports whether the DNS records of record.Host match
// record, see RecordMatchesIps.
func HostnameMatchesRecord(record *update.DnsRecord) bool {
	ips, err := LookupDns(record.Host)
	if err != nil {
		// A host without any records is the goal of a withdraw-only record.
		var dnsErr *net.DNSError
		if !errors.As(err, &dnsErr) || !dnsErr.IsNotFound {
			log.Info().Msgf("Error looking up dns record %s: %v", record.Host, err)
			return false
		}
	}

	return RecordMatchesIps(record, ips)
}

// RecordMatchesIps reports whether ips, the addresses a host resolves to, match
// record: every address the record declares must be present, and no address
// of a withdrawn family may be left over.
func RecordMatchesIps(record *update.DnsRecord, ips []string) bool {
	if !record.HasIpV4() && !record.HasIpV6() && len(record.WithdrawnFamilies) == 0 {
		return false
	}

	var gotV4, gotV6 []netip.Addr
	for _, ip := range ips {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			continue
		}

		addr = addr.Unmap()
		if addr.Is4() {
			gotV4 = append(gotV4, addr)
		} else {
			gotV6 = append(gotV6, addr)
		}
	}

	return familyMatches(record.HasIpV4(), record.IpV4, record.IsWithdrawn(update.AddrFamilyIpv4), gotV4) &&
		familyMatches(record.HasIpV6(), record.IpV6, record.IsWithdrawn(update.AddrFamilyIpv6), gotV6)
}

// familyMatches reports whether got, the addresses of one family a host
// resolves to, match what a record declares for that family.
func familyMatches(declared bool, want string, withdrawn bool, got []netip.Addr) bool {
	if withdrawn {
		return len(got) == 0
	}

	if !declared {
		return true
	}

	wantAddr, err := netip.ParseAddr(want)
	return err == nil && slices.Contains(got, wantAddr.Unmap())
}

func LookupDns(host string) ([]string, error) {
	response, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve host %s: %w", host, err)
	}

	var ips []string
	for _, ip := range response {
		ips = append(ips, ip.String())
	}

	return ips, nil
}
