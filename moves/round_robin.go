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
type playerConditionMet interface {
	//PlayerConditionMet should return whether the condition for the round
	//robin to be over has been met for this player.
	PlayerConditionMet(playerState boardgame.PlayerState) bool
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

Round Robin moves start at a given player and goes around. It will skip
players for whom move.PlayerConditionMet() has already returned
true. When it finds a player whose end condition is not met, it will apply
RoundRobinAction() to them, and then advance to the next player. Every time it
makes a circuit around the list of players, it will increment
RoundRobinRoundCount. By default, once all players have had their player
conditions met (that is, no player is legal to select) the round robin's will
be done applying: its ConditionMet will return nil.

Round Robin keeps track of various properties on the gameState by using the
RoundRobinProperties interface. Generally it's easiest to simply embed
moveinterfaces.RoundRobinBaseGameState in your GameState anonymously to
implement the interface automatically.

The embeding move should implement moveinterfaces.RoundRobinActioner.

*/
type RoundRobin struct {
	ApplyUntil
}

//RoundRobinStarterPlayer by default will return delegate.CurrentPlayer.
//Override this method if you want a different starter.
func (s *RoundRobin) RoundRobinStarterPlayer(state boardgame.State) boardgame.PlayerIndex {
	return state.Game().Manager().Delegate().CurrentPlayerIndex(state)
}

//ConditionMet  goes around and returns nil if all players have had their
//player condition met, meaning that there are no more legal players to
//select. Because this condition is almost always an important base no matter
//the other conditions you are considering (it's not possible to select
//players who have already had their player condition met), if you override
//CondtionMet you should also call this implementation.
func (r *RoundRobin) ConditionMet(state boardgame.State) error {
	conditionsMet, ok := r.TopLevelStruct().(playerConditionMet)

	if !ok {
		//This should be extremely rare since we ourselves have the right method.
		return errors.New("RoundRobin top level struct unexpectedly did not have PlayerConditionMet method")
	}

	for i, player := range state.PlayerStates() {
		if !conditionsMet.PlayerConditionMet(player) {
			return errors.New("Player " + strconv.Itoa(i) + " does not have their player condition met.")
		}
	}

	return nil
}

//PlayerConditionMet is called for each playerState. When advancing
//to the next player, round robin will only pick a player whose condition has
//not yet been met. Once all players have their PlayerconditionMet, then the
//RoundRobin's ConditionMet will return nil, signaling that the RoundRobin is
//done. By default this will return false. If you will use RoundRobin directly
//(as opposed to RoundRobinNumRounds) you will want to override this otherwise
//it will get in an infinite loop.
func (r *RoundRobin) PlayerConditionMet(playerState boardgame.PlayerState) bool {
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

	conditionsMet, ok := r.TopLevelStruct().(playerConditionMet)

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

		if !conditionsMet.PlayerConditionMet(state.PlayerStates()[currentPlayer]) {
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

//Legal is a complex implementation because it needs to figure out when to
//start the round robin. In general you do not want to override this.
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

//Apply is a complex implementation because it needs to figure out when the
//round is already over and handle complicated signalling about who the next
//player is. In general you do not want to override this.
func (r *RoundRobin) Apply(state boardgame.MutableState) error {

	if !r.roundRobinHasStarted(state) {
		if err := r.startRoundRobin(state); err != nil {
			return errors.New("Couldn't start round robin: " + err.Error())
		}
	}

	conditionMetter, ok := r.TopLevelStruct().(moveinterfaces.ConditionMetter)

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

type numRoundser interface {
	NumRounds() int
}

//RoundRobinNumRounds is a subclass of RoundRobin whose ConditionMet checks
//whether RoundRobinRoundCount is greater than or equal to NumRounds(), and if
//it is ends immediately. NumRounds() defaults to 1; if you want to have
//multiple rounds, override NumRounds().
type RoundRobinNumRounds struct {
	RoundRobin
}

func (r *RoundRobinNumRounds) ValidConfiguration(exampleState boardgame.MutableState) error {
	if err := r.RoundRobin.ValidConfiguration(exampleState); err != nil {
		return err
	}
	if _, ok := r.TopLevelStruct().(numRoundser); !ok {
		return errors.New("EmbeddingMove unexpectedly did not implement NumRounds!")
	}
	return nil
}

//NumRounds should return the RoundRobinRoundCount that we are targeting. As
//soon as that RoundCount is reached, our ConditionMet will start returning
//nil, signaling the Round Robin is over. Defaults to 1.
func (r *RoundRobinNumRounds) NumRounds() int {
	return 1
}

//ConditionMet will check if the round count has been reached; if it has it
//will return nil immediately. Otherwise it will fall back on RoundRobin's
//base ConditionMet, returning nil if no players are left to act upon.
func (r *RoundRobinNumRounds) ConditionMet(state boardgame.State) error {
	numRounds, ok := r.TopLevelStruct().(numRoundser)

	if !ok {
		//Unexpected!
		return nil
	}

	roundRobiner, ok := state.GameState().(moveinterfaces.RoundRobinProperties)

	if !ok {
		return nil
	}

	if roundRobiner.RoundRobinRoundCount() >= numRounds.NumRounds() {
		return nil
	}

	return r.RoundRobin.ConditionMet(state)

}
