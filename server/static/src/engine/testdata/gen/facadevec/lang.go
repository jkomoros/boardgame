package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/enum/graph"
)

// The language half of the facade's known answers: Go's integer
// arithmetic, strconv, strings, fmt, the runtime's panic texts, nil-aware
// slices, three behaviors' setters, range enums and grid graphs.

type divCase struct {
	A, B     int
	Quo, Rem int
	QuoPanic string `json:",omitempty"`
	RemPanic string `json:",omitempty"`
}

type mulCase struct {
	A, B    int
	Product int
	// Exact is false when the true product leaves int64 (Go wraps) or the
	// safe-integer range (the facade throws instead).
	Exact bool
}

type atoiCase struct {
	S     string
	Value int
	Err   string
	// Safe is false when Go accepted a value that is not a safe integer:
	// the facade's documented divergence.
	Safe bool
}

type itoaCase struct {
	N int
	S string
}

type compareCase struct {
	A, B    string
	Compare int
	Less    bool
}

type lenCase struct {
	S   string
	Len int
}

type arg struct {
	T string // int, float64, string, bool, nil, []int, []string
	V json.RawMessage
}

type sprintfCase struct {
	Format string
	Args   []arg
	Out    string
}

type stringsCase struct {
	Fn     string
	Args   []json.RawMessage
	Result json.RawMessage `json:",omitempty"`
	Panic  string          `json:",omitempty"`
}

type panicCase struct {
	Kind string
	Args []int `json:",omitempty"`
	Text string
}

type sliceCase struct {
	Name string
	JSON json.RawMessage
}

type rangeEnumVec struct {
	Name         string
	Dimensions   []int
	Values       map[string]string
	DefaultValue int
	ValueToRange []valueToRange
	RangeToValue []rangeToValue
}

type valueToRange struct {
	Key     int
	Indexes []int
}

type rangeToValue struct {
	Indexes []int
	Key     int
	Illegal bool
}

type graphVec struct {
	Enum       string
	Dimensions []int
	Directions []string
	// Neighbors by key, sorted ascending (Go's order is a map's), or null
	// where Go returns nil. Keys -1 and len are included.
	Neighbors map[string][]int
}

type langVectors struct {
	Div          []divCase
	Mul          []mulCase
	Atoi         []atoiCase
	Itoa         []itoaCase
	Compare      []compareCase
	Len          []lenCase
	Sprintf      []sprintfCase
	Strings      []stringsCase
	Panics       []panicCase
	Slices       []sliceCase
	Behaviors    map[string]json.RawMessage
	IllegalValue string
	RangeEnums   []rangeEnumVec
	Graphs       []graphVec
}

// recovered runs f and returns the text of the panic it raised, or "".
func recovered(f func()) (text string) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				text = err.Error()
			} else {
				text = fmt.Sprint(r)
			}
		}
	}()
	f()
	return ""
}

func raw(v interface{}) json.RawMessage {
	blob, err := json.Marshal(v)
	check(err)
	return blob
}

const maxSafe = 1<<53 - 1

func safe(n int) bool {
	return n >= -maxSafe && n <= maxSafe
}

