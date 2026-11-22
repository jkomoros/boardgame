package lint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
)

var (
	examplesOnce   sync.Once
	examplesReport DeclarativeReport
)

// examples runs the lint once over the seven example games.
func examples(t *testing.T) DeclarativeReport {
	t.Helper()
	examplesOnce.Do(func() {
		var inputs []string
		for _, game := range ExampleGames {
			inputs = append(inputs, filepath.Join("..", "..", "..", "examples", game))
		}
		examplesReport = Declarative(inputs, DeclarativeOptions{MaxRefused: -1})
	})
	if !examplesReport.OK || len(examplesReport.Packages) != len(ExampleGames) {
		t.Fatalf("lint of the examples failed: %+v", examplesReport.Diagnostics)
	}
	return examplesReport
}

func examplePackage(t *testing.T, name string) DeclarativePackage {
	t.Helper()
	for _, pkg := range examples(t).Packages {
		if pkg.Name == name {
			return pkg
		}
	}
	t.Fatalf("no package %s", name)
	return DeclarativePackage{}
}

// proposalBaseline is fabric/TRANSPILE-PROPOSAL.md's baseline table, by
// game, as the proposal states it (counted by hand, before this tool).
var proposalBaseline = map[string]Baseline{
	"memory":          {MoveConfigs: 7, DataOnly: 3, WithGameCode: 4, DeclaredPreconditionSpecs: 9, Functions: 16, InSubset: 15, Outside: 1, NeedEngineFeature: 0, NotPorted: 11},
	"blackjack":       {MoveConfigs: 23, DataOnly: 16, WithGameCode: 7, DeclaredPreconditionSpecs: 6, Functions: 22, InSubset: 22, Outside: 0, NeedEngineFeature: 0, NotPorted: 1},
	"pig":             {MoveConfigs: 6, DataOnly: 3, WithGameCode: 3, DeclaredPreconditionSpecs: 4, Functions: 10, InSubset: 10, Outside: 0, NeedEngineFeature: 0, NotPorted: 1},
	"tictactoe":       {MoveConfigs: 4, DataOnly: 3, WithGameCode: 1, DeclaredPreconditionSpecs: 0, Functions: 11, InSubset: 11, Outside: 0, NeedEngineFeature: 0, NotPorted: 6},
	"werewolf":        {MoveConfigs: 11, DataOnly: 4, WithGameCode: 7, DeclaredPreconditionSpecs: 0, Functions: 16, InSubset: 16, Outside: 0, NeedEngineFeature: 0, NotPorted: 1},
	"debuganimations": {MoveConfigs: 13, DataOnly: 1, WithGameCode: 12, DeclaredPreconditionSpecs: 7, Functions: 24, InSubset: 24, Outside: 0, NeedEngineFeature: 0, NotPorted: 1},
	// The proposal gives "about 11 (estimate)" for checkers' engine features.
	"checkers": {MoveConfigs: 7, DataOnly: 4, WithGameCode: 3, DeclaredPreconditionSpecs: 5, Functions: 20, InSubset: 19, Outside: 1, NeedEngineFeature: 11, NotPorted: 0},
}

// measuredDifferences are where the measured row differs from the
// proposal's hand count, with the cause (read from the Go; see
// fabric/transpile/status/static-lint.md). Each entry is the measured value.
var measuredDifferences = map[string]map[string]int{
	// The computed property's closure (ConfigureComputedProperties.func1) is a
	// translation unit of its own, beside the method it calls
	// ((*gameState).CurrentPlayerHasCardsToReveal); the hand count counted
	// the computed property once.
	"memory": {"functions": 17, "inSubset": 16},
	// The same: ConfigureComputedProperties.func1 and (*playerState).HandValue.
	"blackjack": {"functions": 23, "inSubset": 23},
	// Not ported is Diagram plus the four Agent methods: 5.
	// (*gameState).tokenValue is reached by nothing (unreached), not agent code.
	"tictactoe": {"notPorted": 5},
	// The proposal's 11 was an estimate (six move methods, five token
	// methods). Measured: (*moveCrownToken).Apply, (*moveMoveToken).Apply and
	// LegalCustom, (*movePlaceToken).DefaultsForState, and (*token).Legal,
	// AllNextSpaces, LegalCaptureSpaces, FreeNextSpaces. (*movePlaceToken)'s
	// Apply and LegalCustom and (*token).Dynamic use nothing the engine lacks.
	"checkers": {"needEngineFeature": 8},
}

