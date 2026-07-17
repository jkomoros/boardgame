package debuganimations

import (
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
)

func TestManager(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(manager).IsNotNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(game).IsNotNil()

}

func TestHiddenShuffleMoveIsSilentForOtherViewers(t *testing.T) {
	storage := memory.NewStorageManager()
	manager, err := boardgame.NewGameManager(NewDelegate(), storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	move := game.MoveByName("Shuffle Hidden")
	version := move.Info().Version()
	if err := <-game.ProposeMove(move, 0); err != nil {
		t.Fatal(err)
	}
	record, err := storage.Move(game.ID(), version)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := game.MoveJSONForPlayer(0, record)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(actor)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"AnimationKey":"Shuffle Hidden","Version":1}`; got != want {
		t.Fatalf("actor move = %s, want %s", got, want)
	}
	other, err := game.MoveJSONForPlayer(1, record)
	if err != nil {
		t.Fatal(err)
	}
	if other != nil {
		t.Fatalf("other viewer received silent move metadata: %#v", other)
	}
}
