package boardgame

import (
	"testing"

	"github.com/workfit/tester/assert"
)

/*
legal_move_player_test.go pins the "move's player" resolution of "player.X"
paths (legalMovePlayerIndex, legal_path.go; design spec
docs/superpowers/specs/2026-11-22-legality-actor-paths-design.md). The rule is
the PlayerIndex.Equivalent wildcard algebra, narrowed one source at a time:
concrete current player → proposer → the player an admin acts on behalf of.
*/

// legalOnBehalfTestMove wraps an ordinary test move with the
// LegalOnBehalfOfer hook moves.CurrentPlayer/AnyPlayer provide. Embedding the
// Move interface (rather than a concrete type) keeps the wrapper honest: only
// the hook under test is added.
type legalOnBehalfTestMove struct {
	Move
	target PlayerIndex
}

func (m *legalOnBehalfTestMove) LegalOnBehalfOf() PlayerIndex { return m.target }

func TestLegalMovePlayerIndex(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()
	gameState, _ := concreteStates(state)
	plain := game.Manager().ExampleMoveByName("Test")
	onBehalf := &legalOnBehalfTestMove{Move: plain, target: 2}

	tests := []struct {
		name         string
		current      PlayerIndex
		proposer     PlayerIndex
		move         Move
		want         PlayerIndex
		wantFromMove bool
	}{
		// A move that names its own player is that player's move in every
		// phase: the proposer, or whom an admin acts for. The current
		// player is never consulted.
		{"own player, turn-based, current proposes", 1, 1, onBehalf, 1, false},
		{"own player, turn-based, out of turn", 1, 0, onBehalf, 0, false},
		{"own player, turn-based, admin on behalf", 1, AdminPlayerIndex, onBehalf, 2, true},
		{"own player, simultaneous", AnyPlayerIndex, 0, onBehalf, 0, false},
		{"own player, simultaneous, admin on behalf", AnyPlayerIndex, AdminPlayerIndex, onBehalf, 2, true},
		{"own player, observer phase", ObserverPlayerIndex, 3, onBehalf, 3, false},
		{"own player, observer proposer", 1, ObserverPlayerIndex, onBehalf, ObserverPlayerIndex, false},
		// Any other move is the current player's, unchanged in turn-based
		// phases for every proposer.
		{"plain, turn-based, current proposes", 1, 1, plain, 1, false},
		{"plain, turn-based, other proposes", 1, 0, plain, 1, false},
		{"plain, turn-based, admin proposes", 1, AdminPlayerIndex, plain, 1, false},
		// ...with a wildcard current player narrowed by the proposer.
		{"plain, simultaneous", AnyPlayerIndex, 1, plain, 1, false},
		{"plain, admin current", AdminPlayerIndex, 2, plain, 2, false},
		{"plain, simultaneous, admin has nobody to act for", AnyPlayerIndex, AdminPlayerIndex, plain, AdminPlayerIndex, false},
		{"plain, simultaneous, nil move", AnyPlayerIndex, AdminPlayerIndex, nil, AdminPlayerIndex, false},
		// Observer means nobody may move: never narrowed.
		{"plain, observer current", ObserverPlayerIndex, 0, plain, ObserverPlayerIndex, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gameState.CurrentPlayer = tc.current
			got, fromMove := legalMovePlayerIndex(state, tc.move, tc.proposer)
			assert.For(t).ThatActual(got).Equals(tc.want)
			assert.For(t).ThatActual(fromMove).Equals(tc.wantFromMove)
		})
	}
}

func TestResolvePlayerPathDuringSimultaneousPlay(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()
	gameState, players := concreteStates(state)
	gameState.CurrentPlayer = AnyPlayerIndex
	for i, p := range players {
		p.Score = 10 + i
	}
	plain := game.Manager().ExampleMoveByName("Test")

	for _, proposer := range []PlayerIndex{0, 1, 2} {
		ctx := LegalContext{State: state, Move: plain, ProposerPlayerIndex: proposer}
		val, _, err := ctx.ResolvePath("player.Score")
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(val).Equals(10 + int(proposer))
	}

	ctx := LegalContext{State: state, Move: &legalOnBehalfTestMove{Move: plain, target: 1}, ProposerPlayerIndex: AdminPlayerIndex}
	val, _, err := ctx.ResolvePath("player.Score")
	assert.For(t, "admin acting for player 1").ThatActual(err).IsNil()
	assert.For(t, "admin acting for player 1").ThatActual(val).Equals(11)

	for _, target := range []PlayerIndex{ObserverPlayerIndex, AnyPlayerIndex, AdminPlayerIndex, 3} {
		ctx := LegalContext{State: state, Move: &legalOnBehalfTestMove{Move: plain, target: target}, ProposerPlayerIndex: AdminPlayerIndex}
		_, _, err := ctx.ResolvePath("player.Score")
		assert.For(t, "non-concrete on-behalf target fails closed").ThatActual(err).IsNotNil()
	}

	ctx = LegalContext{State: state, Move: plain, ProposerPlayerIndex: AdminPlayerIndex}
	_, _, err = ctx.ResolvePath("player.Score")
	assert.For(t, "admin without an on-behalf player fails closed").ThatActual(err).IsNotNil()
}

