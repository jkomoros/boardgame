package tictactoe

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//TODO: test this!!

// movePlaceToken is the reusable "play a piece into the slot you chose" verb
// plus this game's own bookkeeping. moves.MoveComponentToSlot supplies both
// stacks (the source is on the PLAYER, which is exactly what no reusable move
// could reach before), the complete legality check, and the transfer.
//
//boardgame:codegen
type movePlaceToken struct {
	moves.MoveComponentToSlot
	//Which slot to place the token in. moves.MoveComponentToSlot discovers
	//this field automatically: it is the move's only int property.
	Slot int `sanitize:"self:visible"`
}

func (m *movePlaceToken) DefaultsForState(state boardgame.ImmutableState) {
	game, _ := concreteStates(state)

	m.MoveComponentToSlot.DefaultsForState(state)

	//Default to setting a slot that's empty.
	for i, c := range game.Slots.Components() {
		if c == nil {
			m.Slot = i
			break
		}
	}
}

// Legal() is deliberately absent, and so is the LegalCustom residue this
// move used to carry. Its two gates were "the current player has a token
// left" and "that token may go into move.Slot" -- which together are exactly
// what moves.MoveComponentToSlot contributes as
// legal.MayMoveFirstToSlot("players[move.TargetPlayerIndex].UnusedTokens",
// "game.Slots", "move.Slot"). The residue existed because the catalog could
// not name a LITERAL source index ("the first of UnusedTokens") -- the gap
// this file's previous comment described, and which checkers' movePlaceToken
// independently described too. legal.MayMoveFirstToSlot closes it: "the first
// component" is now a named concept in the catalog rather than the integer 0
// smuggled in as a property path.
//
// The one visible consequence is the empty-source message. It used to come
// from an authored legal.StackNotEmpty carrying "tictactoe.no_tokens_left";
// it now comes from mayMoveFirstToSlot's own no-component branch, so
// ConfigureLegalTemplates (main.go) overrides legal.TemplateNoComponentToMove
// with the same string instead.

// Apply super-calls moves.MoveComponentToSlot.Apply for the transfer itself,
// then does this game's own bookkeeping.
func (m *movePlaceToken) Apply(state boardgame.State) error {

	if err := m.MoveComponentToSlot.Apply(state); err != nil {
		return err
	}

	game, players := concreteStates(state)

	players[m.TargetPlayerIndex.EnsureValid(state)].TokensToPlaceThisTurn--

	game.Phase.SetValue(phaseAfterFirstMove)

	return nil
}
