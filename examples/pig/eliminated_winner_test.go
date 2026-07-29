package pig

import (
	"testing"

	"github.com/jkomoros/boardgame"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

// TestEliminatedPlayerCannotWin pins the default scoring contract in
// base.GameDelegate.CheckGameFinished: a player who is out of play when the
// game-end condition fires is not a candidate for the win.
//
// pig embeds both behaviors.PlayerElimination and behaviors.ScoreBehavior, and
// behaviors.PlayerElimination deliberately does NOT also set the player
// inactive (behaviors/elimination.go). So the inactive-only filter that
// CheckGameFinished used to apply left eliminated players in the winner pool,
// and the highest-scoring eliminated player won the game outright.
func TestEliminatedPlayerCannotWin(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building default game: %v", err)
	}
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatal("CurrentState() was not mutable")
	}

	_, players := concreteStates(state)
	if len(players) < 2 {
		t.Fatalf("expected at least 2 players, got %d", len(players))
	}

	// Player 0 has the top score but is out of play. Player 1 is the only
	// player still standing, so player 1 is the winner.
	players[0].Score = defaultTargetScore + 10
	players[0].Eliminated = true
	players[1].Score = defaultTargetScore
	players[1].Eliminated = false
	for _, p := range players[2:] {
		p.Score = 0
		p.Eliminated = false
	}

	finished, winners := manager.Delegate().CheckGameFinished(state)
	if !finished {
		t.Fatal("expected the game-end condition to be met")
	}
	if len(winners) != 1 || winners[0] != boardgame.PlayerIndex(1) {
		t.Errorf("expected player 1 to be the sole winner, got %v -- an "+
			"eliminated player is still in the winner pool", winners)
	}
}

// TestEliminationDoesNotDisturbNormalScoring is the other half of the
// contract: with nobody eliminated, the highest scorer still wins, and ties
// still produce multiple winners.
func TestEliminationDoesNotDisturbNormalScoring(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building default game: %v", err)
	}
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatal("CurrentState() was not mutable")
	}

	_, players := concreteStates(state)
	for _, p := range players {
		p.Score = defaultTargetScore
		p.Eliminated = false
	}

	finished, winners := manager.Delegate().CheckGameFinished(state)
	if !finished {
		t.Fatal("expected the game-end condition to be met")
	}
	if len(winners) != len(players) {
		t.Errorf("expected every player to tie for the win, got %v", winners)
	}
}
