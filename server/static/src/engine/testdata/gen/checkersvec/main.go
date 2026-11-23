/*
checkersvec writes ../../checkers.vectors.json.gz: what real Go checkers
(examples/checkers) answers on positions constructed directly on a live Go
state. The fuzz corpus plays real games, but only their first 90 versions;
these positions pin each rule directly: steps, captures (blocked, off the
board, over one's own token, refused by a constraint), multi-jumps, crowning,
crowned tokens moving both ways, a finished turn and the end of the game.

A position starts from the state a fresh game has after set-up (the 24
Place Token fix-ups and Start Phase Playing, version 25) and is rearranged
with Go's own stack operations: every token is moved back to UnusedTokens,
then the position's tokens are moved to their spaces, its captured tokens to
players' CapturedTokens; the players' Color, CurrentPlayer, FinishedTurn and
each token's Crowned are set through the generated property readers. The
game's code is never changed and its Legal and Apply are never bypassed:
every answer below comes from calling them.

For each position it records:

  - Move Token's move.Legal text for every (token, space, proposer): token
    is every occupied space and the first two empty ones, space every key,
    proposer each player;
  - the full ledger (GameManager.LegalEvaluateLedger) for the current
    player at every space within two rows and columns of each token;
  - Crown Token's DefaultsForState and its Legal text for the admin at
    every slot -1 .. 64, and Finish Turn's;
  - Apply results, each on a fresh copy: Move Token for the current player
    from every occupied space to every diagonal space one and two away
    (legal or not: Go's Apply runs whatever it is handed), Crown Token at
    every occupied slot and one empty one, Finish Turn, and Place Token
    into one empty space when UnusedTokens has a token. An Apply error is
    its text; a nil error carries the state after;
  - GameEndConditionMet and PlayerScore.

Positions: the real post-set-up position (as Go leaves it: player 0 red,
player 1 black), designed ones for each rule above (plus one with the
players' colors swapped and one mid-jump, where the turn is not over), and
30 random clustered positions on black spaces from a fixed seed. Run from the repository root:

	./scripts/go-local run ./server/static/src/engine/testdata/gen/checkersvec

The output is deterministic (fixed game id, salt and seed; sorted
enumeration; Go map order never reaches it).
*/
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/examples/checkers"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

const (
	gameID   = "CHECKERSVEC00001"
	gameSalt = "checkersvec-secret-salt"
	red      = 0
	black    = 4
	width    = 8
)

type placement struct {
	Space   int
	Color   int
	Crowned bool
}

type positionSpec struct {
	Name string
	// Setup keeps the post-set-up position exactly as Go leaves it.
	Setup        bool
	Tokens       []placement
	Captured     [2][]int
	PlayerColors [2]int
	Current      int
	FinishedTurn [2]bool
}

type entry struct {
	Name            string   `json:"name"`
	Args            []string `json:"args"`
	Verdict         verdict  `json:"verdict"`
	Serializable    bool     `json:"serializable"`
	ClientEvaluable bool     `json:"clientEvaluable"`
	FieldDependent  bool     `json:"fieldDependent"`
	Reads           []read   `json:"reads"`
}

type verdict struct {
	Outcome string   `json:"outcome"`
	Message *message `json:"message,omitempty"`
	Reason  string   `json:"reason,omitempty"`
}

type message struct {
	Template string                                 `json:"template"`
	Bindings map[string]boardgame.LegalBindingValue `json:"bindings,omitempty"`
}

type read struct {
	Path  string `json:"path"`
	Facet string `json:"facet"`
}

type applyVector struct {
	Move   string                 `json:"move"`
	Fields map[string]interface{} `json:"fields"`
	Error  string                 `json:"error"`
	Panic  string                 `json:"panic,omitempty"`
	After  json.RawMessage        `json:"after,omitempty"`
}

