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
	if resolver.watchedInterface == "" {
		return "InterfaceResolver (default route)"
	}
	return fmt.Sprintf("InterfaceResolver (%s)", resolver.watchedInterface)
}

func (resolver *InterfaceResolver) Host() string {
	return resolver.host
}

// interfaceFor returns the configured interface or, if none is configured, the
// interface of addrFamily's default route. It is looked up on every resolve, as
// the default route may change or not exist yet when the client starts.
func (resolver *InterfaceResolver) interfaceFor(addrFamily string) (string, error) {
	if resolver.watchedInterface != "" {
		return resolver.watchedInterface, nil
	}

	return defaultRouteInterface(addrFamily)
}

func (resolver *InterfaceResolver) Resolve() (*update.DnsRecord, error) {
	record := &update.DnsRecord{
		Host:      resolver.host,
		Timestamp: time.Now(),
	}

	var used []string
	if slices.Contains(resolver.addrFamilies, internal.AddrFamilyIpv4) {
		ifName, err := resolver.interfaceFor(internal.AddrFamilyIpv4)
		if err != nil {
			// Without a default route there is no ipv4 to publish, the record can still carry an IPv6.
			log.Debug().Err(err).Str("component", "interface_resolver").Msg("no ipv4 default route")
		} else {
			ipv4, err := interfaceIpV4(ifName)
			if err != nil {
				return nil, err
			}
			record.IpV4 = ipv4
			used = append(used, ifName)
		}
	}

	if slices.Contains(resolver.addrFamilies, internal.AddrFamilyIpv6) {
		// A missing IPv6 address is not an error, the record can still carry an IPv4.
		ifName, err := resolver.interfaceFor(internal.AddrFamilyIpv6)
		if err != nil {
			log.Debug().Err(err).Str("component", "interface_resolver").Msg("no ipv6 default route")
		} else if ipv6, err := util.LocalGlobalIpV6Addr(ifName); err != nil {
			log.Debug().Err(err).Str("component", "interface_resolver").Str("interface", ifName).Msg("no usable ipv6 address")
		} else {
			record.IpV6 = ipv6.String()
			used = append(used, ifName)
		}
	}

	if !record.HasIpV4() && !record.HasIpV6() {
		if resolver.watchedInterface == "" {
			return nil, fmt.Errorf("no address found on the default route interfaces for address families %v", resolver.addrFamilies)
		}
		return nil, fmt.Errorf("no address found on interface %q for address families %v", resolver.watchedInterface, resolver.addrFamilies)
	}

	log.Debug().Str("component", "interface_resolver").Strs("interfaces", used).Msg("resolved addresses from interfaces")
	return record, nil
}

func interfaceIpV4(ifName string) (string, error) {
	ief, err := net.InterfaceByName(ifName)
	if err != nil {
		return "", fmt.Errorf("could not get interface %q: %w", ifName, err)
	}

	addrs, err := ief.Addrs()
	if err != nil {
		return "", fmt.Errorf("could not get addresses of interface %q: %w", ifName, err)
	}

	return pickIpV4(addrs), nil
}

// defaultRouteInterface returns the name of the interface that traffic of
// addrFamily to the internet leaves through. Connecting a udp socket makes the
// kernel look up the route and pick a source address without sending anything,
// the interface holding that source address is the one of the default route.
func defaultRouteInterface(addrFamily string) (string, error) {
	target, ok := serverAddresses[addrFamily]
	if !ok {
		return "", fmt.Errorf("unknown address family %q", addrFamily)
	}

	network := "udp4"
	if addrFamily == internal.AddrFamilyIpv6 {
		network = "udp6"
	}

	conn, err := net.Dial(network, target)
	if err != nil {
		return "", fmt.Errorf("could not find a %s route: %w", addrFamily, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("could not list interfaces: %w", err)
	}

	return interfaceWithIp(ifaces, conn.LocalAddr().(*net.UDPAddr).IP)
}

func interfaceWithIp(ifaces []net.Interface, ip net.IP) (string, error) {
	for _, ief := range ifaces {
		addrs, err := ief.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.Equal(ip) {
				return ief.Name, nil
			}
		}
	}

	return "", fmt.Errorf("no interface holds the source address %s", ip)
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
