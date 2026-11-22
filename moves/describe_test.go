package moves_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/werewolf"
	"github.com/jkomoros/boardgame/moves"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

// outcome is one evaluation of a counter: "" for nil, else the error text
// prefixed with "error: " (so an error with empty text differs from nil).
func outcome(counter moves.ValidCounter, currentCount, length int) string {
	err := counter(currentCount, length)
	if err == nil {
		return ""
	}
	return "error: " + err.Error()
}

// gridDifference returns the first currentCount in 0..CounterGridMax where a
// and b differ at length, or -1.
func gridDifference(a, b moves.ValidCounter, length int) (int, string, string) {
	for c := 0; c <= moves.CounterGridMax; c++ {
		x, y := outcome(a, c, length), outcome(b, c, length)
		if x != y {
			return c, x, y
		}
	}
	return -1, "", ""
}

type stockCounter struct {
	name string
	make func() moves.ValidCounter
}

func stockCounters() []stockCounter {
	result := []stockCounter{
		{"CountAny()", moves.CountAny},
		{"CountAll()", moves.CountAll},
		{"CountInfinite()", moves.CountInfinite},
	}
	params := []int{-3, -1, 0, 1, 2, 3, 4, 5, 7, 8, 9, 63, 64, 65, 66, 100}
	for _, n := range params {
		n := n
		result = append(result,
			stockCounter{fmt.Sprintf("CountAtLeast(%d)", n), func() moves.ValidCounter { return moves.CountAtLeast(n) }},
			stockCounter{fmt.Sprintf("CountAtMost(%d)", n), func() moves.ValidCounter { return moves.CountAtMost(n) }},
			stockCounter{fmt.Sprintf("CountExactly(%d)", n), func() moves.ValidCounter { return moves.CountExactly(n) }},
		)
	}
	for _, min := range params {
		for _, max := range params {
			min, max := min, max
			result = append(result, stockCounter{fmt.Sprintf("CountBetween(%d, %d)", min, max), func() moves.ValidCounter { return moves.CountBetween(min, max) }})
		}
	}
	return result
}

// TestDescribeCounterStockRoundTrip is the test the proposal says the counter
// fit needs: every stock counter, with a spread of parameters, at every
// length 0..8, is described as a stock counter, and that description,
// rebuilt through the stock constructors, is the same function as the
// original on currentCount 0..CounterGridMax at that length (results and
// error texts). Each original and each rebuilt counter is fresh per length,
// because CountAtMost and CountBetween fix a negative max to the first length
// they see.
func TestDescribeCounterStockRoundTrip(t *testing.T) {
	fits := 0
	for _, stock := range stockCounters() {
		for length := 0; length <= 8; length++ {
			desc := moves.DescribeCounter(stock.make(), length)
			if desc.Kind == moves.CounterKindCustom {
				t.Errorf("%s at length %d: described as custom (%s)", stock.name, length, desc.Detail)
				continue
			}
			rebuilt, err := moves.NewCounterFromDescription(desc)
			if err != nil {
				t.Errorf("%s at length %d: %+v does not rebuild: %v", stock.name, length, desc, err)
				continue
			}
			if c, want, got := gridDifference(stock.make(), rebuilt, length); c >= 0 {
				t.Errorf("%s at length %d described as %s; differs at currentCount %d: original %q, rebuilt %q", stock.name, length, describeJSON(t, desc), c, want, got)
			}
			// The description survives JSON.
			var decoded moves.CounterDescription
			if err := json.Unmarshal([]byte(describeJSON(t, desc)), &decoded); err != nil {
				t.Fatal(err)
			}
			rebuiltFromJSON, err := moves.NewCounterFromDescription(decoded)
			if err != nil {
				t.Errorf("%s at length %d: JSON %s does not rebuild: %v", stock.name, length, describeJSON(t, desc), err)
				continue
			}
			if c, _, _ := gridDifference(stock.make(), rebuiltFromJSON, length); c >= 0 {
				t.Errorf("%s at length %d: JSON-decoded description differs at currentCount %d", stock.name, length, c)
			}
			fits++
		}
	}
	t.Logf("%d stock counter fits checked", fits)
}

func describeJSON(t *testing.T, v interface{}) string {
	t.Helper()
	blob, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(blob)
}

