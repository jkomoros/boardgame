package checkers

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Golden-equivalence harness for moveMoveToken and movePlaceToken (design spec
§8/§9, Task 12 brief): for every recorded (state, proposer) pair, asserts that
legacyLegalMoveMoveToken (a hand-copied snapshot of the move's Legal() body
exactly as it read before this migration) and the migrated move's ACTUAL
Legal() (now dispatched through moves.CurrentPlayer.Legal ->
moves.Default.Legal's plan-evaluation short-circuit, since moves.go's Legal()
override was deleted and replaced with LegalCustom) agree on both nil-ness
and, for a non-nil result outside the documented residue-collapse divergence
below, the exact error message string.

Fixture-construction approach follows examples/memory/legal_golden_test.go's
precedent (Task 11): build a real game via manager.NewDefaultGame() and
mutate specific state fields/stacks directly to reach each branch
deterministically. A fresh checkers game is special-cased further: per
legal/conformance_test.go's "checkersDefault" fixture (Task 4/5 precedent),
NewDefaultGame() already runs SeatPlayer and every PlaceToken fixup to
completion, landing in phasePlaying with the standard 24-token starting
layout — no manual setup-phase play is needed to reach a moveMoveToken-legal
state.

movePlaceToken is the mirror case: it is only legal DURING phaseSetup, which
NewDefaultGame() has already run to completion, so its fixtures can't come
from the finished head state. They come instead from genuine HISTORICAL
versions of the same finished game (game.State(v)) — real mid-setup snapshots.
This matters for equivalence: movePlaceToken is added via AddOrderedForPhase,
so FixUpMulti contributes an inProgression base check (walking the
move-progression tape) alongside the inPhase check. At a real historical setup
version that tape is genuinely mid-progression, so proposing another Place
Token there PASSES both base checks — for the legacy oracle and the migrated
plan alike (they call the identical (*Default).legalMoveInProgression /
"inProgression" implementation). Forcing a FINISHED game back into phaseSetup
by mutation, by contrast, leaves the real tape ending in StartPhase, so the
progression check fails and never reaches the gates under test.
*/

// legacyLegalMoveMoveToken is a hand-copied snapshot of moveMoveToken's
// Legal() method exactly as it read before this migration (see moves.go's
// comment block for the original source). It deliberately does NOT call
// m.CurrentPlayer.Legal(state, proposer): that would dispatch through
// moves.Default.Legal, which (post-migration) detects the assembled plan and
// evaluates THAT instead of the frozen chain — defeating the point of an
// independent oracle. Instead it hand-replicates:
//   - moves.Default.legalInPhase's check via the same exported helper that
//     check itself calls (boardgame.LegalInPhaseCheck — moves/default.go's
//     legalInPhase doc comment: "extracted to core so that this frozen
//     chain and legal's inPhase wrapper predicate call exactly one
//     implementation", a Task 7 decision this oracle relies on rather than
//     re-deriving); moveMoveToken's only legal phase is phasePlaying (set
//     via moves.AddForPhase in main.go). legalMoveInProgression and
//     legalStackConstraints are both no-ops for moveMoveToken (no
//     WithLegalMoveProgression/WithSourceProperty+WithDestinationProperty
//     configured), so they are omitted.
//   - moves.CurrentPlayer.Legal's own TargetPlayerIndex checks
//     (moves/current_player.go:37-65), by hand.
//   - moveMoveToken's own body, verbatim.
func legacyLegalMoveMoveToken(m *moveMoveToken, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := boardgame.LegalInPhaseCheck(state, []enum.EnumKey{enum.EnumKey(phasePlaying)}); err != nil {
		return err
	}

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

	p := state.ImmutableCurrentPlayer().(*playerState)
	g := state.ImmutableGameState().(*gameState)

	if err := g.Spaces.MaySwapComponentsByKey(m.TokenIndexToMove.Value(), m.SpaceIndex.Value()); err != nil {
		return err
	}

	c := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value())

	if c == nil {
		return errors.New("That space does not have a component in it")
	}

	t := c.Values().(*token)

	if !p.Color.Equals(t.Color) {
		return errors.New("that token isn't your token to move")
	}

	if !spaceIsBlack(m.SpaceIndex.Value().Int()) {
		return errors.New("you can only move to spaces that are black")
	}

	for _, space := range t.FreeNextSpaces(state, m.TokenIndexToMove.Value().Int()) {
		if m.SpaceIndex.Value().Int() == space {
			return nil
		}
	}

	for _, space := range t.LegalCaptureSpaces(state, m.TokenIndexToMove.Value().Int()) {
		if m.SpaceIndex.Value().Int() == space {
			return nil
		}
	}

	return errors.New("spaceIndex does not represent a legal space for that token to move to")
}

