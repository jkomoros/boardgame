package boardgame

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
)

// MoveType represents a type of a move in a game, and information about that
// MoveType. New Moves are constructed by calling its NewMove() method. Fields
// are hidden to prevent modifying them once a game has been SetUp. New ones
// cannot be created directly; they are created via
// GameManager.AddMoveType(moveTypeConfig).
type moveType struct {
	name                string
	constructor         func() Move
	validator           *StructInflater
	customConfiguration PropertyCollection
	nameSanitization    map[string]Policy
	hiddenAnimationKey  string
	manager             *GameManager
}

// MoveConfig is a collection of information used to create a Move. Your
// delegate's ConfigureMoves() will emit a slice of them to define which moves
// are valid for your game. Typically you'll use moves.Combine, moves.Add,
// moves.AddWithPhase, combined with moves.AutoConfigurer.Configure() to
// generate these. This is an interface and not a concrete struct because other
// packages, like moves, add more behavior to the ones they return. If you want
// just a vanilla one without using the moves package, use NewMoveConfig.
type MoveConfig interface {
	//Name is the name for this type of move. No other Move structs
	//in use in this game should have the same name, but it should be human-
	//friendly. For example, "Place Token" is a reasonable name, as long as no
	//other types of Move-structs will return that name in this game. Required.
	Name() string

	//Constructor should return a non-nil pointer to a zero-valued Move of the
	//given type. Normally this is simply new(MyMoveType). Required. Once the
	//engine initializes a Move it is an identity object and must not be copied
	//by value. The moves you create may not have fields of Stack, Board, or Timer
	//type, but may have enum.Val
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

// NewMoveConfig returns a simple MoveConfig that will return the provided
// parameters from its getters. Typically you don't use this, but rather use
// the output of moves.AutoConfigurer.Config().
func NewMoveConfig(name string, constructor func() Move, customConfiguration PropertyCollection) MoveConfig {
	return defaultMoveConfig{
		name,
		constructor,
		customConfiguration,
	}
}

const newMoveTypeErrNoManagerPassed = "No manager passed, so we can'd do validation"

// NewMoveType takes a MoveConfig and returns a MoveType associated with
// the given manager. The returned move type will not yet have been added to
// the manager in question. In general you don't call this directly, and
// instead use manager.AddMove, which accepts a MoveConfig.
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
	exampleValue := reflect.ValueOf(exampleMove)
	if exampleValue.Kind() != reflect.Ptr || exampleValue.IsNil() {
		return nil, errors.New("Constructor must return a non-nil pointer to a Move")
	}

	// MoveInfo is the single runtime-affiliation channel for a Move. Validate
	// the constructor's SetInfo implementation while setup can still return a
	// useful configuration error rather than allowing broken dispatch later.
	// SetInfo may retain the pointer or copy the value; behavior, not pointer
	// identity, is the contract.
	probeInfo := &MoveInfo{runtime: moveRuntime{concreteMove: exampleMove}}
	exampleMove.SetInfo(probeInfo)
	if exampleMove.Info() == nil || exampleMove.Info().ConcreteMove() != exampleMove {
		return nil, errors.New("Move.SetInfo must preserve the MoveInfo affiliation provided by the engine")
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
		validator, err = newStructInflater(exampleMove, moveTypeIllegalPropTypes, manager.Chest(), PolicyHidden)

		if err != nil {
			return nil, errors.New("Couldn't create validator: " + err.Error())
		}
	} else {
		//moves.DefaultConfig hackily looks for exactly this error string.
		err = errors.New(newMoveTypeErrNoManagerPassed)
	}

	nameSanitization, hiddenAnimationKey, configErr := configuredMoveNameSanitization(config.CustomConfiguration())
	if configErr != nil {
		return nil, configErr
	}
	if manager != nil {
		for propName, policies := range validator.sanitizationPolicy {
			for groupName, policy := range policies {
				if policy == PolicyInvalid {
					return nil, errors.New("Move " + config.Name() + " property " + propName + " had invalid sanitization policy for group " + groupName)
				}
			}
		}
		groupEnum := manager.Delegate().GroupEnum()
		groupNames := validator.sanitizationPolicyGroupNames(groupEnum)
		for groupName := range nameSanitization {
			if groupName == SanitizationDefaultGroup {
				continue
			}
			if groupEnum != nil && groupEnum.ValueFromString(groupName) != enum.IllegalValue {
				continue
			}
			groupNames[groupName] = true
		}
		for groupName := range groupNames {
			if groupName == sanitizationGroupSelf || groupName == sanitizationGroupOther {
				continue
			}
			if _, err := manager.Delegate().ComputedPlayerGroupMembership(groupName, nil, nil); err != nil {
				return nil, errors.New("Move " + config.Name() + " had illegal sanitization group " + groupName + ": " + err.Error())
			}
		}
	}

	return &moveType{
		name:                config.Name(),
		constructor:         config.Constructor(),
		customConfiguration: config.CustomConfiguration(),
		validator:           validator,
		nameSanitization:    nameSanitization,
		hiddenAnimationKey:  hiddenAnimationKey,
		manager:             manager,
	}, err

}

