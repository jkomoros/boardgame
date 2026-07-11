package tictactoe

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//TODO: test this!!

//boardgame:codegen
type movePlaceToken struct {
	moves.CurrentPlayer
	//Which token to place the token
	Slot int
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

// movePlaceToken stays fully imperative (design spec §8 survey, Task 12):
// no catalog predicate can express its one custom gate. The move takes the
// FIRST component of players[TargetPlayerIndex].UnusedTokens (a growable,
// compacted stack — "first" is always index 0) and checks whether it
// MayMoveToSlot the move-field-indexed slot m.Slot in game.Slots — two
// DIFFERENT indices (a fixed 0 for the source, move.Slot for the
// destination). legal.MayMoveToSlot (legal/catalog_stack.go) only supports
// the "mirrored stacks, ONE shared idxField for both the source lookup and
// the destination slot" shape (design spec §8's memory
// HiddenCards/VisibleCards acid test is the shape it was built for); there
// is also no way to express a literal index like "0" in the catalog's path
// grammar (every LegalPropPath must name a real game/player/move property —
// see boardgame/legal_path.go's parseLegalPath). checkers' movePlaceToken
// (examples/checkers/moves.go) has this identical shape and is ALSO
// unmigrated, though for a different, independent reason (it embeds
// moves.FixUpMulti, a v1-seam-unsupported base type) — this catalog gap is
// the more fundamental of the two blockers and is flagged as design
// feedback in the Task 12 report. Left byte-for-byte unchanged from
// pre-Task-12.
func (m *movePlaceToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	first := players[m.TargetPlayerIndex.EnsureValid(state)].UnusedTokens.ImmutableFirst()
	if first == nil {
		return errors.New("there aren't any remaining tokens for the current player to place")
	}

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
