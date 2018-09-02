package checkers

import (
	"github.com/jkomoros/boardgame"
)

const numCards = 10
const exampleCardDeckName = "examplecards"

//boardgame:codegen
type exampleCard struct {
	Value int
}

//newExampleCardDeck returns a new deck for examplecards.
func newExampleCardDeck() *boardgame.Deck {
	deck := boardgame.NewDeck()

	for i := 0; i < numCards; i++ {
		deck.AddComponent(&exampleCard{
			Value: i + 1,
		})
	}

	return deck
}