// NewOrphanMove constructs a Move with its engine-owned MoveInfo affiliation,
// but without attaching it to a GameManager or state. It is the supported
// advanced seam for reusable-framework tests and tools that need a standalone
// move whose embedded behavior can dispatch to the final concrete move.
//
// Manager-backed inflation, configuration validation, and state-dependent
// defaults are deliberately not run. Normal game code should obtain moves from
// a Game or GameManager instead.
func NewOrphanMove(config MoveConfig) (Move, error) {
	throwAwayMoveType, err := newMoveType(config, nil)

	if err != nil {
		// A managerless move type is exactly what this constructor needs. All
		// other errors describe an invalid MoveConfig or Move implementation.
		if err.Error() != newMoveTypeErrNoManagerPassed {
			return nil, errors.New("couldn't create orphan move type: " + err.Error())
		}
	}
	move := throwAwayMoveType.NewMove(nil)
	if move == nil {
		return nil, errors.New("couldn't construct orphan move with valid runtime affiliation")
	}
	return move, nil
}

// OrphanExampleMove is retained for callers using the ManagerInternals seam.
// New code that does not otherwise need a manager should use NewOrphanMove.
func (m *ManagerInternals) OrphanExampleMove(config MoveConfig) (Move, error) {
	return NewOrphanMove(config)
}

// moveRuntime contains the engine-assigned runtime affiliation for one
// constructed move. Consumers must treat it as immutable. It is deliberately
// separate from the descriptive fields on MoveInfo: runtime affiliation must
// not grow into a general-purpose execution context or service locator.
type moveRuntime struct {
	concreteMove Move
	moveType     *moveType
}

// MoveInfo contains descriptive information and engine-assigned affiliation
// for one constructed move. It is fetched via move.Info(). A MoveInfo belongs
// to exactly one Move instance and must not be reused between moves.
type MoveInfo struct {
	runtime   moveRuntime
	version   int
	initiator int
	name      string
	timestamp time.Time
}

