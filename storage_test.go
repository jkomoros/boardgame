package boardgame

import (
	"encoding/json"
	"testing"
)

func TestInMemoryStorageManger(t *testing.T) {

	game := testGame()

	game.SetUp(0)

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	if err := <-game.ProposeMove(move); err != nil {
		t.Fatal("Couldn't make move", err)
	}

	//OK, now test that the manager and SetUp and everyone did the right thing.

	//Manager.Storage() is a InMemoryStorageManager

	//Game.SetUp() should have stored the state to the storages

	storage := game.Manager().Storage()
	manager := game.Manager()

	localGame := storage.Game(manager, game.Id())

	if localGame == nil {
		t.Fatal("Couldn't get game copy out")
	}

	if localGame.Modifiable() {
		t.Error("We asked for a non-modifiable game, got a modifiable one")
	}

	blob, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		t.Fatal("couldn't marshal game", err)
	}

	localBlob, err := json.MarshalIndent(localGame, "", "  ")

	if err != nil {
		t.Fatal("Couldn't marshal localGame", err)
	}

	compareJSONObjects(blob, localBlob, "Comparing game and local game", t)

	state := game.State(0)
	stateBlob, _ := json.MarshalIndent(state, "", "  ")

	localState := localGame.State(0)
	localStateBlob, _ := json.MarshalIndent(localState, "", "  ")

	compareJSONObjects(stateBlob, localStateBlob, "Comparing game version 0", t)

	//Verify that if the game is stored with wrong name that doesn't match manager it won't load up.

	store := storage.(*InMemoryStorageManager)

	record := store.games[game.Id()]

	record.Name = "BOGUS"

	bogusGame := storage.Game(manager, game.Id())

	if bogusGame != nil {
		t.Error("Game shouldn't have come back because name doesn't match")
	}

}
