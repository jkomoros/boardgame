/*
Package seatpresentation is the storage record for the per-(gameID,
playerIndex) display-name + avatar that a phone establishes when it joins a
Table+Hand companion-mode game. Kept in a separate package (parallel to users
and extendedgame) to avoid dependency cycles between storage backends and the
record definition.

Per-seat-not-per-user is deliberate (see spec §5.4): joining game 2 with a
different avatar must not mutate game 1's presentation, and clearing a seat
(V2 Free-Seat) must clear the row without disturbing the user.
*/
package seatpresentation

import (
	"github.com/jkomoros/boardgame"
)

// StorageRecord is the per-seat presentation record. There is at most one
// row per (GameID, PlayerIndex) pair; absent rows mean the seat hasn't been
// claimed via the Table+Hand join flow (the standard solo-mode game-creation
// path doesn't write rows here).
type StorageRecord struct {
	GameID      string
	PlayerIndex boardgame.PlayerIndex
	DisplayName string
	AvatarSlug  string
}
