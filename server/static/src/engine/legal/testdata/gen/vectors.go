package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

/*
Predicate vectors: one resolved catalog predicate, evaluated against one
fixture for one proposer, with everything the Go produced - the verdict
(outcome, template key, bindings, Unknown reason) and the predicate's
construction metadata (reads and their facets, cost, client evaluability,
emitted templates and bindings, read-type contracts).

Evaluation is the corpus harness's: the spec is resolved straight through
legal.DefaultConstructors() with a nil chest and its Evaluate is called
directly (legal/conformance_test.go's resolveSpecViaRegistry). A panic inside
Evaluate is recorded as such rather than recovered into a verdict, because
the recovery (evalLegalPredicate) is unexported core code; plans.go records
what the recovery turns it into, through a real plan.
*/

type exportedMessage struct {
	Template string                                 `json:"template"`
	Bindings map[string]boardgame.LegalBindingValue `json:"bindings,omitempty"`
}

type exportedVerdict struct {
	Outcome string           `json:"outcome"`
	Message *exportedMessage `json:"message,omitempty"`
	Reason  string           `json:"reason,omitempty"`
	// Panic is set instead of everything else when Evaluate panicked.
	Panic string `json:"panic,omitempty"`
}

type exportedRead struct {
	Path  string `json:"path"`
	Facet string `json:"facet"`
}

type exportedPredicate struct {
	Name              string              `json:"name"`
	Args              []string            `json:"args"`
	Reads             []exportedRead      `json:"reads"`
	Cost              int                 `json:"cost"`
	ClientEvaluable   bool                `json:"clientEvaluable"`
	UsesProposer      bool                `json:"usesProposer"`
	Serializable      bool                `json:"serializable"`
	EmittedTemplates  []string            `json:"emittedTemplates"`
	EmittedBindings   map[string][]string `json:"emittedBindings"`
	RequiredReadTypes map[string]string   `json:"requiredReadTypes"`
	AllowedReadTypes  map[string][]string `json:"allowedReadTypes"`
}

type exportedVector struct {
	// Source is "corpus:<file>#<index>" or "extra:<group>".
	Source   string
	Spec     legal.Spec
	Fixture  string
	Proposer int
	Verdict  exportedVerdict
}

type exportedConstructorError struct {
	Spec  legal.Spec `json:"spec"`
	Error string     `json:"error"`
}

func outcomeString(o legal.Outcome) string {
	switch o {
	case legal.Pass:
		return "pass"
	case legal.Fail:
		return "fail"
	case legal.Unknown:
		return "unknown"
	default:
		return "invalid"
	}
}

func facetName(f boardgame.LegalFacet) string {
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
	default:
		panic("gen: unknown facet")
	}
}

func exportVerdict(v legal.Verdict) exportedVerdict {
	out := exportedVerdict{Outcome: outcomeString(v.Outcome), Reason: v.Reason}
	if v.Message != nil {
		out.Message = &exportedMessage{Template: v.Message.Template, Bindings: v.Message.Bindings}
	}
	return out
}

func exportReads(reads []boardgame.LegalRead) []exportedRead {
	out := []exportedRead{}
	for _, r := range reads {
		out = append(out, exportedRead{Path: string(r.Path), Facet: facetName(r.Facet)})
	}
	return out
}

func exportPredicate(p *legal.Predicate) exportedPredicate {
	out := exportedPredicate{
		Name:              p.Name,
		Args:              p.Args,
		Reads:             exportReads(p.Reads),
		Cost:              int(p.Cost),
		ClientEvaluable:   p.ClientEvaluable,
		UsesProposer:      p.UsesProposer,
		Serializable:      p.Serializable(),
		EmittedTemplates:  p.EmittedTemplates,
		EmittedBindings:   map[string][]string{},
		RequiredReadTypes: map[string]string{},
		AllowedReadTypes:  map[string][]string{},
	}
	if out.Args == nil {
		out.Args = []string{}
	}
	for key, names := range p.EmittedBindings {
		if names == nil {
			names = []string{}
		}
		out.EmittedBindings[key] = names
	}
	for path, t := range p.RequiredReadTypes {
		out.RequiredReadTypes[string(path)] = t.String()
	}
	for path, types := range p.AllowedReadTypes {
		for _, t := range types {
			out.AllowedReadTypes[string(path)] = append(out.AllowedReadTypes[string(path)], t.String())
		}
	}
	return out
}

func resolveSpecViaRegistry(spec legal.Spec, registry []*legal.PredicateConstructor, chest *boardgame.ComponentChest) (*legal.Predicate, error) {
	for _, c := range registry {
		if c.Name != spec.Name {
			continue
		}
		resolve := func(sub legal.Spec) (*legal.Predicate, error) {
			return resolveSpecViaRegistry(sub, registry, chest)
		}
		return c.Constructor(spec, chest, resolve)
	}
	return nil, fmt.Errorf("legal: unknown predicate name %q", spec.Name)
}

func evaluateRaw(pred *legal.Predicate, ctx legal.Context) (out exportedVerdict) {
	defer func() {
		if r := recover(); r != nil {
			out = exportedVerdict{Panic: fmt.Sprint(r)}
		}
	}()
	return exportVerdict(pred.Evaluate(ctx))
}

func specKey(spec legal.Spec) string {
	b, err := json.Marshal(spec)
	must(err, "marshaling spec")
	return string(b)
}

type fixtureCache struct {
	built map[string]fixture
}

func (c *fixtureCache) get(name string) fixture {
	if f, ok := c.built[name]; ok {
		return f
	}
	builder, ok := fixtureBuilders[name]
	if !ok {
		builder, ok = extraFixtureBuilders[name]
	}
	if !ok {
		panic("gen: unknown fixture " + name)
	}
	f := builder()
	c.built[name] = f
	return f
}

