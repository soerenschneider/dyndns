package resolvers

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/util"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type InterfaceResolver struct {
	watchedInterface string
	addrFamilies     []string
	host             string
}

func NewInterfaceResolver(watchedInterface, host string, addrFamilies []string) (*InterfaceResolver, error) {
	if len(addrFamilies) == 0 {
		return nil, errors.New("empty addressFamily slice provided")
	}

	return &InterfaceResolver{
		watchedInterface: watchedInterface,
		host:             host,
		addrFamilies:     addrFamilies,
	}, nil
}

func (resolver *InterfaceResolver) Name() string {
	return fmt.Sprintf("InterfaceResolver (%s)", resolver.watchedInterface)
}

func (resolver *InterfaceResolver) Host() string {
	return resolver.host
}

func (resolver *InterfaceResolver) Resolve() (*update.DnsRecord, error) {
	ief, err := net.InterfaceByName(resolver.watchedInterface)
	if err != nil {
		return nil, fmt.Errorf("could not get interface %q: %w", resolver.watchedInterface, err)
	}

	record := &update.DnsRecord{
		Host:      resolver.host,
		Timestamp: time.Now(),
	}

	if slices.Contains(resolver.addrFamilies, internal.AddrFamilyIpv4) {
		addrs, err := ief.Addrs()
		if err != nil {
			return nil, fmt.Errorf("could not get addresses of interface %q: %w", resolver.watchedInterface, err)
		}
		record.IpV4 = pickIpV4(addrs)
	}

	if slices.Contains(resolver.addrFamilies, internal.AddrFamilyIpv6) {
		// A missing IPv6 address is not an error, the record can still carry an IPv4.
		ipv6, err := util.LocalGlobalIpV6Addr(resolver.watchedInterface)
		if err != nil {
			log.Debug().Err(err).Str("component", "interface_resolver").Str("interface", resolver.watchedInterface).Msg("no usable ipv6 address")
		} else {
			record.IpV6 = ipv6.String()
		}
	}

	if !record.HasIpV4() && !record.HasIpV6() {
		return nil, fmt.Errorf("no address found on interface %q for address families %v", resolver.watchedInterface, resolver.addrFamilies)
	}

	return record, nil
}

// pickIpV4 returns the first publicly routable IPv4 address of addrs. If there
// is none it returns the first IPv4 address at all, so the caller can report why
// it is unusable (e.g. CGNAT) instead of seeing nothing.
func pickIpV4(addrs []net.Addr) string {
	var fallback string
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}

		if update.IsPublicIpV4(ip.String()) {
			return ip.String()
		}

		if fallback == "" {
			fallback = ip.String()
		}
	}

	return fallback
}
