package pairs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// wantPairs is how many pairs each hand port splits into.
var wantPairs = map[string]int{"memory": 17, "blackjack": 22, "tictactoe": 9, "pig": 10, "debuganimations": 24, "werewolf": 16}

// noPair are the manifest functions the hand ports have no counterpart for
// (server/static/scripts/pairs-extract.mjs SPECIAL says why).
var noPair = []string{"blackjack.handValue", "tictactoe.(*gameState).tokenValueAtIndex", "tictactoe.rowColToIndex"}

var (
	loadOnce   sync.Once
	loadRoot   string
	loadStored []Pair
	loadErr    error
	manifests  map[string]*gamedef.StaticReport
)

func store(t *testing.T) (string, []Pair, map[string]*gamedef.StaticReport) {
	t.Helper()
	loadOnce.Do(func() {
		if loadRoot, loadErr = ModuleRoot(); loadErr != nil {
			return
		}
		if loadStored, loadErr = Load(loadRoot); loadErr != nil {
			return
		}
		manifests, loadErr = Manifests(loadRoot, Games)
	})
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	return loadRoot, loadStored, manifests
}

func node(t *testing.T) string {
	t.Helper()
	bin, err := Node()
	if err != nil {
		t.Skip("the TypeScript half needs Node: " + err.Error())
	}
	return bin
}

// TestStore: every pair is well formed, the counts are the hand ports',
// and every manifest function is either a pair or listed in noPair.
func TestStore(t *testing.T) {
	_, stored, manifests := store(t)
	counts := map[string]int{}
	have := map[string]bool{}
	for _, p := range stored {
		if err := p.Validate(); err != nil {
			t.Error(err)
		}
		counts[p.Game]++
		have[p.Meta.Func] = true
	}
	total := 0
	for _, game := range Games {
		if counts[game] != wantPairs[game] {
			t.Errorf("%s: %d pairs, want %d", game, counts[game], wantPairs[game])
		}
		total += counts[game]
	}
	t.Logf("%d pairs: %v", total, counts)
	var missing []string
	for _, game := range Games {
		for _, entry := range manifests[game].Functions {
			if !have[entry.Func] {
				missing = append(missing, entry.Func)
			}
		}
	}
	sort.Strings(missing)
	want := append([]string{}, noPair...)
	sort.Strings(want)
	if strings.Join(missing, ",") != strings.Join(want, ",") {
		t.Errorf("manifest functions without a pair: %v, want %v", missing, want)
	}
}

// TestStale: a pair whose Go source has changed. A hand pair is reported
// (the hand port is not regenerated from Go; rebuild the store after
// re-checking the hand port), any other producer's is a failure.
func TestStale(t *testing.T) {
	_, stored, manifests := store(t)
	for _, s := range Stale(stored, manifests) {
		if s.Pair.Meta.Producer == ProducerHand {
			t.Logf("stale hand pair %s/%s: %s", s.Pair.Game, s.Pair.Name, s.Why)
			continue
		}
		t.Errorf("stale %s pair %s/%s: %s", s.Pair.Meta.Producer, s.Pair.Game, s.Pair.Name, s.Why)
	}
}

// TestIndexCurrent: index.ts imports exactly the stored pairs.
func TestIndexCurrent(t *testing.T) {
	root, stored, _ := store(t)
	var ex []Extracted
	for _, p := range stored {
		ex = append(ex, Extracted{Game: p.Game, Func: p.Meta.Func, Name: p.Name})
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(StoreDir), "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != IndexText(ex) {
		t.Error("index.ts is stale; rebuild the store (testdata/gen)")
	}
}

// TestHelperSignatures: a helper pair has the signature its translation
// has (parameter and result types; names aside), since its callers call it
// directly.
func TestHelperSignatures(t *testing.T) {
	root, stored, _ := store(t)
	byGame := map[string][]Pair{}
	for _, p := range stored {
		if p.Meta.Role == "helper" {
			byGame[p.Game] = append(byGame[p.Game], p)
		}
	}
	for game, helpers := range byGame {
		refs := make([]string, len(helpers))
		for i, p := range helpers {
			refs[i] = p.Meta.Func
		}
		translations, err := Translate(root, game, refs, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range helpers {
			tr := translations[p.Meta.Func]
			if !tr.Accepted {
				continue
			}
			want, err := Signature(tr.Code, p.Name)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Signature(string(p.Accepted), p.Name)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("%s: accepted.ts is %s, the translation %s", p.Meta.Func, got, want)
			}
		}
	}
}

// TestStoreTypeChecks: the store type-checks under the engine's sandbox
// options (testdata/pairs/tsconfig.json). Neither the engine's sandbox
// project nor the test project includes it (it is outside server/static).
func TestStoreTypeChecks(t *testing.T) {
	root, _, _ := store(t)
	bin := node(t)
	tsc := filepath.Join(StaticDir(root), "node_modules", "typescript", "bin", "tsc")
	if _, err := os.Stat(tsc); err != nil {
		t.Skip("server/static/node_modules has no TypeScript: " + err.Error())
	}
	cmd := exec.Command(bin, tsc, "--noEmit", "-p", filepath.Join(root, filepath.FromSlash(StoreDir), "tsconfig.json"))
	cmd.Dir = StaticDir(root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("tsc: %v\n%s", err, out)
	}
}

