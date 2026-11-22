package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
)

// The stack half of the facade's known answers: sequences of component and
// stack calls, in Go's own shapes (a ComponentInstance, an error result), on
// real vecgame states. Each operation records what Go returned (an error
// text or nil, a component, a stack and slot), the text of any panic, and
// every top-level property of the state blob that changed. A component is
// named either fresh (Src's ComponentAt(Index)) or from an earlier
// Components() snapshot, which is how a stale handle is exercised.

type compJSON struct {
	Deck  string
	Index int
}

type sop struct {
	Op  string
	Src *vecgame.StackRef `json:",omitempty"`
	Dst *vecgame.StackRef `json:",omitempty"`
	// The component: Snap and Pos name entry Pos of snapshot Snap; with no
	// Snap it is Src's ComponentAt(Index).
	Snap  *int `json:",omitempty"`
	Pos   int  `json:",omitempty"`
	Index int  `json:",omitempty"`
	Slot  int  `json:",omitempty"`
	N     int  `json:",omitempty"`
	// Enum operations: the owner ("game", "player", "dynamic") and key.
	Owner  string `json:",omitempty"`
	Player int    `json:",omitempty"`
	Key    int    `json:",omitempty"`
	// Result is what Go returned, as JSON: an error text or null, a
	// component or null, a list, a bool, an enum key.
	Result json.RawMessage
	// Panic is the text of the panic Go raised, if it did (then Result is null).
	Panic string `json:",omitempty"`
	// Changed holds every top-level property of the state blob that
	// changed ("Game.Draw", "Players[1].Color", "Components.cards[3].Counter",
	// "SecretMoveCount").
	Changed map[string]json.RawMessage `json:",omitempty"`
}

type sscenario struct {
	GameID     string
	SecretSalt string
	Version    int
	NumPlayers int
	Before     json.RawMessage
	Ops        []sop
	After      json.RawMessage
}

// flatten splits a state blob into its top-level properties.
func flatten(blob []byte) map[string]string {
	var parsed struct {
		Game            map[string]json.RawMessage
		Players         []map[string]json.RawMessage
		Components      map[string][]map[string]json.RawMessage
		SecretMoveCount json.RawMessage
	}
	check(json.Unmarshal(blob, &parsed))
	out := map[string]string{}
	for k, v := range parsed.Game {
		out["Game."+k] = string(v)
	}
	for i, p := range parsed.Players {
		for k, v := range p {
			out[fmt.Sprintf("Players[%d].%s", i, k)] = string(v)
		}
	}
	for deck, values := range parsed.Components {
		for i, p := range values {
			for k, v := range p {
				out[fmt.Sprintf("Components.%s[%d].%s", deck, i, k)] = string(v)
			}
		}
	}
	if len(parsed.SecretMoveCount) > 0 {
		out["SecretMoveCount"] = string(parsed.SecretMoveCount)
	}
	return out
}

// immutable resolves a ref, merged stacks included.
func immutable(st boardgame.State, ref vecgame.StackRef) boardgame.ImmutableStack {
	if ref.Group == "game" && ref.Prop == "All" {
		return vecgame.Game(st).All
	}
	if ref.Group == "game" && ref.Prop == "Overlap" {
		return vecgame.Game(st).Overlap
	}
	s, err := vecgame.Resolve(st, ref)
	check(err)
	return s
}

func comp(c boardgame.ImmutableComponentInstance) interface{} {
	if c == nil {
		return nil
	}
	return compJSON{Deck: c.Deck().Name(), Index: c.DeckIndex()}
}

func errResult(err error) interface{} {
	if err == nil {
		return nil
	}
	return err.Error()
}

// valuesOf reads a component's static values through its reader, as
// c.Values().(*T) would.
func valuesOf(c boardgame.ImmutableComponentInstance) map[string]interface{} {
	r := c.Values().Reader()
	out := map[string]interface{}{}
	for name := range r.Props() {
		v, err := r.Prop(name)
		check(err)
		out[name] = v
	}
	return out
}

func rankOf(c boardgame.ImmutableComponentInstance) int {
	if c.Deck().Name() != "cards" {
		return 0
	}
	r, err := c.Values().Reader().IntProp("Rank")
	check(err)
	return r
}

