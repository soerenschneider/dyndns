package notification

import (
	"github.com/soerenschneider/dyndns/internal/common"
)

type Notification interface {
	NotifyUpdatedIpDetected(ip *common.ResolvedIp) error
	NotifyUpdatedIpApplied(ip *common.ResolvedIp) error
}

type DummyNotification struct{}

func (d *DummyNotification) NotifyUpdatedIpDetected(ip *common.ResolvedIp) error {
	return nil
}

func (d *DummyNotification) NotifyUpdatedIpApplied(ip *common.ResolvedIp) error {
	return nil
}
