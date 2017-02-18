/*

	test is a package that is used to run a boardgame/server.StorageManager
	implementation through its paces and verify it does everything correctly.

*/
package test

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"reflect"
	"testing"
)

type StorageManager interface {
	boardgame.StorageManager

	//CleanUp will be called when a given manager is done and can be dispoed of.
	CleanUp()

	//The methods past this point are the same ones that are included in Server.StorageManager
	Close()
	ListGames(manager *boardgame.GameManager, max int) []*boardgame.Game
}

type StorageManagerFactory func() StorageManager

func Test(factory StorageManagerFactory, t *testing.T) {

	storage := factory()

	manager := tictactoe.NewManager(storage)

	game := boardgame.NewGame(manager)

	game.SetUp(0)

	move := game.PlayerMoveByName("Place Token")

	if move == nil {
		t.Fatal("Couldn't find a move")
	}

	if err := <-game.ProposeMove(move); err != nil {
		t.Fatal("Couldn't make move", err)
	}

	//OK, now test that the manager and SetUp and everyone did the right thing.

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

	//TODO: figure out how to test that name is matched when retrieving from store.

	storage.CleanUp()

}

func compareJSONObjects(in []byte, golden []byte, message string, t *testing.T) {

	//recreated in boardgame/state_test.go

	var deserializedIn interface{}
	var deserializedGolden interface{}

	json.Unmarshal(in, &deserializedIn)
	json.Unmarshal(golden, &deserializedGolden)

	if deserializedIn == nil {
		t.Error("In didn't deserialize", message)
	}

	if deserializedGolden == nil {
		t.Error("Golden didn't deserialize", message)
	}

	if !reflect.DeepEqual(deserializedIn, deserializedGolden) {
		t.Error("Got wrong json.", message, "Got", string(in), "wanted", string(golden))
	}
}