// TestDescribeCounterPreference pins the documented preference order where
// stock counters are the same function on the grid.
func TestDescribeCounterPreference(t *testing.T) {
	cases := []struct {
		name    string
		counter moves.ValidCounter
		length  int
		want    string
		detail  bool
	}{
		{"all at 3", moves.CountAll(), 3, `{"Kind":"all"}`, false},
		{"exactly(3) at 3 keeps its own error texts", moves.CountExactly(3), 3, `{"Kind":"exactly","N":3}`, false},
		{"any at 1 is all at 1", moves.CountAny(), 1, `{"Kind":"all"}`, false},
		{"any at 2", moves.CountAny(), 2, `{"Kind":"any"}`, false},
		{"all at 1 (Parallel of one child)", moves.CountAll(), 1, `{"Kind":"all"}`, false},
		{"Optional's counter", moves.CountBetween(0, 1), 1, `{"Kind":"between","Min":0,"Max":1}`, false},
		{"atMost(-1) is the literal length", moves.CountAtMost(-1), 3, `{"Kind":"atMost","N":3}`, false},
		{"between(1,-1) is the literal length", moves.CountBetween(1, -1), 4, `{"Kind":"between","Min":1,"Max":4}`, false},
		{"between(-5,2) has min 0", moves.CountBetween(-5, 2), 4, `{"Kind":"between","Min":0,"Max":2}`, false},
		{"atLeast(0) never fails", moves.CountAtLeast(0), 2, `{"Kind":"infinite"}`, false},
		{"atMost(100) never fails on the grid", moves.CountAtMost(100), 2, `{"Kind":"infinite"}`, false},
		{"exactly(-2) always passed", moves.CountExactly(-2), 2, `{"Kind":"exactly","N":-1}`, false},
		{"atLeast(100) past the grid", moves.CountAtLeast(100), 2, `{"Kind":"atLeast","N":65}`, true},
	}
	for _, tc := range cases {
		desc := moves.DescribeCounter(tc.counter, tc.length)
		hasDetail := desc.Detail != ""
		desc.Detail = ""
		if got := describeJSON(t, desc); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
		if hasDetail != tc.detail {
			t.Errorf("%s: Detail present = %v, want %v", tc.name, hasDetail, tc.detail)
		}
	}
}

// TestDescribeCounterCustom: counters no stock family reproduces are custom.
func TestDescribeCounterCustom(t *testing.T) {
	cases := map[string]moves.ValidCounter{
		"even counts": func(currentCount, length int) error {
			if currentCount%2 == 0 {
				return nil
			}
			return errors.New("odd")
		},
		"exactly(2) with its own error text": func(currentCount, length int) error {
			if currentCount == 2 {
				return nil
			}
			return errors.New("not two")
		},
		"atLeast(2) except one error text": func(currentCount, length int) error {
			if currentCount >= 2 {
				return nil
			}
			if currentCount == 1 {
				return errors.New("almost")
			}
			return errors.New("currentCount not yet greater than min configuration")
		},
		"panics": func(currentCount, length int) error {
			if currentCount > 3 {
				panic("boom")
			}
			return nil
		},
		"nil": nil,
	}
	for name, counter := range cases {
		desc := moves.DescribeCounter(counter, 3)
		if desc.Kind != moves.CounterKindCustom {
			t.Errorf("%s: got %s, want custom", name, describeJSON(t, desc))
		}
		if desc.Detail == "" {
			t.Errorf("%s: custom without Detail", name)
		}
		if _, err := moves.NewCounterFromDescription(desc); err == nil {
			t.Errorf("%s: a custom description rebuilt without error", name)
		}
	}
}

// leaf returns a groupable config for a NoOp move named name.
func leaf(name string) moves.GroupableMoveConfig {
	return moves.NewGroupableMoveConfig(boardgame.NewMoveConfig(name, func() boardgame.Move { return new(moves.NoOp) }, nil))
}

type foreignGroup struct{}

func (foreignGroup) MoveConfigs() []boardgame.MoveConfig { return nil }
func (foreignGroup) Satisfied(tape *moves.MoveGroupHistoryItem) (*moves.MoveGroupHistoryItem, error) {
	return tape, nil
}

