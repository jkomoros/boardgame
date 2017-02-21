package blackjack

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/playingcards"
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
				Hand: createHand(deck, playingcards.Rank2, playingcards.Rank3, playingcards.Rank4),
			},
			9,
		},
		{
			&playerState{
				Hand: createHand(deck, playingcards.RankAce),
			},
			11,
		},
		{
			&playerState{
				Hand: createHand(deck, playingcards.RankAce, playingcards.RankAce),
			},
			12,
		},
		{
			&playerState{
				Hand: createHand(deck, playingcards.RankJack, playingcards.RankKing, playingcards.RankAce),
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

func createHand(deck *boardgame.Deck, ranks ...playingcards.Rank) *boardgame.GrowableStack {
	result := boardgame.NewGrowableStack(deck, 0)

	givenCards := make(map[int]bool)

	for i, rank := range ranks {
		for i, c := range deck.Components() {

			//Skip cards we have already used
			if givenCards[i] {
				continue
			}

			card := c.Values.(*playingcards.Card)

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