// moveMoveTokenGoldenFixture is one (game, move) pair to check every
// proposer worth distinguishing against.
type moveMoveTokenGoldenFixture struct {
	name string
	game *boardgame.Game
	move *moveMoveToken
}

func newMoveMoveTokenGame(t *testing.T) (*boardgame.Game, boardgame.State) {
	t.Helper()
	manager, err := boardgame.NewGameManager(NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("legal_golden: building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal_golden: building default game: %v", err)
	}
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatalf("legal_golden: CurrentState() was not mutable")
	}
	return game, state
}

// moveMoveTokenMove returns a fresh "Move Token" move from game, with
// TokenIndexToMove and SpaceIndex set.
func moveMoveTokenMove(t *testing.T, game *boardgame.Game, tokenIndex, spaceIndex int) *moveMoveToken {
	t.Helper()
	move := game.MoveByName("Move Token")
	if move == nil {
		t.Fatal("legal_golden: no \"Move Token\" move found")
	}
	mv, ok := move.(*moveMoveToken)
	if !ok {
		t.Fatal("legal_golden: \"Move Token\" move was not a *moveMoveToken")
	}
	val, err := mv.ReadSetter().EnumProp("TokenIndexToMove")
	if err != nil {
		t.Fatalf("legal_golden: reading TokenIndexToMove enum prop: %v", err)
	}
	if err := val.SetValue(enum.EnumKey(tokenIndex)); err != nil {
		t.Fatalf("legal_golden: setting TokenIndexToMove: %v", err)
	}
	val, err = mv.ReadSetter().EnumProp("SpaceIndex")
	if err != nil {
		t.Fatalf("legal_golden: reading SpaceIndex enum prop: %v", err)
	}
	if err := val.SetValue(enum.EnumKey(spaceIndex)); err != nil {
		t.Fatalf("legal_golden: setting SpaceIndex: %v", err)
	}
	return mv
}

func firstOccupiedSpace(spaces boardgame.ImmutableStack) int {
	for i := 0; i < spaces.Len(); i++ {
		if spaces.ImmutableComponentAt(i) != nil {
			return i
		}
	}
	return -1
}

func firstEmptySpace(spaces boardgame.ImmutableStack) int {
	for i := 0; i < spaces.Len(); i++ {
		if spaces.ImmutableComponentAt(i) == nil {
			return i
		}
	}
	return -1
}

// firstSpaceWithColorMatch returns the lowest Spaces index whose occupying
// token's Color equals (or, if wantMatch is false, does not equal)
// currentPlayerColor, or -1 if none exists. Mirrors
// legal/conformance_test.go's firstSpaceWithColor helper (same purpose,
// duplicated here so this package's tests don't need to import
// package legal_test's unexported fixtures).
func firstSpaceWithColorMatch(t *testing.T, spaces boardgame.ImmutableStack, currentPlayerColor enum.ImmutableVal, wantMatch bool) int {
	t.Helper()
	for i := 0; i < spaces.Len(); i++ {
		c := spaces.ImmutableComponentAt(i)
		if c == nil {
			continue
		}
		tokenColor, err := c.Values().Reader().ImmutableEnumProp("Color")
		if err != nil {
			t.Fatalf("legal_golden: reading token Color at Spaces[%d]: %v", i, err)
		}
		if tokenColor.Equals(currentPlayerColor) == wantMatch {
			return i
		}
	}
	return -1
}

// firstNonBlackSpace returns the lowest space index for which spaceIsBlack
// is false.
func firstNonBlackSpace() int {
	for i := 0; i < boardSize; i++ {
		if !spaceIsBlack(i) {
			return i
		}
	}
	return -1
}

