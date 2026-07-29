package api

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/workfit/tester/assert"
)

type unmanagedGameStorage struct {
	StorageManager
	games []*extendedgame.CombinedStorageRecord
}

func (s *unmanagedGameStorage) ListGames(
	max int, list listing.Type, userID string, gameType string,
) []*extendedgame.CombinedStorageRecord {
	return s.games
}

// A stored game whose type this server does not manage must not take the whole
// game list down with it.
//
// `managers` is built from config.json at startup; storage is not, and outlives
// every edit to that list. Removing a game from config.json -- or pointing a
// second server with a different list at the same database, which is what a dev
// machine does routinely -- leaves records behind that map to no manager. The
// unguarded `s.managers[game.Name].manager` dereferenced a nil *managerInfo and
// panicked, and because gin's recovery unwinds the whole handler the failure was
// never scoped to the one game: /api/list/game returned HTTP 500 and every page
// that lists games showed an error dialog instead. Observed on a real dev server
// with eight such records: 190 panics, and the app was unusable until they aged
// out of the hundred-most-recent window.
//
// The contract asserted here is that the unmanaged game is still LISTED, with no
// player info -- which is what gamePlayerInfo, alone on this path, was already
// written to do for a nil manager.
func TestListGamesSurvivesAGameTypeThisServerDoesNotManage(t *testing.T) {
	storage := &unmanagedGameStorage{games: []*extendedgame.CombinedStorageRecord{
		{
			GameStorageRecord: boardgame.GameStorageRecord{
				ID:         "a-game-of-a-retired-type",
				Name:       "no-such-game-type",
				NumPlayers: 2,
				Agents:     []string{"", ""},
				SecretSalt: "must-not-leak",
			},
		},
	}}
	server := &Server{
		storage: NewServerStorageManager(storage),
		// Deliberately non-nil and deliberately not containing the game's type:
		// an empty map and a populated one both return nil here, and it was the
		// nil VALUE, not a nil map, that crashed.
		managers: managerMap{"some-other-game": &managerInfo{}},
	}

	result := server.listGamesWithUsers(100, listing.All, "", "")

	assert.For(t).ThatActual(len(result)).Equals(1)
	assert.For(t).ThatActual(result[0].Name).Equals("no-such-game-type")
	assert.For(t).ThatActual(len(result[0].Players)).Equals(0)
	// The sanitization on this path still has to happen for a game nobody can
	// render, or an unmanaged record would be the one that leaked its salt.
	assert.For(t).ThatActual(result[0].SecretSalt).Equals("")
}
