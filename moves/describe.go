package moves

import (
	"fmt"

	"github.com/jkomoros/boardgame"
)

/*
This file is additive and read-only. It turns the unexported progression group
types (serial, repeat, parallelCount, the leaf move config) and the opaque
ValidCounter closures into plain data, for exporters that need a move
progression as data (fabric/TRANSPILE-PROPOSAL.md, "Where the agreed design has
to change" item 2). Nothing here is called by the engine, and nothing here
changes how any group or counter behaves.

What an exporter needs from an installed move, and where it comes from:

  - The progression group: InstalledProgression(move.Info().CustomConfiguration()),
    below. This was not reachable before: the configuration key is unexported.
  - The legal phases: already reachable as InstalledLegalPhases() on any move
    embedding Default (default.go). Only Default reads that configuration, so a
    move that does not embed Default has no phase restriction to report.
  - Whether the move allows multiple in progression: already reachable by
    asserting the move to interfaces.AllowMultipleInProgression and calling
    AllowMultipleInProgression(), exactly as the leaf group's Satisfied does.
*/

// Progression description kinds.
const (
	ProgressionKindMove     = "move"
	ProgressionKindSerial   = "serial"
	ProgressionKindRepeat   = "repeat"
	ProgressionKindParallel = "parallel"
	// ProgressionKindUnknown marks a group that cannot be described as data:
	// RepeatFromProp (its count is read from live state) or a
	// MoveProgressionGroup implementation from outside this package. Detail
	// says which.
	ProgressionKindUnknown = "unknown"
)

// Counter description kinds. The first seven are the stock families in
// count.go; CounterKindCustom means no stock family reproduced the counter.
const (
	CounterKindAny      = "any"
	CounterKindAll      = "all"
	CounterKindInfinite = "infinite"
	CounterKindAtLeast  = "atLeast"
	CounterKindAtMost   = "atMost"
	CounterKindExactly  = "exactly"
	CounterKindBetween  = "between"
	CounterKindCustom   = "custom"
)

// CounterGridMax is the largest currentCount DescribeCounter evaluates. The
// grid is 0..CounterGridMax inclusive.
const CounterGridMax = 64

// ProgressionDescription is a MoveProgressionGroup as plain data. Its JSON
// keys match ProgressionData in fabric/TRANSPILE-PROPOSAL.md:
//
//	{Kind: "move", Name}
//	{Kind: "serial", Children}
//	{Kind: "repeat", Count, Child}
//	{Kind: "parallel", Count, Children}
//
// plus {Kind: "unknown", Detail}, which the proposal's type does not have: a
// group that is not data is reported, never approximated.
type ProgressionDescription struct {
	Kind string `json:"Kind"`
	// Name is the move name, for a "move" leaf.
	Name string `json:"Name,omitempty"`
	// Children are a serial's or parallel's sub-groups, in order.
	Children []ProgressionDescription `json:"Children,omitempty"`
	// Child is a repeat's repeated group. For an "unknown" RepeatFromProp it
	// is the repeated group too, described for information.
	Child *ProgressionDescription `json:"Child,omitempty"`
	// Count is a repeat's or parallel's counter.
	Count *CounterDescription `json:"Count,omitempty"`
	// Detail explains an "unknown" node.
	Detail string `json:"Detail,omitempty"`
}

// CounterDescription is a ValidCounter as plain data, fitted at one length.
// Its JSON keys match CounterData in fabric/TRANSPILE-PROPOSAL.md. N is set
// for atLeast, atMost and exactly; Min and Max for between. For "custom" the
// proposal's Count FnRef is not known here (package moves cannot name the
// function); Detail says why the fit failed, and the exporter's static half
// supplies the function reference.
type CounterDescription struct {
	Kind string `json:"Kind"`
	N    *int   `json:"N,omitempty"`
	Min  *int   `json:"Min,omitempty"`
	Max  *int   `json:"Max,omitempty"`
	// Detail is set for "custom", and for a stock fit whose parameter lies
	// at the grid's edge (see DescribeCounter).
	Detail string `json:"Detail,omitempty"`
}

// InstalledProgression returns the move progression group installed in a
// move's custom configuration by WithLegalMoveProgression or
// AddOrderedForPhase, or nil if none was. Pass
// move.Info().CustomConfiguration() or a MoveConfig's CustomConfiguration().
// Every move of one AddOrderedForPhase call shares the same group: the
// implied Serial of the groups passed to it.
func InstalledProgression(config boardgame.PropertyCollection) MoveProgressionGroup {
	if config == nil {
		return nil
	}
	group, _ := config[configPropLegalMoveProgression].(MoveProgressionGroup)
	return group
}