type stackGen struct {
	rnd        *rand.Rand
	st         boardgame.State
	numPlayers int
	refs       []vecgame.StackRef // physical
	merged     []vecgame.StackRef
	snaps      [][]boardgame.ImmutableComponentInstance
	mutable    []bool // whether each snapshot came from a physical stack
}

func intPtr(i int) *int { return &i }

func (g *stackGen) pick() *vecgame.StackRef {
	r := g.refs[g.rnd.Intn(len(g.refs))]
	return &r
}

func (g *stackGen) pickAny() *vecgame.StackRef {
	if g.rnd.Intn(5) == 0 {
		r := g.merged[g.rnd.Intn(len(g.merged))]
		return &r
	}
	return g.pick()
}

func (g *stackGen) stack(ref *vecgame.StackRef) boardgame.Stack {
	s, err := vecgame.Resolve(g.st, *ref)
	check(err)
	return s
}

// pickFrom prefers a stack with components.
func (g *stackGen) pickFrom() *vecgame.StackRef {
	for tries := 0; tries < 8; tries++ {
		r := g.pick()
		if g.stack(r).NumComponents() > 0 {
			return r
		}
	}
	return g.pick()
}

// pickTo prefers a destination with src's deck.
func (g *stackGen) pickTo(deck string) *vecgame.StackRef {
	for tries := 0; tries < 8; tries++ {
		r := g.pick()
		if g.stack(r).Deck().Name() == deck && g.rnd.Intn(10) != 0 {
			return r
		}
	}
	return g.pick()
}

