package blackjack

import (
	"errors"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
)

func newReshuffleTestGame(t *testing.T) *boardgame.Game {
	t.Helper()
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}
	return game
}

func TestShuffleDiscardIntoDrawIntegration(t *testing.T) {
	game := newReshuffleTestGame(t)
	state := game.CurrentState().(boardgame.State)
	gameState, _ := concreteStates(state)
	if err := gameState.DrawStack.MoveAllTo(gameState.DiscardStack); err != nil {
		t.Fatalf("prepare discard: %v", err)
	}
	discardCount := gameState.DiscardStack.NumComponents()
	shuffleCount := gameState.DrawStack.ShuffleCount()

	move := game.MoveByName("Shuffle Discard Into Draw")
	if move == nil {
		t.Fatal("installed reshuffle move not found")
	}
	if err := move.Legal(state, boardgame.AdminPlayerIndex); err != nil {
		t.Fatalf("Legal: %v", err)
	}
	if err := move.Apply(state); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := gameState.DiscardStack.NumComponents(); got != 0 {
		t.Fatalf("discard count = %d, want 0", got)
	}
	if got := gameState.DrawStack.NumComponents(); got != discardCount {
		t.Fatalf("draw count = %d, want %d", got, discardCount)
	}
	if got := gameState.DrawStack.ShuffleCount(); got != shuffleCount+1 {
		t.Fatalf("shuffle count = %d, want %d", got, shuffleCount+1)
	}
}

func TestShuffleDiscardIntoDrawLegalPreflightsConstraint(t *testing.T) {
	game := newReshuffleTestGame(t)
	game.Manager().Internals().AllowMutableConstraints(game)
	state := game.CurrentState().(boardgame.State)
	gameState, _ := concreteStates(state)
	if err := gameState.DrawStack.MoveAllTo(gameState.DiscardStack); err != nil {
		t.Fatalf("prepare discard: %v", err)
	}
	beforeDiscard := gameState.DiscardStack.NumComponents()
	if err := gameState.DrawStack.AddConstraint(func(boardgame.ImmutableStack, []boardgame.ImmutableComponentInstance, boardgame.ImmutableState) error {
		return errors.New("rejected")
	}); err != nil {
		t.Fatalf("AddConstraint: %v", err)
	}

	move := game.MoveByName("Shuffle Discard Into Draw")
	err := move.Legal(state, boardgame.AdminPlayerIndex)
	if err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("Legal error = %v", err)
	}
	if got := gameState.DiscardStack.NumComponents(); got != beforeDiscard {
		t.Fatalf("Legal mutated discard: got %d, want %d", got, beforeDiscard)
	}
	if got := gameState.DrawStack.NumComponents(); got != 0 {
		t.Fatalf("Legal mutated draw: got %d, want 0", got)
	}
}
