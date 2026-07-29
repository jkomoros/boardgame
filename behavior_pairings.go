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

// activateInactivePlayerMover matches moves.ActivateInactivePlayer
// (interfaces.ActivateInactivePlayerMover).
type activateInactivePlayerMover interface {
	IsActivateInactivePlayerMove() bool
}

// currentPlayerMover matches moves.CurrentPlayer (interfaces.CurrentPlayerMover).
type currentPlayerMover interface {
	IsCurrentPlayerMove() bool
}

// validateBehaviorPairings returns an error if the game's state behaviors imply
// a companion move that the game's move list does not contain. It is called
// during NewGameManager, after moves are installed and the example state
// exists.
func validateBehaviorPairings(exampleState ImmutableState, moveTypes []*moveType) error {

	examplePlayer := exampleState.ImmutablePlayerStates()[0]

	_, playerIsSeater := examplePlayer.(seater)
	_, playerIsInactiver := examplePlayer.(playerInactiver)

	var hasSeatPlayerMove, hasActivateInactivePlayerMove, hasCurrentPlayerMove bool

	for _, mt := range moveTypes {
		testMove := mt.NewMove(exampleState)
		if m, ok := testMove.(seatPlayerMover); ok && m.IsSeatPlayerMove() {
			hasSeatPlayerMove = true
		}
		if m, ok := testMove.(activateInactivePlayerMover); ok && m.IsActivateInactivePlayerMove() {
			hasActivateInactivePlayerMove = true
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

	// Seating + inactivating + current-player-gated play implies an
	// ActivateInactivePlayer move.
	//
	// moves.SeatPlayer.Apply unconditionally calls SetPlayerInactive on whoever
	// it seats, on the theory that someone arriving mid-round should not join
	// that round. Only moves.ActivateInactivePlayer ever clears the flag. An
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
	if playerIsInactiver && hasSeatPlayerMove && hasCurrentPlayerMove && !hasActivateInactivePlayerMove {
		return errors.New("PlayerState implements PlayerInactiver (e.g. embeds behaviors.InactivePlayer) and the game configures a SeatPlayer move and at least one move gated on moves.CurrentPlayer, but no move implements IsActivateInactivePlayerMove (e.g. moves.ActivateInactivePlayer). moves.SeatPlayer marks every player it seats as inactive and only ActivateInactivePlayer undoes that, so every seated player would be permanently ineligible to be the current player and those moves permanently illegal. Add moves.ActivateInactivePlayer -- always legal if the game has no rounds, or legal in whatever phase ends a round")
	}

	return nil
}
