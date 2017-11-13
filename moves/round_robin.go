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
type roundRobinPlayerConditionMet interface {
	//RoundRobinPlayerConditionMet should return whether the condition for the
	//round robin to be over has been met for this player.
	RoundRobinPlayerConditionMet(playerState boardgame.PlayerState) bool
}

/*

RoundRobin is a complicated type of move because a lot of complicated logic
has to be included in its Legal() and Apply(). It doesn't use
TargetPlayerIndex, because DefaultsForState doesn't get a chance to modify the
underlying state, and figuring out the next index to apply on is a non-trivial
calculation.

In addition, we don't want an explicit StartRoundRobin move, which means that
RoundRobin itself is responsible for signaling that it is active and when it
is done.

The first round robin sees that HasStartedRoundRobin is false, and is
therefore legal. (It also verifies that the last round robin is not equal to
its move name, to avoid situations where the round robin is over and it just
starts back up again).

RoundRobinLastPlayer is alway set to the last round robin
move that was applied.

The first Apply() sets up the RoundRobin by configuring all of the various
properties and signaling thta RoundRobinHasStarted is true.

RoundRobin.Legal() and RoundRobin.Apply() both have to keep track of the next
player to apply to. It is the LastPlayer, played forward until we find a
player whose ConditionMet returns nil.

When we apply it, we apply the action and update LastPlayer to this player
index. If searching forward fro the LastPlayer to our index either ended on
the original player or passed over them, then RoundCount is incremented.

After applying, we check the legal logic again. If we're the last move in the
round robin we finalize it by setting the properties to signal the round robin
is over.

*/

/*

RoundRobin is a type of move that goes around every player one by one and does
some action. Other moves in this package embed RoundRobin, and it's more
common to use those directly.

Round Robin moves start at a given player and go around, applying the
RoundRobinAction for each player until the ConditionMet() method returns
nil. Various embedders of the base RoundRobin will override the default
behavior for that method.

Round Robin keeps track of various properties on the gameState by using the
RoundRobinProperties interface. Generally it's easiest to simply embed
moveinterfaces.RoundRobinBaseGameState in your GameState anonymously to
implement the interface automatically.

The embeding move should implement RoundRobinActioner.

*/
type RoundRobin struct {
	ApplyUntil
}

//RoundRobinStarterPlayer by default will return delegate.CurrentPlayer.
//Override this method if you want a different starter.
func (s *RoundRobin) RoundRobinStarterPlayer(state boardgame.State) boardgame.PlayerIndex {
	return state.Game().Manager().Delegate().CurrentPlayerIndex(state)
}

//ConditionMet should return nil when the move can stop being applied. Vy
//default it returns the result of RoundRobinFinishedOneCircuit(). If you want
//other behavior override this method.
func (r *RoundRobin) ConditionMet(state boardgame.State) error {
	return r.RoundRobinFinishedOneCircuit(state)
}

//RoundRobinFinshedOneCircuit returns an error if the RoundRobinRountCount is
//less than 1, meaning it begins returning nil as soon as one full circuit is
//completed. The result of this is a single circuit. It is designed to be
//called directly in your ConditionMet.
func (r *RoundRobin) RoundRobinFinishedOneCircuit(state boardgame.State) error {
	return r.RoundRobinFinishedMultiCircuit(1, state)
}

//RoundRobinFinshedOneCircuit returns an error if the RoundRobinRountCount is
//less than targetCount, meaning it begins returning nil as soon as that many
//full circuits are completed. It is designed to be called directly in your
//ConditionMet.
func (r *RoundRobin) RoundRobinFinishedMultiCircuit(targetCount int, state boardgame.State) error {
	props, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState unexpectedly did not implement RoundRobinProperties")
	}

	if props.RoundRobinRoundCount() < targetCount {
		return errors.New("The round count is " + strconv.Itoa(props.RoundRobinRoundCount()) + " which does not meet the target of " + strconv.Itoa(targetCount))
	}

	return nil
}

//RoundRobinFinishedPlayerConditionsMet returns nil if calling
//RoundRobinPlayerConditionMet on this move, passing each playerState in turn,
//all return true. It's useful, as an example, for going around and making
//sure that every player has at least two cards in their hand, if players may
//have started the round robin with a different number of cards in hand. It is
//designed to be called directly in your ConditionMet.
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

//roundRobinHasStarted returns true if the gameState RoundRobin properties are not their sentinel values.
func (r *RoundRobin) roundRobinHasStarted(state boardgame.State) bool {
	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return false
	}

	return roundRobiner.RoundRobinHasStarted()
}

//startRoundRobin should be called if roundRobinHasStarted is false.
func (r *RoundRobin) startRoundRobin(state boardgame.MutableState) error {

	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState unexpectedly did not implement RoundRobiner interface")
	}

	starter, ok := r.TopLevelStruct().(roundRobinStarterPlayer)

	if !ok {
		//This should be extremely rare, because if we're embedded in it then
		//the struct should have it.
		return errors.New("The top level struct unexpectedly didn't have RoundRobinStarterPlayer")
	}

	starterPlayer := starter.RoundRobinStarterPlayer(state)

	roundRobiner.SetRoundRobinLastPlayer(starterPlayer.Previous(state))
	roundRobiner.SetRoundRobinStarterPlayer(starterPlayer)
	roundRobiner.SetRoundRobinRoundCount(0)
	roundRobiner.SetRoundRobinHasStarted(true)

	return nil
}

