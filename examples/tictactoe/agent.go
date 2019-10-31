package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

//Agent is an agent capable of playing tictactoe.
type Agent struct{}

//Name returns "ai"
func (a *Agent) Name() string {
	return "ai"
}

//DisplayName returns "Robby The Robot"
func (a *Agent) DisplayName() string {
	return "Robby The Robot"
}

//SetUpForGame is not yet implemented.
func (a *Agent) SetUpForGame(game *boardgame.Game, player boardgame.PlayerIndex) (agentState []byte) {
	return nil
}

//ProposeMove is not yet implemented.
func (a *Agent) ProposeMove(game *boardgame.Game, player boardgame.PlayerIndex, agentState []byte) (move boardgame.Move, newState []byte) {
	return nil, nil
}
