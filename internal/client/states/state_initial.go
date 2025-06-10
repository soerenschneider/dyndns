package states

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
)

type initialState struct {
	forceSendUpdate bool
}

func NewInitialState(forceSendUpdate bool) *initialState {
	return &initialState{
		forceSendUpdate: forceSendUpdate,
	}
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
	if state.forceSendUpdate {
		log.Info().Str("component", "state_machine").Str("state", state.Name()).Msg("forceSendUpdate is set, sending update")
		return true
	}
	return context.GetState().EvaluateState(context, resolved)
}

func (state *initialState) WaitInterval() time.Duration {
	return 45 * time.Second
}
