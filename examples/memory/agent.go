package memory

import (
	"github.com/jkomoros/boardgame"
)

type Agent struct{}

func (a *Agent) Name() string {
	return "ai"
}

func (a *Agent) DisplayName() string {
	return "Robby the Robot"
}

func (a *Agent) SetUpForGame(game *boardgame.Game, player boardgame.PlayerIndex) []byte {
	//TODO: do something
	return nil
}

func (a *Agent) ProposeMove(game *boardgame.Game, player boardgame.PlayerIndex, agentState []byte) (move boardgame.Move, newState []byte) {
	//TODO: do something
	return nil, nil
}
