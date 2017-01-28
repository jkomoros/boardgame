package boardgame

import (
	"testing"
)

func TestState(t *testing.T) {
	state := &State{
		0,
		0,
		nil,
		nil,
	}

	if state == nil {
		t.Error("State could not be created")
	}
}
