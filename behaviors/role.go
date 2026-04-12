package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

/*
PlayerRole is a struct that is designed to be embedded in your playerState. It
assumes there is an enum called `role`. This is typically used for when
different players have different roles, for example roleGuesser and
roleClueGiver. If your role enum is combined with the group enum, then
base.GameDelegate.GroupMembership will pick this up automatically.
*/
type PlayerRole struct {
	Role enum.Val `enum:"role"`
}

// GetPlayerRole returns itself, satisfying [HasPlayerRole].
func (p *PlayerRole) GetPlayerRole() *PlayerRole {
	return p
}

// HasPlayerRole is implemented by any SubState that embeds a [PlayerRole]. It
// allows moves and framework code to find the behavior via type assertion.
type HasPlayerRole interface {
	GetPlayerRole() *PlayerRole
}

// PlayerRoleValue is a convenience function that returns the role enum value
// for the given player state, if it implements [HasPlayerRole]. Returns nil if
// the player state does not embed PlayerRole.
func PlayerRoleValue(playerState boardgame.ImmutableSubState) enum.ImmutableVal {
	if role, ok := playerState.(HasPlayerRole); ok {
		return role.GetPlayerRole().Role
	}
	return nil
}
