package boardgame

//A Game represents a specific game between a collection of Players
type Game struct {
	//Name is a string that defines the type of game this is. It is useful as
	//a sanity check to verify that various interface values were actually
	//intended to be used with this game.
	Name string
	//Chest is the ComponentChest used for this game.
	Chest ComponentChest
	//State is the current state of the game.
	State *State

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

type GameNamer interface {
	//GameName returns the string of the type of game we're designed for.
	//Before a move is applied to a game we verify that game.Name() and
	//move.GameName() match. Other types will similarly be gutchecked.
	GameName() string
}

//Game applies the move to the state if it is currently legal.
func (g *Game) ApplyMove(move Move) bool {

	//TODO: test this

	//Verify that the Move is actually designed to be used with this type of
	//game.
	if move.GameName() != g.Name {
		return false
	}

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
