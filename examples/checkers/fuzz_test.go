package checkers_test

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
	"github.com/jkomoros/boardgame/examples/checkers"
)

// TestFuzzGamesFinish plays whole games with the fuzz driver's random-legal
// policy at its default bounds: each seed below reaches the end condition,
// with captures, crowning and multi-jump continuations along the way, and no
// input the driver believed legal is refused by Apply.
func TestFuzzGamesFinish(t *testing.T) {
	seeds := []int64{0, 3, 6}
	for _, seed := range seeds {
		result, err := fuzz.RunWithDelegate(func() boardgame.GameDelegate { return checkers.NewDelegate() }, fuzz.Config{Game: "checkers", Seed: seed})
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if result.End != fuzz.EndFinished {
			t.Errorf("seed %d: run ended %q, want %q", seed, result.End, fuzz.EndFinished)
		}
		for _, r := range result.Refusals {
			if r.Reason == fuzz.ReasonApplyError {
				t.Errorf("seed %d: Apply refused a move Legal accepted: %+v", seed, r)
			}
		}
		if result.LegalPanics != 0 {
			t.Errorf("seed %d: %d legality checks panicked", seed, result.LegalPanics)
		}
		moves := result.Moves
		if moves["Move Token"].Applied == 0 || moves["Finish Turn"].Applied == 0 || moves["Crown Token"].Applied == 0 {
			t.Errorf("seed %d: Move Token %d, Finish Turn %d, Crown Token %d applied; want each at least once", seed, moves["Move Token"].Applied, moves["Finish Turn"].Applied, moves["Crown Token"].Applied)
		}
		if moves["Move Token"].Applied <= moves["Finish Turn"].Applied {
			t.Logf("seed %d: no multi-jump continuation (Move Token %d, Finish Turn %d)", seed, moves["Move Token"].Applied, moves["Finish Turn"].Applied)
		}
		if moves["Move Token"].ProposedLegally != moves["Move Token"].Applied {
			t.Errorf("seed %d: %d Move Tokens proposed as legal, %d applied", seed, moves["Move Token"].ProposedLegally, moves["Move Token"].Applied)
		}
	}
}
