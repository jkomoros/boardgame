package boardgame

import (
	"reflect"
	"testing"
)

func TestInMemoryStorageManger(t *testing.T) {

	manager := NewInMemoryStorageManager()

	game := testGame()

	game.SetUp()

	//Save state 0

	state0 := game.StateWrapper

	if err := manager.SaveState(game, game.StateWrapper); err != nil {
		t.Error("Save state 0 failed", err)
	}

	wrapper := manager.State(game, 0)

	if wrapper == nil {
		t.Error("State didn't return back for 0")
	}

	if !reflect.DeepEqual(wrapper, state0) {
		t.Error("Reconstituted state did not deep equal state. Got", wrapper, "wanted", state0)
	}

	//Try to save again insame slot

	if err := manager.SaveState(game, game.StateWrapper); err == nil {
		t.Error("We didn't get an error trying to save again at same version")
	}

}