// DescribeProgression returns group as plain data. It reads group and never
// calls Satisfied. Every group type this package builds is described:
// Serial; Repeat and Optional (Optional is Repeat(CountBetween(0, 1))) as
// "repeat"; Parallel and ParallelCount as "parallel"; a move config made by
// AutoConfigurer.Config or NewGroupableMoveConfig as "move". RepeatFromProp,
// whose count is resolved from live state, and any MoveProgressionGroup
// implemented outside this package are "unknown" with a Detail.
//
// Counters are described with DescribeCounter at the length the group itself
// passes them: 1 for a repeat (repeatSatisfied calls count(i, 1)), and the
// number of children for a parallel.
func DescribeProgression(group MoveProgressionGroup) ProgressionDescription {
	switch g := group.(type) {
	case nil:
		return ProgressionDescription{Kind: ProgressionKindUnknown, Detail: "nil group"}
	case *defaultMoveConfig:
		if g == nil || g.MoveConfig == nil {
			return ProgressionDescription{Kind: ProgressionKindUnknown, Detail: "nil move config"}
		}
		return ProgressionDescription{Kind: ProgressionKindMove, Name: g.Name()}
	case serial:
		return ProgressionDescription{Kind: ProgressionKindSerial, Children: describeChildren(g)}
	case repeat:
		child := DescribeProgression(g.Child)
		count := DescribeCounter(g.Count, 1)
		return ProgressionDescription{Kind: ProgressionKindRepeat, Count: &count, Child: &child}
	case parallelCount:
		return describeParallel(g)
	case *parallelCount:
		if g == nil {
			return ProgressionDescription{Kind: ProgressionKindUnknown, Detail: "nil parallel group"}
		}
		return describeParallel(*g)
	case repeatFromProp:
		child := DescribeProgression(g.child)
		return ProgressionDescription{
			Kind:   ProgressionKindUnknown,
			Child:  &child,
			Detail: fmt.Sprintf("RepeatFromProp(%q): the repeat count is read from state when the tape is matched, so it is not data", g.path),
		}
	default:
		return ProgressionDescription{
			Kind:   ProgressionKindUnknown,
			Detail: fmt.Sprintf("MoveProgressionGroup of type %T is not built by package moves", group),
		}
	}
}

func describeChildren(children []MoveProgressionGroup) []ProgressionDescription {
	result := make([]ProgressionDescription, len(children))
	for i, child := range children {
		result[i] = DescribeProgression(child)
	}
	return result
}

func describeParallel(p parallelCount) ProgressionDescription {
	count := DescribeCounter(p.Count, len(p.Children))
	return ProgressionDescription{
		Kind:     ProgressionKindParallel,
		Count:    &count,
		Children: describeChildren(p.Children),
	}
}

// counterOutcome is one evaluation of a counter: nil, an error with a text,
// or a panic.
type counterOutcome struct {
	ok       bool
	text     string
	panicked bool
}

func evalCounter(counter ValidCounter, currentCount, length int) (outcome counterOutcome) {
	defer func() {
		if r := recover(); r != nil {
			outcome = counterOutcome{panicked: true, text: fmt.Sprint(r)}
		}
	}()
	err := counter(currentCount, length)
	if err == nil {
		return counterOutcome{ok: true}
	}
	return counterOutcome{text: err.Error()}
}

// counterGrid evaluates counter at currentCount 0..CounterGridMax and length.
func counterGrid(counter ValidCounter, length int) []counterOutcome {
	result := make([]counterOutcome, CounterGridMax+1)
	for c := 0; c <= CounterGridMax; c++ {
		result[c] = evalCounter(counter, c, length)
	}
	return result
}

// matchesGrid reports whether counter, at length, gives want at every point
// of the grid. It stops at the first difference.
func matchesGrid(want []counterOutcome, counter ValidCounter, length int) bool {
	for c, outcome := range want {
		if evalCounter(counter, c, length) != outcome {
			return false
		}
	}
	return true
}

// counterCandidate is one stock family with parameters.
type counterCandidate struct {
	desc CounterDescription
	make func() ValidCounter
}

func intPtr(i int) *int {
	return &i
}

// counterParams is the order DescribeCounter tries a stock parameter in:
// 0..CounterGridMax+1, then -1. On the grid every value above
// CounterGridMax+1 behaves as CounterGridMax+1, and (for atMost and between's
// max, where a negative value means "length") every negative value behaves as
// -1, so these are all the distinct behaviors a parameter can have. Trying the
// non-negative values first means a counter that means "length" is described
// with the literal length when one exists.
func counterParams() []int {
	result := make([]int, 0, CounterGridMax+3)
	for i := 0; i <= CounterGridMax+1; i++ {
		result = append(result, i)
	}
	return append(result, -1)
}

// counterCandidates returns every stock counter DescribeCounter tries, in
// preference order: infinite, all, any, exactly, atLeast, atMost, between;
// within a family, parameters in counterParams order (for between, Min
// outer, Max inner). Min of between only runs 0..CounterGridMax+1: a
// negative Min behaves as 0 on the grid.
func counterCandidates() []counterCandidate {
	params := counterParams()
	candidates := []counterCandidate{
		{CounterDescription{Kind: CounterKindInfinite}, CountInfinite},
		{CounterDescription{Kind: CounterKindAll}, CountAll},
		{CounterDescription{Kind: CounterKindAny}, CountAny},
	}
	for _, n := range params {
		n := n
		candidates = append(candidates, counterCandidate{CounterDescription{Kind: CounterKindExactly, N: intPtr(n)}, func() ValidCounter { return CountExactly(n) }})
	}
	for _, n := range params {
		n := n
		candidates = append(candidates, counterCandidate{CounterDescription{Kind: CounterKindAtLeast, N: intPtr(n)}, func() ValidCounter { return CountAtLeast(n) }})
	}
	for _, n := range params {
		n := n
		candidates = append(candidates, counterCandidate{CounterDescription{Kind: CounterKindAtMost, N: intPtr(n)}, func() ValidCounter { return CountAtMost(n) }})
	}
	for _, min := range params {
		if min < 0 {
			continue
		}
		for _, max := range params {
			min, max := min, max
			candidates = append(candidates, counterCandidate{CounterDescription{Kind: CounterKindBetween, Min: intPtr(min), Max: intPtr(max)}, func() ValidCounter { return CountBetween(min, max) }})
		}
	}
	return candidates
}

