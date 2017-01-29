package boardgame

//ComponentAddress describes a specific Component in the game's ComponentChest.
type ComponentAddress struct {
	Deck  string
	Index int
}

//A Component represents a movable resource in the game. Cards, dice, meeples,
//resource tokens, etc are all components. Values is a struct that stores the
//specific values for the component.
type Component struct {
	Values PropertyReader
	//The name of the game this is associated with. Will be set automatically when added to a Deck.
	GameName string
	//The addresss of the component in the chest. Will be set automatically when added to a Deck.
	Address ComponentAddress
}
