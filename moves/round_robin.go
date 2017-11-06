package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
	"strconv"
)

//We can keep these private because embedders already will have the interface
//satisfied so don't need to be confused by them.
type roundRobinStarterPlayer interface {
	RoundRobinStarterPlayer(state boardgame.State) boardgame.PlayerIndex
}
type roundRobinFinished interface {
	RoundRobinFinished(state boardgame.State) error
}
type roundRobinPlayerConditionMet interface {
	//RoundRobinPlayerConditionMet should return whether the condition for the
	//round robin to be over has been met for this player.
	RoundRobinPlayerConditionMet(playerState boardgame.PlayerState) bool
}

//PhaseToEnterer should be implemented by moves that embed moves.StartPhase to
//configure which phase to enter.
type PhaseToEnterer interface {
	PhaseToEnter(currentPhase int) int
}

//StartRoundRobin is the move you should have in the progression immediately
//before a round robin starts. It sets the NextRoundRobinPlayer to the game's
//CurrentPlayer, getting ready for moves that embed RoundRobin. Because the
//move doesn't change very much, NewStartRoundRobinMoveConfig can be used so
//you don't even need to define an embedding move for it in your own package.
//
//+autoreader
type StartRoundRobin struct {
	Base
}

func (s *StartRoundRobin) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := exampleState.GameState().(moveinterfaces.RoundRobinProperties); !ok {
		return errors.New("GameState does not implement RoundRobiner interface")
	}
	if _, ok := s.TopLevelStruct().(roundRobinStarterPlayer); !ok {
		return errors.New("Embedding Move doesn't have RoundRobinStarterPlayer")
	}
	return nil
}

//RoundRobinStarterPlayer by default will return delegate.CurrentPlayer.
//Override this method if you want a differnt starter.
func (s *StartRoundRobin) RoundRobinStarterPlayer(state boardgame.State) boardgame.PlayerIndex {
	return state.Game().Manager().Delegate().CurrentPlayerIndex(state)
}

//Apply gets the game ready by setting the various starter properties on
//GameState. It sets the starting player for the round robin to the result of
//calling RoundRobinStarterPlayer on the move.
func (s *StartRoundRobin) Apply(state boardgame.MutableState) error {
	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState unexpectedly did not implement RoundRobiner interface")
	}

	starter, ok := s.TopLevelStruct().(roundRobinStarterPlayer)

	if !ok {
		//This should be extremely rare, because if we're embedded in it then
		//the struct should have it.
		return errors.New("The top level struct unexpectedly didn't have RoundRobinStarterPlayer")
	}

	starterPlayer := starter.RoundRobinStarterPlayer(state)

	roundRobiner.SetNextRoundRobinPlayer(starterPlayer)
	roundRobiner.SetRoundRobinStarterPlayer(starterPlayer)
	roundRobiner.SetRoundRobinRoundCount(0)

	return nil
}

//NewStartRoundRobinMoveConfig returns a move config that you can use to
//install a StartRoundRobin into your GameManager without needing to define a
//struct in your own package that embeds StartRoundRobin anonymously.
func NewStartRoundRobinMoveConfig(legalPhases []int) *boardgame.MoveTypeConfig {

	return &boardgame.MoveTypeConfig{
		Name:     "Start Round Robin",
		HelpText: "Gets ready for a round robin",
		MoveConstructor: func() boardgame.Move {
			return new(StartRoundRobin)
		},
		IsFixUp:     true,
		LegalPhases: legalPhases,
	}
}

