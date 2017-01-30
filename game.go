package boardgame

import (
	"errors"
	"strconv"
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
	//into its correct starter stack. As long as you put each component into a
	//Stack, the invariant will be met at the end of SetUp. If any errors are
	//returned SetUp fails. Unlike after the game has been SetUp, you can
	//modify payload directly.
	DistributeComponentToStarterStack(payload StatePayload, c *Component) error

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

//DefaultGameDelegate is a struct that implements stubs for all of
//GameDelegate's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one.
type DefaultGameDelegate struct{}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(payload StatePayload, c *Component) error {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

func (d *DefaultGameDelegate) CheckGameFinished(state StatePayload) (finished bool, winners []int) {
	return false, nil
}

func (d *DefaultGameDelegate) ProposeFixUpMove(state StatePayload) Move {
	return nil
}

//SetUp should be called a single time after all of the member variables are
//set correctly, including Chest. SetUp must be called before ApplyMove can be
//called. Even if an error is returned, the game should be in a consistent state.
func (g *Game) SetUp() error {

	if g.initalized {
		return errors.New("Game already initalized")
	}

	if g.chest == nil {
		return errors.New("No component chest set")
	}

	if g.Delegate == nil {
		g.Delegate = &DefaultGameDelegate{}
	}

	//Distribute all components to their starter locations

	//We'll work on a copy of Payload, so if it fails at some point we can just drop it
	payloadCopy := g.State.Payload.Copy()

	for _, name := range g.Chest().DeckNames() {
		deck := g.Chest().Deck(name)
		for i, component := range deck.Components() {
			if err := g.Delegate.DistributeComponentToStarterStack(payloadCopy, component); err != nil {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ":" + err.Error())
			}
		}
	}

	//If we got to here then the payloadCopy is now the real one.
	g.State.Payload = payloadCopy

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
