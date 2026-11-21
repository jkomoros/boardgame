// Command sanitizevec emits ../../sanitize.vectors.json: per-viewer views
// produced by real Go (Game.JSONForPlayer, which is state.SanitizedForPlayer
// plus the viewer JSON encoding) for many states, viewers and policy
// configurations of the vector game. Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/sanitizevec
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
)

type policies struct {
	Game    map[string]map[string]string
	Player  map[string]map[string]string
	Dynamic map[string]map[string]map[string]string
}

type vecState struct {
	GameID     string
	SecretSalt string
	Version    int
	NumPlayers int
	State      json.RawMessage
}

type vecCase struct {
	Label  string
	State  int // index into States
	Viewer int
	// Policies holds each property whose policy differs from its struct
	// tag; every other property uses the tag in the exported definition.
	Policies policies
	// Outcomes holds every distinct result Go produced over repeated runs.
	// Almost always one. More than one means Go's answer depends on map
	// iteration order (see fabric/status/engine-core.md).
	Outcomes []outcome
}

type outcome struct {
	// Err is Go's error text, or the panic value when Panic is set.
	Err   string          `json:",omitempty"`
	Panic bool            `json:",omitempty"`
	View  json.RawMessage `json:",omitempty"`
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var policyByName = map[string]boardgame.Policy{
	"visible": boardgame.PolicyVisible, "order": boardgame.PolicyOrder, "len": boardgame.PolicyLen,
	"nonempty": boardgame.PolicyNonEmpty, "hidden": boardgame.PolicyHidden, "invalid": boardgame.PolicyInvalid,
}
var policyNames = []string{"visible", "order", "len", "nonempty", "hidden"}

// buildState makes a state with components spread over every kind of stack
// and every scalar set to a non-zero value.
func buildState(h *vecgame.Harness, rnd *rand.Rand, numPlayers int) vecState {
	id := h.UniqueID("SAN")
	salt := fmt.Sprintf("SANSALT%03d", rnd.Intn(1000))
	version := 1 + rnd.Intn(50)
	fresh, err := h.FreshBlob(id+"SETUP", salt, numPlayers)
	check(err)
	_, st, err := h.StateAt(id+"BUILD", salt, version, numPlayers, fresh)
	check(err)
	g := vecgame.Game(st)
	players := vecgame.Players(st)

	check(g.Draw.Shuffle())
	deal := func(dst boardgame.Stack, slot int) {
		if g.Draw.Len() == 0 {
			return
		}
		check(g.Draw.ComponentAt(0).MoveTo(dst, slot))
	}
	deal(g.Limited, 0)
	deal(g.Row, 1)
	deal(g.Row, 3)
	deal(g.Under, 1)
	deal(g.Under, 2)
	deal(g.TagLen, 0)
	if rnd.Intn(2) == 0 {
		deal(g.TagLen, 0)
	}
	deal(g.TagOrder, 0)
	if rnd.Intn(2) == 0 {
		deal(g.TagNonEmpty, 0)
	}
	for i, p := range players {
		deal(p.Hand, 0)
		if i%2 == 0 {
			check(g.Draw.ComponentAt(0).SecretMoveTo(p.Hand, 0))
			deal(p.Slots, 1)
		}
		p.Score = 10 + i
		p.Secret = 100 + i
		p.TeamSecret = fmt.Sprintf("team-%d", i)
		p.RivalNote = fmt.Sprintf("rival-%d", i)
		p.OwnOnly = []int{i, i + 1}
		p.Dyn = i
		p.Color.SetValue(enumKey((i + rnd.Intn(2)) % 3))
	}
	// Card dynamic values: a nested chain (one card holding another) and a
	// held token, so transitive visibility has something to follow.
	if g.Draw.Len() >= 2 {
		outer := g.Row.ComponentAt(1)
		check(g.Draw.ComponentAt(0).MoveTo(vecgame.CardDynamic(st, outer.DeckIndex()).Nest, 0))
		inner := vecgame.CardDynamic(st, outer.DeckIndex()).Nest.ComponentAt(0)
		check(g.Draw.ComponentAt(0).MoveTo(vecgame.CardDynamic(st, inner.DeckIndex()).Nest, 0))
		check(g.Tokens.ComponentAt(0).MoveTo(vecgame.CardDynamic(st, inner.DeckIndex()).Held, 0))
	}
	check(g.Tokens.ComponentAt(0).MoveTo(vecgame.CardDynamic(st, g.Limited.ComponentAt(0).DeckIndex()).Held, 0))
	check(g.Tokens.ComponentAt(0).MoveTo(g.Spaces.SpaceAt(1), 0))
	check(g.Tokens.ComponentAt(0).MoveTo(g.Spaces.SpaceAt(1), 0))
	check(g.Tokens.ComponentAt(0).SecretMoveTo(g.Spaces.SpaceAt(2), 0))
	for i := 0; i < vecgame.NumCards; i++ {
		d := vecgame.CardDynamic(st, i)
		d.Counter = i + 1
		d.Mark.SetValue(enumKey(i % 3))
	}
	g.Count = 7
	g.Name = "named"
	g.Flag = true
	g.Who = boardgame.PlayerIndex(numPlayers - 1)
	g.Ints = []int{3, 1, 4}
	g.Bools = []bool{true, false}
	g.Strings = []string{"x", "y", "z", "w"}
	g.Whos = []boardgame.PlayerIndex{1, 0}
	g.Color.SetValue(vecgame.ColorBlue)
	check(g.Colors.SetValues([]enumKeyT{vecgame.ColorGreen, vecgame.ColorBlue, vecgame.ColorGreen}))
	g.TagHidden = 41
	g.TagRedOnly = "red eyes only"
	g.TagSpaced = 43
	if rnd.Intn(3) == 0 {
		g.Ints = []int{}
		g.Strings = nil
	}

	blob, err := vecgame.Blob(st)
	check(err)
	return vecState{GameID: id, SecretSalt: salt, Version: version, NumPlayers: numPlayers, State: json.RawMessage(blob)}
}

func main() {
	h, err := vecgame.NewHarness()
	check(err)
	rnd := rand.New(rand.NewSource(77001))

	def, err := vecgame.DefinitionFor(h.Manager)
	check(err)
	tagPolicies := func() policies {
		p := policies{Game: map[string]map[string]string{}, Player: map[string]map[string]string{}, Dynamic: map[string]map[string]map[string]string{}}
		for name, prop := range def.Game.Props {
			p.Game[name] = prop.Sanitize
		}
		for name, prop := range def.Player.Props {
			p.Player[name] = prop.Sanitize
		}
		for deck, sub := range def.DynamicComponentValues {
			p.Dynamic[deck] = map[string]map[string]string{}
			for name, prop := range sub.Props {
				p.Dynamic[deck][name] = prop.Sanitize
			}
		}
		return p
	}
	clone := func(p policies) policies {
		var out policies
		blob, _ := json.Marshal(p)
		check(json.Unmarshal(blob, &out))
		return out
	}
	// sparse keeps only the properties whose policy differs from its struct
	// tag; the engine test overlays these on the exported definition.
	sparse := func(p policies) policies {
		base := tagPolicies()
		out := policies{Game: map[string]map[string]string{}, Player: map[string]map[string]string{}, Dynamic: map[string]map[string]map[string]string{}}
		same := func(a, b map[string]string) bool {
			x, _ := json.Marshal(a)
			y, _ := json.Marshal(b)
			return string(x) == string(y)
		}
		for name, m := range p.Game {
			if !same(m, base.Game[name]) {
				out.Game[name] = m
			}
		}
		for name, m := range p.Player {
			if !same(m, base.Player[name]) {
				out.Player[name] = m
			}
		}
		for deck, props := range p.Dynamic {
			for name, m := range props {
				if !same(m, base.Dynamic[deck][name]) {
					if out.Dynamic[deck] == nil {
						out.Dynamic[deck] = map[string]map[string]string{}
					}
					out.Dynamic[deck][name] = m
				}
			}
		}
		return out
	}
	isMerged := func(sub vecgame.SubStateDef, name string) bool { return sub.Props[name].Type == "mergedStack" }

	var states []vecState
	for i := 0; i < 3; i++ {
		states = append(states, buildState(h, rnd, 2+i%3))
	}

	// Each state is loaded once, under the game id the vector records, and
	// every case sanitizes that same loaded state.
	loaded := make([]boardgame.State, len(states))
	for i, vs := range states {
		_, st, err := h.StateAt(vs.GameID, vs.SecretSalt, vs.Version, vs.NumPlayers, vs.State)
		check(err)
		loaded[i] = st
	}

	var cases []vecCase
	unstable := 0
	run := func(label string, stateIndex int, viewer int, p policies) {
		vs := states[stateIndex]
		h.Delegate.Override = func(ref boardgame.StatePropertyRef) (map[string]boardgame.Policy, bool) {
			var source map[string]string
			switch ref.Group {
			case boardgame.StateGroupGame:
				source = p.Game[ref.PropName]
			case boardgame.StateGroupPlayer:
				source = p.Player[ref.PropName]
			case boardgame.StateGroupDynamicComponentValues:
				source = p.Dynamic[ref.DeckName][ref.PropName]
			}
			if source == nil {
				return nil, false
			}
			out := map[string]boardgame.Policy{}
			for group, name := range source {
				out[group] = policyByName[name]
			}
			return out, true
		}
		once := func() (view []byte, errText string, panicked bool) {
			defer func() {
				if r := recover(); r != nil {
					view, errText, panicked = nil, fmt.Sprint(r), true
				}
			}()
			return viewOf(h, vs, loaded[stateIndex], viewer)
		}
		c := vecCase{Label: label, State: stateIndex, Viewer: viewer, Policies: sparse(p)}
		// Go ranges over maps while sanitizing, so one input can have more
		// than one answer. Repeat, and keep every distinct outcome.
		seen := map[string]bool{}
		for i := 0; i < 60; i++ {
			view, errText, panicked := once()
			key := fmt.Sprint(panicked) + "|" + errText + "|" + string(view)
			if seen[key] {
				continue
			}
			seen[key] = true
			o := outcome{Err: errText, Panic: panicked}
			if view != nil {
				o.View = json.RawMessage(view)
			}
			c.Outcomes = append(c.Outcomes, o)
		}
		sort.Slice(c.Outcomes, func(i, j int) bool {
			if c.Outcomes[i].Panic != c.Outcomes[j].Panic {
				return !c.Outcomes[i].Panic
			}
			return string(c.Outcomes[i].View) < string(c.Outcomes[j].View)
		})
		if len(c.Outcomes) > 1 {
			unstable++
		}
		cases = append(cases, c)
	}

	viewers := func(n int) []int { // any, admin, observer, each player, one past the end
		out := []int{-3, -2, -1}
		for i := 0; i <= n; i++ {
			out = append(out, i)
		}
		return out
	}

	for si, vs := range states {
		for _, viewer := range viewers(vs.NumPlayers) {
			run("struct tags", si, viewer, tagPolicies())
		}
		for _, policy := range policyNames {
			// Every game property under one policy (merged stacks stay visible).
			p := tagPolicies()
			for name := range p.Game {
				if !isMerged(def.Game, name) {
					p.Game[name] = map[string]string{"all": policy}
				}
			}
			run("game all:"+policy, si, -1, p)
			run("game all:"+policy, si, 0, p)

			// Every player property hidden from others under one policy.
			p = tagPolicies()
			for name := range p.Player {
				p.Player[name] = map[string]string{"other": policy}
			}
			for _, viewer := range []int{-1, 0, vs.NumPlayers - 1} {
				run("player other:"+policy, si, viewer, p)
			}
			p = tagPolicies()
			for name := range p.Player {
				p.Player[name] = map[string]string{"all": "hidden", "self": policy}
			}
			run("player all:hidden,self:"+policy, si, 0, p)
			run("player all:hidden,self:"+policy, si, 1, p)

			// Every dynamic property under one policy.
			p = tagPolicies()
			for name := range p.Dynamic["cards"] {
				p.Dynamic["cards"][name] = map[string]string{"all": policy}
			}
			run("dynamic all:"+policy, si, -1, p)

			// Dynamic values behind stacks of each policy: only components in
			// a visible stack (directly or through a visible component's own
			// stacks) keep their dynamic values.
			p = tagPolicies()
			for _, name := range []string{"Draw", "Limited", "Row", "Under"} {
				p.Game[name] = map[string]string{"all": policy}
			}
			run("card stacks all:"+policy, si, -1, p)
		}
		// Random mixtures over every group name the game knows.
		groups := []string{"all", "other", "self", "same-color", "different-color", "Red", "Green", "Blue"}
		for k := 0; k < 8; k++ {
			p := tagPolicies()
			randomMap := func() map[string]string {
				out := map[string]string{}
				n := 1 + rnd.Intn(3)
				for i := 0; i < n; i++ {
					out[groups[rnd.Intn(len(groups))]] = policyNames[rnd.Intn(len(policyNames))]
				}
				return out
			}
			names := func(m map[string]map[string]string) []string {
				var out []string
				for name := range m {
					out = append(out, name)
				}
				sort.Strings(out)
				return out
			}
			for _, name := range names(p.Game) {
				if !isMerged(def.Game, name) && rnd.Intn(2) == 0 {
					p.Game[name] = randomMap()
				}
			}
			for _, name := range names(p.Player) {
				if rnd.Intn(2) == 0 {
					p.Player[name] = randomMap()
				}
			}
			for _, name := range names(p.Dynamic["cards"]) {
				if rnd.Intn(2) == 0 {
					p.Dynamic["cards"][name] = randomMap()
				}
			}
			viewer := rnd.Intn(vs.NumPlayers+1) - 1
			run(fmt.Sprintf("random %d", k), si, viewer, clone(p))
		}
	}

	// Edge cases on the first state.
	{
		p := tagPolicies()
		p.Game["All"] = map[string]string{"all": "len"}
		run("merged concatenation with a non-visible policy", 0, -1, p)
		p = tagPolicies()
		p.Game["Overlap"] = map[string]string{"all": "hidden"}
		run("merged overlap with a non-visible policy", 0, -1, p)
		p = tagPolicies()
		p.Game["Count"] = map[string]string{"all": "invalid"}
		run("an invalid policy", 0, -1, p)
		p = tagPolicies()
		p.Game["Count"] = map[string]string{"all": "invalid", "nobody": "hidden"}
		run("an invalid policy beside an inapplicable one", 0, 0, p)
		p = tagPolicies()
		p.Game["Count"] = map[string]string{"nobody": "hidden"}
		run("no applicable group falls back to visible", 0, 0, p)
	}

	fmt.Fprintln(os.Stderr, "cases:", len(cases), "unstable:", unstable)
	for _, c := range cases {
		if len(c.Outcomes) > 1 || c.Outcomes[0].Err != "" {
			var errs []string
			for _, o := range c.Outcomes {
				errs = append(errs, fmt.Sprintf("panic=%v err=%q", o.Panic, o.Err))
			}
			fmt.Fprintln(os.Stderr, " ", c.Label, "viewer", c.Viewer, "outcomes:", len(c.Outcomes), errs)
		}
	}
	blob, err := json.Marshal(struct {
		States []vecState
		Cases  []vecCase
	}{states, cases})
	check(err)
	check(os.WriteFile("server/static/src/engine/testdata/sanitize.vectors.json", append(blob, '\n'), 0644))
}
