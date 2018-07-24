/*

	helpers has generic implementations of finicky methods, like Moves(),
	ListGames() that are appropriate for storage managers who don't get a
	performance boost from well-crafted queries to use.

*/
package helpers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"sort"
)

type AllGamesStorageManager interface {
	api.StorageManager
	//AllGames simply returns all games
	AllGames() []*boardgame.GameStorageRecord
}

//MovesHelper is an implementation for Moves() if the underlying storage
//manager can't do any better than just repeated calls to Move() anyway.
func MovesHelper(s boardgame.StorageManager, gameId string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {

	//There's no efficiency boost for fetching multiple moves at once so just wrap around Move()

	if fromVersion == toVersion {
		fromVersion = fromVersion - 1
	}

	result := make([]*boardgame.MoveStorageRecord, toVersion-fromVersion)

	index := 0
	for i := fromVersion + 1; i <= toVersion; i++ {
		move, err := s.Move(gameId, i)
		if err != nil {
			return nil, err
		}
		result[index] = move
		index++
	}
	return result, nil
}

//ListGamesHelper is an implementation for ListGames() if the underlying
//storage manager can't do any better than walking through each game anyway.
//Note that your StorageManager must implement AllGames().
func ListGamesHelper(s AllGamesStorageManager, max int, list listing.Type, userId string, gameType string) []*extendedgame.CombinedStorageRecord {

	if (list == listing.ParticipatingActive || list == listing.ParticipatingFinished) && userId == "" {
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

		eGame, _ := s.ExtendedGame(game.Id)

		usersForGame := s.UserIdsForGame(game.Id)

		hasUser := false
		numUsers := 0

		for _, user := range usersForGame {
			if user != "" {
				numUsers++
			}
			if userId != "" && user == userId {
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
