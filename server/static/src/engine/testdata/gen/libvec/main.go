// Command libvec emits ../../lib.vectors.json: what the Go packages behind
// the engine's lib/ answer, so lib/enumrange.ts, lib/graph.ts, lib/dice.ts
// and lib/playingcards.ts can be held to them.
//
//   - range enums (enum.Set.AddRange): the values, ValueToRange of every key
//     and of keys the enum lacks, RangeToValue of indexes in and out of the
//     grid, SetRangeValue's error, and AddRange's own errors;
//   - grid graphs (graph.NewEnumGridConnectedness) with each filter
//     combination: every node's neighbors, sorted, since Go's order comes
//     from a map and is not fixed; plus checkers' two graphs;
//   - dice: BasicDie, Min, Max, and sequences of Roll on a seeded generator;
//   - playingcards: every deck constructor's components, the generic
//     values, Card.String and the two enums.
//
// Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/libvec
package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/enum/graph"
)

type rangeToValueCase struct {
	Indexes []int
	Key     int64
	// SetError is RangeVal.SetRangeValue(indexes...)'s error text, or "".
	SetError string
}

type valueToRangeCase struct {
	Key int64
	// Indexes is nil (JSON null) where Go returns nil.
	Indexes []int
}

type valueFromStringCase struct {
	Input string
	Key   int64
}

type rangeVec struct {
	Name            string
	Dimensions      []int
	Values          map[string]string
	DefaultValue    int64
	RangeDimensions []int
	ValueToRange    []valueToRangeCase
	RangeToValue    []rangeToValueCase
	ValueFromString []valueFromStringCase
	MarshalledEnum  json.RawMessage
}

type rangeErrorVec struct {
	Dimensions []int
	Error      string
}

type graphVec struct {
	Name       string
	Dimensions []int
	Directions []string
	// Error is NewEnumGridConnectedness's error text, when it fails.
	Error string `json:",omitempty"`
	// Neighbors is each key's NeighborsByKey, sorted; an empty list where
	// Go returns nil.
	Neighbors map[string][]int `json:",omitempty"`
	// Connected samples ConnectedByKey for every pair (from, to) with from
	// in the first 16 keys.
	Connected [][3]int `json:",omitempty"`
}

type rollVec struct {
	Seed  int64
	Faces []int
	// Steps holds the dynamic values after each roll.
	Steps []rollStep
}

type rollStep struct {
	Value, SelectedFace, RollCount int
}

type dieVec struct {
	Min, Max int
	// Faces is nil when BasicDie returns nil.
	Faces    []int
	DieMin   int64
	DieMax   int64
	HasValue bool
}

type deckVec struct {
	Label      string
	Count      int
	WithJokers bool
	Cards      []map[string]string
	Strings    []string
	Generic    map[string]string
}

