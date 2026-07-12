package tictactoe

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Golden-equivalence harness for movePlaceToken's PARTIAL declarative
migration (Task 7 survey re-check, design spec §6 §3): the
token-availability gate moved to WithPreconditions
(legal.StackNotEmpty("players[move.TargetPlayerIndex].UnusedTokens")); the
MayMoveToSlot check survives as LegalCustom residue (see moves.go's doc
comment). Follows examples/memory and examples/pig's legal_golden_test.go
pattern (Task 11/12 precedent): a hand-copied legacy oracle (replicating
moves.CurrentPlayer.Legal's proposer checks directly, per memory's
precedent, rather than dispatching into the now-migrated chain), crossed
against every proposer worth distinguishing, checked against the migrated
move's ACTUAL Legal().
*/

// legacyLegalMovePlaceToken is a hand-copied snapshot of movePlaceToken's
// Legal() method exactly as it read before this migration (see moves.go's
// comment block for the original source), including moves.CurrentPlayer's
// proposer checks it super-called (current_player.go).
func legacyLegalMovePlaceToken(m *movePlaceToken, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	currentPlayer := state.CurrentPlayerIndex()
	targetPlayerIndex := m.TargetPlayerIndex.EnsureValid(state)

	if !targetPlayerIndex.Valid(state) {
		return errors.New("The specified target player is not valid")
	}
	if targetPlayerIndex < 0 {
		return errors.New("The specified target player is not valid")
	}
	if !targetPlayerIndex.Equivalent(currentPlayer) {
		return errors.New("it's not your turn")
	}
	if !targetPlayerIndex.Equivalent(proposer) {
		return errors.New("it's not your turn")
	}

	game, players := concreteStates(state)

	first := players[m.TargetPlayerIndex.EnsureValid(state)].UnusedTokens.ImmutableFirst()
	if first == nil {
		return errors.New("there aren't any remaining tokens for the current player to place")
	}

	return first.MayMoveToSlot(game.Slots, m.Slot)
}

func newTicTacToeGame(t *testing.T) *boardgame.Game {
	t.Helper()
	manager, err := boardgame.NewGameManager(NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("legal_golden: building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal_golden: building default game: %v", err)
	}
	return game
}

func mustMutableTTState(t *testing.T, game *boardgame.Game) boardgame.State {
	t.Helper()
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatal("legal_golden: CurrentState() was not mutable")
	}
	return state
}

func ticTacToeProposers(t *testing.T, state boardgame.State) map[string]boardgame.PlayerIndex {
	t.Helper()
	currentPlayer := state.CurrentPlayerIndex()
	var otherPlayer boardgame.PlayerIndex = -1
	for i := range state.ImmutablePlayerStates() {
		pIdx := boardgame.PlayerIndex(i)
		if pIdx != currentPlayer {
			otherPlayer = pIdx
			break
		}
	}
	if otherPlayer < 0 {
		t.Fatal("legal_golden: could not find a non-current player")
	}
	return map[string]boardgame.PlayerIndex{
		"currentPlayer": currentPlayer,
		"otherPlayer":   otherPlayer,
		"admin":         boardgame.AdminPlayerIndex,
		"observer":      boardgame.ObserverPlayerIndex,
	}
}

// placeTokenMove returns a fresh "Place Token" move from game, with Slot
// set to slot.
func placeTokenMove(t *testing.T, game *boardgame.Game, slot int) *movePlaceToken {
	t.Helper()
	move := game.MoveByName("Place Token")
	if move == nil {
		t.Fatal("legal_golden: no \"Place Token\" move found")
	}
	pt, ok := move.(*movePlaceToken)
	if !ok {
		t.Fatal("legal_golden: \"Place Token\" move was not a *movePlaceToken")
	}
	pt.Slot = slot
	return pt
}

// knownMessageOrderingDivergence names (fixture, proposer) combinations
// where the migrated plan is EXPECTED to disagree with the legacy oracle on
// WHICH message wins, even though both agree the move is illegal (nil-ness
// always matches). See memory/legal_golden_test.go's doc comment for the
// full architectural explanation (design spec §5's field-independent/
// field-dependent bucketing does not preserve declaration order in
// general). Populated empirically below if the sweep finds a case; left
// empty if not (both the token-availability read and the contributed
// proposerIsCurrentPlayer atom read a move.* path here, so they are BOTH
// field-dependent -- unlike memory's revealCard, there is no
// independent/dependent split to reorder them relative to each other).
var knownMessageOrderingDivergence = map[string]bool{}

func TestGoldenLegalMovePlaceToken(t *testing.T) {
	type fixture struct {
		name string
		game *boardgame.Game
		slot int
	}

	var fixtures []fixture

	// default: a fresh game, targeting an empty slot. Legal.
	{
		game := newTicTacToeGame(t)
		fixtures = append(fixtures, fixture{"default", game, 0})
	}

	// noTokensLeft: the current player's UnusedTokens drained to 0 (moved
	// into other board slots, not the one under test) -- illegal via the
	// migrated declarative gate.
	{
		game := newTicTacToeGame(t)
		state := mustMutableTTState(t, game)
		gs, players := concreteStates(state)
		current := players[state.CurrentPlayerIndex().EnsureValid(state)]
		slotIdx := 1
		for current.UnusedTokens.NumComponents() > 0 {
			if err := current.UnusedTokens.First().MoveTo(gs.Slots, slotIdx); err != nil {
				t.Fatalf("legal_golden: draining UnusedTokens: %v", err)
			}
			slotIdx++
		}
		fixtures = append(fixtures, fixture{"noTokensLeft", game, 0})
	}

	// occupiedSlot: slot 0 is already occupied (by the OTHER player's
	// token, moved there directly, leaving the current player's
	// UnusedTokens untouched) -- illegal via the LegalCustom residue
	// (MayMoveToSlot).
	{
		game := newTicTacToeGame(t)
		state := mustMutableTTState(t, game)
		gs, players := concreteStates(state)
		current := state.CurrentPlayerIndex().EnsureValid(state)
		var otherIdx boardgame.PlayerIndex = -1
		for i := range players {
			pIdx := boardgame.PlayerIndex(i)
			if pIdx != current {
				otherIdx = pIdx
				break
			}
		}
		if otherIdx < 0 {
			t.Fatal("legal_golden: could not find a non-current player")
		}
		other := players[otherIdx]
		if err := other.UnusedTokens.First().MoveTo(gs.Slots, 0); err != nil {
			t.Fatalf("legal_golden: occupying slot 0: %v", err)
		}
		fixtures = append(fixtures, fixture{"occupiedSlot", game, 0})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := placeTokenMove(t, fx.game, fx.slot)

		for proposerName, proposer := range ticTacToeProposers(t, state.(boardgame.State)) {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMovePlaceToken(move, state, proposer)
				actualErr := move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if knownMessageOrderingDivergence[fx.name+"/"+proposerName] {
					return
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}
