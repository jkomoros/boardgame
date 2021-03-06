package boardgame

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jkomoros/boardgame/errors"
	"github.com/sirupsen/logrus"
)

//GameManager defines the logic and machinery for a given game type. It is the
//core game object that represents a game type in the engine. GameManager is a
//struct provided by the core engine; think of your GameDelegate as the game-
//type specific brain that will be plugged into the generic GameManager to
//imbue it with life. GameManagers manage fetching specific games from
//storage, proposing moves, and other lifecycle methods. All games of a
//certain type on a certain computer use the same GameManager.
type GameManager struct {
	delegate                  GameDelegate
	gameValidator             *StructInflater
	playerValidator           *StructInflater
	dynamicComponentValidator map[string]*StructInflater
	chest                     *ComponentChest
	storage                   StorageManager
	agents                    []Agent
	moves                     []*moveType
	movesByName               map[string]*moveType
	agentsByName              map[string]Agent
	modifiableGamesLock       sync.RWMutex
	modifiableGames           map[string]*Game
	timers                    *timerManager
	initialized               bool
	logger                    *logrus.Logger
	variantConfig             VariantConfig
}

//Internals returns a ManagerInternals for this manager. All of the methods on
//a ManagerInternals are designed to be used only in very specific conditions;
//users of this package should almost never do anything with these
func (g *GameManager) Internals() *ManagerInternals {
	return &ManagerInternals{
		g,
	}
}

//ManagerInternals is a special struct that has debug-only methods hanging off
//of it. Some methods need to be exposed outside of the package due to how the
//sub-packages are organized. All of the methods off of this object are
//designed to be used only by other sub-packages, and should be used at your
//own risk.
type ManagerInternals struct {
	manager *GameManager
}

//RecreateGame creates a new game that has the same properties as the provided
//GameStorageRecord. It is very rarely what you want; see NewGame(), Game(),
//and ModifiableGame(). RecreateGame is most useful in debugging or testing
//scenarios where you want a game to have the same ID and SecretSalt as a
//previously created game, so the moves can be applied deterministically with
//the same input. rec generally should be a GameStorageRecord representing a
//game that was created in a different storage pool; if a game with that ID
//already exists in this storage pool RecreateGame will error.
func (m *ManagerInternals) RecreateGame(rec *GameStorageRecord) (*Game, error) {
	return m.manager.recreateGame(rec)
}

//ForceNextTimer forces the next timer to fire even if it's not supposed to
//fire yet. Will return true if there was a timer that was fired, false
//otherwise.
func (m *ManagerInternals) ForceNextTimer() bool {
	return m.manager.timers.ForceNextTimer()
}

//ForceFixUp forces the engine to check if a FixUp move applies, even if no
//player move is waiting to apply. Typically moves are only legal based on the
//state, so if a move hasn't been applied they can't be legal. But in some
//cases, like for server seating players, there's some outside state that might
//have changed that could cause a move to be legal even though the game state
//didn't change.
func (m *ManagerInternals) ForceFixUp(game *Game) DelayedError {
	if game == nil {
		delayed := make(DelayedError, 1)
		delayed <- nil
		return delayed
	}
	return game.triggerFixUp()
}

//AddCommittedCallback adds a function that will be called once the state is
//successfully saved. Typically you'd do something with this in your Move's
//Apply() method if you wanted to note in some external system whether the move
//had actually been successfully committed or not. Will be called back
//immediately after the state is successfully saved to the database and before
//any other fixup moves are called. This is only designed to be called from
//within a move's Apply function.
func (m *ManagerInternals) AddCommittedCallback(st State, callback func()) error {
	s, ok := st.(*state)
	if !ok {
		return errors.New("The State was not the expected type of underlying object")
	}
	s.AddCommittedCallback(callback)
	return nil
}

//StructInflater returns the autp-created StructInflater for the given type of
//property in your state, allowing you to retrieve the inflater in use to
//inspect for e.g. SanitizationPolicy configuration. Typically you don't use
//this directly--it's primarily provided for base.GameDelegate to use.
func (m *ManagerInternals) StructInflater(propRef StatePropertyRef) *StructInflater {

	manager := m.manager

	var validator *StructInflater
	switch propRef.Group {
	case StateGroupGame:
		validator = manager.gameValidator
	case StateGroupPlayer:
		validator = manager.playerValidator
	case StateGroupDynamicComponentValues:
		validator = manager.dynamicComponentValidator[propRef.DeckName]
	}

	return validator

}

//InflateMoveStorageRecord takes a move storage record and turns it into a
//move associated with that game, if possible. Returns nil if not possible.
//You rarely need this; it's exposed primarily for the use of boardgame
///boardgame-util/lib/golden.
func (m *ManagerInternals) InflateMoveStorageRecord(rec *MoveStorageRecord, game *Game) (Move, error) {
	if rec == nil {
		return nil, nil
	}
	return rec.inflate(game)
}