func (f fixture) context(proposer int) legal.Context {
	return legal.Context{
		State:               f.state,
		Move:                f.move,
		ProposerPlayerIndex: boardgame.PlayerIndex(proposer),
		Chest:               f.game.Manager().Chest(),
	}
}

type conformanceCase struct {
	Spec     legal.Spec `json:"spec"`
	Fixture  string     `json:"fixture"`
	Proposer int        `json:"proposer"`
	Expect   string     `json:"expect"`
	Template string     `json:"template,omitempty"`
}

type conformanceFile struct {
	Predicate string            `json:"predicate"`
	Cases     []conformanceCase `json:"cases"`
}

// corpusVectors replays legal/testdata/conformance exactly as
// TestConformanceCorpus does, and panics on the first disagreement: the
// ported fixtures are only trusted because the corpus agrees with them.
func corpusVectors(corpusDir string, fixtures *fixtureCache) []exportedVector {
	paths, err := filepath.Glob(filepath.Join(corpusDir, "*.json"))
	must(err, "globbing corpus")
	if len(paths) == 0 {
		panic("gen: no conformance corpus files under " + corpusDir)
	}
	sort.Strings(paths)
	var out []exportedVector
	for _, path := range paths {
		data, err := os.ReadFile(path)
		must(err, "reading "+path)
		var cf conformanceFile
		must(json.Unmarshal(data, &cf), "parsing "+path)
		for i, c := range cf.Cases {
			pred, err := resolveSpecViaRegistry(c.Spec, legal.DefaultConstructors(), nil)
			must(err, fmt.Sprintf("resolving %s case %d", path, i))
			verdict := evaluateRaw(pred, fixtures.get(c.Fixture).context(c.Proposer))
			if verdict.Outcome != c.Expect {
				panic(fmt.Sprintf("gen: %s case %d: ported fixture %q gave %q, corpus expects %q", path, i, c.Fixture, verdict.Outcome, c.Expect))
			}
			if c.Template != "" && (verdict.Message == nil || verdict.Message.Template != c.Template) {
				panic(fmt.Sprintf("gen: %s case %d: template %+v, corpus expects %q", path, i, verdict.Message, c.Template))
			}
			out = append(out, exportedVector{
				Source:   fmt.Sprintf("corpus:%s#%d", filepath.Base(path), i),
				Spec:     c.Spec,
				Fixture:  c.Fixture,
				Proposer: c.Proposer,
				Verdict:  verdict,
			})
		}
	}
	return out
}

type extraGroup struct {
	name      string
	fixtures  []string
	proposers []int
	specs     []legal.Spec
}

func sp(name string, args ...string) legal.Spec {
	if args == nil {
		args = []string{}
	}
	return legal.Spec{Name: name, Args: args}
}

func sel(stackPath, wire string) []string { return []string{stackPath, wire} }

var memoryStates = []string{
	"memoryDefault", "memoryZeroCardsLeft", "memorySeatFilled", "memoryVisibleOccupied",
	"memoryCardAlreadyRevealed", "memoryCardNeverThere", "memoryNoMove",
	"xMemoryCardIndexHuge", "xMemoryCardIndexNegative", "xMemoryAllRevealed", "xMemoryPlayerZeroWonCards",
}

var memoryTargets = []string{
	"memoryDefault", "memoryTargetPlayerOne", "memoryTargetObserver", "memoryNoMove",
	"xMemoryTargetOutOfBounds", "xMemoryTargetAdmin", "xMemoryTargetAny", "xMemoryPlayerOneInactiveTargeted",
}

var everyProposer = []int{0, 1, 5, -1, -2, -3, -4}

