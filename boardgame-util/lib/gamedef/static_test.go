package gamedef

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
)

func loadFixture(t *testing.T) *StaticReport {
	t.Helper()
	report, err := LoadStatic("./testdata/staticgame", "")
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func refs(list []FunctionRef) []string {
	result := []string{}
	for _, fn := range list {
		result = append(result, fn.Func)
	}
	return result
}

func TestStaticFixtureClassification(t *testing.T) {
	report := loadFixture(t)
	const p = "staticgame."
	wantRoles := map[string]string{
		p + "(*gameDelegate).ConfigureComputedProperties.func1": RoleComputed,
		p + "gappedCount":   RoleComputed,
		p + "countKeyed":    RoleHelper,
		p + "readsMoveName": RoleHelper,
		p + "legalHelper":   RoleHelper,
		p + "(*gameDelegate).DistributeComponentToStarterStack": RoleDistribute,
		p + "(*gameDelegate).FinishSetUp":                       RoleSetUpFinish,
		p + "(*moveCode).Legal":                                 RoleLegal,
		p + "(*moveCode).Apply":                                 RoleApply,
		p + "(*moveGap).Apply":                                  RoleApply,
		p + "(*playerState).TurnDone":                           RoleTurnDone,
		p + "(*playerState).ResetForTurnStart":                  RoleResetForTurnStart,
		p + "(*playerState).ResetForTurnEnd":                    RoleResetForTurnEnd,
		p + "(*gameDelegate).ConfigureMoves.func1":              RoleCounter,
	}
	gotRoles := map[string]string{}
	for _, fn := range report.Functions {
		gotRoles[fn.Func] = fn.Role
		if fn.File != "game.go" || fn.Line == 0 || !strings.HasPrefix(fn.SourceHash, SourceHashPrefix) {
			t.Errorf("%s: bad position or hash: %+v", fn.Func, fn)
		}
		if (fn.Verdict == VerdictSubset) != (len(fn.Refusals) == 0) || fn.Refusals == nil || fn.Calls == nil {
			t.Errorf("%s: verdict %s with refusals %v", fn.Func, fn.Verdict, fn.Refusals)
		}
	}
	if !reflect.DeepEqual(gotRoles, wantRoles) {
		t.Errorf("functions and roles:\n got %v\nwant %v", gotRoles, wantRoles)
	}
	if !sort.SliceIsSorted(report.Functions, func(i, j int) bool { return report.Functions[i].Func < report.Functions[j].Func }) {
		t.Error("Functions are not sorted by Func")
	}

	wantBuckets := map[string]string{
		p + "(*gameDelegate).ConfigureComputedProperties.func1": BucketOutside, // callee-refused: readsMoveName
		p + "readsMoveName":     BucketOutside,   // package-var-read
		p + "(*moveCode).Apply": BucketOutside,   // package-var-write
		p + "countKeyed":        BucketEngineGap, // componentAtKey
		p + "gappedCount":       BucketEngineGap, // callee-refused: countKeyed
		p + "legalHelper":       BucketEngineGap, // callee-refused: countKeyed
		p + "(*moveCode).Legal": BucketEngineGap, // callee-refused: legalHelper
		p + "(*gameDelegate).DistributeComponentToStarterStack": BucketSubset,
		p + "(*gameDelegate).FinishSetUp":                       BucketSubset,
		p + "(*moveGap).Apply":                                  BucketSubset,
		p + "(*playerState).TurnDone":                           BucketSubset,
		p + "(*playerState).ResetForTurnStart":                  BucketSubset,
		p + "(*playerState).ResetForTurnEnd":                    BucketSubset,
		p + "(*gameDelegate).ConfigureMoves.func1":              BucketOutside, // custom-counter
	}
	if !reflect.DeepEqual(report.Buckets, wantBuckets) {
		t.Errorf("buckets:\n got %v\nwant %v", report.Buckets, wantBuckets)
	}

	byFunc := map[string]FunctionEntry{}
	for _, fn := range report.Functions {
		byFunc[fn.Func] = fn
	}
	legal := byFunc[p+"(*moveCode).Legal"]
	if len(legal.Refusals) != 1 || legal.Refusals[0].Reason != refusal.CalleeRefused || legal.Refusals[0].Detail != p+"legalHelper" {
		t.Errorf("Legal refusals = %+v, want one callee-refused naming legalHelper", legal.Refusals)
	}
	if !reflect.DeepEqual(legal.Calls, []string{p + "legalHelper"}) {
		t.Errorf("Legal calls = %v; concreteStates and the stock Legal must not appear", legal.Calls)
	}
	if got := byFunc[p+"countKeyed"].Refusals; len(got) != 1 || got[0].Reason != refusal.EngineGap || got[0].Detail != "componentAtKey" {
		t.Errorf("countKeyed refusals = %+v", got)
	}
	if got := byFunc[p+"(*gameDelegate).ConfigureMoves.func1"].Refusals; len(got) != 1 || got[0].Reason != refusal.CustomCounter {
		t.Errorf("counter refusals = %+v", got)
	}
	if len(report.Counters) != 0 {
		t.Errorf("Counters = %v; empty in v1", report.Counters)
	}

	// Callee first: every callee before its callers.
	position := map[string]int{}
	for i, ref := range report.Order {
		position[ref] = i
	}
	if len(report.Order) != len(report.Functions) {
		t.Fatalf("Order has %d entries, want %d", len(report.Order), len(report.Functions))
	}
	for _, fn := range report.Functions {
		for _, callee := range fn.Calls {
			if position[callee] > position[fn.Func] {
				t.Errorf("%s comes before its callee %s", fn.Func, callee)
			}
		}
	}
	if !reflect.DeepEqual(CalleeFirst(report.Functions), report.Order) {
		t.Error("CalleeFirst disagrees with Order")
	}

	if got, want := refs(report.Erased), []string{p + "concreteStates"}; !reflect.DeepEqual(got, want) {
		t.Errorf("erased = %v, want %v", got, want)
	}
	if got, want := refs(report.NotPorted), []string{p + "(*gameDelegate).Diagram", p + "diagramLines"}; !reflect.DeepEqual(got, want) {
		t.Errorf("not ported = %v, want %v", got, want)
	}
	if got, want := refs(report.Unreached), []string{p + "(*gameDelegate).DefaultNumPlayeres", p + "unused"}; !reflect.DeepEqual(got, want) {
		t.Errorf("unreached = %v, want %v", got, want)
	}
	notCounted := strings.Join(refs(report.NotCounted), ",")
	for _, name := range []string{"(*gameDelegate).ConfigureMoves", "(*gameDelegate).Name", "(*gameDelegate).GameStateConstructor", "(*moveCode).HelpText", "NewDelegate"} {
		if !strings.Contains(notCounted, p+name) {
			t.Errorf("not counted %s lacks %s", notCounted, name)
		}
	}
	for _, list := range [][]FunctionRef{report.NotCounted, report.NotPorted, report.Erased, report.Unreached} {
		for _, fn := range list {
			if strings.Contains(fn.Func, "Reader") || strings.Contains(fn.Func, "ReadSetter") {
				t.Errorf("%s from the generated auto_reader.go was counted", fn.Func)
			}
		}
	}
}

func TestStaticFixtureMoves(t *testing.T) {
	report := loadFixture(t)
	type row struct {
		GoType, Stock string
		Ported, Data  bool
		Specs         int
		Methods       string
		Expanded      string
	}
	var got []row
	for _, config := range report.MoveConfigs {
		got = append(got, row{config.GoType, config.StockType, config.StockPorted, config.DataOnly, config.PreconditionSpecs, strings.Join(config.Methods, ","), config.ExpandedFrom})
	}
	want := []row{
		{"staticgame.moveData", "NoOp", true, true, 0, "", ""},
		{"staticgame.moveCode", "CurrentPlayer", true, false, 2, "Apply,HelpText,Legal", ""},
		{"staticgame.moveNested", "CurrentPlayer", true, false, 0, "Apply,HelpText,Legal", ""},
		{"staticgame.moveGap", "DefaultComponent", false, false, 0, "Apply", ""},
		{"moves.ActivateInactivePlayer", "ActivateInactivePlayer", true, true, 0, "", "moves.DefaultRoundSetup"},
		{"moves.WaitForEnoughPlayers", "WaitForEnoughPlayers", true, true, 0, "", "moves.DefaultRoundSetup"},
		{"moves.InactivateEmptySeat", "InactivateEmptySeat", true, true, 0, "", "moves.DefaultRoundSetup"},
		{"moves.StartPhase", "StartPhase", true, true, 0, "", ""},
		{"moves.NoOp", "NoOp", true, true, 0, "", ""},
		{"moves.StartPhase", "StartPhase", true, true, 0, "", ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("move configs:\n got %+v\nwant %+v", got, want)
	}
	if report.MoveConfigs[0].MoveName != "Data Move" || report.MoveConfigs[2].MoveNameSuffix != "Nested" {
		t.Errorf("name options not read: %+v", report.MoveConfigs[:3])
	}
	nested := report.Moves["staticgame.moveNested"]
	if nested.Bodies.Apply != "staticgame.(*moveCode).Apply" || nested.Bodies.Legal != "staticgame.(*moveCode).Legal" || nested.Bodies.Defaults != "" {
		t.Errorf("moveNested bodies = %+v: a method promoted from an embedded game type is the game's body", nested.Bodies)
	}
	if data := report.Moves["staticgame.moveData"]; data.Bodies != (MoveBodies{}) {
		t.Errorf("moveData bodies = %+v, want none", data.Bodies)
	}
	if _, ok := report.Moves["moves.StartPhase"]; !ok {
		t.Error("a configured stock type has no Moves entry")
	}
}

func TestStaticFixtureTraitsAndSlots(t *testing.T) {
	report := loadFixture(t)
	wantTraits := Traits{
		Game: GameTraits{DrawDiscardPair: &DrawDiscardPair{Draw: "Cards", Discard: "Discards"}, RoundRobin: true},
		Player: PlayerTraits{
			Seat: true, GameScore: GameScoreFromScore,
			Turn: &TurnTraits{
				TurnDone:          "staticgame.(*playerState).TurnDone",
				ResetForTurnStart: "staticgame.(*playerState).ResetForTurnStart",
				ResetForTurnEnd:   "staticgame.(*playerState).ResetForTurnEnd",
			},
		},
	}
	if !reflect.DeepEqual(report.Traits, wantTraits) {
		t.Errorf("traits = %+v, want %+v", report.Traits, wantTraits)
	}
	wantComputed := map[string]FnRef{
		"Refused": "staticgame.(*gameDelegate).ConfigureComputedProperties.func1",
		"Gapped":  "staticgame.gappedCount",
	}
	if !reflect.DeepEqual(report.Computed, wantComputed) {
		t.Errorf("computed = %v", report.Computed)
	}
	if report.SetUp != (SetUpData{Distribute: "staticgame.(*gameDelegate).DistributeComponentToStarterStack", Finish: "staticgame.(*gameDelegate).FinishSetUp"}) {
		t.Errorf("set-up = %+v", report.SetUp)
	}
	if report.End != (EndData{}) || report.CurrentPlayerIndex != "" {
		t.Errorf("end = %+v, current player %q; the fixture declares none", report.End, report.CurrentPlayerIndex)
	}
	if report.PackageName != "staticgame" || !strings.HasSuffix(report.Package, "/testdata/staticgame") {
		t.Errorf("package = %q %q", report.Package, report.PackageName)
	}
}

// TestStaticMatchesStaticReport checks that Static returns the report's
// StaticInfo, and that it is deterministic across loads.
func TestStaticIsDeterministic(t *testing.T) {
	first, err := Static("github.com/jkomoros/boardgame/examples/memory")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Static("github.com/jkomoros/boardgame/examples/memory")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Error("two loads of memory differ")
	}
	if first.PackageName != "memory" || len(first.Functions) == 0 {
		t.Fatalf("memory static = %+v", first)
	}
	apply := first.Moves["memory.moveRevealCard"]
	if apply.StockType != "MoveComponentToSlot" || apply.Bodies.Apply != "memory.(*moveRevealCard).Apply" || apply.Bodies.Defaults != "memory.(*moveRevealCard).DefaultsForState" || apply.Bodies.Legal != "" {
		t.Errorf("memory.moveRevealCard = %+v", apply)
	}
}

func TestStaticRejectsNonGame(t *testing.T) {
	if _, err := LoadStatic("github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal", ""); err == nil {
		t.Fatal("a package with no NewDelegate loaded as a game")
	}
}

func TestSourceHashIgnoresLayout(t *testing.T) {
	// gofmt keeps a one-line body on one line, so compare layouts it does
	// normalize: indentation and spacing.
	a := SourceHash([]byte("func f() int {\n\treturn 1\n}"))
	b := SourceHash([]byte("func f()  int {\n        return   1\n}"))
	c := SourceHash([]byte("func f() int {\n\treturn 2\n}"))
	if a != b {
		t.Errorf("layout changed the hash: %s %s", a, b)
	}
	if a == c || !strings.HasPrefix(a, SourceHashPrefix) || len(a) != len(SourceHashPrefix)+64 {
		t.Errorf("hash %s, other body %s", a, c)
	}
}
