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
	// Vote is the player index this player is voting for. -1 means no vote.
	Vote boardgame.PlayerIndex
}
