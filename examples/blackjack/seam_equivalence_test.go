package blackjack

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal/legaltest"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/storage/memory"
)

// seamOptedInDelegate is this game with every move opted in to declarative
// legality with no authored preconditions, so each move's plan is exactly its
// framework bases' contributions (plus any LegalCustom).
type seamOptedInDelegate struct {
	*gameDelegate
}

func (d *seamOptedInDelegate) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.gameDelegate.ConfigureMoves(), moves.WithLegalPreconditions())
}

// TestDeclarativeEquivalenceEveryMoveOptedIn walks a game with every move
// opted in and checks, at every version, that each move's plan agrees with
// its frozen Legal() (see package legaltest).
func TestDeclarativeEquivalenceEveryMoveOptedIn(t *testing.T) {
	manager, err := boardgame.NewGameManager(&seamOptedInDelegate{NewDelegate().(*gameDelegate)}, memory.NewStorageManager())
	if err != nil {
		t.Fatalf("every move opted in must still boot: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}
	if compared := legaltest.Walk(t, game, 40); compared == 0 {
		t.Fatal("no move was compared")
	}
}
