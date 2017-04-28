package boardgame

import (
	"encoding/json"
	"errors"
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
	delegate            GameDelegate
	chest               *ComponentChest
	storage             StorageManager
	agents              []Agent
	fixUpMoves          []MoveFactory
	playerMoves         []MoveFactory
	fixUpMovesByName    map[string]MoveFactory
	playerMovesByName   map[string]MoveFactory
	agentsByName        map[string]Agent
	modifiableGamesLock sync.RWMutex
	modifiableGames     map[string]*Game
	timers              *timerManager
	initialized         bool
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
		delegate: delegate,
		chest:    chest,
		storage:  storage,
	}

	chest.manager = result

	delegate.SetManager(result)

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
		finished:   record.Finished,
		winners:    record.Winners,
		numPlayers: record.NumPlayers,
		agents:     record.Agents,
		modifiable: false,
		initalized: true,
	}
}

//modifiableGameCreated lets Manager know that a modifiable game was created
//with the given ID, so that manager can vend that later if necessary. It is
//designed to only be called from NewGame.
func (g *GameManager) modifiableGameCreated(game *Game) {
	if !g.initialized {
		return
	}

	g.modifiableGamesLock.RLock()
	_, ok := g.modifiableGames[game.Id()]
	g.modifiableGamesLock.RUnlock()

	if ok {
		panic("modifiableGameCreated collided with existing game")
	}

	id := strings.ToUpper(game.Id())

	g.modifiableGamesLock.Lock()
	g.modifiableGames[id] = game
	g.modifiableGamesLock.Unlock()
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
	Game       json.RawMessage
	Players    []json.RawMessage
	Components map[string][]json.RawMessage
}

