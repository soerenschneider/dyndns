package notification

import (
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type Notification interface {
	NotifyUpdatedIpDetected(ip *update.DnsRecord) error
	NotifyUpdatedIpApplied(ip *update.DnsRecord) error
}

type DummyNotification struct{}

func (d *DummyNotification) NotifyUpdatedIpDetected(ip *update.DnsRecord) error {
	return nil
}

func (d *DummyNotification) NotifyUpdatedIpApplied(ip *update.DnsRecord) error {
	return nil
}
