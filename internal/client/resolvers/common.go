package resolvers

import (
	"github.com/soerenschneider/dyndns/internal/common"
)

type IpResolver interface {
	Resolve() (*common.DnsRecord, error)
	Name() string
	Host() string
}