// Move is the struct that are how all modifications are made to States after
// initialization. Packages define structs that implement different Moves for all
// types of valid modifications. Moves are objects your own packages will
// returen. Use base.Move or moves.Default for a convenient composable base Move
// that will allow you to skip most of the boilerplate overhead. Your Move is
// similar to a SubState in that all of the persistable properties must be one of
// the enumerated types in PropertyType, excluding a few types. Your Moves are
// installed based on what your GameDelegate returns from ConfigureMoves(). See
// MoveConfig for more about things that must be true about structs you return.
// The two primary methods for your game logic are Legal() and Apply().
type Move interface {

	//Legal returns nil if this proposed move is legal at the given state, or
	//an error if the move is not legal. The error message may be shown
	//directly to the end-user so be sure to make it user friendly. proposer
	//is set to the notional player that is proposing the move. proposer might
	//be a valid player index, or AdminPlayerIndex (for example, if it is a
	//FixUpMove it will typically be AdminPlayerIndex). AdminPlayerIndex is
	//always allowed to make any move. It will never be ObserverPlayerIndex,
	//because by definition Observers may not make moves. Note that during
	//simultaneous phases where CurrentPlayerIndex() returns AnyPlayerIndex,
	//the proposer will still be the actual player index (0, 1, 2, ...); the
	//Equivalent() method treats AnyPlayerIndex as a wildcard, so
	//m.TargetPlayerIndex.Equivalent(currentPlayer) will return true when
	//currentPlayer is AnyPlayerIndex. If you want to check that the person
	//proposing is able to apply the move for the given player, and that it is
	//their turn, you would do something like test
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

	//SetInfo will be called after the constructor is called to affiliate the
	//MoveInfo for this exact move instance. Implementations may retain the
	//pointer or a value copy, but Info must preserve all information provided.
	//Initialized moves are engine-owned identity objects and must not themselves
	//be copied by value.
	SetInfo(m *MoveInfo)

	//Moves alos have a ValidConfiguration, because moves, especially
	//sub-classes of the moves package, require set-up that can only be verified
	//at run time (for example, verifying that the embedder implements a certain
	//inteface)
	ConfigurationValidator

	//Moves, like ConfigurableSubStates, must only have all of their
	//important, persistable properties available to be inspected and modified
	//via a PropertyReadSetConfigurer. The game engine will use that interface
	//to create new moves, inflate old moves from storage, and copy moves.
	//Typically you generate this automatically for your moves with `boargame-
	//util codegen`.
	ReadSetConfigurer
}

// configuredMoveStateApplier is implemented by complete framework move
// behaviors that contribute a configured state effect in addition to the
// concrete move's Apply method. Its exported method can be promoted across the
// moves package boundary while the hook itself remains an engine detail.
type configuredMoveStateApplier interface {
	ApplyConfiguredMoveState(State) error
}

// ConfigurationValidator is an interface that certain types must implement.
// These will be called typically during NewGameManager set up, and are an
// opportunity for the structs to report configuration errors that can only be
// discovered at runtime. If an error is reported then NewGameManager will fail,
// which means that the misconfiguration can be detected early almost
// guaranteeing it will be detected by the game package author. For example, many
// moves in the moves package must be embedded in structs that contain particular
// methods in the embedding struct, and that can only be validated at runtime.
// Typically you don't need to implement this yourself; the types of structs that
// have it will have a stub implementation in the base package, and the primary
// beneficiaries of this are more complex embeddable library structs like those
// found in the moves package.
type ConfigurationValidator interface {
	//ValidConfiguration will be checked when the NewGameManager is being set
	//up, and if it returns an error the manager will fail to be created.
	ValidConfiguration(exampleState State) error
}