const baseLibraryName = "github.com/jkomoros/boardgame"

//gamePkgMatchesDelegateName checks, via reflection, that the delegate's
//package name is the same as the delegate.Name(), because many systems assume
//that is the case. If the delegate comes from this package (i.e. it's a test)
//then it returns nil.
func gamePkgMatchesDelegateName(delegate GameDelegate) error {

	path := reflect.ValueOf(delegate).Elem().Type().PkgPath()

	if path == baseLibraryName {
		//Delegates in this package are test delegates, and are always fine as
		//a special case.
		return nil
	}

	pieces := strings.Split(path, "/")

	lastPiece := pieces[len(pieces)-1]

	if lastPiece != delegate.Name() {
		return errors.New("Delegate name (" + delegate.Name() + ") did not match the last part of the package name (" + path + "), which is required.")
	}

	return nil

}

//NewGameManager creates a new game manager with the given GameDelegate. It
//will validate that the various sub-states are reasonable, and will call
//ConfigureMoves and ConfigureAgents and then check that all tiems are
//configured reaasonably. It does a large amount of verification and wiring up
//of your game type to get it ready for use, and will error if any part of the
//configuration appears suspect.
func NewGameManager(delegate GameDelegate, storage StorageManager) (*GameManager, error) {
	if delegate == nil {
		return nil, errors.New("No delegate provided")
	}

	if delegate.Manager() != nil {
		return nil, errors.New("that delegate has already been associated with another game manager")
	}

	matched, err := regexp.MatchString(`^[0-9a-zA-Z]+$`, delegate.Name())

	if err != nil {
		return nil, errors.New("The legal name regexp failed: " + err.Error())
	}

	if !matched {
		return nil, errors.New("your delegate's name contains illegal characters")
	}

	if err := gamePkgMatchesDelegateName(delegate); err != nil {
		return nil, err
	}

	if storage == nil {
		return nil, errors.New("No Storage provided")
	}

	variantConfig := delegate.Variants()

	//Esnure it's initalized
	variantConfig.Initialize()

	if err := variantConfig.Valid(); err != nil {
		return nil, errors.New("The provided Variants config was not valid: " + err.Error())
	}

	chest := newComponentChest(delegate.ConfigureEnums())

	for name, deck := range delegate.ConfigureDecks() {
		if err := chest.addDeck(name, deck); err != nil {
			return nil, errors.New("Couldn't add deck named " + name + ": " + err.Error())
		}
	}

	for name, val := range delegate.ConfigureConstants() {
		if err := chest.addConstant(name, val); err != nil {
			return nil, errors.New("Couldn't add constant named " + name + ": " + err.Error())
		}
	}

	chest.finish()

	result := &GameManager{
		delegate:      delegate,
		chest:         chest,
		storage:       storage,
		logger:        logrus.New(),
		movesByName:   make(map[string]*moveType),
		variantConfig: variantConfig,
	}

	chest.manager = result

	delegate.SetManager(result)

	if !delegate.LegalNumPlayers(delegate.DefaultNumPlayers()) {
		return nil, errors.New("The default number of players is not legal")
	}

	if !delegate.LegalNumPlayers(delegate.MinNumPlayers()) {
		return nil, errors.New("The MinNumPlayers is not legal")
	}

	if !delegate.LegalNumPlayers(delegate.MaxNumPlayers()) {
		return nil, errors.New("The MaxNumPlayers is not legal")
	}

	if err := result.setUpValidators(); err != nil {
		return nil, errors.New("Couldn't configure validators: " + err.Error())
	}

	groupEnum := result.delegate.GroupEnum()

	groupNames := result.gameValidator.sanitizationPolicyGroupNames(groupEnum)
	if len(groupNames) > 0 {
		return nil, errors.New("Game state had illegal group names in sanitization policy: " + fmt.Sprint(groupNames))
	}

	//we can skip playerValidator for now; they're legal to have extra group
	//names. Later, we'll verify that a SanitizedForPlayer state works, which
	//will implicitly test all groupNames are valid.

	for deckName, deckValidator := range result.dynamicComponentValidator {
		groupNames = deckValidator.sanitizationPolicyGroupNames(groupEnum)
		if len(groupNames) > 0 {
			return nil, errors.New("DynamicComponentValues " + deckName + " state had illegal group names in sanitization policy: " + fmt.Sprint(groupNames))
		}
	}

	if err := result.installMoves(delegate.ConfigureMoves()); err != nil {
		return nil, errors.New("Failed to install moves: " + err.Error())
	}

	result.agents = delegate.ConfigureAgents()

	exampleState, err := result.newGame("", "").starterState(delegate.DefaultNumPlayers())

	if err != nil {
		return nil, errors.New("Couldn't get exampleState: " + err.Error())
	}

	if err := verifySubStatesConnectedAndValid(exampleState); err != nil {
		return nil, errors.New("The substates weren't valid: " + err.Error())
	}

	for _, moveType := range result.moves {
		testMove := moveType.NewMove(exampleState)

		if err := testMove.ValidConfiguration(exampleState); err != nil {
			return nil, errors.New(moveType.Name() + " move failed the ValidConfiguration test: " + err.Error())
		}

	}

	//Verify that all of the int values returned by GroupMembership are part of
	//groupEnum. a nil return value is fine.
	if groupMembership := result.delegate.GroupMembership(exampleState.ImmutablePlayerStates()[0]); groupMembership != nil {
		if len(groupMembership) > 0 && groupEnum == nil {
			return nil, errors.New("delegate.GroupMembership returned keys but groupEnum was nil")
		}
		for k := range groupMembership {
			if !groupEnum.Valid(k) {
				return nil, errors.New("delegate.GroupMembership returned an int not in GroupEnum: " + strconv.Itoa(k))
			}
		}
	}

	//This will implicitly check that the extra group names for playerValidator
	//are all handled by computedPlayerGroupMembership.
	if _, err := exampleState.SanitizedForPlayer(0); err != nil {
		return nil, errors.New("Couldn't sanitize for player: " + err.Error())
	}

	result.agentsByName = make(map[string]Agent)
	for _, agent := range result.agents {
		result.agentsByName[strings.ToLower(agent.Name())] = agent
	}

	result.modifiableGames = make(map[string]*Game)

	result.timers = newTimerManager(result)

	//Start ticking timers.
	go func() {
		//TODO: is there a way to turn off timer ticking for a manager we want
		//to throw out?
		for {
			<-time.After(250 * time.Millisecond)
			result.timers.Tick()
		}
	}()

	result.initialized = true

	return result, nil
}

