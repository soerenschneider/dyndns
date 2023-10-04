package notification

import (
	"github.com/soerenschneider/dyndns/internal/common"
)

type Notification interface {
	NotifyUpdatedIpDetected(ip *common.DnsRecord) error
	NotifyUpdatedIpApplied(ip *common.DnsRecord) error
}

type DummyNotification struct{}

func (d *DummyNotification) NotifyUpdatedIpDetected(ip *common.DnsRecord) error {
	return nil
}

func (d *DummyNotification) NotifyUpdatedIpApplied(ip *common.DnsRecord) error {
	return nil
}
