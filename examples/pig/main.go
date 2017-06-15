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

func (g *gameDelegate) EmptyGameState() boardgame.MutableSubState {
	//TODO: implement
	return nil
}

func (g *gameDelegate) EmptyPlayerState(index boardgame.PlayerIndex) boardgame.MutablePlayerState {
	//TODO: implement
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

	//TODO: configure moves here

	manager.SetUp()

	return manager
}
