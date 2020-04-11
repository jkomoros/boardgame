package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
InactivePlayer is a struct designed to be embedded anonymously in your
PlayerStates. It encodes whether a player is Inactive or not. If a player is
Inactive, then base.GameDelegate will return false for that player index for
PlayerMayBeActive. This is useful if you have 'filled' seats but the players
should not be included in the core game logic of your game. For example, if a
player is sitting out a round of a game (perhaps becuase they joined the game
mid-way through a round). See the package doc of this package for more.
*/
type InactivePlayer struct {
	PlayerInactive bool
}

//IsInactive returns whether Inactive is true. Satisfies the PlayerInactiver
//interface in moves/interfaces.
func (i *InactivePlayer) IsInactive() bool {
	return i.PlayerInactive
}

//SetPlayerInactive sets the player to be Inactive. Satisfies the
//PlayerInactiver interface in moves/interfaces.
func (i *InactivePlayer) SetPlayerInactive() {
	i.PlayerInactive = true
}

//SetPlayerActive sets the player to be Active. Satisfies the PlayerInactiver
//interface in moves/interfaces.
func (i *InactivePlayer) SetPlayerActive() {
	i.PlayerInactive = false
}

//PlayerIsInactive is a convenience method that does the cast to
//interfaces.PlayerInactiver, so you don't have to. You can pass any playerState
//to it and it will return true if the player state implements
//interfaces.PlayerInactiver and the IsInactive returns true, false otherwise.
func PlayerIsInactive(playerState boardgame.ImmutableSubState) bool {
	inactiver, ok := playerState.(interfaces.PlayerInactiver)
	if !ok {
		return false
	}
	return inactiver.IsInactive()
}
