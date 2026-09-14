/*
Package helpers has generic implementations of finicky methods, like Moves(),
ListGames() that are appropriate for storage managers who don't get a
performance boost from well-crafted queries to use.
*/
package helpers

import (
	"errors"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
)

// AllGamesStorageManager wraps normal storage manager but also has an AllGames
// method.
type AllGamesStorageManager interface {
	api.StorageManager
	//AllGames simply returns all games
	AllGames() []*boardgame.GameStorageRecord
}

// TimerWakeupsHelper discovers active durable timers by scanning game-head
// records and parsing only their stored timer metadata. It never inflates a
// boardgame.Game or adds one to a manager's warm cache.
func TimerWakeupsHelper(s AllGamesStorageManager, gameName string) ([]boardgame.TimerWakeup, error) {
	if s == nil {
		return nil, errors.New("nil storage manager")
	}
	var result []boardgame.TimerWakeup
	for _, game := range s.AllGames() {
		if game == nil || game.Name != gameName || game.Finished {
			continue
		}
		state, err := s.State(game.ID, game.Version)
		if err != nil {
			return nil, err
		}
		wakeups, err := boardgame.TimerWakeupsFromStateStorage(game.ID, state)
		if err != nil {
			return nil, err
		}
		result = append(result, wakeups...)
	}
	return result, nil
}

// MovesHelper is an implementation for Moves() if the underlying storage
// manager can't do any better than just repeated calls to Move() anyway.
func MovesHelper(s boardgame.StorageManager, gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {

	//There's no efficiency boost for fetching multiple moves at once so just wrap around Move()

	if fromVersion == toVersion {
		fromVersion = fromVersion - 1
	}

	result := make([]*boardgame.MoveStorageRecord, toVersion-fromVersion)

	index := 0
	for i := fromVersion + 1; i <= toVersion; i++ {
		move, err := s.Move(gameID, i)
		if err != nil {
			return nil, err
		}
		result[index] = move
		index++
	}
	return result, nil
}

// ListGamesHelper is an implementation for ListGames() if the underlying
// storage manager can't do any better than walking through each game anyway.
// Note that your StorageManager must implement AllGames().
func ListGamesHelper(s AllGamesStorageManager, max int, list listing.Type, userID string, gameType string) []*extendedgame.CombinedStorageRecord {

	if (list == listing.ParticipatingActive || list == listing.ParticipatingFinished) && userID == "" {
		//If we're filtering to only participating games and there's no userId, then there can't be any games,
		//because the non-user can't be participating in any games.
		return nil
	}

	var result []*extendedgame.CombinedStorageRecord

	for _, game := range s.AllGames() {

		if gameType != "" {
			if game.Name != gameType {
				continue
			}
		}

		eGame, _ := s.ExtendedGame(game.ID)

		usersForGame := s.UserIDsForGame(game.ID)

		hasUser := false
		numUsers := 0

		for _, user := range usersForGame {
			if user != "" {
				numUsers++
			}
			if userID != "" && user == userID {
				hasUser = true
				break
			}
		}

		numAgents := 0

		for _, agent := range game.Agents {
			if agent != "" {
				numAgents++
			}
		}

		hasSlots := game.NumPlayers > (numUsers + numAgents)

		switch list {
		case listing.ParticipatingActive:
			if game.Finished || !hasUser {
				continue
			}
		case listing.ParticipatingFinished:
			if !game.Finished || !hasUser {
				continue
			}
		case listing.VisibleJoinableActive:
			if game.Finished || hasUser || !eGame.Visible || !eGame.Open || !hasSlots {
				continue
			}
		case listing.VisibleActive:
			if game.Finished || hasUser || !eGame.Visible || (eGame.Open && hasSlots) {
				continue
			}
		}

		result = append(result, &extendedgame.CombinedStorageRecord{
			GameStorageRecord: *game,
			StorageRecord:     *eGame,
		})

		if len(result) >= max {
			break
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Modified.After(result[j].Modified)
	})

	return result
}
