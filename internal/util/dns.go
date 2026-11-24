package util

import (
	"errors"
	"fmt"
	"net"
	"net/netip"

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

	wantV4, _ := netip.ParseAddr(record.IpV4)
	wantV6, _ := netip.ParseAddr(record.IpV6)

	var ipv4Confirmed, ipv6Confirmed, staleWithdrawn bool
	for _, ip := range ips {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			continue
		}
		addr = addr.Unmap()

		switch {
		case addr.Is4() && record.IsWithdrawn(update.AddrFamilyIpv4),
			addr.Is6() && record.IsWithdrawn(update.AddrFamilyIpv6):
			staleWithdrawn = true
		case record.HasIpV4() && addr == wantV4:
			ipv4Confirmed = true
		case record.HasIpV6() && addr == wantV6:
			ipv6Confirmed = true
		}
	}

	return !staleWithdrawn &&
		(!record.HasIpV4() || ipv4Confirmed) &&
		(!record.HasIpV6() || ipv6Confirmed)
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
