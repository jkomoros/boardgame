package boardgame

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/jkomoros/boardgame/errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Moves is the set of all move types that are ever legal to apply in this
//game. When a move will be proposed it should copy one of these moves.
//Player moves are moves that can be applied by users. FixUp moves are
//only ever returned by Delegate.ProposeFixUpMove().

//TODO: figure out where the above comment should go for documentation.

//GameManager is a struct that keeps track of configuration that is common
//across multiple games. It is specifically designed to be used with multiple
//games.
type GameManager struct {
	delegate                  GameDelegate
	gameValidator             *readerValidator
	playerValidator           *readerValidator
	dynamicComponentValidator map[string]*readerValidator
	chest                     *ComponentChest
	storage                   StorageManager
	agents                    []Agent
	moves                     []*MoveType
	movesByName               map[string]*MoveType
	agentsByName              map[string]Agent
	modifiableGamesLock       sync.RWMutex
	modifiableGames           map[string]*Game
	timers                    *timerManager
	initialized               bool
	logger                    *logrus.Logger
}

type moveTypeConfingBundleRun struct {
	ordered bool
	phase   int
	configs []*MoveTypeConfig
}

//MoveTypeConfigBundle is a bundle of move type configs to add to a game. You
//return one from your delegate's ConfigMoves method, which is the primary way
//to install moves on a game type. None of the Add* methods return errors
//immediately; NewGameManager will itself error if any of the moves to add are
//illegal for any reason.
type MoveTypeConfigBundle struct {
	runs []*moveTypeConfingBundleRun
}

//NewMoveTypeConfigBundle returns a new empty bundle ready to have moves added
//to it via the various Add* methods.
func NewMoveTypeConfigBundle() *MoveTypeConfigBundle {
	return &MoveTypeConfigBundle{}
}

//AddMove adds the specified move type to the game as a move. Returns the
//bundle itself for convenience so you can chain.
func (m *MoveTypeConfigBundle) AddMove(config *MoveTypeConfig) *MoveTypeConfigBundle {
	m.runs = append(m.runs, &moveTypeConfingBundleRun{
		ordered: false,
		phase:   -1,
		configs: []*MoveTypeConfig{config},
	})
	return m
}

//AddMoves is a simple wrapper around AddMoves. It is useful for move configs
//that are legal in any phase in any order. If you want to configure moves
//that are only legal in certain phases, use AddMovesForPhase. If you want to
//add moves that are only legal in certain phases in certain orders, use
//AddOrderedMovesForPhase instead. Unlike the AddMovesForPhase variants,
//AddMoves doesn't modify the LegalPhases of the movs you add. Returns the
//bundle itself for convenience so you can chain.
func (m *MoveTypeConfigBundle) AddMoves(config ...*MoveTypeConfig) *MoveTypeConfigBundle {
	m.runs = append(m.runs, &moveTypeConfingBundleRun{
		ordered: false,
		phase:   -1,
		configs: config,
	})
	return m
}

//AddMovesForPhase is a convenience wrapper around AddMoves. It is useful to
//install moves that are only legal in a specific phase, but in any order. As
//a convenience, if the move configs you pass do not already affirmatively
//list the phase being configured, then they will have it added to the config
//before adding (as long as the LegalPhases isn't a zero-length slice). This
//means that in most cases you can skip defining LegalPhases, as it will be
//configured automatically. See AddOrderedMovesForPhase for an ordered
//variant. Returns the bundle itself for convenience so you can chain.
func (m *MoveTypeConfigBundle) AddMovesForPhase(phase int, config ...*MoveTypeConfig) *MoveTypeConfigBundle {
	m.runs = append(m.runs, &moveTypeConfingBundleRun{
		ordered: false,
		phase:   phase,
		configs: config,
	})
	return m
}

//AddOrderedMovesForPhase is a variant around AddMovesForPhase that in
//addition to enforcing the moves are only legal in a given phase will also
//set a specific order. (Moves that are legal in every phase will not count in
//the order matching). Will error if your delegate does not implement
//PhaseMoveProgressionSetter (DefaultGameDelegate does by default). Returns
//the bundle itself for convenience so you can chain.
func (m *MoveTypeConfigBundle) AddOrderedMovesForPhase(phase int, config ...*MoveTypeConfig) *MoveTypeConfigBundle {
	m.runs = append(m.runs, &moveTypeConfingBundleRun{
		ordered: true,
		phase:   phase,
		configs: config,
	})
	return m
}

