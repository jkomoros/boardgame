/*
Package interfaces is a collection of interfaces that your objects can implement
to configure how the moves package's base moves operate.

Factored into a separate package primarily just to make the moves
package more clear about which structs are the main embeddable moves.
*/
package interfaces

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// AllowMultipleInProgression is an interface that moves should implement if they
// want to affirmatively communicate to [moves.Default] that in a move progression
// is it legal to apply multiple. If the move does not implement this interface
// then it is considered to only allow one.
type AllowMultipleInProgression interface {
	//AllowMultipleInProgression should return true if the given move is
	//allowed to apply multiple times in order in a move progression.
	AllowMultipleInProgression() bool
}

// LegalComponent should be implemented by ComponentValues that will be used
// with [moves.DefaultComponent].
type LegalComponent interface {
	//Legal will be called on each component, with a legalType related to the
	//move in question (or 0 if WithLegalType hasn't been called). This allows
	//the same component values to participate in multiple
	//moves.DefaultComponent move types. Idiomatically legalType should be a
	//value in an enum created for the purpose of disambiguating different
	//move types to check for legality for. Legal should return nil if it is
	//legal, or an error if the component is not legal.
	Legal(state boardgame.ImmutableState, legalType enum.ImmutableVal) error
}

// PlayerStacker should be implemented by your embedding Move if you embed
// [moves.DealCountComponents]. It will be consulted to figure out where the PlayerStack is
// to deal a component to.
type PlayerStacker interface {
	PlayerStack(playerState boardgame.SubState) boardgame.Stack
}

// GameStacker should be implemented by your emedding Move if you embed
// [moves.DealCountComponents]. It will be consulted to figure out where to draw the
// components from to deal to players.
type GameStacker interface {
	GameStack(gameState boardgame.SubState) boardgame.Stack
}

// SourceStacker should be implemented by moves that need an input stack to
// operate on as primary/source, for example [moves.ShuffleStack].
type SourceStacker interface {
	SourceStack(state boardgame.State) boardgame.Stack
}

// DestinationStacker should be implemented by moves that need a destination
// stack to operate on as primary/source, for example [moves.ApplyUntilCount].
type DestinationStacker interface {
	DestinationStack(state boardgame.State) boardgame.Stack
}

// CurrentPlayerSetter should be implemented by gameStates that use [moves.FinishTurn].
// [behaviors.CurrentPlayerBehavior] satifies this.
type CurrentPlayerSetter interface {
	SetCurrentPlayer(currentPlayer boardgame.PlayerIndex)
}

// TargetCounter should be implemented by moves who should be legal until a
// TargetCount has been reached.
type TargetCounter interface {
	TargetCount(state boardgame.ImmutableState) int
}

// PlayerTurnFinisher is the interface your playerState is expected to adhere
// to when you use [moves.FinishTurn].
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

// RoundRobinProperties should be implemented by your GameState if you use any of
// the [moves.RoundRobin] moves, including StartRoundRobin. You don't have to do anything
// we these other than store them to a property in your gameState and then return
// them via the getters. Generally you simply embed [behaviors.RoundRobin] to
// satisfy this interface for free.
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

// ConditionMetter should be implemented by moves that subclass
// [moves.ApplyUntil].
type ConditionMetter interface {
	//ConditionMet should return nil if the condition has been met, or an
	//error describing why the condition has not yet been met.
	ConditionMet(state boardgame.ImmutableState) error
}

// RoundRobinActioner should be implemented by any moves that embed a
// [moves.RoundRobin] move. It's the action that will be called on the player who is
// next in the round robin.
type RoundRobinActioner interface {
	//RoundRobinAction should do the action for the round robin to given player.
	RoundRobinAction(playerState boardgame.SubState) error
}

// CurrentPhaseSetter should be implemented by you gameState to set the
// CurrentPhase. Must be implemented if you use the [moves.StartPhase] move type.
// [behaviors.PhaseBehavior] implements this.
type CurrentPhaseSetter interface {
	SetCurrentPhase(enum.EnumKey)
}

// BeforeLeavePhaser is an interface to implement on GameState if you want to
// do some action on state before leaving the given phase.
type BeforeLeavePhaser interface {
	BeforeLeavePhase(phase enum.ImmutableVal, state boardgame.State) error
}

// BeforeEnterPhaser is an interface to implement on GameState if you want to
// do some action on state just before entering the givenn state.
type BeforeEnterPhaser interface {
	BeforeEnterPhase(phase enum.ImmutableVal, state boardgame.State) error
}

// PlayerSubmitter is for PlayerStates that track whether a player has
// submitted a selection during a simultaneous play phase. Used by
// [moves.AllPlayersSubmitted] and [moves.ResetAllPlayerSubmissions].
// [behaviors.PlayerSubmission] satisfies this interface.
type PlayerSubmitter interface {
	HasSubmitted() bool
	SetPlayerSubmitted()
	ResetSubmission()
}