// NewCounterFromDescription builds the stock counter desc names. It returns
// an error for "custom" or an unknown kind, and for a missing parameter.
func NewCounterFromDescription(desc CounterDescription) (ValidCounter, error) {
	switch desc.Kind {
	case CounterKindAny:
		return CountAny(), nil
	case CounterKindAll:
		return CountAll(), nil
	case CounterKindInfinite:
		return CountInfinite(), nil
	case CounterKindAtLeast, CounterKindAtMost, CounterKindExactly:
		if desc.N == nil {
			return nil, fmt.Errorf("counter %q has no N", desc.Kind)
		}
		switch desc.Kind {
		case CounterKindAtLeast:
			return CountAtLeast(*desc.N), nil
		case CounterKindAtMost:
			return CountAtMost(*desc.N), nil
		default:
			return CountExactly(*desc.N), nil
		}
	case CounterKindBetween:
		if desc.Min == nil || desc.Max == nil {
			return nil, fmt.Errorf("counter %q needs Min and Max", desc.Kind)
		}
		return CountBetween(*desc.Min, *desc.Max), nil
	}
	return nil, fmt.Errorf("counter kind %q is not a stock counter", desc.Kind)
}

/*
DescribeCounter fits counter, at the given length, to one of the seven stock
families in count.go. It evaluates counter on currentCount 0..CounterGridMax
at that length, and accepts a stock counter only if it gives the same result
and the same error text at every point. If none does, the result is "custom".
A counter that panics anywhere on the grid is "custom".

The description is behaviorally exact at that length, for currentCount in
0..CounterGridMax, and nowhere else. Several stock counters are the same
function on that grid, so DescribeCounter picks among them by a fixed order:
infinite, all, any, exactly, atLeast, atMost, between, and within a family
parameters 0..CounterGridMax+1 and then -1 (Min of between runs only over the
non-negative values). The consequences:

  - all and exactly(length) differ in error text, but all and any are the
    same function at length 1, so a counter that is any at length 1 (every
    repeat counts at length 1) is described as all.
  - atLeast(N) for N <= 0, atMost(N) for N >= CounterGridMax (or N < 0 with
    length >= CounterGridMax) and between(min <= 0, max >= CounterGridMax)
    never fail on the grid, so they are the same function as infinite there
    and are described as infinite.
  - atMost(-1) and between(n, -1) mean "length"; they are described with the
    literal length, which equals them at this length only.
  - A parameter above CounterGridMax+1 is indistinguishable on the grid from
    CounterGridMax+1 and is described as CounterGridMax+1. Such a fit has a
    Detail saying so.

ValidCounter closures from CountAtMost and CountBetween write a negative max
back into their captured variable on first call (count.go), so a counter from
either evaluated at one length and then another keeps the first length. The
groups in this package always evaluate a counter at one length, which is why
fitting at that length is enough; DescribeCounter's own evaluation at length
fixes such a counter's max to length exactly as the group's first evaluation
would.
*/
func DescribeCounter(counter ValidCounter, length int) CounterDescription {
	if counter == nil {
		return CounterDescription{Kind: CounterKindCustom, Detail: "nil counter"}
	}
	want := counterGrid(counter, length)
	for c, outcome := range want {
		if outcome.panicked {
			return CounterDescription{Kind: CounterKindCustom, Detail: fmt.Sprintf("counter panicked at currentCount %d, length %d: %s", c, length, outcome.text)}
		}
	}
	for _, candidate := range counterCandidates() {
		if !matchesGrid(want, candidate.make(), length) {
			continue
		}
		desc := candidate.desc
		if atGridEdge(desc) {
			desc.Detail = fmt.Sprintf("a parameter equals %d, the first value past the grid 0..%d; any larger value behaves the same on the grid, so this description is exact only for currentCount <= %d", CounterGridMax+1, CounterGridMax, CounterGridMax)
		}
		return desc
	}
	return CounterDescription{Kind: CounterKindCustom, Detail: fmt.Sprintf("no stock counter reproduces this counter's results and error texts on currentCount 0..%d at length %d", CounterGridMax, length)}
}

func atGridEdge(desc CounterDescription) bool {
	edge := CounterGridMax + 1
	for _, p := range []*int{desc.N, desc.Min, desc.Max} {
		if p != nil && *p == edge {
			return true
		}
	}
	return false
}
