package boardgame

import (
	"testing"

	"github.com/workfit/tester/assert"
)

/*
This file tests LegalReadEvaluable/LegalReadsEvaluable (legal_evaluable.go),
the per-viewer sanitization half of Task 10's server ledger "evaluable"
formula (design spec §6). It reuses the package's own testGameState/
testPlayerState fixture (main_test.go) and their ALREADY-CONFIGURED default
struct tags — DrawDeck `sanitize:"len"`, MovesLeftThisTurn `sanitize:"hidden"`
(implicitly "other:hidden", per that field's own doc comment) — rather than
sanitizationTestConfig, since the defaults already give both a group-typed
(stack) and a scalar hidden-from-others property to exercise. The
players[move.<Field>].<Prop> tests below are the one exception: exercising
the FacetOccupancy/PolicyOrder cell of facetSurvives' table needs a stack
property sanitized with PolicyOrder, which no default struct tag provides,
so those reach for sanitizationTestConfig (see sanitization_test.go) the
same way TestSanitization does.
*/

func TestLegalReadEvaluableMoveAlwaysEvaluable(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	// A move.* read is evaluable for every viewer, including Observer, even
	// though the read here doesn't correspond to a real move field: move
	// fields are supplied by whoever frames the request, never sanitized.
	read := LegalRead{Path: "move.SomeField", Facet: LegalFacetValues}

	assert.For(t, "observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read)).Equals(true)
	assert.For(t, "player0").ThatActual(LegalReadEvaluable(state, 0, read)).Equals(true)
}

func TestLegalReadEvaluableAdminAlwaysEvaluable(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	// game.DrawDeck is sanitize:"len" for non-owners -- LegalFacetValues
	// would normally NOT survive that for Observer, but Admin is omniscient
	// (mirrors state.SanitizedForPlayer's own Admin bypass).
	read := LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}

	assert.For(t, "admin").ThatActual(LegalReadEvaluable(state, AdminPlayerIndex, read)).Equals(true)
}

func TestLegalReadEvaluableGameFacetSurvival(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	// game.DrawDeck carries `sanitize:"len"`: FacetValues does not survive
	// PolicyLen, but FacetCount does (facetSurvives' own table).
	valuesRead := LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}
	countRead := LegalRead{Path: "game.DrawDeck", Facet: LegalFacetCount}

	assert.For(t, "values for observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, valuesRead)).Equals(false)
	assert.For(t, "count for observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, countRead)).Equals(true)
}

func TestLegalReadEvaluableMissingPropDefaultsVisible(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	// A property with no sanitize tag at all (MyIntSlice) defaults to
	// PolicyVisible ("missing properties treated as PolicyVisible" per
	// subStateSanitizationTransformation's own doc comment), so any facet
	// survives for any viewer.
	read := LegalRead{Path: "game.MyIntSlice", Facet: LegalFacetValues}

	assert.For(t, "observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read)).Equals(true)
}

func TestLegalReadEvaluablePlayerPathSelfVsOther(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	gameState, _ := concreteStates(state)
	gameState.CurrentPlayer = 0

	// player.MovesLeftThisTurn is `sanitize:"hidden"` (implicitly
	// "other:hidden" -- see its doc comment): "player.X" resolves against
	// the CURRENT player (player 0 here), so player 0 viewing it sees their
	// own ("self") value, but any other viewer (including Observer) sees
	// the "other" policy, which is Hidden.
	read := LegalRead{Path: "player.MovesLeftThisTurn", Facet: LegalFacetValues}

	assert.For(t, "self (current player 0)").ThatActual(LegalReadEvaluable(state, 0, read)).Equals(true)
	assert.For(t, "other concrete player").ThatActual(LegalReadEvaluable(state, 1, read)).Equals(false)
	assert.For(t, "observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read)).Equals(false)
}

func TestLegalReadEvaluablePlayersAllRequiresEveryPlayer(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	gameState, _ := concreteStates(state)
	gameState.CurrentPlayer = 0

	// players[*].MovesLeftThisTurn: viewer 0 sees their OWN entry (index 0)
	// as self/visible, but every other player's entry is "other"/hidden
	// from viewer 0's perspective -- a quantifier over all players is only
	// evaluable if EVERY iterated player's value survives, so this must be
	// false even though the viewer's own slot would individually survive.
	hiddenRead := LegalRead{Path: "players[*].MovesLeftThisTurn", Facet: LegalFacetValues}
	assert.For(t, "players[*] with a hidden member").ThatActual(LegalReadEvaluable(state, 0, hiddenRead)).Equals(false)

	// players[*].IsFoo carries no sanitize tag at all, so it survives for
	// every player regardless of viewer.
	visibleRead := LegalRead{Path: "players[*].IsFoo", Facet: LegalFacetValues}
	assert.For(t, "players[*] all visible").ThatActual(LegalReadEvaluable(state, 0, visibleRead)).Equals(true)
}

func TestLegalReadEvaluableMalformedPathFailsClosed(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	read := LegalRead{Path: "nonsense-path-no-dot", Facet: LegalFacetValues}

	assert.For(t, "malformed path").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read)).Equals(false)
}

func TestLegalReadsEvaluableConjunction(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	visible := LegalRead{Path: "game.MyIntSlice", Facet: LegalFacetValues}
	hidden := LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}

	assert.For(t, "all visible").ThatActual(LegalReadsEvaluable(state, ObserverPlayerIndex, []LegalRead{visible})).Equals(true)
	assert.For(t, "one hidden fails the conjunction").ThatActual(LegalReadsEvaluable(state, ObserverPlayerIndex, []LegalRead{visible, hidden})).Equals(false)
	assert.For(t, "empty reads vacuously evaluable").ThatActual(LegalReadsEvaluable(state, ObserverPlayerIndex, nil)).Equals(true)
}

