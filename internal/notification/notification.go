package notification

import (
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"go.uber.org/multierr"
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

// Multi sends every notification to all of its notifiers.
type Multi []Notification

func (m Multi) NotifyUpdatedIpDetected(ip *update.DnsRecord) error {
	var errs error
	for _, n := range m {
		errs = multierr.Append(errs, n.NotifyUpdatedIpDetected(ip))
	}
	return errs
}

func (m Multi) NotifyUpdatedIpApplied(ip *update.DnsRecord) error {
	var errs error
	for _, n := range m {
		errs = multierr.Append(errs, n.NotifyUpdatedIpApplied(ip))
	}
	return errs
}