// extraGroups is where the corpus is thin: every binding a Fail carries,
// every Unknown reason, out-of-range indexes, every path kind, the special
// player indexes, and the move-legality error strings stacks produce.
var extraGroups = []extraGroup{
	{
		name:      "paths",
		fixtures:  []string{"memoryDefault", "memoryNoMove", "xMemoryPlayerZeroWonCards", "memoryTargetPlayerOne", "memoryTargetObserver", "xMemoryTargetOutOfBounds"},
		proposers: []int{0, 1, -1, -2, 7},
		specs: []legal.Spec{
			sp("propAtLeast", "game.NumCards", "20"),
			sp("propAtLeast", "player.CardsLeftToReveal", "2"),
			sp("propAtLeast", "proposer.CardsLeftToReveal", "2"),
			sp("propAtLeast", "players[move.TargetPlayerIndex].CardsLeftToReveal", "2"),
			sp("propAtLeast", "players[move.CardIndex].CardsLeftToReveal", "2"),
			sp("propAtLeast", "players[move.NoSuchField].CardsLeftToReveal", "2"),
			sp("propAtLeast", "players[*].CardsLeftToReveal", "2"),
			sp("propAtLeast", "move.CardIndex", "0"),
			sp("propAtLeast", "move.NoSuchField", "0"),
			sp("propAtLeast", "player.SeatFilled", "1"),
			sp("propAtLeast", "player.NoSuchProp", "1"),
			sp("propAtLeast", "nokind.Whatever", "1"),
			sp("propAtLeast", "nodothere", "1"),
			sp("propAtLeast", "game.", "1"),
			sp("propAtLeast", "players[0].CardsLeftToReveal", "1"),
			sp("propAtLeast", "players[move.].CardsLeftToReveal", "1"),
			sp("propAtLeast", "players[move.TargetPlayerIndex]", "1"),
			sp("propAtLeast", "players[move.TargetPlayerIndex", "1"),
			sp("propAtLeast", "game.HiddenCards[move.CardIndex]", "1"),
			sp("propAtLeast", "game.NumCards[move.CardIndex]", "1"),
			sp("propAtLeast", "move.CardIndex[move.CardIndex]", "1"),
			sp("propAtLeast", "game.NumCards]", "1"),
			sp("propAtLeast", "game.NumCards[0]", "1"),
			sp("propAtLeast", "game.NumCards[move.A.B]", "1"),
			sp("propAtLeast", "player.WonCards[move.CardIndex]", "1"),
			sp("stackCount", "player.WonCards", ">=", "2"),
			sp("stackCount", "proposer.WonCards", "==", "0"),
			sp("stackCount", "players[move.TargetPlayerIndex].WonCards", "<", "2"),
			sp("stackCount", "game.Cards", "==", "20"),
			sp("stackCount", "game.NumCards", "==", "20"),
			sp("stackEmpty", "player.WonCards"),
			sp("stackNotEmpty", "proposer.WonCards"),
			sp("playerBoolAt", "player.SeatFilled", "false"),
			sp("playerBoolAt", "proposer.SeatFilled", "true"),
			sp("playerBoolAt", "players[move.TargetPlayerIndex].SeatFilled", "true"),
			sp("playerBoolAt", "players[move.TargetPlayerIndex].PlayerInactive", "false"),
			sp("playerBoolAt", "proposer.CardsLeftToReveal", "true"),
		},
	},
	{
		name:      "compare",
		fixtures:  []string{"memoryDefault", "memoryZeroCardsLeft", "memoryTargetObserver", "xMemoryTargetAdmin", "xMemoryTargetAny"},
		proposers: []int{0},
		specs: []legal.Spec{
			sp("propAtLeast", "player.CardsLeftToReveal", "-1"),
			sp("propAtLeast", "player.CardsLeftToReveal", "+2"),
			sp("propAtLeast", "player.CardsLeftToReveal", "3"),
			sp("propAtLeast", "game.NumCards", "21"),
			sp("propCompare", "player.CardsLeftToReveal", "==", "0"),
			sp("propCompare", "player.CardsLeftToReveal", "!=", "0"),
			sp("propCompare", "player.CardsLeftToReveal", "<", "2"),
			sp("propCompare", "player.CardsLeftToReveal", "<=", "0"),
			sp("propCompare", "player.CardsLeftToReveal", ">", "2"),
			sp("propCompare", "player.CardsLeftToReveal", ">=", "1"),
			sp("propCompare", "game.NumCards", "<", "007"),
			sp("propEquals", "player.CardsLeftToReveal", "abc"),
			sp("propEquals", "player.CardsLeftToReveal", "+2"),
			sp("propEquals", "player.CardsLeftToReveal", "02"),
			sp("propEquals", "player.CardsLeftToReveal", "true"),
			sp("propEquals", "player.PlayerInactive", "1"),
			sp("propEquals", "player.PlayerInactive", "TRUE"),
			sp("propEquals", "player.SeatFilled", "false"),
			sp("propEquals", "move.TargetPlayerIndex", "admin"),
			sp("propEquals", "move.TargetPlayerIndex", "observer"),
			sp("propEquals", "move.TargetPlayerIndex", "any"),
			sp("propEquals", "move.TargetPlayerIndex", "-2"),
			sp("propEquals", "move.TargetPlayerIndex", "-3"),
			sp("propEquals", "move.TargetPlayerIndex", "+0"),
			sp("propEquals", "move.TargetPlayerIndex", "true"),
			sp("propEquals", "game.HideCardsTimer", "0"),
			sp("propEquals", "game.Cards", "0"),
			sp("propEquals", "move.CardIndex", "0"),
			sp("propNotEquals", "player.CardsLeftToReveal", "abc"),
			sp("propNotEquals", "move.TargetPlayerIndex", "admin"),
			sp("propNotEquals", "move.TargetPlayerIndex", "observer"),
			sp("propNotEquals", "player.SeatFilled", "false"),
			sp("playerBool", "PlayerInactive"),
			sp("playerBool", "PlayerInactive", "false"),
			sp("playerBool", "CardsLeftToReveal"),
			sp("playerBool", "CardsLeftToReveal", "false"),
		},
	},
	{
		name:      "occupancy",
		fixtures:  memoryStates,
		proposers: []int{0},
		specs: []legal.Spec{
			sp("componentPresentAt", "game.HiddenCards", "move.CardIndex"),
			sp("componentPresentAt", "game.VisibleCards", "move.CardIndex"),
			sp("componentPresentAt", "game.Cards", "move.CardIndex"),
			sp("componentPresentAt", "game.HiddenCards", "player.CardsLeftToReveal"),
			sp("componentPresentAt", "player.WonCards", "move.CardIndex"),
			sp("componentPresentAt", "game.NumCards", "move.CardIndex"),
			sp("componentPresentAt", "game.HiddenCards", "player.SeatFilled"),
			sp("componentAbsentAt", "game.HiddenCards", "move.CardIndex"),
			sp("componentAbsentAt", "game.VisibleCards", "move.CardIndex"),
			sp("componentAbsentAt", "player.WonCards", "move.CardIndex"),
			sp("revealableCardAt", "game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
			sp("revealableCardAt", "game.VisibleCards", "game.HiddenCards", "move.CardIndex"),
			sp("revealableCardAt", "game.HiddenCards", "game.NoSuchStack", "move.CardIndex"),
			sp("stackCount", "game.HiddenCards", "==", "20"),
			sp("stackCount", "game.HiddenCards", "!=", "20"),
			sp("stackCount", "game.VisibleCards", "<", "1"),
			sp("stackCount", "game.VisibleCards", "<=", "0"),
			sp("stackCount", "game.VisibleCards", ">", "0"),
			sp("stackCount", "game.Cards", ">=", "21"),
			sp("stackEmpty", "game.VisibleCards"),
			sp("stackEmpty", "game.HiddenCards"),
			sp("stackEmpty", "game.UnusedCards"),
			sp("stackEmpty", "game.Cards"),
			sp("stackNotEmpty", "game.VisibleCards"),
			sp("stackNotEmpty", "game.HiddenCards"),
			sp("stackNotEmpty", "game.Cards"),
			sp("stackNotEmpty", "game.NumCards"),
		},
	},
	{
		name:      "moves",
		fixtures:  memoryStates,
		proposers: []int{0},
		specs: []legal.Spec{
			sp("mayMoveTo", "game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
			sp("mayMoveTo", "game.HiddenCards", "player.WonCards", "move.CardIndex"),
			sp("mayMoveTo", "game.HiddenCards", "game.UnusedCards", "move.CardIndex"),
			sp("mayMoveTo", "game.VisibleCards", "game.HiddenCards", "move.CardIndex"),
			sp("mayMoveTo", "player.WonCards", "game.HiddenCards", "move.CardIndex"),
			sp("mayMoveTo", "player.WonCards", "player.WonCards", "move.CardIndex"),
			sp("mayMoveTo", "game.HiddenCards", "game.NumCards", "move.CardIndex"),
			sp("mayMoveTo", "game.HiddenCards", "game.VisibleCards", "player.CardsLeftToReveal"),
			sp("mayMoveToSlot", "game.HiddenCards", "game.VisibleCards", "move.CardIndex", "move.CardIndex"),
			sp("mayMoveToSlot", "game.HiddenCards", "game.VisibleCards", "player.CardsLeftToReveal", "move.CardIndex"),
			sp("mayMoveToSlot", "game.HiddenCards", "game.VisibleCards", "move.CardIndex", "player.CardsLeftToReveal"),
			sp("mayMoveToSlot", "game.HiddenCards", "player.WonCards", "player.CardsLeftToReveal", "move.CardIndex"),
			sp("mayMoveToSlot", "game.HiddenCards", "player.WonCards", "player.CardsLeftToReveal", "player.CardsLeftToReveal"),
			sp("mayMoveToSlot", "game.VisibleCards", "game.HiddenCards", "move.CardIndex", "move.CardIndex"),
			sp("mayMoveFirstTo", "game.HiddenCards", "game.VisibleCards"),
			sp("mayMoveFirstTo", "game.VisibleCards", "game.HiddenCards"),
			sp("mayMoveFirstTo", "game.HiddenCards", "player.WonCards"),
			sp("mayMoveFirstTo", "player.WonCards", "game.UnusedCards"),
			sp("mayMoveFirstTo", "game.HiddenCards", "game.HiddenCards"),
			sp("mayMoveFirstTo", "game.NoSuchStack", "game.VisibleCards"),
			sp("mayMoveFirstTo", "game.HiddenCards", "game.NumCards"),
			sp("mayMoveFirstToSlot", "game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
			sp("mayMoveFirstToSlot", "game.VisibleCards", "game.HiddenCards", "move.CardIndex"),
			sp("mayMoveFirstToSlot", "game.HiddenCards", "player.WonCards", "move.CardIndex"),
			sp("mayMoveAllTo", "game.HiddenCards", "game.VisibleCards"),
			sp("mayMoveAllTo", "game.VisibleCards", "game.HiddenCards"),
			sp("mayMoveAllTo", "game.HiddenCards", "player.WonCards"),
			sp("mayMoveAllTo", "player.WonCards", "game.UnusedCards"),
			sp("mayMoveAllTo", "game.HiddenCards", "game.HiddenCards"),
			sp("mayMoveAllTo", "game.Cards", "game.UnusedCards"),
			sp("mayMoveAllTo", "game.HiddenCards", "game.NumCards"),
			sp("mayMoveCountTo", "game.HiddenCards", "game.VisibleCards", "player.CardsLeftToReveal"),
			sp("mayMoveCountTo", "game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
			sp("mayMoveCountTo", "game.HiddenCards", "player.WonCards", "move.CardIndex"),
			sp("mayMoveCountTo", "game.VisibleCards", "game.HiddenCards", "game.NumCards"),
			sp("mayMoveCountTo", "game.HiddenCards", "game.VisibleCards", "player.SeatFilled"),
			sp("mayMoveFixedCountTo", "game.HiddenCards", "game.VisibleCards", "0"),
			sp("mayMoveFixedCountTo", "game.HiddenCards", "game.VisibleCards", "1"),
			sp("mayMoveFixedCountTo", "game.HiddenCards", "game.VisibleCards", "20"),
			sp("mayMoveFixedCountTo", "game.HiddenCards", "game.VisibleCards", "21"),
			sp("mayMoveFixedCountTo", "game.VisibleCards", "player.WonCards", "1"),
			sp("mayMoveFixedCountTo", "game.Cards", "game.UnusedCards", "1"),
			sp("maySwapComponents", "game.HiddenCards", "move.CardIndex", "player.CardsLeftToReveal"),
			sp("maySwapComponents", "game.HiddenCards", "player.CardsLeftToReveal", "move.CardIndex"),
			sp("maySwapComponents", "game.HiddenCards", "move.CardIndex", "move.CardIndex"),
			sp("maySwapComponents", "player.WonCards", "move.CardIndex", "player.CardsLeftToReveal"),
			sp("maySwapComponents", "game.Cards", "move.CardIndex", "player.CardsLeftToReveal"),
			sp("maySwapComponents", "game.HiddenCards", "move.CardIndex", "player.SeatFilled"),
			sp("stackConstraints", "HiddenCards", "VisibleCards"),
			sp("stackConstraints", "VisibleCards", "HiddenCards"),
			sp("stackConstraints", "HiddenCards", "HiddenCards"),
			sp("stackConstraints", "HiddenCards", "UnusedCards"),
			sp("stackConstraints", "HiddenCards", "NoSuchStack"),
			sp("stackConstraints", "NoSuchStack", "VisibleCards"),
			sp("stackConstraints", "NumCards", "VisibleCards"),
			sp("stackConstraints", "Cards", "UnusedCards"),
		},
	},
	{
		name:      "componentScalars",
		fixtures:  []string{"memoryDefault", "memoryVisibleOccupied", "memoryVisibleMatchingSparse", "memoryVisibleDifferentSparse", "memoryNoMove", "xMemoryCardIndexHuge", "xMemoryCardIndexNegative", "xMemoryPlayerZeroWonCards"},
		proposers: []int{0},
		specs: []legal.Spec{
			sp("componentPropEquals", append(sel("game.HiddenCards", "first-occupied"), "Type", "anything")...),
			sp("componentPropEquals", append(sel("game.HiddenCards", "first-occupied"), "NoSuchField", "anything")...),
			sp("componentPropEquals", append(sel("game.HiddenCards", "occupied-ordinal:0"), "Type", "anything")...),
			sp("componentPropEquals", append(sel("game.HiddenCards", "occupied-ordinal:99"), "Type", "anything")...),
			sp("componentPropEquals", append(sel("game.HiddenCards", "move-index:move.CardIndex"), "Type", "anything")...),
			sp("componentPropEquals", append(sel("game.NoSuchStack", "first-occupied"), "Type", "anything")...),
			sp("componentPropEquals", append(sel("game.VisibleCards", "occupied-ordinal:1"), "Type", "anything")...),
			sp("componentPropEquals", append(sel("player.WonCards", "first-occupied"), "Type", "anything")...),
			sp("componentPropNotEquals", append(sel("game.HiddenCards", "first-occupied"), "Type", "anything")...),
			sp("componentPropNotEquals", append(sel("game.HiddenCards", "move-index:move.CardIndex"), "Type", "anything")...),
			sp("componentPropNotEquals", append(sel("game.Cards", "first-occupied"), "Type", "anything")...),
			sp("componentPropsEqual", "game.HiddenCards", "first-occupied", "game.HiddenCards", "first-occupied", "Type"),
			sp("componentPropsEqual", "game.HiddenCards", "first-occupied", "game.HiddenCards", "occupied-ordinal:1", "Type"),
			sp("componentPropsEqual", "game.HiddenCards", "move-index:move.CardIndex", "game.VisibleCards", "first-occupied", "Type"),
			sp("componentPropsEqual", "game.VisibleCards", "first-occupied", "game.HiddenCards", "occupied-ordinal:99", "Type"),
			sp("componentPropsEqual", "game.HiddenCards", "first-occupied", "game.HiddenCards", "first-occupied", "NoSuchField"),
			sp("componentPropsNotEqual", "game.HiddenCards", "first-occupied", "game.HiddenCards", "first-occupied", "Type"),
			sp("componentPropsNotEqual", "game.HiddenCards", "first-occupied", "game.HiddenCards", "occupied-ordinal:1", "Type"),
			sp("componentPropsNotEqual", "game.HiddenCards", "move-index:move.CardIndex", "game.VisibleCards", "occupied-ordinal:0", "Type"),
		},
	},
	{
		name:      "proposer",
		fixtures:  memoryTargets,
		proposers: everyProposer,
		specs: []legal.Spec{
			sp("proposerIsCurrentPlayer"),
			sp("proposerIsPlayerFromMove", "TargetPlayerIndex"),
			sp("proposerIsPlayerFromMove", "CardIndex"),
			sp("proposerIsPlayerFromMove", "NoSuchField"),
			sp("propEquals", "move.TargetPlayerIndex", "1"),
			sp("propAtLeast", "proposer.CardsLeftToReveal", "1"),
		},
	},
	{
		name:      "phase",
		fixtures:  []string{"memoryDefault", "checkersDefault", "xBlackjackDefault"},
		proposers: []int{0},
		specs: []legal.Spec{
			sp("inPhase"),
			sp("inPhase", "0"),
			sp("inPhase", "1"),
			sp("inPhase", "7"),
			sp("inPhase", "7", "1"),
			sp("inPhase", "-3", "0"),
		},
	},
	{
		name:      "checkers",
		fixtures:  []string{"checkersDefault", "checkersEmptySpace", "checkersNoMove", "checkersDistinctSpaces", "checkersOwnToken", "checkersOpponentToken", "checkersBlackFirst", "xCheckersCaptured"},
		proposers: []int{0, 1},
		specs: []legal.Spec{
			sp("propEquals", "player.Color", "red"),
			sp("propEquals", "player.Color", " RED "),
			sp("propEquals", "player.Color", "0"),
			sp("propEquals", "player.Color", "4"),
			sp("propEquals", "player.Color", "3"),
			sp("propEquals", "proposer.Color", "Black"),
			sp("propEquals", "game.Phase", "Playing"),
			sp("propEquals", "game.Phase", "Setup"),
			sp("propEquals", "game.Phase", "1"),
			sp("propEquals", "move.TokenIndexToMove", "0_1"),
			sp("propEquals", "move.TokenIndexToMove", "1"),
			sp("propNotEquals", "player.Color", "black"),
			sp("propNotEquals", "player.Color", "Purple"),
			sp("propNotEquals", "game.Phase", "Playing"),
			sp("componentPresentAtKey", "game.Spaces", "move.TokenIndexToMove"),
			sp("componentPresentAtKey", "game.Spaces", "move.SpaceIndex"),
			sp("componentPresentAtKey", "game.Spaces", "move.TargetPlayerIndex"),
			sp("componentPresentAtKey", "game.UnusedTokens", "move.TokenIndexToMove"),
			sp("componentPresentAtKey", "player.CapturedTokens", "move.TokenIndexToMove"),
			sp("componentPresentAtKey", "game.NoSuchStack", "move.TokenIndexToMove"),
			sp("componentPropEqualsCurrentPlayer", "game.Spaces", "move.TokenIndexToMove", "Color"),
			sp("componentPropEqualsCurrentPlayer", "game.Spaces", "move.SpaceIndex", "Color"),
			sp("componentPropEqualsCurrentPlayer", "game.Spaces", "move.TokenIndexToMove", "NoSuchProp"),
			sp("componentPropEqualsCurrentPlayer", "game.Spaces", "move.TokenIndexToMove", "SeatFilled"),
			sp("maySwapComponentsByKey", "game.Spaces", "move.TokenIndexToMove", "move.SpaceIndex"),
			sp("maySwapComponentsByKey", "game.Spaces", "move.SpaceIndex", "move.SpaceIndex"),
			sp("maySwapComponentsByKey", "game.UnusedTokens", "move.TokenIndexToMove", "move.SpaceIndex"),
			sp("maySwapComponentsByKey", "game.Spaces", "move.TargetPlayerIndex", "move.SpaceIndex"),
			sp("maySwapComponentsByKey", "game.Spaces", "move.TokenIndexToMove", "move.TargetPlayerIndex"),
			sp("componentPropEquals", "game.Spaces", "first-occupied", "Color", "black"),
			sp("componentPropEquals", "game.Spaces", "first-occupied", "Color", " RED "),
			sp("componentPropEquals", "game.Spaces", "first-occupied", "Color", "4"),
			sp("componentPropEquals", "game.Spaces", "first-occupied", "Color", "Purple"),
			sp("componentPropEquals", "game.Spaces", "occupied-ordinal:23", "Color", "Red"),
			sp("componentPropEquals", "game.Spaces", "occupied-ordinal:24", "Color", "Red"),
			sp("componentPropNotEquals", "game.Spaces", "occupied-ordinal:12", "Color", "Red"),
			sp("componentPropNotEquals", "game.Spaces", "first-occupied", "Color", "Purple"),
			sp("componentPropsEqual", "game.Spaces", "first-occupied", "game.Spaces", "occupied-ordinal:1", "Color"),
			sp("componentPropsEqual", "game.Spaces", "first-occupied", "game.Spaces", "occupied-ordinal:22", "Color"),
			sp("componentPropsEqual", "game.Spaces", "first-occupied", "player.CapturedTokens", "first-occupied", "Color"),
			sp("componentPropsNotEqual", "game.Spaces", "first-occupied", "game.Spaces", "occupied-ordinal:22", "Color"),
			sp("componentPropsNotEqual", "game.Spaces", "first-occupied", "game.Spaces", "occupied-ordinal:1", "Color"),
			sp("stackCount", "game.Spaces", "==", "24"),
			sp("stackCount", "player.CapturedTokens", ">", "0"),
			sp("stackCount", "proposer.CapturedTokens", ">", "0"),
			sp("stackEmpty", "game.UnusedTokens"),
			sp("mayMoveFirstTo", "game.Spaces", "game.UnusedTokens"),
			sp("mayMoveAllTo", "game.Spaces", "game.UnusedTokens"),
			sp("mayMoveFixedCountTo", "game.Spaces", "game.UnusedTokens", "3"),
			sp("stackConstraints", "Spaces", "UnusedTokens"),
			sp("stackConstraints", "UnusedTokens", "Spaces"),
		},
	},
	{
		name:      "allActivePlayers",
		fixtures:  []string{"blackjackAllFinished", "blackjackOneUnfinished", "blackjackInactiveSkipped", "xBlackjackDefault", "xBlackjackPlayerOneEliminated", "xBlackjackEveryoneInactive"},
		proposers: []int{0},
		specs: []legal.Spec{
			legal.AllActivePlayers(legal.PlayerBool("Stood")),
			legal.AllActivePlayers(legal.PlayerBoolIs("Stood", false)),
			legal.AllActivePlayers(legal.PlayerBoolIs("Eliminated", false)),
			legal.AllActivePlayers(legal.Any(legal.PlayerBool("Eliminated"), legal.PlayerBool("Stood"))),
			legal.AllActivePlayers(legal.Any(legal.PlayerBool("NoSuchProp"), legal.PlayerBool("Stood"))),
			legal.AllActivePlayers(legal.Any(legal.PlayerBool("Stood"), legal.PlayerBool("NoSuchProp"))),
			legal.AllActivePlayers(legal.Any(legal.PlayerBool("NoSuchProp"), legal.PlayerBool("AlsoMissing"))),
			legal.AllActivePlayers(legal.PlayerBool("Stood")).WithMessage("gen.everyone_must_stand"),
			legal.AllActivePlayers(legal.PropAtLeast("player.Stood", 1)),
		},
	},
	{
		name:      "allActivePlayersInts",
		fixtures:  []string{"memoryDefault", "memoryZeroCardsLeft", "xMemoryPlayerZeroWonCards", "xMemoryPlayerOneInactiveTargeted"},
		proposers: []int{0},
		specs: []legal.Spec{
			legal.AllActivePlayers(legal.PropAtLeast("player.CardsLeftToReveal", 1)),
			legal.AllActivePlayers(legal.PropAtLeast("player.CardsLeftToReveal", 2)),
			legal.AllActivePlayers(legal.PropCompare("player.CardsLeftToReveal", "<=", 1)),
			legal.AllActivePlayers(legal.PropCompare("player.CardsLeftToReveal", "!=", 1)),
			legal.AllActivePlayers(legal.PropCompare("player.NoSuchProp", "==", 1)),
			legal.AllActivePlayers(legal.Any(legal.PropAtLeast("player.CardsLeftToReveal", 2), legal.PlayerBool("SeatFilled"), legal.PropCompare("player.CardsLeftToReveal", "==", 0))),
			legal.AllActivePlayers(legal.Any(legal.PropCompare("player.CardsLeftToReveal", "==", 1), legal.PlayerBoolIs("PlayerInactive", true))),
		},
	},
	{
		name:      "messageOverride",
		fixtures:  []string{"memoryZeroCardsLeft", "memoryCardNeverThere", "memoryCardAlreadyRevealed", "memoryVisibleOccupied", "memoryTargetPlayerOne", "memoryTargetObserver"},
		proposers: []int{0},
		specs: []legal.Spec{
			legal.PropAtLeast("player.CardsLeftToReveal", 1).WithMessage("gen.override"),
			legal.PropCompare("player.CardsLeftToReveal", ">", 0).WithMessage("gen.override"),
			legal.PropEquals("player.CardsLeftToReveal", "2").WithMessage("gen.override"),
			legal.PropNotEquals("player.CardsLeftToReveal", "0").WithMessage("gen.override"),
			legal.PlayerBool("SeatFilled").WithMessage("gen.override"),
			legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", true).WithMessage("gen.override"),
			legal.StackCount("game.VisibleCards", ">", 5).WithMessage("gen.override"),
			legal.StackEmpty("game.HiddenCards").WithMessage("gen.override"),
			legal.StackNotEmpty("game.UnusedCards").WithMessage("gen.override"),
			legal.ComponentPresentAt("game.HiddenCards", "move.CardIndex").WithMessage("gen.override"),
			legal.ComponentAbsentAt("game.HiddenCards", "move.CardIndex").WithMessage("gen.override"),
			legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex").WithMessage("gen.override"),
			legal.MayMoveTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex").WithMessage("gen.override"),
			legal.MayMoveToSameSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex").WithMessage("gen.override"),
			legal.MayMoveFirstTo("game.HiddenCards", "game.HiddenCards").WithMessage("gen.override"),
			legal.MayMoveAllTo("game.HiddenCards", "game.HiddenCards").WithMessage("gen.override"),
			legal.MayMoveFixedCountTo("game.HiddenCards", "game.HiddenCards", 2).WithMessage("gen.override"),
			legal.MaySwapComponents("game.HiddenCards", "move.CardIndex", "move.CardIndex").WithMessage("gen.override"),
			legal.ProposerIsCurrentPlayer().WithMessage("gen.override"),
			legal.ProposerIsPlayerFromMove("TargetPlayerIndex").WithMessage("gen.override"),
			legal.StackConstraints("HiddenCards", "HiddenCards").WithMessage("gen.override"),
			legal.InPhase(3).WithMessage("gen.override"),
			legal.ComponentPropEquals(legal.FirstOccupied("game.HiddenCards"), "Type", "anything").WithMessage("gen.override"),
			legal.ComponentPropsNotEqual(legal.FirstOccupied("game.HiddenCards"), legal.FirstOccupied("game.HiddenCards"), "Type").WithMessage("gen.override"),
			legal.PlayerHasSubmitted(legal.CurrentPlayer()),
			legal.PlayerHasNotSubmitted(legal.Proposer()),
			legal.PlayerIsActive(legal.PlayerFromMove("TargetPlayerIndex")),
			legal.PlayerIsInactive(legal.CurrentPlayer()),
			legal.PlayerSeatIsFilled(legal.Proposer()),
			legal.PlayerSeatIsClosed(legal.CurrentPlayer()),
			legal.PlayerIsAdmin(legal.Proposer()),
		},
	},
}

// constructorErrorSpecs are specs every catalog constructor must reject.
// With a nil chest (as the corpus harness resolves) propEquals' typo guard
// is off; plans.go covers that guard through a real manager.
var constructorErrorSpecs = []legal.Spec{
	sp("noSuchPredicate", "a"),
	sp("propAtLeast"),
	sp("propAtLeast", "player.X"),
	sp("propAtLeast", "player.X", "one"),
	sp("propAtLeast", "player.X", "1.5"),
	sp("propAtLeast", "player.X", " 1"),
	sp("propAtLeast", "player.X", ""),
	sp("propAtLeast", "player.X", "99999999999999999999"),
	sp("propAtLeast", "player.X", "1_000"),
	sp("propAtLeast", "player.X", "0x10"),
	sp("propCompare", "player.X", "=="),
	sp("propCompare", "player.X", "=", "1"),
	sp("propCompare", "player.X", "==", "x"),
	sp("playerBool"),
	sp("playerBool", "A", "B", "C"),
	sp("playerBool", "A", "yes"),
	sp("playerBool", "A", "True"),
	sp("playerBoolAt", "player.A"),
	sp("playerBoolAt", "game.A", "true"),
	sp("playerBoolAt", "players[*].A", "true"),
	sp("playerBoolAt", "player.A", "maybe"),
	sp("propEquals", "player.A"),
	sp("propNotEquals", "player.A", "1", "2"),
	sp("stackCount", "game.S", ">="),
	sp("stackCount", "game.S", "~", "1"),
	sp("stackCount", "game.S", ">=", "n"),
	sp("stackEmpty"),
	sp("stackNotEmpty", "a", "b"),
	sp("componentPresentAt", "game.S"),
	sp("componentAbsentAt", "game.S"),
	sp("componentPresentAtKey", "game.S"),
	sp("mayMoveTo", "game.A", "game.B"),
	sp("mayMoveToSlot", "game.A", "game.B", "move.I"),
	sp("mayMoveFirstTo", "game.A"),
	sp("mayMoveFirstToSlot", "game.A", "game.B"),
	sp("mayMoveAllTo", "game.A"),
	sp("mayMoveCountTo", "game.A", "game.B"),
	sp("mayMoveFixedCountTo", "game.A", "game.B"),
	sp("mayMoveFixedCountTo", "game.A", "game.B", "x"),
	sp("mayMoveFixedCountTo", "game.A", "game.B", "-1"),
	sp("maySwapComponents", "game.A", "move.I"),
	sp("maySwapComponentsByKey", "game.A", "move.I"),
	legal.Spec{Name: "allActivePlayers"},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("stackEmpty", "game.S")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("playerBool"), sp("playerBool", "A")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("playerBool")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("playerBool", "A", "nope")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propAtLeast", "game.X", "1")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propAtLeast", "player.", "1")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propAtLeast", "player.X")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propAtLeast", "player.X", "q")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propCompare", "player.X", "==")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propCompare", "move.X", "==", "1")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propCompare", "player.X", "=~", "1")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{sp("propCompare", "player.X", "==", "q")}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{{Name: "any", Sub: []legal.Spec{sp("playerBool", "A")}}}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{{Name: "any", Sub: []legal.Spec{sp("playerBool", "A"), sp("propAtLeast", "player.A", "1")}}}},
	legal.Spec{Name: "allActivePlayers", Sub: []legal.Spec{{Name: "any", Sub: []legal.Spec{sp("playerBool", "A"), {Name: "any", Sub: []legal.Spec{sp("playerBool", "A"), sp("playerBool", "B")}}}}}},
	sp("proposerIsCurrentPlayer", "x"),
	sp("proposerIsPlayerFromMove"),
	sp("proposerIsPlayerFromMove", ""),
	sp("proposerIsPlayerFromMove", "A", "B"),
	sp("revealableCardAt", "game.A", "game.B"),
	sp("componentPropEqualsCurrentPlayer", "game.A", "move.K"),
	sp("inPhase", "Playing"),
	sp("inPhase", "1", "x"),
	sp("stackConstraints", "A"),
	sp("componentPropEquals", "game.S", "first-occupied", "F"),
	sp("componentPropEquals", "", "first-occupied", "F", "v"),
	sp("componentPropEquals", "game.S", "last-occupied", "F", "v"),
	sp("componentPropEquals", "game.S", "occupied-ordinal:-1", "F", "v"),
	sp("componentPropEquals", "game.S", "occupied-ordinal:x", "F", "v"),
	sp("componentPropEquals", "game.S", "occupied-ordinal:", "F", "v"),
	sp("componentPropEquals", "game.S", "move-index:game.I", "F", "v"),
	sp("componentPropEquals", "game.S", "move-index:move.", "F", "v"),
	sp("componentPropEquals", "game.S", "first-occupied", "", "v"),
	sp("componentPropNotEquals", "game.S", "first-occupied", "F"),
	sp("componentPropsEqual", "game.S", "first-occupied", "game.T", "first-occupied"),
	sp("componentPropsEqual", "game.S", "nope", "game.T", "first-occupied", "F"),
	sp("componentPropsEqual", "game.S", "first-occupied", "game.T", "nope", "F"),
	sp("componentPropsEqual", "game.S", "first-occupied", "game.T", "first-occupied", ""),
	sp("componentPropsNotEqual", "game.S", "first-occupied", "", "first-occupied", "F"),
}

// vectorsFile is written column-wise to keep ~4000 vectors small: a vector is
// [source, spec index, fixture, proposer, verdict index], where the indexes
// point into the de-duplicated specs and verdicts tables. predicates is
// parallel to specs.
type vectorsFile struct {
	Specs             []legal.Spec
	Predicates        []exportedPredicate
	Verdicts          []exportedVerdict
	Vectors           [][]interface{}
	ConstructorErrors []exportedConstructorError
}

func buildVectors(corpusDir string, fixtures *fixtureCache) vectorsFile {
	out := vectorsFile{}
	specIndex := map[string]int{}
	verdictIndex := map[string]int{}
	record := func(source string, spec legal.Spec, fixtureName string, proposer int) {
		pred, err := resolveSpecViaRegistry(spec, legal.DefaultConstructors(), nil)
		must(err, "resolving "+specKey(spec))
		si, ok := specIndex[specKey(spec)]
		if !ok {
			si = len(out.Specs)
			specIndex[specKey(spec)] = si
			out.Specs = append(out.Specs, spec)
			out.Predicates = append(out.Predicates, exportPredicate(pred))
		}
		verdict := evaluateRaw(pred, fixtures.get(fixtureName).context(proposer))
		verdictKey := string(marshal(verdict))
		vi, ok := verdictIndex[verdictKey]
		if !ok {
			vi = len(out.Verdicts)
			verdictIndex[verdictKey] = vi
			out.Verdicts = append(out.Verdicts, verdict)
		}
		out.Vectors = append(out.Vectors, []interface{}{source, si, fixtureName, proposer, vi})
	}
	for _, v := range corpusVectors(corpusDir, fixtures) {
		record(v.Source, v.Spec, v.Fixture, v.Proposer)
	}
	for _, group := range extraGroups {
		for _, spec := range group.specs {
			for _, fixtureName := range group.fixtures {
				for _, proposer := range group.proposers {
					record("extra:"+group.name, spec, fixtureName, proposer)
				}
			}
		}
	}
	for _, spec := range constructorErrorSpecs {
		_, err := resolveSpecViaRegistry(spec, legal.DefaultConstructors(), nil)
		if err == nil {
			panic("gen: expected a constructor error for " + specKey(spec))
		}
		out.ConstructorErrors = append(out.ConstructorErrors, exportedConstructorError{Spec: spec, Error: err.Error()})
	}
	return out
}
