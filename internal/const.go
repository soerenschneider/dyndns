package internal

import "github.com/soerenschneider/dyndns/v2/pkg/update"

const (
	AddrFamilyIpv6 = update.AddrFamilyIpv6
	AddrFamilyIpv4 = update.AddrFamilyIpv4

	ResolverHttp      = "http"
	ResolverInterface = "interface"

	ServerMode = "server"
	ClientMode = "client"
	EdgeMode   = "edge"
)
