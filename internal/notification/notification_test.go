package notification

import (
	"errors"
	"testing"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type recordingNotification struct {
	detected, applied int
	err               error
}

func (r *recordingNotification) NotifyUpdatedIpDetected(*update.DnsRecord) error {
	r.detected++
	return r.err
}

func (r *recordingNotification) NotifyUpdatedIpApplied(*update.DnsRecord) error {
	r.applied++
	return r.err
}

func TestMulti(t *testing.T) {
	failing := &recordingNotification{err: errors.New("smtp down")}
	working := &recordingNotification{}
	multi := Multi{failing, working}

	if err := multi.NotifyUpdatedIpDetected(&update.DnsRecord{}); err == nil {
		t.Error("NotifyUpdatedIpDetected() error = nil, want the failing notifier's error")
	}
	if err := multi.NotifyUpdatedIpApplied(&update.DnsRecord{}); err == nil {
		t.Error("NotifyUpdatedIpApplied() error = nil, want the failing notifier's error")
	}

	// A failing notifier must not keep the others from being notified.
	for _, n := range []*recordingNotification{failing, working} {
		if n.detected != 1 || n.applied != 1 {
			t.Errorf("notifier called detected=%d applied=%d, want 1 and 1", n.detected, n.applied)
		}
	}
}
