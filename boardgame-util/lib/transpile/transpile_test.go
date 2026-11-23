package transpile

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func TestMangleRoundTrip(t *testing.T) {
	cases := map[string]string{
		"pig.(*moveRollDice).Apply":                                "pig_moveRollDice_Apply",
		"pig.concreteStates":                                       "pig_concreteStates",
		"tictactoe.(gameState).tokenValue":                         "tictactoe_gameState$v_tokenValue",
		"memory.(*gameDelegate).ConfigureComputedProperties.func1": "memory_gameDelegate_ConfigureComputedProperties$1",
		"my_pkg.(*some_type).Do_it":                                "my__pkg_some__type_Do__it",
		"a.b__c.func12":                                            "a_b____c$12",
		"x.(T).M.func3":                                            "x_T$v_M$3",
	}
	for ref, want := range cases {
		got, err := Mangle(ref)
		if err != nil {
			t.Errorf("Mangle(%q): %v", ref, err)
			continue
		}
		if got != want {
			t.Errorf("Mangle(%q) = %q, want %q", ref, got, want)
		}
		back, err := Unmangle(got)
		if err != nil || back != ref {
			t.Errorf("Unmangle(%q) = %q, %v; want %q", got, back, err, ref)
		}
	}
	for _, bad := range []string{"nopkg", "p._leading", "p.trailing_", "p.(*T)", "p.(*T).", "p.a-b"} {
		if _, err := Mangle(bad); err == nil {
			t.Errorf("Mangle(%q) should fail", bad)
		}
	}
}

func loadFixture(t *testing.T) *Package {
	t.Helper()
	p, err := Load("./testdata/emitgame", "")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, got, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test -update)", err)
	}
	if !bytes.Equal(got, want) {
		gotLines, wantLines := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
			var g, w string
			if i < len(gotLines) {
				g = gotLines[i]
			}
			if i < len(wantLines) {
				w = wantLines[i]
			}
			if g != w {
				t.Fatalf("%s differs at line %d:\n got: %s\nwant: %s\n(run go test -update after reviewing)", name, i+1, g, w)
			}
		}
	}
}

// TestGolden holds the emitted TypeScript of the fixture's functions, one
// rule each (see testdata/emitgame/game.go), to a reviewed golden file.
func TestGolden(t *testing.T) {
	p := loadFixture(t)
	out, err := p.Generate([]byte("null"), nil, "00000000")
	if err != nil {
		t.Fatal(err)
	}
	for _, fn := range out.Functions {
		if len(fn.Refusals) > 0 {
			t.Errorf("%s refused: %+v", fn.Func, fn.Refusals)
		}
	}
	checkGolden(t, "emitgame.bodies.golden", out.Files["bodies.gen.ts"])
}

