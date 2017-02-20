/*

blackjack implements a simple blackjack game. This example is interesting
because it has hidden state.

*/
package blackjack

import (
	"github.com/jkomoros/boardgame"
)

const gameDisplayname = "Blackjack"
const gameName = "blackjack"

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return gameName
}

func (g *gameDelegate) DisplayName() string {
	return gameDisplayname
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 4
}

func NewManager(storage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	cards := &boardgame.Deck{}

	ranks := []Rank{RankAce, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10, RankJack, RankQueen, RankKing}
	suits := []Suit{SuitClubs, SuitDiamonds, SuitHearts, SuitSpades}

	for _, rank := range ranks {
		for _, suit := range suits {
			cards.AddComponent(&Card{
				Suit: suit,
				Rank: rank,
			})
		}
	}

	//Add two Jokers
	cards.AddComponentMulti(&Card{
		Suit: SuitJokers,
		Rank: RankJoker,
	}, 2)

	chest.AddDeck("cards", cards)

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		panic("No manager returned")
	}

	//TODO: add moves

	manager.SetUp()

	return manager
}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := boardgame.NewGame(manager)

	if err := game.SetUp(0); err != nil {
		panic(err)
	}

	return game
}
