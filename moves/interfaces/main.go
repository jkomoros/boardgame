/*

Package interfaces is a collection of interfaces that your objects can implement
to configure how the moves package's base moves operate.

Factored into a separate package primarily just to make the moves package more
clear about which structs are the main embeddable moves.

*/
package interfaces

import (
	"github.com/jkomoros/boardgame"
)

//AllowMultipleInProgression is an interface that moves should implement if they
//want to affirmatively communicate to moves.Default that in a move progression
//is it legal to apply multiple. If the move does not implement this interface
//then it is considered to only allow one.
type AllowMultipleInProgression interface {
	//AllowMultipleInProgression should return true if the given move is
	//allowed to apply multiple times in order in a move progression.
	AllowMultipleInProgression() bool
}

//LegalComponent should be implemented by ComponentValues that will be used
//with moves.DefaultComponent.
type LegalComponent interface {
	//Legal will be called on each component, with a legalType related to the
	//move in question (or 0 if WithLegalType hasn't been called). This allows
	//the same component values to participate in multiple
	//moves.DefaultComponent move types. Idiomatically legalType should be a
	//value in an enum created for the purpose of disambiguating different
	//move types to check for legality for. Legal should return nil if it is
	//legal, or an error if the component is not legal.
	Legal(state boardgame.ImmutableState, legalType int) error
}

//PlayerStacker should be implemented by your embedding Move if you embed
//DealComponents. It will be consulted to figure out where the PlayerStack is
//to deal a component to.
type PlayerStacker interface {
	PlayerStack(playerState boardgame.SubState) boardgame.Stack
}

//GameStacker should be implemented by your emedding Move if you embed
//DealComponents. It will be consulted to figure out where to draw the
//components from to deal to players.
type GameStacker interface {
	GameStack(gameState boardgame.SubState) boardgame.Stack
}

//SourceStacker should be implemented by moves that need an input stack to
//operate on as primary/source, for example ShuffleStack.
type SourceStacker interface {
	SourceStack(state boardgame.State) boardgame.Stack
}

//DestinationStacker should be implemented by moves that need a destination
//stack to operate on as primary/source, for example ApplyUntilCount.
type DestinationStacker interface {
	DestinationStack(state boardgame.State) boardgame.Stack
}

//CurrentPlayerSetter should be implemented by gameStates that use FinishTurn.
//behaviors.CurrentPlayerBehavior satifies this.
type CurrentPlayerSetter interface {
	SetCurrentPlayer(currentPlayer boardgame.PlayerIndex)
}

//TargetCounter should be implemented by moves who should be legal until a
//TargetCount has been reached.
type TargetCounter interface {
	TargetCount(state boardgame.ImmutableState) int
}

//PlayerTurnFinisher is the interface your playerState is expected to adhere
//to when you use FinishTurn.
type PlayerTurnFinisher interface {
	//TurnDone should return nil when the turn is done, or a descriptive error
	//if the turn is not done.
	TurnDone() error
	//ResetForTurnStart will be called when this player begins their turn.
	ResetForTurnStart() error
	//ResetForTurnEnd will be called right before the CurrentPlayer is
	//advanced to the next player.
	ResetForTurnEnd() error
}

//RoundRobinProperties should be implemented by your GameState if you use any of
//the RoundRobin moves, including StartRoundRobin. You don't have to do anything
//we these other than store them to a property in your gameState and then return
//them via the getters. Generally you simply embed behaviors.RoundRobin to
//satisfy this interface for free.
type RoundRobinProperties interface {
	//The last successfully applied round robin player
	RoundRobinLastPlayer() boardgame.PlayerIndex
	//The index of the player we started the round robin on.
	RoundRobinStarterPlayer() boardgame.PlayerIndex
	//How many complete times around the round robin we've been. Increments
	//each time NextRoundRobinPlayer is StarterPlayer.
	RoundRobinRoundCount() int
	//RoundRobinHasStarted is true if the first move of a RoundRobin has been
	//applied.
	RoundRobinHasStarted() bool

	SetRoundRobinLastPlayer(nextPlayer boardgame.PlayerIndex)
	SetRoundRobinStarterPlayer(index boardgame.PlayerIndex)
	SetRoundRobinRoundCount(count int)
	SetRoundRobinHasStarted(hasStarted bool)
}

//ConditionMetter should be implemented by moves that subclass
//moves.ApplyUntil.
type ConditionMetter interface {
	//ConditionMet should return nil if the condition has been met, or an
	//error describing why the condition has not yet been met.
	ConditionMet(state boardgame.ImmutableState) error
}

//RoundRobinActioner should be implemented by any moves that embed a
//RoundRobin move. It's the action that will be called on the player who is
//next in the round robin.
type RoundRobinActioner interface {
	//RoundRobinAction should do the action for the round robin to given player.
	RoundRobinAction(playerState boardgame.SubState) error
}

//CurrentPhaseSetter should be implemented by you gameState to set the
//CurrentPhase. Must be implemented if you use the StartPhase move type.
//behaviors.PhaseBehavior implements this.
type CurrentPhaseSetter interface {
	SetCurrentPhase(int)
}

//BeforeLeavePhaser is an interface to implement on GameState if you want to
//do some action on state before leaving the given phase.
type BeforeLeavePhaser interface {
	BeforeLeavePhase(phase int, state boardgame.State) error
}

//BeforeEnterPhaser is an interface to implement on GameState if you want to
//do some action on state just before entering the givenn state.
type BeforeEnterPhaser interface {
	BeforeEnterPhase(phase int, state boardgame.State) error
}

//PlayerInactiver is for PlayerStates that encode whether that player is
//Inactive and whether that might be changed. See the package doc of
//boardgame/behaviors for more on the notion of inactive players.
type PlayerInactiver interface {
	IsInactive() bool
	SetPlayerInactive()
	SetPlayerActive()
}

//Seater is for PlayerStates that interace with moves like SeatPlayer. See the
//package doc of boardgame/behaviors for more on the notion of seats.
type Seater interface {
	SeatIsFilled() bool
	SeatIsClosed() bool
	SetSeatFilled()
	SetSeatClosed()
}

//SeatPlayerMover should be implemented for moves that are SeatPlayer moves,
//returing true from IsSeatPlayerMove(). Typically you use moves.SeatPlayer
//directly, which implements this interface, but you might also want to embed
//that move in another move to, for example, change the DefaultsForState logic.
//This is the way that even those embedded moves can be detected as being
//intended as SeatPlayer moves.
type SeatPlayerMover interface {
	IsSeatPlayerMove() bool
}

//SeatPlayerSignaler is the way that moves.SeatPlayer and the server coordinate
//about where to seat a player.
type SeatPlayerSignaler interface {
	//The index of the seat to sit the palyer in. If it's not a valid index
	//(e.g. AdminPlayerIndex), will use the next available seat.
	SeatIndex() boardgame.PlayerIndex
	//The callback that should be called when the move is committed
	Committed()
}
