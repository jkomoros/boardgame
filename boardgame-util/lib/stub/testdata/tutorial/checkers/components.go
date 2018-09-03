package checkers

import (
	"github.com/jkomoros/boardgame"
)

const numCards = 10
const exampleCardDeckName = "examplecards"

//boardgame:codegen
type exampleCard struct {
	boardgame.BaseComponentValues
	Value int
}

//boardgame:codegen
type exampleCardDynamicValues struct {
	boardgame.BaseSubState
	boardgame.BaseComponentValues
	DynamicValue int
}

//newExampleCardDeck returns a new deck for examplecards.
func newExampleCardDeck() *boardgame.Deck {
	deck := boardgame.NewDeck()

	for i := 0; i < numCards; i++ {
		deck.AddComponent(&exampleCard{
			Value: i + 1,
		})
	}

	//Set the value to return whenever the stack is sanitized. If we didn't
	//set this then sometimes the ComponentValues in a stack would be nil when
	//they are sanitized, which is error-prone for methods. It's always best
	//to set a reasonable generic value so that methods can always assume non-
	//nil ComponentValues.
	deck.SetGenericValues(&exampleCard{
		Value: 0,
	})

	return deck
}
