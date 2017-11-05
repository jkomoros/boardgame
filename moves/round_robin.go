package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//RoundRobiner should be implemented by your GameState if you use any of the
//RoundRobin moves, including StartRoundRobin.
type RoundRobiner interface {
	NextRoundRobinPlayer() boardgame.PlayerIndex
	SetNextRoundRobinPlayer(nextPlayer boardgame.PlayerIndex)
}

//StartRoundRobin is the move you should have in the progression immediately
//before a round robin starts. It sets the NextRoundRobinPlayer to the game's
//CurrentPlayer, getting ready for moves that embed RoundRobin.
type StartRoundRobin struct {
	Base
}

func (s *StartRoundRobin) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := exampleState.GameState().(RoundRobiner); !ok {
		return errors.New("GameState does not implement RoundRobiner interface")
	}
	return nil
}

//Apply gets the game ready for a round robin by calling
//gameState.SetNextRoundRobinPlayer to CurrentPlayerIndex.
func (s *StartRoundRobin) Apply(state boardgame.MutableState) error {
	roundRobiner, ok := state.GameState().(RoundRobiner)

	if !ok {
		return errors.New("GameState unexpectedly did not implement RoundRobiner interface")
	}

	currentPlayer := state.Game().Manager().Delegate().CurrentPlayerIndex(state)

	roundRobiner.SetNextRoundRobinPlayer(currentPlayer)

	return nil
}