/*

RoundRobin is a type of move that goes around every player one by one and does
some action. Other moves in this package embed RoundRobin, and it's more
common to use those directly.

Round Robin moves start at a given player and go around, applying the
RoundRobinAction for each player until the RoundRobinFinished() method returns
true. Various embedders of the base RoundRobin will override the default
behavior for that method.

Round Robin keeps track of various properties on the gameState by using the
RoundRobinProperties interface. Generally it's easiest to simply embed
moveinterfaces.RoundRobinBaseGameState in your GameState anonymously to
implement the interface automatically.

The embeding move should implement RoundRobinActioner.

A round robin phase must be immediately preceded by StartRoundRobin, which
sets various properties the round robin needs to operate before it starts.

*/
type RoundRobin struct {
	Base
	TargetPlayerIndex boardgame.PlayerIndex
}

//AllowMultipleInProgression returns true because RoundRobins go until the end
//condition of the round robin is met.
func (r *RoundRobin) AllowMultipleInProgression() bool {
	return true
}

//DefaultsForState sets the TargetPlayerIndex to NextRoundRobinPlayer, unless
//that player already has their PlayerConditionMet, in which case it advances
//until it finds a TargetPlayerIndex where the condition is not yet met.
func (r *RoundRobin) DefaultsForState(state boardgame.State) {
	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return
	}

	targetPlayer := roundRobiner.NextRoundRobinPlayer()

	//If the PlayerConditionMet for that player is already true, we know that
	//we shouldn't land on them. Cycle around until we find one for which
	//PlayerConditionMet returns false, unless none of them are true, in which
	//case just leave it with the original target.

	//RoundRobin moves whose Finished() routine look for something other than
	//PlayerConditionMet will still work fine, because their
	//PlayerConditionMet will always return false.

	conditionsMet, ok := r.TopLevelStruct().(roundRobinPlayerConditionMet)

	if !ok {
		//This should be extremely rare since we ourselves have the right method.
		return
	}

	counter := 0

	//Advance around, but if we loop back just leave it.
	for counter <= len(state.PlayerStates()) {

		if !conditionsMet.RoundRobinPlayerConditionMet(state.PlayerStates()[targetPlayer]) {
			break
		}

		targetPlayer = targetPlayer.Next(state)
		counter++
	}

	r.TargetPlayerIndex = targetPlayer
}

//Legal returns whether the super's Legal returns an error, and will return an
//error if the RoundRobinFinished() method on this move returns true.
func (r *RoundRobin) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := r.Base.Legal(state, proposer); err != nil {
		return err
	}

	finisher, ok := r.TopLevelStruct().(roundRobinFinished)

	if !ok {
		//This should be extremely rare since we ourselves have the right method.
		return errors.New("RoundRobin top level struct unexpectedly did not have RoundRobinFinished method")
	}

	if err := finisher.RoundRobinFinished(state); err != nil {
		return errors.New("The round robin has met its finish condition: " + err.Error())
	}

	return nil

}

//RoundRobinFinished will be consulted by the Legal() method. If it returns an
//error then the round robin is considered finished. By default it returns the
//result of RoundRobinFinishedOneCircuit(). If you want other behavior
//override this method.
func (r *RoundRobin) RoundRobinFinished(state boardgame.State) error {
	return r.RoundRobinFinishedOneCircuit(state)
}

//RoundRobinFinshedOneCircuit returns an error if the RoundRobinRountCount is
//1 or higher, meaning as soon as one full circuit is completed. It is
//designed to be called directly in your RoundRobinFinished
func (r *RoundRobin) RoundRobinFinishedOneCircuit(state boardgame.State) error {
	return r.RoundRobinFinishedMultiCircuit(1, state)
}

//RoundRobinFinshedOneCircuit returns an error if the RoundRobinRountCount is
//targetCount or higher, meaning as soon as that many full circuits are
//completed. It is designed to be called directly in your RoundRobinFinished
func (r *RoundRobin) RoundRobinFinishedMultiCircuit(targetCount int, state boardgame.State) error {
	props, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState unexpectedly did not implement RoundRobinProperties")
	}

	if props.RoundRobinRoundCount() >= targetCount {
		return errors.New("The round count is " + strconv.Itoa(props.RoundRobinRoundCount()) + " which meets the target of " + strconv.Itoa(targetCount))
	}

	return nil
}

