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
func (g *GameManager) LoadGame(name string, id string, modifiable bool, version int, finished bool, winners []int) *Game {

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
		modifiable: modifiable,
		id:         id,
		finished:   finished,
		winners:    winners,
		initalized: true,
	}

	if result.Modifiable() {
		go result.mainLoop()
	}

	return result

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
