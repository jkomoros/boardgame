package boardgame

//A Game represents a specific game between a collection of Players
type Game struct {
	//Name is a string that defines the type of game this is. It is useful as
	//a sanity check to verify that various interface values were actually
	//intended to be used with this game.
	Name string

	//Delegate is an (optional) way to override behavior at key game states.
	Delegate GameDelegate
	//State is the current state of the game.
	State *State
	//Finished is whether the came has been completed. If it is over, the
	//Winners will be set.
	Finished bool
	//Winners is the player indexes who were winners. Typically, this will be
	//one player, but it could be multiple in the case of tie, or 0 in the
	//case of a draw.
	Winners []int

	chest *ComponentChest

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

//TODO: Create a NewGame()

//GameDelegate is called at various points in the game lifecycle. It is one of
//the primary ways that a specific game controls behavior over and beyond
//Moves and their Legal states.
type GameDelegate interface {
	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state StatePayload) (finished bool, winners []int)

	//ProposeFixUpMove is called after a move has been applied. It may return
	//a FixUp move, which will be applied before any other moves are applied.
	//If it returns nil, we may take the next move off of the queue. FixUp
	//moves are useful for things like shuffling a discard deck back into a
	//draw deck, or other moves that are necessary to get the GameState back
	//into reasonable shape.
	ProposeFixUpMove(state StatePayload) Move
}

type GameNamer interface {
	//GameName returns the string of the type of game we're designed for.
	//Before a move is applied to a game we verify that game.Name() and
	//move.GameName() match. Other types will similarly be gutchecked.
	GameName() string
}

//Chest is the ComponentChest in use for this game.
func (g *Game) Chest() *ComponentChest {
	return g.chest
}

//SetChest is the way to associate the given Chest with this game.
func (g *Game) SetChest(chest *ComponentChest) {
	chest.game = g
	g.chest = chest
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
		finished, winners := g.Delegate.CheckGameFinished(g.State.Payload)

		if finished {
			g.Finished = true
			g.Winners = winners
			//TODO: persist to database here.
		}
	}

	//TDOO: once we have a ProposedMove queue, instead of running this
	//syncrhounously, we should just inject the move (if it exists) into the
	//MoveQueue.
	if g.Delegate != nil {
		move := g.Delegate.ProposeFixUpMove(g.State.Payload)

		if move != nil {
			g.ApplyMove(move)
		}
	}

	return true

}