// TestDescribeProgressionKinds covers every group constructor in package
// moves.
func TestDescribeProgressionKinds(t *testing.T) {
	group := moves.Serial(
		leaf("A"),
		moves.Optional(leaf("B")),
		moves.ParallelCount(moves.CountAtLeast(1), leaf("C"), leaf("D")),
		moves.Parallel(leaf("E"), leaf("F")),
		moves.Repeat(moves.CountExactly(2), leaf("A")),
	)
	want := `{"Kind":"serial","Children":[` +
		`{"Kind":"move","Name":"A"},` +
		`{"Kind":"repeat","Child":{"Kind":"move","Name":"B"},"Count":{"Kind":"between","Min":0,"Max":1}},` +
		`{"Kind":"parallel","Children":[{"Kind":"move","Name":"C"},{"Kind":"move","Name":"D"}],"Count":{"Kind":"atLeast","N":1}},` +
		`{"Kind":"parallel","Children":[{"Kind":"move","Name":"E"},{"Kind":"move","Name":"F"}],"Count":{"Kind":"all"}},` +
		`{"Kind":"repeat","Child":{"Kind":"move","Name":"A"},"Count":{"Kind":"exactly","N":2}}]}`
	if got := describeJSON(t, moves.DescribeProgression(group)); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}

	unknowns := map[string]moves.MoveProgressionGroup{
		"RepeatFromProp": moves.RepeatFromProp("game.Count", leaf("A")),
		"foreign":        foreignGroup{},
		"nil":            nil,
	}
	for name, g := range unknowns {
		desc := moves.DescribeProgression(g)
		if desc.Kind != moves.ProgressionKindUnknown || desc.Detail == "" {
			t.Errorf("%s: got %s, want unknown with a Detail", name, describeJSON(t, desc))
		}
	}
	// Nested: the unknown is not hidden by a parent.
	nested := moves.DescribeProgression(moves.Serial(leaf("A"), moves.RepeatFromProp("game.Count", leaf("B"))))
	if nested.Children[1].Kind != moves.ProgressionKindUnknown || nested.Children[1].Child == nil || nested.Children[1].Child.Name != "B" {
		t.Errorf("nested RepeatFromProp: got %s", describeJSON(t, nested))
	}
	// A custom counter inside a group is reported as custom.
	custom := moves.DescribeProgression(moves.ParallelCount(func(c, l int) error { return errors.New("never") }, leaf("A")))
	if custom.Count == nil || custom.Count.Kind != moves.CounterKindCustom {
		t.Errorf("custom counter in parallel: got %s", describeJSON(t, custom))
	}

	if moves.InstalledProgression(nil) != nil {
		t.Error("InstalledProgression(nil) is not nil")
	}
}

type exampleGame struct {
	name        string
	newDelegate func() boardgame.GameDelegate
}

var exampleGames = []exampleGame{
	{"memory", memory.NewDelegate},
	{"blackjack", blackjack.NewDelegate},
	{"pig", pig.NewDelegate},
	{"tictactoe", tictactoe.NewDelegate},
	{"werewolf", werewolf.NewDelegate},
	{"debuganimations", debuganimations.NewDelegate},
	{"checkers", checkers.NewDelegate},
}

// rebuild turns a description back into a group with the stock constructors,
// using leaves (by name) for move nodes.
func rebuild(t *testing.T, desc moves.ProgressionDescription, leaves map[string]boardgame.MoveConfig) moves.MoveProgressionGroup {
	t.Helper()
	children := func() []moves.MoveProgressionGroup {
		result := make([]moves.MoveProgressionGroup, len(desc.Children))
		for i, child := range desc.Children {
			result[i] = rebuild(t, child, leaves)
		}
		return result
	}
	switch desc.Kind {
	case moves.ProgressionKindMove:
		config := leaves[desc.Name]
		if config == nil {
			t.Fatalf("leaf %q is not among the group's MoveConfigs", desc.Name)
		}
		return moves.NewGroupableMoveConfig(config)
	case moves.ProgressionKindSerial:
		return moves.Serial(children()...)
	case moves.ProgressionKindRepeat:
		count, err := moves.NewCounterFromDescription(*desc.Count)
		if err != nil {
			t.Fatal(err)
		}
		return moves.Repeat(count, rebuild(t, *desc.Child, leaves))
	case moves.ProgressionKindParallel:
		count, err := moves.NewCounterFromDescription(*desc.Count)
		if err != nil {
			t.Fatal(err)
		}
		return moves.ParallelCount(count, children()...)
	}
	t.Fatalf("cannot rebuild %s", desc.Kind)
	return nil
}

