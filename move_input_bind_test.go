package boardgame_test

import (
	"testing"

	"github.com/jkomoros/boardgame"
	gameMemory "github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/storage/memory"
)

func TestMoveWithInputUsesCreatorContract(t *testing.T) {
	manager, err := boardgame.NewGameManager(gameMemory.NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		input map[string]interface{}
		valid bool
	}{
		{"explicit zero", map[string]interface{}{"CardIndex": 0}, true},
		{"ordinary index", map[string]interface{}{"CardIndex": 2}, true},
		{"missing", nil, false},
		{"null", map[string]interface{}{"CardIndex": nil}, false},
		{"fraction", map[string]interface{}{"CardIndex": 1.5}, false},
		{"string number", map[string]interface{}{"CardIndex": "1"}, false},
		{"extra", map[string]interface{}{"CardIndex": 0, "Typo": 1}, false},
		{"context", map[string]interface{}{"CardIndex": 0, "TargetPlayerIndex": 1}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			move, err := game.MoveWithInput("Reveal Card", test.input)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v: move=%v error=%v", test.valid, move, err)
			}
			if err != nil && move != nil {
				t.Fatal("returned partially bound move")
			}
		})
	}
	move, err := game.MoveWithInput("Reveal Card", map[string]interface{}{"CardIndex": 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := <-game.ProposeMoveAtVersion(move, game.CurrentState().CurrentPlayerIndex(), game.Version()); err != nil {
		t.Fatalf("bound move did not execute through canonical legality: %v", err)
	}
	if move, err := game.MoveWithInput("Capture Cards", nil); err == nil || move != nil {
		t.Fatalf("creator binder exposed engine-owned fix-up: move=%T err=%v", move, err)
	}
}
