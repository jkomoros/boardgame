package boardgame

import (
	"encoding/json"
	"errors"
)

//MoveType represents a type of a move in a game, and information about that
//MoveType. New Moves are constructed by calling its NewMove() method. Fields
//are hidden to prevent modifying them once a game has been SetUp.
type MoveType struct {
	name           string
	helpText       string
	constructor    func(*MoveType) Move
	immediateFixUp func(state State) Move
}

//MoveTypeConfig is a collection of information used to create a MoveType.
type MoveTypeConfig struct {
	//Name is the name for this type of move. No other Move structs
	//in use in this game should have the same name, but it should be human-
	//friendly. For example, "Place Token" is a reasonable name, as long as no
	//other types of Move-structs will return that name in this game. Required.
	Name string
	//HelpText is a human-readable sentence describing what the move does.
	//HelpText should be the same for all moves of the same type, and
	//should not vary with the Move's specific properties. For example, the
	//HelpText for "Place Token" might be "Places the current user's token
	//in the specified slot on the board."
	HelpText string
	//Constructor should return a zero-valued Move of the given type. Normally
	//very simple. Required.
	MoveConstructor func(*MoveType) Move

	//If ImmediateFixUp is defined and returns a Move, it will immediately be
	//applied (if Legal) to the game before Delegate's ProposeFixUp is
	//consulted. The move returned need not have been registered with the
	//GameManager via AddFixUpMove, and if the returned move is not legal is
	//fine, it just won't be applied. ImmediateFixUp is useful when you've
	//broken a fixup task into multiple moves only so the observable semantics
	//are granular enough, and saves awkward and error-prone signaling in
	//State fields. When in doubt, just return nil for this method, or do not
	//supply one.
	ImmediateFixUp func(State) Move
}

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications. The Move should be JSON-able (that is, all persistable state
//should be in public fields).
type Move interface {

	//Type is the MoveType that this Move is.
	Type() *MoveType

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

	//Sets the move to have reasonable defaults for the given state.For
	//example, if the Move has a TargetPlayerIndex property, a reasonable
	//default is state.CurrentPlayer().
	DefaultsForState(state State)

	//Description is a human-readable prose description of the effects that
	//this particular move will have, including any move configuration. For
	//example the Description for "Place Token" might be "Player 0 places a
	//token in position 3".
	Description() string

	ReadSetter() PropertyReadSetter
}

//StorageRecordForMove returns a MoveStorageRecord. Can't hang off of Move
//itself since Moves are provided by users of the library.
func StorageRecordForMove(move Move) *MoveStorageRecord {

	blob, err := json.MarshalIndent(move, "", "\t")

	if err != nil {
		return nil
	}

	return &MoveStorageRecord{
		Name: move.Type().Name(),
		Blob: blob,
	}
}

//NewMoveType returns a new MoveType based on the given config.
func NewMoveType(config *MoveTypeConfig) (*MoveType, error) {
	if config == nil {
		return nil, errors.New("No config provided")
	}

	if config.Name == "" {
		return nil, errors.New("No name provided")
	}

	if config.MoveConstructor == nil {
		return nil, errors.New("No MoveConstructor provided")
	}

	return &MoveType{
		name:           config.Name,
		helpText:       config.HelpText,
		constructor:    config.MoveConstructor,
		immediateFixUp: config.ImmediateFixUp,
	}, nil

}

//Name returns the unique name for this type of move.
func (m *MoveType) Name() string {
	return m.name
}

//HelpText is a human-readable sentence describing what the move does.
func (m *MoveType) HelpText() string {
	return m.helpText
}

func (m *MoveType) ImmediateFixUp(state State) Move {
	if m.immediateFixUp == nil {
		return nil
	}
	return m.immediateFixUp(state)
}

//NewMove returns a new move of this type, with defaults set for the given
//state.
func (m *MoveType) NewMove(state State) Move {
	move := m.constructor(m)
	if move == nil {
		return nil
	}
	move.DefaultsForState(state)
	return move
}

//BaseMove is an optional, convenience struct designed to be embedded
//anonymously in your own Moves. It implements no-op methods for many of the
//required methods on Moves (although it can't implement the ones that require
//access to the top level struct, like Copy() and ReadSetter()). Legal and
//Apply are not covered, because every Move should implement their own, and if
//this implemented them it would obscure errors where for example your Legal()
//was incorrectly named and thus not used.
type BaseMove struct {
	MoveType *MoveType
}

func (d *BaseMove) Type() *MoveType {
	return d.MoveType
}

func (d *BaseMove) DefaultsForState(state State) {
	return
}

func (d *BaseMove) Description() string {
	return d.Type().HelpText()
}
