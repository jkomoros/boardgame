/*

moveinterfaces is a collection of interfaces that your objects can implement
to configure how the moves package's base moves operate.

Factored into a separate package primarily just to make the moves package more
clear about which structs are the main embeddable moves.

*/
package moveinterfaces

import (
	"github.com/jkomoros/boardgame"
)

//AutoConfigurableMove is the interface that moves passed to moves.AutoConfig
//must implement. These methods are interrogated to set the move name,
//helptext, and isFixUp to good values. moves.Base defines powerful stubs for
//these, so any moves that embed moves.Base (or embed a move that embeds
//moves.Base, etc) satisfy this interface.
type AutoConfigurableMove interface {
	//DefaultConfigMoves all must implement all Move methods.
	boardgame.Move
	//The name for the move type
	MoveTypeName() string
	//The HelpText to use.
	MoveTypeHelpText() string
	//Whether the move should be a fix up.
	MoveTypeIsFixUp() bool
}

//RoundRobinBaseGameState is designed to be embedded in your GameState
//anonymously to automatically satisfy the RoundRobinProperties interface,
//making it easy to use RoundRobin-basd moves. Because this embeds
//boardgame.BaseSubState itself, you should embed this INSTEAD of
//boardgame.BaseSubState.
type RoundRobinBaseGameState struct {
	boardgame.BaseSubState
	RRLastPlayer    boardgame.PlayerIndex
	RRStarterPlayer boardgame.PlayerIndex
	RRRoundCount    int
	RRHasStarted    bool
}

func (r *RoundRobinBaseGameState) RoundRobinLastPlayer() boardgame.PlayerIndex {
	return r.RRLastPlayer
}

func (r *RoundRobinBaseGameState) RoundRobinStarterPlayer() boardgame.PlayerIndex {
	return r.RRStarterPlayer
}

func (r *RoundRobinBaseGameState) RoundRobinRoundCount() int {
	return r.RRRoundCount
}

func (r *RoundRobinBaseGameState) RoundRobinHasStarted() bool {
	return r.RRHasStarted
}

func (r *RoundRobinBaseGameState) SetRoundRobinLastPlayer(nextPlayer boardgame.PlayerIndex) {
	r.RRLastPlayer = nextPlayer
}
func (r *RoundRobinBaseGameState) SetRoundRobinStarterPlayer(index boardgame.PlayerIndex) {
	r.RRStarterPlayer = index
}

func (r *RoundRobinBaseGameState) SetRoundRobinRoundCount(count int) {
	r.RRRoundCount = count
}

func (r *RoundRobinBaseGameState) SetRoundRobinHasStarted(val bool) {
	r.RRHasStarted = val
}

//Moves should implement AllowMultipleInProgression if they want to
//affirmatively communicate to moves.Base that in a move progression is it
//legal to apply multiple. If the move does not implement this interface then
//it is considered to only allow one.
type AllowMultipleInProgression interface {
	//AllowMultipleInProgression should return true if the given move is
	//allowed to apply multiple times in order in a move progression.
	AllowMultipleInProgression() bool
}

//PlayerStacker should be implemented by your embedding Move if you embed
//DealComponents. It will be consulted to figure out where the PlayerStack is
//to deal a component to.
type PlayerStacker interface {
	PlayerStack(playerState boardgame.MutablePlayerState) boardgame.MutableStack
}

//GameStacker should be implemented by your emedding Move if you embed
//DealComponents. It will be consulted to figure out where to draw the
//components from to deal to players.
type GameStacker interface {
	GameStack(gameState boardgame.MutableSubState) boardgame.MutableStack
}

//SourceStacker should be implemented by moves that need an input stack to
//operate on as primary/source, for example ShuffleStack.
type SourceStacker interface {
	SourceStack(state boardgame.MutableState) boardgame.MutableStack
}

//SourceStacker should be implemented by moves that need a destination stack
//to operate on as primary/source, for example ApplyUntilCount.
type DestinationStacker interface {
	DestinationStack(state boardgame.MutableState) boardgame.MutableStack
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
	ConditionMet(state boardgame.State) error
}

//RoundRobinActioner should be implemented by any moves that embed a
//RoundRobin move. It's the action that will be called on the player who is
//next in the round robin.
type RoundRobinActioner interface {
	//RoundRobinAction should do the action for the round robin to given player.
	RoundRobinAction(playerState boardgame.MutablePlayerState) error
}

//CurrentPhaseSetter should be implemented by you gameState to set the
//CurrentPhase. Must be implemented if you use the StartPhase move type.
type CurrentPhaseSetter interface {
	SetCurrentPhase(int)
}

//BeforeLeavePhaser is an interface to implement on GameState if you want to
//do some action on state before leaving the given phase.
type BeforeLeavePhaser interface {
	BeforeLeavePhase(phase int, state boardgame.MutableState) error
}

//BeforeEnterPhaser is an interface to implement on GameState if you want to
//do some action on state just before entering the givenn state.
type BeforeEnterPhaser interface {
	BeforeEnterPhase(phase int, state boardgame.MutableState) error
}
