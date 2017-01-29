package boardgame

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications.
type Move interface {
	//Legal returns true if this proposed move is legal
	Legal(state *State) bool
	//Apply applies the move to the state in game (checking first whether it's legal)
	Apply(game *Game) bool
	PropertyReader
}
