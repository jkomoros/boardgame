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

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers == 2
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) GameStateConstructor() boardgame.MutableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.MutablePlayerState {
	return &playerState{
		playerIndex: playerIndex,
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)

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

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) {
	game, _ := concreteStates(state)

	game.DrawStack.Shuffle()

}

func (g *gameDelegate) Diagram(state boardgame.State) string {
	return "Not implemented"
}

var policy *boardgame.StatePolicy

func (g *gameDelegate) StateSanitizationPolicy() *boardgame.StatePolicy {

	if policy == nil {
		policy = &boardgame.StatePolicy{
			Game: map[string]boardgame.GroupPolicy{
				"DrawStack": {
					boardgame.GroupAll: boardgame.PolicyOrder,
				},
				"HiddenCard": {
					boardgame.GroupAll: boardgame.PolicyOrder,
				},
				"FirstShortStack": {
					boardgame.GroupAll: boardgame.PolicyOrder,
				},
				"SecondShortStack": {
					boardgame.GroupAll: boardgame.PolicyOrder,
				},
				"FanDiscard": {
					boardgame.GroupAll: boardgame.PolicyOrder,
				},
				"HiddenStack": {
					boardgame.GroupAll: boardgame.PolicyNonEmpty,
				},
			},
		}
	}

	return policy

}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {
	//This debug game is never finished
	return false, nil
}

func MustNewManager(storage boardgame.StorageManager) *boardgame.GameManager {

	manager, err := NewManager(storage)

	if err != nil {
		panic("Couldn't create manager: " + err.Error())
	}

	return manager

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

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		return nil, errors.New("No manager returned")
	}

	moveTypeConfigs := []*boardgame.MoveTypeConfig{
		&moveMoveCardBetweenShortStacksConfig,
		&moveMoveCardBetweenDrawAndDiscardStacksConfig,
		&moveFlipHiddenCardConfig,
		&moveMoveCardBetweenFanStacksConfig,
		&moveVisibleShuffleCardsConfig,
		&moveMoveBetweenHiddenConfig,
		&moveMoveBetweenHiddenConfig,
	}

	if err := manager.BulkAddMoveTypes(moveTypeConfigs); err != nil {
		return nil, errors.New("Couldn't create moves: " + err.Error())
	}

	if err := manager.SetUp(); err != nil {
		return nil, errors.New("Couldn't set up manager: " + err.Error())
	}

	return manager, nil
}
