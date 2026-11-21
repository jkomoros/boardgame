package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	blackjackgame "github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	memorygame "github.com/jkomoros/boardgame/examples/memory"
	piggame "github.com/jkomoros/boardgame/examples/pig"
	tictactoegame "github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Plan vectors: whole legal plans, evaluated by the real plan engine
(GameManager.LegalEvaluateLedger / LegalEvaluatePlan / LegalRenderVerdict and
move.Legal itself).

The plan engine is unexported core code, reachable only through a move type
installed on a real GameManager. Two sources of plans are used:

  - The example games' own opted-in moves, exactly as they ship.
  - Synthetic moves added to a real example game by wrapping its delegate
    (synthDelegate below). They embed moves.Default / moves.CurrentPlayer,
    declare whatever preconditions a vector needs, and never apply. This is
    what lets a vector put the `any` compositor at the top of a plan, order
    several failing predicates against each other, bypass for admin,
    suppress a contributed check, or end in LegalCustom - none of which an
    example game happens to do in the right combination.

Every plan is recorded with its assembled spec list (rebuilt here from the
exported ContributedPreconditions/DeclaredPreconditions and checked, name by
name, against the ledger the real plan produced).
*/

// customModes maps a synthetic move's name to what its LegalCustom returns.
var customModes = map[string]string{}

func runCustom(name string) error {
	switch customModes[name] {
	case "nil":
		return nil
	case "plain":
		return errors.New("custom says no")
	case "structured":
		return legal.Errorf("gen.custom", map[string]boardgame.LegalBindingValue{
			"count": legal.Int(3),
			"who":   legal.String("nobody"),
			"flag":  legal.Bool(true),
		})
	case "structuredBare":
		return legal.Errorf("gen.custom_bare", nil)
	case "unregistered":
		return legal.Errorf("gen.not_in_table", map[string]boardgame.LegalBindingValue{"n": legal.Int(1)})
	case "panic":
		panic("custom blew up")
	default:
		panic("gen: no custom mode for " + name)
	}
}

type genCurrentPlayer struct{ moves.CurrentPlayer }

func (g *genCurrentPlayer) Apply(state boardgame.State) error { return nil }

type genDefault struct{ moves.Default }

func (g *genDefault) Apply(state boardgame.State) error { return nil }

type genCurrentPlayerCustom struct{ moves.CurrentPlayer }

func (g *genCurrentPlayerCustom) Apply(state boardgame.State) error { return nil }
func (g *genCurrentPlayerCustom) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return runCustom(g.Info().Name())
}

type genDefaultCustom struct{ moves.Default }

func (g *genDefaultCustom) Apply(state boardgame.State) error { return nil }
func (g *genDefaultCustom) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return runCustom(g.Info().Name())
}

// synthMove is one synthetic move type to add to a wrapped game.
type synthMove struct {
	name string
	// base is "default", "currentPlayer", "defaultCustom" or
	// "currentPlayerCustom".
	base       string
	customMode string
	options    []moves.CustomConfigurationOption
}

// policyOverride replaces the sanitization policy of one property.
type policyOverride func(groupMembership map[string]bool) boardgame.Policy

// synthDelegate wraps an example game's delegate. It must forward the
// optional interfaces core consumes by type assertion, since embedding the
// GameDelegate interface does not promote them.
type synthDelegate struct {
	boardgame.GameDelegate
	extraMoves     []synthMove
	extraTemplates map[string]string
	// policies is keyed "game.<Prop>" or "player.<Prop>".
	policies map[string]policyOverride
}

// Name must match this package's name (NewGameManager checks it).
func (s *synthDelegate) Name() string { return "main" }

func (s *synthDelegate) ConfigureLegalTemplates() map[string]string {
	out := map[string]string{}
	if tc, ok := s.GameDelegate.(legal.TemplateConfigurer); ok {
		for k, v := range tc.ConfigureLegalTemplates() {
			out[k] = v
		}
	}
	for k, v := range s.extraTemplates {
		out[k] = v
	}
	return out
}

