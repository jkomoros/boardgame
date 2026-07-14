package debuganimations

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Golden-equivalence harness for the five count-gated moves migrated in Task 7
(design spec §6): moveShuffleHidden, moveVisibleShuffleCards,
moveShuffleCards, moveStartMoveAllComponentsToHidden, and
moveStartMoveAllComponentsToVisible. Follows examples/memory and
examples/pig's legal_golden_test.go pattern (Task 11/12 precedent): a
hand-copied legacy oracle per move (or shared gate, where two moves had
byte-for-byte identical Legal() bodies), crossed against every proposer
worth distinguishing, checked against the migrated move's ACTUAL Legal().

None of these five move types embed moves.CurrentPlayer (all embed
moves.Default directly) and none declare WithLegalPhases, a progression, or
source/destination stack properties, so moves.Default.Legal's own
phase/progression/stack-constraint checks are no-ops for all of them (as
observed for memory's moveRevealCard) and are omitted from the oracles
below. Consequently the migrated result is identical for every proposer
tested; the proposer axis is still swept per the design spec §9 golden
harness shape, and every fixture/proposer combination agrees, so there is no
knownDivergence map for this file (no bucket-reordering can bite when no
predicate reads a move.* or player-current-player path at all).
*/

func newDebugAnimationsGame(t *testing.T) *boardgame.Game {
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

// debugAnimationsProposers returns every proposer worth distinguishing for
// this game (2 players, plus admin and observer wildcards).
func debugAnimationsProposers() map[string]boardgame.PlayerIndex {
	return map[string]boardgame.PlayerIndex{
		"player0":  0,
		"player1":  1,
		"admin":    boardgame.AdminPlayerIndex,
		"observer": boardgame.ObserverPlayerIndex,
	}
}

// drainStack moves every component out of from and into to.
func drainStack(t *testing.T, from, to boardgame.Stack) {
	t.Helper()
	for from.NumComponents() > 0 {
		if err := from.First().MoveToNextSlot(to); err != nil {
			t.Fatalf("legal_golden: draining stack: %v", err)
		}
	}
}

// moveNComponents moves exactly n components from from to to.
func moveNComponents(t *testing.T, from, to boardgame.Stack, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if from.NumComponents() == 0 {
			t.Fatalf("legal_golden: source stack ran out of components while moving %d", n)
		}
		if err := from.First().MoveToNextSlot(to); err != nil {
			t.Fatalf("legal_golden: moving component: %v", err)
		}
	}
}

// mustMutableState returns game's current state, failing the test if it
// isn't mutable (it always should be for a freshly built game).
func mustMutableState(t *testing.T, game *boardgame.Game) boardgame.State {
	t.Helper()
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatal("legal_golden: CurrentState() was not mutable")
	}
	return state
}

/**************************************************
 *
 * moveShuffleHidden golden coverage
 *
 **************************************************/

// legacyLegalMoveShuffleHidden is a hand-copied snapshot of
// moveShuffleHidden's Legal() method exactly as it read before this
// migration (see moves.go's comment block for the original source).
func legacyLegalMoveShuffleHidden(state boardgame.ImmutableState) error {
	game := state.ImmutableGameState().(*gameState)
	if game.FanDiscard.NumComponents() < 1 {
		return errors.New("FanDiscard has no cards to shuffle")
	}
	return nil
}