//verifyValidConfigurationOnStruct verifies that if there are any sub-structs
//that have been embedded that satisfy ValidConfiguration that they have that
//checked. This is an expensive method that uses reflection and so should not be
//called except during NewGameManager.
func verifyValidConfigurationOnStruct(state State, strct interface{}) error {

	v := reflect.ValueOf(strct).Elem()
	t := reflect.TypeOf(v.Interface())

	if t.Kind() != reflect.Struct {
		return errors.New("expected strct to be a struct or a pointer to a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type
		if fieldType.Kind() != reflect.Struct {
			continue
		}
		//Calling interface on a non-exported field will panic; check if we can.
		if !v.Field(i).CanInterface() {
			continue
		}

		//Often the structs are embedded directly (that is, not a pointer). But
		//the methods for ValidConfiguration often take a pointer receiver.
		embeddedStructValue := v.Field(i)
		if embeddedStructValue.Type().Kind() != reflect.Ptr {
			embeddedStructValue = embeddedStructValue.Addr()
		}
		embeddedStruct := embeddedStructValue.Interface()

		validator, ok := embeddedStruct.(ConfigurationValidator)
		if !ok {
			continue
		}
		if err := validator.ValidConfiguration(state); err != nil {
			return errors.New("Struct field " + fieldType.Name() + " had a valid configuration that wasn't valid: " + err.Error())
		}
	}

	return nil
}

func verifySubStatesConnectedAndValid(exampleState State) error {
	if exampleState.GameState().State() != exampleState {
		return errors.New("GameState returned different state")
	}
	if err := verifyValidConfigurationOnStruct(exampleState, exampleState.GameState()); err != nil {
		return err
	}
	for i, pState := range exampleState.PlayerStates() {
		if pState.State() != exampleState {
			return errors.New("PlayerState " + strconv.Itoa(i) + " returned different state")
		}
		if i == 0 {
			//Only need to bother doing the expensive reflection checking on one.
			if err := verifyValidConfigurationOnStruct(exampleState, pState); err != nil {
				return err
			}
		}
	}
	for deckName, values := range exampleState.DynamicComponentValues() {
		for i, value := range values {
			if value.State() != exampleState {
				return errors.New("DynamicComponentValues " + deckName + " " + strconv.Itoa(i) + " returned different state")
			}
			if i == 0 {
				//Only need to bother doing the expensive reflection checking on one.
				if err := verifyValidConfigurationOnStruct(exampleState, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//Variants returns the VariantConfig for this game type. A simple wrapper
//around gameDelegate.Variants() that calls Initialize() on that result and
//also memoizes it for performance.
func (g *GameManager) Variants() VariantConfig {
	return g.variantConfig
}

func (g *GameManager) installMoves(configs []MoveConfig) error {

	if configs == nil {
		return errors.New("No move configs provided")
	}

	for i, config := range configs {
		if err := g.addMove(config); err != nil {
			return errors.New("Move Config " + strconv.Itoa(i) + " could not be installed: " + err.Error())
		}
	}

	return nil
}

func (g *GameManager) setUpValidators() error {
	if g.chest == nil {
		return errors.New("No chest provided")
	}

	if g.storage == nil {
		return errors.New("Storage not provided")
	}

	fakeState := &state{}

	exampleGameState := g.delegate.GameStateConstructor()

	if exampleGameState == nil {
		return errors.New("GameStateConstructor returned nil")
	}

	exampleGameState.ConnectContainingState(nil, StatePropertyRef{Group: StateGroupGame})

	validator, err := NewStructInflater(exampleGameState, nil, g.chest)

	if err != nil {
		return errors.New("Could not validate empty game state: " + err.Error())
	}

	//Technically we don't need to do this test inflation now, but we might as
	//well catch these problems at SetUp instead of later.

	if err = validator.Inflate(exampleGameState, fakeState); err != nil {
		return errors.New("Couldn't auto inflate empty game state: " + err.Error())
	}

	if err = validator.Valid(exampleGameState); err != nil {
		return errors.New("Default infflated empty game state was not valid: " + err.Error())
	}

	g.gameValidator = validator

	examplePlayerState := g.delegate.PlayerStateConstructor(0)

	if examplePlayerState == nil {
		return errors.New("PlayerStateConstructor returned nil")
	}

	examplePlayerState.ConnectContainingState(nil, StatePropertyRef{Group: StateGroupPlayer})

	validator, err = NewStructInflater(examplePlayerState, nil, g.chest)

	if err != nil {
		return errors.New("Could not validate empty player state: " + err.Error())
	}

	if err = validator.Inflate(examplePlayerState, fakeState); err != nil {
		return errors.New("Couldn't auto inflate empty player state: " + err.Error())
	}

	if err = validator.Valid(examplePlayerState); err != nil {
		return errors.New("Default infflated empty player state was not valid: " + err.Error())
	}

	g.playerValidator = validator

	g.dynamicComponentValidator = make(map[string]*StructInflater)

	for _, deckName := range g.chest.DeckNames() {
		deck := g.chest.Deck(deckName)

		exampleDynamicComponentValue := g.delegate.DynamicComponentValuesConstructor(deck)

		if exampleDynamicComponentValue == nil {
			continue
		}

		exampleDynamicComponentValue.ConnectContainingState(nil, StatePropertyRef{Group: StateGroupDynamicComponentValues, DeckName: deckName})

		validator, err = NewStructInflater(exampleDynamicComponentValue, nil, g.chest)

		if err != nil {
			return errors.New("Could not validate empty dynamic component state for " + deckName + ": " + err.Error())
		}

		if err = validator.Inflate(exampleDynamicComponentValue, fakeState); err != nil {
			return errors.New("Couldn't auto inflate empty dynamic component state for " + deckName + ": " + err.Error())
		}

		if err = validator.Valid(exampleDynamicComponentValue); err != nil {
			return errors.New("Default infflated empty dynamic component state for " + deckName + " was not valid: " + err.Error())
		}

		g.dynamicComponentValidator[deckName] = validator
	}

	return nil
}

//computedPlayerGroupMembership is how we compute special group names for player
//states. player is the player state being prepared, and viewingAsPlayer is the
//state for the viewing as player, or an empty map if the viewingAsPlayer is
//obdserver.
func (g *GameManager) computedPlayerGroupMembership(groupName string, player, viewingAsPlayer PlayerIndex, playerMembership, viewingAsPlayerMembership map[int]bool) (bool, error) {
	if groupName == sanitizationGroupSelf {
		if player == viewingAsPlayer {
			return true, nil
		}
		return false, nil
	} else if groupName == sanitizationGroupOther {
		if player != viewingAsPlayer {
			return true, nil
		}
		return false, nil
	}

	return g.Delegate().ComputedPlayerGroupMembership(groupName, playerMembership, viewingAsPlayerMembership)
}

//Logger returns the logrus.Logger that is in use for this game. This is a
//reasonable place to emit info or debug information specific to your game.
//This is initialized to a default logger when NewGameManager is called, and
//calls to SetLogger will fail if the logger is nil, so this will always
//return a non-nil logger. Change the logger in use for this GameManager via
//SetLogger().
func (g *GameManager) Logger() *logrus.Logger {
	return g.logger
}

//SetLogger configures the manager to use the given logger (which can be
//accessed via GameManager.Logger()) Will fail if logger is nil.
func (g *GameManager) SetLogger(logger *logrus.Logger) {
	if logger == nil {
		return
	}
	g.logger = logger
}

//NewDefaultGame returns a NewGame with everything set to default. Simple
//sugar for NewGame(0, nil, nil).
func (g *GameManager) NewDefaultGame() (*Game, error) {
	return g.NewGame(0, nil, nil)
}

//NewGame returns a new specific game instation that is set up with these
//options, persisted to the datastore, starter state created, first round of
//fix up moves applied, and in general ready for the first move to be
//proposed. The variant will be passed to delegate.Variant().NewVariant(). If
//the game you want to access has already been created, use GameManager.Game()
//or ModifiableGame().
func (g *GameManager) NewGame(numPlayers int, variantValues map[string]string, agentNames []string) (*Game, error) {
	return g.createGame("", "", numPlayers, variantValues, agentNames)
}

func (g *GameManager) createGame(id, secretSalt string, numPlayers int, variantValues map[string]string, agentNames []string) (*Game, error) {
	result, err := g.newGameImpl(id, secretSalt)

	if err != nil {
		return nil, err
	}

	if err := result.setUp(numPlayers, variantValues, agentNames); err != nil {
		return nil, err
	}

	return result, nil
}

//recreateGame is designed to be called by Internals().RecreateGame. See its
//documentation.
func (g *GameManager) recreateGame(rec *GameStorageRecord) (*Game, error) {

	if rec == nil {
		return nil, errors.New("No GameStorageRecord provided")
	}

	if rec.ID == "" {
		return nil, errors.New("That Id is not valid")
	}

	if rec.SecretSalt == "" {
		return nil, errors.New("that secret salt is not valid")
	}

	if other, _ := g.Storage().Game(rec.ID); other != nil {
		return nil, errors.New("A game with that Id already exists in this storage pool. Did you mean to use manager.NewGame, manager.Game(), or manager.ModifiableGame instead?")
	}

	return g.createGame(rec.ID, rec.SecretSalt, rec.NumPlayers, rec.Variant, rec.Agents)

}

//newGameImpl is NewGame, but without calling SetUp. Broken out only for tests
//internal to this package.
func (g *GameManager) newGameImpl(id, secretSalt string) (*Game, error) {
	result := g.newGame(id, secretSalt)

	if err := g.modifiableGameCreated(result); err != nil {
		return nil, errors.New("Couldn't warn that a modifiable game was created: " + err.Error())
	}

	return result, nil
}

//newGame is the inner portion of creating a valid game object, but we don't
//yet tell the system that it exists because we expect to throw it out before
//saving it. You almost never want this, use NewGame instead. If id or
//secretSalt are "", then reasonable ones will be created automatically.
func (g *GameManager) newGame(id, secretSalt string) *Game {
	if g == nil {
		return nil
	}

	if id == "" {
		id = randomString(gameIDLength, nil)
	}

	if secretSalt == "" {
		secretSalt = randomString(gameIDLength, nil)
	}

	return &Game{
		manager: g,
		//TODO: set the size of chan based on something more reasonable.
		//Note: this is also set similarly in manager.ModifiableGame
		proposedMoves:  make(chan *proposedMoveItem, 20),
		fixUpTriggered: make(chan DelayedError, 10),
		id:             id,
		secretSalt:     secretSalt,
		modifiable:     true,
	}
}

func (g *GameManager) gameFromStorageRecord(record *GameStorageRecord) *Game {

	//Sanity check that this game actually does match with this manager.
	if record.Name != g.Delegate().Name() {
		return nil
	}

	return &Game{
		manager:    g,
		version:    record.Version,
		id:         record.ID,
		secretSalt: record.SecretSalt,
		finished:   record.Finished,
		winners:    record.Winners,
		numPlayers: record.NumPlayers,
		created:    record.Created,
		agents:     record.Agents,
		variant:    record.Variant,
		modifiable: false,
		initalized: true,
	}
}

//modifiableGameCreated lets Manager know that a modifiable game was created
//with the given ID, so that manager can vend that later if necessary. It is
//designed to only be called from NewGame.
func (g *GameManager) modifiableGameCreated(game *Game) error {
	if !g.initialized {
		return errors.New("Game is not setup yet")
	}

	g.modifiableGamesLock.RLock()
	_, ok := g.modifiableGames[game.ID()]
	g.modifiableGamesLock.RUnlock()

	if ok {
		return errors.New("modifiableGameCreated collided with existing game")
	}

	id := strings.ToUpper(game.ID())

	g.modifiableGamesLock.Lock()
	g.modifiableGames[id] = game
	g.modifiableGamesLock.Unlock()

	return nil
}

//ModifiableGame returns a modifiable Game with the given ID. Either it
//returns one it already knows about that is resident in memory, or it creates
//a modifiable version from storage (if one is stored in storage). If a game
//cannot be created from those ways, it will return nil. The primary way to
//avoid race conditions with the same underlying game being stored to the
//store is that only one modifiable copy of a Game should exist at a time. It
//is up to the specific user of boardgame to ensure that is the case. As long
//as manager.Game is used, a single manager in a given application binary will
//not allow multiple modifiable versions of a single game to be "checked out".
//However, if there could be multiple managers loaded up at the same time for
//the same store, it's possible to have a race condition. For example, it
//makes sense to have only a single server that takes in proposed moves from a
//queue and then applies them to a modifiable version of the given game.
func (g *GameManager) ModifiableGame(id string) *Game {

	id = strings.ToUpper(id)

	g.modifiableGamesLock.RLock()
	game := g.modifiableGames[id]
	g.modifiableGamesLock.RUnlock()

	if game != nil {
		return game
	}

	//Let's try to load up from storage.

	gameRecord, _ := g.storage.Game(id)

	if gameRecord == nil {
		//Nah, we've never seen that game.
		return nil
	}

	game = g.gameFromStorageRecord(gameRecord)

	//Only SetUp() and us are allowed to kick off a game's mainLoop.
	game.modifiable = true
	//TODO: set the size of chan based on something more reasonable.
	//Note: this is also set similarly in NewGame
	game.proposedMoves = make(chan *proposedMoveItem, 20)
	game.fixUpTriggered = make(chan DelayedError, 10)
	go game.mainLoop()

	g.modifiableGamesLock.Lock()
	g.modifiableGames[id] = game
	g.modifiableGamesLock.Unlock()

	return game

}

//Game fetches a new non-modifiable copy of the given game from storage. If
//you want a modifiable version, see ModifiableGame. You'd use this method
//instead of ModifiableGame in situations where you're on a read-only servant
//binary.
func (g *GameManager) Game(id string) *Game {
	record, err := g.storage.Game(id)

	if err != nil {
		return nil
	}

	return g.gameFromStorageRecord(record)
}

type refriedState struct {
	Game            json.RawMessage
	Players         []json.RawMessage
	Components      map[string][]json.RawMessage
	SecretMoveCount map[string][]int
	Version         int
}

//playerStateConstructor is a simple wrapper around
//delegate.PlayerStateConstructor that just verifies that stacks are inflated.
func (g *GameManager) playerStateConstructor(state *state, player PlayerIndex) (ConfigurableSubState, error) {

	playerState := g.delegate.PlayerStateConstructor(player)

	if playerState == nil {
		return nil, errors.New("PlayerStateConstructor returned nil for " + strconv.Itoa(int(player)))
	}

	if err := g.playerValidator.Inflate(playerState, state); err != nil {
		return nil, errors.New("Couldn't auto-inflate empty player state: " + err.Error())
	}

	if err := g.playerValidator.Valid(playerState); err != nil {
		return nil, errors.New("Player State was not valid: " + err.Error())
	}

	return playerState, nil

}

//GameStateConstructor is a simple wrapper around
//delegate.GameStateConstructor that just verifies that stacks are inflated.
func (g *GameManager) gameStateConstructor(state *state) (ConfigurableSubState, error) {

	gameState := g.delegate.GameStateConstructor()

	if gameState == nil {
		return nil, errors.New("GameStateConstructor returned nil")
	}

	if err := g.gameValidator.Inflate(gameState, state); err != nil {
		return nil, errors.New("Couldn't auto-inflate empty game state: " + err.Error())
	}

	if err := g.gameValidator.Valid(gameState); err != nil {
		return nil, errors.New("game State was not valid: " + err.Error())
	}

	return gameState, nil

}

func (g *GameManager) dynamicComponentValuesConstructor(state *state) (map[string][]ConfigurableSubState, error) {
	result := make(map[string][]ConfigurableSubState)

	for _, deckName := range g.Chest().DeckNames() {

		deck := g.Chest().Deck(deckName)

		if deck == nil {
			return nil, errors.New("Couldn't find deck for " + deckName)
		}

		values := g.Delegate().DynamicComponentValuesConstructor(deck)
		if values == nil {
			continue
		}

		//Check outside the loop if it has containing component so we can skip
		//it within the loop if doesn't, for minor savings.
		_, hasContainingComponent := values.(ComponentValues)

		validator := g.dynamicComponentValidator[deckName]

		if validator == nil {
			return nil, errors.New("Unexpectedly couldn't find validator for deck " + deckName)
		}

		arr := make([]ConfigurableSubState, len(deck.Components()))
		for i := 0; i < len(deck.Components()); i++ {
			arr[i] = g.Delegate().DynamicComponentValuesConstructor(deck)

			if hasContainingComponent {
				componentValues, ok := arr[i].(ComponentValues)

				containingComponent := deck.ComponentAt(i)

				//It would be unexpected if we couldn't cast (that would imply
				//we were gettind different shaped values for the same deck
				//name over time), but check to be safe.
				if ok {
					componentValues.SetContainingComponent(containingComponent)
				}
			}

			if err := validator.Inflate(arr[i], state); err != nil {
				return nil, errors.New("Couldn't auto-inflate dynamic compoonent values for " + deckName + " " + strconv.Itoa(i) + ": " + err.Error())
			}

			if err := validator.Valid(arr[i]); err != nil {
				return nil, errors.New("Dynamic compoonent values for " + deckName + " " + strconv.Itoa(i) + " was not valid: " + err.Error())
			}

		}
		result[deckName] = arr

	}

	return result, nil
}

//StateFromBlob takes a state that was serialized in storage and reinflates
//it. Storage sub-packages should call this to recover a real State object
//given a serialized state blob. Note: the state that is returned does not
//have its game property set.
func (g *GameManager) stateFromRecord(record StateStorageRecord, version int) (*state, error) {
	//At this point, no extra state is stored in the blob other than in props.

	//We can't just delegate to StateProps to unmarshal itself, because it
	//needs a reference to delegate to inflate, and only we have that.
	var refried refriedState

	if err := json.Unmarshal(record, &refried); err != nil {
		return nil, err
	}

	//Old state blobs might have their own version, but new ones don't encode
	//it. (It's always implied, and it's just a random thing in state blobs that
	//changes that doens't need to)
	refried.Version = version

	result, err := g.emptyState(len(refried.Players))

	if err != nil {
		return nil, errors.New("Couldn't create an empty state: " + err.Error())
	}

	if refried.SecretMoveCount != nil {
		result.secretMoveCount = refried.SecretMoveCount
	}
	result.version = refried.Version

	if err := json.Unmarshal(refried.Game, result.gameState); err != nil {
		return nil, errors.New("Unmarshal of GameState failed: " + err.Error())
	}

	for i, blob := range refried.Players {

		if err := json.Unmarshal(blob, result.playerStates[i]); err != nil {
			return nil, errors.New("Unmarshal into player state failed for " + strconv.Itoa(i) + " player: " + err.Error())
		}

	}

	for deckName, values := range refried.Components {
		resultDeckValues := result.dynamicComponentValues[deckName]
		//TODO: detect the case where the emptycompontentvalues has decknames that are not in the JSON.
		if resultDeckValues == nil {
			return nil, errors.New("The empty dynamic component state didn't have deck name: " + deckName)
		}
		if len(values) != len(resultDeckValues) {
			return nil, errors.New("The empty dynamic component state for deck " + deckName + " had wrong length. Got " + strconv.Itoa(len(values)) + " wanted " + strconv.Itoa(len(resultDeckValues)))
		}
		for i := 0; i < len(values); i++ {

			value := values[i]
			resultDeckValue := resultDeckValues[i]

			if err := json.Unmarshal(value, resultDeckValue); err != nil {
				return nil, errors.New("Error unmarshaling component state for deck " + deckName + " index " + strconv.Itoa(i) + ": " + err.Error())
			}
		}
	}

	return result, nil

}

//proposeMoveOnGame is how non-modifiable games should tell the manager they
//have a move they want to make on a given move ID. For now it's just a simple
//wrapper around ModifiableGame, but in multi-server situations, in the future
//it would conceivably do an RPC or something. Note that game.triggerFixUp()
//also does this kind of dispatching.
func (g *GameManager) proposeMoveOnGame(nonModifiableGame *Game, move Move, proposer PlayerIndex) DelayedError {

	//The chan that the core logic will tell us the move is done in.
	errChan := make(DelayedError, 1)
	//The chan that we'll erturn on after refreshing the game.
	finalErrChan := make(DelayedError, 1)

	go func() {
		result := <-errChan
		nonModifiableGame.Refresh()
		finalErrChan <- result
	}()

	//ModifiableGame could take awhile if it has to be fetched from storage,
	//so we'll run all of this in a goroutine since we're returning a
	//DelayedError anyway.

	go func() {
		game := g.ModifiableGame(nonModifiableGame.ID())

		if game == nil {
			errChan <- errors.New("There was no game with that ID")
			return
		}

		workItem := &proposedMoveItem{
			move:     move,
			ch:       errChan,
			proposer: proposer,
		}

		game.proposedMoves <- workItem

	}()

	return finalErrChan

}

//ExampleState will return a fully-constructed state for this game, with a
//single player and no specific game object associated. This is a convenient
//way to inspect the final shape of your State objects using your various
//Constructor() methods and after tag-based inflation. Primarily useful for
//meta- programming approaches, used often in the moves package.
func (g *GameManager) ExampleState() ImmutableState {
	state, err := g.emptyState(1)
	if err != nil {
		return nil
	}
	return state
}

//emptyState returns an empty state for this game with this number of players.
//This is the canonical way to create a new state object with all of the right
//auto-inflation and everything.
func (g *GameManager) emptyState(numPlayers int) (*state, error) {
	stateCopy := &state{
		manager: g,
		//Other users should set the game to a real thing
		game:            nil,
		version:         0,
		secretMoveCount: make(map[string][]int),
	}

	gameState, err := g.gameStateConstructor(stateCopy)

	if err != nil {
		return nil, err
	}

	stateCopy.gameState = gameState

	playerStates := make([]ConfigurableSubState, numPlayers)

	for i := 0; i < numPlayers; i++ {
		playerState, err := g.playerStateConstructor(stateCopy, PlayerIndex(i))

		if err != nil {
			return nil, err
		}

		playerStates[i] = playerState
	}

	stateCopy.playerStates = playerStates

	dynamic, err := g.dynamicComponentValuesConstructor(stateCopy)

	if err != nil {
		return nil, errors.New("Couldn't create empty dynamic component values: " + err.Error())
	}

	stateCopy.dynamicComponentValues = dynamic

	if err := stateCopy.setStateForSubStates(); err != nil {
		return nil, errors.New("error connecting sub states: " + err.Error())
	}

	return stateCopy, nil
}

func (g *GameManager) addMove(config MoveConfig) error {

	moveType, err := newMoveType(config, g)

	if err != nil {
		return err
	}

	if g.initialized {
		return errors.New("gameManager has already been SetUp so no new moves may be added")
	}

	moveName := strings.ToLower(moveType.Name())

	//TODO: theoeretically if the move name has already been added we want to
	//replace it with the new one. ... But that requires splicing out things
	//and will be error prone, so need to do it carefully.

	if g.movesByName[moveName] != nil {
		return errors.New(moveType.Name() + " was already installed as a move and cannot be installed again.")
	}

	g.moves = append(g.moves, moveType)
	g.movesByName[moveName] = moveType

	return nil
}

//ExampleMoves returns a list of example moves, which are moves not initalized
//based on a state. The list of moves is based on what your
//GameDelegate.ConfigureMoves() returned.
func (g *GameManager) ExampleMoves() []Move {

	mTypes := g.moveTypes()

	if mTypes == nil {
		return nil
	}

	result := make([]Move, len(mTypes))

	for i, mType := range mTypes {
		result[i] = mType.NewMove(nil)
	}

	return result

}

//ExampleMoveByName returns an example move with that name, but without
//initializing it with a state. See also ExampleMoves().
func (g *GameManager) ExampleMoveByName(name string) Move {

	mType := g.moveTypeByName(name)

	if mType == nil {
		return nil
	}

	return mType.NewMove(nil)
}

//Agents returns a slice of all agents configured on this Manager via
//GameDelegate.ConfigureAgents.
func (g *GameManager) Agents() []Agent {
	if !g.initialized {
		return nil
	}

	return g.agents
}

//MoveTypes returns all moves that are valid in this game: all of the Moves
//that have been added via AddMove during initalization. Returns nil until
//game.SetUp() has been called.
func (g *GameManager) moveTypes() []*moveType {
	if !g.initialized {
		return nil
	}

	return g.moves
}

//AgentByName will return the agent with the given name, or nil if one doesn't
//exist. See also Agents()
func (g *GameManager) AgentByName(name string) Agent {

	if !g.initialized {
		return nil
	}

	name = strings.ToLower(name)

	return g.agentsByName[name]
}

//MoveTypeByName returns the MoveType of that name from game.MoveTypes(), if
//it exists. Names are considered without regard to case.  Will return a copy.
func (g *GameManager) moveTypeByName(name string) *moveType {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.movesByName[name]

	if move == nil {
		return nil
	}

	return move
}

//Chest is the ComponentChest in use for this game.
func (g *GameManager) Chest() *ComponentChest {
	return g.chest
}

//Storage is the StorageManager games that was associated with this
//GameManager in NewGameManager.
func (g *GameManager) Storage() StorageManager {
	return g.storage
}

//Delegate returns the GameDelegate configured for these games, that was
//associated with this GameManager in NewGameManager.
func (g *GameManager) Delegate() GameDelegate {
	return g.delegate
}
