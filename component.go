package boardgame

//A Component represents a movable resource in the game. Cards, dice, meeples,
//resource tokens, etc are all components.
type Component interface {
	//The name of the Deck in the game's ComponentChest that we are part of.
	Deck() string
	//The index that this component resides at in its deck.
	DeckIndex() string
	PropertyReader
}