// movePlayerCountingPlan is a one-step field-independent plan reading
// player.Score, counting evaluations so tests can observe memo hits.
func movePlayerCountingPlan(calls *int) *legalPlan {
	pred := &LegalPredicate{
		Name:  "countingPlayerScore",
		Reads: []LegalRead{{Path: "player.Score", Facet: LegalFacetValues}},
		Evaluate: func(ctx LegalContext) LegalVerdict {
			*calls++
			val, _, err := ctx.ResolvePath("player.Score")
			if err != nil {
				return LegalVerdict{Outcome: LegalUnknown, Reason: err.Error()}
			}
			if val.(int) > 10 {
				return LegalVerdict{Outcome: LegalPass}
			}
			return LegalVerdict{Outcome: LegalFail, Reason: "low score"}
		},
	}
	return buildLegalPlanFromPredicates("MovePlayerMemo", []*LegalPredicate{pred}, nil, nil)
}

// TestLegalPlanMemoKeysOnMovePlayer: during simultaneous play an admin
// acting for two different players must not share a memoized player.X
// verdict, even though state version and proposer are identical.
func TestLegalPlanMemoKeysOnMovePlayer(t *testing.T) {
	game := newLegalMemoTestGame(t)
	state := game.CurrentState()
	gameState, players := concreteStates(state)
	gameState.CurrentPlayer = AnyPlayerIndex
	players[0].Score = 0
	players[1].Score = 20

	calls := 0
	plan := movePlayerCountingPlan(&calls)
	step := plan.ordered[0]
	assert.For(t, "player.X stays memoizable").ThatActual(step.fieldDependent).Equals(false)
	assert.For(t).ThatActual(step.readsMovePlayer).Equals(true)
	assert.For(t).ThatActual(plan.independentReadsMovePlayer).Equals(true)

	plain := game.MoveByName("Legal Memo Test Move")
	forPlayer := func(p PlayerIndex) LegalContext {
		return LegalContext{State: state, Move: &legalOnBehalfTestMove{Move: plain, target: p}, ProposerPlayerIndex: AdminPlayerIndex}
	}

	v, _ := plan.evaluate(forPlayer(0), false)
	assert.For(t).ThatActual(v.Outcome).Equals(LegalFail)
	v, _ = plan.evaluate(forPlayer(1), false)
	assert.For(t, "different on-behalf player is a memo miss").ThatActual(v.Outcome).Equals(LegalPass)
	assert.For(t).ThatActual(calls).Equals(2)
	v, _ = plan.evaluate(forPlayer(1), false)
	assert.For(t, "same on-behalf player is a memo hit").ThatActual(v.Outcome).Equals(LegalPass)
	assert.For(t).ThatActual(calls).Equals(2)
}

// TestLegalLedgerMarksOnBehalfPlayerReadsFieldDependent: a player.X entry is
// provisional on move fields exactly when its player came from the move.
func TestLegalLedgerMarksOnBehalfPlayerReadsFieldDependent(t *testing.T) {
	game := newLegalMemoTestGame(t)
	state := game.CurrentState()
	gameState, _ := concreteStates(state)
	calls := 0
	plan := movePlayerCountingPlan(&calls)
	move := &legalOnBehalfTestMove{Move: game.MoveByName("Legal Memo Test Move"), target: 1}

	gameState.CurrentPlayer = 0
	_, entries := plan.evaluate(LegalContext{State: state, Move: move, ProposerPlayerIndex: 1}, true)
	assert.For(t, "player proposes").ThatActual(entries[0].FieldDependent).Equals(false)

	_, entries = plan.evaluate(LegalContext{State: state, Move: move, ProposerPlayerIndex: AdminPlayerIndex}, true)
	assert.For(t, "admin on behalf").ThatActual(entries[0].FieldDependent).Equals(true)

	plain := game.MoveByName("Legal Memo Test Move")
	_, entries = plan.evaluate(LegalContext{State: state, Move: plain, ProposerPlayerIndex: AdminPlayerIndex}, true)
	assert.For(t, "a move without its own player reads the current player").ThatActual(entries[0].FieldDependent).Equals(false)
}

