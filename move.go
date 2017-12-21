package boardgame

import (
	"encoding/json"
	"github.com/jkomoros/boardgame/errors"
	"time"
)

//MoveType represents a type of a move in a game, and information about that
//MoveType. New Moves are constructed by calling its NewMove() method. Fields
//are hidden to prevent modifying them once a game has been SetUp. New ones
//cannot be created directly; they are created via
//GameManager.AddMoveType(moveTypeConfig).
type MoveType struct {
	name        string
	helpText    string
	constructor func() Move
	legalPhases []int
	isFixUp     bool
	validator   *readerValidator
	manager     *GameManager
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
	//very simple: just a new(MyMoveType). Required. If you have a field that
	//is an enum.Var you'll need to initalize it with a variable for the
	//correct enum (generally necessitating a multi-line constructor),
	//otherwise your move will fail to register. If that field has a struct
	//tag of `enum:"ENUMNAME"` then so long as ENUMNAME is the name of a valid
	//enum in the game's Chest.Enums, it will auto-instantiate a enum.Var for
	//the correct enum for that field, allowing you to maintain a single-line
	//constructor.
	MoveConstructor func() Move

	//LegalPhases is the value that will be returned from
	//MoveType.LegalPhases. It is primarily used by moves.Base to see if the
	//move is legal in the current phase, as determined by
	//delegate.CurrentPhase(). See moves.Base for how this information is
	//used. If you use manager.AddMovesForPhase or
	//manager.AddOrderedMovesForPhase to install your moves, you likely can
	//leave this as nil and have it set correctly. If you want your move to be
	//valid in any phase but don't want them to modify LegalPhases, configure
	//this to be a zero-length slice instead of nil and they'll leave it
	//untouched.
	LegalPhases []int

	//If IsFixUp is true, the moveType will be a FixUp move--that is, players
	//may not propose it, only ProposeFixUp moves may.
	IsFixUp bool
}

//NewMoveType takes a MoveTypeConfig and returns a MoveType associated with
//the given manager. The returned move type will not yet have been added to
//the manager in question. In general you don't call this directly, and
//instead use manager.AddMove, which accepts a MoveTypeConfig.
func (m *MoveTypeConfig) NewMoveType(manager *GameManager) (*MoveType, error) {
	if m == nil {
		return nil, errors.New("No config provided")
	}

	if m.Name == "" {
		return nil, errors.New("No name provided")
	}

	if m.MoveConstructor == nil {
		return nil, errors.New("No MoveConstructor provided")
	}

	exampleMove := m.MoveConstructor()

	if exampleMove == nil {
		return nil, errors.New("MoveConstructor returned nil")
	}

	readSetter := exampleMove.ReadSetter()

	if readSetter == nil {
		return nil, errors.New("MoveConstructor's readsetter returned nil")
	}

	validator, err := newReaderValidator(readSetter, readSetter, exampleMove, moveTypeIllegalPropTypes, manager.Chest(), false)

	if err != nil {
		return nil, errors.New("Couldn't create validator: " + err.Error())
	}

	return &MoveType{
		name:        m.Name,
		helpText:    m.HelpText,
		constructor: m.MoveConstructor,
		isFixUp:     m.IsFixUp,
		legalPhases: m.LegalPhases,
		validator:   validator,
		manager:     manager,
	}, nil

}

//MoveInfo is an object that contains meta-information about a move.
type MoveInfo struct {
	moveType  *MoveType
	version   int
	initiator int
	timestamp time.Time
}

//Move's are how all modifications are made to Game States after
//initialization. Packages define structs that implement Move for all
//modifications. The Move should be JSON-able (that is, all persistable state
//should be in public fields), and it may not include Timers, Stacks, or
//EnumValues. Use moves.Base for a convenient composable base Move that will
//allow you to skip most of the boilerplate overhead.
type Move interface {

	//Info returns the MoveInfo objec that was affiliated with this object by
	//SetInfo.
	Info() *MoveInfo

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

	//SetInfo will be called after the constructor is called to set the
	//information, including what type the move is. Splitting this out allows
	//the basic constructors not in the base classes to be very small, because
	//in most cases you'll compose a moves.Base.
	SetInfo(m *MoveInfo)

	//TopLevelStruct should return the value that was set via
	//SetTopLevelStruct. It returns the Move that is at the top of the
	//embedding chain (because structs that are embedded anonymously can only
	//access themselves and not their embedders). This is useful because in a
	//number of cases embeded moves (for example moves in the moves package)
	//need to consult a method on their embedder.
	TopLevelStruct() Move

	//SetTopLevelStruct is called right after the move is constructed.
	//Splitting this out allows the basic constructors not in the base classes
	//to be very small, because in most cases you'll compose a moves.Base.
	SetTopLevelStruct(m Move)

	//ValidConfiguration will be checked when the game manager is SetUp, and
	//if it returns an error the manager will fail to SetUp. Some Moves,
	//especially sub-classes of moves in the moves package, require set up
	//that can only be verified at run time (for example, verifying that the
	//embedder implements a certain inteface). This is a useful way to detect
	//those misconfigurations at the earliest moment. In most cases you never
	//need to implement this yourself; moves in the moves package that need it
	//will implement it.
	ValidConfiguration(exampleState MutableState) error

	ReadSetConfigurer
}

