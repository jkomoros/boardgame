package boardgame

import (
	"errors"
	"strings"
)

//Moves is the set of all move types that are ever legal to apply in this
//game. When a move will be proposed it should copy one of these moves.
//Player moves are moves that can be applied by users. FixUp moves are
//only ever returned by Delegate.ProposeFixUpMove().

//TODO: figure out where the above comment should go for documentation.

//GameManager is a struct that keeps track of configuration that is common
//across multiple games. It is specifically designed to be used with multiple
//games.
type GameManager struct {
	delegate          GameDelegate
	chest             *ComponentChest
	storage           StorageManager
	fixUpMoves        []Move
	playerMoves       []Move
	fixUpMovesByName  map[string]Move
	playerMovesByName map[string]Move
	modifiableGames   map[string]*Game
	initialized       bool
}

//NewGameManager creates a new game manager with the given delegate.
func NewGameManager(delegate GameDelegate, chest *ComponentChest, storage StorageManager) *GameManager {
	if delegate == nil {
		return nil
	}

	if chest == nil {
		return nil
	}

	//Make sure the chest is no longer open for modification. If finish was
	//already called, this will be a no-op.
	chest.Finish()

	if storage == nil {
		return nil
	}

	result := &GameManager{
		delegate: delegate,
		chest:    chest,
		storage:  storage,
	}

	chest.manager = result

	delegate.SetManager(result)

	return result
}

//LoadGame is used to provide back a real Game instance based on state that
//was stored in storage, that is ready to use like any other game (that is, it
//operates like SetUp has already been called). If you want a new game, use
//NewGame.
func (g *GameManager) LoadGame(name string, id string, version int, finished bool, winners []int) *Game {

	//It feels really weird that this is exposed, but I think something like
	//it has to be so that others can implement their own StorageManagers
	//without being able to modify Game's internal fields.

	//Sanity check that this game actually does match with this manager.
	if name != g.delegate.Name() {
		return nil
	}

	result := &Game{
		manager:    g,
		version:    version,
		id:         id,
		finished:   finished,
		winners:    winners,
		modifiable: false,
		initalized: true,
	}

	return result

}

//modifiableGameCreated lets Manager know that a modifiable game was created
//with the given ID, so that manager can vend that later if necessary. It is
//designed to only be called from NewGame.
func (g *GameManager) modifiableGameCreated(game *Game) {
	if !g.initialized {
		return
	}

	if _, ok := g.modifiableGames[game.Id()]; ok {
		panic("modifiableGameCreated collided with existing game")
	}

	id := strings.ToUpper(game.Id())

	g.modifiableGames[id] = game
}

//ModifiableGameForId returns a modifiable game with the given ID. Either it
//returns one it already knows about, or it creates a modifiable version from
//storage (if one is stored in storage). If a game cannot be created from
//those ways, it will return nil. The primary way to avoid race
//conditions with the same underlying game being stored to the store is
//that only one modifiable copy of a Game should exist at a time. It is up
//to the specific user of boardgame to ensure that is the case. As long as
//manager.LoadGame is used, a single manager will not allow multiple
//modifiable versions of a single game to be "checked out".  However, if
//there could be multiple managers loaded up at the same time for the same
//store, it's possible to have a race condition. For example, it makes
//sense to have only a single server that takes in proposed moves from a
//queue and then applies them to a modifiable version of the given game.
func (g *GameManager) ModifiableGame(id string) *Game {

	id = strings.ToUpper(id)

	game := g.modifiableGames[id]

	if game != nil {
		return game
	}

	//Let's try to load up from storage.

	game, _ = g.storage.Game(g, id)

	if game == nil {
		//Nah, we've never seen that game.
		return nil
	}

	//Only SetUp() and us are allowed to kick off a game's mainLoop.
	game.modifiable = true
	//TODO: set the size of chan based on something more reasonable.
	//Note: this is also set similarly in NewGame
	game.proposedMoves = make(chan *proposedMoveItem, 20)
	go game.mainLoop()

	g.modifiableGames[id] = game

	return game

}