// TestFuzzFieldsMeasured: every pair records its seeds, and every measured
// field is consistent; a null count carries a reason.
func TestFuzzFieldsMeasured(t *testing.T) {
	root, stored, _ := store(t)
	for _, p := range stored {
		seeds, err := Seeds(root, p.Game)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(seeds) != fmt.Sprint(p.Meta.Fuzz.Seeds) {
			t.Errorf("%s: seeds %v, the corpus has %v", p.Meta.Func, p.Meta.Fuzz.Seeds, seeds)
		}
		f := p.Meta.Fuzz
		if (f.BlocksCovered == nil) != (f.BlocksTotal == nil) || (f.MutantsKilled == nil) != (f.MutantsTotal == nil) {
			t.Errorf("%s: a count without its total", p.Meta.Func)
		}
		if f.BlocksTotal != nil && (*f.BlocksCovered > *f.BlocksTotal || len(f.Unreached) != *f.BlocksTotal-*f.BlocksCovered) {
			t.Errorf("%s: %d of %d blocks with %d unreached listed", p.Meta.Func, *f.BlocksCovered, *f.BlocksTotal, len(f.Unreached))
		}
		if f.MutantsTotal != nil && (*f.MutantsKilled > *f.MutantsTotal || len(f.Survivors) != *f.MutantsTotal-*f.MutantsKilled) {
			t.Errorf("%s: %d of %d mutants killed with %d survivors listed", p.Meta.Func, *f.MutantsKilled, *f.MutantsTotal, len(f.Survivors))
		}
		if f.BlocksTotal == nil {
			t.Errorf("%s: block coverage not measured; run testdata/gen -coverage", p.Meta.Func)
		}
		if f.MutantsTotal == nil && f.Why == "" {
			t.Errorf("%s: no mutant counts and no reason", p.Meta.Func)
		}
	}
}

func workDir(t *testing.T) string {
	t.Helper()
	// The translated modules import the engine by absolute path, so they
	// can live anywhere.
	return t.TempDir()
}

// TestRuleAgainstPairs is the proof of a translator rule
// (fabric/TRANSPILE-PROPOSAL.md, "Turning a refusal reason into a rule"),
// run for every refusal any stored pair carried: once the translator
// accepts the Go source of every pair that carried it, each translation
// substituted for its accepted.ts must pass the game's corpora, and the
// proof counts only pairs whose Go blocks the corpus all reached. A refusal
// the translator still makes is reported as not yet a rule.
func TestRuleAgainstPairs(t *testing.T) {
	root, stored, _ := store(t)
	bin := node(t)
	carried := RefusalsCarried(stored)
	keys := make([][2]string, 0, len(carried))
	for key := range carried {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i][0]+keys[i][1] < keys[j][0]+keys[j][1] })
	if len(keys) == 0 {
		t.Fatal("no stored pair carries a refusal; the proof has nothing to check")
	}
	for _, key := range keys {
		proof, err := ProveRule(root, stored, key[0], key[1], Options{Node: bin, WorkDir: workDir(t)})
		if err != nil {
			t.Fatal(err)
		}
		if refused := proof.StillRefused(); len(refused) > 0 {
			for _, r := range refused {
				t.Logf("%s %s: not yet a rule: the translator still refuses %s (%s)", key[0], key[1], r.Pair.Meta.Func, strings.Join(r.Refusals, "; "))
			}
			continue
		}
		for _, r := range proof.Pairs {
			for _, replay := range r.Replays {
				if !replay.Passed {
					t.Errorf("%s %s: %s: the translation fails the corpus: %s", key[0], key[1], r.Pair.Meta.Func, replay.Failure)
				}
			}
			if !r.CoverageComplete {
				t.Logf("%s %s: %s does not count as proof: its Go blocks are not all reached (%v)", key[0], key[1], r.Pair.Meta.Func, r.Pair.Meta.Fuzz.Unreached)
			}
		}
		if !proof.Proven() {
			t.Errorf("%s %s: accepted everywhere but not proven (no pair with complete coverage, or a failure above)", key[0], key[1])
		} else {
			t.Logf("%s %s: proven on %d pairs", key[0], key[1], len(proof.Pairs))
		}
	}
}

// expectedSubstitutionFailures are pairs whose translation is known to fail
// the corpus when substituted, each with its root cause. When one starts
// passing, remove it.
var expectedSubstitutionFailures = map[string]string{}

