package vecgame

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

// Harness boots the vector game on an in-memory store.
type Harness struct {
	Delegate *Delegate
	Manager  *boardgame.GameManager
	Storage  *memory.StorageManager
	counter  int
}

func NewHarness() (*Harness, error) {
	delegate := &Delegate{}
	storage := memory.NewStorageManager()
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		return nil, err
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	return &Harness{Delegate: delegate, Manager: manager, Storage: storage}, nil
}

// FreshBlob runs the real set-up for a new game and returns its version-0
// state as the storage layer would persist it.
func (h *Harness) FreshBlob(id, salt string, numPlayers int) (boardgame.StateStorageRecord, error) {
	game, err := h.Manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name:       h.Delegate.Name(),
		ID:         id,
		SecretSalt: salt,
		NumPlayers: numPlayers,
	})
	if err != nil {
		return nil, err
	}
	return Blob(game.CurrentState())
}

// StateAt stores blob as the state of game (id, salt) at exactly version, then
// loads it back through the manager the way the server does, and returns a
// mutable copy. The copy keeps the version and the game's salt, so ids and
// state.Rand() are the real ones for that game at that version.
func (h *Harness) StateAt(id, salt string, version, numPlayers int, blob []byte) (*boardgame.Game, boardgame.State, error) {
	rec := &boardgame.GameStorageRecord{
		Name:       h.Delegate.Name(),
		ID:         id,
		SecretSalt: salt,
		NumPlayers: numPlayers,
		Version:    version,
	}
	if err := h.Storage.SaveGameAndCurrentState(rec, blob, nil); err != nil {
		return nil, nil, err
	}
	game := h.Manager.Game(id)
	if game == nil {
		return nil, nil, errors.New("game did not load")
	}
	current := game.CurrentState()
	if current == nil {
		return nil, nil, errors.New("state did not load")
	}
	copied, err := current.Copy(false)
	if err != nil {
		return nil, nil, err
	}
	mutable, ok := copied.(boardgame.State)
	if !ok {
		return nil, nil, errors.New("copy was not mutable")
	}
	return game, mutable, nil
}

// Blob is the state's storage record, compacted.
func Blob(state boardgame.ImmutableState) (boardgame.StateStorageRecord, error) {
	rec := state.StorageRecord()
	var generic interface{}
	if err := json.Unmarshal(rec, &generic); err != nil {
		return nil, err
	}
	return json.Marshal(generic)
}

// UniqueID returns a fresh game id with the given prefix.
func (h *Harness) UniqueID(prefix string) string {
	h.counter++
	const hex = "0123456789ABCDEF"
	n := h.counter
	out := ""
	for i := 0; i < 6; i++ {
		out = string(hex[n%16]) + out
		n /= 16
	}
	return prefix + out
}
