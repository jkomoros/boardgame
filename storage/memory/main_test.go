package memory

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/internal/test"
)

func TestStorageManager(t *testing.T) {

	test.Test(func() test.StorageManager {
		return NewStorageManager()
	}, "memory", "", t)

}

func TestSequentialGameVersion(t *testing.T) {
	test.SequentialGameVersionTest(func() test.StorageManager {
		return NewStorageManager()
	}, "", t)
}

func TestCoreRecordsAreCopiedAtStorageBoundary(t *testing.T) {
	storage := NewStorageManager()
	game := &boardgame.GameStorageRecord{
		ID:      "copy-test",
		Version: 0,
		Winners: []boardgame.PlayerIndex{1},
		Agents:  []string{"bot"},
		Variant: boardgame.Variant{"mode": "original"},
	}
	state := boardgame.StateStorageRecord(`{"value":"original"}`)
	move := &boardgame.MoveStorageRecord{Version: 0, Blob: []byte(`{"choice":"original"}`)}
	if err := storage.SaveGameAndCurrentState(game, state, move); err != nil {
		t.Fatal(err)
	}

	game.Winners[0] = 2
	game.Agents[0] = "changed"
	game.Variant["mode"] = "changed"
	state[10] = 'X'
	move.Blob[11] = 'X'

	storedGame, err := storage.Game(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	storedState, err := storage.State(game.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	storedMove, err := storage.Move(game.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(storedGame.Winners, []boardgame.PlayerIndex{1}) ||
		!reflect.DeepEqual(storedGame.Agents, []string{"bot"}) || storedGame.Variant["mode"] != "original" {
		t.Fatalf("mutating save input changed stored game: %#v", storedGame)
	}
	if string(storedState) != `{"value":"original"}` || string(storedMove.Blob) != `{"choice":"original"}` {
		t.Fatalf("mutating save input changed stored state or move: %s, %s", storedState, storedMove.Blob)
	}

	storedGame.Variant["mode"] = "returned"
	storedState[10] = 'Y'
	storedMove.Blob[11] = 'Y'
	reloadedGame, _ := storage.Game(game.ID)
	reloadedState, _ := storage.State(game.ID, 0)
	reloadedMove, _ := storage.Move(game.ID, 0)
	if reloadedGame.Variant["mode"] != "original" || string(reloadedState) != `{"value":"original"}` || string(reloadedMove.Blob) != `{"choice":"original"}` {
		t.Fatal("mutating returned records changed stored values")
	}
}