//NewGameManager creates a new game manager with the given delegate. It will
//validate that the various sub-states are reasonable, and will call
//ConfigureMoves and ConfigureAgents and then check that all tiems are
//configured reaasonably.
func NewGameManager(delegate GameDelegate, chest *ComponentChest, storage StorageManager) (*GameManager, error) {
	if delegate == nil {
		return nil, errors.New("No delegate provided")
	}

	if chest == nil {
		return nil, errors.New("No chest provided")
	}

	matched, err := regexp.MatchString(`^[0-9a-zA-Z_-]+$`, delegate.Name())

	if err != nil {
		return nil, errors.New("The legal name regexp failed: " + err.Error())
	}

	if !matched {
		return nil, errors.New("Your delegate's name contains illegal characters.")
	}

	//Make sure the chest is no longer open for modification. If finish was
	//already called, this will be a no-op.
	chest.Finish()

	if storage == nil {
		return nil, errors.New("No Storage provided")
	}

	result := &GameManager{
		delegate:    delegate,
		chest:       chest,
		storage:     storage,
		logger:      logrus.New(),
		movesByName: make(map[string]*MoveType),
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

	if err := result.installMoveTypeConfigBundle(delegate.ConfigureMoves()); err != nil {
		return nil, errors.New("Failed to install moves: " + err.Error())
	}

	result.agents = delegate.ConfigureAgents()

	exampleState, err := result.newGame().starterState(delegate.DefaultNumPlayers())

	if err != nil {
		return nil, errors.New("Couldn't get exampleState: " + err.Error())
	}

	for _, moveType := range result.moves {
		testMove := moveType.NewMove(exampleState)

		if err := testMove.ValidConfiguration(exampleState); err != nil {
			return nil, errors.New(moveType.Name() + " move failed the ValidConfiguration test: " + err.Error())
		}

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

func (g *GameManager) installMoveTypeConfigBundle(m *MoveTypeConfigBundle) error {
	if m == nil {
		return errors.New("No bundle provided")
	}

	for _, run := range m.runs {
		if run.phase < 0 {

			if err := g.addMoves(run.configs...); err != nil {
				return err
			}

			continue
		}

		if !run.ordered {
			if err := g.addMovesForPhase(run.phase, run.configs...); err != nil {
				return err
			}
			continue
		}

		if err := g.addOrderedMovesForPhase(run.phase, run.configs...); err != nil {
			return err
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

	reader := exampleGameState.Reader()

	if reader == nil {
		return errors.New("GameStateConstructor's returned value returned nil for Reader")
	}

	validator, err := newReaderValidator(reader, exampleGameState.ReadSetter(), exampleGameState, nil, g.chest, false)

	if err != nil {
		return errors.New("Could not validate empty game state: " + err.Error())
	}

	//Technically we don't need to do this test inflation now, but we might as
	//well catch these problems at SetUp instead of later.

	readSetConfigurer := exampleGameState.ReadSetConfigurer()

	if readSetConfigurer == nil {
		return errors.New("GameStateConstructor's returned value returned nil for ReadSetConfigurer")
	}

	if err = validator.AutoInflate(readSetConfigurer, fakeState); err != nil {
		return errors.New("Couldn't auto inflate empty game state: " + err.Error())
	}

	if err = validator.Valid(reader); err != nil {
		return errors.New("Default infflated empty game state was not valid: " + err.Error())
	}

	g.gameValidator = validator

	examplePlayerState := g.delegate.PlayerStateConstructor(0)

	if examplePlayerState == nil {
		return errors.New("PlayerStateConstructor returned nil")
	}

	reader = examplePlayerState.Reader()

	if reader == nil {
		return errors.New("PlayerStateConstructor's returned value returned nil for Reader")
	}

	validator, err = newReaderValidator(reader, examplePlayerState.ReadSetter(), examplePlayerState, nil, g.chest, true)

	if err != nil {
		return errors.New("Could not validate empty player state: " + err.Error())
	}

	readSetConfigurer = examplePlayerState.ReadSetConfigurer()

	if readSetConfigurer == nil {
		return errors.New("PlayerStateConstructor's returned value returned nil for ReadSetConfigurer")
	}

	if err = validator.AutoInflate(readSetConfigurer, fakeState); err != nil {
		return errors.New("Couldn't auto inflate empty player state: " + err.Error())
	}

	if err = validator.Valid(reader); err != nil {
		return errors.New("Default infflated empty player state was not valid: " + err.Error())
	}

	g.playerValidator = validator

	g.dynamicComponentValidator = make(map[string]*readerValidator)

	for i, deckName := range g.chest.DeckNames() {
		deck := g.chest.Deck(deckName)

		exampleDynamicComponentValue := g.delegate.DynamicComponentValuesConstructor(deck)

		if exampleDynamicComponentValue == nil {
			continue
		}

		reader = exampleDynamicComponentValue.Reader()

		if reader == nil {
			return errors.New("DynamicComponentValue for " + deckName + " " + strconv.Itoa(i) + " reader returned nil")
		}

		validator, err = newReaderValidator(reader, exampleDynamicComponentValue.ReadSetter(), exampleDynamicComponentValue, nil, g.chest, false)

		if err != nil {
			return errors.New("Could not validate empty dynamic component state for " + deckName + ": " + err.Error())
		}

		readSetConfigurer = exampleDynamicComponentValue.ReadSetConfigurer()

		if readSetConfigurer == nil {
			return errors.New("DynamicComponentValue for " + deckName + " " + strconv.Itoa(i) + " readSetConfigurer returned nil")
		}

		if err = validator.AutoInflate(readSetConfigurer, fakeState); err != nil {
			return errors.New("Couldn't auto inflate empty dynamic component state for " + deckName + ": " + err.Error())
		}

		if err = validator.Valid(reader); err != nil {
			return errors.New("Default infflated empty dynamic component state for " + deckName + " was not valid: " + err.Error())
		}

		g.dynamicComponentValidator[deckName] = validator
	}

	return nil
}

//Logger returns the logrus.Logger that is in use for this game. This is a
//reasonable place to emit info or debug information specific to your game.
//This is initialized to a default logger when NewGameManager is called, and
//calls to SetLogger will fail if the logger is nil, so this will always
//return a non-nil logger.
func (g *GameManager) Logger() *logrus.Logger {
	return g.logger
}

//SetLogger configures the manager to use the given logger. Will fail if
//logger is nil.
func (g *GameManager) SetLogger(logger *logrus.Logger) {
	if logger == nil {
		return
	}
	g.logger = logger
}

//NewGame returns a new game. You must call SetUp before using it.
func (g *GameManager) NewGame() *Game {

	result := g.newGame()

	if err := g.modifiableGameCreated(result); err != nil {
		g.logger.Warn("Couldn't warn that a modifiable game was created: " + err.Error())
		return nil
	}

	return result

}

//newGame is the inner portion of creating a valid game object, but we don't
//yet tell the system that it exists because we expect to throw it out before
//saving it. You almost never want this, use NewGame instead.
func (g *GameManager) newGame() *Game {
	if g == nil {
		return nil
	}

	return &Game{
		manager: g,
		//TODO: set the size of chan based on something more reasonable.
		//Note: this is also set similarly in manager.ModifiableGame
		proposedMoves: make(chan *proposedMoveItem, 20),
		id:            randomString(gameIDLength),
		secretSalt:    randomString(gameIDLength),
		modifiable:    true,
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
		id:         record.Id,
		secretSalt: record.SecretSalt,
		finished:   record.Finished,
		winners:    record.Winners,
		numPlayers: record.NumPlayers,
		created:    record.Created,
		agents:     record.Agents,
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
	_, ok := g.modifiableGames[game.Id()]
	g.modifiableGamesLock.RUnlock()

	if ok {
		return errors.New("modifiableGameCreated collided with existing game")
	}

	id := strings.ToUpper(game.Id())

	g.modifiableGamesLock.Lock()
	g.modifiableGames[id] = game
	g.modifiableGamesLock.Unlock()

	return nil
}

//ModifiableGameForId returns a modifiable game with the given ID. Either it
//returns one it already knows about, or it creates a modifiable version from
//storage (if one is stored in storage). If a game cannot be created from
//those ways, it will return nil. The primary way to avoid race
//conditions with the same underlying game being stored to the store is
//that only one modifiable copy of a Game should exist at a time. It is up
//to the specific user of boardgame to ensure that is the case. As long as
//manager.LoadGame is used, a single manager will not allow multiple
//modifiable versions of a single game to be "checked out".  However, if
//there could be multiple managers loaded up at the same time for the same
//store, it's possible to have a race condition. For example, it makes
//sense to have only a single server that takes in proposed moves from a
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
	go game.mainLoop()

	g.modifiableGamesLock.Lock()
	g.modifiableGames[id] = game
	g.modifiableGamesLock.Unlock()

	return game

}

//Game fetches a new non-modifiable copy of the given game from storage. If
//you want a modifiable version, use ModifiableGame.
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
func (g *GameManager) playerStateConstructor(state *state, player PlayerIndex) (ConfigurablePlayerState, error) {

	playerState := g.delegate.PlayerStateConstructor(player)

	if playerState == nil {
		return nil, errors.New("PlayerStateConstructor returned nil for " + strconv.Itoa(int(player)))
	}

	readSetConfigurer := playerState.ReadSetConfigurer()

	if readSetConfigurer == nil {
		return nil, errors.New("PlayerState ReadSetConfigurer returned nil")
	}

	if err := g.playerValidator.AutoInflate(readSetConfigurer, state); err != nil {
		return nil, errors.New("Couldn't auto-inflate empty player state: " + err.Error())
	}

	if err := g.playerValidator.Valid(readSetConfigurer); err != nil {
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

	readSetConfigurer := gameState.ReadSetConfigurer()

	if readSetConfigurer == nil {
		return nil, errors.New("GameState ReadSetConfigurer returned nil")
	}

	if err := g.gameValidator.AutoInflate(readSetConfigurer, state); err != nil {
		return nil, errors.New("Couldn't auto-inflate empty game state: " + err.Error())
	}

	if err := g.gameValidator.Valid(readSetConfigurer); err != nil {
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

		validator := g.dynamicComponentValidator[deckName]

		if validator == nil {
			return nil, errors.New("Unexpectedly couldn't find validator for deck " + deckName)
		}

		arr := make([]ConfigurableSubState, len(deck.Components()))
		for i := 0; i < len(deck.Components()); i++ {
			arr[i] = g.Delegate().DynamicComponentValuesConstructor(deck)

			readSetConfigurer := arr[i].ReadSetConfigurer()

			if readSetConfigurer == nil {
				return nil, errors.New("ReadSetConfigurer for dynamic component values for " + deckName + " " + strconv.Itoa(i) + " was nil")
			}

			if err := validator.AutoInflate(readSetConfigurer, state); err != nil {
				return nil, errors.New("Couldn't auto-inflate dynamic compoonent values for " + deckName + " " + strconv.Itoa(i) + ": " + err.Error())
			}

			if err := validator.Valid(readSetConfigurer); err != nil {
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
func (g *GameManager) stateFromRecord(record StateStorageRecord) (*state, error) {
	//At this point, no extra state is stored in the blob other than in props.

	//We can't just delegate to StateProps to unmarshal itself, because it
	//needs a reference to delegate to inflate, and only we have that.
	var refried refriedState

	if err := json.Unmarshal(record, &refried); err != nil {
		return nil, err
	}

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
//it would conceivably do an RPC or something.
func (g *GameManager) proposeMoveOnGame(id string, move Move, proposer PlayerIndex) DelayedError {

	errChan := make(DelayedError, 1)

	//ModifiableGame could take awhile if it has to be fetched from storage,
	//so we'll run all of this in a goroutine since we're returning a
	//DelayedError anyway.

	go func() {
		game := g.ModifiableGame(id)

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

	return errChan

}

//ExampleState will return a fully-constructed state for this game, with a
//single player and no specific game object associated. This is a convenient
//way to inspect the final shape of your State objects using your various
//Constructor() methods and after tag-based inflation. Primarily useful for
//meta- programming approaches, used often in the moves package.
func (g *GameManager) ExampleState() MutableState {
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

	playerStates := make([]ConfigurablePlayerState, numPlayers)

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

	stateCopy.setStateForSubStates()

	return stateCopy, nil
}

func (g *GameManager) addMoves(config ...*MoveTypeConfig) error {

	if len(config) == 0 {
		return errors.New("No moveTypeConfigs provided")
	}

	for i, theConfig := range config {
		if err := g.addMove(theConfig); err != nil {
			return errors.New("Config " + strconv.Itoa(i) + " failed with error: " + err.Error())
		}
	}

	return nil

}

func (g *GameManager) addMovesForPhase(phase int, config ...*MoveTypeConfig) error {

	for i, moveConfig := range config {

		//If the moveConfig isn't a zero-len slice then if the phase isn't
		//affirmatively listed in the config as being legal we should add it.
		if moveConfig.LegalPhases == nil || len(moveConfig.LegalPhases) > 0 {

			hasTargetPhase := false

			for _, legalPhase := range moveConfig.LegalPhases {
				if legalPhase == phase {
					hasTargetPhase = true
					break
				}
			}

			if !hasTargetPhase {
				//If we didn't explicitly say that the given phase we're
				//configuring is legal on this move type, add it.

				//Note that in cases where the move type is legal in ALL phases,
				//this will lock it to only being legal in this move progression.
				//That's generally what you want--but not always.
				moveConfig.LegalPhases = append(moveConfig.LegalPhases, phase)
			}
		}

		if err := g.addMove(moveConfig); err != nil {
			return errors.New("Couldn't add " + strconv.Itoa(i) + " move config: " + err.Error())
		}
	}

	return nil
}

func (g *GameManager) addOrderedMovesForPhase(phase int, config ...*MoveTypeConfig) error {
	progressionSetter, ok := g.Delegate().(PhaseMoveProgressionSetter)
	if !ok {
		return errors.New("The delegate doest not implement PhaseMoveProgressionSetter, making this conveience method ineffective. Use AddGeneralMoves instead.")
	}

	var moveProgression []string

	for _, moveConfig := range config {

		moveProgression = append(moveProgression, moveConfig.Name)

	}

	if err := g.addMovesForPhase(phase, config...); err != nil {
		return err
	}

	progressionSetter.SetPhaseMoveProgression(phase, moveProgression)

	return nil
}

func (g *GameManager) addMove(config *MoveTypeConfig) error {

	moveType, err := config.NewMoveType(g)

	if err != nil {
		return err
	}

	if g.initialized {
		return errors.New("GameManager has already been SetUp so no new moves may be added.")
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

//Agents returns a slice of all agents configured on this Manager via
//GameDelegate.ConfigureAgents. Will return nil before SetUp is called.
func (g *GameManager) Agents() []Agent {
	if !g.initialized {
		return nil
	}

	return g.agents
}

//MoveTypes returns all moves that are valid in this game: all of the Moves
//that have been added via AddMove during initalization. Returns nil until
//game.SetUp() has been called.
func (g *GameManager) MoveTypes() []*MoveType {
	if !g.initialized {
		return nil
	}

	return g.moves
}

//AgentByName will return the agent with the given name, or nil if one doesn't
//exist. Will return nil before SetUp is called.
func (g *GameManager) AgentByName(name string) Agent {

	if !g.initialized {
		return nil
	}

	name = strings.ToLower(name)

	return g.agentsByName[name]
}

//MoveTypeByName returns the MoveType of that name from game.MoveTypes(), if
//it exists. Names are considered without regard to case.  Will return a copy.
func (g *GameManager) MoveTypeByName(name string) *MoveType {
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

//Chest is the ComponentChest in use for this game. Will return nil until
//SetUp() called.
func (g *GameManager) Chest() *ComponentChest {
	return g.chest
}

//Storage is the StorageManager games that use this manager should use.
func (g *GameManager) Storage() StorageManager {
	return g.storage
}

//Delegate returns the GameDelegate configured for these games.
func (g *GameManager) Delegate() GameDelegate {
	return g.delegate
}
