/*

	debuganimations is a very simple debug "game" designed to allow us to
	exercise component animations very directly and purely, in order to build
	and debug that system.

*/
package debuganimations

import (
	"github.com/jkomoros/boardgame"
)

//go:generate autoreader

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "debuganimations"
}

func (g *gameDelegate) DisplayName() string {
	return "Animations Debugger"
}

func (g *gameDelegate) Description() string {
	return "A game type designed to test all of the stack animations in one place"
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 2
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: playerIndex,
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)

	if c.Deck().Name() == tokensDeckName {

		if game.TokensTo.NumComponents() < 9 {
			return game.TokensTo, nil
		}

		if game.SanitizedTokensFrom.NumComponents() < 10 {
			return game.SanitizedTokensFrom, nil
		}

		if game.SanitizedTokensTo.NumComponents() < 9 {
			return game.SanitizedTokensTo, nil
		}

		return game.TokensFrom, nil
	}

	if game.FirstShortStack.NumComponents() < 1 {
		return game.FirstShortStack, nil
	}

	if game.SecondShortStack.NumComponents() < 1 {
		return game.SecondShortStack, nil
	}

	if game.DiscardStack.NumComponents() < 2 {
		return game.DiscardStack, nil
	}

	if game.HiddenCard.NumComponents() < 1 {
		return game.HiddenCard, nil
	}

	if game.FanStack.NumComponents() < 6 {
		return game.FanStack, nil
	}

	if game.FanDiscard.NumComponents() < 3 {
		return game.FanDiscard, nil
	}

	if game.VisibleStack.NumComponents() < 5 {
		return game.VisibleStack, nil
	}

	if game.HiddenStack.NumComponents() < 4 {
		return game.HiddenStack, nil
	}

	return game.DrawStack, nil

}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil

}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	return "Not implemented"
}

func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {
	//This debug game is never finished
	return false, nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	return []boardgame.MoveConfig{
		moveMoveCardBetweenShortStacksConfig,
		moveMoveCardBetweenDrawAndDiscardStacksConfig,
		moveFlipHiddenCardConfig,
		moveMoveCardBetweenFanStacksConfig,
		moveVisibleShuffleCardsConfig,
		moveShuffleCardsConfig,
		moveMoveBetweenHiddenConfig,
		moveMoveTokenConfig,
		moveMoveTokenSanitizedConfig,
	}
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {

	cards := boardgame.NewDeck()

	for _, val := range cardNames {
		for i := 0; i < 3; i++ {
			cards.AddComponent(&cardValue{
				Type: val,
			})
		}
	}

	cards.SetGenericValues(&cardValue{
		Type: "<hidden>",
	})

	tokens := boardgame.NewDeck()

	for i := 0; i < 38; i++ {
		tokens.AddComponent(nil)
	}

	return map[string]*boardgame.Deck{
		cardsDeckName:  cards,
		tokensDeckName: tokens,
	}
}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
