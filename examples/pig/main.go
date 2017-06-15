/*
 *
 * pig is a very simple game involving dice rolls.
 *
 */
package pig

import (
	"github.com/jkomoros/boardgame"
)

//go:generate autoreader

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "pig"
}

func (g *gameDelegate) DisplayName() string {
	return "Pig"
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers > 0 && numPlayers < 6
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) EmptyGameState() boardgame.MutableSubState {
	dice := g.Manager().Chest().Deck(diceDeckName)

	if dice == nil {
		return nil
	}

	return &gameState{
		CurrentPlayer: 0,
		Die:           boardgame.NewSizedStack(dice, 1),
	}
}

func (g *gameDelegate) EmptyPlayerState(index boardgame.PlayerIndex) boardgame.MutablePlayerState {
	return &playerState{
		playerIndex: index,
		TotalScore:  0,
		RoundScore:  0,
		DieCounted:  true,
		Done:        false,
		Busted:      false,
	}
}

func (g *gameDelegate) EmptyDynamicComponentValues(deck *boardgame.Deck) boardgame.MutableSubState {
	if deck.Name() == diceDeckName {
		return &dieDynamicValue{
			Value: 1,
		}
	}
	return nil
}

func NewManager(storage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	dice := boardgame.NewDeck()

	dice.AddComponent(DefaultDie())

	chest.AddDeck(diceDeckName, dice)

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		panic("No manager returned")
	}

	manager.AddPlayerMoveFactory(MoveRollDiceFactory)
	manager.AddPlayerMoveFactory(MoveDoneTurnFactory)

	manager.AddFixUpMoveFactory(MoveCountDieFactory)
	manager.AddFixUpMoveFactory(MoveAdvanceNextPlayerFactory)

	manager.SetUp()

	return manager
}
