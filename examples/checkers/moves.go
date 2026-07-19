package checkers

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type movePlaceToken struct {
	moves.FixUpMulti
	TargetIndex enum.RangeVal `enum:"spaces"`
}

func (m *movePlaceToken) DefaultsForState(state boardgame.ImmutableState) {

	game := state.ImmutableGameState().(*gameState)

	if game.UnusedTokens.NumComponents() <= 0 {
		return
	}

	nextToken := game.UnusedTokens.ComponentAt(0)

	nextTokenVals := nextToken.Values().(*token)

	//Red starts at top
	fromBottom := false

	if nextTokenVals.Color.Value() == colorBlack {
		fromBottom = true
	}

	startIndex := 0
	increment := 1
	endCondition := game.Spaces.Len()

	if fromBottom {
		startIndex = game.Spaces.Len() - 1
		increment = -1
		endCondition = 0
	}

	for i := startIndex; i != endCondition; i += increment {
		//We're only allowed to put tokens on black spaces
		if !spaceIsBlack(i) {
			continue
		}
		if game.Spaces.ComponentAt(i) == nil {
			m.TargetIndex.SetValue(enum.EnumKey(i))
			return
		}
	}

}

// Legal() is deliberately absent: this move opted into declarative legality
// via a PARTIAL migration (design spec §8). Only the FIRST of its three
// original gates is declarative; the other two stay imperative in LegalCustom
// below, in their ORIGINAL order. The original imperative body (kept only as
// legacyLegalMovePlaceToken, a private copy in legal_golden_test.go, for
// golden-equivalence testing) read:
//
//	if err := m.FixUpMulti.Legal(state, proposer); err != nil {
//		return err
//	}
//	game := state.ImmutableGameState().(*gameState)
//	first := game.UnusedTokens.ImmutableFirst()
//	if first == nil {
//		return errors.New("No more components to place")
//	}
//	if err := first.MayMoveToSlot(game.Spaces, m.TargetIndex.Value().Int()); err != nil {
//		return err
//	}
//	if !spaceIsBlack(m.TargetIndex.Value().Int()) {
//		return errors.New("The proposed space is not black")
//	}
//	return nil
//
// Migration mapping:
//   - "No more components to place" (first == nil, i.e.
//     UnusedTokens.NumComponents() == 0) is exactly legal.StackNotEmpty's Pass
//     condition, so it becomes legal.StackNotEmpty("game.UnusedTokens").
//     WithMessage("checkers.no_more_components"), declared via
//     WithLegalPreconditions in main.go's ConfigureMoves. FixUpMulti's phase +
//     move-progression checks are contributed base-first ahead of it,
//     unchanged. This single precondition is also what opts the move in,
//     satisfying the boot rule that a LegalCustom move must declare at least
//     one WithLegalPreconditions spec.
//   - MayMoveToSlot stays in LegalCustom because its source is the FIXED index
//     0 ("first" of UnusedTokens), rather than a move field. The declarative
//     legal.MayMoveToSlot supports distinct source/destination fields, but not
//     a literal source index. Its native error is returned verbatim.
//   - spaceIsBlack stays imperative in LegalCustom too, in its ORIGINAL order
//     (AFTER MayMoveToSlot). checkers already registers a
//     "checkers.spaceIsBlack" predicate (moveMoveToken uses it), but migrating
//     this gate to it would REORDER it — a declarative atom evaluates before
//     LegalCustom — changing which message wins for a move that fails both
//     gates; keeping the imperative call here preserves byte-for-byte order.
//     Its native error is returned verbatim.
//
// LegalCustom's first deref (game.UnusedTokens.ImmutableFirst()) is guaranteed
// non-nil: the StackNotEmpty precondition runs base-first (before LegalCustom)
// and rejects the move when UnusedTokens is empty, so control only reaches
// here with a non-nil First.
func (m *movePlaceToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	game := state.ImmutableGameState().(*gameState)

	first := game.UnusedTokens.ImmutableFirst()

	if err := first.MayMoveToSlot(game.Spaces, m.TargetIndex.Value().Int()); err != nil {
		return err
	}

	if !spaceIsBlack(m.TargetIndex.Value().Int()) {
		return errors.New("The proposed space is not black")
	}

	return nil
}

func (m *movePlaceToken) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	return game.UnusedTokens.First().MoveTo(game.Spaces, m.TargetIndex.Value().Int())
}

//boardgame:codegen
type moveMoveToken struct {
	moves.CurrentPlayer
	TokenIndexToMove enum.RangeVal `enum:"spaces"`
	SpaceIndex       enum.RangeVal `enum:"spaces"`
}