// TestSubstitutionProof is the general form: for every stored pair whose Go
// source the translator accepts, its translation substituted for
// accepted.ts must pass the game's conformance and fuzz corpora. It is the
// translator checked against every known answer in the store.
func TestSubstitutionProof(t *testing.T) {
	if testing.Short() {
		t.Skip("the general substitution proof replays every game's corpus once per pair")
	}
	root, stored, _ := store(t)
	bin := node(t)
	proofs, err := Substitute(root, stored, Games, Options{Node: bin, WorkDir: workDir(t)})
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, g := range proofs {
		if !g.Baseline.Passed {
			t.Errorf("%s: the store itself fails: %s", g.Game, g.Baseline.Failure)
		}
		pairs, translated, substituted, passing, equal, failing, helpers := g.Counts()
		lines = append(lines, fmt.Sprintf("| %s | %d | %d | %d | %d | %d | %d | %d |", g.Game, pairs, translated, substituted, passing, helpers, equal, failing))
		for _, r := range g.Results {
			if r.TextuallyEqual {
				t.Logf("%s: the translator's text equals accepted.ts's (whitespace aside)", r.Pair.Meta.Func)
			}
			if !r.Translation.Accepted {
				var why []string
				for _, refusal := range r.Translation.Refusals {
					why = append(why, refusal.Reason+" "+refusal.Detail)
				}
				t.Logf("%s: not translated: %s", r.Pair.Meta.Func, strings.Join(why, "; "))
			}
			if r.Replay == nil || r.Replay.Passed {
				if r.Replay != nil {
					if expected, ok := expectedSubstitutionFailures[r.Pair.Meta.Func]; ok {
						t.Errorf("%s passes now; remove it from expectedSubstitutionFailures (%s)", r.Pair.Meta.Func, expected)
					}
				}
				continue
			}
			if expected, ok := expectedSubstitutionFailures[r.Pair.Meta.Func]; ok {
				t.Logf("%s fails as expected (%s): %s", r.Pair.Meta.Func, expected, r.Replay.Failure)
				continue
			}
			t.Errorf("%s: the translation fails the corpus: %s", r.Pair.Meta.Func, r.Replay.Failure)
		}
		t.Logf("%s: baseline %d documents, %d versions, %d rejections", g.Game, g.Baseline.Documents, g.Baseline.Versions, g.Baseline.Rejections)
	}
	t.Log("\n| Game | Pairs | Translated | Substituted | Passing | Helpers (through callers) | Textually equal | Failing |\n|---|---|---|---|---|---|---|---|\n" + strings.Join(lines, "\n"))
}

// TestBrokenRuleFailsProof: a deliberately wrong translation must fail the
// proof. The translator's output for blackjack.handValue is changed in
// memory so an ace is promoted while the hand is at 11 too (`>= 11` becomes
// `> 11`: the "fix" of Go's ten-plus-ace bug); every pair whose translation
// bundles handValue must then fail the corpus.
func TestBrokenRuleFailsProof(t *testing.T) {
	root, stored, _ := store(t)
	bin := node(t)
	rewrote := 0
	broken := func(fn *transpile.Function) {
		if fn.Func != "blackjack.handValue" {
			return
		}
		if !strings.Contains(fn.Code, "currentValue >= 11") {
			t.Fatalf("the translation of blackjack.handValue no longer has `currentValue >= 11`:\n%s", fn.Code)
		}
		fn.Code = strings.Replace(fn.Code, "currentValue >= 11", "currentValue > 11", 1)
		rewrote++
	}
	// The pairs whose translation bundles handValue (its callers, through
	// (*playerState).HandValue).
	var candidates []Pair
	var refs []string
	for _, p := range stored {
		if p.Game == "blackjack" && p.Meta.Role != "helper" {
			candidates = append(candidates, p)
			refs = append(refs, p.Meta.Func)
		}
	}
	translations, err := Translate(root, "blackjack", refs, nil)
	if err != nil {
		t.Fatal(err)
	}
	var subjects []Pair
	for _, p := range candidates {
		for _, callee := range translations[p.Meta.Func].Callees {
			if callee == "blackjack.handValue" {
				subjects = append(subjects, p)
			}
		}
	}
	if len(subjects) == 0 {
		t.Fatal("no stored blackjack pair's translation calls handValue")
	}
	rewrote = 0
	proof, err := SubstituteGame(root, "blackjack", subjects, Options{Node: bin, WorkDir: workDir(t), Rewrite: broken})
	if err != nil {
		t.Fatal(err)
	}
	if rewrote != 1 {
		t.Fatalf("rewrote %d translations of blackjack.handValue", rewrote)
	}
	failed := 0
	for _, r := range proof.Results {
		bundles := false
		for _, callee := range r.Translation.Callees {
			bundles = bundles || callee == "blackjack.handValue"
		}
		if r.Replay == nil {
			continue
		}
		switch {
		case bundles && r.Replay.Passed:
			t.Errorf("%s bundles the broken handValue and still passes the corpus", r.Pair.Meta.Func)
		case bundles:
			failed++
			t.Logf("%s fails as it must: %s", r.Pair.Meta.Func, r.Replay.Failure)
		case !r.Replay.Passed:
			t.Errorf("%s does not call handValue but fails: %s", r.Pair.Meta.Func, r.Replay.Failure)
		}
	}
	if failed == 0 {
		t.Fatal("no substitution bundled the broken handValue")
	}
}
