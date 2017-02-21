/*

blackjack implements a simple blackjack game. This example is interesting
because it has hidden state.

*/
package blackjack

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/playingcards"
)

const targetScore = 21

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

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) error {

	s := state.(*mainState)

	card := c.Values.(*playingcards.Card)

	if card.Rank == playingcards.RankJoker {
		s.Game.UnusedCards.InsertFront(c)
	} else {
		s.Game.DrawStack.InsertFront(c)
	}

	return nil

}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []int) {

	s := state.(*mainState)

	for _, player := range s.Players {
		if !player.Busted && !player.Stood {
			return false, nil
		}
	}

	//OK, everyone has either busted or Stood. So who won?

	maxScore := 0

	for _, player := range s.Players {
		if player.Busted {
			continue
		}
		handValue := player.HandValue()
		if handValue > maxScore {
			maxScore = handValue
		}
	}

	//OK, now who got the maxScore?

	var result []int

	for i, player := range s.Players {
		if player.Busted {
			continue
		}
		handValue := player.HandValue()
		if handValue == maxScore {
			result = append(result, i)
		}
	}

	return true, result

}

func (g *gameDelegate) StartingState(numPlayers int) boardgame.State {
	cards := g.Manager().Chest().Deck("cards")

	if cards == nil {
		return nil
	}

	result := &mainState{
		Game: &gameState{
			DiscardStack:  boardgame.NewGrowableStack(cards, 0),
			DrawStack:     boardgame.NewGrowableStack(cards, 0),
			UnusedCards:   boardgame.NewGrowableStack(cards, 0),
			CurrentPlayer: 0,
		},
	}

	for i := 0; i < numPlayers; i++ {
		player := &playerState{
			playerIndex: i,
			Hand:        boardgame.NewGrowableStack(cards, 0),
			Busted:      false,
			Stood:       false,
		}
		result.Players = append(result.Players, player)
	}

	return result
}

func (g *gameDelegate) FinishSetUp(state boardgame.State) {
	s := state.(*mainState)

	s.Game.DrawStack.Shuffle()
}

func (g *gameDelegate) StateFromBlob(blob []byte) (boardgame.State, error) {
	result := &mainState{}
	if err := json.Unmarshal(blob, result); err != nil {
		return nil, err
	}

	result.Game.DrawStack.Inflate(g.Manager().Chest())
	result.Game.DiscardStack.Inflate(g.Manager().Chest())
	result.Game.UnusedCards.Inflate(g.Manager().Chest())

	for i, player := range result.Players {
		player.playerIndex = i
		player.Hand.Inflate(g.Manager().Chest())
	}

	return result, nil
}

func NewManager(storage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	chest.AddDeck("cards", playingcards.NewDeck(false))

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		panic("No manager returned")
	}

	manager.AddPlayerMove(&MoveCurrentPlayerHit{})
	manager.AddPlayerMove(&MoveCurrentPlayerStand{})

	manager.AddFixUpMove(&MoveShuffleDiscardToDraw{})
	manager.AddFixUpMove(&MoveAdvanceNextPlayer{})

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