// moveMoveTokenGoldenFixtures builds the table of (state, move) fixtures the
// golden test sweeps every proposer against.
func moveMoveTokenGoldenFixtures(t *testing.T) []moveMoveTokenGoldenFixture {
	t.Helper()

	var fixtures []moveMoveTokenGoldenFixture

	// default: a fresh game (already in phasePlaying with the standard
	// 24-token layout — see this file's doc comment). TokenIndexToMove is
	// the current player's own token; SpaceIndex is one of its free
	// neighboring spaces (t.FreeNextSpaces). Legal for the current player.
	// Not every own-colored token has a free next space in the starting
	// layout (a token pinned at the board's outer edge, in its own
	// movement direction, has none), so this scans for the first one that
	// does rather than assuming the lowest-index match works.
	{
		game, state := newMoveMoveTokenGame(t)
		g, _ := concreteStates(state)
		playerColor, err := state.ImmutableCurrentPlayer().Reader().ImmutableEnumProp("Color")
		if err != nil {
			t.Fatalf("legal_golden: reading current player's Color: %v", err)
		}
		tokenIdx := -1
		var freeSpaces []int
		for i := 0; i < g.Spaces.Len(); i++ {
			c := g.Spaces.ImmutableComponentAt(i)
			if c == nil {
				continue
			}
			tokenColor, err := c.Values().Reader().ImmutableEnumProp("Color")
			if err != nil {
				t.Fatalf("legal_golden: reading token Color at Spaces[%d]: %v", i, err)
			}
			if !tokenColor.Equals(playerColor) {
				continue
			}
			tok := c.Values().(*token)
			spaces := tok.FreeNextSpaces(state, i)
			if len(spaces) > 0 {
				tokenIdx = i
				freeSpaces = spaces
				break
			}
		}
		if tokenIdx < 0 {
			t.Fatal("legal_golden: no current-player token with a free next space found")
		}
		move := moveMoveTokenMove(t, game, tokenIdx, freeSpaces[0])
		fixtures = append(fixtures, moveMoveTokenGoldenFixture{"default", game, move})
	}

	// emptySpace: TokenIndexToMove names an unoccupied space —
	// legal.ComponentPresentAtKey's Fail branch ("checkers.no_token_there").
	{
		game, state := newMoveMoveTokenGame(t)
		g, _ := concreteStates(state)
		emptyIdx := firstEmptySpace(g.Spaces)
		if emptyIdx < 0 {
			t.Fatal("legal_golden: no empty space found")
		}
		move := moveMoveTokenMove(t, game, emptyIdx, 0)
		fixtures = append(fixtures, moveMoveTokenGoldenFixture{"emptySpace", game, move})
	}

	// opponentToken: TokenIndexToMove names a space occupied by the
	// OPPOSING color — legal.ComponentPropEqualsCurrentPlayer's Fail branch
	// ("checkers.not_your_token").
	{
		game, state := newMoveMoveTokenGame(t)
		g, _ := concreteStates(state)
		playerColor, err := state.ImmutableCurrentPlayer().Reader().ImmutableEnumProp("Color")
		if err != nil {
			t.Fatalf("legal_golden: reading current player's Color: %v", err)
		}
		oppIdx := firstSpaceWithColorMatch(t, g.Spaces, playerColor, false)
		if oppIdx < 0 {
			t.Fatal("legal_golden: no space occupied by an opposing color")
		}
		move := moveMoveTokenMove(t, game, oppIdx, 0)
		fixtures = append(fixtures, moveMoveTokenGoldenFixture{"opponentToken", game, move})
	}

	// nonBlackDest: TokenIndexToMove names the current player's own token,
	// but SpaceIndex names a non-black space — the game-registered
	// "checkers.spaceIsBlack" predicate's Fail branch
	// ("checkers.black_spaces_only").
	{
		game, state := newMoveMoveTokenGame(t)
		g, _ := concreteStates(state)
		playerColor, err := state.ImmutableCurrentPlayer().Reader().ImmutableEnumProp("Color")
		if err != nil {
			t.Fatalf("legal_golden: reading current player's Color: %v", err)
		}
		tokenIdx := firstSpaceWithColorMatch(t, g.Spaces, playerColor, true)
		if tokenIdx < 0 {
			t.Fatal("legal_golden: no space occupied by the current player's own color")
		}
		whiteIdx := firstNonBlackSpace()
		if whiteIdx < 0 {
			t.Fatal("legal_golden: no non-black space found")
		}
		move := moveMoveTokenMove(t, game, tokenIdx, whiteIdx)
		fixtures = append(fixtures, moveMoveTokenGoldenFixture{"nonBlackDest", game, move})
	}

	// unreachableDest: TokenIndexToMove/SpaceIndex both legitimate (own
	// token, black space), but SpaceIndex is not reachable by a free move
	// or a capture — LegalCustom's residue Fail branch
	// ("checkers.illegal_dest"). Chosen as the diametrically opposite
	// corner of the board from the token, which is never adjacent to it.
	{
		game, state := newMoveMoveTokenGame(t)
		g, _ := concreteStates(state)
		playerColor, err := state.ImmutableCurrentPlayer().Reader().ImmutableEnumProp("Color")
		if err != nil {
			t.Fatalf("legal_golden: reading current player's Color: %v", err)
		}
		tokenIdx := firstSpaceWithColorMatch(t, g.Spaces, playerColor, true)
		if tokenIdx < 0 {
			t.Fatal("legal_golden: no space occupied by the current player's own color")
		}
		// spacesEnum is a boardWidth x boardWidth range enum; corner (0,0)
		// is index 0 and is black (spaceIsBlack(0) == true). If tokenIdx
		// itself happens to be 0, fall back to the opposite corner.
		dest := 0
		if tokenIdx == dest {
			dest = boardSize - 1
			if !spaceIsBlack(dest) {
				dest = boardSize - 2
			}
		}
		if !spaceIsBlack(dest) {
			t.Fatalf("legal_golden: chosen unreachable-dest space %d is not black", dest)
		}
		tok := g.Spaces.ImmutableComponentAt(tokenIdx).Values().(*token)
		for _, s := range tok.FreeNextSpaces(state, tokenIdx) {
			if s == dest {
				t.Fatalf("legal_golden: chosen unreachable-dest space %d is unexpectedly a free next space of token %d", dest, tokenIdx)
			}
		}
		for _, s := range tok.LegalCaptureSpaces(state, tokenIdx) {
			if s == dest {
				t.Fatalf("legal_golden: chosen unreachable-dest space %d is unexpectedly a legal capture space of token %d", dest, tokenIdx)
			}
		}
		move := moveMoveTokenMove(t, game, tokenIdx, dest)
		fixtures = append(fixtures, moveMoveTokenGoldenFixture{"unreachableDest", game, move})
	}

	// sameSpace: TokenIndexToMove == SpaceIndex. Both the legacy method and
	// legal.MaySwapComponentsByKey hit "i and j were the same" before the
	// component-present check.
	{
		game, state := newMoveMoveTokenGame(t)
		g, _ := concreteStates(state)
		playerColor, err := state.ImmutableCurrentPlayer().Reader().ImmutableEnumProp("Color")
		if err != nil {
			t.Fatalf("legal_golden: reading current player's Color: %v", err)
		}
		tokenIdx := firstSpaceWithColorMatch(t, g.Spaces, playerColor, true)
		if tokenIdx < 0 {
			t.Fatal("legal_golden: no space occupied by the current player's own color")
		}
		move := moveMoveTokenMove(t, game, tokenIdx, tokenIdx)
		fixtures = append(fixtures, moveMoveTokenGoldenFixture{"sameSpace", game, move})
	}

	return fixtures
}