// StorageRecordForMove returns a MoveStorageRecord. Can't hang off of Move
// itself since Moves are provided by users of the library.
func StorageRecordForMove(move Move, currentPhase enum.EnumKey, proposer PlayerIndex) *MoveStorageRecord {

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

// Name returns the name of the move type that this move is, based on the value
// passed in the affiliated MoveConfig from your GameDelegate.ConfigureMoves().
// Calling manager.ExampleMove() with that string value will return a similar
// struct.
func (m *MoveInfo) Name() string {
	return m.name
}

// ConcreteMove returns the exact Move instance produced by the constructor
// and affiliated with this MoveInfo by the engine. Embedded framework moves
// use it to dispatch to methods and capabilities implemented by the final
// composed move. Ordinary game moves normally do not need to call it.
//
// ConcreteMove returns nil when called on a nil MoveInfo. It is available on
// every engine-created move before inflation, configuration validation, and
// defaults are run.
func (m *MoveInfo) ConcreteMove() Move {
	if m == nil {
		return nil
	}
	return m.runtime.concreteMove
}

// Version returns the version of this move--or the version that it will be
// when successfully committed.
func (m *MoveInfo) Version() int {
	return m.version
}

// Timestamp returns the time that the given move was made.
func (m *MoveInfo) Timestamp() time.Time {
	return m.timestamp
}

// CustomConfiguration returns the configuration object associated with this
// move when it was installed from its MoveConfig.CustomConfiguration().
func (m *MoveInfo) CustomConfiguration() PropertyCollection {
	if m.runtime.moveType == nil {
		return nil
	}
	return m.runtime.moveType.customConfiguration
}

// SanitizationPolicy resolves the move property's sanitize tag for the given
// memberships. Move properties are hidden unless explicitly configured.
func (m *MoveInfo) SanitizationPolicy(propName string, groupMembership map[string]bool) Policy {
	if m == nil || m.runtime.moveType == nil || m.runtime.moveType.validator == nil {
		return PolicyInvalid
	}
	return ResolveSanitizationPolicy(m.runtime.moveType.validator.PropertySanitizationPolicy(propName), groupMembership, PolicyHidden)
}

// Initiator returns the move version that initiated this causal chain: the
// player Move that was applied that led to this chain of fix up moves as
// proposed by GameDelegate.ProposeFixUpMove. The Initiator of a PlayerMove is
// its own version, so this value will be less than or equal to its own
// version. The value of Initator is unspecified until after the move has been
// successfully committed.
func (m *MoveInfo) Initiator() int {
	return m.initiator
}

var moveTypeIllegalPropTypes = map[PropertyType]bool{
	TypeStack: true,
	TypeBoard: true,
	TypeTimer: true,
}

// Name returns the unique name for this type of move.
func (m *moveType) Name() string {
	return m.name
}

// NewMove returns a new move of this type, with defaults set for the given
// state. If state is nil, then DefaultsForState will not be called.
func (m *moveType) NewMove(state ImmutableState) Move {
	move := m.constructor()
	if move == nil {
		return nil
	}
	moveValue := reflect.ValueOf(move)
	if moveValue.Kind() != reflect.Ptr || moveValue.IsNil() {
		return nil
	}

	info := &MoveInfo{
		runtime: moveRuntime{
			concreteMove: move,
			moveType:     m,
		},
		name: m.Name(),
	}

	if state != nil {
		info.version = state.Version() + 1
	}

	move.SetInfo(info)
	installedInfo := move.Info()
	if installedInfo == nil || installedInfo.ConcreteMove() != move || installedInfo.runtime.moveType != m {
		if m.manager != nil {
			m.manager.Logger().Error("Move.SetInfo did not preserve the MoveInfo affiliation provided by the engine")
		}
		return nil
	}

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

// We implement a private stub of base.Move in this package just for the
// convience of our own test structs.
type baseMove struct {
	info *MoveInfo
}

// baseFixUpMove is same as baseMove but returns true for IsFixUp.
type baseFixUpMove struct {
	baseMove
}

func (d *baseMove) HelpText() string {
	return "Unimplemented"
}

func (d *baseMove) SetInfo(m *MoveInfo) {
	d.info = m
}

func (d *baseMove) Info() *MoveInfo {
	return d.info
}

func (d *baseMove) IsFixUp() bool {
	return false
}

func (d *baseFixUpMove) IsFixUp() bool {
	return true
}

// DefaultsForState doesn't do anything
func (d *baseMove) DefaultsForState(state ImmutableState) {
	return
}

// Description defaults to returning the Type's HelpText()
func (d *baseMove) Description() string {
	return d.Info().ConcreteMove().HelpText()
}

func (d *baseMove) ValidConfiguration(exampleState State) error {
	return nil
}
