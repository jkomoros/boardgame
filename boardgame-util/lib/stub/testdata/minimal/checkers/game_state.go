package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type gameState struct {
	boardgame.BaseSubState
	//Use RoundRobinGameStateProperties so roundrobin moves can be used without any changes
	moves.RoundRobinGameStateProperties
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}
