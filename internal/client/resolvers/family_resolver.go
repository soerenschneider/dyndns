package resolvers

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"go.uber.org/multierr"
)

// FamilyResolver resolves IPv4 and IPv6 with distinct resolvers, e.g. the
// public IPv4 via http behind NAT and the IPv6 straight from the interface.
type FamilyResolver struct {
	host string
	ipv4 IpResolver
	ipv6 IpResolver
}

func NewFamilyResolver(host string, ipv4, ipv6 IpResolver) (*FamilyResolver, error) {
	if ipv4 == nil || ipv6 == nil {
		return nil, errors.New("both an ipv4 and an ipv6 resolver must be provided")
	}

	return &FamilyResolver{
		host: host,
		ipv4: ipv4,
		ipv6: ipv6,
	}, nil
}

func (resolver *FamilyResolver) Name() string {
	return fmt.Sprintf("FamilyResolver (ipv4: %s, ipv6: %s)", resolver.ipv4.Name(), resolver.ipv6.Name())
}

func (resolver *FamilyResolver) Host() string {
	return resolver.host
}

// Resolve takes the IPv4 from the ipv4 resolver and the IPv6 from the ipv6
// resolver. A failing family is logged and left empty; Resolve only fails if
// neither family yields an address.
func (resolver *FamilyResolver) Resolve() (*update.DnsRecord, error) {
	record := &update.DnsRecord{
		Host:      resolver.host,
		Timestamp: time.Now(),
	}

	var errs error
	if resolved, err := resolver.ipv4.Resolve(); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("ipv4 via %s: %w", resolver.ipv4.Name(), err))
	} else if resolved != nil {
		record.IpV4 = resolved.IpV4
	}

	if resolved, err := resolver.ipv6.Resolve(); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("ipv6 via %s: %w", resolver.ipv6.Name(), err))
	} else if resolved != nil {
		record.IpV6 = resolved.IpV6
	}

	if !record.HasIpV4() && !record.HasIpV6() {
		if errs == nil {
			errs = errors.New("neither ipv4 nor ipv6 address resolved")
		}
		return nil, errs
	}

	if errs != nil {
		log.Warn().Err(errs).Str("component", "family_resolver").Str("host", resolver.host).Msg("could not resolve all address families")
	}

	return record, nil
}
