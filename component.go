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
	Values   PropertyReader
	GameName string
	Address  ComponentAddress
}
