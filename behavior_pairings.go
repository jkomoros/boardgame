package boardgame

import (
	"github.com/jkomoros/boardgame/errors"
)

/*
Some behaviors only work in pairs: embedding one in your state creates an
obligation to configure a companion move, and the framework misbehaves silently
if you don't. Prose describing that obligation is not an API -- a boot error is.
This file collects those checks so that NewGameManager refuses a game whose
behaviors are half-wired, rather than shipping a game that compiles, boots, and
then does the wrong thing with no error.

The interfaces below are declared structurally rather than imported from
moves/interfaces, because moves imports boardgame and not the other way around.
Any move embedding the relevant moves.* type at any depth satisfies them, so
subclasses are detected too.
*/

// seater matches behaviors.Seat (interfaces.Seater).
type seater interface {
	SeatIsFilled() bool
}

// playerInactiver matches behaviors.InactivePlayer (interfaces.PlayerInactiver).
type playerInactiver interface {
	SetPlayerInactive()
	SetPlayerActive()
}

// seatPlayerMover matches moves.SeatPlayer (interfaces.SeatPlayerMover).
type seatPlayerMover interface {
	IsSeatPlayerMove() bool
}

// seatedPlayerActivator matches any move that can activate an inactive player
// in a FILLED seat -- moves.ActivateInactivePlayer and moves.ActivateFilledSeat
// both do (interfaces.SeatedPlayerActivator).
type seatedPlayerActivator interface {
	ActivatesSeatedPlayers() bool
}

// currentPlayerMover matches moves.CurrentPlayer (interfaces.CurrentPlayerMover).
type currentPlayerMover interface {
	IsCurrentPlayerMove() bool
}

// emptySeatActivator matches any move that clears the inactive flag on a seat
// nobody is sitting in -- moves.ActivateEmptySeat and moves.ActivateInactivePlayer
// both do (interfaces.EmptySeatActivator). moves.ActivateFilledSeat
// deliberately does not.
type emptySeatActivator interface {
	ActivatesEmptySeats() bool
}

// emptySeatInactivator matches moves.InactivateEmptySeat
// (interfaces.EmptySeatInactivator).
type emptySeatInactivator interface {
	InactivatesEmptySeats() bool
}

// unrestrictedMove matches moves.Default's InstalledUnrestricted, which reports
// whether a move was installed with neither a legal-phase restriction nor a
// move progression -- i.e. whether it is a candidate at every point of every
// phase.
type unrestrictedMove interface {
	InstalledUnrestricted() bool
}

// validateBehaviorPairings returns an error if the game's state behaviors imply
// a companion move that the game's move list does not contain. It is called
// during NewGameManager, after moves are installed and the example state
// exists.
func validateBehaviorPairings(exampleState ImmutableState, moveTypes []*moveType) error {

	examplePlayer := exampleState.ImmutablePlayerStates()[0]

	_, playerIsSeater := examplePlayer.(seater)
	_, playerIsInactiver := examplePlayer.(playerInactiver)

	var hasSeatPlayerMove, hasSeatedPlayerActivator, hasCurrentPlayerMove bool

	for _, mt := range moveTypes {
		testMove := mt.NewMove(exampleState)
		if m, ok := testMove.(seatPlayerMover); ok && m.IsSeatPlayerMove() {
			hasSeatPlayerMove = true
		}
		if m, ok := testMove.(seatedPlayerActivator); ok && m.ActivatesSeatedPlayers() {
			hasSeatedPlayerActivator = true
		}
		if m, ok := testMove.(currentPlayerMover); ok && m.IsCurrentPlayerMove() {
			hasCurrentPlayerMove = true
		}
	}

	// Seater implies a SeatPlayer move. Without it the server has no way to put
	// anyone in a seat, and NumSeatedActivePlayers() is permanently 0.
	if playerIsSeater && !hasSeatPlayerMove {
		return errors.New("PlayerState implements Seater (e.g. embeds behaviors.Seat) but no move implements IsSeatPlayerMove (e.g. moves.SeatPlayer). Without it, the server cannot seat players and NumSeatedActivePlayers() will always return 0")
	}

	// Seating + inactivating + current-player-gated play implies a move that
	// activates seated players.
	//
	// moves.SeatPlayer.Apply unconditionally calls SetPlayerInactive on whoever
	// it seats, on the theory that someone arriving mid-round should not join
	// that round. Only an activation move ever clears the flag. An
	// inactive player fails PlayerMayBeActive, which makes PlayerIndex.Valid
	// false, which makes moves.CurrentPlayer.Legal fail with "The specified
	// target player is not valid" -- permanently. So every seated player is a
	// silent permanent spectator and the game deadlocks the moment auto-seating
	// fills the seats.
	//
	// The current-player clause is what keeps this from being a false alarm: a
	// game whose moves are all moves.Default (examples/debuganimations) never
	// consults PlayerIndex.Valid, so inactivity costs it nothing and it needs no
	// activation move.
	//
	// Ordinary in-process tests cannot catch this, which is why it shipped in
	// four example games and is why the check belongs at boot:
	// manager.NewDefaultGame() leaves players unseated, because SeatPlayer needs
	// the server's rendezvous injection, so nothing is ever marked inactive and
	// every move stays legal.
	if playerIsInactiver && hasSeatPlayerMove && hasCurrentPlayerMove && !hasSeatedPlayerActivator {
		return errors.New("PlayerState implements PlayerInactiver (e.g. embeds behaviors.InactivePlayer) and the game configures a SeatPlayer move and at least one move gated on moves.CurrentPlayer, but no move implements ActivatesSeatedPlayers (e.g. moves.ActivateFilledSeat or moves.ActivateInactivePlayer). moves.SeatPlayer marks every player it seats as inactive and only an activation move undoes that, so every seated player would be permanently ineligible to be the current player and those moves permanently illegal. Add moves.ActivateFilledSeat, which activates only the seats a real player is in and is therefore safe to leave always legal even alongside moves.InactivateEmptySeat/moves.DefaultRoundSetup. moves.ActivateInactivePlayer is the alternative for a game that closes no seats, or inside a round-setup phase where reopening the empty ones is the point -- do not leave THAT one always legal in a game that inactivates its empty seats, because it reopens them and the round waits on players who do not exist")
	}

	return validateEmptySeatActivationLoop(exampleState, moveTypes)
}

