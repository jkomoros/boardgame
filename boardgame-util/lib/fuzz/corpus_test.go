package fuzz

import (
	"strings"
	"testing"
)

const repoRoot = "../../.."

// TestCorpusIsCurrent regenerates all 48 checked-in fuzz documents and the
// index on the real engine (about half a minute) and compares them with
// conformance/fuzz. If the engine, a game or the driver changed on purpose,
// regenerate with:
//
//	./scripts/go-local run ./boardgame-util fuzz corpus
func TestCorpusIsCurrent(t *testing.T) {
	configs, err := CorpusConfigs()
	if err != nil {
		t.Fatal(err)
	}
	problems, index, err := CheckCorpus(repoRoot, configs, true)
	if err != nil {
		t.Fatalf("generate fuzz corpus: %v", err)
	}
	if len(problems) > 0 {
		t.Fatalf("the fuzz corpus is stale; regenerate with\n\t./scripts/go-local run ./boardgame-util fuzz corpus\n%s", strings.Join(problems, "\n"))
	}
	assertCorpusCoverage(t, index)
}

// expectedNeverApplied are the move types the checked-in corpus does not
// apply, with the reason. Every other move type of every game must be
// applied at least once across its seeds.
var expectedNeverApplied = map[string]string{
	"blackjack/Shuffle Discard Into Draw": "fires only once the 52-card draw stack is exhausted, which takes several full rounds (conformance/blackjack/scenario-reshuffle covers it)",
	"blackjack/Force Finish Turn":         "legal only for the admin, and the driver never proposes as the admin",
	"debuganimations/Seat Player":         "the corpus takes only unseated debuganimations seeds (seated, its players are never activated)",
}

func assertCorpusCoverage(t *testing.T, index *Index) {
	t.Helper()
	if len(index.Games) != len(CorpusGames) {
		t.Fatalf("index has %d games, want %d", len(index.Games), len(CorpusGames))
	}
	for _, game := range index.Games {
		if game.Seeds != CorpusSeeds {
			t.Errorf("%s: %d seeds, want %d", game.Game, game.Seeds, CorpusSeeds)
		}
		refusals := 0
		for _, count := range game.RejectionsByKind {
			refusals += count
		}
		if refusals == 0 {
			t.Errorf("%s: no refusal recorded", game.Game)
		}
		for _, name := range game.NeverApplied {
			if _, ok := expectedNeverApplied[game.Game+"/"+name]; !ok {
				t.Errorf("%s: move %q is never applied and has no recorded reason", game.Game, name)
			}
		}
	}
}