type positionVector struct {
	Name            string          `json:"name"`
	Version         int             `json:"version"`
	State           json.RawMessage `json:"state"`
	MoveLegal       [][4]int        `json:"moveLegal"`
	MoveLedgers     [][]interface{} `json:"moveLedgers"`
	CrownDefault    int             `json:"crownDefault"`
	CrownLegal      [][2]int        `json:"crownLegal"`
	FinishTurnLegal int             `json:"finishTurnLegal"`
	Applies         []applyVector   `json:"applies"`
	ConditionMet    bool            `json:"conditionMet"`
	Scores          []int           `json:"scores"`
}

type output struct {
	Note string `json:"note"`
	Game struct {
		ID         string `json:"id"`
		SecretSalt string `json:"secretSalt"`
		NumPlayers int    `json:"numPlayers"`
	} `json:"game"`
	Texts     []string         `json:"texts"`
	Entries   []entry          `json:"entries"`
	Positions []positionVector `json:"positions"`
	Summary   map[string]int   `json:"summary"`
}

type tables struct {
	texts     []string
	textIndex map[string]int
	entries   []entry
	entryKey  map[string]int
}

// text interns a non-empty error text; "" (Go's nil) is -1.
func (t *tables) text(s string) int {
	if s == "" {
		return -1
	}
	if i, ok := t.textIndex[s]; ok {
		return i
	}
	t.texts = append(t.texts, s)
	t.textIndex[s] = len(t.texts) - 1
	return len(t.texts) - 1
}