type output struct {
	Ranges      []rangeVec
	RangeErrors []rangeErrorVec
	Graphs      []graphVec
	Dice        []dieVec
	Rolls       []rollVec
	Decks       []deckVec
	Enums       map[string]json.RawMessage
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var directionFuncs = map[string]graph.EdgeFilter{
	"DirectionUp":            graph.DirectionUp,
	"DirectionDown":          graph.DirectionDown,
	"DirectionLeft":          graph.DirectionLeft,
	"DirectionRight":         graph.DirectionRight,
	"DirectionPerpendicular": graph.DirectionPerpendicular,
	"DirectionDiagonal":      graph.DirectionDiagonal,
}

func keysOf(e enum.Enum) []enum.EnumKey {
	keys := e.Values()
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func rangeVector(name string, dims ...int) rangeVec {
	set := enum.NewSet()
	e := set.MustAddRange(name, dims...)
	v := rangeVec{Name: name, Dimensions: dims, Values: map[string]string{}, DefaultValue: int64(e.DefaultValue()), RangeDimensions: e.RangeDimensions()}
	keys := keysOf(e)
	for _, k := range keys {
		v.Values[itoa(int64(k))] = e.String(k)
	}
	probe := []enum.EnumKey{-1, enum.EnumKey(len(keys)), enum.EnumKey(len(keys) + 7), enum.IllegalValue}
	for _, k := range append(keys, probe...) {
		v.ValueToRange = append(v.ValueToRange, valueToRangeCase{Key: int64(k), Indexes: e.ValueToRange(k)})
	}
	// Every in-grid tuple, plus one step outside each edge, plus tuples of
	// the wrong arity.
	var tuples [][]int
	var walk func(prefix []int)
	walk = func(prefix []int) {
		if len(prefix) == len(dims) {
			tuples = append(tuples, append([]int(nil), prefix...))
			return
		}
		d := dims[len(prefix)]
		for i := -1; i <= d; i++ {
			walk(append(prefix, i))
		}
	}
	if len(keys) <= 100 {
		walk(nil)
	} else {
		for _, k := range keys {
			tuples = append(tuples, e.ValueToRange(k))
		}
		tuples = append(tuples, []int{-1, 0}, []int{0, -1}, []int{dims[0], 0}, []int{0, dims[len(dims)-1]})
	}
	tuples = append(tuples, []int{}, []int{0, 0, 0, 0}, []int{1})
	for _, t := range tuples {
		c := rangeToValueCase{Indexes: t, Key: int64(e.RangeToValue(t...))}
		val := e.NewRangeVal()
		if err := val.SetRangeValue(t...); err != nil {
			c.SetError = err.Error()
		}
		v.RangeToValue = append(v.RangeToValue, c)
	}
	for _, s := range []string{"0", "1", "+1", "-0", "007", " 0_1 ", "0_1", "1_0", "0-1", "", "x", "2", "5", "99", "0_0_0"} {
		v.ValueFromString = append(v.ValueFromString, valueFromStringCase{Input: s, Key: int64(e.ValueFromString(s))})
	}
	blob, err := json.Marshal(e)
	check(err)
	v.MarshalledEnum = blob
	return v
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func graphVector(name string, dims []int, directions ...string) graphVec {
	v := graphVec{Name: name, Dimensions: dims, Directions: directions}
	set := enum.NewSet()
	e := set.MustAddRange(name, dims...)
	var filters []graph.EdgeFilter
	for _, d := range directions {
		filters = append(filters, directionFuncs[d])
	}
	g, err := graph.NewEnumGridConnectedness(e, filters...)
	if err != nil {
		v.Error = err.Error()
		return v
	}
	v.Neighbors = map[string][]int{}
	keys := keysOf(e)
	for _, k := range keys {
		ns := []int{}
		for _, n := range g.NeighborsByKey(k) {
			ns = append(ns, int(n))
		}
		sort.Ints(ns)
		v.Neighbors[itoa(int64(k))] = ns
	}
	for i, from := range keys {
		if i >= 16 {
			break
		}
		for _, to := range keys {
			c := 0
			if g.ConnectedByKey(from, to) {
				c = 1
			}
			v.Connected = append(v.Connected, [3]int{int(from), int(to), c})
		}
	}
	return v
}

// newDie gives a DynamicValue whose containing component is a die with
// these faces, through a real deck, as Roll requires.
func newDie(faces []int) *dice.DynamicValue {
	deck := boardgame.NewDeck()
	deck.AddComponent(&dice.Value{Faces: faces})
	d := &dice.DynamicValue{}
	d.SetContainingComponent(deck.Components()[0])
	return d
}

func cardValues(deck *boardgame.Deck) ([]map[string]string, []string) {
	var cards []map[string]string
	var strs []string
	for _, c := range deck.Components() {
		card := c.Values().(*playingcards.Card)
		cards = append(cards, map[string]string{"Suit": card.Suit.String(), "Rank": card.Rank.String()})
		strs = append(strs, card.String())
	}
	return cards, strs
}

func deckVector(label string, count int, withJokers bool, multi bool) deckVec {
	var deck *boardgame.Deck
	if multi {
		deck = playingcards.NewDeckMulti(count, withJokers)
	} else {
		deck = playingcards.NewDeck(withJokers)
	}
	cards, strs := cardValues(deck)
	generic := deck.GenericComponent().Values().(*playingcards.Card)
	return deckVec{Label: label, Count: count, WithJokers: withJokers, Cards: cards, Strings: strs,
		Generic: map[string]string{"Suit": generic.Suit.String(), "Rank": generic.Rank.String()}}
}

func main() {
	var out output

	out.Ranges = append(out.Ranges,
		rangeVector("single", 3),
		rangeVector("double", 2, 3),
		rangeVector("triple", 2, 3, 4),
		rangeVector("unit", 1, 1),
		rangeVector("spaces", 8, 8),
		rangeVector("wide", 12),
	)
	for _, dims := range [][]int{{}, {0}, {3, -1}, {2, 0, 2}} {
		set := enum.NewSet()
		_, err := set.AddRange("bad", dims...)
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		out.RangeErrors = append(out.RangeErrors, rangeErrorVec{Dimensions: dims, Error: msg})
	}

	names := []string{"DirectionUp", "DirectionDown", "DirectionLeft", "DirectionRight", "DirectionPerpendicular", "DirectionDiagonal"}
	out.Graphs = append(out.Graphs, graphVector("grid", []int{8, 8}))
	for _, n := range names {
		out.Graphs = append(out.Graphs, graphVector("grid", []int{5, 4}, n))
	}
	// checkers' two graphs (examples/checkers/components.go).
	out.Graphs = append(out.Graphs,
		graphVector("spaces", []int{8, 8}, "DirectionDown", "DirectionDiagonal"),
		graphVector("spaces", []int{8, 8}, "DirectionUp", "DirectionDiagonal"),
		graphVector("grid", []int{3, 3}, "DirectionUp", "DirectionLeft"),
		graphVector("grid", []int{1, 1}),
		graphVector("grid", []int{1, 5}, "DirectionRight"),
		graphVector("line", []int{5}),
		graphVector("cube", []int{2, 2, 2}),
	)

	for _, mm := range [][2]int{{1, 6}, {0, 1}, {3, 3}, {5, 2}, {-2, 2}, {1, 20}} {
		d := dice.BasicDie(mm[0], mm[1])
		v := dieVec{Min: mm[0], Max: mm[1], HasValue: d != nil}
		if d != nil {
			v.Faces = d.Faces
			v.DieMin = int64(d.Min())
			v.DieMax = int64(d.Max())
		}
		out.Dice = append(out.Dice, v)
	}
	empty := &dice.Value{}
	out.Dice = append(out.Dice, dieVec{Min: 0, Max: 0, Faces: []int{}, DieMin: int64(empty.Min()), DieMax: int64(empty.Max()), HasValue: true})
	out.Dice = append(out.Dice, dieVec{Min: 1, Max: 6, Faces: dice.DefaultDie().Faces, DieMin: int64(dice.DefaultDie().Min()), DieMax: int64(dice.DefaultDie().Max()), HasValue: true})

	for _, c := range []struct {
		seed  int64
		faces []int
	}{
		{1, []int{1, 2, 3, 4, 5, 6}},
		{42, []int{1, 2, 3, 4, 5, 6}},
		{-7, []int{2, 4, 8}},
		{123456789, []int{10}},
		{99, []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144}},
	} {
		r := rand.New(rand.NewSource(c.seed))
		d := newDie(c.faces)
		v := rollVec{Seed: c.seed, Faces: c.faces}
		for i := 0; i < 25; i++ {
			d.Roll(r)
			v.Steps = append(v.Steps, rollStep{Value: d.Value, SelectedFace: d.SelectedFace, RollCount: d.RollCount})
		}
		out.Rolls = append(out.Rolls, v)
	}

	out.Decks = append(out.Decks,
		deckVector("NewDeck(false)", 1, false, false),
		deckVector("NewDeck(true)", 1, true, false),
		deckVector("NewDeckMulti(2, false)", 2, false, true),
		deckVector("NewDeckMulti(0, true)", 0, true, true),
	)
	out.Enums = map[string]json.RawMessage{}
	for name, e := range map[string]enum.Enum{"Suit": playingcards.SuitEnum, "Rank": playingcards.RankEnum} {
		blob, err := json.Marshal(e)
		check(err)
		out.Enums[name] = blob
	}

	blob, err := json.MarshalIndent(out, "", " ")
	check(err)
	check(os.WriteFile("server/static/src/engine/testdata/lib.vectors.json", append(blob, '\n'), 0644))
}
