package boardgame

import (
	"errors"
	"strconv"
	"strings"
)

//A Game represents a specific game between a collection of Players. Create a
//new one with NewGame().
type Game struct {
	//Name is a string that defines the type of game this is. The name should
	//be unique but human readable. Good examples are "Tic Tac Toe",
	//"Blackjack".
	Name string

	//Delegate is an (optional) way to override behavior at key game states.
	Delegate GameDelegate
	//State is the current state of the game.
	StateWrapper *StateWrapper
	//Finished is whether the came has been completed. If it is over, the
	//Winners will be set.
	Finished bool
	//Winners is the player indexes who were winners. Typically, this will be
	//one player, but it could be multiple in the case of tie, or 0 in the
	//case of a draw.
	Winners []int

	//Moves is the set of all move types that are ever legal to apply in this
	//game. When a move will be proposed it should copy one of these moves.
	moves       []Move
	movesByName map[string]Move

	//Initalized is set to True after SetUp is called.
	initalized bool
	chest      *ComponentChest

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

//NewGame returns a new game. You must set a Chest and call AddMove with all
//moves, before calling SetUp. Then the game can be used.
func NewGame(name string, initialState State, optionalDelegate GameDelegate) *Game {

	return &Game{
		Name:         name,
		Delegate:     optionalDelegate,
		StateWrapper: newStarterStateWrapper(initialState),
	}

}

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
	DistributeComponentToStarterStack(state State, c *Component) error

	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state State) (finished bool, winners []int)

	//ProposeFixUpMove is called after a move has been applied. It may return
	//a FixUp move, which will be applied before any other moves are applied.
	//If it returns nil, we may take the next move off of the queue. FixUp
	//moves are useful for things like shuffling a discard deck back into a
	//draw deck, or other moves that are necessary to get the GameState back
	//into reasonable shape.
	ProposeFixUpMove(state State) Move

	//SetGame is called during game.SetUp and passes a reference to the Game
	//that the delegate is part of.
	SetGame(game *Game)
}

//DefaultGameDelegate is a struct that implements stubs for all of
//GameDelegate's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one.
type DefaultGameDelegate struct {
	Game *Game
}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(state State, c *Component) error {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

func (d *DefaultGameDelegate) CheckGameFinished(state State) (finished bool, winners []int) {
	return false, nil
}

func (d *DefaultGameDelegate) ProposeFixUpMove(state State) Move {
	return nil
}

func (d *DefaultGameDelegate) SetGame(game *Game) {
	d.Game = game
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

	g.Delegate.SetGame(g)

	//Distribute all components to their starter locations

	//We'll work on a copy of Payload, so if it fails at some point we can just drop it
	stateCopy := g.StateWrapper.State.Copy()

	for _, name := range g.Chest().DeckNames() {
		deck := g.Chest().Deck(name)
		for i, component := range deck.Components() {
			if err := g.Delegate.DistributeComponentToStarterStack(stateCopy, component); err != nil {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ":" + err.Error())
			}
		}
	}

	g.movesByName = make(map[string]Move)
	for _, move := range g.moves {
		g.movesByName[strings.ToLower(move.Name())] = move
	}

	//If we got to here then the payloadCopy is now the real one.
	g.StateWrapper.State = stateCopy

	//TODO: do other set-up work, including FinishSetUp

	g.initalized = true

	return nil
}

//AddMove adds the specified move to the game. It may only be called during
//initalization.
func (g *Game) AddMove(move Move) {

	if g.initalized {
		return
	}
	g.moves = append(g.moves, move)
}

//Moves returns all moves that are valid in this game--all of the Moves that
//have been added via AddMove during initalization. Returns nil until
//game.SetUp() has been called.
func (g *Game) Moves() []Move {
	if !g.initalized {
		return nil
	}
	return g.moves
}

//MoveByName returns the Move of that name from game.Moves(), if it exists.
//Names are considered without regard to case.
func (g *Game) MoveByName(name string) Move {
	if !g.initalized {
		return nil
	}
	name = strings.ToLower(name)
	return g.movesByName[name]
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
		//If Finish was not already called in Chest it must be now--we can't
		//have it changing anymore. This will be a no-op if Finish() was
		//already called.

		//TODO: test that a chest that has not yet had finish called will when
		//added to a game.
		chest.Finish()
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

	//Verify that the Move is actually configured to be part of this game.
	if g.MoveByName(move.Name()) == nil {
		return errors.New("That move is not configured for this game.")
	}

	if err := move.Legal(g.StateWrapper.State); err != nil {
		//It's not legal, reject.
		return errors.New("The move was not legal: " + err.Error())
	}

	//TODO: keep track of historical states
	//TODO: persist new states to database here

	newState := g.StateWrapper.State.Copy()

	if err := move.Apply(newState); err != nil {
		return errors.New("The move's apply function returned an error:" + err.Error())
	}

	newStateWrapper := &StateWrapper{
		Version: g.StateWrapper.Version + 1,
		Schema:  g.StateWrapper.Schema,
		State:   newState,
	}

	g.StateWrapper = newStateWrapper

	//Check to see if that move made the game finished.
	if g.Delegate != nil {
		finished, winners := g.Delegate.CheckGameFinished(g.StateWrapper.State)

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
		move := g.Delegate.ProposeFixUpMove(g.StateWrapper.State)

		if move != nil {
			g.ApplyMove(move)
		}
	}

	return nil

}