// Legal() is deliberately absent: this move opted into declarative legality
// (design spec §8's checkers acid test) via the moves.WithLegalPreconditions call
// in main.go's ConfigureMoves, plus the LegalCustom escape hatch just below
// for the one piece of residue the catalog cannot express (the capture-graph
// walk). moves.CurrentPlayer.Legal (promoted, since this type no longer
// overrides it) calls moves.Default.Legal, which detects the assembled plan
// and evaluates THAT instead of the frozen chain — the plan is: the phase
// check + proposer check (both contributed by moves.CurrentPlayer, unchanged
// from before), then the four authored gates below, then LegalCustom.
// The original imperative body (kept only as legacyLegalMoveMoveToken, a
// private copy in legal_golden_test.go, for golden-equivalence testing) read:
//
//	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
//		return err
//	}
//	p := state.ImmutableCurrentPlayer().(*playerState)
//	g := state.ImmutableGameState().(*gameState)
//	if err := g.Spaces.MaySwapComponentsByKey(m.TokenIndexToMove.Value(), m.SpaceIndex.Value()); err != nil {
//		return err
//	}
//	c := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value())
//	if c == nil {
//		return errors.New("That space does not have a component in it")
//	}
//	t := c.Values().(*token)
//	if !p.Color.Equals(t.Color) {
//		return errors.New("that token isn't your token to move")
//	}
//	if !spaceIsBlack(m.SpaceIndex.Value().Int()) {
//		return errors.New("you can only move to spaces that are black")
//	}
//	//If it's one of the legal spaces, great.
//	for _, space := range t.FreeNextSpaces(state, m.TokenIndexToMove.Value().Int()) {
//		if m.SpaceIndex.Value().Int() == space {
//			return nil
//		}
//	}
//	for _, space := range t.LegalCaptureSpaces(state, m.TokenIndexToMove.Value().Int()) {
//		if m.SpaceIndex.Value().Int() == space {
//			return nil
//		}
//	}
//	return errors.New("spaceIndex does not represent a legal space for that token to move to")
//
// Migration notes (design spec §8):
//   - "That space does not have a component in it" -> legal.ComponentPresentAtKey
//     ("checkers.no_token_there").
//   - "that token isn't your token to move" -> legal.ComponentPropEqualsCurrentPlayer
//     ("checkers.not_your_token").
//   - "you can only move to spaces that are black" -> the game-registered
//     "checkers.spaceIsBlack" predicate (ConfigurePredicateConstructors,
//     below), default template "checkers.black_spaces_only".
//   - g.Spaces.MaySwapComponentsByKey's i/j bounds+distinctness check is now
//     legal.MaySwapComponentsByKey, first in the authored plan so it retains
//     the original imperative check order. The FreeNextSpaces/
//     LegalCaptureSpaces walk stays hard-custom: no catalog predicate can
//     express a graph search, and by the time LegalCustom runs the remaining
//     three gates above have already guaranteed
//     TokenIndexToMove names a present, current-player-owned token and
//     SpaceIndex names a black space, so MaySwapComponentsByKey's bounds
//     checks are already complete. Every residue failure is therefore a
//     genuinely unreachable destination and uses "checkers.illegal_dest".
func (m *moveMoveToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	g := state.ImmutableGameState().(*gameState)

	c := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value())

	t := c.Values().(*token)

	//If it's one of the legal spaces, great.
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

	return legal.Errorf("checkers.illegal_dest", nil)
}

func (m *moveMoveToken) Apply(state boardgame.State) error {

	g := state.GameState().(*gameState)

	p := state.CurrentPlayer().(*playerState)

	if err := g.Spaces.SwapComponentsByKey(m.TokenIndexToMove.Value(), m.SpaceIndex.Value()); err != nil {
		return errors.New("Couldn't move token: " + err.Error())
	}

	startIndexes := m.TokenIndexToMove.RangeValue()

	if startIndexes == nil || len(startIndexes) != 2 {
		return errors.New("Couldn't get indexes for token space")
	}

	finishIndexes := m.SpaceIndex.RangeValue()

	if finishIndexes == nil || len(finishIndexes) != 2 {
		return errors.New("Couldn't get indexes for finish space")
	}

	middleIndexes := []int{
		finishIndexes[0] - startIndexes[0],
		finishIndexes[1] - startIndexes[1],
	}

	middleSpace := spacesEnum.RangeToValue(middleIndexes...)

	if middleSpace == enum.IllegalValue {
		return errors.New("Invalid result from range to value")
	}

	c := g.Spaces.ComponentAtKey(middleSpace)

	tokenCaptured := false

	if c != nil {

		tokenValues := c.Values().(*token)

		if !tokenValues.Color.Equals(p.Color) {
			tokenCaptured = true
			if err := g.Spaces.ComponentAtKey(middleSpace).MoveToLastSlot(p.CapturedTokens); err != nil {
				return errors.New("Couldn't capture token: " + err.Error())
			}
		}

	}

	//The turn is over if a token wasn't captured
	if !tokenCaptured {
		p.FinishedTurn = true
	} else {
		//The turn is also over if there isn't another cpature space to move
		//to.
		t := g.Spaces.ComponentAtKey(m.SpaceIndex.Value()).Values().(*token)
		if len(t.LegalCaptureSpaces(state, m.SpaceIndex.Value().Int())) == 0 {
			p.FinishedTurn = true
		}
	}

	return nil

}

//boardgame:codegen
type moveCrownToken struct {
	moves.DefaultComponent
}

func (m *moveCrownToken) Apply(state boardgame.State) error {
	g := state.GameState().(*gameState)

	c := g.Spaces.ComponentAt(m.ComponentIndex)

	if c == nil {
		return errors.New("No token at that space")
	}

	d := c.DynamicValues().(*tokenDynamic)

	d.Crowned = true

	return nil
}
