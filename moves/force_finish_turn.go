package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
)

/*
ForceFinishTurn is a FixUp variant of FinishTurn that bypasses the
TurnDone() check so a server (or other AdminPlayerIndex caller) can force
the current player's turn to end even when the player hasn't satisfied the
game's normal turn-end condition.

ForceFinishTurn is only meaningful for turn-based games where
CurrentPlayerIndex returns a valid player index and the GameState
implements [interfaces.CurrentPlayerSetter]. For simultaneous-action games
(where CurrentPlayerIndex returns AnyPlayerIndex), there is no "current
player" to advance — omit ForceFinishTurn from those games' move lists.
See examples/werewolf for a companion-capable game that correctly omits it.

The motivating case is the Table+Hand companion-mode SkipTurn host action
(spec §9.3): a phone drops mid-turn with required moves still pending; the
host wants to advance past them. Vanilla FinishTurn refuses because
TurnDone() returns an error; ForceFinishTurn accepts and lets the existing
Apply() do its work — which already calls ResetForTurnEnd via
interfaces.PlayerTurnFinisher (moves/finish_turn.go:90) and advances to the
next player.

Restricting Legal() to AdminPlayerIndex-only proposers is the security
contract — regular players can't bypass turn rules. This matches the
existing precedent in moves/seat_player.go where admin-only Legal guards
are how the framework distinguishes "server-initiated" from
"player-initiated" moves.

Usage in a server-initiated move proposal:

	game.ProposeMove(forceFinishTurnMove, boardgame.AdminPlayerIndex)

When registering in a game's move list via auto config, you MUST pass
moves.WithIsFixUp(false) AND moves.WithMoveName("Force Finish Turn"):

	auto.MustConfig(
	    new(moves.ForceFinishTurn),
	    moves.WithMoveName("Force Finish Turn"),
	    moves.WithIsFixUp(false),
	)

Both options are load-bearing:
  - The name override is required because ForceFinishTurn embeds
    FinishTurn, and the auto-configurator's default name derivation
    yields "Finish Turn" which clashes with the parent move's
    registration.
  - WithIsFixUp(false) is required because FinishTurn (the embedded
    type) is a FixUp, and FixUp moves are auto-proposed by the
    framework's fixup pipeline. Since ForceFinishTurn.Legal returns
    nil for AdminPlayerIndex (the same identity the fixup pipeline
    proposes under), an auto-proposed ForceFinishTurn would advance
    the current player every fixup pass — i.e. infinite recursion
    in the move loop. The host-SkipTurn flow proposes it directly,
    bypassing the fixup pipeline.

boardgame:codegen
*/
type ForceFinishTurn struct {
	FinishTurn
}

// Legal accepts the move only when proposed by AdminPlayerIndex. We
// deliberately do NOT call FinishTurn.Legal — that's exactly the check
// (TurnDone) we want to bypass. We don't need to call Default.Legal either
// because the admin-only gate is stricter (server-initiated moves don't
// have the same player-side legality concerns).
func (f *ForceFinishTurn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if proposer != boardgame.AdminPlayerIndex {
		return errors.New("ForceFinishTurn can only be proposed by AdminPlayerIndex")
	}
	currentPlayerIndex := state.CurrentPlayerIndex()
	if !currentPlayerIndex.Valid(state) || currentPlayerIndex < 0 {
		return errors.New("Current player is not valid")
	}
	return nil
}

// Apply is inherited from FinishTurn — it already calls ResetForTurnEnd
// via interfaces.PlayerTurnFinisher and advances the current player via
// CurrentPlayerSetter. We deliberately do NOT override; doing so would
// risk double-invoking ResetForTurnEnd.
