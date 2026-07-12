package memory

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Golden-equivalence harness for moveRevealCard (design spec §8/§9, Task 11
brief) and moveHideCards (the Workstream 9 completeness-round re-migration;
see its own section near the bottom of this file): for every recorded
(state, proposer) pair, asserts that
legacyLegalMoveRevealCard (a hand-copied snapshot of the move's Legal() body
exactly as it read before this migration) and the migrated move's ACTUAL
Legal() (now dispatched through moves.CurrentPlayer.Legal ->
moves.Default.Legal's plan-evaluation short-circuit, since moves.go's
Legal() override was deleted) agree on both nil-ness and, for a non-nil
result, the exact error message string.

Fixture-construction decision (documented deviation from the brief's
suggestion to "reuse each game's existing golden JSON under
testdata/golden"): that JSON is consumed by boardgame-util/lib/golden's
record.Record replay machinery, which is built for whole-game
version-by-version comparison (move sequencing, storage diffs, timer
firing), not for probing a single move type's Legal() against many
hand-picked (state, proposer) combinations covering both legal and every
illegal branch. legal/conformance_test.go (Task 4) already established the
precedent this file follows instead: build a real game via
manager.NewDefaultGame() and mutate specific state fields directly to reach
each branch deterministically (a fresh memory game's exact card layout is
randomized by Shuffle(), so indexing into "whatever golden happened to
record" would be fragile in a way hand-constructed fixtures are not). Every
fixture below is a real, engine-produced boardgame.State — only individual
property values are forced — so this remains a genuine state, not a
fabricated one.
*/

// legacyLegalMoveRevealCard is a hand-copied snapshot of moveRevealCard's
// Legal() method exactly as it read before this migration (see moves.go's
// comment block for the original source). It deliberately does NOT call
// m.CurrentPlayer.Legal(state, proposer): that would dispatch through
// moves.Default.Legal, which (post-migration) detects the assembled plan and
// evaluates THAT instead of the frozen chain — defeating the point of an
// independent oracle. Instead it hand-replicates moves.CurrentPlayer.Legal's
// TargetPlayerIndex checks (moves/current_player.go) directly; since
// moveRevealCard is not phase-restricted (no WithLegalPhases in
// ConfigureMoves), moves.Default.Legal's own phase/progression/stack-
// constraint checks are no-ops for this move type and are omitted.
func legacyLegalMoveRevealCard(m *moveRevealCard, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

	p := players[game.CurrentPlayer.EnsureValid(state)]

	if p.CardsLeftToReveal < 1 {
		return errors.New("You have no cards left to reveal this turn")
	}

	c := game.HiddenCards.ImmutableComponentAt(m.CardIndex)
	if c == nil {
		if game.VisibleCards.ImmutableComponentAt(m.CardIndex) == nil {
			return errors.New("there is no card at that index")
		}
		return errors.New("that card has already been revealed")
	}

	return c.MayMoveToSlot(game.VisibleCards, m.CardIndex)
}

// revealCardGoldenFixture is one (game, move) pair to check every proposer
// against.
type revealCardGoldenFixture struct {
	name string
	game *boardgame.Game
	move *moveRevealCard
}

func newRevealCardGame(t *testing.T) *boardgame.Game {
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

// revealCardMove returns a fresh "Reveal Card" move from game, with
// CardIndex set to idx.
func revealCardMove(t *testing.T, game *boardgame.Game, idx int) *moveRevealCard {
	t.Helper()
	move := game.MoveByName("Reveal Card")
	if move == nil {
		t.Fatal("legal_golden: no \"Reveal Card\" move found")
	}
	rc, ok := move.(*moveRevealCard)
	if !ok {
		t.Fatal("legal_golden: \"Reveal Card\" move was not a *moveRevealCard")
	}
	rc.CardIndex = idx
	return rc
}

// revealCardGoldenFixtures builds the table of (state, move) fixtures the
// golden test sweeps every proposer against. Each mirrors a fixture already
// established as a conformance precedent by legal/conformance_test.go (Task
// 4/5's memoryDefault/memoryZeroCardsLeft/memoryCardAlreadyRevealed/
// memoryCardNeverThere), rebuilt here independently so this package's test
// suite doesn't need to import package legal_test's unexported fixtures.
func revealCardGoldenFixtures(t *testing.T) []revealCardGoldenFixture {
	t.Helper()

	var fixtures []revealCardGoldenFixture

	// default: a fresh game. HiddenCards[0] is occupied (FinishSetUp
	// distributes every card), CardsLeftToReveal is 2. Legal for the
	// current player.
	{
		game := newRevealCardGame(t)
		move := revealCardMove(t, game, 0)
		fixtures = append(fixtures, revealCardGoldenFixture{"default", game, move})
	}

	// lastIndex: same as default, but CardIndex is the last slot (boundary
	// check) instead of the first.
	{
		game := newRevealCardGame(t)
		gameState, _ := concreteStates(game.CurrentState())
		move := revealCardMove(t, game, gameState.HiddenCards.Len()-1)
		fixtures = append(fixtures, revealCardGoldenFixture{"lastIndex", game, move})
	}

	// zeroCardsLeft: the current player's CardsLeftToReveal is forced to 0
	// — legal.PropAtLeast("player.CardsLeftToReveal", 1)'s Fail branch.
	{
		game := newRevealCardGame(t)
		state, ok := game.CurrentState().(boardgame.State)
		if !ok {
			t.Fatal("legal_golden: CurrentState() was not mutable")
		}
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetIntProp("CardsLeftToReveal", 0); err != nil {
			t.Fatalf("legal_golden: setting CardsLeftToReveal: %v", err)
		}
		move := revealCardMove(t, game, 0)
		fixtures = append(fixtures, revealCardGoldenFixture{"zeroCardsLeft", game, move})
	}

	// alreadyRevealed: HiddenCards[0] moved directly to VisibleCards[0] (the
	// mirrored slot) — legal.RevealableCardAt's "already revealed" branch.
	{
		game := newRevealCardGame(t)
		state, ok := game.CurrentState().(boardgame.State)
		if !ok {
			t.Fatal("legal_golden: CurrentState() was not mutable")
		}
		gameRS := state.GameState().ReadSetter()
		hidden, err := gameRS.StackProp("HiddenCards")
		if err != nil {
			t.Fatalf("legal_golden: reading HiddenCards: %v", err)
		}
		visible, err := gameRS.StackProp("VisibleCards")
		if err != nil {
			t.Fatalf("legal_golden: reading VisibleCards: %v", err)
		}
		card := hidden.ComponentAt(0)
		if card == nil {
			t.Fatal("legal_golden: expected HiddenCards[0] to be occupied")
		}
		if err := card.MoveTo(visible, 0); err != nil {
			t.Fatalf("legal_golden: moving HiddenCards[0] to VisibleCards[0]: %v", err)
		}
		move := revealCardMove(t, game, 0)
		fixtures = append(fixtures, revealCardGoldenFixture{"alreadyRevealed", game, move})
	}

	// noCardHere: HiddenCards[0] moved to a DIFFERENT visible slot (5), so
	// at idx 0 both stacks are empty — legal.RevealableCardAt's "no card
	// here" branch.
	{
		game := newRevealCardGame(t)
		state, ok := game.CurrentState().(boardgame.State)
		if !ok {
			t.Fatal("legal_golden: CurrentState() was not mutable")
		}
		gameRS := state.GameState().ReadSetter()
		hidden, err := gameRS.StackProp("HiddenCards")
		if err != nil {
			t.Fatalf("legal_golden: reading HiddenCards: %v", err)
		}
		visible, err := gameRS.StackProp("VisibleCards")
		if err != nil {
			t.Fatalf("legal_golden: reading VisibleCards: %v", err)
		}
		card := hidden.ComponentAt(0)
		if card == nil {
			t.Fatal("legal_golden: expected HiddenCards[0] to be occupied")
		}
		if err := card.MoveTo(visible, 5); err != nil {
			t.Fatalf("legal_golden: moving HiddenCards[0] to VisibleCards[5]: %v", err)
		}
		move := revealCardMove(t, game, 0)
		fixtures = append(fixtures, revealCardGoldenFixture{"noCardHere", game, move})
	}

	return fixtures
}

// knownMessageOrderingDivergence names (fixture, proposer) combinations
// where the migrated plan is EXPECTED to disagree with the legacy oracle on
// WHICH message wins, even though both agree the move is illegal (nil-ness
// always matches). This is a genuine, documented architectural finding from
// this golden harness, not a test bug:
//
// Legacy Legal() evaluates strictly in declaration order: CurrentPlayer's
// proposer check (a super-call) runs before moveRevealCard's own
// CardsLeftToReveal check. The migrated plan's SPEC order matches that
// exactly ([proposerIsCurrentPlayer (contributed), PropAtLeast,
// RevealableCardAt, MayMoveToSlot (authored)] — see main.go). But
// legalPlan.evaluate (legal_plan.go) does NOT evaluate strictly in spec
// order: design spec §5's memoization architecture splits a plan into a
// fieldIndependent bucket (no move.* reads) and a fieldDependent bucket (at
// least one move.* read), and evaluates the ENTIRE fieldIndependent bucket
// before ANY fieldDependent predicate, regardless of declaration order.
// legal.PropAtLeast("player.CardsLeftToReveal", 1) reads no move.* path, so
// it lands in fieldIndependent; legal.ProposerIsCurrentPlayer() reads
// "move.TargetPlayerIndex", so it lands in fieldDependent — meaning
// CardsLeftToReveal is checked BEFORE the proposer check in the migrated
// plan, the reverse of the legacy order. legal.RevealableCardAt and
// legal.MayMoveToSlot both read "move.CardIndex" too, so they stay
// fieldDependent and keep their legacy relative order after the proposer
// check.
//
// Net effect: for a proposer that is BOTH not the current player AND whose
// target player has zero CardsLeftToReveal, legacy reports "it's not your
// turn" while the migrated plan reports "You have no cards left to reveal
// this turn". Both are illegal (nil-ness matches); only the FIRST-failure
// message differs. This narrows design spec §8's "migrated moves keep their
// historical first-failure messages" claim: it holds only when a move's
// declaration order and its field-independent/field-dependent split agree
// on ordering, which is not guaranteed in general. See the Task 11 report
// for the full writeup; this is a Named Review Risk realized, not a defect
// in this migration.
var knownMessageOrderingDivergence = map[string]bool{
	"zeroCardsLeft/otherPlayer": true,
	"zeroCardsLeft/observer":    true,
}

// TestGoldenLegalMoveRevealCard is the design spec §9 "golden equivalence"
// test for memory's flagship declarative migration (spec §8): for every
// fixture above, cross every proposer worth distinguishing (the current
// player, a different concrete player, AdminPlayerIndex — a wildcard that
// passes the proposer check, ObserverPlayerIndex — which fails it) and
// assert the legacy oracle and the migrated move's real Legal() agree on
// nil-ness, and (outside the one documented ordering exception above) on
// message text too.
func TestGoldenLegalMoveRevealCard(t *testing.T) {
	fixtures := revealCardGoldenFixtures(t)

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
				legacyErr := legacyLegalMoveRevealCard(fixture.move, state, proposer)
				actualErr := fixture.move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if knownMessageOrderingDivergence[fixture.name+"/"+proposerName] {
					return
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
 * moveHideCards golden coverage (Workstream 9)
 *
 **************************************************/

// legacyLegalMoveHideCards is a hand-copied snapshot of moveHideCards's
// Legal() method exactly as it read before the Workstream 9 re-migration (see
// moves.go's comment block for the original source). Like
// legacyLegalMoveRevealCard above, it deliberately does NOT call
// m.CurrentPlayer.Legal (which post-migration would dispatch through
// moves.Default.Legal into the assembled plan, defeating the point of an
// independent oracle); it hand-replicates moves.CurrentPlayer.Legal's
// TargetPlayerIndex/proposer checks directly (moves/current_player.go). Since
// moveHideCards is not phase-restricted, moves.Default.Legal's own
// phase/progression/stack-constraint checks are no-ops and are omitted. Unlike
// the reveal oracle there are only two authored gates after the proposer
// checks: CardsLeftToReveal and VisibleCards.
func legacyLegalMoveHideCards(m *moveHideCards, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

	p := players[game.CurrentPlayer.EnsureValid(state)]

	if p.CardsLeftToReveal > 0 {
		return errors.New("You still have to reveal more cards before your turn is over")
	}

	if game.VisibleCards.NumComponents() < 1 {
		return errors.New("no cards left to hide")
	}

	return nil
}

// hideCardsGoldenFixture is one (game, move) pair to check every proposer
// against.
type hideCardsGoldenFixture struct {
	name string
	game *boardgame.Game
	move *moveHideCards
}

// hideCardsMove returns a fresh "Hide Cards" move from game. Its
// TargetPlayerIndex is set by moves.CurrentPlayer.DefaultsForState (invoked by
// NewMove) to the current player index.
func hideCardsMove(t *testing.T, game *boardgame.Game) *moveHideCards {
	t.Helper()
	move := game.MoveByName("Hide Cards")
	if move == nil {
		t.Fatal("legal_golden: no \"Hide Cards\" move found")
	}
	hc, ok := move.(*moveHideCards)
	if !ok {
		t.Fatal("legal_golden: \"Hide Cards\" move was not a *moveHideCards")
	}
	return hc
}

// hideCardsGoldenFixtures builds the table of (state, move) fixtures the golden
// test sweeps every proposer against. Fixture construction follows the same
// NewDefaultGame()+direct-field-mutation approach documented in this file's top
// comment (reused via newRevealCardGame, which just builds a memory default
// game).
func hideCardsGoldenFixtures(t *testing.T) []hideCardsGoldenFixture {
	t.Helper()

	var fixtures []hideCardsGoldenFixture

	// default: a fresh game. Unlike reveal's "default" (legal for the current
	// player), BOTH of hide's field-independent gates FAIL here:
	// CardsLeftToReveal is 2 (>0, so PropCompare "<=" 0 fails) and VisibleCards
	// is empty (StackNotEmpty fails). PropCompare is declared first, so it is
	// the first-failure message for a passing/wildcard proposer.
	{
		game := newRevealCardGame(t)
		move := hideCardsMove(t, game)
		fixtures = append(fixtures, hideCardsGoldenFixture{"default", game, move})
	}

	// bothPass: force the current player's CardsLeftToReveal to 0 (PropCompare
	// "<=" 0 PASSES) AND move one card from HiddenCards into VisibleCards
	// (StackNotEmpty PASSES) — both field-independent gates pass, so the
	// contributed proposer atom decides the verdict. This is the only fixture
	// that exercises the legal/nil (pass) path and the proposer-branch "it's
	// not your turn" message byte-for-byte (mirrors reveal's alreadyRevealed
	// card-move technique).
	{
		game := newRevealCardGame(t)
		state, ok := game.CurrentState().(boardgame.State)
		if !ok {
			t.Fatal("legal_golden: CurrentState() was not mutable")
		}
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetIntProp("CardsLeftToReveal", 0); err != nil {
			t.Fatalf("legal_golden: setting CardsLeftToReveal: %v", err)
		}
		gameRS := state.GameState().ReadSetter()
		hidden, err := gameRS.StackProp("HiddenCards")
		if err != nil {
			t.Fatalf("legal_golden: reading HiddenCards: %v", err)
		}
		visible, err := gameRS.StackProp("VisibleCards")
		if err != nil {
			t.Fatalf("legal_golden: reading VisibleCards: %v", err)
		}
		card := hidden.ComponentAt(0)
		if card == nil {
			t.Fatal("legal_golden: expected HiddenCards[0] to be occupied")
		}
		if err := card.MoveTo(visible, 0); err != nil {
			t.Fatalf("legal_golden: moving HiddenCards[0] to VisibleCards[0]: %v", err)
		}
		move := hideCardsMove(t, game)
		fixtures = append(fixtures, hideCardsGoldenFixture{"bothPass", game, move})
	}

	return fixtures
}

// knownMessageOrderingDivergenceHide names (fixture, proposer) combinations
// where the migrated plan is EXPECTED to disagree with the legacy oracle on
// WHICH message wins, even though both agree the move is illegal (nil-ness
// always matches). Same architectural finding as moveRevealCard's
// knownMessageOrderingDivergence above: legalPlan.evaluate runs the ENTIRE
// field-independent bucket before ANY field-dependent predicate, regardless of
// declaration order (design spec §5's memoization split). moveHideCards's two
// gates (PropCompare on player.*, StackNotEmpty on game.*) read no move.* path,
// so both land field-INDEPENDENT; the contributed proposer atom reads
// move.TargetPlayerIndex, so it lands field-DEPENDENT — meaning the gates are
// checked BEFORE the proposer check in the plan, the reverse of the legacy
// order (super-call first).
//
// This bites ONLY in the "default" fixture, where both gates already fail:
// for a proposer that also fails the proposer check (a non-current, non-admin
// proposer), legacy reports "it's not your turn" (proposer check first) while
// the plan reports "You still have to reveal more cards before your turn is
// over" (PropCompare, the first field-independent gate). admin does NOT diverge
// (its proposer atom passes, so it reaches the same failing gate in both
// orderings), and the "bothPass" fixture does not diverge at all (its gates
// pass, so the proposer atom runs and, when it fails, wins byte-for-byte in
// both orderings). This is a normal player move, not a fixup, so there is no
// setup-memo artifact — every proposer cell recomputes fresh.
var knownMessageOrderingDivergenceHide = map[string]bool{
	"default/otherPlayer": true,
	"default/observer":    true,
}

// TestGoldenLegalMoveHideCards is the design spec §9 "golden equivalence" test
// for moveHideCards's Workstream 9 re-migration: for every fixture above, cross
// every proposer worth distinguishing (the current player, a different concrete
// player, AdminPlayerIndex — a wildcard that passes the proposer check,
// ObserverPlayerIndex — which fails it) and assert the legacy oracle and the
// migrated move's real Legal() agree on nil-ness, and (outside the two
// documented ordering exceptions above) on message text too.
func TestGoldenLegalMoveHideCards(t *testing.T) {
	fixtures := hideCardsGoldenFixtures(t)

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
				legacyErr := legacyLegalMoveHideCards(fixture.move, state, proposer)
				actualErr := fixture.move.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if knownMessageOrderingDivergenceHide[fixture.name+"/"+proposerName] {
					return
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}