//verifyReaderStacks goes through each property in Reader that is a stack or
//timer, and verifies that it is non-nil, and its state property is set to the
//given state.
func verifyReaderObjects(reader PropertyReader, state *state) error {
	for propName, propType := range reader.Props() {
		switch propType {
		case TypeGrowableStack:
			val, err := reader.GrowableStackProp(propName)
			if val == nil {
				return errors.New("GrowableStack Prop " + propName + " was nil")
			}
			if err != nil {
				return errors.New("GrowableStack prop " + propName + " had unexpected error: " + err.Error())
			}
			val.statePtr = state
		case TypeSizedStack:
			val, err := reader.SizedStackProp(propName)
			if val == nil {
				return errors.New("SizedStackProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("SizedStack prop " + propName + " had unexpected error: " + err.Error())
			}
			val.statePtr = state
		case TypeTimer:
			val, err := reader.TimerProp(propName)
			if val == nil {
				return errors.New("TimerProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("TimerProp " + propName + " had unexpected error: " + err.Error())
			}
			val.statePtr = state
		}
	}
	return nil
}

//emptyPlayerState is a simple wrapper around delegate.EmptyPlayerState that
//just verifies that stacks are inflated.
func (g *GameManager) emptyPlayerState(state *state, player PlayerIndex) (MutablePlayerState, error) {

	playerState := g.delegate.EmptyPlayerState(player)

	if playerState == nil {
		return nil, errors.New("EmptyPlayerState returned nil for " + strconv.Itoa(int(player)))
	}

	if err := verifyReaderObjects(playerState.Reader(), state); err != nil {
		return nil, err
	}

	return playerState, nil

}

//emptyGameState is a simple wrapper around delegate.EmptyPlayerState that
//just verifies that stacks are inflated.
func (g *GameManager) emptyGameState(state *state) (MutableBaseState, error) {

	gameState := g.delegate.EmptyGameState()

	if gameState == nil {
		return nil, errors.New("EmptyGameState returned nil")
	}

	if err := verifyReaderObjects(gameState.Reader(), state); err != nil {
		return nil, err
	}

	return gameState, nil

}

func (g *GameManager) emptyDynamicComponentValues(state *state) (map[string][]MutableDynamicComponentValues, error) {
	result := make(map[string][]MutableDynamicComponentValues)

	for _, deckName := range g.Chest().DeckNames() {

		deck := g.Chest().Deck(deckName)

		if deck == nil {
			return nil, errors.New("Couldn't find deck for " + deckName)
		}

		values := g.Delegate().EmptyDynamicComponentValues(deck)
		if values == nil {
			continue
		}
		arr := make([]MutableDynamicComponentValues, len(deck.Components()))
		for i := 0; i < len(deck.Components()); i++ {
			arr[i] = values.Copy()
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

	result := &state{}

	game, err := g.emptyGameState(result)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(refried.Game, game); err != nil {
		return nil, errors.New("Unmarshal of GameState failed: " + err.Error())
	}

	for propName, propType := range game.Reader().Props() {
		switch propType {
		case TypeSizedStack:
			stack, err := game.Reader().SizedStackProp(propName)
			if err != nil {
				return nil, errors.New("Unable to inflate stack " + propName + " in game.")
			}
			stack.Inflate(g.Chest())
		case TypeGrowableStack:
			stack, err := game.Reader().GrowableStackProp(propName)
			if err != nil {
				return nil, errors.New("Unable to inflate stack " + propName + " in game.")
			}
			stack.Inflate(g.Chest())
		}
	}

	result.gameState = game

	for i, blob := range refried.Players {
		player, err := g.emptyPlayerState(result, PlayerIndex(i))

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(blob, player); err != nil {
			return nil, errors.New("Unmarshal into player state failed for " + strconv.Itoa(i) + " player: " + err.Error())
		}

		for propName, propType := range player.Reader().Props() {
			switch propType {
			case TypeSizedStack:
				stack, err := player.Reader().SizedStackProp(propName)
				if err != nil {
					return nil, errors.New("Unable to inflate stack " + propName + " in player " + strconv.Itoa(i))
				}
				stack.Inflate(g.Chest())
			case TypeGrowableStack:
				stack, err := player.Reader().GrowableStackProp(propName)
				if err != nil {
					return nil, errors.New("Unable to inflate stack " + propName + " in player " + strconv.Itoa(i))
				}
				stack.Inflate(g.Chest())
			}
		}

		result.playerStates = append(result.playerStates, player)
	}

	dynamic, err := g.emptyDynamicComponentValues(result)

	if err != nil {
		return nil, errors.New("Couldn't create empty dynamic component values: " + err.Error())
	}

	result.dynamicComponentValues = dynamic

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
				case TypeSizedStack:
					stack, err := resultDeckValue.Reader().SizedStackProp(propName)
					if err != nil {
						return nil, errors.New("Unable to inflate stack " + propName + " in deck " + deckName + " component " + strconv.Itoa(i))
					}
					stack.Inflate(g.Chest())
				case TypeGrowableStack:
					stack, err := resultDeckValue.Reader().GrowableStackProp(propName)
					if err != nil {
						return nil, errors.New("Unable to inflate stack " + propName + " in deck " + deckName + " component " + strconv.Itoa(i))
					}
					stack.Inflate(g.Chest())
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

	//Verify that the shape of the computed property collections fits with the config.
	if config := g.delegate.ComputedPropertiesConfig(); config != nil {

		if global := config.Global; global != nil {
			if collection := g.delegate.EmptyComputedGlobalPropertyCollection(); collection != nil {
				//Verify the shape has slots for all of the configed properties
				for propName, propConfig := range global {
					propType := propConfig.PropType
					if collection.Reader().Props()[propName] != propType {
						return errors.New("The global property collection the delegate returns has a mismatch for property " + propName)
					}
				}
			}
		}

		if player := config.Player; player != nil {
			if collection := g.delegate.EmptyComputedPlayerPropertyCollection(); collection != nil {
				//Verify the shape has slots for all of the configed properties
				for propName, propConfig := range player {
					propType := propConfig.PropType
					if collection.Reader().Props()[propName] != propType {
						return errors.New("The global property collection the delegate returns has a mismatch for property " + propName)
					}
				}
			}
		}

	}

	g.agentsByName = make(map[string]Agent)
	for _, agent := range g.agents {
		g.agentsByName[strings.ToLower(agent.Name())] = agent
	}

	g.playerMovesByName = make(map[string]MoveFactory)
	for _, factory := range g.playerMoves {
		sampleMove := factory(nil)
		g.playerMovesByName[strings.ToLower(sampleMove.Name())] = factory
	}

	g.fixUpMovesByName = make(map[string]MoveFactory)
	for _, factory := range g.fixUpMoves {
		sampleMove := factory(nil)
		g.fixUpMovesByName[strings.ToLower(sampleMove.Name())] = factory
	}

	g.modifiableGames = make(map[string]*Game)

	g.timers = newTimerManager()

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

//AddAgent is called before set up to configure an agent that is available to
//play in games.
func (g *GameManager) AddAgent(agent Agent) {
	if g.initialized {
		return
	}
	g.agents = append(g.agents, agent)
}

//AddPlayerMoveFactories adds the specified move factory to the game as a move
//that Players can make. It may only be called during initalization.
func (g *GameManager) AddPlayerMoveFactory(factory MoveFactory) {

	if g.initialized {
		return
	}
	g.playerMoves = append(g.playerMoves, factory)
}

//AddFixUpMoveFactory adds a move factory that can only be legally made by
//GameDelegate as a FixUp move. It can only be called during initialization.
func (g *GameManager) AddFixUpMoveFactory(factory MoveFactory) {
	if g.initialized {
		return
	}
	g.fixUpMoves = append(g.fixUpMoves, factory)
}

//Agents returns a slice of all agents configured on this Manager. Will return
//nil before SetUp is called.
func (g *GameManager) Agents() []Agent {
	if !g.initialized {
		return nil
	}

	return g.agents
}

//PlayerMoveFactories returns all moves that are valid in this game to be made my
//players--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called.
func (g *GameManager) PlayerMoveFactories() []MoveFactory {
	if !g.initialized {
		return nil
	}

	return g.playerMoves
}

//FixUpMoveFactoriess returns all move factories that are valid in this game
//to be made as fixup moves--all of the Moves that have been added via
//AddPlayerMove  during initalization. Returns nil until game.SetUp() has been
//called.
func (g *GameManager) FixUpMoveFactories() []MoveFactory {

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

//PlayerMoveByName returns the MoveFactory of that name from
//game.PlayerMoves(), if it exists. Names are considered without regard to
//case.  Will return a copy.
func (g *GameManager) PlayerMoveFactoryByName(name string) MoveFactory {
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

//FixUpMoveFactoryByName returns the MoveFactory of that name from
//game.FixUpMoves(), if it exists. Names are considered without regard to
//case.  Will return a copy.
func (g *GameManager) FixUpMoveFactoryByName(name string) MoveFactory {
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
