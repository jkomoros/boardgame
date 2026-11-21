// Command stackvec emits ../../stacks.vectors.json: scripted sequences of
// stack operations run against real boardgame states, with the state blob
// before, the stacks that changed after every operation, each error Go
// returned, and the blob after. Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/stackvec
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
)

type op struct {
	Op    string
	Src   *vecgame.StackRef `json:",omitempty"`
	Dst   *vecgame.StackRef `json:",omitempty"`
	Index int               `json:"Index"`
	Slot  int               `json:"Slot"`
	// NoComponent is set when Src has no component at Index. Go's
	// component-centric API cannot even express that call, so no Go error
	// text exists; the engine must simply refuse.
	NoComponent bool `json:",omitempty"`
	// Err is Go's error text, or "" when the operation succeeded.
	Err string
	// Changed holds the JSON of every stack (physical or merged) whose
	// serialization changed, keyed by owner path.
	Changed map[string]json.RawMessage `json:",omitempty"`
	// SecretMoveCount is present when it changed.
	SecretMoveCount map[string][]int `json:",omitempty"`
}

type scenario struct {
	GameID     string
	SecretSalt string
	Version    int
	NumPlayers int
	Before     json.RawMessage
	Ops        []op
	After      json.RawMessage
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// stacksOf extracts every stack's JSON from a state blob by owner path.
func stacksOf(blob []byte, numPlayers int) (map[string]json.RawMessage, map[string][]int) {
	var parsed struct {
		Game            map[string]json.RawMessage
		Players         []map[string]json.RawMessage
		Components      map[string][]map[string]json.RawMessage
		SecretMoveCount map[string][]int
	}
	check(json.Unmarshal(blob, &parsed))
	out := map[string]json.RawMessage{}
	for _, ref := range vecgame.AllStackRefs(numPlayers) {
		var raw json.RawMessage
		switch ref.Group {
		case "game":
			raw = parsed.Game[ref.Prop]
			if ref.Space != nil {
				var board struct{ Spaces []json.RawMessage }
				check(json.Unmarshal(raw, &board))
				raw = board.Spaces[*ref.Space]
			}
		case "player":
			raw = parsed.Players[ref.Player][ref.Prop]
		default:
			raw = parsed.Components[ref.Deck][ref.Index][ref.Prop]
		}
		out[ref.Path()] = raw
	}
	out["Game.All"] = parsed.Game["All"]
	out["Game.Overlap"] = parsed.Game["Overlap"]
	return out, parsed.SecretMoveCount
}

func main() {
	h, err := vecgame.NewHarness()
	check(err)
	rnd := rand.New(rand.NewSource(20261121))

	var scenarios []scenario
	opCounts := map[string]int{}
	errCount := 0

	for n := 0; n < 24; n++ {
		numPlayers := 2 + n%3
		id := h.UniqueID("VEC")
		salt := fmt.Sprintf("SALT%04d", rnd.Intn(10000))
		version := rnd.Intn(400)
		if n < 3 {
			version = n // include versions 0, 1, 2
		}
		fresh, err := h.FreshBlob(id+"SETUP", salt, numPlayers)
		check(err)
		_, st, err := h.StateAt(id, salt, version, numPlayers, fresh)
		check(err)

		before, err := vecgame.Blob(st)
		check(err)
		sc := scenario{GameID: id, SecretSalt: salt, Version: version, NumPlayers: numPlayers, Before: json.RawMessage(before)}
		prevStacks, prevCounts := stacksOf(before, numPlayers)
		refs := vecgame.AllStackRefs(numPlayers)

		pick := func() *vecgame.StackRef {
			r := refs[rnd.Intn(len(refs))]
			return &r
		}
		// pickFrom prefers a stack that has components, so sequences make progress.
		pickFrom := func() *vecgame.StackRef {
			for tries := 0; tries < 8; tries++ {
				r := pick()
				s, err := vecgame.Resolve(st, *r)
				check(err)
				if s.NumComponents() > 0 {
					return r
				}
			}
			return pick()
		}
		// pickTo prefers a destination of the same deck as src.
		pickTo := func(src *vecgame.StackRef) *vecgame.StackRef {
			s, err := vecgame.Resolve(st, *src)
			check(err)
			for tries := 0; tries < 8; tries++ {
				r := pick()
				d, err := vecgame.Resolve(st, *r)
				check(err)
				if d.Deck() == s.Deck() && rnd.Intn(12) != 0 {
					return r
				}
				if rnd.Intn(15) == 0 {
					return r // occasionally the wrong deck, or the same stack
				}
			}
			return pick()
		}
		componentIndex := func(s boardgame.Stack) int {
			if rnd.Intn(14) == 0 {
				return rnd.Intn(s.Len()+3) - 1 // sometimes out of range or empty
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

		kinds := []string{
			"move", "move", "move", "move", "secretMove", "secretMove", "moveFirst", "moveLast", "moveNext",
			"slideFirst", "slideLast", "moveAll", "moveCount", "moveCount", "shuffle", "shuffle", "publicShuffle",
			"swap", "expand", "contract", "setSize", "sizeToFit",
		}

		for i := 0; i < 60; i++ {
			o := op{Op: kinds[rnd.Intn(len(kinds))]}
			var opErr error
			switch o.Op {
			case "move", "secretMove", "moveFirst", "moveLast", "moveNext":
				o.Src = pickFrom()
				o.Dst = pickTo(o.Src)
				src, err := vecgame.Resolve(st, *o.Src)
				check(err)
				dst, err := vecgame.Resolve(st, *o.Dst)
				check(err)
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
				}
			case "slideFirst", "slideLast":
				o.Src = pickFrom()
				src, err := vecgame.Resolve(st, *o.Src)
				check(err)
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
			case "moveAll", "moveCount":
				o.Src = pickFrom()
				o.Dst = pickTo(o.Src)
				src, err := vecgame.Resolve(st, *o.Src)
				check(err)
				dst, err := vecgame.Resolve(st, *o.Dst)
				check(err)
				if o.Op == "moveAll" {
					opErr = src.MoveAllTo(dst)
				} else {
					o.Index = rnd.Intn(src.NumComponents()+3) - 1
					opErr = src.MoveCountTo(dst, o.Index)
				}
			case "shuffle", "publicShuffle":
				o.Src = pickFrom()
				src, err := vecgame.Resolve(st, *o.Src)
				check(err)
				if o.Op == "shuffle" {
					opErr = src.Shuffle()
				} else {
					opErr = src.PublicShuffle()
				}
			case "swap":
				o.Src = pickFrom()
				src, err := vecgame.Resolve(st, *o.Src)
				check(err)
				o.Index = rnd.Intn(src.Len()+2) - 1
				o.Slot = rnd.Intn(src.Len()+2) - 1
				opErr = src.SwapComponents(o.Index, o.Slot)
			case "expand", "contract", "setSize", "sizeToFit":
				o.Src = pick()
				// Row and Under back the Overlap view. Resizing one of them
				// makes the view invalid, and Go then panics serializing the
				// state (mergedStack.IDs indexes past the shorter part,
				// stack.go:1274), so there is no Go answer to record.
				for o.Src.Prop == "Row" || o.Src.Prop == "Under" {
					o.Src = pick()
				}
				src, err := vecgame.Resolve(st, *o.Src)
				check(err)
				o.Index = rnd.Intn(8) - 1
				switch o.Op {
				case "expand":
					opErr = src.ExpandSize(o.Index)
				case "contract":
					opErr = src.ContractSize(o.Index)
				case "setSize":
					opErr = src.SetSize(o.Index)
				case "sizeToFit":
					opErr = src.SizeToFit()
				}
			}
			if opErr != nil {
				o.Err = opErr.Error()
				errCount++
			}
			opCounts[o.Op]++

			blob, err := vecgame.Blob(st)
			check(err)
			stacks, counts := stacksOf(blob, numPlayers)
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
			if (o.Err != "" || o.NoComponent) && (o.Changed != nil || o.SecretMoveCount != nil) {
				panic("a failed operation changed the state: " + o.Op + ": " + o.Err)
			}
			prevStacks, prevCounts = stacks, counts
			sc.Ops = append(sc.Ops, o)
		}
		after, err := vecgame.Blob(st)
		check(err)
		sc.After = json.RawMessage(after)
		scenarios = append(scenarios, sc)
	}

	fmt.Fprintln(os.Stderr, "ops:", opCounts, "errors:", errCount)
	blob, err := json.Marshal(struct{ Scenarios []scenario }{scenarios})
	check(err)
	check(os.WriteFile("server/static/src/engine/testdata/stacks.vectors.json", append(blob, '\n'), 0644))
}