func langVecs() langVectors {
	var out langVectors

	// ----- integer arithmetic ------------------------------------------
	operands := []int{7, -7, 0, 1, -1, 6, -6, maxSafe, -maxSafe}
	divisors := []int{2, -2, 1, -1, 0, 3, -3, 7}
	for _, a := range operands {
		for _, b := range divisors {
			c := divCase{A: a, B: b}
			c.QuoPanic = recovered(func() { c.Quo = a / b })
			c.RemPanic = recovered(func() { c.Rem = a % b })
			out.Div = append(out.Div, c)
		}
	}
	factors := []int{0, 1, -1, 3, -3, 1 << 26, -(1 << 26), 1 << 27, 94906265, maxSafe}
	for _, a := range factors {
		for _, b := range factors {
			hi, lo := bits.Mul64(uint64(absInt(a)), uint64(absInt(b)))
			product := a * b
			exact := hi == 0 && lo <= maxSafe
			out.Mul = append(out.Mul, mulCase{A: a, B: b, Product: product, Exact: exact})
		}
	}

	// ----- strconv -----------------------------------------------------
	for _, s := range []string{
		"0", "-0", "+0", "7", "123", "-123", "+5", "007", "", "+", "-", "+-1", " 1", "1 ", "1_000", "0x10", "12a", "1e3", "1.5",
		"9007199254740991", "-9007199254740991", "9007199254740992", "-9007199254740993",
		"9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809",
		"18446744073709551615", "18446744073709551616", "99999999999999999999x", "9999999999999999999x",
		"00000000000000000000000000001", "-00000000000000000000000000000", "\u0661\u0662", "x\"y", "tab\t", "\u00e9", "\U0001f6b4",
		"123456789012345678", "1234567890123456789",
	} {
		v, err := strconv.Atoi(s)
		c := atoiCase{S: s, Value: v, Safe: true}
		if err != nil {
			c.Err = err.Error()
		} else {
			c.Safe = safe(v)
		}
		out.Atoi = append(out.Atoi, c)
	}
	for _, n := range []int{0, -1, 1, 42, -42, 1000000, maxSafe, -maxSafe} {
		out.Itoa = append(out.Itoa, itoaCase{N: n, S: strconv.Itoa(n)})
	}

	// ----- comparison and length ---------------------------------------
	words := []string{"", "a", "b", "ab", "B", "Z", "z", "\u00e9", "e\u0301", "\uffff", "\ue000", "\ud7ff", "\U0001f600", "\U0001f6b4", "a\U0001f600", "a\uffff", "\U0001F600b", "\u65e5\u672c"}
	for _, a := range words {
		for _, b := range words {
			out.Compare = append(out.Compare, compareCase{A: a, B: b, Compare: strings.Compare(a, b), Less: a < b})
		}
		out.Len = append(out.Len, lenCase{S: a, Len: len(a)})
	}
	if !utf8.ValidString(strings.Join(words, "")) {
		panic("invalid UTF-8 in words")
	}

	// ----- fmt ---------------------------------------------------------
	I := func(n int) arg { return arg{T: "int", V: raw(n)} }
	F := func(f float64) arg { return arg{T: "float64", V: raw(strconv.FormatFloat(f, 'g', -1, 64))} }
	S := func(s string) arg { return arg{T: "string", V: raw(s)} }
	B := func(b bool) arg { return arg{T: "bool", V: raw(b)} }
	N := arg{T: "nil", V: raw(nil)}
	IS := func(xs ...int) arg { return arg{T: "[]int", V: raw(xs)} }
	SS := func(xs ...string) arg { return arg{T: "[]string", V: raw(xs)} }
	formats := []struct {
		f    string
		args []arg
	}{
		{"plain", nil},
		{"%d", []arg{I(42)}},
		{"%d", []arg{I(-7)}},
		{"%d/%d", []arg{I(1), I(3)}},
		{"Round: %d/%d", []arg{I(2), I(5)}},
		{"%5d|%-5d|%05d|%+d|%+d", []arg{I(42), I(42), I(-42), I(3), I(-3)}},
		{"%0-5d|%-05d|", []arg{I(7), I(7)}},
		{"%s", []arg{S("hi")}},
		{"%s and %s", []arg{S("a"), S("\U0001f6b4")}},
		{"%6s|%-6s|%06s", []arg{S("ab"), S("ab"), S("ab")}},
		{"%3s|", []arg{S("\u65e5\u672c\u8a9e")}},
		{"%v %v %v", []arg{I(5), S("x"), B(true)}},
		{"%v", []arg{B(false)}},
		{"%v", []arg{N}},
		{"%+v|%5v|%-5v|%05v", []arg{I(5), I(5), S("ab"), I(5)}},
		{"%v %v %v %v %v %v %v %v", []arg{F(1.5), F(0.1), F(-2.25), F(1e-5), F(123456.5), F(1234567.5), F(2.5e-7), F(-0.000123)}},
		{"%v", []arg{F(3.0000001)}},
		{"%q", []arg{S("hello")}},
		{"%q", []arg{S("a\"b\\c\nd\te\x01\x7f")}},
		{"%q", []arg{S("\u00e9\U0001f6b4\u65e5\u672c\u00a0\u200b\u2028")}},
		{"%q %q %q %q", []arg{I(65), I('\''), I('"'), I(10)}},
		{"%q %q", []arg{I(0x1F6B4), I(-1)}},
		{"%v", []arg{IS(1, 2, 3)}},
		{"%d", []arg{IS(4, 5)}},
		{"%v", []arg{SS("a", "b")}},
		{"%s", []arg{SS("a", "b")}},
		{"%q", []arg{SS("a", "b")}},
		{"%v", []arg{IS([]int{}...)}},
		{"%d", []arg{S("x")}},
		{"%d", []arg{B(true)}},
		{"%d", []arg{N}},
		{"%d", []arg{F(1.5)}},
		{"%s", []arg{I(5)}},
		{"%s", []arg{B(true)}},
		{"%s", []arg{N}},
		{"%q", []arg{B(false)}},
		{"%q", []arg{N}},
		{"%d", []arg{SS("a")}},
		{"%d %d", []arg{I(1)}},
		{"%d", []arg{I(1), I(2)}},
		{"%d", []arg{I(1), S("x"), N, B(true)}},
		{"", []arg{I(1)}},
		{"100%%", nil},
		{"%d%%", []arg{I(50)}},
		{"trailing %", nil},
		{"trailing %5", nil},
		{"%z", []arg{I(1)}},
		{"%!", []arg{I(1)}},
		{"%v", []arg{F(math.Inf(1))}},
		{"%v", []arg{F(math.Inf(-1))}},
	}
	for _, f := range formats {
		values := make([]interface{}, len(f.args))
		for i, a := range f.args {
			values[i] = goValue(a)
		}
		out.Sprintf = append(out.Sprintf, sprintfCase{Format: f.f, Args: f.args, Out: fmt.Sprintf(f.f, values...)})
	}

	// ----- strings -----------------------------------------------------
	addStrings := func(fn string, args []interface{}, call func() interface{}) {
		c := stringsCase{Fn: fn}
		for _, a := range args {
			c.Args = append(c.Args, raw(a))
		}
		var result interface{}
		c.Panic = recovered(func() { result = call() })
		if c.Panic == "" {
			c.Result = raw(result)
		}
		out.Strings = append(out.Strings, c)
	}
	for _, s := range []string{"", "ABC", "Hello, World", "\u00c0\u00c9\u00ce", "\u03a3\u0391\u03a3", "\u039f\u0394\u039f\u03a3", "\u0130stanbul", "\u01c5", "\u1e9e", "\U0001f600X", "\ua64a", "\U00010400"} {
		addStrings("ToLower", []interface{}{s}, func() interface{} { return strings.ToLower(s) })
	}
	for _, s := range []string{"", " a ", "\t\n a b \r\v\f", "\u0085x\u0085", "\u00a0x\u00a0", "\ufeffx\ufeff", "\u2003x\u3000", "\u200bx", "\u1680x\u2028\u2029\u202f\u205f", "   ", "x"} {
		addStrings("TrimSpace", []interface{}{s}, func() interface{} { return strings.TrimSpace(s) })
	}
	splits := [][2]string{{"a,b,c", ","}, {"a,b,", ","}, {"", ","}, {"abc", ""}, {"\U0001f6b4\u00e9", ""}, {"", ""}, {"a--b----c", "--"}, {"abc", "abc"}, {"abc", "x"}}
	for _, sp := range splits {
		s, sep := sp[0], sp[1]
		addStrings("Split", []interface{}{s, sep}, func() interface{} { return strings.Split(s, sep) })
	}
	joins := []struct {
		elems []string
		sep   string
	}{{nil, ","}, {[]string{}, ","}, {[]string{"a"}, ","}, {[]string{"a", "b", "c"}, ", "}, {[]string{"", ""}, "-"}}
	for _, j := range joins {
		elems, sep := j.elems, j.sep
		addStrings("Join", []interface{}{elems, sep}, func() interface{} { return strings.Join(elems, sep) })
	}
	for _, r := range []struct {
		s string
		n int
	}{{"ab", 3}, {"ab", 0}, {"", 5}, {"\U0001f6b4", 2}, {"a", -1}} {
		s, n := r.s, r.n
		addStrings("Repeat", []interface{}{s, n}, func() interface{} { return strings.Repeat(s, n) })
	}
	pairs := [][2]string{{"hello", "he"}, {"hello", "lo"}, {"hello", ""}, {"", ""}, {"", "a"}, {"hello", "hello!"}, {"\U0001f6b4\u00e9", "\U0001f6b4"}, {"\U0001f6b4\u00e9", "\u00e9"}, {"abc", "b"}, {"abc", "d"}}
	for _, p := range pairs {
		a, b := p[0], p[1]
		addStrings("HasPrefix", []interface{}{a, b}, func() interface{} { return strings.HasPrefix(a, b) })
		addStrings("HasSuffix", []interface{}{a, b}, func() interface{} { return strings.HasSuffix(a, b) })
		addStrings("Contains", []interface{}{a, b}, func() interface{} { return strings.Contains(a, b) })
	}

	// ----- runtime panics ----------------------------------------------
	for _, ix := range [][2]int{{3, 3}, {5, 3}, {-1, 3}, {0, 0}, {-5, 0}, {2, 3}} {
		i, n := ix[0], ix[1]
		xs := make([]int, n)
		out.Panics = append(out.Panics, panicCase{Kind: "index", Args: []int{i, n}, Text: recovered(func() { _ = xs[i] })})
	}
	for _, sl := range [][3]int{{0, 4, 3}, {2, 1, 3}, {-1, 2, 3}, {0, -1, 3}, {-2, -1, 3}, {4, 5, 3}, {4, 3, 3}, {0, 3, 3}, {1, 2, 3}, {0, 1, 0}} {
		lo, hi, n := sl[0], sl[1], sl[2]
		xs := make([]int, n)
		out.Panics = append(out.Panics, panicCase{Kind: "slice", Args: []int{lo, hi, n}, Text: recovered(func() { _ = xs[lo:hi] })})
	}
	out.Panics = append(out.Panics, panicCase{Kind: "nilDeref", Text: recovered(func() {
		var p *struct{ X int }
		_ = p.X
	})})
	out.Panics = append(out.Panics, panicCase{Kind: "nilInterface", Text: recovered(func() {
		var c boardgame.ImmutableComponentInstance
		_ = c.DeckIndex()
	})})
	out.Panics = append(out.Panics, panicCase{Kind: "nilMap", Text: recovered(func() {
		var m map[string]int
		m["a"] = 1
	})})
	zero := 0
	out.Panics = append(out.Panics, panicCase{Kind: "divide", Text: recovered(func() { _ = 1 / zero })})
	out.Panics = append(out.Panics, panicCase{Kind: "modulo", Text: recovered(func() { _ = 1 % zero })})
	negative := -1
	out.Panics = append(out.Panics, panicCase{Kind: "makeslice", Args: []int{negative}, Text: recovered(func() { _ = make([]int, negative) })})

	// ----- nil-aware slices --------------------------------------------
	type holder struct{ X []int }
	slice := func(name string, f func() []int) {
		out.Slices = append(out.Slices, sliceCase{Name: name, JSON: raw(holder{f()})})
	}
	slice("var", func() []int { var x []int; return x })
	slice("appendNothingToNil", func() []int { var x, none []int; x = append(x, none...); return x })
	slice("appendToNil", func() []int { var x []int; x = append(x, 1, 2); return x })
	slice("literalEmpty", func() []int { return []int{} })
	slice("makeZero", func() []int { return make([]int, 0) })
	slice("makeThree", func() []int { return make([]int, 3) })
	slice("appendNothingToEmpty", func() []int { x := []int{}; var none []int; x = append(x, none...); return x })
	slice("assignNil", func() []int { x := []int{1}; x = nil; return x })
	slice("subsliceNil", func() []int { var x []int; return x[0:0] })
	slice("subsliceEmpty", func() []int { x := []int{1, 2, 3}; return x[1:1] })
	slice("subsliceMiddle", func() []int { x := []int{1, 2, 3}; return x[1:3] })
	slice("lenNil", func() []int { var x []int; return []int{len(x)} })
	slice("copyShort", func() []int { dst := make([]int, 2); n := copy(dst, []int{7, 8, 9}); return append(dst, n) })
	slice("copyLong", func() []int { dst := []int{1, 2, 3, 4}; n := copy(dst, []int{7, 8}); return append(dst, n) })
	slice("copyNil", func() []int { dst := []int{1}; var src []int; n := copy(dst, src); return append(dst, n) })
	slice("copyOverlap", func() []int { x := []int{1, 2, 3, 4}; copy(x[1:], x); return x })
	slice("aliasAfterAppend", func() []int { a := []int{1, 2}; b := a; a = append(a, 3); return []int{len(a), len(b)} })

	// ----- behaviors ---------------------------------------------------
	out.Behaviors = map[string]json.RawMessage{}
	var seat behaviors.Seat
	seat.SetSeatFilled()
	out.Behaviors["SetSeatFilled"] = raw(seat)
	var elim behaviors.PlayerElimination
	elim.SetEliminated()
	out.Behaviors["SetEliminated"] = raw(elim)
	var inactive behaviors.InactivePlayer
	inactive.SetPlayerInactive()
	out.Behaviors["SetPlayerInactive"] = raw(inactive)

	// ----- range enums and grid graphs ---------------------------------
	out.IllegalValue = strconv.FormatInt(int64(enum.IllegalValue), 10)
	set := enum.NewSet()
	ranges := []struct {
		name string
		dims []int
	}{{"line", []int{5}}, {"grid", []int{3, 4}}, {"board", []int{8, 8}}, {"cube", []int{2, 3, 2}}}
	enums := map[string]enum.RangeEnum{}
	for _, r := range ranges {
		e := set.MustAddRange(r.name, r.dims...)
		enums[r.name] = e
		vec := rangeEnumVec{Name: r.name, Dimensions: e.RangeDimensions(), Values: map[string]string{}, DefaultValue: int(e.DefaultValue())}
		for _, k := range e.Values() {
			vec.Values[strconv.Itoa(int(k))] = e.String(k)
		}
		n := len(e.Values())
		for k := -1; k <= n; k++ {
			vec.ValueToRange = append(vec.ValueToRange, valueToRange{Key: k, Indexes: e.ValueToRange(enum.EnumKey(k))})
		}
		var tuples [][]int
		switch len(r.dims) {
		case 1:
			for i := -1; i <= r.dims[0]; i++ {
				tuples = append(tuples, []int{i})
			}
			tuples = append(tuples, []int{0, 0}, []int{})
		case 2:
			for i := -1; i <= r.dims[0]; i++ {
				for j := -1; j <= r.dims[1]; j++ {
					tuples = append(tuples, []int{i, j})
				}
			}
			tuples = append(tuples, []int{1}, []int{1, 1, 1}, []int{})
		default:
			for i := -1; i <= r.dims[0]; i++ {
				for j := 0; j <= r.dims[1]; j++ {
					for k := 0; k < r.dims[2]; k++ {
						tuples = append(tuples, []int{i, j, k})
					}
				}
			}
			tuples = append(tuples, []int{1, 1})
		}
		for _, t := range tuples {
			k := e.RangeToValue(t...)
			c := rangeToValue{Indexes: t, Illegal: k == enum.IllegalValue}
			if !c.Illegal {
				c.Key = int(k)
			}
			vec.RangeToValue = append(vec.RangeToValue, c)
		}
		out.RangeEnums = append(out.RangeEnums, vec)
	}
	filters := map[string]graph.EdgeFilter{
		"DirectionUp": graph.DirectionUp, "DirectionDown": graph.DirectionDown, "DirectionLeft": graph.DirectionLeft,
		"DirectionRight": graph.DirectionRight, "DirectionPerpendicular": graph.DirectionPerpendicular, "DirectionDiagonal": graph.DirectionDiagonal,
	}
	directionSets := [][]string{
		nil, {"DirectionUp"}, {"DirectionDown"}, {"DirectionLeft"}, {"DirectionRight"}, {"DirectionLeft", "DirectionRight"},
		{"DirectionPerpendicular"}, {"DirectionDiagonal"}, {"DirectionDown", "DirectionDiagonal"}, {"DirectionUp", "DirectionDiagonal"},
		{"DirectionUp", "DirectionDown"},
	}
	for _, name := range []string{"grid", "board"} {
		e := enums[name]
		for _, dirs := range directionSets {
			var fs []graph.EdgeFilter
			for _, d := range dirs {
				fs = append(fs, filters[d])
			}
			g := graph.MustNewEnumGridConnectedness(e, fs...)
			vec := graphVec{Enum: name, Dimensions: e.RangeDimensions(), Directions: dirs, Neighbors: map[string][]int{}}
			if vec.Directions == nil {
				vec.Directions = []string{}
			}
			n := len(e.Values())
			for k := -1; k <= n; k++ {
				keys := g.NeighborsByKey(enum.EnumKey(k))
				var ints []int
				for _, key := range keys {
					ints = append(ints, int(key))
				}
				sort.Ints(ints)
				vec.Neighbors[strconv.Itoa(k)] = ints
			}
			out.Graphs = append(out.Graphs, vec)
		}
	}
	return out
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// goValue is the Go value an arg describes.
func goValue(a arg) interface{} {
	switch a.T {
	case "int":
		var n int
		check(json.Unmarshal(a.V, &n))
		return n
	case "float64":
		var text string
		check(json.Unmarshal(a.V, &text))
		f, err := strconv.ParseFloat(text, 64)
		check(err)
		return f
	case "string":
		var s string
		check(json.Unmarshal(a.V, &s))
		return s
	case "bool":
		var b bool
		check(json.Unmarshal(a.V, &b))
		return b
	case "nil":
		return nil
	case "[]int":
		var xs []int
		check(json.Unmarshal(a.V, &xs))
		return xs
	case "[]string":
		var xs []string
		check(json.Unmarshal(a.V, &xs))
		return xs
	}
	panic("unknown arg type " + a.T)
}
