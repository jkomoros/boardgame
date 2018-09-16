/*

interfaces is a collection of interfaces that your objects can implement
to configure how the moves package's base moves operate.

Factored into a separate package primarily just to make the moves package more
clear about which structs are the main embeddable moves.

*/
package interfaces

import (
	"github.com/jkomoros/boardgame"
)

//MoveGroupHistoryItem is a singly-linked list (referred to in various
//comments as a "tape") that is passed to MoveProgressionGroup.Satisfied(). It
//represents a list of all of the moves that have applied so far since
//game.CurrentPhase() last changed.
type MoveGroupHistoryItem struct {
	MoveName string
	Rest     *MoveGroupHistoryItem
}

//MoveProgressionGroup is an object that can be used to define a valid move
//progression. moves.AutoConfigurer().Config() returns objects that fit this
//interface.
type MoveProgressionGroup interface {
	//MoveConfigs should return the full enumeration of contained MoveConfigs
	//within this Group, from left to right and top to bottom. This is used by
	//moves.AddOrderedForPhase to know which MoveConfigs contained within it
	//to install.
	MoveConfigs() []boardgame.MoveConfig

	//Satisfied reads the tape and returns an error if the sequence was not
	//valid (did not match, for example, or the group was configured in an
	//invalid way in general). If it returns a nil error, it should also
	//return the rest of the tape representing the items it did not yet
	//consume. If passed a nil tape, it should immediately return nil, nil. If
	//the top-level MoveProgressionGroup consumes the entire tape and doesn't
	//return an error then the progression is considered valid.
	Satisfied(tape *MoveGroupHistoryItem) (error, *MoveGroupHistoryItem)
}

//ValidCounter is the signature of objects in the moves/count package. It is
//expected within groups in the move/groups package for items like
//ParallelCount. currentCount is the value of the counter in question, and
//length is the context-specific length of the important item, often the
//number of children in the parrent group. If ValidCounter returns nil, the
//count is considered valid and complete; if it is not valid it should return
//a descriptive error. Typically these functions are closures that close over
//configuration options.
type ValidCounter func(currentCount, length int) error

//Moves should implement AllowMultipleInProgression if they want to
//affirmatively communicate to moves.Base that in a move progression is it
//legal to apply multiple. If the move does not implement this interface then
//it is considered to only allow one.
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
	PlayerStack(playerState boardgame.PlayerState) boardgame.Stack
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

//SourceStacker should be implemented by moves that need a destination stack
//to operate on as primary/source, for example ApplyUntilCount.
type DestinationStacker interface {
	DestinationStack(state boardgame.State) boardgame.Stack
}

//CurrentPlayerSetter should be implemented by gameStates that use FinishTurn.
type CurrentPlayerSetter interface {
	SetCurrentPlayer(currentPlayer boardgame.PlayerIndex)
}

//TargetCounter should be implemented by moves who should be legal until a
//TargetCount has been reached.
type TargetCounter interface {
	TargetCount() int
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

//RoundRobinProperties should be implemented by your GameState if you use any
//of the RoundRobin moves, including StartRoundRobin. You don't have to do
//anything we these other than store them to a property in your gameState and
//then return them via the getters. Generally you simply embed
//RoundRobinBaseGameState to satisfy this interface for free.
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
	RoundRobinAction(playerState boardgame.PlayerState) error
}

//CurrentPhaseSetter should be implemented by you gameState to set the
//CurrentPhase. Must be implemented if you use the StartPhase move type.
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
