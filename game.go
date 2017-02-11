package boardgame

import (
	"errors"
	"strconv"
	"strings"
)

//maxRecurseCount is the number of fixUp moves that can be considered normal--
//anything more than that and we'll panic because the delegate is likely going
//to return fixup moves forever.
const maxRecurseCount = 50

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
	//Player moves are moves that can be applied by users. FixUp moves are
	//only ever returned by Delegate.ProposeFixUpMove().
	playerMoves       []Move
	fixUpMoves        []Move
	playerMovesByName map[string]Move
	fixUpMovesByName  map[string]Move

	//Proposed moves is where moves that have been proposed but have not yet been applied go.
	proposedMoves chan *proposedMoveItem

	//Initalized is set to True after SetUp is called.
	initalized bool
	chest      *ComponentChest

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

type DelayedError chan error

type proposedMoveItem struct {
	move Move
	//Ch is the channel we should either return an error on and then close, or
	//send nil and close.
	ch DelayedError
}

//NewGame returns a new game. You must set a Chest and call AddMove with all
//moves, before calling SetUp. Then the game can be used.
func NewGame(name string, initialState State, optionalDelegate GameDelegate) *Game {

	result := &Game{
		Name:         name,
		Delegate:     optionalDelegate,
		StateWrapper: newStarterStateWrapper(initialState),
		//TODO: set the size of chan based on something more reasonable.
		proposedMoves: make(chan *proposedMoveItem, 20),
	}

	return result

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
//set correctly, including Chest. SetUp must be called before ProposeMove can be
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

	g.playerMovesByName = make(map[string]Move)
	for _, move := range g.playerMoves {
		g.playerMovesByName[strings.ToLower(move.Name())] = move
	}

	g.fixUpMovesByName = make(map[string]Move)
	for _, move := range g.fixUpMoves {
		g.fixUpMovesByName[strings.ToLower(move.Name())] = move
	}

	//If we got to here then the payloadCopy is now the real one.
	g.StateWrapper.State = stateCopy

	//TODO: do other set-up work, including FinishSetUp

	go g.mainLoop()

	g.initalized = true

	return nil
}

//MainLoop should be run in a goroutine. It is what takes moves off of
//proposedMoves and applies them. It is the only method that may call
//applyMove.
func (g *Game) mainLoop() {

	for item := range g.proposedMoves {
		if item == nil {
			return
		}
		item.ch <- g.applyMove(item.move, false, 0)
		close(item.ch)
	}

}

//AddPlayerMove adds the specified move to the game as a move that Players can
//make. It may only be called during initalization.
func (g *Game) AddPlayerMove(move Move) {

	if g.initalized {
		return
	}
	g.playerMoves = append(g.playerMoves, move)
}

//AddFixUpMove adds a move that can only be legally made by GameDelegate as a
//FixUp move. It can only be called during initialization.
func (g *Game) AddFixUpMove(move Move) {
	if g.initalized {
		return
	}
	g.fixUpMoves = append(g.fixUpMoves, move)
}

//PlayerMoves returns all moves that are valid in this game to be made my
//players--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies, with them already set to the proper
//DefaultsForState.
func (g *Game) PlayerMoves() []Move {
	if !g.initalized {
		return nil
	}

	result := make([]Move, len(g.playerMoves))

	for i, move := range g.playerMoves {
		result[i] = move.Copy()
		result[i].DefaultsForState(g.StateWrapper.State)
	}

	return result
}

//FixUpMoves returns all moves that are valid in this game to be made as fixup
//moves--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies, with them already set to the proper
//DefaultsForState.
func (g *Game) FixUpMoves() []Move {

	//TODO: test all of these fixup moves

	if !g.initalized {
		return nil
	}

	result := make([]Move, len(g.fixUpMoves))

	for i, move := range g.fixUpMoves {
		result[i] = move.Copy()
		result[i].DefaultsForState(g.StateWrapper.State)
	}

	return result
}

//PlayerMoveByName returns the Move of that name from game.PlayerMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy
//with defaults already set for current game state by move.DefaultsForState.
func (g *Game) PlayerMoveByName(name string) Move {
	if !g.initalized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.playerMovesByName[name]

	if move == nil {
		return nil
	}

	result := move.Copy()
	result.DefaultsForState(g.StateWrapper.State)
	return result
}

//FixUpMoveByName returns the Move of that name from game.FixUpMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy
//with defaults already set for current game state by move.DefaultsForState.
func (g *Game) FixUpMoveByName(name string) Move {
	if !g.initalized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.fixUpMovesByName[name]

	if move == nil {
		return nil
	}

	result := move.Copy()
	result.DefaultsForState(g.StateWrapper.State)
	return result
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

//ProposedMove is the way to propose a move to the game. DelayedError will
//return an error in the future if the move was unable to be applied, or nil
//if the move was applied successfully. DelayedError will only resolve once
//any applicable FixUp moves have been applied already. Note: DelayedError
//won't return anything until after SetUp has been called.
func (g *Game) ProposeMove(move Move) DelayedError {

	errChan := make(DelayedError, 1)

	workItem := &proposedMoveItem{
		move: move,
		ch:   errChan,
	}

	g.proposedMoves <- workItem

	return errChan

}

//Game applies the move to the state if it is currently legal. May only be
//called by mainLoop. Propose moves with game.ProposeMove instead.
func (g *Game) applyMove(move Move, isFixUp bool, recurseCount int) error {

	if !g.initalized {
		return errors.New("The game has not been initalized.")
	}

	if g.Finished {
		return errors.New("Game was already finished")
	}

	if isFixUp {
		if g.FixUpMoveByName(move.Name()) == nil {
			return errors.New("That move is not configured as a Fix Up move for this game.")
		}
	} else {

		//Verify that the Move is actually configured to be part of this game.
		if g.PlayerMoveByName(move.Name()) == nil {
			return errors.New("That move is not configured as a Player move for this game.")
		}
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

	if g.Delegate != nil {

		if recurseCount > maxRecurseCount {
			panic("We recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal. Quitting.")
		}

		move := g.Delegate.ProposeFixUpMove(g.StateWrapper.State)

		if move != nil {
			//We apply the move immediately. This ensures that when
			//DelayedError resolves, all of the fix up moves have been
			//applied.
			g.applyMove(move, true, recurseCount+1)
		}
	}

	return nil

}
