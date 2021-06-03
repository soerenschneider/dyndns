package client

import "testing"

func Test_ipNotConfirmedState_PerformIpLookup(t *testing.T) {
	state := NewIpNotConfirmedState()
	for o := 0; o < 5; o++ {
		for i := 0; i < 9; i++ {
			if state.PerformIpLookup() {
				t.Fail()
			}
		}
		if !state.PerformIpLookup() {
			t.Fail()
		}
	}
}
