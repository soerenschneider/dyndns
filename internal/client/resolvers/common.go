package resolvers

import (
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type IpResolver interface {
	Resolve() (*update.DnsRecord, error)
	Name() string
	Host() string
}
