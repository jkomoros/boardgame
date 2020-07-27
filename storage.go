package boardgame

import (
	"encoding/json"
	"errors"
	"time"
)

//StateStorageRecord is a record representing a state that can be written to
//storage and later returned. It is an encoded json blob, and can be written
//directly to storage with no modification. Typically you don't use this
//representation directly, instead fetching a game from the GameManager and
//then using State() for a fully-inflated state.
type StateStorageRecord json.RawMessage

//MoveStorageRecord is a record representing the Move that was made to get the
//game to its most recent version. It pops out various fields that
//StorageManagers could conceivably want to understand. Typically you don't
//use this directly, but instead fetch information for moves from game.Moves()
//and game.Move().
type MoveStorageRecord struct {
	Name      string
	Version   int
	Initiator int
	//The Phase as returned by Delegate.CurrentPhase() for the state the move
	//was in before it was applied. This is captured in this field because
	//moves in the moves package need to quickly inspect this value without
	//fully inflating the move structs.
	Phase int
	//The player index of the proposer of the move.
	Proposer  PlayerIndex
	Timestamp time.Time
	//The actual JSON serialized blob representing the properties of the move.
	Blob json.RawMessage
}

//Inflate takes a move storage record and turns it into a move associated with
//that game, if possible. Returns nil if not possible.
func (m *MoveStorageRecord) inflate(game *Game) (Move, error) {

	if game == nil {
		return nil, errors.New("Game was nil")
	}

	move := game.MoveByName(m.Name)

	if move == nil {
		return nil, errors.New("Couldn't find a move with name: " + m.Name)
	}

	if err := json.Unmarshal(m.Blob, move); err != nil {
		return nil, errors.New("Couldn't unmarshal move: " + err.Error())
	}

	move.Info().version = m.Version
	move.Info().initiator = m.Initiator
	move.Info().timestamp = m.Timestamp

	return move, nil
}

//GameStorageRecord is a simple struct with public fields representing the
//important aspects of a game that should be serialized to storage. The fields
//are broken out specifically so that the storage layer can understand these
//properties in queries. Typically you don't use this struct directly, instead
//getting an inflated version via something like GameManager.ModifiableGame()
//and then using the associated methods on the struct to get at the undelying
//values.
type GameStorageRecord struct {
	//Name is the type of the game, from its manager. Used for sanity
	//checking.
	Name string
	ID   string
	//SecretSalt for this game for things like component Ids. Should never be
	//transmitted to an insecure or untrusted environment; the only way to
	//access it outside this package is via this field, because it must be
	//able to be persisted to and read from storage.
	SecretSalt string `json:",omitempty"`
	Version    int
	Winners    []PlayerIndex
	Finished   bool
	Created    time.Time
	//Modified is updated every time a new move is applied.
	Modified time.Time
	//NumPlayers is the reported number of players when it was created.
	//Primarily for convenience to storage layer so they know how many players
	//are in the game.
	NumPlayers int
	Agents     []string
	Variant    Variant
}

//StorageManager is the interface that storage layers implement. The core
//engine expects one of these to be passed in via NewGameManager as the place
//to store and retrieve game information. A number of different
//implementations are available in boardgame/storage that can all be used.
//Typically you don't use this interface directly--it's defined just to
//formalize the interface between the core engine and the underlying storage
//layer.
type StorageManager interface {
	//State returns the StateStorageRecord for the game at the given version,
	//or nil.
	State(gameID string, version int) (StateStorageRecord, error)

	//Move returns the MoveStorageRecord for the game at the given version, or
	//nil.
	Move(gameID string, version int) (*MoveStorageRecord, error)

	//Moves is like Move but returns all moves from fromVersion (exclusive) to
	//toVersion (inclusive). If fromVersion == toVersion, should return
	//toVersion. In many storage subsystems this is cheaper than repeated
	//calls to Move, which is why it's broken out separately.
	Moves(gameID string, fromVersion, toVersion int) ([]*MoveStorageRecord, error)

	//Game fetches the GameStorageRecord with the given ID from the store, if
	//it exists.
	Game(id string) (*GameStorageRecord, error)

	//AgentState retrieves the most recent state for the given agent
	AgentState(gameID string, player PlayerIndex) ([]byte, error)

	//SaveGameAndCurrentState stores the game and the current state (at
	//game.Version()) into the store at the same time in a transaction. Move
	//is normally provided but will be be nil if game.Version() is 0, denoting
	//the initial state for a game.
	SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error

	//SaveAgentState saves the agent state for the given player
	SaveAgentState(gameID string, player PlayerIndex, state []byte) error

	//PlayerMoveApplied is called after a PlayerMove and all of its resulting
	//FixUp moves have been applied. Most StorageManagers don't need to do
	//anything here; it's primarily useful as a callback to signal that a run
	//of moves has been applied, e.g. in the server.
	PlayerMoveApplied(game *GameStorageRecord) error

	//FetchInjectedDataForGame is an override point for other layers to inject
	//triggers for bits of game logic to call into. dataType should be the name
	//of the package that publishes the data type, to avoid collissions (for
	//example, 'github.com/jkomoros/boardgame/server/api.PlayerToSeat'). Things,
	//like server, will override this method to add new data types. Base storage
	//managers need only return nil in all cases.
	FetchInjectedDataForGame(gameID string, dataType string) interface{}
}