//StorageRecordForMove returns a MoveStorageRecord. Can't hang off of Move
//itself since Moves are provided by users of the library.
func StorageRecordForMove(move Move, currentPhase int) *MoveStorageRecord {

	blob, err := json.MarshalIndent(move, "", "\t")

	if err != nil {
		return nil
	}

	return &MoveStorageRecord{
		Name:      move.Info().Type().Name(),
		Version:   move.Info().version,
		Initiator: move.Info().initiator,
		Timestamp: move.Info().timestamp,
		Phase:     currentPhase,
		Blob:      blob,
	}
}

//Type returns the MoveType of this Move.
func (m *MoveInfo) Type() *MoveType {
	return m.moveType
}

//Version returns the version of this move--or the version that it will be
//when successfully committed.
func (m *MoveInfo) Version() int {
	return m.version
}

//Timestamp returns the time that the given move was made.
func (m *MoveInfo) Timestamp() time.Time {
	return m.timestamp
}

//Initiator returns the move version that initiated this causal chain: the
//PlayerMove that was applied that led to this chain of FixUp moves. The
//Initiator of a PlayerMove is its own version, so this value will be less
//than or equal to its own version. The value of Initator is unspecified until
//after the move has been successfully committed.
func (m *MoveInfo) Initiator() int {
	return m.initiator
}

var moveTypeIllegalPropTypes = map[PropertyType]bool{
	TypeStack: true,
	TypeTimer: true,
}

//Name returns the unique name for this type of move.
func (m *MoveType) Name() string {
	return m.name
}

//HelpText is a human-readable sentence describing what the move does.
func (m *MoveType) HelpText() string {
	return m.helpText
}

func (m *MoveType) IsFixUp() bool {
	return m.isFixUp
}

func (m *MoveType) LegalPhases() []int {
	return m.legalPhases
}

//NewMove returns a new move of this type, with defaults set for the given
//state. If state is nil, then DefaultsForState will not be called.
func (m *MoveType) NewMove(state State) Move {
	move := m.constructor()
	if move == nil {
		return nil
	}

	info := &MoveInfo{
		moveType: m,
		version:  state.Version() + 1,
	}

	move.SetInfo(info)
	move.SetTopLevelStruct(move)

	readSetConfigurer := move.ReadSetConfigurer()

	if readSetConfigurer == nil {
		//This shouldn't happen because we verified that ReadSetter returned
		//non-nil when the movetype was registered.
		m.manager.Logger().Error("ReadSetConfigurer for move unexpectedly returned nil")
		return nil
	}

	if err := m.validator.AutoInflate(readSetConfigurer, state); err != nil {
		m.manager.Logger().Error("AutoInflate had an error: " + err.Error())
		return nil
	}

	if err := m.validator.Valid(readSetConfigurer); err != nil {
		m.manager.Logger().Error("Move was not valid: " + err.Error())
		return nil
	}

	if state != nil {
		move.DefaultsForState(state)
	}
	return move
}

//We implement a private stub of moves.Base in this package just for the
//convience of our own test structs.
type baseMove struct {
	info           MoveInfo
	topLevelStruct Move
}

func (d *baseMove) SetInfo(m *MoveInfo) {
	d.info = *m
}

func (d *baseMove) Info() *MoveInfo {
	return &d.info
}

func (d *baseMove) SetTopLevelStruct(m Move) {
	d.topLevelStruct = m
}

func (d *baseMove) TopLevelStruct() Move {
	return d.topLevelStruct
}

//DefaultsForState doesn't do anything
func (d *baseMove) DefaultsForState(state State) {
	return
}

//Description defaults to returning the Type's HelpText()
func (d *baseMove) Description() string {
	return d.Info().Type().HelpText()
}

func (d *baseMove) ValidConfiguration(exampleState MutableState) error {
	return nil
}
