package werewolf

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves"
)

// moveCastVote is a non-fixup move where a player votes for who to eliminate.
// During the day phase, any alive player may vote. During the night phase,
// only alive werewolves may vote.
//
//boardgame:codegen
type moveCastVote struct {
	moves.AnyPlayer
	// VoteTarget is the player index this player wants to eliminate.
	VoteTarget boardgame.PlayerIndex
}

func (m *moveCastVote) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	voter := players[m.TargetPlayerIndex]

	if voter.Eliminated {
		return errors.New("eliminated players cannot vote")
	}

	if voter.Vote >= 0 {
		return errors.New("you have already voted this phase")
	}

	phase := game.Phase.Value()

	if phase == phaseNight {
		if voter.Role.Value() != roleWerewolf {
			return errors.New("only werewolves may act during the night")
		}
	} else if phase != phaseDay {
		return errors.New("voting is not allowed during this phase")
	}

	// Validate the target
	target := m.VoteTarget
	if target < 0 || int(target) >= len(players) {
		return errors.New("invalid vote target")
	}

	targetPlayer := players[target]
	if targetPlayer.Eliminated {
		return errors.New("cannot vote for an eliminated player")
	}

	if behaviors.PlayerIsInactive(targetPlayer) {
		return errors.New("cannot vote for an inactive player")
	}

	// During day phase, you cannot vote for yourself
	if phase == phaseDay && target == m.TargetPlayerIndex {
		return errors.New("you cannot vote for yourself during the day")
	}

	return nil
}

func (m *moveCastVote) Apply(state boardgame.State) error {
	_, players := concreteStates(state)
	voter := players[m.TargetPlayerIndex]
	voter.Vote = m.VoteTarget
	return nil
}

// moveResolveVotes is a fixup move that triggers when all eligible voters
// have voted. It tallies votes, eliminates the target (if any), resets votes,
// and transitions the phase.
//
//boardgame:codegen
type moveResolveVotes struct {
	moves.FixUp
}

func (m *moveResolveVotes) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	phase := game.Phase.Value()
	if phase != phaseDay && phase != phaseNight {
		return errors.New("not in a voting phase")
	}

	// Check that all eligible voters have voted
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Eliminated {
			continue
		}
		if phase == phaseNight && p.Role.Value() != roleWerewolf {
			// Villagers don't vote at night
			continue
		}
		if p.Vote < 0 {
			return errors.New("not all eligible players have voted")
		}
	}

	return nil
}

func (m *moveResolveVotes) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	phase := game.Phase.Value()

	// Tally votes
	voteCounts := make(map[boardgame.PlayerIndex]int)
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Eliminated {
			continue
		}
		if phase == phaseNight && p.Role.Value() != roleWerewolf {
			continue
		}
		if p.Vote >= 0 {
			voteCounts[p.Vote]++
		}
	}

	// Find the player with most votes
	var maxVotes int
	var eliminated boardgame.PlayerIndex = -1
	tied := false

	for target, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			eliminated = target
			tied = false
		} else if count == maxVotes {
			tied = true
		}
	}

	// During day phase, ties mean no elimination
	if phase == phaseDay && tied {
		eliminated = -1
	}

	// Eliminate the target (if any)
	if eliminated >= 0 && int(eliminated) < len(players) {
		players[eliminated].SetEliminated()
	}

	// Reset all votes
	for _, p := range players {
		p.Vote = -1
	}

	// Transition phase
	if phase == phaseDay {
		game.SetCurrentPhase(phaseNight)
	} else {
		game.RoundNumber++
		game.SetCurrentPhase(phaseDay)
	}

	return nil
}
