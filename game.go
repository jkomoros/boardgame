package boardgame

//A Game represents a specific game between a collection of Players
type Game struct {
	//Name is a string that defines the type of game this is. It is useful as
	//a sanity check to verify that various interface values were actually
	//intended to be used with this game.
	Name string

	//Delegate is an (optional) way to override behavior at key game states.
	Delegate GameDelegate
	//Chest is the ComponentChest used for this game.
	Chest *ComponentChest
	//State is the current state of the game.
	State *State
	//Finished is whether the came has been completed. If it is over, the
	//Winners will be set.
	Finished bool
	//Winners is the player indexes who were winners. Typically, this will be
	//one player, but it could be multiple in the case of tie, or 0 in the
	//case of a draw.
	Winners []int

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

//GameDelegate is called at various points in the game lifecycle. It is one of
//the primary ways that a specific game controls behavior over and beyond
//Moves and their Legal states.
type GameDelegate interface {
	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state *State) (finished bool, winners []int)
}

type GameNamer interface {
	//GameName returns the string of the type of game we're designed for.
	//Before a move is applied to a game we verify that game.Name() and
	//move.GameName() match. Other types will similarly be gutchecked.
	GameName() string
}

//Game applies the move to the state if it is currently legal.
func (g *Game) ApplyMove(move Move) bool {

	if g.Finished {
		return false
	}

	//Verify that the Move is actually designed to be used with this type of
	//game.
	if move.GameName() != g.Name {
		return false
	}

	if !move.Legal(g.State.Payload) {
		//It's not legal, reject.
		return false
	}

	//TODO: keep track of historical states
	//TODO: persist new states to database here
	newStatePayload := move.Apply(g.State.Payload)

	if newStatePayload == nil {
		return false
	}

	newState := &State{
		Version: g.State.Version + 1,
		Schema:  g.State.Schema,
		Payload: newStatePayload,
	}

	g.State = newState

	//Check to see if that move made the game finished.
	if g.Delegate != nil {
		finished, winners := g.Delegate.CheckGameFinished(g.State)

		if finished {
			g.Finished = true
			g.Winners = winners
			//TODO: persist to database here.
		}
	}

	return true

}