// tapeResult is a group's verdict on a tape: the error text, or how many
// items were left unconsumed.
func tapeResult(group moves.MoveProgressionGroup, names []string) string {
	var tape *moves.MoveGroupHistoryItem
	for i := len(names) - 1; i >= 0; i-- {
		tape = &moves.MoveGroupHistoryItem{MoveName: names[i], Rest: tape}
	}
	rest, err := group.Satisfied(tape)
	if err != nil {
		return "error: " + err.Error()
	}
	left := 0
	for ; rest != nil; rest = rest.Rest {
		left++
	}
	return fmt.Sprintf("rest %d", left)
}

// forEachTape calls f with every tape of length 1..maxLen over alphabet.
func forEachTape(alphabet []string, maxLen int, f func([]string)) {
	var walk func(prefix []string)
	walk = func(prefix []string) {
		if len(prefix) > 0 {
			f(prefix)
		}
		if len(prefix) == maxLen {
			return
		}
		for _, name := range alphabet {
			walk(append(append([]string(nil), prefix...), name))
		}
	}
	walk(nil)
}

func walkDescription(desc moves.ProgressionDescription, f func(moves.ProgressionDescription)) {
	f(desc)
	for _, child := range desc.Children {
		walkDescription(child, f)
	}
	if desc.Child != nil {
		walkDescription(*desc.Child, f)
	}
}

// TestDescribeProgressionExamples describes every installed progression of
// every example game. Each must have no unknown node and no custom counter,
// name only installed moves, and, rebuilt through the stock constructors,
// give the same verdict as the installed group on every tape of up to five
// moves drawn from its own moves plus one foreign name.
func TestDescribeProgressionExamples(t *testing.T) {
	withProgression := map[string]int{}
	for _, game := range exampleGames {
		manager, err := boardgame.NewGameManager(game.newDelegate(), storagememory.NewStorageManager())
		if err != nil {
			t.Fatalf("%s: NewGameManager: %v", game.name, err)
		}
		installed := map[string]bool{}
		for _, move := range manager.ExampleMoves() {
			installed[move.Info().Name()] = true
		}
		checked := map[string]bool{}
		for _, move := range manager.ExampleMoves() {
			group := moves.InstalledProgression(move.Info().CustomConfiguration())
			if group == nil {
				continue
			}
			withProgression[game.name]++
			desc := moves.DescribeProgression(group)
			blob := describeJSON(t, desc)
			walkDescription(desc, func(node moves.ProgressionDescription) {
				if node.Kind == moves.ProgressionKindUnknown {
					t.Errorf("%s %q: unknown node: %s", game.name, move.Info().Name(), node.Detail)
				}
				if node.Count != nil && node.Count.Kind == moves.CounterKindCustom {
					t.Errorf("%s %q: custom counter: %s", game.name, move.Info().Name(), node.Count.Detail)
				}
				if node.Kind == moves.ProgressionKindMove && !installed[node.Name] {
					t.Errorf("%s %q: progression names %q, which is not installed", game.name, move.Info().Name(), node.Name)
				}
			})
			if !strings.Contains(blob, fmt.Sprintf(`"Name":%q`, move.Info().Name())) {
				t.Errorf("%s %q: its own progression does not name it: %s", game.name, move.Info().Name(), blob)
			}
			if checked[blob] {
				continue
			}
			checked[blob] = true
			t.Logf("%s: %s", game.name, blob)

			leaves := map[string]boardgame.MoveConfig{}
			var alphabet []string
			for _, config := range group.MoveConfigs() {
				if leaves[config.Name()] == nil {
					alphabet = append(alphabet, config.Name())
				}
				leaves[config.Name()] = config
			}
			alphabet = append(alphabet, "Not In This Progression")
			rebuilt := rebuild(t, desc, leaves)
			tapes := 0
			forEachTape(alphabet, 5, func(names []string) {
				tapes++
				if want, got := tapeResult(group, names), tapeResult(rebuilt, names); want != got {
					t.Errorf("%s %q: tape %q: installed group %s, rebuilt from description %s", game.name, move.Info().Name(), names, want, got)
				}
			})
			t.Logf("%s: %d tapes agree", game.name, tapes)
		}
		manager.Internals().Close()
	}
	for _, name := range []string{"blackjack", "werewolf", "checkers"} {
		if withProgression[name] == 0 {
			t.Errorf("%s: no move has a progression", name)
		}
	}
	t.Logf("moves with a progression, by game: %v", withProgression)
}
