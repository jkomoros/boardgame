/*

memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"github.com/jkomoros/boardgame"
)

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "memory"
}

func (g *gameDelegate) DisplayName() string {
	return "Memory"
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers < 4 && numPlayers > 1
}

func (g *gameDelegate) EmptyGameState() boardgame.MutableGameState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	return &gameState{
		CurrentPlayer: 0,
		HiddenCards:   boardgame.NewSizedStack(cards, len(cards.Components())),
		RevealedCards: boardgame.NewSizedStack(cards, len(cards.Components())),
	}
}

func (g *gameDelegate) EmptyPlayerState(playerIndex int) boardgame.MutablePlayerState {

	cards := g.Manager().Chest().Deck(cardsDeckName)

	if cards == nil {
		return nil
	}

	return &playerState{
		playerIndex: playerIndex,
		WonCards:    boardgame.NewGrowableStack(cards, 0),
	}
}

func NewManager(storage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	cards := boardgame.NewDeck()

	for _, val := range cardNames {
		cards.AddComponentMulti(&cardValue{
			Type: val,
		}, 2)
	}

	chest.AddDeck(cardsDeckName, cards)

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		panic("No manager returned")
	}

	//TODO: add moves

	manager.SetUp()

	return manager
}
