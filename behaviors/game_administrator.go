package behaviors

import (
	"github.com/jkomoros/boardgame"
)

/*
GameAdministrator is a struct designed to be embedded in your playerState. It
tracks whether a given player has game-admin authority — the ability to perform
host-like actions such as starting the game, reassigning teams, or kicking
players.

By default, the first player seated (typically the game creator) is
automatically marked as admin by [moves.SeatPlayer]. Admin status can be
changed at runtime by directly setting IsAdmin on the behavior.

Moves that should be restricted to the game admin can embed [moves.AdminPlayer]
as their base type (parallel to [moves.CurrentPlayer] and [moves.AnyPlayer]),
or individual moves can be configured with [moves.WithRequireAdmin].

If your playerState does not embed GameAdministrator, admin checks are skipped
and all players have equal authority (the existing default behavior).
*/
type GameAdministrator struct {
	IsAdmin bool
}

// GetGameAdministrator returns itself, satisfying [HasGameAdministrator].
func (g *GameAdministrator) GetGameAdministrator() *GameAdministrator {
	return g
}

// IsGameAdmin returns whether this player is currently the game administrator.
func (g *GameAdministrator) IsGameAdmin() bool {
	return g.IsAdmin
}

// SetGameAdmin marks this player as the game administrator.
func (g *GameAdministrator) SetGameAdmin() {
	g.IsAdmin = true
}

// ClearGameAdmin removes game administrator status from this player.
func (g *GameAdministrator) ClearGameAdmin() {
	g.IsAdmin = false
}

// HasGameAdministrator is implemented by any SubState that embeds a
// [GameAdministrator]. It allows moves and framework code to discover the
// behavior via type assertion.
type HasGameAdministrator interface {
	GetGameAdministrator() *GameAdministrator
}

// PlayerIsAdmin is a convenience function that returns true if the player
// state implements [HasGameAdministrator] and IsAdmin is true. If the player
// state does not embed GameAdministrator, returns false.
func PlayerIsAdmin(playerState boardgame.ImmutableSubState) bool {
	if admin, ok := playerState.(HasGameAdministrator); ok {
		return admin.GetGameAdministrator().IsAdmin
	}
	return false
}
