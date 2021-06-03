package resolvers

import "dyndns/internal/common"

type CliResolver struct {
	ipv4 string
	ipv6 string
	host string
}

func NewCliResolver(ipv4, ipv6, host string) (*CliResolver, error) {
	return &CliResolver{
		ipv4: ipv4,
		ipv6: ipv6,
		host: host,
	}, nil
}

func (resolver *CliResolver) Name() string {
	return "CLI"
}

func (resolver *CliResolver) Host() string {
	return resolver.host
}

func (resolver *CliResolver) Resolve() (*common.ResolvedIp, error) {
	return &common.ResolvedIp{
		IpV4: resolver.ipv4,
		IpV6: resolver.ipv6,
		Host: resolver.host,
	}, nil
}