// TestGoldenRules checks that the golden holds each rule's translation.
func TestGoldenRules(t *testing.T) {
	p := loadFixture(t)
	out, err := p.Generate([]byte("null"), nil, "00000000")
	if err != nil {
		t.Fatal(err)
	}
	text := string(out.Files["bodies.gen.ts"])
	for _, want := range []string{
		"return 10;",                   // dim*dim + phasePlay, folded
		"let x: number = idiv(a, b);",  // integer division
		"x = x + imod(a, b);",          // %
		"x = imul(x, b);",              // *=
		"strCompare(a, b) < 0",         // string order
		"utf8Len(s) > 1",               // len(string)
		"xs = appendTo(xs, n, n + 1);", // nil-aware append
		"let ys: number[] | null = makeSlice(n, 0);",
		"setAt(ys, 0, at(xs, 1));",                          // bounds-checked index
		"new Map<string, number>()",                         // make(map)
		"let ok: boolean = $m4.has($k3);",                   // comma-ok read
		"counts.delete('b');",                               // delete
		"const $r1 = emitgame_pair(state, 1);",              // tuple results
		"start(2000, nonNil(state.moveByName('Move All')))", // whole milliseconds
		"enumKey(state, 'game', 'Phase') === 0",
		"state.setPhase(1);",
		"setEliminated(state, at(players, next));",
		"state.rand().intn(lenOf(players))",
		"let from: StackRef | null = null;",
		"moveToFirstSlot(first(nonNil(from)), gameStack(state, 'Discards'));",
		"const c: DynamicRef = { deck, index };",
		"throw new Error(nonNil('unknown deck'));",
		"(values(state, c).Value as number) > 10", // values handle
		"if (first_ === null) {",
		"sprintf('score %d is too high for %s'",
		"'could not move: ' + nonNil(err)",
		"let err: string | null = base();",                            // stock Legal
		"stockApply(state, m)",                                        // stock Apply
		"const $r1 = emitgame_gameDelegate_CheckGameFinished(state);", // delegate dispatch
		"variantValue(state, 'label')",
		"continue outer;",
		"case 1:\n\t\tcase 2: {",
		"switch (true) {",
		"return doubled;",
		"Math.min(a, b) + Math.max(a, b)",
		"for (let $i",
		"playerIndexes(state)",
		"components(gameStack(state, 'Cards'))",
		"return [finished, winners];",
		"emitgame_gameDelegate_ConfigureComputedProperties$1(state: RulesState, player_: number): number",
		"imul((player(state, player_).Score as number), 2)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the fixture's output lacks %q", want)
		}
	}
	// DefaultsForState's leading stock call is dropped.
	if !strings.Contains(text, "emitgame_moveAll_DefaultsForState(state: RulesState, m: MoveInstance): void {\n\tm.fields.Amount = 1;\n}") {
		t.Errorf("the stock DefaultsForState call was not dropped")
	}
}

func TestDeterministicAndFirstPreserved(t *testing.T) {
	p := loadFixture(t)
	first, err := p.Generate([]byte("null"), nil, "aaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	again, err := p.Generate([]byte("null"), nil, "aaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range FileNames {
		if !bytes.Equal(first.Files[name], again.Files[name]) {
			t.Errorf("%s differs between two runs", name)
		}
	}
	// A later run at another commit keeps every unchanged block's first:.
	later, err := p.Generate([]byte("null"), first.Files["bodies.gen.ts"], "bbbbbbbb")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Files["bodies.gen.ts"], later.Files["bodies.gen.ts"]) {
		t.Errorf("an unchanged body's first: commit was not kept")
	}
	// A block whose code changed is recorded at the new commit; the others
	// keep theirs.
	edited := bytes.Replace(first.Files["bodies.gen.ts"], []byte("return 10;"), []byte("return 12;"), 1)
	changed, err := p.Generate([]byte("null"), edited, "cccccccc")
	if err != nil {
		t.Fatal(err)
	}
	text := string(changed.Files["bodies.gen.ts"])
	if !strings.Contains(text, "go:emitgame.folded sha256:") || strings.Count(text, "first:cccccccc") != 1 || !strings.Contains(text, "first:aaaaaaaa") {
		t.Errorf("only the changed block should move to the new commit:\n%s", text)
	}
	if blocks := parseBodies(first.Files["bodies.gen.ts"]); len(blocks) != len(first.Functions) {
		t.Errorf("parsed %d blocks of %d functions", len(blocks), len(first.Functions))
	}
}

// TestHeadersMatchLint: every header is one the declarative lint reads, with
// the function's current source hash.
func TestHeadersMatchLint(t *testing.T) {
	p := loadFixture(t)
	out, err := p.Generate([]byte("null"), nil, "00000000")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, entry := range p.Static.Functions {
		hashes[entry.Func] = entry.SourceHash
	}
	blocks := 0
	for _, line := range strings.Split(string(out.Files["bodies.gen.ts"]), "\n") {
		m := header.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		blocks++
		if hashes[m[2]] != m[3] {
			t.Errorf("header for %s names %s, the source is %s", m[2], m[3], hashes[m[2]])
		}
	}
	if blocks != len(p.Static.Functions) {
		t.Errorf("%d headers for %d functions", blocks, len(p.Static.Functions))
	}
}

