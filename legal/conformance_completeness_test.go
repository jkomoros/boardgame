package legal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

/*
TestConformanceCorpusCompleteness is Task 14's corpus-completeness
meta-test: every predicate name registered by legal.DefaultConstructors()
(the universal catalog) or moves.FrameworkConstructors() (the one framework
predicate -- "inProgression" -- that must live in package moves; see
moves/catalog_framework.go's doc comment for why) must have a conformance
file under testdata/conformance/ with >= 3 cases (TestConformanceCorpus,
conformance_test.go, already enforces the >=3-cases part for every file that
exists; this test enforces that a file exists at all, for every registered
name).

Exemptions are named explicitly below, never silently skipped: a predicate
lands in conformanceExemptions only when a real conformance fixture is
infeasible from this package with today's fixture machinery, with a comment
explaining why. Removing an entry from this list (by adding the
corresponding corpus file) is always preferred to leaving it exempted.
*/

// conformanceExemptions names registered predicates that legitimately have
// no testdata/conformance/*.json file. Each entry must carry a reason.
var conformanceExemptions = map[string]string{
	"inProgression": "moves/catalog_framework.go's inProgressionConstructor.Evaluate looks up a LIVE move instance via " +
		"ctx.State.Game().MoveByName(moveName) and calls that instance's unexported legalMoveInProgression method " +
		"(promoted from moves.Default) -- unlike every other catalog predicate, its correctness depends on a specific " +
		"move type's own WithLegalMoveProgression configuration and move-tape history, not just fixture state. " +
		"legal_test (this package) has no fixture-construction path that drives real move proposals through a " +
		"progression the way moves-package tests already do (see moves/inprogression_test.go and " +
		"moves/legal_tape_memo_test.go, which exercise this predicate's real semantics end to end from the package " +
		"that owns MoveProgressionGroup). Adding it here would mean either reimplementing that harness a second time " +
		"or reaching into moves-package internals this package cannot import (package legal cannot import package " +
		"moves -- see the design spec's layering diagram). Exempted rather than faked with a trivial/misleading fixture.",
}

// registeredPredicateNames returns every predicate name legal.DefaultConstructors()
// and moves.FrameworkConstructors() register, deduplicated.
func registeredPredicateNames() []string {
	seen := make(map[string]bool)
	var names []string
	add := func(cs []*legal.PredicateConstructor) {
		for _, c := range cs {
			if seen[c.Name] {
				continue
			}
			seen[c.Name] = true
			names = append(names, c.Name)
		}
	}
	add(legal.DefaultConstructors())
	add(moves.FrameworkConstructors())
	return names
}

// corpusPredicateNames returns the "predicate" field of every JSON file
// under testdata/conformance/.
func corpusPredicateNames(t *testing.T) map[string]bool {
	t.Helper()
	paths, err := filepath.Glob("testdata/conformance/*.json")
	if err != nil {
		t.Fatalf("legal: globbing conformance corpus: %v", err)
	}
	out := make(map[string]bool, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("legal: reading %s: %v", path, err)
		}
		var cf struct {
			Predicate string `json:"predicate"`
		}
		if err := json.Unmarshal(data, &cf); err != nil {
			t.Fatalf("legal: parsing %s: %v", path, err)
		}
		out[cf.Predicate] = true
	}
	return out
}

func TestConformanceCorpusCompleteness(t *testing.T) {
	corpus := corpusPredicateNames(t)

	for _, name := range registeredPredicateNames() {
		if corpus[name] {
			if _, exempt := conformanceExemptions[name]; exempt {
				t.Errorf("legal: %q has BOTH a conformance file and an exemption entry -- remove the stale exemption", name)
			}
			continue
		}
		if reason, exempt := conformanceExemptions[name]; exempt {
			if reason == "" {
				t.Errorf("legal: exemption for %q has no reason recorded", name)
			}
			continue
		}
		t.Errorf("legal: registered predicate %q has no testdata/conformance/*.json file and is not in conformanceExemptions", name)
	}

	// Guard the exemption list itself against drift: every exempted name
	// must actually be a currently-registered predicate (an exemption for a
	// predicate that no longer exists is dead weight, not honesty).
	registered := make(map[string]bool)
	for _, name := range registeredPredicateNames() {
		registered[name] = true
	}
	for name := range conformanceExemptions {
		if !registered[name] {
			t.Errorf("legal: conformanceExemptions has stale entry %q, which is not a currently registered predicate name", name)
		}
	}
}
