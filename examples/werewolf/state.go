package werewolf

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
)

//boardgame:codegen
const (
	phaseGathering = iota
	phaseDay
	phaseNight
)

//boardgame:codegen
const (
	roleVillager = iota
	roleWerewolf
)

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.PhaseBehavior
	RoundNumber int
}

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.Seat
	behaviors.InactivePlayer
	behaviors.PlayerElimination
	behaviors.PlayerRole
	// DayVote is public: daytime voting is deliberately visible to everyone.
	// -1 means no vote.
	DayVote boardgame.PlayerIndex
	// NightVote must never identify a werewolf or their target to another
	// player or an observer. The owner can still see it on their Hand view.
	NightVote boardgame.PlayerIndex `sanitize:"other:hidden"`
	// FellowWolves is populated when roles are assigned. Keeping this on each
	// player makes the wolf-team mechanic available to the owner's Hand view
	// without exposing any other player's Role.
	FellowWolves []boardgame.PlayerIndex `sanitize:"other:hidden"`
}