var exampleGames = []string{"memory", "blackjack", "pig", "tictactoe", "werewolf", "debuganimations", "checkers"}

// TestExamplesAgreeWithLint: for the seven example games, the functions the
// translator writes are exactly those the subset checker accepts
// (gamedef.LoadStatic's verdicts), and the emitter refuses none of them.
func TestExamplesAgreeWithLint(t *testing.T) {
	if testing.Short() {
		t.Skip("loads seven games")
	}
	totalAccepted, totalRefused := 0, 0
	for _, game := range exampleGames {
		p, err := Load("github.com/jkomoros/boardgame/examples/"+game, "")
		if err != nil {
			t.Fatal(err)
		}
		verdicts := map[string]string{}
		for _, entry := range p.Static.Functions {
			verdicts[entry.Func] = entry.Verdict
		}
		functions := p.Translate()
		if len(functions) != len(verdicts) {
			t.Errorf("%s: %d translated or refused, %d in the manifest", game, len(functions), len(verdicts))
		}
		accepted, refused := 0, 0
		for _, fn := range functions {
			if fn.EmitterRefused {
				t.Errorf("%s: the checker accepts %s but the emitter refuses it: %+v", game, fn.Func, fn.Refusals)
			}
			translated := len(fn.Refusals) == 0
			if translated != (verdicts[fn.Func] == gamedef.VerdictSubset) {
				t.Errorf("%s: %s translated=%v, verdict %s", game, fn.Func, translated, verdicts[fn.Func])
			}
			for _, r := range fn.Refusals {
				if !ValidReason(r.Reason) {
					t.Errorf("%s: %s refused for %q, not a closed reason", game, fn.Func, r.Reason)
				}
			}
			if translated {
				accepted++
			} else {
				refused++
			}
		}
		t.Logf("%s: %d translated, %d refused", game, accepted, refused)
		totalAccepted += accepted
		totalRefused += refused
	}
	t.Logf("total: %d translated, %d refused", totalAccepted, totalRefused)
}

// TestCheckedInCurrent is `transpile --check` for the generated games that
// are checked in.
func TestCheckedInCurrent(t *testing.T) {
	root, err := ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, game := range exampleGames {
		dir := filepath.Join(root, filepath.FromSlash(GamesDir), game)
		existing, err := os.ReadFile(filepath.Join(dir, "bodies.gen.ts"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		p, err := Load("github.com/jkomoros/boardgame/examples/"+game, "")
		if err != nil {
			t.Fatal(err)
		}
		definition, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(DefinitionsDir), game+".json"))
		if err != nil {
			t.Fatal(err)
		}
		out, err := p.Generate(definition, existing, "ffffffff")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range FileNames {
			got, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil || !bytes.Equal(got, out.Files[name]) {
				t.Errorf("%s/%s is stale; run boardgame-util transpile ./examples/%s", game, name, game)
			}
		}
		var refused []string
		for _, fn := range out.Functions {
			if len(fn.Refusals) > 0 {
				refused = append(refused, fn.Func)
			}
		}
		if want := pendingRefused[game]; strings.Join(refused, " ") != strings.Join(want, " ") {
			t.Errorf("%s: checked-in game has refused %v, want exactly the pending %v", game, refused, want)
		}
	}
}

// pendingRefused are the functions a checked-in generated game may lack:
// the translator refuses them and they are bound for the agent loop. Until
// each has a body, the game's test supplies a test-side stand-in
// (games/<game>/<game>.gen.test.ts) and the generated index.ts fails to
// load with a GameSpecError naming the FnRef.
var pendingRefused = map[string][]string{
	"memory":   {"memory.(*moveStartHideCardsTimer).Apply"},
	"checkers": {"checkers.(*gameDelegate).ConfigurePredicateConstructors"},
}
