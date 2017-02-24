package boardgame

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications.
type Move interface {
	//Legal returns nil if this proposed move is legal, or an error if the
	//move is not legal
	Legal(state State) error

	//Apply applies the move to the state. It is handed a copy of the state to
	//modify. If error is non-nil it will not be applied to the game. It
	//should not be called directly; use Game.ApplyMove.
	Apply(state State) error

	//Copy creates a new move based on this one.
	Copy() Move

	//DefaultsForState should set this move up so that obvious defaults, given
	//the state, are set. For example, for moves that have a
	//TargetPlayerIndex, it makes sense to have this set that to
	//game.CurrentPlayerIndex. Note: this will modify the move!
	DefaultsForState(state State)

	//Name should return the name for this type of move. No other Move structs
	//in use in this game should have the same name, but it should be human-
	//friendly. For example, "Place Token" is a reasonable name, as long as no
	//other types of Move-structs will return that name in this game. Name()
	//should be the same for every Move of the same type.
	Name() string

	//Description is a human-readable sentence describing what the move does.
	//Description should be the same for all moves of the same type, and
	//should not vary with the Move's specific properties. For example, the
	//Description for "Place Token" might be "Places the current user's token
	//in the specified slot on the board."
	Description() string

	ReadSetter() PropertyReadSetter
}
