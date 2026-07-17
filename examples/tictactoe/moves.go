package tictactoe

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//TODO: test this!!

//boardgame:codegen
type movePlaceToken struct {
	moves.CurrentPlayer
	//Which token to place the token
	Slot int `sanitize:"self:visible"`
}

func (m *movePlaceToken) DefaultsForState(state boardgame.ImmutableState) {
	game, _ := concreteStates(state)

	m.CurrentPlayer.DefaultsForState(state)

	//Default to setting a slot that's empty.
	for i, c := range game.Slots.Components() {
		if c == nil {
			m.Slot = i
			break
		}
	}
}

// movePlaceToken: PARTIALLY migrated (Task 7 survey re-check, design spec
// §6 §3). Task 12's survey found this move's Legal() body has two
// independent gates and left both fully imperative because no catalog
// predicate could express either. Design spec §6 §3's new
// "players[move.<Field>].<Prop>" path kind closes the gap for the first
// gate:
//
//  1. Token availability — "is players[TargetPlayerIndex].UnusedTokens
//     non-empty" — is now a plain legal.StackNotEmpty read on
//     "players[move.TargetPlayerIndex].UnusedTokens" (TargetPlayerIndex is
//     a PlayerIndex-typed field on the embedded moves.CurrentPlayer, so it
//     satisfies spec §3's <Field> requirement exactly). Declared via
//     WithLegalPreconditions in main.go; Legal() itself is deleted.
//
//  2. The MayMoveToSlot check is NOT migratable and survives as
//     LegalCustom residue, below: it compares a FIXED index (0 — "first" of
//     UnusedTokens) against move.Slot for the destination — two DIFFERENT
//     indices for source and destination. legal.MayMoveToSlot only
//     supports the mirrored-stacks shape (ONE shared idxField for both the
//     source lookup and the destination slot — design spec §8's memory
//     HiddenCards/VisibleCards acid test is the shape it was built for),
//     and the path grammar still has no literal-index syntax (spec §6
//     doesn't add one; every LegalPropPath must name a real
//     game/player/move/players[move.Field] property, never a literal like
//     "0" — see boardgame/legal_path.go's parseLegalPath). checkers'
//     movePlaceToken (examples/checkers/moves.go) has this identical
//     residual shape and is ALSO unmigrated, for the same fundamental
//     reason (plus, independently, its moves.FixUpMulti base type).
func (m *movePlaceToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	game, players := concreteStates(state)

	first := players[m.TargetPlayerIndex.EnsureValid(state)].UnusedTokens.ImmutableFirst()

	return first.MayMoveToSlot(game.Slots, m.Slot)

}

func (m *movePlaceToken) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	u := players[m.TargetPlayerIndex.EnsureValid(state)]

	if err := u.UnusedTokens.First().MoveTo(game.Slots, m.Slot); err != nil {
		return err
	}

	u.TokensToPlaceThisTurn--

	game.Phase.SetValue(phaseAfterFirstMove)

	return nil
}
