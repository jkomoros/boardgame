/*

	debuganimations is a very simple debug "game" designed to allow us to
	exercise component animations very directly and purely, in order to build
	and debug that system.

*/
package debuganimations

import (
	"errors"
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

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)

	if c.Deck.Name() == tokensDeckName {

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

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) error {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

	return nil

}

func (g *gameDelegate) Diagram(state boardgame.State) string {
	return "Not implemented"
}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {
	//This debug game is never finished
	return false, nil
}

func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
	return boardgame.NewMoveTypeConfigBundle().AddMoves(
		&moveMoveCardBetweenShortStacksConfig,
		&moveMoveCardBetweenDrawAndDiscardStacksConfig,
		&moveFlipHiddenCardConfig,
		&moveMoveCardBetweenFanStacksConfig,
		&moveVisibleShuffleCardsConfig,
		&moveShuffleCardsConfig,
		&moveMoveBetweenHiddenConfig,
		&moveMoveBetweenHiddenConfig,
		&moveMoveTokenConfig,
		&moveMoveTokenSanitizedConfig,
	)
}

func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(nil)

	cards := boardgame.NewDeck()

	for _, val := range cardNames {
		cards.AddComponentMulti(&cardValue{
			Type: val,
		}, 3)
	}

	cards.SetShadowValues(&cardValue{
		Type: "<hidden>",
	})

	if err := chest.AddDeck(cardsDeckName, cards); err != nil {
		return nil, errors.New("Couldn't add deck: " + err.Error())
	}

	tokens := boardgame.NewDeck()

	tokens.AddComponentMulti(nil, 38)

	if err := chest.AddDeck(tokensDeckName, tokens); err != nil {
		return nil, errors.New("Couldnt add deck: " + err.Error())
	}

	return boardgame.NewGameManager(&gameDelegate{}, chest, storage)

}
