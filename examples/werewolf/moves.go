package werewolf

import (
	"errors"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

const voteDuration = 45 * time.Second

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
// Declarative-legality survey re-check (Task 7, design spec §6): Task 6
// (design spec §5) widened the v1 seam to include moves.StartPhase
// (verified structurally, moves/seam_source_test.go, to declare no Legal()
// override of its own), so the base-type block Task 12's survey originally
// cited here no longer applies. In practice that doesn't create any work:
// this move has no Legal() override of its own (it relies entirely on
// moves.StartPhase's -> moves.Default's frozen chain) and declares no
// custom gate, so there is nothing to opt into WithLegalPreconditions for — an
// open seam doesn't manufacture content where none exists. Unaffected by
// this task.
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

	populateFellowWolves(players)

	return startVoteTimer(state)
}

func startVoteTimer(state boardgame.State) error {
	if !timedVotingEnabled(state) {
		return nil
	}
	game, _ := concreteStates(state)
	moveName := "Resolve Day Votes on Timeout"
	if game.Phase.Value() == phaseNight {
		moveName = "Resolve Night Votes on Timeout"
	}
	completion := state.Game().MoveByNameForState(moveName, state)
	if completion == nil {
		return errors.New("timed vote resolver was not configured")
	}
	game.VoteTimer.Start(voteDuration, completion)
	return nil
}

func populateFellowWolves(players []*playerState) {
	var werewolves []boardgame.PlayerIndex
	for i, p := range players {
		p.FellowWolves = nil
		if behaviors.PlayerIsInactive(p) || p.Role.Value() != roleWerewolf {
			continue
		}
		werewolves = append(werewolves, boardgame.PlayerIndex(i))
	}
	for _, wolf := range werewolves {
		for _, fellow := range werewolves {
			if fellow != wolf {
				players[wolf].FellowWolves = append(players[wolf].FellowWolves, fellow)
			}
		}
	}
}

func voteForPhase(player *playerState, phase enum.EnumKey) boardgame.PlayerIndex {
	if phase == phaseNight {
		return player.NightVote
	}
	return player.DayVote
}

// moveCastVote is a non-fixup move where a player votes for who to eliminate.
// During the day phase, any alive player may vote. During the night phase,
// only alive werewolves may vote.
//
// Declarative legality: moves.AnyPlayer contributes its own check (target
// defaulting, seat and proposer validity) as the framework check
// moves.PreconditionAnyPlayer, after the phase gate AddForPhase contributes.
// The first gate this move adds -- an eliminated player cannot vote -- is a
// declared precondition (see main.go). The rest stays LegalCustom, in its
// original order: "already voted" reads a different field per phase, and the
// night-phase "only werewolves may act" branch is a disjunction of
// conjunctions ((phase==night AND role==werewolf) OR phase==day), which the
// catalog deliberately cannot express. legal_golden_test.go compares this
// against a snapshot of the Legal() it replaced.
//
// The proposer-relative reads here (the voter's own Role, sanitized from
// other players) remain server-side: LegalCustom's verdict is server-only.
//
//boardgame:codegen
type moveCastVote struct {
	moves.AnyPlayer
	// VoteTarget is the player index this player wants to eliminate.
	VoteTarget boardgame.PlayerIndex `sanitize:"self:visible,same-role:visible"`
}

// LegalCustom runs after the contributed phase and AnyPlayer checks and the
// declared eliminated-voter gate have passed.
func (m *moveCastVote) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	voter := players[m.TargetPlayerIndex]

	phase := game.Phase.Value()

	if voteForPhase(voter, phase) >= 0 {
		return errors.New("you have already voted this phase")
	}

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
	game, players := concreteStates(state)
	voter := players[m.TargetPlayerIndex]
	if game.Phase.Value() == phaseNight {
		voter.NightVote = m.VoteTarget
	} else {
		voter.DayVote = m.VoteTarget
	}
	return nil
}

// moveResolveVotes is a fixup move that triggers when all eligible voters
// have voted. It tallies votes, eliminates the target (if any), resets votes,
// and transitions the phase.
//
// Declarative-legality survey re-check (Task 7, design spec §6): Task 6
// (design spec §5) widened the v1 seam to include moves.FixUp (verified
// structurally to declare no Legal() override of its own), closing the
// base-type block Task 12's survey originally cited here. That does NOT
// unblock a migration, though: this move's one substantive gate — "every
// eligible voter has voted" — is a per-player quantifier whose per-player
// exemption depends on two properties legal.AllActivePlayers' inner leaf
// restriction (buildAllActivePlayersLeaf, legal/catalog_players.go) cannot
// read at all: Vote is boardgame.PlayerIndex-typed, not int (propCompare's
// inner leaf only reads reader.IntProp), and Role is enum-typed (no enum
// leaf kind is accepted inside AllActivePlayers — only playerBool,
// propAtLeast, and propCompare are, none of them enum-aware). Design spec
// §2's typed-equality legal.PropEquals/PropNotEquals would handle Role's
// enum comparison fine as a STANDALONE top-level predicate, but neither is
// in AllActivePlayers' supported inner-leaf set, so it can't be composed
// into the quantifier this check actually needs. This is a genuine,
// reportable catalog gap (AllActivePlayers' inner leaf kinds not growing
// alongside spec §6's new scalar predicates), not a base-type problem
// anymore.
//
// Separately, and moot given the above: this move's OTHER check ("not in a
// voting phase") is already fully redundant with the "inPhase" precondition
// each of this move's two per-phase registrations (main.go's
// moves.AddForPhase(phaseDay, ...) / moves.AddForPhase(phaseNight, ...))
// gets contributed automatically from its own single-phase
// WithLegalPhases call: moves.Default.Legal's legalInPhase check runs
// FIRST (via this move's `m.FixUp.Legal(state, proposer)` super-call) and
// already rejects any phase outside the one this specific registration was
// configured for, so the manual phase != phaseDay && phase != phaseNight
// check below can never actually fire for either registration. Left
// byte-for-byte unchanged (verified-redundant dead code is still not this
// task's file scope to remove).
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
		if voteForPhase(p, phase) < 0 {
			return errors.New("not all eligible players have voted")
		}
	}

	return nil
}

func (m *moveResolveVotes) Apply(state boardgame.State) error {
	return finishVoteResolution(state)
}

// moveResolveVotesOnTimeout deliberately has a distinct move identity but no
// author input. Embedding the ordinary resolver reuses its generated property
// plumbing while Legal keeps the partial-vote path exclusive to the timer's
// AdminPlayerIndex proposal.
type moveResolveVotesOnTimeout struct {
	moveResolveVotes
}

func (m *moveResolveVotesOnTimeout) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	if proposer != boardgame.AdminPlayerIndex {
		return errors.New("only the timer may resolve incomplete votes")
	}
	game, _ := concreteStates(state)
	if !timedVotingEnabled(state) || !game.VoteTimer.Active() {
		return errors.New("timed vote resolution is not active")
	}
	return nil
}

func (m *moveResolveVotesOnTimeout) Apply(state boardgame.State) error {
	return finishVoteResolution(state)
}

func finishVoteResolution(state boardgame.State) error {
	game, _ := concreteStates(state)
	game.VoteTimer.Cancel()
	if err := resolveVotes(state); err != nil {
		return err
	}
	if finished, _ := state.Manager().Delegate().CheckGameFinished(state); finished {
		return nil
	}
	return startVoteTimer(state)
}

func resolveVotes(state boardgame.State) error {
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
		vote := voteForPhase(p, phase)
		if vote >= 0 {
			voteCounts[vote]++
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
		p.DayVote = -1
		p.NightVote = -1
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
