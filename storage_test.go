package boardgame

import (
	"reflect"
	"testing"
)

func TestInMemoryStorageManger(t *testing.T) {

	manager := NewInMemoryStorageManager()

	game := testGame()

	game.SetUp(0)

	//Save state 0

	state0 := game.storage.State(game, game.Version())

	if err := manager.SaveState(game, 0, 0, state0); err != nil {
		t.Error("Save state 0 failed", err)
	}

	savedState0 := manager.State(game, 0)

	if !reflect.DeepEqual(savedState0, state0) {
		t.Error("Reconstituted state did not deep equal state. Got", savedState0, "wanted", state0)
	}

	//Try to save again insame slot

	if err := manager.SaveState(game, 0, 0, state0); err == nil {
		t.Error("We didn't get an error trying to save again at same version")
	}

}