//Game fetches a new non-modifiable copy of the given game from storage. If
//you want a modifiable version, use ModifiableGame.
func (g *GameManager) Game(id string) *Game {
	if result, err := g.storage.Game(g, id); err == nil {
		return result
	}
	return nil
}

//proposeMoveOnGame is how non-modifiable games should tell the manager they
//have a move they want to make on a given move ID. For now it's just a simple
//wrapper around ModifiableGame, but in multi-server situations, in the future
//it would conceivably do an RPC or something.
func (g *GameManager) proposeMoveOnGame(id string, move Move) DelayedError {

	errChan := make(DelayedError, 1)

	//ModifiableGame could take awhile if it has to be fetched from storage,
	//so we'll run all of this in a goroutine since we're returning a
	//DelayedError anyway.

	go func() {
		game := g.ModifiableGame(id)

		if game == nil {
			errChan <- errors.New("There was no game with that ID")
			return
		}

		workItem := &proposedMoveItem{
			move: move,
			ch:   errChan,
		}

		game.proposedMoves <- workItem

	}()

	return errChan

}

//SetUp should be called before this Manager is used. It locks in moves,
//chest, storage, etc.
func (g *GameManager) SetUp() error {

	if g.chest == nil {
		return errors.New("No chest provided")
	}

	if g.storage == nil {
		return errors.New("Storage not provided")
	}

	g.playerMovesByName = make(map[string]Move)
	for _, move := range g.playerMoves {
		g.playerMovesByName[strings.ToLower(move.Name())] = move
	}

	g.fixUpMovesByName = make(map[string]Move)
	for _, move := range g.fixUpMoves {
		g.fixUpMovesByName[strings.ToLower(move.Name())] = move
	}

	g.modifiableGames = make(map[string]*Game)

	g.initialized = true

	return nil
}

//AddPlayerMove adds the specified move to the game as a move that Players can
//make. It may only be called during initalization.
func (g *GameManager) AddPlayerMove(move Move) {

	if g.initialized {
		return
	}
	g.playerMoves = append(g.playerMoves, move)
}

//AddFixUpMove adds a move that can only be legally made by GameDelegate as a
//FixUp move. It can only be called during initialization.
func (g *GameManager) AddFixUpMove(move Move) {
	if g.initialized {
		return
	}
	g.fixUpMoves = append(g.fixUpMoves, move)
}

//PlayerMoves returns all moves that are valid in this game to be made my
//players--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies.
func (g *GameManager) PlayerMoves() []Move {
	if !g.initialized {
		return nil
	}

	result := make([]Move, len(g.playerMoves))

	for i, move := range g.playerMoves {
		result[i] = move.Copy()
	}

	return result
}

//FixUpMoves returns all moves that are valid in this game to be made as fixup
//moves--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies.
func (g *GameManager) FixUpMoves() []Move {

	//TODO: test all of these fixup moves

	if !g.initialized {
		return nil
	}

	result := make([]Move, len(g.fixUpMoves))

	for i, move := range g.fixUpMoves {
		result[i] = move.Copy()
	}

	return result
}

//PlayerMoveByName returns the Move of that name from game.PlayerMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy.
func (g *GameManager) PlayerMoveByName(name string) Move {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.playerMovesByName[name]

	if move == nil {
		return nil
	}

	return move.Copy()
}

//FixUpMoveByName returns the Move of that name from game.FixUpMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy.
func (g *GameManager) FixUpMoveByName(name string) Move {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.fixUpMovesByName[name]

	if move == nil {
		return nil
	}

	return move.Copy()
}

//Chest is the ComponentChest in use for this game. Will return nil until
//SetUp() called.
func (g *GameManager) Chest() *ComponentChest {
	return g.chest
}

//Storage is the StorageManager games that use this manager should use.
func (g *GameManager) Storage() StorageManager {
	return g.storage
}

//Delegate returns the GameDelegate configured for these games.
func (g *GameManager) Delegate() GameDelegate {
	return g.delegate
}
