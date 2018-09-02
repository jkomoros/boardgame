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

func (g *gameDelegate) ConfigureConstants() map[string]interface{} {

	//ConfigureConstants isn't needed very often. It's useful to ensure a
	//constant value is available client-side, or if you want to use the value
	//in a struct tag.

	return map[string]interface{}{
		"numCards": numCards,
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

//values for the config setup
const (
	configKeyTargetCardsLeft = "targetcardsleft"
)

const (
	configTargetCardsLeftDefault = "default"
	configTargetCardsLeftShort   = "short"
)

func (g *gameDelegate) Configs() map[string][]string {

	//configs are the legal configuration options that will be show in the new
	//game dialog. Display names and description are returned in
	//ConfigKeyDisplay and ConfigValueDisplay.

	return map[string][]string{
		configKeyTargetCardsLeft: []string{
			configTargetCardsLeftDefault,
			configTargetCardsLeftShort,
		},
	}
}

func (g *gameDelegate) ConfigKeyDisplay(key string) (displayName, description string) {
	if key == configKeyTargetCardsLeft {
		return "Target Cards Left", "Whether or not the target cards left is the default"
	}
	return "", ""
}

func (g *gameDelegate) ConfigValueDisplay(key, val string) (displayName, description string) {
	if key == configKeyTargetCardsLeft {
		switch val {
		case configTargetCardsLeftShort:
			return "Short", "A short game that ends when 2 cards are left"
		default:
			return "Default", "A normal-length game that ends when no cards are left"
		}
	}
	return "", ""
}

func (g *gameDelegate) BeginSetUp(state boardgame.State, config boardgame.GameConfig) error {

	//This is the only time that config is passed in, so we need to interpret
	//it now and set it as a property in GameState.
	targetCardsLeftVal := config[configKeyTargetCardsLeft]
	if targetCardsLeftVal == "" {
		targetCardsLeftVal = configTargetCardsLeftDefault
	}

	var targetCardsLeft int

	switch targetCardsLeftVal {
	case configTargetCardsLeftShort:
		targetCardsLeft = 2
	case configTargetCardsLeftDefault:
		targetCardsLeft = 0
	default:
		return errors.New("Unknown value for " + configKeyTargetCardsLeft + ": " + targetCardsLeftVal)
	}

	game := state.GameState().(*gameState)
	game.TargetCardsLeft = targetCardsLeft

	return nil

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