func TestGoldenLegalMoveShuffleHidden(t *testing.T) {
	type fixture struct {
		name string
		game *boardgame.Game
	}

	var fixtures []fixture

	// default: a fresh game. FanDiscard has 3 cards (FinishSetUp's
	// distribution) — legal.
	{
		game := newDebugAnimationsGame(t)
		fixtures = append(fixtures, fixture{"default", game})
	}

	// empty: FanDiscard drained to 0 — illegal.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		drainStack(t, gs.FanDiscard, gs.DrawStack)
		fixtures = append(fixtures, fixture{"empty", game})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := fx.game.MoveByName("Shuffle Hidden")
		if move == nil {
			t.Fatal("legal_golden: no \"Shuffle Hidden\" move found")
		}

		for proposerName, proposer := range debugAnimationsProposers() {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMoveShuffleHidden(state)
				actualErr := move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}

/**************************************************
 *
 * moveVisibleShuffleCards / moveShuffleCards golden coverage
 *
 * Both moves had byte-for-byte identical Legal() bodies (a single
 * game.FanStack.NumComponents() > 1 threshold), so they share one oracle
 * and one fixture builder, exercised against both move names.
 *
 **************************************************/

// legacyLegalFanStackShuffle is a hand-copied snapshot of the Legal() body
// shared by moveVisibleShuffleCards and moveShuffleCards before this
// migration.
func legacyLegalFanStackShuffle(state boardgame.ImmutableState) error {
	game, _ := concreteStates(state)
	if game.FanStack.NumComponents() > 1 {
		return nil
	}
	return errors.New("Aren't enough cards to shuffle")
}

func testGoldenFanStackShuffle(t *testing.T, moveName string) {
	type fixture struct {
		name string
		game *boardgame.Game
	}

	var fixtures []fixture

	// default: a fresh game. FanStack has 6 cards — legal (6 > 1).
	{
		game := newDebugAnimationsGame(t)
		fixtures = append(fixtures, fixture{"default", game})
	}

	// oneCard: FanStack drained down to exactly 1 card — illegal (1 > 1 is
	// false), the boundary case.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		moveNComponents(t, gs.FanStack, gs.DrawStack, gs.FanStack.NumComponents()-1)
		fixtures = append(fixtures, fixture{"oneCard", game})
	}

	// empty: FanStack drained to 0 — illegal.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		drainStack(t, gs.FanStack, gs.DrawStack)
		fixtures = append(fixtures, fixture{"empty", game})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := fx.game.MoveByName(moveName)
		if move == nil {
			t.Fatalf("legal_golden: no %q move found", moveName)
		}

		for proposerName, proposer := range debugAnimationsProposers() {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalFanStackShuffle(state)
				actualErr := move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}

func TestGoldenLegalMoveVisibleShuffleCards(t *testing.T) {
	testGoldenFanStackShuffle(t, "Visible Shuffle")
}

func TestGoldenLegalMoveShuffleCards(t *testing.T) {
	testGoldenFanStackShuffle(t, "Shuffle")
}

/**************************************************
 *
 * moveStartMoveAllComponentsToHidden golden coverage
 *
 **************************************************/

// legacyLegalMoveStartMoveAllComponentsToHidden is a hand-copied snapshot
// of moveStartMoveAllComponentsToHidden's Legal() method exactly as it read
// before this migration.
func legacyLegalMoveStartMoveAllComponentsToHidden(state boardgame.ImmutableState) error {
	game := state.ImmutableGameState().(*gameState)
	if game.AllVisibleStack.NumComponents() < 1 {
		return errors.New("No components in visible stack to move")
	}
	if game.AllHiddenStack.NumComponents() > 0 {
		return errors.New("The hidden stack already has items. Use the 'To Visible' move")
	}
	return nil
}

func TestGoldenLegalMoveStartMoveAllComponentsToHidden(t *testing.T) {
	type fixture struct {
		name string
		game *boardgame.Game
	}

	var fixtures []fixture

	// default: a fresh game. AllVisibleStack has 4 cards, AllHiddenStack is
	// never filled by distribution (0) — legal.
	{
		game := newDebugAnimationsGame(t)
		fixtures = append(fixtures, fixture{"default", game})
	}

	// visibleEmpty: AllVisibleStack drained to 0 — illegal, first check.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		drainStack(t, gs.AllVisibleStack, gs.DrawStack)
		fixtures = append(fixtures, fixture{"visibleEmpty", game})
	}

	// hiddenOccupied: AllVisibleStack left at its default 4 (nonempty, so
	// the first check passes), but AllHiddenStack is given a card too —
	// illegal, second check.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		moveNComponents(t, gs.DrawStack, gs.AllHiddenStack, 1)
		fixtures = append(fixtures, fixture{"hiddenOccupied", game})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := fx.game.MoveByName("Start Move All Components To Hidden")
		if move == nil {
			t.Fatal("legal_golden: no \"Start Move All Components To Hidden\" move found")
		}

		for proposerName, proposer := range debugAnimationsProposers() {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMoveStartMoveAllComponentsToHidden(state)
				actualErr := move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}

/**************************************************
 *
 * moveStartMoveAllComponentsToVisible golden coverage
 *
 **************************************************/

// legacyLegalMoveStartMoveAllComponentsToVisible is a hand-copied snapshot
// of moveStartMoveAllComponentsToVisible's Legal() method exactly as it
// read before this migration (mirror of the Hidden variant above).
func legacyLegalMoveStartMoveAllComponentsToVisible(state boardgame.ImmutableState) error {
	game := state.ImmutableGameState().(*gameState)
	if game.AllHiddenStack.NumComponents() < 1 {
		return errors.New("No components in hidden stack to move")
	}
	if game.AllVisibleStack.NumComponents() > 0 {
		return errors.New("The visible stack already has items. Use the 'To Hidden' move")
	}
	return nil
}

func TestGoldenLegalMoveStartMoveAllComponentsToVisible(t *testing.T) {
	type fixture struct {
		name string
		game *boardgame.Game
	}

	var fixtures []fixture

	// default: a fresh game. AllHiddenStack is never filled by distribution
	// (0) — illegal, first check. (This is the ONLY move of the five whose
	// out-of-the-box default state is illegal, since distribution favors
	// AllVisibleStack.)
	{
		game := newDebugAnimationsGame(t)
		fixtures = append(fixtures, fixture{"defaultHiddenEmpty", game})
	}

	// visibleOccupied: AllHiddenStack given a card (first check would now
	// pass), but AllVisibleStack is left at its default 4 (nonempty) —
	// illegal, second check.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		moveNComponents(t, gs.DrawStack, gs.AllHiddenStack, 1)
		fixtures = append(fixtures, fixture{"visibleOccupied", game})
	}

	// legal: AllVisibleStack drained to 0 AND AllHiddenStack given a card —
	// both checks pass.
	{
		game := newDebugAnimationsGame(t)
		state := mustMutableState(t, game)
		gs, _ := concreteStates(state)
		drainStack(t, gs.AllVisibleStack, gs.DrawStack)
		moveNComponents(t, gs.DrawStack, gs.AllHiddenStack, 1)
		fixtures = append(fixtures, fixture{"legal", game})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := fx.game.MoveByName("Start Move All Components To Visible")
		if move == nil {
			t.Fatal("legal_golden: no \"Start Move All Components To Visible\" move found")
		}

		for proposerName, proposer := range debugAnimationsProposers() {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMoveStartMoveAllComponentsToVisible(state)
				actualErr := move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}
