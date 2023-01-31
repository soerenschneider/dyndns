package notification

import (
	"github.com/soerenschneider/dyndns/internal/common"
)

type Notification interface {
	NotifyUpdatedIpDetected(ip *common.ResolvedIp) error
	NotifyUpdatedIpApplied(ip *common.ResolvedIp) error
}
