package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/roundrobinhelpers"
)

//boardgame:codegen
type gameState struct {
	//Use roundrobinhelpers so roundrobin moves can be used without any changes
	roundrobinhelpers.BaseGameState
	//DefaultGameDelegate will automatically return this from CurrentPlayerIndex
	CurrentPlayer boardgame.PlayerIndex
	//DefaultGameDelegate will automatically return this from PhaseEnum, CurrentPhase.
	Phase    enum.Val        `enum:"Phase"`
	DrawDeck boardgame.Stack `stack:"examplecards"`
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

func (g *gameState) CardsDone() bool {
	//It's common to hang computed properties and methods off of gameState and
	//playerState to use in logic elsewhere.

	return g.DrawDeck.Len() == g.TargetCardsLeft
}
