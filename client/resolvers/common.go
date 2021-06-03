package resolvers

import (
	"dyndns/internal/common"
)

type IpResolver interface {
	Resolve() (*common.ResolvedIp, error)
	Name() string
	Host() string
}
