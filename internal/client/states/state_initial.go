package states

import (
	"time"

	"github.com/soerenschneider/dyndns/internal/common"
)

type initialState struct{}

func NewInitialState() *initialState {
	return &initialState{}
}

func (state *initialState) String() string {
	return "initialState"
}

func (state *initialState) Name() string {
	return state.String()
}

func (state *initialState) EvaluateState(context Client, resolved *common.DnsRecord) bool {
	// This is just a dummy state, we'll immediately set the next state and invoke it
	context.SetState(NewIpNotConfirmedState())
	return context.GetState().EvaluateState(context, resolved)
}

func (state *initialState) WaitInterval() time.Duration {
	return 45 * time.Second
}