func (s *synthDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {
	if cc, ok := s.GameDelegate.(legal.ConstructorConfigurer); ok {
		return cc.ConfigurePredicateConstructors()
	}
	return nil
}

// base.GameDelegate.CheckGameFinished asserts these three on the manager's
// delegate (the wrapper), not on itself.
type gameFinisher interface {
	GameEndConditionMet(state boardgame.ImmutableState) bool
	PlayerScore(pState boardgame.ImmutableSubState) int
	LowScoreWins() bool
}

func (s *synthDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	return s.GameDelegate.(gameFinisher).GameEndConditionMet(state)
}

func (s *synthDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	return s.GameDelegate.(gameFinisher).PlayerScore(pState)
}

func (s *synthDelegate) LowScoreWins() bool {
	return s.GameDelegate.(gameFinisher).LowScoreWins()
}

func (s *synthDelegate) SanitizationPolicy(prop boardgame.StatePropertyRef, groupMembership map[string]bool) boardgame.Policy {
	key := ""
	switch prop.Group {
	case boardgame.StateGroupGame:
		key = "game." + prop.PropName
	case boardgame.StateGroupPlayer:
		key = "player." + prop.PropName
	}
	if override, ok := s.policies[key]; ok {
		return override(groupMembership)
	}
	return s.GameDelegate.SanitizationPolicy(prop, groupMembership)
}

func (s *synthDelegate) ConfigureMoves() []boardgame.MoveConfig {
	configs := s.GameDelegate.ConfigureMoves()
	auto := moves.NewAutoConfigurer(s)
	for _, m := range s.extraMoves {
		options := append([]moves.CustomConfigurationOption{moves.WithMoveName(m.name)}, m.options...)
		var example moves.AutoConfigurableMove
		switch m.base {
		case "default":
			example = new(genDefault)
		case "currentPlayer":
			example = new(genCurrentPlayer)
		case "defaultCustom":
			example = new(genDefaultCustom)
		case "currentPlayerCustom":
			example = new(genCurrentPlayerCustom)
		default:
			panic("gen: unknown synthetic base " + m.base)
		}
		if m.customMode != "" {
			customModes[m.name] = m.customMode
		}
		configs = append(configs, auto.MustConfig(example, options...))
	}
	return configs
}

var synthTemplates = map[string]string{
	"gen.no_cards":         "you have {value} cards left but need {min}",
	"gen.any_override":     "neither alternative held",
	"gen.custom":           "custom rejected {count} for {who} ({flag})",
	"gen.custom_bare":      "custom rejected it",
	"gen.seat":             "the seat must be filled",
	"gen.everyone":         "everyone needs a card left",
	"custom says no":       "the custom check said no",
	"gen.unused_with_bind": "unused {value}",
}

func pre(specs ...legal.Spec) moves.CustomConfigurationOption {
	return moves.WithLegalPreconditions(specs...)
}

var synthMemoryMoves = []synthMove{
	{name: "Gen Interleaved", base: "currentPlayer", options: []moves.CustomConfigurationOption{pre(
		legal.PropEquals("move.TargetPlayerIndex", "0"),
		legal.PropAtLeast("player.CardsLeftToReveal", 1).WithMessage("gen.no_cards"),
		legal.StackEmpty("game.VisibleCards"),
		legal.PlayerBoolAt(legal.PlayerFromMove("TargetPlayerIndex"), "SeatFilled", false),
		legal.StackCount("game.HiddenCards", ">=", 20),
		legal.ComponentAbsentAt("game.VisibleCards", "player.CardsLeftToReveal"),
	)}},
	{name: "Gen Any", base: "currentPlayer", options: []moves.CustomConfigurationOption{pre(
		legal.Any(legal.PlayerBool("SeatFilled"), legal.PropAtLeast("player.CardsLeftToReveal", 2)),
		legal.Any(legal.StackEmpty("game.HiddenCards"), legal.StackNotEmpty("game.VisibleCards")).WithMessage("gen.any_override"),
		legal.Any(legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", true), legal.PlayerBool("SeatFilled")),
		legal.Any(legal.PropEquals("move.TargetPlayerIndex", "1"), legal.StackCount("game.VisibleCards", ">", 3), legal.PlayerBoolIs("PlayerInactive", true)),
	)}},
	{name: "Gen Any Default", base: "default", options: []moves.CustomConfigurationOption{pre(
		legal.Any(legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", true), legal.StackEmpty("game.HiddenCards")),
		legal.Any(legal.StackEmpty("game.HiddenCards"), legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", true)),
		legal.Any(legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", false), legal.PlayerBoolAt(legal.Proposer(), "PlayerInactive", true)),
		// Two unknown subs with different names: the reason names the first.
		legal.Any(legal.StackEmpty("game.HiddenCards"), legal.PropAtLeast("proposer.CardsLeftToReveal", 9), legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", true)),
	)}},
	{name: "Gen Stack Constraints", base: "default", options: []moves.CustomConfigurationOption{
		moves.WithSourceProperty("HiddenCards"), moves.WithDestinationProperty("VisibleCards"),
		pre(legal.StackNotEmpty("game.HiddenCards")),
	}},
	{name: "Gen Stack Constraints Same", base: "default", options: []moves.CustomConfigurationOption{
		moves.WithSourceProperty("HiddenCards"), moves.WithDestinationProperty("HiddenCards"),
		pre(legal.StackEmpty("game.HiddenCards"), legal.StackNotEmpty("game.VisibleCards")),
	}},
	{name: "Gen Suppressed", base: "default", options: []moves.CustomConfigurationOption{
		moves.WithSourceProperty("HiddenCards"), moves.WithDestinationProperty("HiddenCards"),
		moves.WithoutLegalPrecondition(moves.PreconditionStackConstraints),
		pre(legal.StackNotEmpty("game.VisibleCards")),
	}},
	{name: "Gen Suppressed Only", base: "default", options: []moves.CustomConfigurationOption{
		moves.WithSourceProperty("HiddenCards"), moves.WithDestinationProperty("HiddenCards"),
		moves.WithoutLegalPrecondition(moves.PreconditionStackConstraints),
	}},
	{name: "Gen Empty", base: "default", options: []moves.CustomConfigurationOption{pre()}},
	{name: "Gen Custom Nil", base: "currentPlayerCustom", customMode: "nil", options: []moves.CustomConfigurationOption{pre(legal.StackNotEmpty("game.HiddenCards"))}},
	{name: "Gen Custom Plain", base: "currentPlayerCustom", customMode: "plain", options: []moves.CustomConfigurationOption{pre(legal.StackNotEmpty("game.HiddenCards"))}},
	{name: "Gen Custom Structured", base: "currentPlayerCustom", customMode: "structured", options: []moves.CustomConfigurationOption{pre(legal.StackNotEmpty("game.HiddenCards"), legal.PropAtLeast("player.CardsLeftToReveal", 1))}},
	{name: "Gen Custom Bare", base: "defaultCustom", customMode: "structuredBare", options: []moves.CustomConfigurationOption{pre(legal.StackEmpty("game.VisibleCards"))}},
	{name: "Gen Custom Unregistered", base: "defaultCustom", customMode: "unregistered"},
	{name: "Gen Custom Panic", base: "defaultCustom", customMode: "panic", options: []moves.CustomConfigurationOption{pre(legal.StackNotEmpty("game.HiddenCards"))}},
	{name: "Gen Custom Only", base: "defaultCustom", customMode: "plain"},
	{name: "Gen Admin", base: "currentPlayer", options: []moves.CustomConfigurationOption{pre(
		legal.ProposerIsPlayerFromMove("TargetPlayerIndex"),
		legal.PlayerSeatIsFilled(legal.Proposer()).WithMessage("gen.seat"),
		legal.PlayerIsActive(legal.CurrentPlayer()),
		legal.PlayerIsActive(legal.PlayerFromMove("TargetPlayerIndex")),
	)}},
	{name: "Gen Proposer Unknown", base: "default", options: []moves.CustomConfigurationOption{pre(
		legal.PlayerBoolAt(legal.Proposer(), "SeatFilled", false),
		legal.StackEmpty("game.HiddenCards"),
	)}},
	{name: "Gen Not Client Evaluable", base: "default", options: []moves.CustomConfigurationOption{pre(
		legal.AllActivePlayers(legal.PropAtLeast("player.CardsLeftToReveal", 1)).WithMessage("gen.everyone"),
		legal.MayMoveAllTo("game.HiddenCards", "game.VisibleCards"),
		legal.MayMoveFixedCountTo("game.HiddenCards", "game.VisibleCards", 21),
		legal.MayMoveCountTo("game.VisibleCards", "game.HiddenCards", "player.CardsLeftToReveal"),
		legal.ComponentPropsEqual(legal.FirstOccupied("game.HiddenCards"), legal.OccupiedOrdinal("game.HiddenCards", 1), "Type"),
		legal.ComponentPropEquals(legal.FirstOccupied("game.VisibleCards"), "Type", "anything"),
	)}},
	{name: "Gen Hidden Reads", base: "currentPlayer", options: []moves.CustomConfigurationOption{pre(
		legal.StackCount("game.HiddenCards", "==", 5),
		legal.StackEmpty("game.HiddenCards"),
		legal.ComponentAbsentAt("game.HiddenCards", "player.CardsLeftToReveal"),
		legal.MayMoveFirstTo("game.VisibleCards", "game.HiddenCards"),
		legal.PropAtLeast("player.CardsLeftToReveal", 5),
		legal.PropAtLeast("proposer.CardsLeftToReveal", 5),
		legal.PropAtLeast("players[move.TargetPlayerIndex].CardsLeftToReveal", 5),
		legal.PropEquals("move.TargetPlayerIndex", "1"),
		legal.PropAtLeast("game.NumCards", 100),
		legal.StackNotEmpty("player.WonCards"),
	)}},
}

var synthCheckersMoves = []synthMove{
	{name: "Gen Phase Playing", base: "currentPlayer", options: []moves.CustomConfigurationOption{
		moves.WithLegalPhases(enum.EnumKey(1)),
		pre(legal.PropEquals("player.Color", "Black"), legal.ComponentPropEquals(legal.FirstOccupied("game.Spaces"), "Color", "Red")),
	}},
	{name: "Gen Phase Setup", base: "currentPlayer", options: []moves.CustomConfigurationOption{
		moves.WithLegalPhases(enum.EnumKey(0)),
		pre(legal.PropEquals("player.Color", "Black"), legal.StackEmpty("game.UnusedTokens")),
	}},
	{name: "Gen Phase Either", base: "default", options: []moves.CustomConfigurationOption{
		moves.WithLegalPhases(enum.EnumKey(0), enum.EnumKey(1)),
		moves.WithSourceProperty("Spaces"), moves.WithDestinationProperty("UnusedTokens"),
		pre(legal.PropEquals("game.Phase", "Playing"), legal.PropNotEquals("proposer.Color", "Red")),
	}},
	{name: "Gen Phase Custom", base: "currentPlayerCustom", customMode: "structured", options: []moves.CustomConfigurationOption{
		moves.WithLegalPhases(enum.EnumKey(1)),
		pre(legal.StackCount("game.Spaces", "==", 24)),
	}},
}

type exportedEntryStatic struct {
	Name            string         `json:"name"`
	Args            []string       `json:"args"`
	Serializable    bool           `json:"serializable"`
	ClientEvaluable bool           `json:"clientEvaluable"`
	FieldDependent  bool           `json:"fieldDependent"`
	Reads           []exportedRead `json:"reads"`
}

type exportedPlanMove struct {
	Game         string       `json:"game"`
	Name         string       `json:"name"`
	Contributed  []legal.Spec `json:"contributed"`
	Authored     []legal.Spec `json:"authored"`
	Suppressions []string     `json:"suppressions"`
	// Specs is contributed minus suppressed, then authored: the plan.
	Specs      []legal.Spec `json:"specs"`
	HasCustom  bool         `json:"hasCustom"`
	CustomMode string       `json:"customMode,omitempty"`
	// Entries is the static half of the ledger: one per plan predicate, then
	// "custom" when HasCustom.
	Entries []exportedEntryStatic `json:"entries"`
	// GameRegistered names predicates this game registers itself; the
	// TypeScript side stubs them with the recorded verdicts.
	GameRegistered []string `json:"gameRegistered,omitempty"`
}

type exportedPrecondition struct {
	Name        string           `json:"name"`
	Args        []string         `json:"args,omitempty"`
	Verdict     string           `json:"verdict"`
	Message     *exportedMessage `json:"message,omitempty"`
	Evaluable   bool             `json:"evaluable"`
	Provisional bool             `json:"provisional,omitempty"`
}

// A scenario is written as one row of indexes into de-duplicated tables,
// because ~1800 scenarios repeat the same few hundred verdicts and ledger
// lines. The row is:
//
//	[move, fixture, moveArgs, proposer, overall, [verdicts], [preconditions],
//	 rendered, hotPath, legalError]
//
// move indexes moves; fixture, moveArgs, overall/verdicts, preconditions and
// the three texts index fixtureNames, moveArgs, verdicts, preconditions and
// texts. verdicts is the dynamic half of the ledger, parallel to the move's
// entries. rendered is GameManager.LegalRenderVerdict(overall); hotPath is
// LegalEvaluatePlan's error text; legalError is move.Legal(state, proposer)'s
// (-1 when the scenario has no move to call it on). preconditions is the wire
// ledger for viewer == proposer, built the way server/api's
// buildPreconditionEntry builds it.
type scenarioTables struct {
	fixtureNames  []string
	moveArgs      []*exportedMove
	verdicts      []exportedVerdict
	preconditions []exportedPrecondition
	texts         []string
	index         map[string]int
}

func (t *scenarioTables) intern(kind string, value interface{}, add func() int) int {
	key := kind + string(marshal(value))
	if i, ok := t.index[key]; ok {
		return i
	}
	i := add()
	t.index[key] = i
	return i
}

func (t *scenarioTables) fixture(name string) int {
	return t.intern("f", name, func() int { t.fixtureNames = append(t.fixtureNames, name); return len(t.fixtureNames) - 1 })
}

func (t *scenarioTables) move(m *exportedMove) int {
	return t.intern("m", m, func() int { t.moveArgs = append(t.moveArgs, m); return len(t.moveArgs) - 1 })
}

func (t *scenarioTables) verdict(v exportedVerdict) int {
	return t.intern("v", v, func() int { t.verdicts = append(t.verdicts, v); return len(t.verdicts) - 1 })
}

func (t *scenarioTables) precondition(p exportedPrecondition) int {
	return t.intern("p", p, func() int { t.preconditions = append(t.preconditions, p); return len(t.preconditions) - 1 })
}

func (t *scenarioTables) text(s string) int {
	return t.intern("t", s, func() int { t.texts = append(t.texts, s); return len(t.texts) - 1 })
}

type exportedAssemblyError struct {
	Name         string            `json:"name"`
	Base         string            `json:"base"`
	Authored     []legal.Spec      `json:"authored"`
	Suppressions []string          `json:"suppressions"`
	Templates    map[string]string `json:"templates"`
	// Error is NewGameManager's error with its fixed "Failed to assemble
	// declarative legality plans: " prefix removed.
	Error string `json:"error"`
}

type plansFile struct {
	fixtures       map[string]exportedFixture
	games          map[string]exportedGame
	moves          []exportedPlanMove
	tables         *scenarioTables
	scenarios      [][]interface{}
	views          map[string]map[string]json.RawMessage
	assemblyErrors []exportedAssemblyError
}

func (p plansFile) sections() []section {
	return []section{
		{key: "comment", scalar: generatedComment},
		{key: "shape", scalar: "scenarios[i] = [move, fixture, moveArgs, proposer, overall, [verdicts], [preconditions], rendered, hotPath, legalError]; every element but proposer indexes the table of the same name (overall and verdicts index verdicts; rendered, hotPath and legalError index texts, legalError -1 when there is no move)"},
		{key: "moves", list: anyList(p.moves)},
		{key: "views", object: anyMap(p.views)},
		{key: "assemblyErrors", list: anyList(p.assemblyErrors)},
		{key: "fixtureNames", list: anyList(p.tables.fixtureNames)},
		{key: "moveArgs", list: anyList(p.tables.moveArgs)},
		{key: "verdicts", list: anyList(p.tables.verdicts)},
		{key: "preconditions", list: anyList(p.tables.preconditions)},
		{key: "texts", list: anyList(p.tables.texts)},
		{key: "scenarios", list: anyList(p.scenarios)},
	}
}

func containsInt(list []int, want int) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type contributor interface {
	ContributedPreconditions() []legal.Spec
}

type declarer interface {
	DeclaredPreconditions() ([]legal.Spec, []string)
}

func nonNilSpecs(in []legal.Spec) []legal.Spec {
	if in == nil {
		return []legal.Spec{}
	}
	return in
}

// viewVisibility returns the Visibility block of the state JSON the server
// would send viewer.
func viewVisibility(game *boardgame.Game, state boardgame.ImmutableState, viewer int) json.RawMessage {
	blob, err := game.JSONForPlayer(boardgame.PlayerIndex(viewer), state)
	must(err, "JSONForPlayer")
	data, err := json.Marshal(blob.(map[string]interface{})["CurrentState"])
	must(err, "marshaling viewer state")
	var decoded struct {
		Visibility json.RawMessage
	}
	must(json.Unmarshal(data, &decoded), "decoding viewer state")
	return decoded.Visibility
}

// planGame is one manager + game whose opted-in moves are recorded.
type planGame struct {
	key string
	// delegate builds a fresh delegate: one can only ever back one manager.
	delegate func() boardgame.GameDelegate
	// states are named arrangements of a fresh game.
	states map[string]func(game *boardgame.Game, state boardgame.State)
	// onlySynthetic limits recording to moves whose name starts with "Gen ".
	onlySynthetic bool
	proposers     []int
}

func memoryPlanStates() map[string]func(*boardgame.Game, boardgame.State) {
	return map[string]func(*boardgame.Game, boardgame.State){
		"fresh": func(*boardgame.Game, boardgame.State) {},
		"zeroCardsRevealed": func(game *boardgame.Game, state boardgame.State) {
			must(state.CurrentPlayer().ReadSetter().SetIntProp("CardsLeftToReveal", 0), "setting CardsLeftToReveal")
			must(gameStack(state, "HiddenCards").ComponentAt(0).MoveTo(gameStack(state, "VisibleCards"), 0), "revealing")
			must(gameStack(state, "HiddenCards").ComponentAt(3).MoveTo(gameStack(state, "VisibleCards"), 3), "revealing")
		},
		"seatedPlayerOneTurn": func(game *boardgame.Game, state boardgame.State) {
			must(state.GameState().ReadSetter().SetPlayerIndexProp("CurrentPlayer", 1), "setting CurrentPlayer")
			must(state.PlayerStates()[1].ReadSetter().SetBoolProp("SeatFilled", true), "setting SeatFilled")
			must(state.PlayerStates()[0].ReadSetter().SetIntProp("CardsLeftToReveal", 0), "setting CardsLeftToReveal")
		},
		"allRevealed": func(game *boardgame.Game, state boardgame.State) {
			hidden, visible := gameStack(state, "HiddenCards"), gameStack(state, "VisibleCards")
			for i := 0; i < hidden.Len(); i++ {
				if c := hidden.ComponentAt(i); c != nil {
					must(c.MoveTo(visible, i), "revealing a card")
				}
			}
			must(state.PlayerStates()[1].ReadSetter().SetBoolProp("PlayerInactive", true), "setting PlayerInactive")
		},
	}
}

func freshOnly() map[string]func(*boardgame.Game, boardgame.State) {
	return map[string]func(*boardgame.Game, boardgame.State){"fresh": func(*boardgame.Game, boardgame.State) {}}
}

func buildPlans() plansFile {
	out := plansFile{
		fixtures: map[string]exportedFixture{},
		games:    map[string]exportedGame{},
		views:    map[string]map[string]json.RawMessage{},
		tables:   &scenarioTables{index: map[string]int{}},
	}
	hiddenFromOthers := func(groups map[string]bool) boardgame.Policy {
		if groups["self"] {
			return boardgame.PolicyVisible
		}
		return boardgame.PolicyHidden
	}
	games := []planGame{
		{key: "memory", delegate: memorygame.NewDelegate, states: memoryPlanStates(), proposers: []int{0, 1, -1, -2}},
		{key: "checkers", delegate: checkers.NewDelegate, states: freshOnly(), proposers: []int{0, 1, -1, -2}},
		{key: "blackjack", delegate: blackjackgame.NewDelegate, states: freshOnly(), proposers: []int{0, 1, -1, -2}},
		{key: "pig", delegate: piggame.NewDelegate, states: freshOnly(), proposers: []int{0, 1, -1, -2}},
		{key: "tictactoe", delegate: tictactoegame.NewDelegate, states: freshOnly(), proposers: []int{0, 1, -1, -2}},
		{
			key: "synthMemory",
			delegate: func() boardgame.GameDelegate {
				return &synthDelegate{GameDelegate: memorygame.NewDelegate(), extraMoves: synthMemoryMoves, extraTemplates: synthTemplates}
			},
			states:        memoryPlanStates(),
			onlySynthetic: true,
			proposers:     []int{0, 1, -1, -2, -3},
		},
		{
			// The same synthetic plans, on a game whose sanitization hides
			// things: this is what makes the per-viewer ledger interesting.
			key: "synthMemoryHidden",
			delegate: func() boardgame.GameDelegate {
				return &synthDelegate{
					GameDelegate:   memorygame.NewDelegate(),
					extraMoves:     synthMemoryMoves,
					extraTemplates: synthTemplates,
					policies: map[string]policyOverride{
						"game.HiddenCards":         func(map[string]bool) boardgame.Policy { return boardgame.PolicyLen },
						"game.VisibleCards":        func(map[string]bool) boardgame.Policy { return boardgame.PolicyNonEmpty },
						"game.NumCards":            func(map[string]bool) boardgame.Policy { return boardgame.PolicyHidden },
						"player.CardsLeftToReveal": hiddenFromOthers,
						"player.WonCards":          func(map[string]bool) boardgame.Policy { return boardgame.PolicyOrder },
					},
				}
			},
			states:        map[string]func(*boardgame.Game, boardgame.State){"fresh": memoryPlanStates()["fresh"], "seatedPlayerOneTurn": memoryPlanStates()["seatedPlayerOneTurn"]},
			onlySynthetic: true,
			proposers:     []int{0, 1, -1, -2},
		},
		{
			key: "synthCheckers",
			delegate: func() boardgame.GameDelegate {
				return &synthDelegate{GameDelegate: checkers.NewDelegate(), extraMoves: synthCheckersMoves, extraTemplates: synthTemplates}
			},
			states:        freshOnly(),
			onlySynthetic: true,
			proposers:     []int{0, 1, -1, -2},
		},
	}

	defaultNames := map[string]bool{"any": true, "inProgression": true}
	for _, c := range legal.DefaultConstructors() {
		defaultNames[c.Name] = true
	}

	for _, pg := range games {
		stateNames := sortedKeys(pg.states)
		var manager *boardgame.GameManager
		moveIndex := map[string]int{}
		for _, stateName := range stateNames {
			game, state := newGame(pg.key, pg.delegate())
			manager = game.Manager()
			pg.states[stateName](game, state)
			fixtureName := "plan/" + pg.key + "/" + stateName
			out.fixtures[fixtureName] = exportFixtureWith(pg.key, state, nil)
			numPlayers := len(state.ImmutablePlayerStates())
			// Every proposer is also a viewer: the server builds the ledger
			// for the requesting player.
			viewers := append([]int{}, pg.proposers...)
			for i := 0; i < numPlayers; i++ {
				if !containsInt(viewers, i) {
					viewers = append(viewers, i)
				}
			}
			out.views[fixtureName] = map[string]json.RawMessage{}
			for _, viewer := range viewers {
				out.views[fixtureName][fmt.Sprint(viewer)] = viewVisibility(game, state, viewer)
			}

			var moveNames []string
			for _, mv := range manager.ExampleMoves() {
				name := mv.Info().Name()
				if !manager.LegalPlanAssembled(name) {
					continue
				}
				if pg.onlySynthetic && !strings.HasPrefix(name, "Gen ") {
					continue
				}
				moveNames = append(moveNames, name)
			}
			sort.Strings(moveNames)

			for _, name := range moveNames {
				// A fresh move bound to this state, the way the server's move
				// forms build one.
				variants := []boardgame.Move{game.MoveByName(name), nil}
				// For moves with a TargetPlayerIndex, also target the other
				// seat and the observer.
				if _, ok := game.MoveByName(name).Reader().Props()["TargetPlayerIndex"]; ok {
					for _, target := range []boardgame.PlayerIndex{1, boardgame.ObserverPlayerIndex} {
						mv := game.MoveByName(name)
						must(mv.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", target), "setting TargetPlayerIndex")
						variants = append(variants, mv)
					}
				}
				for _, move := range variants {
					for _, proposer := range pg.proposers {
						verdict, entries, ok := manager.LegalEvaluateLedger(name, state, move, boardgame.PlayerIndex(proposer))
						if !ok {
							panic("gen: no plan for " + name)
						}

						idx, seen := moveIndex[name]
						if !seen {
							example := game.MoveByName(name)
							pm := exportedPlanMove{Game: pg.key, Name: name, CustomMode: customModes[name], Suppressions: []string{}}
							var contributed, authored []legal.Spec
							if c, ok := example.(contributor); ok {
								contributed = c.ContributedPreconditions()
							}
							var suppressions []string
							if d, ok := example.(declarer); ok {
								authored, suppressions = d.DeclaredPreconditions()
							}
							suppressed := map[string]bool{}
							for _, s := range suppressions {
								suppressed[s] = true
								pm.Suppressions = append(pm.Suppressions, s)
							}
							pm.Contributed, pm.Authored = nonNilSpecs(contributed), nonNilSpecs(authored)
							pm.Specs = []legal.Spec{}
							for _, s := range contributed {
								if !suppressed[s.Name] {
									pm.Specs = append(pm.Specs, s)
								}
							}
							pm.Specs = append(pm.Specs, authored...)
							_, pm.HasCustom = example.(boardgame.CustomLegaler)
							wantEntries := len(pm.Specs)
							if pm.HasCustom {
								wantEntries++
							}
							if len(entries) != wantEntries {
								panic(fmt.Sprintf("gen: %s/%s: rebuilt %d specs but the plan has %d entries", pg.key, name, wantEntries, len(entries)))
							}
							for i, e := range entries {
								if i < len(pm.Specs) && pm.Specs[i].Name != e.Name {
									panic(fmt.Sprintf("gen: %s/%s: rebuilt spec %d is %q but the plan's entry is %q", pg.key, name, i, pm.Specs[i].Name, e.Name))
								}
								args := e.Args
								if args == nil {
									args = []string{}
								}
								pm.Entries = append(pm.Entries, exportedEntryStatic{
									Name: e.Name, Args: args, Serializable: e.Serializable,
									ClientEvaluable: e.ClientEvaluable, FieldDependent: e.FieldDependent, Reads: exportReads(e.Reads),
								})
								if !defaultNames[e.Name] && e.Name != "custom" {
									pm.GameRegistered = append(pm.GameRegistered, e.Name)
								}
							}
							idx = len(out.moves)
							moveIndex[name] = idx
							out.moves = append(out.moves, pm)
						}

						_, hotErr := manager.LegalEvaluatePlan(name, state, move, boardgame.PlayerIndex(proposer))
						legalError := -1
						if move != nil {
							legalError = out.tables.text(errText(move.Legal(state, boardgame.PlayerIndex(proposer))))
						}
						verdicts, preconditions := []int{}, []int{}
						for _, e := range entries {
							verdicts = append(verdicts, out.tables.verdict(exportVerdict(e.Verdict)))
							// server/api buildPreconditionEntry, restated.
							evaluable := e.ClientEvaluable && e.Serializable && boardgame.LegalReadsEvaluable(state, boardgame.PlayerIndex(proposer), e.Reads)
							p := exportedPrecondition{Name: e.Name, Args: e.Args, Verdict: outcomeString(e.Verdict.Outcome), Evaluable: evaluable, Provisional: e.FieldDependent}
							if msg := e.Verdict.Message; msg != nil {
								p.Message = &exportedMessage{Template: msg.Template}
								if evaluable {
									p.Message.Bindings = msg.Bindings
								}
							}
							preconditions = append(preconditions, out.tables.precondition(p))
						}
						scenario := []interface{}{
							idx, out.tables.fixture(fixtureName), out.tables.move(exportMove(move)), proposer,
							out.tables.verdict(exportVerdict(verdict)), verdicts, preconditions,
							out.tables.text(manager.LegalRenderVerdict(verdict)), out.tables.text(errText(hotErr)), legalError,
						}
						out.scenarios = append(out.scenarios, scenario)
					}
				}
			}
		}
		out.games[pg.key] = exportGame(pg.key, manager)
	}

	out.assemblyErrors = buildAssemblyErrors()
	return out
}

// assemblyErrorCases are move configurations NewGameManager must reject.
var assemblyErrorCases = []struct {
	name         string
	base         string
	authored     []legal.Spec
	suppressions []moves.PreconditionName
	templates    map[string]string
}{
	{name: "unknown predicate", base: "default", authored: []legal.Spec{{Name: "noSuchPredicate", Args: []string{"a"}}}},
	{name: "constructor rejects args", base: "default", authored: []legal.Spec{{Name: "propAtLeast", Args: []string{"player.CardsLeftToReveal", "lots"}}}},
	{name: "any with one sub", base: "default", authored: []legal.Spec{legal.Any(legal.StackEmpty("game.HiddenCards"))}},
	{name: "nested any", base: "default", authored: []legal.Spec{legal.Any(legal.StackEmpty("game.HiddenCards"), legal.Any(legal.StackEmpty("game.VisibleCards"), legal.StackEmpty("game.UnusedCards")))}},
	{name: "any with a bad sub", base: "default", authored: []legal.Spec{legal.Any(legal.StackEmpty("game.HiddenCards"), legal.Spec{Name: "stackEmpty"})}},
	{name: "any with conflicting types", base: "default", authored: []legal.Spec{legal.Any(legal.PropAtLeast("player.SeatFilled", 1), legal.PlayerBool("SeatFilled"))}},
	{name: "missing game property", base: "default", authored: []legal.Spec{legal.StackEmpty("game.NoSuchStack")}},
	{name: "missing player property", base: "default", authored: []legal.Spec{legal.PropAtLeast("player.NoSuchProp", 1)}},
	{name: "missing proposer property", base: "default", authored: []legal.Spec{legal.PropAtLeast("proposer.NoSuchProp", 1)}},
	{name: "missing move field", base: "currentPlayer", authored: []legal.Spec{legal.ComponentPresentAt("game.HiddenCards", "move.NoSuchField")}},
	{name: "missing players[move] field", base: "currentPlayer", authored: []legal.Spec{legal.PropAtLeast("players[move.NoSuchField].CardsLeftToReveal", 1)}},
	{name: "malformed path", base: "default", authored: []legal.Spec{legal.PropAtLeast("nodothere", 1)}},
	{name: "unknown path kind", base: "default", authored: []legal.Spec{legal.PropAtLeast("nokind.X", 1)}},
	{name: "quantifier path on a leaf", base: "default", authored: []legal.Spec{legal.PropAtLeast("players[*].NoSuchProp", 1)}},
	{name: "wrong type int", base: "default", authored: []legal.Spec{legal.PropAtLeast("player.SeatFilled", 1)}},
	{name: "wrong type stack", base: "default", authored: []legal.Spec{legal.StackEmpty("game.NumCards")}},
	{name: "wrong type allowed set", base: "default", authored: []legal.Spec{legal.PropEquals("game.HiddenCards", "0")}},
	{name: "wrong type move field", base: "currentPlayer", authored: []legal.Spec{legal.ComponentPresentAt("game.HiddenCards", "move.TargetPlayerIndex")}},
	{name: "indexed path on a non-board", base: "currentPlayer", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards[move.TargetPlayerIndex]")}},
	{name: "propEquals typo guard", base: "default", authored: []legal.Spec{legal.PropEquals("player.CardsLeftToReveal", "Blakc")}},
	{name: "unregistered message override", base: "default", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards").WithMessage("gen.not_registered")}},
	{name: "unregistered any override", base: "default", authored: []legal.Spec{legal.Any(legal.StackEmpty("game.HiddenCards"), legal.StackEmpty("game.VisibleCards")).WithMessage("gen.not_registered")}},
	{name: "placeholder the predicate never binds", base: "default", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards").WithMessage("gen.unused_with_bind")}},
	{name: "placeholder lost when branches collapse", base: "currentPlayer", authored: []legal.Spec{legal.MayMoveFirstTo("game.HiddenCards", "game.VisibleCards").WithMessage("gen.unused_with_bind")}},
	{name: "default body overridden with an unbound placeholder", base: "default", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards")}, templates: map[string]string{"legal.stack_empty": "the stack has {count} left"}},
	{name: "admin bypass without proposer dependence", base: "default", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards").WithAdminBypass()}},
	{name: "admin bypass reading the current player", base: "default", authored: []legal.Spec{legal.PlayerBoolAt(legal.CurrentPlayer(), "SeatFilled", true).WithAdminBypass()}},
	{name: "unknown admin policy", base: "default", authored: []legal.Spec{{Name: "stackEmpty", Args: []string{"game.HiddenCards"}, AdminPolicy: "sometimes"}}},
	{name: "unmatched suppression with none contributed", base: "default", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards")}, suppressions: []moves.PreconditionName{moves.PreconditionInPhase}},
	{name: "unmatched suppression", base: "currentPlayer", authored: []legal.Spec{legal.StackEmpty("game.HiddenCards")}, suppressions: []moves.PreconditionName{"inphase"}},
}

func buildAssemblyErrors() []exportedAssemblyError {
	var out []exportedAssemblyError
	for _, c := range assemblyErrorCases {
		options := []moves.CustomConfigurationOption{moves.WithLegalPreconditions(c.authored...)}
		suppressions := []string{}
		for _, s := range c.suppressions {
			options = append(options, moves.WithoutLegalPrecondition(s))
			suppressions = append(suppressions, string(s))
		}
		templates := map[string]string{}
		for k, v := range synthTemplates {
			templates[k] = v
		}
		for k, v := range c.templates {
			templates[k] = v
		}
		delegate := &synthDelegate{
			GameDelegate:   memorygame.NewDelegate(),
			extraMoves:     []synthMove{{name: "Gen Bad", base: c.base, options: options}},
			extraTemplates: templates,
		}
		_, err := boardgame.NewGameManager(delegate, storagememory.NewStorageManager())
		if err == nil {
			panic("gen: expected NewGameManager to reject: " + c.name)
		}
		const prefix = "Failed to assemble declarative legality plans: "
		text := err.Error()
		at := strings.Index(text, prefix)
		if at < 0 {
			panic("gen: " + c.name + ": unexpected boot error: " + text)
		}
		out = append(out, exportedAssemblyError{
			Name: c.name, Base: c.base, Authored: c.authored, Suppressions: suppressions,
			Templates: templates, Error: text[at+len(prefix):],
		})
	}
	return out
}
