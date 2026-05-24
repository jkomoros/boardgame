package behaviors

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
PlayerTeam is a struct designed to be embedded anonymously in your PlayerStates.
It tracks which team a player belongs to via an enum field. It is a [Connectable]
behavior that is automatically connected by the framework.

The game must define a "team" enum. The Team field uses the standard enum struct
tag convention:

	type playerState struct {
	    base.SubState
	    behaviors.PlayerTeam
	}

[PlayerTeam] provides helpers for finding teammates and opponents that account
for inactive players. If a game has asymmetric roles within teams (e.g.
Codenames spymaster vs guesser), use both PlayerTeam and [PlayerRole]
together -- they compose cleanly.
*/
type PlayerTeam struct {
	container boardgame.SubState
	// Team is hidden from observers (e.g. the Table+Hand projector at
	// ObserverPlayerIndex) by default — safe default for games where team
	// affiliation is part of the hidden information. Games where teams
	// are publicly known (the common case for Codenames-style games)
	// override at the embedding site with `sanitize:"all:visible"`.
	// See PlayerRole for the same pattern with more documentation.
	Team enum.Val `enum:"team" sanitize:"other:hidden"`
}

// ConnectBehavior stores a reference to the containing SubState.
func (p *PlayerTeam) ConnectBehavior(containingSubState boardgame.SubState) {
	p.container = containingSubState
}

// ValidConfiguration returns an error if the behavior hasn't been properly
// connected.
func (p *PlayerTeam) ValidConfiguration(example boardgame.State) error {
	if p.container == nil {
		return errors.New("PlayerTeam: ConnectBehavior hasn't been called. See the behaviors package doc for more on initializing Connectable behaviors")
	}
	return nil
}

// IsOnTeam returns true if this player's Team value matches the given key.
// Satisfies [interfaces.TeamMember].
func (p *PlayerTeam) IsOnTeam(teamKey enum.EnumKey) bool {
	return p.Team.Value() == teamKey
}

// TeamMembers returns the player indices of all players on the same team as
// this player (including this player). Inactive players are excluded.
func (p *PlayerTeam) TeamMembers() []boardgame.PlayerIndex {
	if p.container == nil {
		return nil
	}
	myTeam := p.Team.Value()
	var result []boardgame.PlayerIndex
	for i, ps := range p.container.State().ImmutablePlayerStates() {
		if PlayerIsInactive(ps) {
			continue
		}
		if member, ok := ps.(interfaces.TeamMember); ok {
			if member.IsOnTeam(myTeam) {
				result = append(result, boardgame.PlayerIndex(i))
			}
		}
	}
	return result
}

// Opponents returns the player indices of all players NOT on the same team as
// this player. Inactive players are excluded.
func (p *PlayerTeam) Opponents() []boardgame.PlayerIndex {
	if p.container == nil {
		return nil
	}
	myTeam := p.Team.Value()
	var result []boardgame.PlayerIndex
	for i, ps := range p.container.State().ImmutablePlayerStates() {
		if PlayerIsInactive(ps) {
			continue
		}
		if member, ok := ps.(interfaces.TeamMember); ok {
			if !member.IsOnTeam(myTeam) {
				result = append(result, boardgame.PlayerIndex(i))
			}
		}
	}
	return result
}

// GetPlayerTeam returns itself, satisfying [HasPlayerTeam].
func (p *PlayerTeam) GetPlayerTeam() *PlayerTeam {
	return p
}

// HasPlayerTeam is implemented by any SubState that embeds a PlayerTeam. It
// allows moves and framework code to find the behavior via type assertion.
type HasPlayerTeam interface {
	GetPlayerTeam() *PlayerTeam
}

// PlayerTeamValue is a convenience function that returns the team enum value
// for the given player state, if it implements [HasPlayerTeam]. Returns nil if
// the player state does not embed PlayerTeam.
func PlayerTeamValue(playerState boardgame.ImmutableSubState) enum.ImmutableVal {
	if team, ok := playerState.(HasPlayerTeam); ok {
		return team.GetPlayerTeam().Team
	}
	return nil
}
