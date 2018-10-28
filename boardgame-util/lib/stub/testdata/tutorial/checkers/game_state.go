package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type gameState struct {
	base.SubState
	//Use RoundRobinGameStateProperties so roundrobin moves can be used without any changes
	moves.RoundRobinGameStateProperties
	//DefaultGameDelegate will automatically return this from CurrentPlayerIndex
	CurrentPlayer boardgame.PlayerIndex
	//DefaultGameDelegate will automatically return this from PhaseEnum, CurrentPhase.
	Phase     enum.Val        `enum:"Phase"`
	DrawStack boardgame.Stack `stack:"examplecards" sanitize:"len"`
	//This is where the example config is stored in BeginSetup. We use it in
	//gameState.CardsDone().
	TargetCardsLeft int
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

func (g *gameState) CardsDone() bool {
	//It's common to hang computed properties and methods off of gameState and
	//playerState to use in logic elsewhere.

	return g.DrawStack.Len() == g.TargetCardsLeft
}
