package boardgame

import (
	"errors"
)

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

	//Initalized is set to True after SetUp is called.
	initalized bool
	chest      *ComponentChest

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

//TODO: Create a NewGame()

//GameDelegate is called at various points in the game lifecycle. It is one of
//the primary ways that a specific game controls behavior over and beyond
//Moves and their Legal states.
type GameDelegate interface {

	//DistributeComponentToStarterStack is called during set up to establish
	//the Deck/Stack invariant that every component in the chest is placed in
	//precisely one Stack. Game will call this on each component in the Chest
	//in order. This is where the logic goes to make sure each Component goes
	//into its correct starter stack.
	DistributeComponentToStarterStack(c *Component)

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

//SetUp should be called a single time after all of the member variables are
//set correctly, including Chest. SetUp must be called before ApplyMove can be
//called.
func (g *Game) SetUp() error {

	if g.initalized {
		return errors.New("Game already initalized")
	}

	if g.chest == nil {
		return errors.New("No component chest set")
	}

	//Distribute all components to their starter locations
	if g.Delegate != nil {
		for _, name := range g.Chest().DeckNames() {
			deck := g.Chest().Deck(name)
			for _, component := range deck.Components() {
				g.Delegate.DistributeComponentToStarterStack(component)
			}
		}
	}

	//TODO: do other set-up work, including FinishSetUp

	g.initalized = true

	return nil
}

//Chest is the ComponentChest in use for this game.
func (g *Game) Chest() *ComponentChest {
	return g.chest
}

//SetChest is the way to associate the given Chest with this game.
func (g *Game) SetChest(chest *ComponentChest) {
	//We are only allowed to change the chest before the game is SetUp.
	if g.initalized {
		return
	}
	if chest != nil {
		chest.game = g
	}
	g.chest = chest
}

//Game applies the move to the state if it is currently legal.
func (g *Game) ApplyMove(move Move) error {

	if !g.initalized {
		return errors.New("The game has not been initalized.")
	}

	if g.Finished {
		return errors.New("Game was already finished")
	}

	//Verify that the Move is actually designed to be used with this type of
	//game.
	if move.GameName() != g.Name {
		return errors.New("The move expected a game of a different name")
	}

	if err := move.Legal(g.State.Payload); err != nil {
		//It's not legal, reject.
		return errors.New("The move was not legal: " + err.Error())
	}

	//TODO: keep track of historical states
	//TODO: persist new states to database here
	newStatePayload := move.Apply(g.State.Payload)

	if newStatePayload == nil {
		return errors.New("The move's Apply function did not return a modified state")
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

	return nil

}
