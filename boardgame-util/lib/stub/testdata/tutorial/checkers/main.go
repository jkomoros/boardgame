package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/with"
)

/*

Call the code generation for readers and enums here, so "go generate" will generate code correctly.

*/
//go:generate boardgame-util codegen

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "checkers"
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() == exampleCardDeckName {
		return new(exampleCardDynamicValues)
	}
	return nil
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	game := state.ImmutableGameState().(*gameState)
	if c.Deck().Name() == exampleCardDeckName {
		return game.DrawDeck, nil
	}
	return nil, errors.New("Unknown deck: " + c.Deck().Name())

}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game := state.GameState().(*gameState)
	return game.DrawDeck.Shuffle()
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		exampleCardDeckName: newExampleCardDeck(),
	}
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	//DefaultGameDelegate's CheckGameFinished checks this method and if true
	//looks at the score to see who won.

	//In this example, the game is over once all of the cards are gone.
	return state.ImmutableGameState().(*gameState).CardsDone()
}

func (g *gameDelegate) ComputedPlayerProperties(player boardgame.ImmutablePlayerState) boardgame.PropertyCollection {

	//ComputedProperties are mostly useful when a given state object's
	//computed property is useful clientside, too.

	p := player.(*playerState)

	return boardgame.PropertyCollection{
		"GameScore": p.GameScore(),
	}
}

func (g *gameDelegate) ComputedGlobalProperties(state boardgame.ImmutableState) boardgame.PropertyCollection {

	//ComputedProperties are mostly useful when a given state object's
	//computed property is useful clientside, too.

	game := state.ImmutableGameState().(*gameState)

	return boardgame.PropertyCollection{
		"CardsDone": game.CardsDone(),
	}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.Add(
			auto.MustConfig(new(moves.NoOp),
				with.MoveName("Example No Op Move"),
				with.HelpText("This move is an example that is always legal and does nothing. It exists to show how to return moves and make sure 'go test' works from the beginning, but you should remove it."),
			),
		),
	)

}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