//RoundRobinFinishedPlayerConditionsMet returns an error if calling
//RoundRobinPlayerConditionMet on this move, passing each playerState in turn,
//all return true. It's useful, as an example, for going around and making
//sure that every player has at least two cards in their hand, if players may
//have started the round robin with a different number of cards in hand. It is
//designed to be called directly in your RoundRobinFinished.
func (r *RoundRobin) RoundRobinFinishedPlayerConditionsMet(state boardgame.State) error {

	conditionsMet, ok := r.TopLevelStruct().(roundRobinPlayerConditionMet)

	if !ok {
		//This should be extremely rare since we ourselves have the right method.
		return errors.New("RoundRobin top level struct unexpectedly did not have RoundRobinPlayerConditionMet method")
	}

	for i, player := range state.PlayerStates() {
		if !conditionsMet.RoundRobinPlayerConditionMet(player) {
			return errors.New("Player " + strconv.Itoa(i) + " does not have their player condition met.")
		}
	}

	return nil

}

//RoundRobinPlayerConditionStackTargetSizeMet is designed to be used as a one-
//liner within your RoundRobinPlayerConditionMet. It checks that the
//PlayerStack from the given PlayerState has the target size and returns true
//if so. Useful for cases where you want to have a DrawComponent that draws
//until each player has targetSize components, but players might start with
//different numbers of components already in hand. See DrawComponents
//documentation for an example.
func (r *RoundRobin) RoundRobinPlayerConditionStackTargetSizeMet(targetSize int, playerState boardgame.PlayerState) bool {
	playerStacker, ok := playerState.(moveinterfaces.PlayerStacker)

	if !ok {
		//Hmmm, guess it doesn't match PlayerStacker
		return false
	}

	//TODO: it's a total hack that we up-cast the playerState because we
	//happen to know it's definitely a MutablePlayerState.
	playerStack := playerStacker.PlayerStack(playerState.(boardgame.MutablePlayerState))

	return playerStack.NumComponents() == targetSize

}

//RoundRobinPlayerConditionMet is called for each playerState by
//RoundRobinFinishedPlayerConditionMet. If all of them return true, the round
//robin is over. The default simply returns false in all cases; you should override it.
func (r *RoundRobin) RoundRobinPlayerConditionMet(playerState boardgame.PlayerState) bool {
	return false
}

func (r *RoundRobin) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := exampleState.GameState().(moveinterfaces.RoundRobinProperties); !ok {
		return errors.New("GameState does not implement RoundRobiner interface")
	}

	embeddingMove := r.TopLevelStruct()

	if _, ok := embeddingMove.(moveinterfaces.RoundRobinActioner); !ok {
		return errors.New("Embedding move doesn't implement RoundRobinActioner")
	}

	if _, ok := embeddingMove.(roundRobinFinished); !ok {
		return errors.New("Embedding move doesn't implement RoundRobinFinished")
	}

	return nil
}

func (r *RoundRobin) Apply(state boardgame.MutableState) error {

	embeddingMove := r.TopLevelStruct()

	actioner, ok := embeddingMove.(moveinterfaces.RoundRobinActioner)

	if !ok {
		return errors.New("Embedding move doesn't implement RoundRobinActioner")
	}

	currentPlayer := state.MutablePlayerStates()[r.TargetPlayerIndex]

	if err := actioner.RoundRobinAction(currentPlayer); err != nil {
		return errors.New("RoundRobinAction returned error: " + err.Error())
	}

	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState does not implement RoundRobiner interface")
	}

	nextPlayer := r.TargetPlayerIndex.Next(state)

	roundRobiner.SetNextRoundRobinPlayer(nextPlayer)

	if nextPlayer == roundRobiner.RoundRobinStarterPlayer() {
		roundRobiner.SetRoundRobinRoundCount(roundRobiner.RoundRobinRoundCount() + 1)
	}

	return nil

}
