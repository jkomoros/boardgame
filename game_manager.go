package boardgame

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/jkomoros/boardgame/errors"
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
	fixUpMoves                []*MoveType
	playerMoves               []*MoveType
	fixUpMovesByName          map[string]*MoveType
	playerMovesByName         map[string]*MoveType
	agentsByName              map[string]Agent
	modifiableGamesLock       sync.RWMutex
	modifiableGames           map[string]*Game
	timers                    *timerManager
	initialized               bool
	logger                    *logrus.Logger
}

//NewGameManager creates a new game manager with the given delegate.
func NewGameManager(delegate GameDelegate, chest *ComponentChest, storage StorageManager) *GameManager {
	if delegate == nil {
		return nil
	}

	if chest == nil {
		return nil
	}

	//Make sure the chest is no longer open for modification. If finish was
	//already called, this will be a no-op.
	chest.Finish()

	if storage == nil {
		return nil
	}

	result := &GameManager{
		delegate:          delegate,
		chest:             chest,
		storage:           storage,
		logger:            logrus.New(),
		fixUpMovesByName:  make(map[string]*MoveType),
		playerMovesByName: make(map[string]*MoveType),
	}

	chest.manager = result

	delegate.SetManager(result)

	return result
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

	if g == nil {
		return nil
	}

	result := &Game{
		manager: g,
		//TODO: set the size of chan based on something more reasonable.
		//Note: this is also set similarly in manager.ModifiableGame
		proposedMoves: make(chan *proposedMoveItem, 20),
		id:            randomString(gameIDLength),
		secretSalt:    randomString(gameIDLength),
		modifiable:    true,
	}

	if err := g.modifiableGameCreated(result); err != nil {
		g.logger.Warn("Couldn't warn that a modifiable game was created: " + err.Error())
		return nil
	}

	return result

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

	reader := playerState.Reader()

	if reader == nil {
		return nil, errors.New("PlayerState Reader returned nil")
	}

	if err := g.playerValidator.Valid(reader); err != nil {
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

	reader := gameState.Reader()

	if reader == nil {
		return nil, errors.New("GameState reader returned nil")
	}

	if err := g.gameValidator.Valid(reader); err != nil {
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

			reader := arr[i].Reader()

			if reader == nil {
				return nil, errors.New("Reader for dynamic component values for " + deckName + " " + strconv.Itoa(i) + " was nil")
			}

			if err := validator.Valid(reader); err != nil {
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

	result := &state{
		secretMoveCount: refried.SecretMoveCount,
		version:         refried.Version,
	}

	if result.secretMoveCount == nil {
		result.secretMoveCount = make(map[string][]int)
	}

	game, err := g.gameStateConstructor(result)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(refried.Game, game); err != nil {
		return nil, errors.New("Unmarshal of GameState failed: " + err.Error())
	}

	for propName, propType := range game.Reader().Props() {
		switch propType {
		case TypeStack:
			stack, err := game.Reader().StackProp(propName)
			if err != nil {
				return nil, errors.New("Unable to inflate stack " + propName + " in game.")
			}
			stack.inflate(g.Chest())
		}
	}

	result.gameState = game

	for i, blob := range refried.Players {
		player, err := g.playerStateConstructor(result, PlayerIndex(i))

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(blob, player); err != nil {
			return nil, errors.New("Unmarshal into player state failed for " + strconv.Itoa(i) + " player: " + err.Error())
		}

		for propName, propType := range player.Reader().Props() {
			switch propType {
			case TypeStack:
				stack, err := player.Reader().StackProp(propName)
				if err != nil {
					return nil, errors.New("Unable to inflate stack " + propName + " in player " + strconv.Itoa(i))
				}
				stack.inflate(g.Chest())
			}
		}

		result.playerStates = append(result.playerStates, player)
	}

	dynamic, err := g.dynamicComponentValuesConstructor(result)

	if err != nil {
		return nil, errors.New("Couldn't create empty dynamic component values: " + err.Error())
	}

	result.dynamicComponentValues = dynamic

	result.setStateForSubStates()

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

			for propName, propType := range resultDeckValue.Reader().Props() {
				switch propType {
				case TypeStack:
					stack, err := resultDeckValue.Reader().StackProp(propName)
					if err != nil {
						return nil, errors.New("Unable to inflate stack " + propName + " in deck " + deckName + " component " + strconv.Itoa(i))
					}
					stack.inflate(g.Chest())
				}
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

//SetUp should be called before this Manager is used. It locks in moves,
//chest, storage, etc.
func (g *GameManager) SetUp() error {

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

	validator, err := newReaderValidator(reader, exampleGameState, nil, g.chest, false)

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

	validator, err = newReaderValidator(reader, examplePlayerState, nil, g.chest, true)

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

		validator, err = newReaderValidator(reader, exampleDynamicComponentValue, nil, g.chest, false)

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

	for _, moveType := range g.fixUpMoves {
		//We don't create a real move because we just need to run the
		//ValidConfiguration(), and creating a whole fake state is a pain.
		testMove := moveType.constructor()

		if err := testMove.ValidConfiguration(); err != nil {
			return errors.New(moveType.Name() + " move failed the ValidConfiguration test: " + err.Error())
		}

	}

	for _, moveType := range g.playerMoves {
		//We don't create a real move because we just need to run the
		//ValidConfiguration(), and creating a whole fake state is a pain.
		testMove := moveType.constructor()

		if err := testMove.ValidConfiguration(); err != nil {
			return errors.New(moveType.Name() + " move failed the ValidConfiguration test: " + err.Error())
		}

	}

	g.agentsByName = make(map[string]Agent)
	for _, agent := range g.agents {
		g.agentsByName[strings.ToLower(agent.Name())] = agent
	}

	g.modifiableGames = make(map[string]*Game)

	g.timers = newTimerManager(g)

	//Start ticking timers.
	go func() {
		//TODO: is there a way to turn off timer ticking for a manager we want
		//to throw out?
		for {
			<-time.After(250 * time.Millisecond)
			g.timers.Tick()
		}
	}()

	g.initialized = true

	return nil
}

//BulkAddMoveTypes is a convenience wrapper around AddPlayerMoveType and
//AddFixUpMoveType. Will error if any of the configs do not produce a valid
//MoveType.
func (g *GameManager) BulkAddMoveTypes(moveTypeConfigs []*MoveTypeConfig) error {

	if moveTypeConfigs == nil {
		return errors.New("No moveTypeConfigs provided")
	}

	for i, config := range moveTypeConfigs {
		if err := g.AddMoveType(config); err != nil {
			return errors.New("Config " + strconv.Itoa(i) + " failed with error: " + err.Error())
		}
	}

	return nil

}

//AddAgent is called before set up to configure an agent that is available to
//play in games.
func (g *GameManager) AddAgent(agent Agent) {
	if g.initialized {
		return
	}
	g.agents = append(g.agents, agent)
}

//AddMoveType adds the specified move type to the game as a move. It may only
//be called during initalization.
func (g *GameManager) AddMoveType(config *MoveTypeConfig) error {

	moveType, err := newMoveType(config, g)

	if err != nil {
		return err
	}

	if g.initialized {
		return errors.New("GameManager has already been SetUp so no new moves may be added.")
	}

	moveName := strings.ToLower(moveType.Name())

	if moveType.IsFixUp() {

		if g.fixUpMovesByName[moveName] != nil {
			//If it's already been added that's OK
			return nil
		}

		g.fixUpMoves = append(g.fixUpMoves, moveType)
		g.fixUpMovesByName[moveName] = moveType
	} else {

		if g.playerMovesByName[moveName] != nil {
			//If it's already been added that's OK
			return nil
		}

		g.playerMoves = append(g.playerMoves, moveType)
		g.playerMovesByName[moveName] = moveType
	}

	return nil
}

//Agents returns a slice of all agents configured on this Manager. Will return
//nil before SetUp is called.
func (g *GameManager) Agents() []Agent {
	if !g.initialized {
		return nil
	}

	return g.agents
}

//PlayerMoves returns all moves that are valid in this game to be made my
//players--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called.
func (g *GameManager) PlayerMoveTypes() []*MoveType {
	if !g.initialized {
		return nil
	}

	return g.playerMoves
}

//FixUpMoveTypes returns all move types that are valid in this game
//to be made as fixup moves--all of the Moves that have been added via
//AddPlayerMove  during initalization. Returns nil until game.SetUp() has been
//called.
func (g *GameManager) FixUpMoveTypes() []*MoveType {

	if !g.initialized {
		return nil
	}

	return g.fixUpMoves
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

//PlayerMoveByName returns the MoveType of that name from
//game.PlayerMoves(), if it exists. Names are considered without regard to
//case.  Will return a copy.
func (g *GameManager) PlayerMoveTypeByName(name string) *MoveType {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.playerMovesByName[name]

	if move == nil {
		return nil
	}

	return move
}

//FixUpMoveTypeByName returns the MoveType of that name from
//game.FixUpMoves(), if it exists. Names are considered without regard to
//case.  Will return a copy.
func (g *GameManager) FixUpMoveTypeByName(name string) *MoveType {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.fixUpMovesByName[name]

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
