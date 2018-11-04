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
type moveType struct {
	name                string
	constructor         func() Move
	validator           *StructInflater
	customConfiguration PropertyCollection
	manager             *GameManager
}

//MoveConfig is a collection of information used to create a Move. Your
//delegate's ConfigureMoves() will emit a slice of them. Typically you'll use
//moves.Combine, moves.Add, moves.AddWithPhase, combined with
//moves.AutoConfigurer.Configure() to generate these. This is an interface and
//not a concrete struct because other packages, like moves, add more behavior
//to the ones they return. If you want just a vanilla one without using the
//moves package, use NewMoveConfig.
type MoveConfig interface {
	//Name is the name for this type of move. No other Move structs
	//in use in this game should have the same name, but it should be human-
	//friendly. For example, "Place Token" is a reasonable name, as long as no
	//other types of Move-structs will return that name in this game. Required.
	Name() string

	//Constructor should return a zero-valued Move of the given type. Normally
	//very simple: just a new(MyMoveType). Required. The moves you create may
	//not have an fields of Stack, Board, or Timer type, but may have enum.Val
	//type. Those fields must be non-nil; like delegate.GameStateConstructor
	//and others, a StructInflater will be created for each move type, which
	//allows you to provide inflation configuration via struct tags. See
	//StructInflater for more. Like ConfigurableSubState, all of the
	//properties to persist must be accessible via their ReadSetConfigurer, as
	//that is how the core engine serializes them, re-inflates them from
	//storage, and copies them.
	Constructor() func() Move

	//CustomConfiguration is an optional PropertyCollection. Some move types--
	//especially in the `moves` package--stash configuration options here that
	//will change how all moves of this type behave. Individual moves would
	//reach through via Info().CustomConfiguration() to retrieve the
	//values stored there. Different move types will store different types of
	//information there--to avoid a collision the convention is to use a
	//string name that starts with your fully qualified package name, then a
	//dot, then the propertyname, like so:
	//"github.com/jkomoros/boardgame/moves.MoveName". Those strings are often
	//encoded as package-private constants, and a
	//interfaces.CustomConfigurationOption functor factory is provided to set
	//those from outside the package. Generally you don't use this directly,
	//but moves.AutoConfigurer will help you set these for what specific
	//moves in that package expect.
	CustomConfiguration() PropertyCollection
}

type defaultMoveConfig struct {
	name                string
	constructor         func() Move
	customConfiguration PropertyCollection
}

func (d defaultMoveConfig) Name() string {
	return d.name
}

func (d defaultMoveConfig) Constructor() func() Move {
	return d.constructor
}

func (d defaultMoveConfig) CustomConfiguration() PropertyCollection {
	return d.customConfiguration
}

//NewMoveConfig returns a simple MoveConfig that will return the provided
//parameters from its getters. Typically you don't use this, but rather use
//the output of moves.AutoConfigurer.Config().
func NewMoveConfig(name string, constructor func() Move, customConfiguration PropertyCollection) MoveConfig {
	return defaultMoveConfig{
		name,
		constructor,
		customConfiguration,
	}
}

const newMoveTypeErrNoManagerPassed = "No manager passed, so we can'd do validation"

//NewMoveType takes a MoveConfig and returns a MoveType associated with
//the given manager. The returned move type will not yet have been added to
//the manager in question. In general you don't call this directly, and
//instead use manager.AddMove, which accepts a MoveConfig.
func newMoveType(config MoveConfig, manager *GameManager) (*moveType, error) {
	if config == nil {
		return nil, errors.New("No config provided")
	}

	if config.Name() == "" {
		return nil, errors.New("No name provided")
	}

	if config.Constructor() == nil {
		return nil, errors.New("No MoveConstructor provided")
	}

	exampleMove := config.Constructor()()

	if exampleMove == nil {
		return nil, errors.New("Constructor returned nil")
	}

	readSetter := exampleMove.ReadSetter()

	if readSetter == nil {
		return nil, errors.New("Constructor's readsetter returned nil")
	}

	var validator *StructInflater
	var err error

	//moves.Defaultconfig will call this without a manager. So return a half-
	//useful object in that case... but also an error so anyone else who
	//checks the error will ignore the half-useful move type.
	if manager != nil {
		validator, err = NewStructInflater(exampleMove, moveTypeIllegalPropTypes, manager.Chest())

		if err != nil {
			return nil, errors.New("Couldn't create validator: " + err.Error())
		}
	} else {
		//moves.DefaultConfig hackily looks for exactly this error string.
		err = errors.New(newMoveTypeErrNoManagerPassed)
	}

	return &moveType{
		name:                config.Name(),
		constructor:         config.Constructor(),
		customConfiguration: config.CustomConfiguration(),
		validator:           validator,
		manager:             manager,
	}, err

}

