package boardgame

//A MoveFactory takes a state and returns a Move. The state may be nil, in
//which case it should just be an empty (generally, zero-valued) Move of the
//given type. If state is non-nil, the move that is returned should be set to
//reasonable defaults for the given state. For example, if the Move has a
//TargetPlayerIndex property, a reasonable default is state.CurrentPlayer().
type MoveFactory func(state State) Move

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications.
type Move interface {
	//Legal returns nil if this proposed move is legal, or an error if the
	//move is not legal. The error message may be shown directly to the end-
	//user so be sure to make it user friendly. proposer is set to the
	//notional player that is proposing the move. proposer might be a valid
	//player index, or AdminPlayerIndex (for example, if it is a FixUpMove it
	//will typically be AdminPlayerIndex). AdminPlayerIndex is always allowed
	//to make any move. It will never be ObserverPlayerIndex, because by
	//definition Observers may not make moves. If you want to check that the
	//person proposing is able to apply the move for the given player, and
	//that it is their turn, you would do something like test
	//m.TargetPlayerIndex.Equivalent(proposer),
	//m.TargetPlayerIndex.Equivalent(game.CurrentPlayer).
	Legal(state State, proposer PlayerIndex) error

	//Apply applies the move to the state. It is handed a copy of the state to
	//modify. If error is non-nil it will not be applied to the game. It
	//should not be called directly; use Game.ProposeMove.
	Apply(state MutableState) error

	//If ImmediateFixUp returns a Move, it will immediately be applied (if
	//Legal) to the game before Delegate's ProposeFixUp is consulted. The move
	//returned need not have been registered with the GameManager via
	//AddFixUpMove, and if the returned move is not legal is fine, it just
	//won't be applied. ImmediateFixUp is useful when you've broken a fixup
	//task into multiple moves only so the observable semantics are granular
	//enough, and saves awkward and error-prone signaling in State fields.
	//When in doubt, just return nil for this method.
	ImmediateFixUp(state State) Move

	//Name should return the name for this type of move. No other Move structs
	//in use in this game should have the same name, but it should be human-
	//friendly. For example, "Place Token" is a reasonable name, as long as no
	//other types of Move-structs will return that name in this game. Name()
	//should be the same for every Move of the same type, so this method
	//should generally return a constant.
	Name() string

	//Description is a human-readable sentence describing what the move does.
	//Description should be the same for all moves of the same type, and
	//should not vary with the Move's specific properties. For example, the
	//Description for "Place Token" might be "Places the current user's token
	//in the specified slot on the board."
	Description() string

	ReadSetter() PropertyReadSetter
}

//DefaultMove is an optional, convenience struct designed to be embedded
//anonymously in your own Moves. It implements no-op methods for many of the
//required methods on Moves (although it can't implement the ones that require
//access to the top level struct, like Copy() and ReadSetter()). Legal and
//Apply are not covered, because every Move should implement their own, and if
//this implemented them it would obscure errors where for example your Legal()
//was incorrectly named and thus not used.
type DefaultMove struct {
	MoveName        string
	MoveDescription string
}

func (d *DefaultMove) ImmediateFixUp(state State) Move {
	return nil
}

func (d *DefaultMove) Name() string {
	return d.MoveName
}

func (d *DefaultMove) Description() string {
	return d.MoveDescription
}