func baselineField(b Baseline, name string) int {
	switch name {
	case "functions":
		return b.Functions
	case "inSubset":
		return b.InSubset
	case "outside":
		return b.Outside
	case "needEngineFeature":
		return b.NeedEngineFeature
	case "notPorted":
		return b.NotPorted
	}
	panic(name)
}

// TestDeclarativeReproducesBaseline: the first four columns equal the
// proposal's table exactly; every other column equals it unless a
// documented difference names the measured value.
func TestDeclarativeReproducesBaseline(t *testing.T) {
	var total Baseline
	for _, game := range ExampleGames {
		got := examplePackage(t, game).Baseline
		want := proposalBaseline[game]
		if got.MoveConfigs != want.MoveConfigs || got.DataOnly != want.DataOnly ||
			got.WithGameCode != want.WithGameCode || got.DeclaredPreconditionSpecs != want.DeclaredPreconditionSpecs {
			t.Errorf("%s: first four columns %d/%d/%d/%d, proposal %d/%d/%d/%d", game,
				got.MoveConfigs, got.DataOnly, got.WithGameCode, got.DeclaredPreconditionSpecs,
				want.MoveConfigs, want.DataOnly, want.WithGameCode, want.DeclaredPreconditionSpecs)
		}
		for _, field := range []string{"functions", "inSubset", "outside", "needEngineFeature", "notPorted"} {
			expected := baselineField(want, field)
			if measured, ok := measuredDifferences[game][field]; ok {
				if measured == expected {
					t.Errorf("%s %s: documented difference %d equals the proposal; remove it", game, field, measured)
				}
				expected = measured
			}
			if got := baselineField(got, field); got != expected {
				t.Errorf("%s %s = %d, want %d", game, field, got, expected)
			}
		}
		total.add(got)
	}
	if examples(t).Totals != total {
		t.Errorf("totals %+v, want the sum %+v", examples(t).Totals, total)
	}
	if total.MoveConfigs != 71 || total.DataOnly != 34 || total.WithGameCode != 37 || total.DeclaredPreconditionSpecs != 31 {
		t.Errorf("total first four columns %+v, proposal 71/34/37/31", total)
	}
}