//OrphanExampleMove returns a move from the config that will be similar to a
//real move, in terms of struct-based auto-inflation, etc. This is exposed
//primarily for moves.AutoConfigrer, and generally shouldn't be used by
//others.
func (m *ManagerInternals) OrphanExampleMove(config MoveConfig) (Move, error) {
	throwAwayMoveType, err := newMoveType(config, nil)

	if err != nil {
		//Look for exatly the single kind of error we're OK with. Yes, this is a hack.
		if err.Error() != newMoveTypeErrNoManagerPassed {
			return nil, errors.New("Couldn't create intermediate move type: " + err.Error())
		}
	}
	return throwAwayMoveType.NewMove(nil), nil
}

//MoveInfo is an object that contains meta-information about a move.
type MoveInfo struct {
	moveType  *moveType
	version   int
	initiator int
	name      string
	timestamp time.Time
}

//Move's are how all modifications are made to States after initialization.
//Packages define structs that implement different Moves for all types of
//valid modifications. Moves are objects your own packages will returen. Use
//base.Move or moves.Default for a convenient composable base Move that will
//allow you to skip most of the boilerplate overhead. Your Move is similar to
//a SubState in that all of the persistable properties must be one of the
//enumerated types in PropertyType, excluding a few types. Your Moves are
//installed based on what your GameDelegate returns from ConfigureMoves(). See
//MoveConfig for more about things that must be true about structs you return.
//The two primary methods for your game logic are Legal() and Apply().
type Move interface {

	//Legal returns nil if this proposed move is legal at the given state, or
	//an error if the move is not legal. The error message may be shown
	//directly to the end- user so be sure to make it user friendly. proposer
	//is set to the notional player that is proposing the move. proposer might
	//be a valid player index, or AdminPlayerIndex (for example, if it is a
	//FixUpMove it will typically be AdminPlayerIndex). AdminPlayerIndex is
	//always allowed to make any move. It will never be ObserverPlayerIndex,
	//because by definition Observers may not make moves. If you want to check
	//that the person proposing is able to apply the move for the given
	//player, and that it is their turn, you would do something like test
	//m.TargetPlayerIndex.Equivalent(proposer),
	//m.TargetPlayerIndex.Equivalent(game.CurrentPlayer). Legal is one of the
	//most key parts of logic for your game type. It is important for fix up
	//moves in particular to have carefully-designed Legal() methods, as the
	//ProposeFixUpMove on base.GameDelegate (which you almost always use)
	//walks through each move and returns the first one that is legal at this
	//game state--so if one of your moves is erroneously legal more often than
	//it should be it could be mistakenly applied, perhaps in an infinite
	//loop!
	Legal(state ImmutableState, proposer PlayerIndex) error

	//Apply applies the move to the state by modifying hte right properties.
	//It is handed a copy of the state to modify. If error is non-nil it will
	//not be applied to the game. It should not be called directly; use
	//Game.ProposeMove. Legal() will have been called before and returned nil.
	//Apply is the only place (outside of some of the Game initalization logic
	//on GameDelegate) where you are allowed to modify the state direclty and
	//are passed a State, not an ImmutableState.
	Apply(state State) error

	//All of the methods below this point are typically provided by base.Move
	//and not necessary to be modified.

	//Sets the move to have reasonable defaults for the given state.For
	//example, if the Move has a TargetPlayerIndex property, a reasonable
	//default is state.CurrentPlayer(). DefaultsForState is used to set
	//reasonable defaults for fix up moves. Typically you can skip this.
	DefaultsForState(state ImmutableState)

	//HelpText is a human-readable sentence describing what the move does in
	//general. HelpText should be the same for all moves of the same type, and
	//should not vary with the Move's specific properties. For example, the
	//HelpText for "Place Token" might be "Places the current user's token in
	//the specified slot on the board." Primarily useful just to show to a
	//user in an interface.
	HelpText() string

	//Info returns the MoveInfo object that was affiliated with this object by
	//SetInfo. It includes information about when the move was applied, the
	//name of the move, and other information.
	Info() *MoveInfo

	//SetInfo will be called after the constructor is called to set the
	//information, including what type the move is.
	SetInfo(m *MoveInfo)

	//TopLevelStruct should return the value that was set via
	//SetTopLevelStruct. It returns the Move that is at the top of the
	//embedding chain (because structs that are embedded anonymously can only
	//access themselves and not their embedders). This is useful because in a
	//number of cases embedded moves (for example moves in the moves package)
	//need to consult a method on their embedder to see if any of their
	//behavior should be overridden.
	TopLevelStruct() Move

	//SetTopLevelStruct is called right after the move is constructed, with
	//the top-level struct. This should be returned from TopLevelStruct.
	SetTopLevelStruct(m Move)

	//ValidConfiguration will be checked when the NewGameManager is being set
	//up, and if it returns an error the manager will fail to be created. Some
	//Moves, especially sub-classes of moves in the moves package, require set
	//up that can only be verified at run time (for example, verifying that
	//the embedder implements a certain inteface). This is a useful way to
	//detect those misconfigurations at the earliest moment. In most cases you
	//never need to implement this yourself; moves in the moves package that
	//need it will implement it.
	ValidConfiguration(exampleState State) error

	//Moves, like ConfigurableSubStates, must only have all of their
	//important, persistable properties available to be inspected and modified
	//via a PropertyReadSetConfigurer. The game engine will use that interface
	//to create new moves, inflate old moves from storage, and copy moves.
	//Typically you generate this automatically for your moves with `boargame-
	//util codegen`.
	ReadSetConfigurer
}