// PlayerInactiver is for PlayerStates that encode whether that player is
// Inactive and whether that might be changed. See the package doc of
// [behaviors] for more on the notion of inactive players.
type PlayerInactiver interface {
	IsInactive() bool
	SetPlayerInactive()
	SetPlayerActive()
}

// Seater is for PlayerStates that interace with moves like [moves.SeatPlayer]. See the
// package doc of [behaviors] for more on the notion of seats.
type Seater interface {
	SeatIsFilled() bool
	SeatIsClosed() bool
	SetSeatFilled()
	SetSeatClosed()
}

// SeatPlayerMover should be implemented for moves that are [moves.SeatPlayer] moves,
// returing true from IsSeatPlayerMove(). Typically you use [moves.SeatPlayer]
// directly, which implements this interface, but you might also want to embed
// that move in another move to, for example, change the DefaultsForState logic.
// This is the way that even those embedded moves can be detected as being
// intended as SeatPlayer moves.
type SeatPlayerMover interface {
	IsSeatPlayerMove() bool
}

// SeatPlayerSignaler is the way that [moves.SeatPlayer] and the server coordinate
// about where to seat a player.
type SeatPlayerSignaler interface {
	//The index of the seat to sit the palyer in. If it's not a valid index
	//(e.g. AdminPlayerIndex), will use the next available seat.
	SeatIndex() boardgame.PlayerIndex
	//The callback that should be called when the move is committed
	Committed()
}

// SpaceValidator is optionally implemented by moves that embed [moves.MoveOnGraph].
// Called during [moves.MoveOnGraph.Legal] for each space in the computed path. The
// playerState has access to the full state via ImmutableState().
type SpaceValidator interface {
	SpaceIsLegal(playerState boardgame.ImmutableSubState, spaceIndex enum.ImmutableVal) error
}

// MovementBudgeter is optionally implemented by moves that embed [moves.MoveOnGraph].
// Controls movement budget checking (Legal) and decrementing (Apply).
type MovementBudgeter interface {
	MovesRemaining(playerState boardgame.ImmutableSubState) int
	ConsumeMovement(playerState boardgame.SubState, pathLength int) error
}

// FreeMovePredicate is optionally implemented by moves that embed [moves.MoveOnGraph].
// If a target satisfies IsFreeMove, the framework skips budget and adjacency
// checks and moves the token directly (teleport).
type FreeMovePredicate interface {
	IsFreeMove(playerState boardgame.ImmutableSubState, targetSpaceIndex enum.ImmutableVal) bool
}

// PlayerOrderer is implemented by gameStates that want to define a custom
// player order. When implemented, [boardgame.PlayerIndex.Next] and
// [boardgame.PlayerIndex.Previous] will follow the custom order instead of
// sequential order. [behaviors.PlayerOrderBehavior] satisfies this interface.
type PlayerOrderer interface {
	PlayerOrder() []boardgame.PlayerIndex
}

// FreeMoveApplier is optionally implemented by moves that embed [moves.MoveOnGraph].
// When a free move is made, ApplyFreeMove is called to handle game-specific
// cleanup (e.g., resetting a card-based MoveToRoom value).
type FreeMoveApplier interface {
	ApplyFreeMove(playerState boardgame.SubState, targetSpaceIndex enum.ImmutableVal) error
}

// PlayerEliminator is for PlayerStates that track whether a player has been
// eliminated (knocked out) from play. Used by game logic to check elimination
// status. [behaviors.PlayerElimination] satisfies this interface.
type PlayerEliminator interface {
	IsEliminated() bool
	SetEliminated()
	ClearEliminated()
}

// TurnBudgeter is for PlayerStates that track how many actions a player has
// remaining in their current turn. [behaviors.MoveBudget] satisfies this
// interface.
type TurnBudgeter interface {
	HasMovesLeft() bool
	ConsumeMove()
	ResetMovesTo(n int)
}

// TeamMember is for PlayerStates that track team membership via an enum.
// [behaviors.PlayerTeam] satisfies this interface.
type TeamMember interface {
	IsOnTeam(enum.EnumKey) bool
}

// AdvanceCondition is optionally implemented by moves that embed [moves.AdvanceToken].
// It gates whether the advancement should happen.
type AdvanceCondition interface {
	ShouldAdvance(state boardgame.ImmutableState) error
}

// PostAdvanceHandler is optionally implemented by moves that embed
// [moves.AdvanceToken]. It runs game-specific side effects after the token has been
// advanced.
type PostAdvanceHandler interface {
	AfterAdvance(state boardgame.State, previousIndex, newIndex enum.ImmutableVal) error
}
