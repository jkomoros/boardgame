package boardgame

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications.
type Move interface {
	//Legal returns true if this proposed move is legal
	Legal(state *State) bool
	//Apply applies the move to the state and returns a new state object. It
	//should not be called directly; use Game.ApplyMove.
	Apply(state *State) *State
	//GameName returns the string of the type of game we're designed for.
	//Before a move is applied to a game we verify that game.Name() and
	//move.GameName() match.
	GameName() string
	PropertyReader
}
