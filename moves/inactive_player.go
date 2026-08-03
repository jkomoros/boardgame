package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
The three activation moves -- [ActivateInactivePlayer], [ActivateEmptySeat] and
[ActivateFilledSeat] -- are one verb with three seat filters. Their Apply
bodies are the same call, their DefaultsForState bodies differ only in which
seats they will select, and their Legal bodies differ only in the message they
give for a seat they refuse. That single implementation lives here, in
activationTarget; the three types supply a filter and nothing else.

They stay three named types rather than one type with an option because the
name at the call site IS the rule -- moves.ActivateFilledSeat says what it
does, where moves.ActivateSeat(moves.WithSeatFilter(...)) would need reading
twice -- and because ActivateInactivePlayer is the right answer for the many
games that never inactivate a seat. But there is exactly one implementation,
and adding a fourth filter should mean adding a filter, not a fourth copy.
*/

// seatRequirement is which seats an activationTarget is willing to touch.
type seatRequirement int

const (
	// seatAnySeat does not consult interfaces.Seater at all, so it works on a
	// playerState that has no seat behavior.
	seatAnySeat seatRequirement = iota
	// seatMustBeEmpty targets unfilled seats only.
	seatMustBeEmpty
	// seatMustBeFilled targets filled seats only.
	seatMustBeFilled
)

// activationTarget is the one thing the three activation moves disagree
// about: which seats they will activate, and what they say about a seat they
// refuse.
type activationTarget struct {
	requirement seatRequirement
	// wrongSeatErr is returned for a player whose seat fails requirement.
	wrongSeatErr string
	// alreadyActiveErr is returned for a player who is already active.
	alreadyActiveErr string
}

// seatMatches reports whether player's seat satisfies this target's
// requirement. A playerState that does not implement interfaces.Seater matches
// only seatAnySeat.
func (t activationTarget) seatMatches(player boardgame.ImmutableSubState) bool {
	if t.requirement == seatAnySeat {
		return true
	}
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return false
	}
	if t.requirement == seatMustBeFilled {
		return seat.SeatIsFilled()
	}
	return !seat.SeatIsFilled()
}

// defaultTarget is the shared DefaultsForState body: the first player this
// target would activate, or -1 if there is none. Returning -1 rather than
// leaving TargetPlayerIndex alone is deliberate -- Legal rejects it either
// way, and the caller only overwrites on a hit.
func (t activationTarget) defaultTarget(state boardgame.ImmutableState) boardgame.PlayerIndex {
	for i, p := range state.ImmutablePlayerStates() {
		if !t.seatMatches(p) {
			continue
		}
		if !behaviors.PlayerIsInactive(p) {
			continue
		}
		return boardgame.PlayerIndex(i)
	}
	return -1
}

// legalTarget is the shared seat half of Legal, and is also what Apply
// re-checks so that Apply resolves the target exactly the way Legal just
// checked it rather than resolving it a second, different way.
//
// Deliberately no EnsureValid, for all three filters. EnsureValid advances
// past any player GameDelegate.PlayerMayBeActive rejects, and an INACTIVE
// player is precisely such a player -- which is the only kind these moves
// exist to touch. Running the target through it walked away from the seat
// DefaultsForState had just chosen and onto some active player, so the
// inactive check below then failed with "Player is already active" and the
// move could never apply in any game where at least one player may be active.
// See SeatPlayer's playerIndex comment for the same hazard, same reason.
func (t activationTarget) legalTarget(state boardgame.ImmutableState, targetPlayerIndex boardgame.PlayerIndex) error {
	if targetPlayerIndex < 0 || int(targetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	if t.requirement != seatAnySeat {
		if _, ok := player.(interfaces.Seater); !ok {
			return errors.New("Player state didn't implement interfaces.Seater")
		}
	}
	if !t.seatMatches(player) {
		return errors.New(t.wrongSeatErr)
	}
	if !behaviors.PlayerIsInactive(player) {
		return errors.New(t.alreadyActiveErr)
	}
	return nil
}

// applyActivation is the shared Apply body for all three moves: the single
// call that is the entire point of the verb.
func applyActivation(state boardgame.State, targetPlayerIndex boardgame.PlayerIndex) error {
	if targetPlayerIndex < 0 || int(targetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	inactiver, ok := player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver")
	}
	inactiver.SetPlayerActive()
	return nil
}

// requirePlayerInactiver is the shared ValidConfiguration check.
func requirePlayerInactiver(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	if _, ok := player.(interfaces.PlayerInactiver); !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver. behaviors.InactivePlayer implements it for free")
	}
	return nil
}

