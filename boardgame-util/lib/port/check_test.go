package port

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// testNode is the Node that runs the TypeScript half: $BOARDGAME_NODE, else
// node on PATH. The tests that need it skip unless it is 22.18 or later
// (the version that runs TypeScript); the repository's is 24.
func testNode(t *testing.T) Node {
	t.Helper()
	binary := os.Getenv("BOARDGAME_NODE")
	if binary == "" {
		binary = "node"
	}
	out, err := exec.Command(binary, "--version").Output()
	if err != nil {
		t.Skipf("no node to run the TypeScript half (%v); set BOARDGAME_NODE", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor := 0
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if major < 22 || major == 22 && minor < 18 {
		t.Skipf("node %s cannot run TypeScript; set BOARDGAME_NODE to Node 22.18 or later", version)
	}
	return Node{Binary: binary, Root: moduleRoot(t)}
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return text
}

func reportField(t *testing.T, r *CheckResult, key string) string {
	t.Helper()
	d := r.Node.Differential
	if d == nil || d.First == nil {
		t.Fatalf("no first difference:\n%s", r.Text())
	}
	var report map[string]json.RawMessage
	if err := json.Unmarshal(d.First.Report, &report); err != nil {
		t.Fatalf("report: %v", err)
	}
	return string(report[key])
}

// TestCheckPigAsAgentWritten: one of pig's mechanically translated
// functions, treated as if an agent had written it, PASSes in a correct
// (and differently written) version and fails with the first difference
// when one comparison is mutated: version 1, field state, the busted
// player's Eliminated, with the arguments of the last call.
func TestCheckPigAsAgentWritten(t *testing.T) {
	node := testNode(t)
	g := loadExample(t, "pig")
	opts := CheckOptions{Node: node, Examples: DefaultExamples(g.Root)}
	const ref = "pig.(*moveCountDie).Apply"
	good, err := g.Check(ref, readTestdata(t, "pig_moveCountDie_Apply.ts"), opts)
	if err != nil {
		t.Fatal(err)
	}
	if good.Verdict != VerdictPass {
		t.Fatalf("the correct version: %s", good.Text())
	}
	if !strings.HasSuffix(good.Text(), "\nPASS\n") {
		t.Errorf("port check does not print PASS last:\n%s", good.Text())
	}
	bad, err := g.Check(ref, readTestdata(t, "pig_moveCountDie_Apply.mutant.ts"), opts)
	if err != nil {
		t.Fatal(err)
	}
	if bad.Verdict != VerdictFail {
		t.Fatalf("the mutated version: %s", bad.Text())
	}
	if got := reportField(t, bad, "field"); got != `"state"` {
		t.Errorf("field = %s, want state\n%s", got, bad.Text())
	}
	if got := reportField(t, bad, "path"); !strings.HasSuffix(got, `.Eliminated"`) {
		t.Errorf("path = %s, want a player's Eliminated\n%s", got, bad.Text())
	}
	if last := bad.Node.Differential.LastCall; last == nil || last.Calls == 0 || !bytes.Contains(last.Arguments, []byte(`"Count Die"`)) {
		t.Errorf("the arguments of the last call are missing:\n%s", bad.Text())
	}
	// The candidate is assembled in place of the translator's block, in
	// the switch, under the candidate producer.
	if !bytes.Contains(good.Bodies, []byte("// @generated agent:candidate brief:"+good.Brief.SHA256+" go:"+ref)) {
		t.Errorf("the assembled bodies carry no candidate block")
	}
}

// TestCheckStaticAndTypes: what the static and type steps refuse.
func TestCheckStaticAndTypes(t *testing.T) {
	node := testNode(t)
	g := loadExample(t, "pig")
	opts := CheckOptions{Node: node, Examples: DefaultExamples(g.Root)}
	const ref = "pig.(*moveCountDie).Apply"
	good := string(readTestdata(t, "pig_moveCountDie_Apply.ts"))
	cases := map[string]struct{ text, want string }{
		"a statement at module scope": {good + "\nconst x = 1 * 2;\n", "neither an import nor the function"},
		"a second function":           {good + "\nexport function other(): void {\n}\n", "2 exported functions"},
		"a forbidden import":          {strings.Replace(good, "'../../go/slice.ts'", "'../../rules/stock.ts'", 1), "imports at from ../../rules/stock.ts"},
		"a wrong signature":           {strings.Replace(good, "m: MoveInstance): string | null", "m: MoveInstance): void", 1), "the signature is"},
		"ambient randomness":          {strings.Replace(good, "if (value === 1)", "if (value === 1 && Math.random() < 2)", 1), "uses Math.random"},
		"a type error":                {strings.Replace(good, "+ value;", "+ valu;", 1), "Cannot find name 'valu'"},
	}
	for name, c := range cases {
		result, err := g.Check(ref, []byte(c.text), opts)
		if err != nil {
			t.Fatal(err)
		}
		if result.Verdict != VerdictFail || !strings.Contains(result.Text(), c.want) {
			t.Errorf("%s: want FAIL mentioning %q, got:\n%s", name, c.want, result.Text())
		}
	}
}

// TestCheckMemory is the proposal's acceptance for the harness: memory's
// refused function PASSes in a version written from the hand port
// (games/memory.ts starts the hide-cards timer with the Hide Cards move)
// against the facade, and gives the first difference for a mutated one.
// While memory's generated game does not load (the stock registry builds
// MoveComponentToSlot only with a SlotField the definition lacks) the check
// stops at that precondition, and this test says so and skips.
func TestCheckMemory(t *testing.T) {
	node := testNode(t)
	g := loadExample(t, "memory")
	opts := CheckOptions{Node: node, Examples: DefaultExamples(g.Root)}
	const ref = "memory.(*moveStartHideCardsTimer).Apply"
	good, err := g.Check(ref, readTestdata(t, "memory_moveStartHideCardsTimer_Apply.ts"), opts)
	if err != nil {
		t.Fatal(err)
	}
	if good.Node.Static == nil || !good.Node.Static.OK || good.Node.Types == nil || !good.Node.Types.OK {
		t.Fatalf("the correct version fails the static or type step:\n%s", good.Text())
	}
	if good.Verdict == VerdictPrecondition {
		t.Skipf("memory's generated game cannot be checked yet:\n%s", good.Text())
	}
	if good.Verdict != VerdictPass {
		t.Fatalf("the correct version: %s", good.Text())
	}
	for _, mutant := range []string{"memory_moveStartHideCardsTimer_Apply.wrong-move.ts", "memory_moveStartHideCardsTimer_Apply.duration.ts"} {
		bad, err := g.Check(ref, readTestdata(t, mutant), opts)
		if err != nil {
			t.Fatal(err)
		}
		if bad.Verdict != VerdictFail || bad.Node.Differential == nil || bad.Node.Differential.First == nil {
			t.Errorf("%s: want the first difference, got:\n%s", mutant, bad.Text())
		}
	}
}

// TestCheckCheckersPredicate: a translation of checkers' predicate
// constructor written from its brief alone passes the static and type
// steps and, now that checkers has a checked-in fuzz corpus, the
// differential step. (Before the corpus existed the check stopped, by
// name, at checkers having no recordings.)
func TestCheckCheckersPredicate(t *testing.T) {
	node := testNode(t)
	g := loadExample(t, "checkers")
	result, err := g.Check("checkers.(*gameDelegate).ConfigurePredicateConstructors", readTestdata(t, "checkers_gameDelegate_ConfigurePredicateConstructors.ts"), CheckOptions{Node: node, Examples: DefaultExamples(g.Root)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Node.Static == nil || !result.Node.Static.OK || result.Node.Types == nil || !result.Node.Types.OK {
		t.Fatalf("static or types:\n%s", result.Text())
	}
	if result.Verdict == VerdictPrecondition || result.Verdict == VerdictFail || result.Node.Differential == nil || result.Node.Differential.First != nil {
		t.Errorf("want the differential step to pass, got:\n%s", result.Text())
	}
}
