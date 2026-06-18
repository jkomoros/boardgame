package werewolf

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves"
)

// moveBeginGame transitions gathering → day AND assigns roles. Role
// assignment must happen here — not in FinishSetUp — because games with
// more slots than MinNumPlayers can legally start with empty seats
// (WaitForEnoughPlayers fires at 4 seated even in a 7-slot game), and
// InactivateEmptySeat marks those seats inactive just before this move.
// Assigning roles across all slots at setup time could put a werewolf on
// a never-filled seat, making the game instantly won or unwinnable
// (GameEndConditionMet only counts active players). Assigning here, only
// among active players, with the wolf count derived from the ACTIVE
// player count, closes that hole.
//
//boardgame:codegen
type moveBeginGame struct {
	moves.StartPhase
}

func (m *moveBeginGame) Apply(state boardgame.State) error {
	if err := m.StartPhase.Apply(state); err != nil {
		return err
	}

	_, players := concreteStates(state)

	var active []*playerState
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		active = append(active, p)
	}

	numWerewolves := 1
	if len(active) >= 6 {
		numWerewolves = 2
	}

	indices := state.Rand().Perm(len(active))

	for i, p := range active {
		isWerewolf := false
		for j := 0; j < numWerewolves; j++ {
			if indices[j] == i {
				isWerewolf = true
				break
			}
		}
		if isWerewolf {
			p.Role.SetValue(roleWerewolf)
		} else {
			p.Role.SetValue(roleVillager)
		}
	}

	return nil
}

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

	if target == m.TargetPlayerIndex {
		return errors.New("you cannot vote for yourself")
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

	// Find the player with most votes. Iterate by player index (not map
	// order) so tie-breaks are deterministic — golden-style replays would
	// otherwise produce different victims per run.
	var maxVotes int
	var eliminated boardgame.PlayerIndex = -1
	tied := false

	for i := 0; i < len(players); i++ {
		target := boardgame.PlayerIndex(i)
		count := voteCounts[target]
		if count == 0 {
			continue
		}
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