// TestLegalReadEvaluablePlayerPathDuringSimultaneousPlay: with a wildcard
// current player, player.X is judged like proposer.X — conservatively over
// every player — because the viewer-keyed API has no trusted proposer.
func TestLegalReadEvaluablePlayerPathDuringSimultaneousPlay(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()
	gameState, _ := concreteStates(state)

	visible := LegalRead{Path: "player.IsFoo", Facet: LegalFacetValues}
	hiddenForOthers := LegalRead{Path: "player.MovesLeftThisTurn", Facet: LegalFacetValues}

	for _, current := range []PlayerIndex{AnyPlayerIndex, AdminPlayerIndex} {
		gameState.CurrentPlayer = current
		assert.For(t, "universally visible").ThatActual(LegalReadEvaluable(state, 0, visible)).Equals(true)
		assert.For(t, "own value is still conservative").ThatActual(LegalReadEvaluable(state, 0, hiddenForOthers)).Equals(false)
		assert.For(t, "admin remains omniscient").ThatActual(LegalReadEvaluable(state, AdminPlayerIndex, hiddenForOthers)).Equals(true)
	}

	gameState.CurrentPlayer = ObserverPlayerIndex
	assert.For(t, "nobody may move").ThatActual(LegalReadEvaluable(state, 0, visible)).Equals(false)
}

// TestLegalLedgerEvaluabilityForMovesMadeForTheirOwnPlayer: a move made for
// its own player that nothing ties to the current player (an AnyPlayer-style
// move) reads "player.X" as the proposer, so the ledger must judge those
// reads like proposer.X. A CurrentPlayer-style move (its plan carries
// proposerIsCurrentPlayer) and a move without a player of its own keep the
// precise current-player judgement.
func TestLegalLedgerEvaluabilityForMovesMadeForTheirOwnPlayer(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()
	gameState, _ := concreteStates(state)
	gameState.CurrentPlayer = 0

	read := LegalRead{Path: "player.MovesLeftThisTurn", Facet: LegalFacetValues}
	pass := func(ctx LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalPass} }
	reader := &LegalPredicate{Name: "readsHiddenPlayerProp", Reads: []LegalRead{read}, Evaluate: pass}
	turnCheck := &LegalPredicate{Name: "proposerIsCurrentPlayer", Evaluate: pass}

	plain := game.Manager().ExampleMoveByName("Test")
	ownPlayer := &legalOnBehalfTestMove{Move: plain, target: 0}

	entryFor := func(move Move, preds ...*LegalPredicate) LegalVerdictEntry {
		plan := buildLegalPlanFromPredicates("", preds, nil, move)
		_, entries := plan.evaluate(LegalContext{State: state, Move: move, ProposerPlayerIndex: 0}, true)
		return entries[len(entries)-1]
	}

	anyPlayerStyle := entryFor(ownPlayer, reader)
	assert.For(t).ThatActual(anyPlayerStyle.MovePlayerIsProposer).Equals(true)
	assert.For(t, "AnyPlayer-style: conservative").ThatActual(anyPlayerStyle.ReadsEvaluable(state, 0)).Equals(false)

	currentPlayerStyle := entryFor(ownPlayer, turnCheck, reader)
	assert.For(t).ThatActual(currentPlayerStyle.MovePlayerIsProposer).Equals(false)
	assert.For(t, "CurrentPlayer-style: own value visible").ThatActual(currentPlayerStyle.ReadsEvaluable(state, 0)).Equals(true)
	assert.For(t, "CurrentPlayer-style: hidden from others").ThatActual(currentPlayerStyle.ReadsEvaluable(state, 1)).Equals(false)

	plainEntry := entryFor(plain, reader)
	assert.For(t).ThatActual(plainEntry.MovePlayerIsProposer).Equals(false)
	assert.For(t, "plain: own value visible").ThatActual(plainEntry.ReadsEvaluable(state, 0)).Equals(true)
}
