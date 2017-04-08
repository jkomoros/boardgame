package boardgame

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
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
	fixUpMoves          []Move
	playerMoves         []Move
	fixUpMovesByName    map[string]Move
	playerMovesByName   map[string]Move
	modifiableGamesLock sync.RWMutex
	modifiableGames     map[string]*Game
	initialized         bool
}

//ManagerCollection is a way to grab a reference to a specific manager based
//on a game name. It's passed in to some storage methods by server and others.
type ManagerCollection interface {
	Get(name string) *GameManager
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

//verifyReaderStacks goes through each property in Reader that is a stack, and
//verifies that it is non-nil, and its state property is set to the given
//state.
func verifyReaderStacks(reader PropertyReader, state *state) error {
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
		}
	}
	return nil
}

//emptyPlayerState is a simple wrapper around delegate.EmptyPlayerState that
//just verifies that stacks are inflated.
func (g *GameManager) emptyPlayerState(state *state, playerIndex int) (MutablePlayerState, error) {

	playerState := g.delegate.EmptyPlayerState(playerIndex)

	if playerState == nil {
		return nil, errors.New("EmptyPlayerState returned nil for " + strconv.Itoa(playerIndex))
	}

	if err := verifyReaderStacks(playerState.Reader(), state); err != nil {
		return nil, err
	}

	return playerState, nil

}

//emptyGameState is a simple wrapper around delegate.EmptyPlayerState that
//just verifies that stacks are inflated.
func (g *GameManager) emptyGameState(state *state) (MutableGameState, error) {

	gameState := g.delegate.EmptyGameState()

	if gameState == nil {
		return nil, errors.New("EmptyGameState returned nil")
	}

	if err := verifyReaderStacks(gameState.Reader(), state); err != nil {
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
//given a serialized state blob.
func (g *GameManager) StateFromBlob(blob []byte) (State, error) {
	//At this point, no extra state is stored in the blob other than in props.

	//We can't just delegate to StateProps to unmarshal itself, because it
	//needs a reference to delegate to inflate, and only we have that.
	var refried refriedState

	if err := json.Unmarshal(blob, &refried); err != nil {
		return nil, err
	}

	result := &state{
		manager: g,
	}

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

	result.game = game

	for i, blob := range refried.Players {
		player, err := g.emptyPlayerState(result, i)

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

		result.players = append(result.players, player)
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
func (g *GameManager) proposeMoveOnGame(id string, move Move) DelayedError {

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
			move: move,
			ch:   errChan,
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

	g.playerMovesByName = make(map[string]Move)
	for _, move := range g.playerMoves {
		g.playerMovesByName[strings.ToLower(move.Name())] = move
	}

	g.fixUpMovesByName = make(map[string]Move)
	for _, move := range g.fixUpMoves {
		g.fixUpMovesByName[strings.ToLower(move.Name())] = move
	}

	g.modifiableGames = make(map[string]*Game)

	g.initialized = true

	return nil
}

//AddPlayerMove adds the specified move to the game as a move that Players can
//make. It may only be called during initalization.
func (g *GameManager) AddPlayerMove(move Move) {

	if g.initialized {
		return
	}
	g.playerMoves = append(g.playerMoves, move)
}

//AddFixUpMove adds a move that can only be legally made by GameDelegate as a
//FixUp move. It can only be called during initialization.
func (g *GameManager) AddFixUpMove(move Move) {
	if g.initialized {
		return
	}
	g.fixUpMoves = append(g.fixUpMoves, move)
}

//PlayerMoves returns all moves that are valid in this game to be made my
//players--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies.
func (g *GameManager) PlayerMoves() []Move {
	if !g.initialized {
		return nil
	}

	result := make([]Move, len(g.playerMoves))

	for i, move := range g.playerMoves {
		result[i] = move.Copy()
	}

	return result
}

//FixUpMoves returns all moves that are valid in this game to be made as fixup
//moves--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies.
func (g *GameManager) FixUpMoves() []Move {

	//TODO: test all of these fixup moves

	if !g.initialized {
		return nil
	}

	result := make([]Move, len(g.fixUpMoves))

	for i, move := range g.fixUpMoves {
		result[i] = move.Copy()
	}

	return result
}

//PlayerMoveByName returns the Move of that name from game.PlayerMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy.
func (g *GameManager) PlayerMoveByName(name string) Move {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.playerMovesByName[name]

	if move == nil {
		return nil
	}

	return move.Copy()
}

//FixUpMoveByName returns the Move of that name from game.FixUpMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy.
func (g *GameManager) FixUpMoveByName(name string) Move {
	if !g.initialized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.fixUpMovesByName[name]

	if move == nil {
		return nil
	}

	return move.Copy()
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
