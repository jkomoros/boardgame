package boardgame

//A Game represents a specific game between a collection of Players
type Game struct {
	//Chest is the ComponentChest used for this game.
	Chest ComponentChest
	//State is the current state of the game.
	State *State

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}
