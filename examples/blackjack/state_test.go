package blackjack

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestHandValue(t *testing.T) {

	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	chest := manager.Chest()

	deck := chest.Deck("cards")

	tests := []struct {
		components []boardgame.Component
		expected   int
	}{
		{
			createHand(t, deck, playingcards.Rank2, playingcards.Rank3, playingcards.Rank4),
			9,
		},
		{
			createHand(t, deck, playingcards.RankAce),
			11,
		},
		{
			createHand(t, deck, playingcards.RankAce, playingcards.RankAce),
			12,
		},
		{
			createHand(t, deck, playingcards.RankJack, playingcards.RankKing, playingcards.RankAce),
			21,
		},
	}

	for i, test := range tests {
		result := handValue(test.components)

		if result != test.expected {
			t.Error("Test", i, "Failed. Got", result, "wanted", test.expected)
		}
	}
}

func createHand(t *testing.T, deck *boardgame.Deck, ranks ...int) []boardgame.Component {
	var result []boardgame.Component

	givenCards := make(map[int]bool)

	for i, rank := range ranks {
		for i, c := range deck.Components() {

			//Skip cards we have already used
			if givenCards[i] {
				continue
			}

			card := c.Values().(*playingcards.Card)

			if card.Rank.Value() == rank {
				//Found one!
				result = append(result, c)
				givenCards[i] = true
				break
			}
		}
		if len(result) <= i {
			//Didn't find a card, must not be any left!
			panic("Wasn't possible to fulfill the cards you asked for")
		}
	}

	return result

}
