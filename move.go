package boardgame

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications.
type Move interface {
	//Legal returns true if this proposed move is legal
	Legal(state *State) bool
	//TODO: figure out how to get a string describing why it's not legal out

	//Apply applies the move to the state and returns a new state object. It
	//should not be called directly; use Game.ApplyMove.
	Apply(state *State) *State
	GameNamer
	PropertyReader
	JSONer
}