/*
validateEmptySeatActivationLoop returns an error if the game configures both a
move that closes empty seats and a move that reopens them, with either one
installed unrestricted -- no legal phases and no move progression.

moves.ActivateInactivePlayer's doc comment has warned about this since the move
was written ("If your game ALSO closes its empty seats for the round ... do not
make this move always legal"), and the seating-deadlock remediation above says
it again in its last sentence. Both are prose, and prose about a required
invariant is not enforcement.

Both moves are FixUpMulti, so the framework auto-proposes both, and an
unrestricted move is a candidate at every point of every phase -- a move with no
progression group is deliberately not filtered by a phase's progression, "since
they may show up at any time in the phase". So the two undo each other, and the
game gets one of two failures depending on nothing more than the ORDER of the
entries in ConfigureMoves. Both were observed on a real fixture with this check
disabled (moves/empty_seat_loop_test.go documents each):

  - If the unrestricted move is reachable before the game leaves the phase, the
    fix-up pass ping-pongs until game.applyMove exhausts maxRecurseCount.
    NewDefaultGame itself fails with ErrTooManyFixUps -- the game cannot be
    created at all.
  - If it is not, the reopening simply wins, and play begins with empty seats
    marked ACTIVE. Turn order then hands the turn to a seat nobody is sitting
    in, and since no user is behind it, nobody can ever propose its move. That
    is verbatim the harm ActivateInactivePlayer's doc comment names: "the round
    starts waiting on players who do not exist."

Which of the two a creator gets is a coin flip decided by list order, which is
precisely why this belongs at boot rather than in a comment. It is also why no
working game can trip the check: both outcomes are unplayable.

The unrestricted clause is what keeps this from being a false alarm on the
idiomatic configuration. moves.DefaultRoundSetup deliberately puts
ActivateInactivePlayer and InactivateEmptySeat in ONE ordered progression, where
each fires at its own position and neither is legal at the other's; both
therefore carry a progression and neither is unrestricted. A game that
phase-scopes the two apart is fine for the same reason.

moves.ActivateFilledSeat -- the move the seating-deadlock error above recommends
leaving always legal, and the one that makes drop-in joining possible -- touches
no empty seat and does not implement interfaces.EmptySeatActivator, so it is
invisible here, exactly as intended.
*/
func validateEmptySeatActivationLoop(exampleState ImmutableState, moveTypes []*moveType) error {

	type installedMove struct {
		name         string
		unrestricted bool
	}

	var activators, inactivators []installedMove

	for _, mt := range moveTypes {
		testMove := mt.NewMove(exampleState)

		unrestricted := false
		if m, ok := testMove.(unrestrictedMove); ok {
			unrestricted = m.InstalledUnrestricted()
		}

		if m, ok := testMove.(emptySeatActivator); ok && m.ActivatesEmptySeats() {
			activators = append(activators, installedMove{mt.Name(), unrestricted})
		}
		if m, ok := testMove.(emptySeatInactivator); ok && m.InactivatesEmptySeats() {
			inactivators = append(inactivators, installedMove{mt.Name(), unrestricted})
		}
	}

	for _, activator := range activators {
		for _, inactivator := range inactivators {
			if !activator.unrestricted && !inactivator.unrestricted {
				continue
			}

			culprit := activator.name
			if inactivator.unrestricted && !activator.unrestricted {
				culprit = inactivator.name
			}

			return errors.New("the move \"" + inactivator.name +
				"\" inactivates empty seats and the move \"" + activator.name +
				"\" reopens them (it implements interfaces.EmptySeatActivator), but \"" + culprit +
				"\" is installed unrestricted -- no legal phases and no move progression, so it is a candidate at every point of every phase. " +
				"Both are fix-up moves, which the framework proposes on its own: one would close an empty seat and the other would immediately reopen it. " +
				"Depending only on the order of your ConfigureMoves entries, that either recurses until NewGame fails with ErrTooManyFixUps, or leaves play " +
				"starting with empty seats marked ACTIVE, so turn order stops on a seat nobody is sitting in and no one can ever propose its move. " +
				"Restrict them to phases where they cannot both apply, or put them in a single ordered move progression the way moves.DefaultRoundSetup does. " +
				"If what you wanted was an always-legal activation so a player who joins mid-game can take a turn, use moves.ActivateFilledSeat: " +
				"it activates only seats a real player is sitting in, so it never fights with an empty-seat inactivator")
		}
	}

	return nil
}
