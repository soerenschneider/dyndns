package dns

import "dyndns/internal/common"

const defaultRecordTtl = 60

type Propagator interface {
	PropagateChange(ip common.ResolvedIp) error
}
