package behaviors

/*
InactivePlayer is a struct designed to be embedded anonymously in your
PlayerStates. It encodes whether a player is Inactive or not. If a player is
Inactive, then base.GameDelegate will return false for that player index for
PlayerMayBeActive. This is useful if you have 'filled' seats but the players
should not be included in the core game logic of your game. For example, if a
player is sitting out a round of a game (perhaps becuase they joined the game
mid-way through a round).
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
