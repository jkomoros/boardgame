package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
)

//boardgame:codegen
type gameState struct {
	base.SubState
	//Use behaviors.RoundRobin so roundrobin moves can be used without any changes
	behaviors.RoundRobin
	//base.GameDelegate will automatically return this from CurrentPlayerIndex
	CurrentPlayer boardgame.PlayerIndex
	//base.GameDelegate will automatically return this from PhaseEnum, CurrentPhase.
	Phase enum.Val `enum:"phase"`
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (g *gameState) SetCurrentPhase(phase int) {
	g.Phase.SetValue(phase)
}

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	//Having this setter allows us to work moves.With moves.TurnDone
	g.CurrentPlayer = currentPlayer
}