// TestDeclarativeOutsideFunctions: exactly the proposal's two functions are
// outside the subset, for its reasons.
func TestDeclarativeOutsideFunctions(t *testing.T) {
	type outside struct{ game, fn, reason, detail string }
	var got []outside
	for _, pkg := range examples(t).Packages {
		for _, fn := range pkg.Functions {
			if fn.Bucket != gamedef.BucketOutside {
				continue
			}
			for _, r := range fn.Refusals {
				got = append(got, outside{pkg.Name, fn.Func, r.Reason, r.Detail})
			}
		}
	}
	want := []outside{
		{"checkers", "checkers.(*gameDelegate).ConfigurePredicateConstructors", refusal.PredicateConstructor, "checkers.spaceIsBlack"},
		{"memory", "memory.(*moveStartHideCardsTimer).Apply", refusal.PackageVarRead, "memory.hideCardMoveName"},
	}
	if len(got) != len(want) {
		t.Fatalf("outside functions %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("outside %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestDeclarativeGoldens compares each game's report with its checked-in
// golden. Regenerate with
// ./scripts/go-local run ./boardgame-util/lib/lint/testdata/gen/declarative
func TestDeclarativeGoldens(t *testing.T) {
	for _, pkg := range examples(t).Packages {
		got, err := MarshalDeclarativePackage(pkg)
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join("testdata", "declarative", pkg.Name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s: report differs from testdata/declarative/%s.json; regenerate with ./scripts/go-local run ./boardgame-util/lib/lint/testdata/gen/declarative and review the diff", pkg.Name, pkg.Name)
		}
	}
}

func TestDeclarativeReport(t *testing.T) {
	report := examples(t)
	werewolf := examplePackage(t, "werewolf")
	if len(werewolf.ImperativeLegal) != 6 || werewolf.Baseline.DeclaredPreconditionSpecs != 0 {
		t.Errorf("werewolf: %d imperative Legal configs and %d specs, want six and none", len(werewolf.ImperativeLegal), werewolf.Baseline.DeclaredPreconditionSpecs)
	}
	memory := examplePackage(t, "memory")
	kinds := map[string]int{}
	for _, warning := range memory.Warnings {
		kinds[warning.Kind]++
	}
	if kinds["global-rand"] != 2 {
		t.Errorf("memory warnings %+v, want the agent's two global math/rand calls", memory.Warnings)
	}
	codes := map[string]int{}
	for _, diagnostic := range report.Diagnostics {
		codes[diagnostic.Code]++
	}
	if codes[CodeNeedsTranslation]+codes[CodeRefused] != report.Totals.Functions {
		t.Errorf("BGLINT0101 %d + BGLINT0102 %d != %d functions", codes[CodeNeedsTranslation], codes[CodeRefused], report.Totals.Functions)
	}
	if codes[CodeRefused] != report.Refused || codes[CodeNondeterministic] != 2 || codes[CodeStaleGenerated] != 0 {
		t.Errorf("codes %v, refused %d", codes, report.Refused)
	}
	// Functions are listed callee first.
	for _, pkg := range report.Packages {
		seen := map[string]bool{}
		for _, fn := range pkg.Functions {
			for _, callee := range fn.Calls {
				if !seen[callee] {
					t.Errorf("%s: %s is listed before its callee %s", pkg.Name, fn.Func, callee)
				}
			}
			seen[fn.Func] = true
		}
	}
	var text bytes.Buffer
	WriteDeclarativeText(&text, report)
	for _, want := range []string{"memory (github.com/jkomoros/boardgame/examples/memory)", "package-var-read memory.hideCardMoveName", "imperative Legal overrides", "totals: move configs 71"} {
		if !strings.Contains(text.String(), want) {
			t.Errorf("text report lacks %q", want)
		}
	}
	if _, err := json.Marshal(report); err != nil {
		t.Fatal(err)
	}
}

func TestDeclarativeMaxRefused(t *testing.T) {
	checkers := filepath.Join("..", "..", "..", "examples", "checkers")
	refused := Declarative([]string{checkers}, DeclarativeOptions{MaxRefused: -1})
	if !refused.OK || refused.MaxRefused != nil || refused.Refused != 9 {
		t.Fatalf("no limit: ok %v, limit %v, refused %d (want 9: one outside, eight needing engine features)", refused.OK, refused.MaxRefused, refused.Refused)
	}
	if report := Declarative([]string{checkers}, DeclarativeOptions{MaxRefused: 8}); report.OK {
		t.Error("9 refused passed --max-refused 8")
	}
	if report := Declarative([]string{checkers}, DeclarativeOptions{MaxRefused: 9}); !report.OK || report.MaxRefused == nil || *report.MaxRefused != 9 {
		t.Error("9 refused failed --max-refused 9")
	}
}

func TestDeclarativeBadPackage(t *testing.T) {
	report := Declarative([]string{filepath.Join("testdata", "no-such-package")}, DeclarativeOptions{MaxRefused: -1})
	if report.OK || len(report.Diagnostics) == 0 || report.Diagnostics[0].Code != CodePackage {
		t.Errorf("missing package: %+v", report)
	}
}

func TestStaleGenerated(t *testing.T) {
	hash := "sha256:" + strings.Repeat("a", 64)
	other := "sha256:" + strings.Repeat("b", 64)
	src := strings.Join([]string{
		"// @generated transpile@3 go:pig.(*moveRollDice).Apply " + hash + " first:f0646494",
		"export function pig_moveRollDice_Apply() {}",
		"// @generated agent:model brief:sha256:00 go:pig.(*moveDoneTurn).Apply " + other + " first:f0646494",
		"// @generated transpile@3 go:pig.gone " + hash,
		"// a comment that is not a header",
	}, "\n")
	headers, err := ParseGeneratedHeaders(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 3 || headers[0].Func != "pig.(*moveRollDice).Apply" || headers[0].Line != 1 || headers[1].SourceHash != other {
		t.Fatalf("headers = %+v", headers)
	}
	functions := []gamedef.FunctionEntry{
		{Func: "pig.(*moveRollDice).Apply", SourceHash: hash},
		{Func: "pig.(*moveDoneTurn).Apply", SourceHash: hash},
	}
	diagnostics := StaleGenerated("pig", "bodies.gen.ts", headers, functions)
	if len(diagnostics) != 2 || diagnostics[0].Line != 3 || diagnostics[1].Line != 4 {
		t.Fatalf("stale = %+v, want the changed hash (line 3) and the vanished function (line 4)", diagnostics)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Code != CodeStaleGenerated {
			t.Errorf("code %s", diagnostic.Code)
		}
	}
}