// handle chooses a component for op: from a snapshot (mutable ones only when
// mustMutate), or fresh from a stack. It returns nil for an empty slot.
func (g *stackGen) handle(o *sop, mustMutate bool) boardgame.ImmutableComponentInstance {
	var candidates []int
	for i, snap := range g.snaps {
		if len(snap) > 0 && (!mustMutate || g.mutable[i]) {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) > 0 && g.rnd.Intn(3) == 0 {
		k := candidates[g.rnd.Intn(len(candidates))]
		o.Snap = intPtr(k)
		o.Pos = g.rnd.Intn(len(g.snaps[k]))
		// Mostly a real component; sometimes whatever the slot held.
		for tries := 0; tries < 6 && g.snaps[k][o.Pos] == nil && g.rnd.Intn(5) != 0; tries++ {
			o.Pos = g.rnd.Intn(len(g.snaps[k]))
		}
		return g.snaps[k][o.Pos]
	}
	if mustMutate || g.rnd.Intn(4) != 0 {
		o.Src = g.pickFrom()
	} else {
		o.Src = g.pickAny()
	}
	s := immutable(g.st, *o.Src)
	if g.rnd.Intn(12) == 0 {
		o.Index = g.rnd.Intn(s.Len()+3) - 1
	} else {
		var filled []int
		for i, c := range s.ImmutableComponents() {
			if c != nil {
				filled = append(filled, i)
			}
		}
		if len(filled) > 0 {
			o.Index = filled[g.rnd.Intn(len(filled))]
		}
	}
	return s.ImmutableComponentAt(o.Index)
}

func (g *stackGen) slotFor(d boardgame.ImmutableStack) int {
	if g.rnd.Intn(8) == 0 {
		return g.rnd.Intn(d.Len()+3) - 1
	}
	if d.ImmutableSizedStack() != nil {
		var empty []int
		for i, c := range d.ImmutableComponents() {
			if c == nil {
				empty = append(empty, i)
			}
		}
		if len(empty) > 0 {
			return empty[g.rnd.Intn(len(empty))]
		}
		return 0
	}
	return g.rnd.Intn(d.Len() + 1)
}

func mut(c boardgame.ImmutableComponentInstance) boardgame.ComponentInstance {
	if c == nil {
		return nil
	}
	return c.(boardgame.ComponentInstance)
}

// run performs one operation and returns its result.
func (g *stackGen) run(o *sop) interface{} {
	st := g.st
	switch o.Op {
	case "move", "secretMove", "moveFirst", "moveLast", "moveNext":
		c := mut(g.handle(o, true))
		deck := "cards"
		if c != nil {
			deck = c.Deck().Name()
		}
		o.Dst = g.pickTo(deck)
		dst := g.stack(o.Dst)
		o.Slot = g.slotFor(dst)
		switch o.Op {
		case "move":
			return errResult(c.MoveTo(dst, o.Slot))
		case "secretMove":
			return errResult(c.SecretMoveTo(dst, o.Slot))
		case "moveFirst":
			return errResult(c.MoveToFirstSlot(dst))
		case "moveLast":
			return errResult(c.MoveToLastSlot(dst))
		default:
			return errResult(c.MoveToNextSlot(dst))
		}
	case "slideFirst", "slideLast":
		c := mut(g.handle(o, true))
		if o.Op == "slideFirst" {
			return errResult(c.SlideToFirstSlot())
		}
		return errResult(c.SlideToLastSlot())
	case "mayMove", "mayMoveSlot":
		c := g.handle(o, false)
		deck := "cards"
		if c != nil {
			deck = c.Deck().Name()
		}
		o.Dst = g.pickTo(deck)
		if g.rnd.Intn(6) == 0 {
			r := g.merged[g.rnd.Intn(len(g.merged))]
			o.Dst = &r
		}
		dst := immutable(st, *o.Dst)
		if o.Op == "mayMove" {
			return errResult(c.MayMoveTo(dst))
		}
		o.Slot = g.slotFor(dst)
		return errResult(c.MayMoveToSlot(dst, o.Slot))
	case "containing":
		c := g.handle(o, false)
		s, slot, err := c.ContainingImmutableStack()
		if err != nil {
			return map[string]interface{}{"Path": nil, "Slot": slot, "Err": err.Error()}
		}
		return map[string]interface{}{"Path": g.pathOf(s), "Slot": slot, "Err": nil}
	case "values":
		c := g.handle(o, false)
		return valuesOf(c)
	case "dynamic", "dynamicBump", "markGet", "markSet":
		c := g.handle(o, false)
		if c != nil && c.Deck().Name() != "cards" {
			// tokens have no dynamic values; Go's type assertion on the nil
			// SubState has no facade counterpart. Ask a card instead.
			o.Snap = nil
			o.Src = &vecgame.StackRef{Group: "game", Prop: "Draw"}
			o.Index = 0
			c = immutable(st, *o.Src).ImmutableComponentAt(0)
		}
		switch o.Op {
		case "dynamic":
			v, err := c.ImmutableDynamicValues().Reader().IntProp("Counter")
			check(err)
			return v
		case "dynamicBump":
			rs := mut(c).DynamicValues().ReadSetter()
			v, err := rs.IntProp("Counter")
			check(err)
			check(rs.SetIntProp("Counter", v+1))
			return v + 1
		case "markGet":
			v, err := c.ImmutableDynamicValues().Reader().ImmutableEnumProp("Mark")
			check(err)
			return int(v.Value())
		default:
			v, err := mut(c).DynamicValues().ReadSetter().EnumProp("Mark")
			check(err)
			o.Key = g.rnd.Intn(5) - 1
			return errResult(v.SetValue(enum.EnumKey(o.Key)))
		}
	case "colorGet", "colorSet":
		var v enum.Val
		if g.rnd.Intn(2) == 0 {
			o.Owner = "game"
			v = vecgame.Game(st).Color
		} else {
			o.Owner = "player"
			o.Player = g.rnd.Intn(g.numPlayers)
			v = playerColor(st, o.Player)
		}
		if o.Op == "colorGet" {
			return int(v.Value())
		}
		o.Key = g.rnd.Intn(5) - 1
		return errResult(v.SetValue(enum.EnumKey(o.Key)))
	case "colorEquals":
		o.Player = g.rnd.Intn(g.numPlayers)
		return vecgame.Game(st).Color.Equals(playerColor(st, o.Player))
	case "moveAll", "moveCount":
		o.Src = g.pickFrom()
		src := g.stack(o.Src)
		o.Dst = g.pickTo(src.Deck().Name())
		dst := g.stack(o.Dst)
		if o.Op == "moveAll" {
			return errResult(src.MoveAllTo(dst))
		}
		o.N = g.rnd.Intn(src.NumComponents()+3) - 1
		return errResult(src.MoveCountTo(dst, o.N))
	case "mayMoveAll", "mayMoveCount":
		o.Src = g.pickAny()
		src := immutable(st, *o.Src)
		o.Dst = g.pickTo(src.Deck().Name())
		if g.rnd.Intn(6) == 0 {
			r := g.merged[g.rnd.Intn(len(g.merged))]
			o.Dst = &r
		}
		dst := immutable(st, *o.Dst)
		if o.Op == "mayMoveAll" {
			return errResult(src.MayMoveAllTo(dst))
		}
		o.N = g.rnd.Intn(src.NumComponents()+3) - 1
		return errResult(src.MayMoveCountTo(dst, o.N))
	case "shuffle", "publicShuffle":
		o.Src = g.pickFrom()
		if o.Op == "shuffle" {
			return errResult(g.stack(o.Src).Shuffle())
		}
		return errResult(g.stack(o.Src).PublicShuffle())
	case "swap", "swapByKey":
		o.Src = g.pickFrom()
		s := g.stack(o.Src)
		o.Index = g.rnd.Intn(s.Len()+2) - 1
		o.Slot = g.rnd.Intn(s.Len()+2) - 1
		if o.Op == "swap" {
			return errResult(s.SwapComponents(o.Index, o.Slot))
		}
		return errResult(s.SwapComponentsByKey(enum.EnumKey(o.Index), enum.EnumKey(o.Slot)))
	case "sortRank":
		// By Rank, empty slots last: many ties, so sort.Sort's order among
		// equals shows, and less reads values through its components.
		o.Src = g.pickFrom()
		return errResult(g.stack(o.Src).SortComponents(func(a, b boardgame.ImmutableComponentInstance) bool {
			if a == nil {
				return false
			}
			if b == nil {
				return true
			}
			return rankOf(a) < rankOf(b)
		}))
	case "expand", "contract", "setSize", "sizeToFit":
		o.Src = g.pick()
		for o.Src.Prop == "Row" || o.Src.Prop == "Under" {
			o.Src = g.pick()
		}
		s := g.stack(o.Src)
		o.N = g.rnd.Intn(8) - 1
		switch o.Op {
		case "expand":
			return errResult(s.ExpandSize(o.N))
		case "contract":
			return errResult(s.ContractSize(o.N))
		case "setSize":
			return errResult(s.SetSize(o.N))
		default:
			return errResult(s.SizeToFit())
		}
	case "first", "last", "componentAt", "componentAtKey", "components":
		o.Src = g.pickAny()
		s := immutable(st, *o.Src)
		switch o.Op {
		case "first":
			return comp(s.ImmutableFirst())
		case "last":
			return comp(s.ImmutableLast())
		case "componentAt":
			o.Index = g.rnd.Intn(s.Len()+3) - 1
			return comp(s.ImmutableComponentAt(o.Index))
		case "componentAtKey":
			o.Index = g.rnd.Intn(s.Len()+3) - 1
			return comp(s.ImmutableComponentAtKey(enum.EnumKey(o.Index)))
		default:
			snap := s.ImmutableComponents()
			g.snaps = append(g.snaps, snap)
			_, isMerged := s.(boardgame.MergedStack)
			g.mutable = append(g.mutable, !isMerged)
			out := []interface{}{}
			for _, c := range snap {
				out = append(out, comp(c))
			}
			return out
		}
	}
	panic("unknown op " + o.Op)
}

func playerColor(st boardgame.State, i int) enum.Val {
	v, err := st.PlayerStates()[i].ReadSetter().EnumProp("Color")
	check(err)
	return v
}

// pathOf names a stack by its vecgame ref's owner path.
func (g *stackGen) pathOf(s boardgame.ImmutableStack) string {
	for _, ref := range append(append([]vecgame.StackRef{}, g.refs...), vecgame.AllStackRefs(g.numPlayers)...) {
		if immutable(g.st, ref) == s {
			return ref.Path()
		}
	}
	// A dynamic stack of a card the ref list does not name.
	dyn := g.st.DynamicComponentValues()["cards"]
	for i := range dyn {
		for _, prop := range []string{"Held", "Nest"} {
			ref := vecgame.StackRef{Group: "dynamic", Deck: "cards", Index: i, Prop: prop}
			if immutable(g.st, ref) == s {
				return ref.Path()
			}
		}
	}
	panic("stack not found")
}

// step runs one op, recovering a panic, and records what changed.
func (g *stackGen) step(o sop, prev map[string]string, perform func(o *sop) interface{}) (sop, map[string]string) {
	var result interface{}
	o.Panic = recovered(func() { result = perform(&o) })
	o.Result = raw(result)
	blob, err := vecgame.Blob(g.st)
	check(err)
	now := flatten(blob)
	for k, v := range now {
		if prev[k] != v {
			if o.Changed == nil {
				o.Changed = map[string]json.RawMessage{}
			}
			o.Changed[k] = json.RawMessage(v)
		}
	}
	for k := range prev {
		if _, ok := now[k]; !ok {
			panic("a property disappeared: " + k)
		}
	}
	failed := o.Panic != ""
	if s, ok := result.(string); ok && s != "" {
		failed = true
	}
	if failed && o.Changed != nil && o.Op != "dynamicBump" {
		panic("a failed operation changed the state: " + o.Op)
	}
	return o, now
}

func stackVecs() interface{} {
	h, err := vecgame.NewHarness()
	check(err)
	rnd := rand.New(rand.NewSource(20261122))
	var scenarios []sscenario
	counts := map[string]int{}
	panics := 0
	failures := 0

	kinds := []string{
		"move", "move", "move", "secretMove", "moveFirst", "moveLast", "moveNext", "slideFirst", "slideLast",
		"mayMove", "mayMove", "mayMoveSlot", "mayMoveSlot", "containing", "containing", "values", "dynamic", "dynamicBump",
		"markGet", "markSet", "colorGet", "colorSet", "colorEquals", "moveAll", "moveCount", "mayMoveAll", "mayMoveCount",
		"shuffle", "publicShuffle", "swap", "swapByKey", "sortRank", "expand", "contract", "setSize", "sizeToFit",
		"first", "last", "componentAt", "componentAtKey", "components", "components", "components",
	}

	for n := 0; n < 13; n++ {
		numPlayers := 2 + n%3
		id := h.UniqueID("FAC")
		salt := fmt.Sprintf("SALT%04d", rnd.Intn(10000))
		version := rnd.Intn(300)
		fresh, err := h.FreshBlob(id+"SETUP", salt, numPlayers)
		check(err)
		_, st, err := h.StateAt(id, salt, version, numPlayers, fresh)
		check(err)
		before, err := vecgame.Blob(st)
		check(err)
		g := &stackGen{rnd: rnd, st: st, numPlayers: numPlayers, refs: vecgame.AllStackRefs(numPlayers),
			merged: []vecgame.StackRef{{Group: "game", Prop: "All"}, {Group: "game", Prop: "Overlap"}}}
		sc := sscenario{GameID: id, SecretSalt: salt, Version: version, NumPlayers: numPlayers, Before: json.RawMessage(before)}
		prev := flatten(before)

		if n == 0 {
			sc.Ops, prev = g.scripted(prev)
		}
		for i := 0; i < 70; i++ {
			var o sop
			o, prev = g.step(sop{Op: kinds[rnd.Intn(len(kinds))]}, prev, g.run)
			counts[o.Op]++
			if o.Panic != "" {
				panics++
			}
			if s := string(o.Result); len(s) > 2 && s[0] == '"' {
				failures++
			}
			sc.Ops = append(sc.Ops, o)
		}
		after, err := vecgame.Blob(st)
		check(err)
		sc.After = json.RawMessage(after)
		scenarios = append(scenarios, sc)
	}
	var names []string
	for k := range counts {
		names = append(names, k)
	}
	sort.Strings(names)
	fmt.Fprintln(os.Stderr, "facade stack ops:", len(names), "kinds; panics:", panics, "error results:", failures)
	return struct{ Scenarios []sscenario }{scenarios}
}

// scripted is the fixed prefix of scenario 0: the cases a random sequence
// would rarely reach.
func (g *stackGen) scripted(prev map[string]string) ([]sop, map[string]string) {
	var ops []sop
	add := func(o sop, f func(o *sop) interface{}) {
		o, prev = g.step(o, prev, f)
		ops = append(ops, o)
	}
	draw := vecgame.StackRef{Group: "game", Prop: "Draw"}
	hand := vecgame.StackRef{Group: "player", Player: 0, Prop: "Hand"}
	slots := vecgame.StackRef{Group: "player", Player: 0, Prop: "Slots"}
	limited := vecgame.StackRef{Group: "game", Prop: "Limited"}
	all := vecgame.StackRef{Group: "game", Prop: "All"}
	overlap := vecgame.StackRef{Group: "game", Prop: "Overlap"}
	tokens := vecgame.StackRef{Group: "game", Prop: "Tokens"}
	snapOf := func(ref vecgame.StackRef) {
		add(sop{Op: "components", Src: &ref}, func(o *sop) interface{} {
			s := immutable(g.st, ref)
			snap := s.ImmutableComponents()
			g.snaps = append(g.snaps, snap)
			_, isMerged := s.(boardgame.MergedStack)
			g.mutable = append(g.mutable, !isMerged)
			out := []interface{}{}
			for _, c := range snap {
				out = append(out, comp(c))
			}
			return out
		})
	}
	fromSnap := func(op string, snap, pos int, dst *vecgame.StackRef, slot int) {
		add(sop{Op: op, Snap: intPtr(snap), Pos: pos, Dst: dst, Slot: slot}, func(o *sop) interface{} {
			c := g.snaps[snap][pos]
			switch op {
			case "move":
				return errResult(mut(c).MoveTo(g.stack(dst), slot))
			case "mayMove":
				return errResult(c.MayMoveTo(immutable(g.st, *dst)))
			case "containing":
				s, at, err := c.ContainingImmutableStack()
				if err != nil {
					return map[string]interface{}{"Path": nil, "Slot": at, "Err": err.Error()}
				}
				return map[string]interface{}{"Path": g.pathOf(s), "Slot": at, "Err": nil}
			case "values":
				return valuesOf(c)
			}
			panic(op)
		})
	}

	// A snapshot outlives the moves made from it: each handle is its
	// component, wherever that component now is.
	snapOf(draw)                         // snapshot 0
	fromSnap("move", 0, 0, &hand, 0)     // first card to the hand
	fromSnap("containing", 0, 0, nil, 0) // now in the hand
	fromSnap("move", 0, 0, &limited, 0)  // and from the hand on
	fromSnap("containing", 0, 1, nil, 0) // the second card shifted to slot 0
	fromSnap("move", 0, 1, &slots, 1)    // into a sized stack's second slot
	// A sized stack contracted past an empty slot leaves the state's
	// component index stale: Go then refuses to move that component.
	snapOf(slots) // snapshot 1: [nil, card]
	add(sop{Op: "contract", Src: &slots, N: 1}, func(o *sop) interface{} { return errResult(g.stack(&slots).ContractSize(1)) })
	fromSnap("containing", 1, 1, nil, 0)
	fromSnap("move", 1, 1, &draw, 0)
	fromSnap("mayMove", 1, 1, &draw, 0)
	fromSnap("values", 1, 1, nil, 0)
	// A nil handle panics in every method.
	fromSnap("move", 1, 0, &draw, 0)
	fromSnap("mayMove", 1, 0, &draw, 0)
	fromSnap("containing", 1, 0, nil, 0)
	fromSnap("values", 1, 0, nil, 0)
	// Merged stacks are views: never a destination, never a mover.
	snapOf(all) // snapshot 2
	fromSnap("mayMove", 2, 0, &all, 0)
	fromSnap("mayMove", 2, 0, &overlap, 0)
	fromSnap("mayMove", 2, 0, &tokens, 0)
	add(sop{Op: "mayMoveAll", Src: &all, Dst: &hand}, func(o *sop) interface{} {
		return errResult(immutable(g.st, all).MayMoveAllTo(immutable(g.st, hand)))
	})
	add(sop{Op: "mayMoveCount", Src: &all, Dst: &hand, N: 1}, func(o *sop) interface{} {
		return errResult(immutable(g.st, all).MayMoveCountTo(immutable(g.st, hand), 1))
	})
	add(sop{Op: "mayMoveAll", Src: &hand, Dst: &all}, func(o *sop) interface{} {
		return errResult(immutable(g.st, hand).MayMoveAllTo(immutable(g.st, all)))
	})
	add(sop{Op: "mayMoveCount", Src: &hand, Dst: &overlap, N: 0}, func(o *sop) interface{} {
		return errResult(immutable(g.st, hand).MayMoveCountTo(immutable(g.st, overlap), 0))
	})
	return ops, prev
}