func (t *tables) entry(e entry) int {
	blob, err := json.Marshal(e)
	check(err)
	key := string(blob)
	if i, ok := t.entryKey[key]; ok {
		return i
	}
	t.entries = append(t.entries, e)
	t.entryKey[key] = len(t.entries) - 1
	return len(t.entries) - 1
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func outcome(o boardgame.LegalOutcome) string {
	switch o {
	case boardgame.LegalPass:
		return "pass"
	case boardgame.LegalFail:
		return "fail"
	case boardgame.LegalUnknown:
		return "unknown"
	}
	return "invalid"
}

func facet(f boardgame.LegalFacet) string {
	switch f {
	case boardgame.LegalFacetValues:
		return "values"
	case boardgame.LegalFacetCount:
		return "count"
	case boardgame.LegalFacetOccupancy:
		return "occupancy"
	case boardgame.LegalFacetOrder:
		return "order"
	case boardgame.LegalFacetNonEmpty:
		return "nonempty"
	}
	return "invalid"
}

func exportVerdict(v boardgame.LegalVerdict) verdict {
	out := verdict{Outcome: outcome(v.Outcome), Reason: v.Reason}
	if v.Message != nil {
		out.Message = &message{Template: v.Message.Template}
		if len(v.Message.Bindings) > 0 {
			out.Message.Bindings = v.Message.Bindings
		}
	}
	return out
}

func exportEntry(e boardgame.LegalVerdictEntry) entry {
	out := entry{Name: e.Name, Args: e.Args, Verdict: exportVerdict(e.Verdict), Serializable: e.Serializable, ClientEvaluable: e.ClientEvaluable, FieldDependent: e.FieldDependent, Reads: []read{}}
	if out.Args == nil {
		out.Args = []string{}
	}
	for _, r := range e.Reads {
		out.Reads = append(out.Reads, read{Path: string(r.Path), Facet: facet(r.Facet)})
	}
	return out
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// guarded runs f, turning a panic into "panic: <value>".
func guarded(f func() error) (text string) {
	defer func() {
		if r := recover(); r != nil {
			text = "panic: " + fmt.Sprint(r)
		}
	}()
	return errText(f())
}

func blob(state boardgame.ImmutableState) json.RawMessage {
	var generic interface{}
	check(json.Unmarshal(state.StorageRecord(), &generic))
	out, err := json.Marshal(generic)
	check(err)
	return out
}

func copyState(state boardgame.State) boardgame.State {
	copied, err := state.Copy(false)
	check(err)
	return copied.(boardgame.State)
}

func spaceString(key int) string {
	return fmt.Sprintf("%d_%d", key/width, key%width)
}

func stackProp(sub boardgame.SubState, name string) boardgame.Stack {
	s, err := sub.ReadSetter().StackProp(name)
	check(err)
	return s
}

func tokenColor(c boardgame.ImmutableComponentInstance) int {
	v, err := c.Values().Reader().ImmutableEnumProp("Color")
	check(err)
	return int(v.Value())
}

func setCrowned(c boardgame.ComponentInstance, crowned bool) {
	check(c.DynamicValues().ReadSetter().SetBoolProp("Crowned", crowned))
}

// takeToken moves a token of color from UnusedTokens with move.
func takeToken(unused boardgame.Stack, color int, move func(boardgame.ComponentInstance) error) boardgame.ComponentInstance {
	for _, c := range unused.Components() {
		if c != nil && tokenColor(c) == color {
			check(move(c))
			return c
		}
	}
	panic(fmt.Sprintf("checkersvec: no unused token of color %d left", color))
}

func construct(base boardgame.State, spec positionSpec) boardgame.State {
	st := copyState(base)
	if spec.Setup {
		return st
	}
	game := st.GameState()
	spaces := stackProp(game, "Spaces")
	unused := stackProp(game, "UnusedTokens")
	for i := 0; i < spaces.Len(); i++ {
		if c := spaces.ComponentAt(i); c != nil {
			check(c.MoveToLastSlot(unused))
		}
	}
	for _, c := range unused.Components() {
		setCrowned(c, false)
	}
	for _, p := range spec.Tokens {
		c := takeToken(unused, p.Color, func(c boardgame.ComponentInstance) error { return c.MoveTo(spaces, p.Space) })
		setCrowned(c, p.Crowned)
	}
	for i, colors := range spec.Captured {
		captured := stackProp(st.PlayerStates()[i], "CapturedTokens")
		for _, color := range colors {
			takeToken(unused, color, func(c boardgame.ComponentInstance) error { return c.MoveToLastSlot(captured) })
		}
	}
	for i, player := range st.PlayerStates() {
		color, err := player.ReadSetter().EnumProp("Color")
		check(err)
		check(color.SetValue(enum.EnumKey(spec.PlayerColors[i])))
		check(player.ReadSetter().SetBoolProp("FinishedTurn", spec.FinishedTurn[i]))
	}
	check(game.ReadSetter().SetPlayerIndexProp("CurrentPlayer", boardgame.PlayerIndex(spec.Current)))
	return st
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func chebyshev(a, b int) int {
	dr, dc := abs(a/width-b/width), abs(a%width-b%width)
	if dr > dc {
		return dr
	}
	return dc
}

func moveToken(game *boardgame.Game, st boardgame.ImmutableState, token, space int) boardgame.Move {
	mv := game.MoveByNameForState("Move Token", st)
	rs := mv.ReadSetter()
	tokenVal, err := rs.EnumProp("TokenIndexToMove")
	check(err)
	check(tokenVal.SetValue(enum.EnumKey(token)))
	spaceVal, err := rs.EnumProp("SpaceIndex")
	check(err)
	check(spaceVal.SetValue(enum.EnumKey(space)))
	return mv
}

func fieldsOf(mv boardgame.Move) map[string]interface{} {
	out := map[string]interface{}{}
	reader := mv.Reader()
	for name, t := range reader.Props() {
		switch t {
		case boardgame.TypeEnum:
			v, err := reader.ImmutableEnumProp(name)
			check(err)
			out[name] = v.String()
		case boardgame.TypeInt:
			v, err := reader.IntProp(name)
			check(err)
			out[name] = v
		case boardgame.TypePlayerIndex:
			v, err := reader.PlayerIndexProp(name)
			check(err)
			out[name] = int(v)
		case boardgame.TypeBool:
			v, err := reader.BoolProp(name)
			check(err)
			out[name] = v
		}
	}
	return out
}

func apply(st boardgame.State, build func(boardgame.State) boardgame.Move, summary map[string]int) applyVector {
	working := copyState(st)
	mv := build(working)
	out := applyVector{Move: mv.Info().Name(), Fields: fieldsOf(mv)}
	text := guarded(func() error { return mv.Apply(working) })
	summary["applies"]++
	if len(text) > 7 && text[:7] == "panic: " {
		out.Panic = text
		summary["applyPanics"]++
		return out
	}
	out.Error = text
	if text == "" {
		out.After = blob(working)
		summary["appliesOk"]++
	}
	return out
}

func record(manager *boardgame.GameManager, game *boardgame.Game, st boardgame.State, spec positionSpec, t *tables, summary map[string]int) positionVector {
	pv := positionVector{Name: spec.Name, Version: st.Version(), State: blob(st), MoveLegal: [][4]int{}, MoveLedgers: [][]interface{}{}, CrownLegal: [][2]int{}, Applies: []applyVector{}}
	spaces := stackProp(st.GameState(), "Spaces")
	current := int(st.CurrentPlayerIndex())

	var occupied, tokens []int
	empties := 0
	for i := 0; i < spaces.Len(); i++ {
		if spaces.ComponentAt(i) != nil {
			occupied = append(occupied, i)
			tokens = append(tokens, i)
		} else if empties < 2 {
			tokens = append(tokens, i)
			empties++
		}
	}
	sort.Ints(tokens)

	// Move Token: Legal for every (token, space, proposer).
	for _, token := range tokens {
		for space := 0; space < spaces.Len(); space++ {
			for proposer := 0; proposer < len(st.PlayerStates()); proposer++ {
				mv := moveToken(game, st, token, space)
				text := guarded(func() error { return mv.Legal(st, boardgame.PlayerIndex(proposer)) })
				if text == "" {
					summary["legalMoveTokens"]++
				}
				pv.MoveLegal = append(pv.MoveLegal, [4]int{token, space, proposer, t.text(text)})
			}
		}
	}
	// Ledgers near each token, for the current player.
	for _, token := range tokens {
		for space := 0; space < spaces.Len(); space++ {
			if chebyshev(token, space) > 2 {
				continue
			}
			mv := moveToken(game, st, token, space)
			v, entries, ok := manager.LegalEvaluateLedger("Move Token", st, mv, boardgame.PlayerIndex(current))
			if !ok {
				panic("checkersvec: Move Token has no plan")
			}
			indexes := []int{}
			for _, e := range entries {
				indexes = append(indexes, t.entry(exportEntry(e)))
			}
			verdictEntry := entry{Args: []string{}, Reads: []read{}, Verdict: exportVerdict(v)}
			pv.MoveLedgers = append(pv.MoveLedgers, []interface{}{token, space, current, t.entry(verdictEntry), indexes})
			summary["ledgers"]++
		}
	}
	// Crown Token and Finish Turn, for the admin.
	crown := game.MoveByNameForState("Crown Token", st)
	ci, err := crown.Reader().IntProp("ComponentIndex")
	check(err)
	pv.CrownDefault = ci
	for slot := -1; slot <= spaces.Len(); slot++ {
		mv := game.MoveByNameForState("Crown Token", st)
		check(mv.ReadSetter().SetIntProp("ComponentIndex", slot))
		text := guarded(func() error { return mv.Legal(st, boardgame.AdminPlayerIndex) })
		if text == "" {
			summary["legalCrowns"]++
		}
		pv.CrownLegal = append(pv.CrownLegal, [2]int{slot, t.text(text)})
	}
	finish := game.MoveByNameForState("Finish Turn", st)
	finishText := guarded(func() error { return finish.Legal(st, boardgame.AdminPlayerIndex) })
	if finishText == "" {
		summary["legalFinishTurns"]++
	}
	pv.FinishTurnLegal = t.text(finishText)

	// Apply.
	for _, token := range occupied {
		for space := 0; space < spaces.Len(); space++ {
			dr, dc := abs(token/width-space/width), abs(token%width-space%width)
			if dr != dc || (dr != 1 && dr != 2) {
				continue
			}
			token, space := token, space
			pv.Applies = append(pv.Applies, apply(st, func(w boardgame.State) boardgame.Move { return moveToken(game, w, token, space) }, summary))
		}
	}
	crownSlots := append([]int{}, occupied...)
	for i := 0; i < spaces.Len(); i++ {
		if spaces.ComponentAt(i) == nil {
			crownSlots = append(crownSlots, i)
			break
		}
	}
	for _, slot := range crownSlots {
		slot := slot
		pv.Applies = append(pv.Applies, apply(st, func(w boardgame.State) boardgame.Move {
			mv := game.MoveByNameForState("Crown Token", w)
			check(mv.ReadSetter().SetIntProp("ComponentIndex", slot))
			return mv
		}, summary))
	}
	pv.Applies = append(pv.Applies, apply(st, func(w boardgame.State) boardgame.Move { return game.MoveByNameForState("Finish Turn", w) }, summary))
	if stackProp(st.GameState(), "UnusedTokens").NumComponents() > 0 {
		for i := 0; i < spaces.Len(); i++ {
			if spaces.ComponentAt(i) != nil {
				continue
			}
			target := i
			pv.Applies = append(pv.Applies, apply(st, func(w boardgame.State) boardgame.Move {
				mv := game.MoveByNameForState("Place Token", w)
				v, err := mv.ReadSetter().EnumProp("TargetIndex")
				check(err)
				check(v.SetValue(enum.EnumKey(target)))
				return mv
			}, summary))
			break
		}
	}

	// base.GameDelegate's end hooks, which checkers overrides.
	delegate := manager.Delegate().(interface {
		GameEndConditionMet(state boardgame.ImmutableState) bool
		PlayerScore(pState boardgame.ImmutableSubState) int
	})
	pv.ConditionMet = delegate.GameEndConditionMet(st)
	for _, p := range st.ImmutablePlayerStates() {
		pv.Scores = append(pv.Scores, delegate.PlayerScore(p))
	}
	summary["positions"]++
	return pv
}

func designed() []positionSpec {
	rb := [2]int{red, black}
	return []positionSpec{
		{Name: "setup", Setup: true},
		{Name: "black-step", Tokens: []placement{{41, black, false}, {18, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "red-step", Tokens: []placement{{41, black, false}, {18, red, false}}, PlayerColors: rb, Current: 0},
		{Name: "black-capture", Tokens: []placement{{41, black, false}, {34, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "red-capture", Tokens: []placement{{27, red, false}, {36, black, false}}, PlayerColors: rb, Current: 0},
		{Name: "black-capture-then-more", Tokens: []placement{{41, black, false}, {34, red, false}, {20, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "mid-jump", Tokens: []placement{{27, black, false}, {20, red, false}, {54, black, false}}, PlayerColors: rb, Current: 1},
		{Name: "capture-landing-occupied", Tokens: []placement{{41, black, false}, {34, red, false}, {27, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "capture-off-board", Tokens: []placement{{9, black, false}, {0, red, false}, {2, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "jump-own-token", Tokens: []placement{{41, black, false}, {34, black, false}}, PlayerColors: rb, Current: 1},
		{Name: "capture-refused-by-same-color", Tokens: []placement{{41, black, false}, {34, red, false}}, Captured: [2][]int{nil, {black}}, PlayerColors: rb, Current: 1},
		{Name: "crowned-black-both-ways", Tokens: []placement{{27, black, true}, {36, red, false}, {18, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "crowned-red-both-ways", Tokens: []placement{{27, red, true}, {36, black, false}, {18, black, false}}, PlayerColors: rb, Current: 0},
		{Name: "crown-candidates", Tokens: []placement{{2, black, false}, {4, black, true}, {57, red, false}, {59, red, true}, {29, red, false}}, PlayerColors: rb, Current: 1},
		{Name: "turn-finished", Tokens: []placement{{41, black, false}, {18, red, false}}, PlayerColors: rb, Current: 1, FinishedTurn: [2]bool{false, true}},
		{Name: "game-over", Tokens: []placement{{41, black, false}}, Captured: [2][]int{nil, {red, red, red, red, red, red, red, red, red, red, red, red}}, PlayerColors: rb, Current: 0},
		{Name: "last-capture", Tokens: []placement{{41, black, false}, {34, red, false}}, Captured: [2][]int{nil, {red, red, red, red, red, red, red, red, red, red, red}}, PlayerColors: rb, Current: 1},
		{Name: "colors-swapped", Tokens: []placement{{41, black, false}, {18, red, false}}, PlayerColors: [2]int{black, red}, Current: 1},
	}
}

func random(n int) []positionSpec {
	rng := rand.New(rand.NewSource(20261122))
	colorSets := [][2]int{{red, black}, {red, black}, {black, red}}
	var out []positionSpec
	for i := 0; i < n; i++ {
		r0, c0 := rng.Intn(width-3), rng.Intn(width-3)
		k := 2 + rng.Intn(6)
		used := map[int]bool{}
		spec := positionSpec{Name: fmt.Sprintf("random-%02d", i), PlayerColors: colorSets[rng.Intn(len(colorSets))], Current: rng.Intn(2)}
		counts := map[int]int{}
		for len(spec.Tokens) < k {
			space := (r0+rng.Intn(4))*width + c0 + rng.Intn(4)
			if used[space] || (space/width+space%width)%2 != 0 {
				continue
			}
			used[space] = true
			color := red
			if rng.Intn(2) == 0 {
				color = black
			}
			counts[color]++
			spec.Tokens = append(spec.Tokens, placement{Space: space, Color: color, Crowned: rng.Intn(4) == 0})
		}
		if rng.Intn(3) == 0 {
			p := rng.Intn(2)
			color := red
			if rng.Intn(2) == 0 {
				color = black
			}
			spec.Captured[p] = []int{color}
		}
		spec.FinishedTurn = [2]bool{rng.Intn(4) == 0, rng.Intn(4) == 0}
		out = append(out, spec)
	}
	return out
}

func main() {
	storage := memstore.NewStorageManager()
	manager, err := boardgame.NewGameManager(checkers.NewDelegate(), storage)
	check(err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	game, err := manager.Internals().RecreateGame(&boardgame.GameStorageRecord{Name: "checkers", ID: gameID, SecretSalt: gameSalt, NumPlayers: 2})
	check(err)
	if game.Version() != 25 {
		panic(fmt.Sprintf("checkersvec: set-up ended at version %d, not 25", game.Version()))
	}
	baseImmutable, err := game.CurrentState().Copy(false)
	check(err)
	base := baseImmutable.(boardgame.State)

	t := &tables{textIndex: map[string]int{}, entryKey: map[string]int{}}
	summary := map[string]int{}
	out := output{Note: "Generated by testdata/gen/checkersvec from examples/checkers; see its doc comment.", Texts: []string{}}
	out.Game.ID, out.Game.SecretSalt, out.Game.NumPlayers = gameID, gameSalt, 2
	for _, spec := range append(designed(), random(30)...) {
		st := construct(base, spec)
		out.Positions = append(out.Positions, record(manager, game, st, spec, t, summary))
	}
	for _, pv := range out.Positions {
		for _, a := range pv.Applies {
			if a.Move == "Move Token" && a.After != nil {
				summary["moveTokenAppliesOk"]++
			}
		}
	}
	out.Texts, out.Entries, out.Summary = t.texts, t.entries, summary

	var raw bytes.Buffer
	enc := json.NewEncoder(&raw)
	enc.SetEscapeHTML(false)
	check(enc.Encode(out))
	var zipped bytes.Buffer
	zw, err := gzip.NewWriterLevel(&zipped, gzip.BestCompression)
	check(err)
	_, err = zw.Write(raw.Bytes())
	check(err)
	check(zw.Close())

	dest := filepath.Join("server", "static", "src", "engine", "testdata", "checkers.vectors.json.gz")
	check(os.WriteFile(dest, zipped.Bytes(), 0o644))
	summaryJSON, _ := json.Marshal(summary)
	fmt.Fprintf(os.Stderr, "checkersvec: %d positions, %d texts, %d entries, %d bytes raw, %d gzipped -> %s\n%s\n", len(out.Positions), len(out.Texts), len(out.Entries), raw.Len(), zipped.Len(), dest, summaryJSON)
}
