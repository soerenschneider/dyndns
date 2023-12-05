package states

import (
	"time"

	"github.com/soerenschneider/dyndns/internal/common"
)

type State interface {
	// EvaluateState evaluates the current state and returns true if the client should proceed sending a change request
	// using the currently detected ip
	EvaluateState(client Client, ip *common.DnsRecord) bool
	// WaitInterval returns the amount of time to sleep after a tick.
	WaitInterval() time.Duration
	Name() string
}

type Client interface {
	SetState(newState State)
	GetState() State
	GetLastStateChange() time.Time
}
