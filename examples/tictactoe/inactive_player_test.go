package tictactoe

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/storage/memory"
)

// TestSeatedPlayersGetActivated pins the deadlock that hit memory and pig
// (fixed 2026-07-25) and that this game had the identical shape for.
//
// Mechanism: playerState embeds behaviors.InactivePlayer and the game
// configures moves.SeatPlayer, so SeatPlayer.Apply marks every seated player
// INACTIVE. An inactive player makes base.GameDelegate.PlayerMayBeActive
// false (base/game_delegate.go), which makes PlayerIndex.Valid false
// (state.go), which makes moves.CurrentPlayer.Legal fail with "The specified
// target player is not valid" -- permanently, because only
// moves.ActivateInactivePlayer undoes the inactivity SeatPlayer stamps. This
// game is round-less, so per that move's own doc it should ALWAYS be legal.
//
// Ordinary tests miss this because manager.NewDefaultGame() leaves players
// unseated (SeatPlayer needs the server's rendezvous injection), so nothing
// is ever marked inactive and every move stays legal in-process.
//
// The invariant asserted here is the general one, so a game that legitimately
// has no CurrentPlayer-gated moves (debuganimations, whose moves are all
// moves.Default) passes without needing the activation move.
func TestSeatedPlayersGetActivated(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatal("couldn't create manager:", err)
	}

	// The mechanism's precondition: an inactive player really is ineligible
	// to be a CurrentPlayer target.
	inactive := &playerState{}
	inactive.SetPlayerInactive()
	if manager.Delegate().PlayerMayBeActive(inactive) {
		t.Fatal("an inactive player must not be eligible to be active; " +
			"the seated-player deadlock analysis assumes it is not")
	}

	seats, activates, currentPlayerGated := false, false, false
	for _, move := range manager.ExampleMoves() {
		switch move.(type) {
		case *moves.SeatPlayer:
			seats = true
		case *moves.ActivateInactivePlayer:
			activates = true
		}
		if embedsCurrentPlayer(reflect.TypeOf(move)) {
			currentPlayerGated = true
		}
	}

	if seats && currentPlayerGated && !activates {
		t.Error("this game seats players (marking them inactive) and gates " +
			"moves on CurrentPlayer, but never configures " +
			"moves.ActivateInactivePlayer: every seated player stays " +
			"inactive forever and those moves are permanently illegal")
	}
}

// embedsCurrentPlayer reports whether t embeds moves.CurrentPlayer at any
// depth. moves.CurrentPlayer's own marker method is package-private to
// moves, so structural inspection is the only way to ask this from here.
func embedsCurrentPlayer(t reflect.Type) bool {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return false
	}
	if t == reflect.TypeOf(moves.CurrentPlayer{}) {
		return true
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && embedsCurrentPlayer(field.Type) {
			return true
		}
	}
	return false
}
