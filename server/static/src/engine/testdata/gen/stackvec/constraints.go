package main

// The constraint vectors: ../../constraints.vectors.json. Scripted stack
// operations on live states of three games whose stacks carry struct-tag
// constraints: cgame (every constraint kind, ./cgame), tictactoe
// (`same(Value)`) and checkers (`same(Color)`). Each scenario records the
// state before, each operation with Go's error text (or result), the stacks
// that changed, and the state after, exactly as stacks.vectors.json does,
// plus the by-key operations and the May... forms. Generation continues,
// scenario by scenario, until every operation has both accepted and refused a
// component on a destination of every constraint kind (and, for the
// transfers, refused a transfer whose components each pass alone), so the
// coverage is a property of the file, not of luck; the number of scenarios is
// bounded and the generator fails if the bound is reached first.
//
// It also records Go's error text for every malformed constraint tag in
// cgame.BadStates, through boardgame.NewStructInflater.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/constraints"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/stackvec/cgame"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

type cop struct {
	Op          string
	Src         *vecgame.StackRef `json:",omitempty"`
	Dst         *vecgame.StackRef `json:",omitempty"`
	Index       int               `json:"Index"`
	Slot        int               `json:"Slot"`
	NoComponent bool              `json:",omitempty"`
	Err         string
	// Result is componentAtKey's deck index, or null for Go's nil.
	Result *int `json:",omitempty"`
	// Constraint names the constraint kind of the destination ("" for none).
	Constraint string `json:",omitempty"`
	// Count is how many components a transfer moves.
	Count int `json:",omitempty"`
	// OrderDependent marks a refused transfer of more than one component
	// whose components would each have been accepted alone.
	OrderDependent  bool                       `json:",omitempty"`
	Changed         map[string]json.RawMessage `json:",omitempty"`
	SecretMoveCount map[string][]int           `json:",omitempty"`
}

type cscenario struct {
	Game       string
	GameID     string
	SecretSalt string
	Version    int
	NumPlayers int
	Before     json.RawMessage
	Ops        []cop
	After      json.RawMessage
}

type tagError struct {
	Tag string
	Err string
}

type cgameHarness struct {
	name     string
	delegate boardgame.GameDelegate
	manager  *boardgame.GameManager
	storage  *memstore.StorageManager
	def      *definition
}