// TestLegalReadEvaluablePlayersMoveFieldAllVisible covers the
// pathPlayersMoveField case LegalReadEvaluable used to silently fall through
// to false for (the whole-round review's Fix 1): a "players[move.<Field>].X"
// read of a property with no sanitize tag at all (IsFoo, visible for every
// player regardless of viewer) must be evaluable true, exactly like the
// equivalent players[*].X read would be. The move field name itself
// (SomeField) doesn't need to correspond to a real move property here --
// LegalReadEvaluable only parses the path's grammar/kind, it never resolves
// the field against an actual move (see LegalReadEvaluable's doc comment on
// move.X reads being about grammar, not resolution).
func TestLegalReadEvaluablePlayersMoveFieldAllVisible(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	read := LegalRead{Path: "players[move.SomeField].IsFoo", Facet: LegalFacetValues}

	assert.For(t, "observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read)).Equals(true)
	assert.For(t, "player0").ThatActual(LegalReadEvaluable(state, 0, read)).Equals(true)
}

// TestLegalReadEvaluablePlayersMoveFieldHiddenForOther covers the same
// pathPlayersMoveField gap as the test above, but for a property that's
// PolicyHidden for non-owning viewers (MovesLeftThisTurn, `sanitize:"hidden"`
// == implicitly "other:hidden"): since the move field could point at ANY
// player, evaluability must require the facet to survive for EVERY player
// (mirroring pathPlayersAll's own "requires every player" test), so a
// non-owning viewer sees this as not evaluable even though the field might,
// at evaluation time, happen to resolve to that viewer's own player index.
func TestLegalReadEvaluablePlayersMoveFieldHiddenForOther(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	read := LegalRead{Path: "players[move.SomeField].MovesLeftThisTurn", Facet: LegalFacetValues}

	assert.For(t, "non-owning concrete viewer").ThatActual(LegalReadEvaluable(state, 1, read)).Equals(false)
	assert.For(t, "observer").ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read)).Equals(false)
}

func TestLegalReadEvaluableProposerIsConservativeAcrossPlayers(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	visible := LegalRead{Path: "proposer.IsFoo", Facet: LegalFacetValues}
	hiddenForOthers := LegalRead{Path: "proposer.MovesLeftThisTurn", Facet: LegalFacetValues}

	assert.For(t, "universally visible proposer property").ThatActual(LegalReadEvaluable(state, 0, visible)).Equals(true)
	assert.For(t, "own value is still conservative").ThatActual(LegalReadEvaluable(state, 0, hiddenForOthers)).Equals(false)
	assert.For(t, "admin remains omniscient").ThatActual(LegalReadEvaluable(state, AdminPlayerIndex, hiddenForOthers)).Equals(true)
}

// TestLegalReadEvaluablePlayersMoveFieldOccupancySurvivesOrder exercises the
// facet dimension of the pathPlayersMoveField fix: facetSurvives' table says
// LegalFacetOccupancy (unlike LegalFacetValues) survives PolicyOrder, so a
// players[move.<Field>].<Prop> read of an Order-sanitized stack property
// must be evaluable for FacetOccupancy even though FacetValues on the same
// read is not. This needs sanitizationTestConfig (see the file doc comment)
// since no default struct tag sanitizes a player Stack with PolicyOrder.
func TestLegalReadEvaluablePlayersMoveFieldOccupancySurvivesOrder(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()

	policy := &sanitizationTestConfig{
		Player: map[string]string{
			"Hand": "order",
		},
	}
	policy.Install(game.Manager())

	occupancyRead := LegalRead{Path: "players[move.SomeField].Hand", Facet: LegalFacetOccupancy}
	valuesRead := LegalRead{Path: "players[move.SomeField].Hand", Facet: LegalFacetValues}

	assert.For(t, "occupancy survives PolicyOrder for every player").ThatActual(LegalReadEvaluable(state, 1, occupancyRead)).Equals(true)
	assert.For(t, "values does not survive PolicyOrder").ThatActual(LegalReadEvaluable(state, 1, valuesRead)).Equals(false)
}
