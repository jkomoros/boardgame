package boardgame

//A Component represents a movable resource in the game. Cards, dice, meeples,
//resource tokens, etc are all components. Values is a struct that stores the
//specific values for the component.
type Component struct {
	Values ComponentValues
	//The deck we're a part of.
	Deck *Deck
	//The index we are in the deck we're in.
	DeckIndex int
}

type ComponentValues interface {
	PropertyReader
}