func newCHarness(delegate boardgame.GameDelegate) *cgameHarness {
	storage := memstore.NewStorageManager()
	manager, err := boardgame.NewGameManager(delegate, storage)
	check(err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	def, err := definitionFor(manager)
	check(err)
	return &cgameHarness{name: delegate.Name(), delegate: delegate, manager: manager, storage: storage, def: def}
}

// stateAt sets up a fresh game, stores its current state as the state of a
// game (id, salt) at version, and loads that back the way the server does,
// returning a mutable copy with the real id, salt and version.
func (h *cgameHarness) stateAt(id, salt string, version, numPlayers int) boardgame.State {
	fresh, err := h.manager.Internals().RecreateGame(&boardgame.GameStorageRecord{
		Name: h.name, ID: id + "SETUP", SecretSalt: salt, NumPlayers: numPlayers,
	})
	check(err)
	blob, err := vecgame.Blob(fresh.CurrentState())
	check(err)
	rec := &boardgame.GameStorageRecord{Name: h.name, ID: id, SecretSalt: salt, NumPlayers: numPlayers, Version: version}
	check(h.storage.SaveGameAndCurrentState(rec, blob, nil))
	game := h.manager.Game(id)
	if game == nil {
		panic("game did not load")
	}
	copied, err := game.CurrentState().Copy(false)
	check(err)
	return copied.(boardgame.State)
}

// stackRefs lists every physical stack of a state, from the definition:
// game and player stacks and board spaces, and the stacks of the first
// four components' dynamic values.
func (h *cgameHarness) stackRefs(numPlayers int) []vecgame.StackRef {
	var out []vecgame.StackRef
	add := func(base vecgame.StackRef, sub subStateDef) {
		names := make([]string, 0, len(sub.Props))
		for name := range sub.Props {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			prop := sub.Props[name]
			switch prop.Type {
			case "stack", "sizedStack":
				r := base
				r.Prop = name
				out = append(out, r)
			case "board":
				for i := 0; i < prop.Size; i++ {
					space := i
					r := base
					r.Prop = name
					r.Space = &space
					out = append(out, r)
				}
			}
		}
	}
	add(vecgame.StackRef{Group: "game"}, h.def.Game)
	for p := 0; p < numPlayers; p++ {
		add(vecgame.StackRef{Group: "player", Player: p}, h.def.Player)
	}
	decks := make([]string, 0, len(h.def.DynamicComponentValues))
	for deck := range h.def.DynamicComponentValues {
		decks = append(decks, deck)
	}
	sort.Strings(decks)
	for _, deck := range decks {
		for i := 0; i < 4; i++ {
			add(vecgame.StackRef{Group: "dynamic", Deck: deck, Index: i}, h.def.DynamicComponentValues[deck])
		}
	}
	return out
}

// mergedRefs lists the game state's merged stacks.
func (h *cgameHarness) mergedRefs() []vecgame.StackRef {
	var out []vecgame.StackRef
	for name, prop := range h.def.Game.Props {
		if prop.Type == "mergedStack" {
			out = append(out, vecgame.StackRef{Group: "game", Prop: name})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Prop < out[j].Prop })
	return out
}

func (h *cgameHarness) propFor(ref vecgame.StackRef) propDef {
	switch ref.Group {
	case "game":
		return h.def.Game.Props[ref.Prop]
	case "player":
		return h.def.Player.Props[ref.Prop]
	}
	return h.def.DynamicComponentValues[ref.Deck].Props[ref.Prop]
}

func subStateOf(st boardgame.State, ref vecgame.StackRef) boardgame.SubState {
	switch ref.Group {
	case "game":
		return st.GameState()
	case "player":
		return st.PlayerStates()[ref.Player]
	}
	return st.DynamicComponentValues()[ref.Deck][ref.Index]
}

func resolveStack(st boardgame.State, ref vecgame.StackRef) boardgame.Stack {
	rs := subStateOf(st, ref).ReadSetter()
	if ref.Space != nil {
		board, err := rs.BoardProp(ref.Prop)
		check(err)
		return board.SpaceAt(*ref.Space)
	}
	stack, err := rs.StackProp(ref.Prop)
	check(err)
	return stack
}

func resolveView(st boardgame.State, ref vecgame.StackRef) boardgame.ImmutableStack {
	view, err := subStateOf(st, ref).Reader().ImmutableStackProp(ref.Prop)
	check(err)
	return view
}

// allStacks extracts every stack's JSON (physical, merged and board
// spaces) from a state blob, keyed by owner path.
func allStacks(blob []byte) (map[string]json.RawMessage, map[string][]int) {
	var parsed struct {
		Game            map[string]json.RawMessage
		Players         []map[string]json.RawMessage
		Components      map[string][]map[string]json.RawMessage
		SecretMoveCount map[string][]int
	}
	check(json.Unmarshal(blob, &parsed))
	out := map[string]json.RawMessage{}
	walk := func(prefix string, sub map[string]json.RawMessage) {
		for name, raw := range sub {
			var probe map[string]json.RawMessage
			if json.Unmarshal(raw, &probe) != nil || probe == nil {
				continue
			}
			if _, ok := probe["Indexes"]; ok {
				out[prefix+"."+name] = raw
				continue
			}
			if spaces, ok := probe["Spaces"]; ok {
				var list []json.RawMessage
				check(json.Unmarshal(spaces, &list))
				for i, space := range list {
					out[fmt.Sprintf("%s.%s[%d]", prefix, name, i)] = space
				}
			}
		}
	}
	walk("Game", parsed.Game)
	for i, p := range parsed.Players {
		walk(fmt.Sprintf("Players[%d]", i), p)
	}
	for deck, values := range parsed.Components {
		for i, v := range values {
			walk(fmt.Sprintf("DynamicComponentValues[%s][%d]", deck, i), v)
		}
	}
	return out, parsed.SecretMoveCount
}

// constraintKind is the kind of a destination's constraints: the single
// constructor name, "combo" for several, "" for none.
func constraintKind(prop propDef) string {
	switch len(prop.Constraints) {
	case 0:
		return ""
	case 1:
		return prop.Constraints[0][:strings.Index(prop.Constraints[0], "(")]
	}
	return "combo"
}

func isConstraintError(text string) bool {
	for _, marker := range []string{"which exceeds max of", "has mixed values", "duplicate value", "distinct values, which exceeds"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

var moveOps = []string{"move", "secretMove", "moveFirst", "moveLast", "moveNext", "mayMoveTo", "mayMoveToSlot"}
var transferOps = []string{"moveAll", "moveCount", "mayMoveAllTo", "mayMoveCountTo"}
var constraintKinds = []string{"max", "same", "unique", "maxdistinct"}

// requiredCells is the coverage the cgame scenarios must reach.
func requiredCells() []string {
	var cells []string
	for _, op := range append(append([]string{}, moveOps...), transferOps...) {
		for _, kind := range constraintKinds {
			cells = append(cells, op+"|"+kind+"|accept", op+"|"+kind+"|refuse")
		}
	}
	for _, op := range transferOps {
		for _, kind := range constraintKinds {
			cells = append(cells, op+"|"+kind+"|order", op+"|"+kind+"|acceptMany")
		}
	}
	for _, op := range []string{"swapByKey", "maySwapByKey"} {
		cells = append(cells, op+"|accept", op+"|refuse")
	}
	cells = append(cells, "componentAtKey|component", "componentAtKey|nil", "componentAtKey|merged",
		"swap|constrained", "slideFirst|constrained", "shuffle|constrained", "move|combo|refuse")
	return cells
}

func cellsOf(o cop) []string {
	var cells []string
	switch o.Op {
	case "swapByKey", "maySwapByKey":
		if o.Err == "" {
			cells = append(cells, o.Op+"|accept")
		} else {
			cells = append(cells, o.Op+"|refuse")
		}
	case "componentAtKey":
		if o.Result != nil {
			cells = append(cells, "componentAtKey|component")
		} else {
			cells = append(cells, "componentAtKey|nil")
		}
		if o.Src != nil && o.Src.Prop == "Both" && o.Result != nil {
			cells = append(cells, "componentAtKey|merged")
		}
	case "swap", "slideFirst", "shuffle":
		if o.Constraint != "" && o.Err == "" && !o.NoComponent {
			cells = append(cells, o.Op+"|constrained")
		}
	}
	if o.Constraint == "" || o.NoComponent {
		return cells
	}
	if o.Err == "" {
		cells = append(cells, o.Op+"|"+o.Constraint+"|accept")
		if (o.Op == "moveAll" || o.Op == "moveCount" || o.Op == "mayMoveAllTo" || o.Op == "mayMoveCountTo") && o.Count > 1 {
			cells = append(cells, o.Op+"|"+o.Constraint+"|acceptMany")
		}
	} else if isConstraintError(o.Err) {
		cells = append(cells, o.Op+"|"+o.Constraint+"|refuse")
		if o.OrderDependent {
			cells = append(cells, o.Op+"|"+o.Constraint+"|order")
		}
	}
	return cells
}

// transferCount is how many components a transfer moves.
func transferCount(o cop, src boardgame.Stack) int {
	if o.Op == "moveAll" || o.Op == "mayMoveAllTo" {
		return src.NumComponents()
	}
	return o.Index
}

// orderDependent reports whether each of the first count components of src
// would be accepted by dst alone.
func orderDependent(src, dst boardgame.Stack, count int) bool {
	if count < 2 {
		return false
	}
	seen := 0
	for _, c := range src.Components() {
		if c == nil {
			continue
		}
		if seen == count {
			break
		}
		if c.MayMoveTo(dst) != nil {
			return false
		}
		seen++
	}
	return seen == count
}

func (h *cgameHarness) scenario(rnd *rand.Rand, id, salt string, version, numPlayers, numOps int, kinds []string) cscenario {
	st := h.stateAt(id, salt, version, numPlayers)
	before, err := vecgame.Blob(st)
	check(err)
	sc := cscenario{Game: h.name, GameID: id, SecretSalt: salt, Version: version, NumPlayers: numPlayers, Before: json.RawMessage(before)}
	prevStacks, prevCounts := allStacks(before)
	refs := h.stackRefs(numPlayers)
	merged := h.mergedRefs()

	pick := func() *vecgame.StackRef {
		r := refs[rnd.Intn(len(refs))]
		return &r
	}
	pickFrom := func() *vecgame.StackRef {
		for tries := 0; tries < 10; tries++ {
			r := pick()
			if resolveStack(st, *r).NumComponents() > 0 {
				return r
			}
		}
		return pick()
	}
	// pickTo prefers a constrained destination of the source's deck.
	pickTo := func(src *vecgame.StackRef) *vecgame.StackRef {
		s := resolveStack(st, *src)
		for tries := 0; tries < 12; tries++ {
			r := pick()
			d := resolveStack(st, *r)
			if d.Deck() != s.Deck() {
				continue
			}
			if len(h.propFor(*r).Constraints) > 0 || rnd.Intn(5) == 0 {
				return r
			}
		}
		return pick()
	}
	componentIndex := func(s boardgame.Stack) int {
		if rnd.Intn(16) == 0 {
			return rnd.Intn(s.Len()+3) - 1
		}
		var filled []int
		for i, c := range s.Components() {
			if c != nil {
				filled = append(filled, i)
			}
		}
		if len(filled) == 0 {
			return 0
		}
		return filled[rnd.Intn(len(filled))]
	}
	slotIndex := func(d boardgame.Stack) int {
		if rnd.Intn(10) == 0 {
			return rnd.Intn(d.Len()+3) - 1
		}
		if d.SizedStack() != nil {
			var empty []int
			for i, c := range d.Components() {
				if c == nil {
					empty = append(empty, i)
				}
			}
			if len(empty) == 0 {
				return 0
			}
			return empty[rnd.Intn(len(empty))]
		}
		return rnd.Intn(d.Len() + 1)
	}

	for i := 0; i < numOps; i++ {
		o := cop{Op: kinds[rnd.Intn(len(kinds))]}
		var opErr error
		switch o.Op {
		case "move", "secretMove", "moveFirst", "moveLast", "moveNext", "mayMoveTo", "mayMoveToSlot":
			o.Src = pickFrom()
			o.Dst = pickTo(o.Src)
			o.Constraint = constraintKind(h.propFor(*o.Dst))
			src := resolveStack(st, *o.Src)
			dst := resolveStack(st, *o.Dst)
			o.Index = componentIndex(src)
			o.Slot = slotIndex(dst)
			c := src.ComponentAt(o.Index)
			if c == nil {
				o.NoComponent = true
				break
			}
			switch o.Op {
			case "move":
				opErr = c.MoveTo(dst, o.Slot)
			case "secretMove":
				opErr = c.SecretMoveTo(dst, o.Slot)
			case "moveFirst":
				opErr = c.MoveToFirstSlot(dst)
			case "moveLast":
				opErr = c.MoveToLastSlot(dst)
			case "moveNext":
				opErr = c.MoveToNextSlot(dst)
			case "mayMoveTo":
				opErr = c.MayMoveTo(dst)
			case "mayMoveToSlot":
				opErr = c.MayMoveToSlot(dst, o.Slot)
			}
		case "slideFirst", "slideLast":
			o.Src = pickFrom()
			o.Constraint = constraintKind(h.propFor(*o.Src))
			src := resolveStack(st, *o.Src)
			o.Index = componentIndex(src)
			c := src.ComponentAt(o.Index)
			if c == nil {
				o.NoComponent = true
				break
			}
			if o.Op == "slideFirst" {
				opErr = c.SlideToFirstSlot()
			} else {
				opErr = c.SlideToLastSlot()
			}
		case "moveAll", "moveCount", "mayMoveAllTo", "mayMoveCountTo":
			o.Src = pickFrom()
			o.Dst = pickTo(o.Src)
			o.Constraint = constraintKind(h.propFor(*o.Dst))
			src := resolveStack(st, *o.Src)
			dst := resolveStack(st, *o.Dst)
			if o.Op == "moveCount" || o.Op == "mayMoveCountTo" {
				max := src.NumComponents()
				if max > 4 {
					max = 4
				}
				o.Index = rnd.Intn(max+2) - 1
			}
			count := transferCount(o, src)
			o.Count = count
			switch o.Op {
			case "moveAll":
				opErr = src.MoveAllTo(dst)
			case "moveCount":
				opErr = src.MoveCountTo(dst, o.Index)
			case "mayMoveAllTo":
				opErr = src.MayMoveAllTo(dst)
			case "mayMoveCountTo":
				opErr = src.MayMoveCountTo(dst, o.Index)
			}
			if opErr != nil && isConstraintError(opErr.Error()) {
				o.OrderDependent = orderDependent(src, dst, count)
			}
		case "shuffle":
			o.Src = pickFrom()
			o.Constraint = constraintKind(h.propFor(*o.Src))
			opErr = resolveStack(st, *o.Src).Shuffle()
		case "swap", "swapByKey", "maySwapByKey":
			o.Src = pickFrom()
			o.Constraint = constraintKind(h.propFor(*o.Src))
			src := resolveStack(st, *o.Src)
			o.Index = rnd.Intn(src.Len()+2) - 1
			o.Slot = rnd.Intn(src.Len()+2) - 1
			switch o.Op {
			case "swap":
				opErr = src.SwapComponents(o.Index, o.Slot)
			case "swapByKey":
				opErr = src.SwapComponentsByKey(enumKey(o.Index), enumKey(o.Slot))
			case "maySwapByKey":
				opErr = src.MaySwapComponentsByKey(enumKey(o.Index), enumKey(o.Slot))
			}
		case "componentAtKey":
			if len(merged) > 0 && rnd.Intn(3) == 0 {
				r := merged[rnd.Intn(len(merged))]
				o.Src = &r
				view := resolveView(st, r)
				o.Index = rnd.Intn(view.Len()+3) - 1
				if c := view.ImmutableComponentAtKey(enumKey(o.Index)); c != nil {
					idx := c.DeckIndex()
					o.Result = &idx
				}
			} else {
				o.Src = pickFrom()
				src := resolveStack(st, *o.Src)
				o.Index = rnd.Intn(src.Len()+3) - 1
				c := src.ComponentAtKey(enumKey(o.Index))
				if c != nil {
					idx := c.DeckIndex()
					o.Result = &idx
				}
				if (c == nil) != (src.ImmutableComponentAtKey(enumKey(o.Index)) == nil) {
					panic("ComponentAtKey and ImmutableComponentAtKey disagree")
				}
			}
		}
		if opErr != nil {
			o.Err = opErr.Error()
		}

		blob, err := vecgame.Blob(st)
		check(err)
		stacks, counts := allStacks(blob)
		for path, raw := range stacks {
			if string(prevStacks[path]) != string(raw) {
				if o.Changed == nil {
					o.Changed = map[string]json.RawMessage{}
				}
				o.Changed[path] = raw
			}
		}
		if !reflect.DeepEqual(counts, prevCounts) {
			o.SecretMoveCount = counts
		}
		if (o.Err != "" || o.NoComponent || strings.HasPrefix(o.Op, "may") || o.Op == "componentAtKey") && (o.Changed != nil || o.SecretMoveCount != nil) {
			panic("an operation that must not change the state did: " + o.Op + ": " + o.Err)
		}
		prevStacks, prevCounts = stacks, counts
		sc.Ops = append(sc.Ops, o)
	}
	after, err := vecgame.Blob(st)
	check(err)
	sc.After = json.RawMessage(after)
	return sc
}

func enumKey(i int) enum.EnumKey { return enum.EnumKey(i) }

// tagErrors records Go's error for each malformed constraint tag.
func tagErrors(h *cgameHarness) []tagError {
	ccMap := map[string]*boardgame.StackConstraintConstructor{}
	for _, cc := range constraints.DefaultConstructors() {
		ccMap[cc.Name] = cc
	}
	var out []tagError
	for _, bad := range cgame.BadStates() {
		field, ok := reflect.TypeOf(bad).Elem().FieldByName("S")
		if !ok {
			panic("bad state has no S")
		}
		_, err := boardgame.NewStructInflater(bad, nil, h.manager.Chest(), ccMap)
		if err == nil {
			panic(errors.New("tag was accepted: " + field.Tag.Get("stack")))
		}
		out = append(out, tagError{Tag: field.Tag.Get("stack"), Err: err.Error()})
	}
	return out
}

func constraintVectors() {
	rnd := rand.New(rand.NewSource(20261122))
	cg := newCHarness(&cgame.Delegate{})
	ttt := newCHarness(tictactoe.NewDelegate())
	ck := newCHarness(checkers.NewDelegate())

	all := []string{
		"move", "move", "move", "secretMove", "moveFirst", "moveLast", "moveNext", "mayMoveTo", "mayMoveTo", "mayMoveToSlot",
		"moveAll", "moveCount", "moveCount", "mayMoveAllTo", "mayMoveCountTo", "mayMoveCountTo",
		"slideFirst", "slideLast", "shuffle", "swap", "swapByKey", "maySwapByKey", "componentAtKey",
	}

	need := map[string]bool{}
	for _, cell := range requiredCells() {
		need[cell] = true
	}
	var scenarios []cscenario
	const maxScenarios = 200
	n := 0
	for ; len(need) > 0; n++ {
		if n == maxScenarios {
			var missing []string
			for cell := range need {
				missing = append(missing, cell)
			}
			sort.Strings(missing)
			panic(fmt.Sprintf("coverage not reached in %d scenarios; missing %v", maxScenarios, missing))
		}
		numPlayers := 1 + n%3
		version := rnd.Intn(300)
		if n < 2 {
			version = n
		}
		sc := cg.scenario(rnd, cg.manager.Delegate().Name()+fmt.Sprintf("%04d", n), fmt.Sprintf("SALT%04d", rnd.Intn(10000)), version, numPlayers, 80, all)
		useful := false
		for _, o := range sc.Ops {
			for _, cell := range cellsOf(o) {
				if need[cell] {
					delete(need, cell)
					useful = true
				}
			}
		}
		if useful || n < 4 {
			scenarios = append(scenarios, sc)
		}
	}
	cgameScenarios := len(scenarios)

	for _, h := range []*cgameHarness{ttt, ck} {
		for i := 0; i < 4; i++ {
			sc := h.scenario(rnd, h.name+fmt.Sprintf("%04d", i), fmt.Sprintf("SALT%04d", rnd.Intn(10000)), rnd.Intn(300), 2, 40, all)
			scenarios = append(scenarios, sc)
		}
	}

	out := struct {
		Games     map[string]*definition
		Scenarios []cscenario
		TagErrors []tagError
	}{
		Games:     map[string]*definition{cg.name: cg.def, ttt.name: ttt.def, ck.name: ck.def},
		Scenarios: scenarios,
		TagErrors: tagErrors(cg),
	}
	fmt.Fprintln(os.Stderr, "constraint scenarios:", cgameScenarios, "cgame of", n, "generated;", len(scenarios)-cgameScenarios, "tictactoe/checkers")
	blob, err := json.Marshal(out)
	check(err)
	check(os.WriteFile("server/static/src/engine/testdata/constraints.vectors.json", append(blob, '\n'), 0644))
}