// activateAnySeat is ActivateInactivePlayer's filter: every inactive player,
// seated or not. It is the union of [ActivateEmptySeat]'s filter and
// [ActivateFilledSeat]'s.
var activateAnySeat = activationTarget{
	requirement:      seatAnySeat,
	wrongSeatErr:     "",
	alreadyActiveErr: "The selected player is not inactive; there must be no inactive players to activate",
}

/*
ActivateInactivePlayer is a fixup move that activates EVERY currently inactive
player -- it is the union of [ActivateEmptySeat] and [ActivateFilledSeat], and
the right move for the many games that never inactivate a seat.

Choose it when your game marks players inactive only through
[SeatPlayer] (which inactivates whoever it seats, on the theory that someone
arriving mid-round should not join that round). Typically you would have it
either ALWAYS be legal -- meaning a player is included in play as soon as they
are seated -- or legal in the specific phase that ends a round, for example the
SetUpForNextRound phase.

If your game ALSO closes its empty seats for the round (via
[InactivateEmptySeat], which is what [DefaultRoundSetup] configures), do not
make this move always legal: it activates those empty seats too, and the round
starts waiting on players who do not exist. Use it inside round setup, where
reopening the seats is the point, and use [ActivateFilledSeat] for the
always-legal activation that lets a mid-game joiner start playing. That pairing
is what makes drop-in joining possible.

Designed to be used with behaviors.InactivePlayer, on its own directly. For
more on inactive players, see the package doc of boardgame/behaviors.

boardgame:codegen
*/
type ActivateInactivePlayer struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

// ActivatesSeatedPlayers returns true: this move activates inactive players
// whose seat is filled (among others), so it undoes SeatPlayer's inactivation
// of the players SeatPlayer actually seats. The framework looks for it at boot
// to verify that a game which inactivates the players it seats has some way of
// activating them again. Implements interfaces.SeatedPlayerActivator.
//
// The method is named for the CAPABILITY rather than for any one move type,
// because more than one move type has it -- [ActivateFilledSeat] returns true
// here too, and a method named after a concrete type would be lying at that
// call site.
func (a *ActivateInactivePlayer) ActivatesSeatedPlayers() bool {
	return true
}

// ActivatesEmptySeats returns true: this move is the union of
// ActivateFilledSeat and ActivateEmptySeat, so it reopens empty seats as a side
// effect of activating everyone. That is precisely the property this move's own
// doc comment warns about, and implementing
// interfaces.EmptySeatActivator is what turns that warning into
// InactivateEmptySeat's boot check.
func (a *ActivateInactivePlayer) ActivatesEmptySeats() bool {
	return true
}

// DefaultsForState sets TargetPlayerIndex to the next player who is currently
// marked as inactive, according to interfaces.PlayerInactiver.
func (a *ActivateInactivePlayer) DefaultsForState(state boardgame.ImmutableState) {
	if target := activateAnySeat.defaultTarget(state); target >= 0 {
		a.TargetPlayerIndex = target
	}
}

// Legal verifies that TargetPlayerIndex is set to a player whose InActive returns true.
func (a *ActivateInactivePlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}
	return activateAnySeat.legalTarget(state, a.TargetPlayerIndex)
}

// Apply sets the TargetPlayerIndex to be active via SetPlayerActive.
func (a *ActivateInactivePlayer) Apply(state boardgame.State) error {
	if err := activateAnySeat.legalTarget(state, a.TargetPlayerIndex); err != nil {
		return err
	}
	return applyActivation(state, a.TargetPlayerIndex)
}

// ValidConfiguration checks that player states implement interfaces.PlayerInactiver
func (a *ActivateInactivePlayer) ValidConfiguration(exampleState boardgame.State) error {
	return requirePlayerInactiver(exampleState)
}

// FallbackHelpText returns "Activates any players who are not currently active."
func (a *ActivateInactivePlayer) FallbackHelpText() string {
	return "Activates any players who are not currently active."
}

// FallbackName returns "Activate Inactive Player"
func (a *ActivateInactivePlayer) FallbackName(m *boardgame.GameManager) string {
	return "Activate Inactive Player"
}
