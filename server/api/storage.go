package api

import (
	"errors"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/seatpresentation"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/jkomoros/boardgame/server/api/users"
)

// Note: these are also duplicated in moves/seat_player.go
const playerToSeatRendevousDataType = "github.com/jkomoros/boardgame/server/api.PlayerToSeat"
const willSeatPlayerRendevousDataType = "github.com/jkomoros/boardgame/server/api.WillSeatPlayer"

// StorageManager extends the base boardgame.StorageManager with a few more
// methods necessary to make server work. When creating a new Server, you need
// to pass in a ServerStorageManager, which wraps one of these objects and thus
// implements these methods, too.
type StorageManager interface {

	//StorageManager extends the boardgame.StorageManager interface. Those
	//methods have two additional semantic expectations, however:
	//SaveGameAndCurrentState should create an ExtendedGameStorageRecord on
	//the first save of a game.
	boardgame.StorageManager

	//Name returns the name of the storage manager type, for example "memory", "bolt", or "mysql"
	Name() string

	//WithManagers is called during set up with references to all of the
	//managers. Will be called before Connect() is called.
	WithManagers(managers []*boardgame.GameManager)

	//Connect will be called before issuing any other substantive calls. The
	//config string is specific to the type of storage layer, which can be
	//interrogated with Nmae().
	Connect(config string) error

	//ExtendedGame is like Game(), but it returns an extended storage record
	//with additional fields necessary for Server.
	ExtendedGame(id string) (*extendedgame.StorageRecord, error)

	CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error)

	//UpdateExtendedGame updates the extended game with the given Id.
	UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error

	//Close should be called before the server is shut down.
	Close()

	//ListGames should list up to max games, in descending order based on the
	//LastActivity. If gameType is not "", only returns games that are that
	//gameType. If gameType is "", all gametypes are fine.
	ListGames(max int, list listing.Type, userID string, gameType string) []*extendedgame.CombinedStorageRecord

	//UserIDsForGame returns an array whose length equals game.NumPlayers.
	//Each one is either empty if there is no user in that slot yet, or the
	//uid representing the user.
	UserIDsForGame(gameID string) []string

	SetPlayerForGame(gameID string, playerIndex boardgame.PlayerIndex, userID string) error

	//GameByRoomCode looks up a gameID by its CompanionRoomCode (see
	//extendedgame.StorageRecord and spec §6.1). Returns "" and a nil error
	//if no live game has that code (caller should treat as 404). Codes
	//are exclusive while the game is active and for 24h after Finished;
	//implementations may choose to filter out games beyond the grace
	//period or rely on the caller to check Finished + Modified.
	GameByRoomCode(code string) (gameID string, err error)

	//SeatPresentation returns the per-(gameID, playerIndex) display name and
	//avatar set by the phone at join time. Returns nil with no error if no
	//row exists (e.g., solo-mode game or seat not yet claimed). See spec §5.4.
	SeatPresentation(gameID string, playerIndex boardgame.PlayerIndex) (*seatpresentation.StorageRecord, error)

	//SetSeatPresentation upserts the seat presentation for
	//(rec.GameID, rec.PlayerIndex).
	SetSeatPresentation(rec *seatpresentation.StorageRecord) error

	//ClearSeatPresentation removes any row for (gameID, playerIndex). No
	//error if no row exists. Called when a host frees a seat (V2) so the
	//next joiner doesn't inherit the prior phone's name + avatar.
	ClearSeatPresentation(gameID string, playerIndex boardgame.PlayerIndex) error

	// CompanionTableLease returns the durable lease for the game's shared
	// Table display, or nil when it has never had one.
	CompanionTableLease(gameID string) (*tablelease.StorageRecord, error)

	// CompareAndSwapCompanionTableLease atomically replaces the lease only
	// when its current generation equals expectedGeneration. A missing lease
	// has generation zero. Storage assigns expectedGeneration+1 to replacement
	// and returns a defensive copy of the resulting current record. This is the
	// cross-process arbitration boundary for claims, renewals, and releases.
	CompareAndSwapCompanionTableLease(gameID string, expectedGeneration uint64, replacement *tablelease.StorageRecord) (current *tablelease.StorageRecord, swapped bool, err error)

	//Store or update all fields
	UpdateUser(user *users.StorageRecord) error

	GetUserByID(uid string) *users.StorageRecord

	GetUserByCookie(cookie string) *users.StorageRecord

	//If user is nil, the cookie should be deleted if it exists. If the user
	//does not yet exist, it should be added to the database.
	ConnectCookieToUser(cookie string, user *users.StorageRecord) error

	//Note: whenever you add methods here, also add them to boardgame/storage/test/StorageManager
}

