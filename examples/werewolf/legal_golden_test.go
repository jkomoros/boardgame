package werewolf

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
)

/*
Golden-equivalence harness for moveCastVote's declarative migration, following
examples/memory/legal_golden_test.go's precedent: legacyCastVoteLegal is a
verbatim snapshot of the Legal() the migration deleted, and for every
(version, move, proposer, TargetPlayerIndex, VoteTarget) the test reaches, it
must agree with the migrated move's actual Legal() on the first failure's
text (or both pass).

The snapshot super-calls m.AnyPlayer.Legal exactly as the original did. That
call would evaluate the migrated move's plan, so the snapshot runs under
LegalWithPlansSuspended, which gives it AnyPlayer's frozen imperative chain:
the pre-migration behavior end to end.
*/

func legacyCastVoteLegal(m *moveCastVote, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	voter := players[m.TargetPlayerIndex]

	if voter.Eliminated {
		return errors.New("eliminated players cannot vote")
	}

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

func describeLegal(fn func() error) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("panic: %v", r)
		}
	}()
	if err := fn(); err != nil {
		return "error: " + err.Error()
	}
	return "legal"
}

func compareCastVoteAt(t *testing.T, game *boardgame.Game, state boardgame.ImmutableState) int {
	t.Helper()
	manager := game.Manager()
	n := len(state.ImmutablePlayerStates())
	indexes := []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex}
	for i := 0; i <= n; i++ {
		indexes = append(indexes, boardgame.PlayerIndex(i))
	}
	proposers := append([]boardgame.PlayerIndex{boardgame.AdminPlayerIndex}, indexes...)

	compared := 0
	for _, name := range []string{"Cast Vote", "Cast Night Vote"} {
		for _, proposer := range proposers {
			for _, targetPlayer := range indexes {
				for _, voteTarget := range indexes {
					legacy := game.MoveByNameForState(name, state).(*moveCastVote)
					migrated := game.MoveByNameForState(name, state).(*moveCastVote)
					for _, m := range []*moveCastVote{legacy, migrated} {
						m.TargetPlayerIndex = targetPlayer
						m.VoteTarget = voteTarget
					}

					var want string
					manager.LegalWithPlansSuspended(func() {
						want = describeLegal(func() error { return legacyCastVoteLegal(legacy, state, proposer) })
					})
					got := describeLegal(func() error { return migrated.Legal(state, proposer) })
					compared++

					// A legacy panic (an index the frozen chain never
					// guarded) must still be refused by the plan, which
					// fails closed instead of panicking.
					if len(want) >= 6 && want[:6] == "panic:" {
						if got == "legal" {
							t.Errorf("version %d %q proposer %v target %v vote %v: legacy panicked (%v) but migrated is legal", state.Version(), name, proposer, targetPlayer, voteTarget, want)
						}
						continue
					}
					if want != got {
						t.Errorf("version %d %q proposer %v target %v vote %v: legacy %v, migrated %v", state.Version(), name, proposer, targetPlayer, voteTarget, want, got)
					}
				}
			}
		}
	}
	return compared
}

// TestGoldenLegalMoveCastVote plays a seated game through a day vote into the
// night (so states include already-voted, eliminated and night-phase voters),
// comparing at every version reached.
func TestGoldenLegalMoveCastVote(t *testing.T) {
	_, game := newSeatedGame(t, 4)

	for step := 0; step < 16 && !game.Finished(); step++ {
		proposed := false
		state := game.CurrentState()
		for _, name := range []string{"Cast Vote", "Cast Night Vote"} {
			for voter := 0; voter < len(state.ImmutablePlayerStates()) && !proposed; voter++ {
				for target := 0; target < len(state.ImmutablePlayerStates()) && !proposed; target++ {
					move := game.MoveByName(name).(*moveCastVote)
					move.TargetPlayerIndex = boardgame.PlayerIndex(voter)
					move.VoteTarget = boardgame.PlayerIndex(target)
					if move.Legal(state, boardgame.PlayerIndex(voter)) != nil {
						continue
					}
					if err := <-game.ProposeMove(move, boardgame.PlayerIndex(voter)); err == nil {
						proposed = true
					}
				}
			}
		}
		if !proposed {
			break
		}
	}

	compared := 0
	phases := make(map[enum.EnumKey]bool)
	for version := 0; version <= game.Version(); version++ {
		state := game.State(version)
		if state == nil {
			continue
		}
		g, _ := concreteStates(state)
		phases[g.Phase.Value()] = true
		compared += compareCastVoteAt(t, game, state)
	}
	if !phases[phaseDay] || !phases[phaseNight] {
		t.Fatalf("the walk should reach both day and night, reached %v", phases)
	}
	t.Logf("compared %d (version, move, proposer, target, vote) cases", compared)
}