//finishRoundRobin should be called in the Apply method of the last round robin move.
func (r *RoundRobin) finishRoundRobin(state boardgame.MutableState) error {
	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState unexpectedly did not implement RoundRobiner interface")
	}

	roundRobiner.SetRoundRobinHasStarted(false)

	return nil
}

//nextPlayerIndex returns the next playerIndex that the round robin will
//operate on.  Also returns roundSkip true if the player we end on is the
//first player in the round robin, or if we skipped over them to find the next
//valid player.
func (r *RoundRobin) nextPlayerIndex(state boardgame.State) (player boardgame.PlayerIndex, roundSkip bool) {

	var currentPlayer boardgame.PlayerIndex

	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return boardgame.ObserverPlayerIndex, true
	}

	if r.roundRobinHasStarted(state) {

		currentPlayer = roundRobiner.RoundRobinLastPlayer()
	} else {

		starterPlayer, ok := r.TopLevelStruct().(roundRobinStarterPlayer)

		if !ok {
			return boardgame.ObserverPlayerIndex, true
		}

		currentPlayer = starterPlayer.RoundRobinStarterPlayer(state).Previous(state)
	}

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
		return boardgame.ObserverPlayerIndex, true
	}

	counter := 0
	roundSkip = false

	//Advance around, but if we loop back just leave it.
	for counter <= len(state.PlayerStates()) {

		currentPlayer = currentPlayer.Next(state)

		if currentPlayer == roundRobiner.RoundRobinStarterPlayer() {
			roundSkip = true
		}

		if !conditionsMet.RoundRobinPlayerConditionMet(state.PlayerStates()[currentPlayer]) {
			break
		}

		counter++
	}

	if counter > len(state.PlayerStates()) {
		//No players are legal
		return currentPlayer, true
	}

	return currentPlayer, roundSkip

}

func (r *RoundRobin) ValidConfiguration(exampleState boardgame.MutableState) error {

	if err := r.ApplyUntil.ValidConfiguration(exampleState); err != nil {
		return err
	}

	if _, ok := exampleState.GameState().(moveinterfaces.RoundRobinProperties); !ok {
		return errors.New("GameState does not implement RoundRobiner interface")
	}

	if _, ok := r.TopLevelStruct().(roundRobinStarterPlayer); !ok {
		return errors.New("Embedding move doesn't implement RoundRobinStarterPlayer")
	}

	embeddingMove := r.TopLevelStruct()

	if _, ok := embeddingMove.(moveinterfaces.RoundRobinActioner); !ok {
		return errors.New("Embedding move doesn't implement RoundRobinActioner")
	}

	return nil
}

//lastMoveName returns the name of the last move that was applied to this game.
func (r *RoundRobin) lastMoveName(state boardgame.State) string {
	moves := state.Game().MoveRecords(state.Version())

	if len(moves) == 0 {
		return ""
	}

	return moves[len(moves)-1].Name
}

func (r *RoundRobin) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	//We run the base legal first to see if this phase is even legal for us.
	//We can't run ApplyUntil until later, because it will say it's not legal
	//if the round robin hasn't started.
	if err := r.Base.Legal(state, proposer); err != nil {
		return err
	}

	if !r.roundRobinHasStarted(state) {

		//If the round robin hasn't started, it's legal to start--as long as
		//the last move applied was not us (otherwise we'd just infinite loop
		//in them).

		if r.TopLevelStruct().Info().Type().Name() == r.lastMoveName(state) {
			return errors.New("Can't start this round robin move because the last move was also part of this round robin.")
		}

		return nil
	}

	//We do ApplyUntil down here because it will check ConditionMet, which
	//will be true before the game starts in many cases.
	if err := r.ApplyUntil.Legal(state, proposer); err != nil {
		return err
	}

	//If the round robin has started, then it's legal--as soon as it isn't
	//legal, we've turned off the round robing.

	return nil
}

func (r *RoundRobin) Apply(state boardgame.MutableState) error {

	if !r.roundRobinHasStarted(state) {
		if err := r.startRoundRobin(state); err != nil {
			return errors.New("Couldn't start round robin: " + err.Error())
		}
	}

	conditionMetter, ok := r.TopLevelStruct().(conditionMetter)

	if !ok {
		return errors.New("Top level struct unexpectedly did not implement condition met")
	}

	if conditionMetter.ConditionMet(state) == nil {
		return errors.New("The round robin was found to be finished in our Apply, but it should have been marked finished before!")
	}

	nextPlayer, _ := r.nextPlayerIndex(state)

	actioner, ok := r.TopLevelStruct().(moveinterfaces.RoundRobinActioner)

	if !ok {
		return errors.New("Embedding move doesn't implement RoundRobinActioner")
	}

	playerToAction := state.MutablePlayerStates()[nextPlayer]

	if err := actioner.RoundRobinAction(playerToAction); err != nil {
		return errors.New("RoundRobinAction returned error: " + err.Error())
	}

	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return errors.New("GameState does not implement RoundRobiner interface")
	}

	roundRobiner.SetRoundRobinLastPlayer(nextPlayer)

	nextPlayer, roundSkip := r.nextPlayerIndex(state)

	if roundSkip {
		roundRobiner.SetRoundRobinRoundCount(roundRobiner.RoundRobinRoundCount() + 1)
	}

	if conditionMetter.ConditionMet(state) == nil {
		if err := r.finishRoundRobin(state); err != nil {
			return errors.New("Couldn't finish round robin when it was done: " + err.Error())
		}
	}

	return nil

}