//StorageRecordForMove returns a MoveStorageRecord. Can't hang off of Move
//itself since Moves are provided by users of the library.
func StorageRecordForMove(move Move, currentPhase int, proposer PlayerIndex) *MoveStorageRecord {

	blob, err := json.MarshalIndent(move, "", "\t")

	if err != nil {
		return nil
	}

	return &MoveStorageRecord{
		Name:      move.Info().Name(),
		Version:   move.Info().version,
		Initiator: move.Info().initiator,
		Timestamp: move.Info().timestamp,
		Phase:     currentPhase,
		Proposer:  proposer,
		Blob:      blob,
	}
}

//Name returns the name of the move type that this move is. Calling
//manager.ExampleMove() with that string value will return a similar struct.
func (m *MoveInfo) Name() string {
	return m.name
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

//CustomConfiguration returns the configuration object associated with this
//move when it was installed from its MoveConfig.CustomConfiguration().
func (m *MoveInfo) CustomConfiguration() PropertyCollection {
	if m.moveType == nil {
		return nil
	}
	return m.moveType.customConfiguration
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
	TypeBoard: true,
	TypeTimer: true,
}

//Name returns the unique name for this type of move.
func (m *moveType) Name() string {
	return m.name
}

//NewMove returns a new move of this type, with defaults set for the given
//state. If state is nil, then DefaultsForState will not be called.
func (m *moveType) NewMove(state ImmutableState) Move {
	move := m.constructor()
	if move == nil {
		return nil
	}

	info := &MoveInfo{
		moveType: m,
		name:     m.Name(),
	}

	if state != nil {
		info.version = state.Version() + 1
	}

	move.SetInfo(info)
	move.SetTopLevelStruct(move)

	//validator might be nil if we have a half-functioning MoveType. (Like
	//what will be returned, along with an error, when NewMoveType is called
	//during moves.DefaultConfig)
	if m.validator != nil {
		if err := m.validator.Inflate(move, state); err != nil {
			m.manager.Logger().Error("AutoInflate had an error: " + err.Error())
			return nil
		}

		if err := m.validator.Valid(move); err != nil {
			m.manager.Logger().Error("Move was not valid: " + err.Error())
			return nil
		}
	}

	if state != nil {
		move.DefaultsForState(state)
	}
	return move
}

//We implement a private stub of base.Move in this package just for the
//convience of our own test structs.
type baseMove struct {
	info           MoveInfo
	topLevelStruct Move
}

//baseFixUpMove is same as baseMove but returns true for IsFixUp.
type baseFixUpMove struct {
	baseMove
}

func (d *baseMove) HelpText() string {
	return "Unimplemented"
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

func (d *baseMove) IsFixUp() bool {
	return false
}

func (d *baseFixUpMove) IsFixUp() bool {
	return true
}

//DefaultsForState doesn't do anything
func (d *baseMove) DefaultsForState(state ImmutableState) {
	return
}

//Description defaults to returning the Type's HelpText()
func (d *baseMove) Description() string {
	return d.TopLevelStruct().HelpText()
}

func (d *baseMove) ValidConfiguration(exampleState State) error {
	return nil
}