// TestGoldenLegalMoveMoveToken is the design spec §9 "golden equivalence"
// test for checkers' declarative migration (spec §8): for every fixture
// above, cross every proposer worth distinguishing (the current player, a
// different concrete player, AdminPlayerIndex — a wildcard that passes the
// proposer check, ObserverPlayerIndex — which fails it) and assert the
// legacy oracle and the migrated move's real Legal() agree on nil-ness, and
// on message text too.
func TestGoldenLegalMoveMoveToken(t *testing.T) {
	fixtures := moveMoveTokenGoldenFixtures(t)

	for _, fixture := range fixtures {
		fixture := fixture
		state := fixture.game.CurrentState()
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
			t.Fatalf("legal_golden[%s]: could not find a non-current player", fixture.name)
		}

		proposers := map[string]boardgame.PlayerIndex{
			"currentPlayer": currentPlayer,
			"otherPlayer":   otherPlayer,
			"admin":         boardgame.AdminPlayerIndex,
			"observer":      boardgame.ObserverPlayerIndex,
		}

		for proposerName, proposer := range proposers {
			t.Run(fixture.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMoveMoveToken(fixture.move, state, proposer)
				actualErr := fixture.move.Legal(state, proposer)

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

// TestGoldenLegalMoveMoveTokenWrongPhase directly exercises the field-
// independent phase-check bucket, forcing the game back to phaseSetup
// (moveMoveToken is only legal in phasePlaying, contributed via
// moves.AddForPhase in main.go). This is the one predicate in
// moveMoveToken's plan that is field-INDEPENDENT (legal.InPhase reads no
// move.* path) — confirming it stays first in evaluation order, matching
// legacy's Default.legalInPhase, which also ran first (before
// CurrentPlayer's own proposer check).
func TestGoldenLegalMoveMoveTokenWrongPhase(t *testing.T) {
	game, state := newMoveMoveTokenGame(t)
	g := state.GameState().(*gameState)
	g.SetCurrentPhase(enum.EnumKey(phaseSetup))

	move := moveMoveTokenMove(t, game, 0, 0)

	legacyErr := legacyLegalMoveMoveToken(move, state, state.CurrentPlayerIndex())
	actualErr := move.Legal(state, state.CurrentPlayerIndex())

	if legacyErr == nil || actualErr == nil {
		t.Fatalf("expected both legacy and actual to be illegal in phaseSetup: legacy=%v actual=%v", legacyErr, actualErr)
	}
	if legacyErr.Error() != actualErr.Error() {
		t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
	}
}

/**************************************************
 *
 * movePlaceToken golden coverage
 *
 **************************************************/

// legacyLegalMovePlaceToken is a hand-copied snapshot of movePlaceToken's
// Legal() method exactly as it read before its PARTIAL declarative migration
// (see moves.go's comment block for the original source). Like the
// moveMoveToken oracle above, it deliberately does NOT call
// m.FixUpMulti.Legal(state, proposer): post-migration that dispatches through
// moves.Default.Legal, which detects the assembled plan and evaluates THAT
// instead of the frozen chain — defeating an independent oracle. Instead it
// hand-replicates the pieces of moves.Default.Legal's frozen chain that matter
// for these fixtures:
//   - legalInPhase via the same exported core helper that check itself calls
//     (boardgame.LegalInPhaseCheck — the Task 7 decision the moveMoveToken
//     oracle also relies on); movePlaceToken's only legal phase is phaseSetup
//     (set via moves.AddOrderedForPhase in main.go).
//   - legalStackConstraints is a no-op (movePlaceToken configures no
//     WithSourceProperty+WithDestinationProperty), so it is omitted.
//   - legalMoveInProgression is NOT a structural no-op (movePlaceToken IS
//     ordered), but it is unexported and cannot be called from this package
//     without dispatching through the migrated plan. It is omitted here on the
//     strength of a fixture invariant, not a shortcut: every fixture below is
//     a genuine mid-setup HISTORICAL version where the progression check
//     PASSES, and — because the frozen chain and the migrated "inProgression"
//     predicate call the one identical implementation — legacy and migrated
//     can never disagree on that atom's verdict, which in these fixtures is
//     always Pass, so it is never the deciding atom. (See this file's doc
//     comment.)
//
// FixUpMulti contributes no proposer/current-player check (only
// moves.CurrentPlayer does), so proposer is unused: every proposer sees the
// same verdict, which the sweep below confirms.
func legacyLegalMovePlaceToken(m *movePlaceToken, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := boardgame.LegalInPhaseCheck(state, []enum.EnumKey{enum.EnumKey(phaseSetup)}); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	first := game.UnusedTokens.ImmutableFirst()
	if first == nil {
		return errors.New("No more components to place")
	}

	if err := first.MayMoveToSlot(game.Spaces, m.TargetIndex.Value().Int()); err != nil {
		return err
	}

	if !spaceIsBlack(m.TargetIndex.Value().Int()) {
		return errors.New("The proposed space is not black")
	}

	return nil
}

// placeTokenMove returns a fresh "Place Token" move from game, with
// TargetIndex set.
func placeTokenMove(t *testing.T, game *boardgame.Game, target int) *movePlaceToken {
	t.Helper()
	move := game.MoveByName("Place Token")
	if move == nil {
		t.Fatal("legal_golden: no \"Place Token\" move found")
	}
	mv, ok := move.(*movePlaceToken)
	if !ok {
		t.Fatal("legal_golden: \"Place Token\" move was not a *movePlaceToken")
	}
	val, err := mv.ReadSetter().EnumProp("TargetIndex")
	if err != nil {
		t.Fatalf("legal_golden: reading TargetIndex enum prop: %v", err)
	}
	if err := val.SetValue(enum.EnumKey(target)); err != nil {
		t.Fatalf("legal_golden: setting TargetIndex: %v", err)
	}
	return mv
}

// firstInProgressSetupVersion returns the lowest historical version of game
// whose state is mid-phaseSetup (UnusedTokens still non-empty) AND has at
// least one occupied space, one empty black space, and one empty non-black
// space — everything the legal / occupiedDest / nonBlackDest fixtures need to
// target. See this file's doc comment for why a real historical setup version
// (not a mutated-back-to-setup finished game) is what keeps the contributed
// inPhase + inProgression base checks genuinely passing.
func firstInProgressSetupVersion(t *testing.T, game *boardgame.Game) int {
	t.Helper()
	for v := 0; v <= game.Version(); v++ {
		g, _ := concreteStates(game.State(v))
		if g.Phase.Value() != phaseSetup {
			continue
		}
		if g.UnusedTokens.NumComponents() == 0 {
			continue
		}
		if firstOccupiedSpace(g.Spaces) < 0 {
			continue
		}
		hasEmptyBlack, hasEmptyNonBlack := false, false
		for i := 0; i < g.Spaces.Len(); i++ {
			if g.Spaces.ImmutableComponentAt(i) != nil {
				continue
			}
			if spaceIsBlack(i) {
				hasEmptyBlack = true
			} else {
				hasEmptyNonBlack = true
			}
		}
		if hasEmptyBlack && hasEmptyNonBlack {
			return v
		}
	}
	t.Fatal("legal_golden: no in-progress phaseSetup version with the needed spaces found")
	return -1
}

// lastSetupVersion returns the highest historical version of game still in
// phaseSetup. That is the instant just before StartPhase fires: UnusedTokens
// is empty (every token placed) but the phase has not yet advanced, so
// proposing another Place Token is still in-progression — the only natural
// state that reaches movePlaceToken's "No more components to place" gate.
func lastSetupVersion(t *testing.T, game *boardgame.Game) int {
	t.Helper()
	last := -1
	for v := 0; v <= game.Version(); v++ {
		g, _ := concreteStates(game.State(v))
		if g.Phase.Value() == phaseSetup {
			last = v
		}
	}
	if last < 0 {
		t.Fatal("legal_golden: no phaseSetup version found")
	}
	return last
}

// firstEmptyBlackSpace returns the lowest index that is both empty in spaces
// and a black space, or -1 if none.
func firstEmptyBlackSpace(spaces boardgame.ImmutableStack) int {
	for i := 0; i < spaces.Len(); i++ {
		if spaces.ImmutableComponentAt(i) == nil && spaceIsBlack(i) {
			return i
		}
	}
	return -1
}

// placeTokenGoldenFixture is one (state, move) pair to check every proposer
// worth distinguishing against.
type placeTokenGoldenFixture struct {
	name  string
	state boardgame.ImmutableState
	move  *movePlaceToken
}

// placeTokenGoldenFixtures builds the table the golden test sweeps every
// proposer against. All four are drawn from historical versions of one
// finished game (see this file's doc comment): three from a single in-progress
// setup version differing only in TargetIndex, and one from the last setup
// version (UnusedTokens empty).
func placeTokenGoldenFixtures(t *testing.T) []placeTokenGoldenFixture {
	t.Helper()

	game, _ := newMoveMoveTokenGame(t)

	var fixtures []placeTokenGoldenFixture

	setupV := firstInProgressSetupVersion(t, game)
	inProgress := game.State(setupV)
	g, _ := concreteStates(inProgress)

	emptyBlack, emptyNonBlack := -1, -1
	for i := 0; i < g.Spaces.Len(); i++ {
		if g.Spaces.ImmutableComponentAt(i) != nil {
			continue
		}
		if spaceIsBlack(i) && emptyBlack < 0 {
			emptyBlack = i
		}
		if !spaceIsBlack(i) && emptyNonBlack < 0 {
			emptyNonBlack = i
		}
	}
	occupied := firstOccupiedSpace(g.Spaces)
	if emptyBlack < 0 || emptyNonBlack < 0 || occupied < 0 {
		t.Fatalf("legal_golden: in-progress setup version %d missing needed spaces (emptyBlack=%d emptyNonBlack=%d occupied=%d)", setupV, emptyBlack, emptyNonBlack, occupied)
	}

	// legal: TargetIndex is an empty black space — StackNotEmpty passes,
	// MayMoveToSlot passes (slot empty), spaceIsBlack passes. Legal.
	fixtures = append(fixtures, placeTokenGoldenFixture{"legal", inProgress, placeTokenMove(t, game, emptyBlack)})

	// occupiedDest: TargetIndex is an occupied space — StackNotEmpty passes,
	// then MayMoveToSlot fails ("slot N is already occupied"), so spaceIsBlack
	// is never reached. LegalCustom residue.
	fixtures = append(fixtures, placeTokenGoldenFixture{"occupiedDest", inProgress, placeTokenMove(t, game, occupied)})

	// nonBlackDest: TargetIndex is an empty non-black space — StackNotEmpty
	// passes, MayMoveToSlot passes (slot empty), then spaceIsBlack fails ("The
	// proposed space is not black"). LegalCustom residue, second gate.
	fixtures = append(fixtures, placeTokenGoldenFixture{"nonBlackDest", inProgress, placeTokenMove(t, game, emptyNonBlack)})

	// noTokensLeft: last setup version (UnusedTokens empty) — the declarative
	// StackNotEmpty precondition fails ("No more components to place") before
	// LegalCustom runs. TargetIndex is a still-empty black space so the failure
	// is unambiguously the emptiness gate, not the dest gates.
	noTok := game.State(lastSetupVersion(t, game))
	gNo, _ := concreteStates(noTok)
	dest := firstEmptyBlackSpace(gNo.Spaces)
	if dest < 0 {
		t.Fatal("legal_golden: no empty black space in the last setup version")
	}
	fixtures = append(fixtures, placeTokenGoldenFixture{"noTokensLeft", noTok, placeTokenMove(t, game, dest)})

	return fixtures
}

// TestGoldenLegalMovePlaceToken is the design spec §9 "golden equivalence"
// test for checkers' PARTIAL movePlaceToken migration (spec §8): for every
// fixture above, cross every proposer worth distinguishing and assert the
// legacy oracle and the migrated move's real Legal() agree on both nil-ness
// and message text. Because the exact legacy order was preserved (StackNotEmpty
// declarative first, then MayMoveToSlot, then spaceIsBlack in LegalCustom),
// there is NO message-ordering divergence to whitelist — verified empirically,
// including for the AdminPlayerIndex proposer: unlike pig's moveCountDie, these
// fixtures come from HISTORICAL versions rather than a ReadSetter mutation of
// the head state, so the fixup-move setup memo (keyed by state version) never
// goes stale against them, and admin recomputes fresh and matches.
func TestGoldenLegalMovePlaceToken(t *testing.T) {
	fixtures := placeTokenGoldenFixtures(t)

	for _, fixture := range fixtures {
		fixture := fixture
		state := fixture.state
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
			t.Fatalf("legal_golden[%s]: could not find a non-current player", fixture.name)
		}

		proposers := map[string]boardgame.PlayerIndex{
			"currentPlayer": currentPlayer,
			"otherPlayer":   otherPlayer,
			"admin":         boardgame.AdminPlayerIndex,
			"observer":      boardgame.ObserverPlayerIndex,
		}

		for proposerName, proposer := range proposers {
			t.Run(fixture.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMovePlaceToken(fixture.move, state, proposer)
				actualErr := fixture.move.Legal(state, proposer)

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
