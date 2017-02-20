package blackjack

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"testing"
)

func TestHandValue(t *testing.T) {

	manager := NewManager(memory.NewStorageManager())

	chest := manager.Chest()

	deck := chest.Deck("cards")

	tests := []struct {
		state    *playerState
		expected int
	}{
		{
			&playerState{
				Hand: createHand(deck, Rank2, Rank3, Rank4),
			},
			9,
		},
		{
			&playerState{
				Hand: createHand(deck, RankAce),
			},
			11,
		},
		{
			&playerState{
				Hand: createHand(deck, RankAce, RankAce),
			},
			12,
		},
		{
			&playerState{
				Hand: createHand(deck, RankJack, RankKing, RankAce),
			},
			21,
		},
	}

	for i, test := range tests {
		result := test.state.HandValue()

		if result != test.expected {
			t.Error("Test", i, "Failed. Got", result, "wanted", test.expected)
		}
	}
}

func createHand(deck *boardgame.Deck, ranks ...Rank) *boardgame.GrowableStack {
	result := boardgame.NewGrowableStack(deck, 0)

	givenCards := make(map[int]bool)

	for i, rank := range ranks {
		for i, c := range deck.Components() {

			//Skip cards we have already used
			if givenCards[i] {
				continue
			}

			card := c.Values.(*Card)

			if card.Rank == rank {
				//Found one!
				result.InsertBack(c)
				givenCards[i] = true
				break
			}
		}
		if result.Len() <= i {
			//Didn't find a card, must not be any left!
			panic("Wasn't possible to fulfill the cards you asked for")
		}
	}

	return result

}