// ServerStorageManager implements the ServerStorage interface by wrapping an
// object that supports StorageManager.
type ServerStorageManager struct {
	StorageManager
	server *Server
}

// SaveChatMessage delegates to the underlying storage if it implements
// ChatStorageManager. This allows EmitSystemMessage to work through the
// ServerStorageManager wrapper.
func (s *ServerStorageManager) SaveChatMessage(msg *boardgame.ChatMessage) error {
	if cs, ok := s.StorageManager.(boardgame.ChatStorageManager); ok {
		return cs.SaveChatMessage(msg)
	}
	return nil // silently skip if chat not supported
}

// ChatMessages delegates to the underlying storage if it implements
// ChatStorageManager.
func (s *ServerStorageManager) ChatMessages(gameID string, channel string, sinceID string, limit int) ([]*boardgame.ChatMessage, error) {
	if cs, ok := s.StorageManager.(boardgame.ChatStorageManager); ok {
		return cs.ChatMessages(gameID, channel, sinceID, limit)
	}
	return nil, nil
}

// NewServerStorageManager takes an object that implements StorageManager and
// wraps it.
func NewServerStorageManager(manager StorageManager) *ServerStorageManager {
	return &ServerStorageManager{
		manager,
		nil,
	}
}

// PlayerMoveApplied notifies all clients connected vie an active WebSocket for
// that game that the game has been modified.
func (s *ServerStorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {

	//Do the wrapped manager's PlayerMoveApplied in case it has one.
	if err := s.StorageManager.PlayerMoveApplied(game); err != nil {
		return err
	}

	server := s.server

	if server == nil {
		return errors.New("no server configured. The storage manager should be added to a Server before it's used")
	}

	//Notify the web sockets that the game was changed
	server.notifier.gameChanged(game)

	// Check if seats have reopened (e.g., between rounds) and the game
	// should be made joinable again.
	server.maybeReopenGame(game)

	// Auto-emit system message when game finishes (once per game)
	server.mu.Lock()
	alreadyEmitted := server.gameOverEmitted[game.ID]
	if game.Finished && !alreadyEmitted {
		server.gameOverEmitted[game.ID] = true
	}
	server.mu.Unlock()
	if game.Finished && !alreadyEmitted {
		server.emitSystemChatMessage(game.ID, game.Version, "Game over!")
	}

	return nil

}

// emitSystemChatMessage saves a system chat message if chat storage is available.
func (s *Server) emitSystemChatMessage(gameID string, version int, body string) {
	cs := s.chatStorage()
	if cs == nil {
		return
	}
	msg := &boardgame.ChatMessage{
		GameID:    gameID,
		Version:   version,
		Sender:    boardgame.AdminPlayerIndex,
		Channel:   "all",
		Body:      body,
		Timestamp: time.Now(),
	}
	if err := cs.SaveChatMessage(msg); err != nil {
		s.logger.Warnln("Failed to emit system chat message:", err)
	}
}

// FetchInjectedDataForGame is where the server signals to SeatPlayer that
// there's a player to be seated.
func (s *ServerStorageManager) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	if dataType == willSeatPlayerRendevousDataType {
		//This data type should return anything non-nil to signal, yes, I am a
		//context that will pass you SeatPlayers when there's a player to seat.
		return true
	}
	if dataType == playerToSeatRendevousDataType {
		s.server.mu.Lock()
		slice := s.server.playersToSeat[gameID]
		var result interface{}
		if len(slice) > 0 {
			//The item's Committed() will remove itself from the list.
			result = slice[0]
		}
		s.server.mu.Unlock()
		if result != nil {
			return result
		}
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}
