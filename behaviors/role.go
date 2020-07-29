package behaviors

import (
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
