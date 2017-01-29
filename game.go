package boardgame

//A Game represents a specific game between a collection of Players
type Game struct {
	//Chest is the ComponentChest used for this game.
	Chest ComponentChest
	//State is the current state of the game.
	State *State

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: a Name property that moves can gutcheck apply

	//TODO: an array of Player objects.
}

//Game applies the move to the state if it is currently legal.
func (g *Game) ApplyMove(move Move) bool {

	//TODO: test this

	if !move.Legal(g.State) {
		//It's not legal, reject.
		return false
	}

	//TODO: keep track of historical states
	//TODO: persist new states to database here
	newState := move.Apply(g.State)

	if newState == nil {
		return false
	}

	//Make sure that the version number monotonically increases.
	newState.Version = g.State.Version + 1

	g.State = newState

	return true

}
