/*
Package gamedef is the exported definition of a game: one JSON document per
game, GameData, holding everything about the game that is already data, so
that the TypeScript engine can load it (server/static/src/engine/rules/data.ts
mirrors every type here, key for key). See fabric/TRANSPILE-PROPOSAL.md, "The
exported definition".

The exporter has two halves that are joined. The run-time half (Export,
export.go) reads the live GameManager. The static half (Static, static.go)
reads the game's package with go/types, and knows what reflection cannot:
which methods a move type declares itself rather than promotes from a stock
move, which behaviors the state structs embed, and every function that
reaches engine output (FunctionEntry).

# JSON shape

Keys are the Go field names, capitalized like the state JSON and the
existing GameDefinition. A field tagged omitempty is optional in the
TypeScript type (`Name?:`); a field without it is always present, and a nil
pointer, slice or map in such a field is `null`. Only GameData.Constants,
GameData.Variants, MoveData.Stock, MoveData.LegalPhases,
MoveData.Progression and DeckDef.GenericValues may be null; the exporter
writes every other map and slice non-nil, so it serializes as `{}` or `[]`
when empty (Moves, Functions, Computed, Predicates, Agents, Fields,
Plan.Specs, Stock.Config, FunctionEntry.Calls, FunctionEntry.Refusals,
Initial.Game, Initial.Players, Initial.Dynamic, LegalTemplates, ...).

# Determinism

The document is a pure function of the game's Go source and live manager:

  - every map is a Go map, and encoding/json writes map keys sorted, so no
    map order leaks into the output;
  - Moves are in registration order (GameManager.Moves()), the order the
    engine tries fix-ups in;
  - Functions are sorted by Func; FunctionEntry.Calls are sorted and
    deduplicated; Computed and Predicates are in the order the delegate
    declares them; Agents in the delegate's order;
  - no wall-clock time, host name, absolute path or commit appears anywhere
    (File fields are relative to the package directory).

Stream 1's `gamedef export --check` relies on this: a checked-in definition
is current exactly when re-exporting reproduces its bytes.
*/
package gamedef

import (
	"encoding/json"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
)

// SchemaVersion is GameData.SchemaVersion. Bump it whenever GameData changes
// shape, and update rules/data.ts together.
const SchemaVersion = 1

// Exporter names the producer in GameData.Source.Exporter. It changes when
// the exporter's output for an unchanged game would change.
const Exporter = "gamedef@1"

// FnRef names a function by its Go identity: "blackjack.(*moveCurrentPlayerHit).Apply"
// for a method with a pointer receiver, "blackjack.(moveX).M" for a value
// receiver, "blackjack.handValue" for a function. The package part is the
// package name, not the import path (GameData.Source.Package has that).
// The empty FnRef means "none" and is omitted from the JSON.
type FnRef = string

// GameData is the whole exported definition of one game.
type GameData struct {
	SchemaVersion int
	Source        Source

	Name              string
	DisplayName       string
	Description       string
	MinNumPlayers     int
	MaxNumPlayers     int
	DefaultNumPlayers int

	// Definition is exactly what testdata/gen/vecgame DefinitionFor emits
	// today, plus GenericValues, ComponentProps, Range and Constraints.
	Definition Definition
	// Constants is the chest's constants; null when there are none.
	Constants map[string]any
	// ComponentEnums are enums component values use that the chest does not
	// ship (GameSpec.componentEnums).
	ComponentEnums map[string]EnumDef
	// PhaseEnum and GroupEnum name the chest enums for phases and
	// sanitization groups; "" when there is none.
	PhaseEnum string
	GroupEnum string

	// Initial is the constructors' output: non-zero starting values.
	Initial Initial
	Traits  Traits
	// Variants is the conformance header's variantConfig, verbatim.
	Variants json.RawMessage
	// Computed is the computed properties, in declaration order.
	Computed []ComputedData
	// LegalTemplates is ConfigureLegalTemplates: template key to text.
	LegalTemplates map[string]string
	// Predicates are the custom legal predicate constructors.
	Predicates []PredicateData
	// ComponentLegal is deck name to the Legal method
	// (interfaces.LegalComponent) of that deck's component values type,
	// which moves.DefaultComponent calls; role RoleComponent. Omitted when
	// no deck's values have one. The static half fills it.
	ComponentLegal map[string]FnRef `json:",omitempty"`
	// Moves are in registration order.
	Moves []MoveData
	SetUp SetUpData
	End   EndData
	// CurrentPlayerIndex is the delegate's CurrentPlayerIndex override, if
	// the game declares one.
	CurrentPlayerIndex FnRef `json:",omitempty"`
	// Agents lists the game's agents. The engine does not run them.
	Agents []AgentData
	// Functions is every game-written function that reaches engine output,
	// with the translator's verdict, sorted by Func.
	Functions []FunctionEntry
}

// Source says where a GameData came from.
type Source struct {
	// Package is the game package's import path.
	Package string
	// Exporter is the producer, Exporter at the time of export.
	Exporter string
}

// Initial is what the state constructors set: only non-zero values, read
// through each sub-state's generated property reader.
type Initial struct {
	// Game is GameStateConstructor()'s non-zero properties.
	Game map[string]any
	// Players has one entry per player index 0 .. MaxNumPlayers-1:
	// PlayerStateConstructor(i)'s non-zero properties.
	Players []map[string]any
	// Dynamic is deck name to one entry per component, by deck index:
	// DynamicComponentValuesConstructor's non-zero properties.
	Dynamic map[string][]map[string]any
}

// Traits are what the embedded behaviors and the interface methods a game
// declares contribute (GameTraits and PlayerTraits in rules/types.ts).
type Traits struct {
	Game   GameTraits
	Player PlayerTraits
}

// GameTraits are the game state's behaviors.
type GameTraits struct {
	// DrawDiscardPair is behaviors.DrawDiscardPair's two stack names.
	DrawDiscardPair *DrawDiscardPair `json:",omitempty"`
	// RoundRobin is behaviors.RoundRobin embedded.
	RoundRobin bool `json:",omitempty"`
	// PlayerOrder is a declared interfaces.PlayerOrderer method.
	PlayerOrder FnRef `json:",omitempty"`
}

// DrawDiscardPair names behaviors.DrawDiscardPair's two stack properties.
type DrawDiscardPair struct {
	Draw    string
	Discard string
}

// GameScoreFromScore is PlayerTraits.GameScore when the score comes from
// behaviors.ScoreBehavior's Score property rather than a game method.
const GameScoreFromScore = "score"

// PlayerTraits are the player state's behaviors and methods.
type PlayerTraits struct {
	Seat        bool `json:",omitempty"`
	Inactive    bool `json:",omitempty"`
	Elimination bool `json:",omitempty"`
	Role        bool `json:",omitempty"`
	Color       bool `json:",omitempty"`
	// GameScore is the declared GameScore method, or GameScoreFromScore.
	GameScore FnRef `json:",omitempty"`
	// Turn is interfaces.PlayerTurnFinisher, when the player state has it.
	Turn *TurnTraits `json:",omitempty"`
}

// TurnTraits are the three interfaces.PlayerTurnFinisher methods.
type TurnTraits struct {
	TurnDone          FnRef
	ResetForTurnStart FnRef
	ResetForTurnEnd   FnRef
}

// ComputedData is one computed property.
type ComputedData struct {
	Name string
	// Scope is "global" or "player".
	Scope string
	// Type is a ComputedType of rules/types.ts ("bool", "int", "intSlice", ...).
	Type    string
	Compute FnRef
}

// PredicateData is one custom legal predicate constructor.
type PredicateData struct {
	Name      string
	Construct FnRef
}

// SetUpData are the set-up hooks.
type SetUpData struct {
	// Begin is BeginSetUp, when the game declares it.
	Begin FnRef `json:",omitempty"`
	// Distribute is DistributeComponentToStarterStack; every game has it.
	Distribute FnRef
	// Finish is FinishSetUp, when the game declares it.
	Finish FnRef `json:",omitempty"`
}

// EndData is how the game ends. With none set, the game never ends.
type EndData struct {
	ConditionMet FnRef `json:",omitempty"`
	PlayerScore  FnRef `json:",omitempty"`
	LowScoreWins bool  `json:",omitempty"`
	// Check is a CheckGameFinished override; it replaces the others.
	Check FnRef `json:",omitempty"`
}

// AgentData is one boardgame.Agent the delegate ships.
type AgentData struct {
	Name        string
	DisplayName string
}

// Move bases (MoveData.Base).
const (
	BaseDefault       = "default"
	BaseCurrentPlayer = "currentPlayer"
	BaseAnyPlayer     = "anyPlayer"
)

// MoveData is one configured move.
type MoveData struct {
	Name     string
	HelpText string
	IsFixUp  bool
	// GoType is the move's Go type, "pkg.typeName": "blackjack.moveCurrentPlayerHit".
	GoType string
	// Stock is the nearest stock ancestor and its options, or null for a
	// move whose only ancestor is moves.Default.
	Stock *StockData
	// Base is BaseDefault, BaseCurrentPlayer or BaseAnyPlayer.
	Base string
	// Fields is field name to its description, for fields beyond what Base
	// supplies.
	Fields map[string]FieldData
	// NameSanitization is moves.WithMoveNameSanitization's tag.
	NameSanitization string `json:",omitempty"`
	// HiddenAnimationKey is the alias shown to viewers who may not see the name.
	HiddenAnimationKey string `json:",omitempty"`
	// LegalPhases is moves.WithLegalPhases' enum keys, or null for any phase.
	LegalPhases []int
	// Progression is moves.WithLegalMoveProgression's group, or null.
	Progression *ProgressionData
	// AllowMultipleInProgression is interfaces.AllowMultipleInProgression.
	AllowMultipleInProgression bool
	// Plan is the legal plan Go assembled.
	Plan PlanData
	// Choices is the conformance header's choice-projection entry, verbatim.
	Choices              json.RawMessage `json:",omitempty"`
	SeatsPlayer          bool            `json:",omitempty"`
	GatheringStart       bool            `json:",omitempty"`
	EndsPhaseProgression bool            `json:",omitempty"`
	// Bodies are the methods the game's own type declares (not promoted
	// from a stock ancestor).
	Bodies MoveBodies
}

// StockData is a move's nearest stock ancestor.
type StockData struct {
	// Type is the stock type's name in package moves: "CurrentPlayer",
	// "DealCountComponents", "StartPhase", ...
	Type string
	// Config is the move config's CustomConfiguration(), keys as stored,
	// values serialized by kind: a string, an int, a bool, an enum key (int),
	// a LegalSpec list, an input field list.
	Config map[string]any
}

// FieldData describes one move field (MoveFieldSpec in rules/types.ts).
type FieldData struct {
	// Type is "int", "bool", "string", "playerIndex" or "enum".
	Type string
	// Enum names the enum, for Type "enum".
	Enum string `json:",omitempty"`
	// Sanitize is the field's sanitize struct tag.
	Sanitize string `json:",omitempty"`
	// Input is the MoveInputDisposition: "required", "server-defaulted",
	// "context-owned" or "unsupported".
	Input string
	// Codec is the MoveInputCodec: "integer", "boolean", "enum",
	// "player-index", "string", or "".
	Codec string
}

// PlanData is the legal plan Go assembled for a move: base checks,
// contributions, authored specs, with suppressions applied.
type PlanData struct {
	// Declared is whether the move has declared legality.
	Declared bool
	// Specs is (*GameManager).LegalPlanSpecs(name), used verbatim.
	Specs []boardgame.LegalSpec
	// HasCustom is whether a LegalCustom check runs last.
	HasCustom bool
}

// MoveBodies are the move methods the game declares. An empty FnRef means
// the method is the stock ancestor's (or moves.Default's).
type MoveBodies struct {
	Apply       FnRef `json:",omitempty"`
	Legal       FnRef `json:",omitempty"`
	LegalCustom FnRef `json:",omitempty"`
	// Defaults is DefaultsForState.
	Defaults FnRef `json:",omitempty"`
}

// Progression kinds (ProgressionData.Kind).
const (
	ProgressionMove     = "move"
	ProgressionSerial   = "serial"
	ProgressionRepeat   = "repeat"
	ProgressionParallel = "parallel"
)

// ProgressionData is a move progression group, as data. Which fields are
// meaningful depends on Kind:
//
//	move:     Name
//	serial:   Children
//	repeat:   Count, Child
//	parallel: Count, Children
//
// MarshalJSON writes exactly those keys (plus Kind), so a serial group with
// no children still has "Children": [].
type ProgressionData struct {
	Kind     string
	Name     string
	Children []ProgressionData
	Count    *CounterData
	Child    *ProgressionData
}

// MarshalJSON writes Kind and the fields its kind uses, in that order.
func (p ProgressionData) MarshalJSON() ([]byte, error) {
	children := p.Children
	if children == nil {
		children = []ProgressionData{}
	}
	switch p.Kind {
	case ProgressionMove:
		return json.Marshal(struct {
			Kind string
			Name string
		}{p.Kind, p.Name})
	case ProgressionSerial:
		return json.Marshal(struct {
			Kind     string
			Children []ProgressionData
		}{p.Kind, children})
	case ProgressionRepeat:
		return json.Marshal(struct {
			Kind  string
			Count *CounterData
			Child *ProgressionData
		}{p.Kind, p.Count, p.Child})
	case ProgressionParallel:
		return json.Marshal(struct {
			Kind     string
			Count    *CounterData
			Children []ProgressionData
		}{p.Kind, p.Count, children})
	}
	type plain ProgressionData
	return json.Marshal(plain(p))
}

// Counter kinds (CounterData.Kind): the seven stock families of
// moves/count.go, plus custom.
const (
	CounterAny      = "any"
	CounterAll      = "all"
	CounterInfinite = "infinite"
	CounterAtLeast  = "atLeast"
	CounterAtMost   = "atMost"
	CounterExactly  = "exactly"
	CounterBetween  = "between"
	CounterCustom   = "custom"
)

// CounterData is a moves.ValidCounter, as data. Which fields are
// meaningful depends on Kind:
//
//	any, all, infinite:        none
//	atLeast, atMost, exactly:  N
//	between:                   Min, Max
//	custom:                    Count (the closure, to translate)
//
// MarshalJSON writes exactly those keys (plus Kind), so N: 0 is kept.
type CounterData struct {
	Kind  string
	N     int
	Min   int
	Max   int
	Count FnRef
}

// MarshalJSON writes Kind and the fields its kind uses, in that order.
func (c CounterData) MarshalJSON() ([]byte, error) {
	switch c.Kind {
	case CounterAny, CounterAll, CounterInfinite:
		return json.Marshal(struct{ Kind string }{c.Kind})
	case CounterAtLeast, CounterAtMost, CounterExactly:
		return json.Marshal(struct {
			Kind string
			N    int
		}{c.Kind, c.N})
	case CounterBetween:
		return json.Marshal(struct {
			Kind string
			Min  int
			Max  int
		}{c.Kind, c.Min, c.Max})
	case CounterCustom:
		return json.Marshal(struct {
			Kind  string
			Count FnRef
		}{c.Kind, c.Count})
	}
	type plain CounterData
	return json.Marshal(plain(c))
}

// Function roles (FunctionEntry.Role). Each names the slot the function
// fills; RoleHelper is an in-package function one of the others calls.
const (
	RoleApply              = "apply"
	RoleLegal              = "legal"
	RoleLegalCustom        = "legalCustom"
	RoleDefaults           = "defaults"
	RoleComputed           = "computed"
	RoleSetUpBegin         = "setUpBegin"
	RoleDistribute         = "distribute"
	RoleSetUpFinish        = "setUpFinish"
	RoleConditionMet       = "conditionMet"
	RolePlayerScore        = "playerScore"
	RoleCheckGameFinished  = "checkGameFinished"
	RoleCurrentPlayerIndex = "currentPlayerIndex"
	RolePlayerOrder        = "playerOrder"
	RoleGameScore          = "gameScore"
	RoleTurnDone           = "turnDone"
	RoleResetForTurnStart  = "resetForTurnStart"
	RoleResetForTurnEnd    = "resetForTurnEnd"
	RolePredicate          = "predicate"
	RoleCounter            = "counter"
	RoleComponent          = "component"
	RoleHelper             = "helper"
)

// Verdicts (FunctionEntry.Verdict).
const (
	VerdictSubset  = "subset"
	VerdictRefused = "refused"
)

// FunctionEntry is one game-written function that reaches engine output.
type FunctionEntry struct {
	Func FnRef
	// Role is one of the Role constants.
	Role string
	// File is relative to the package directory; Line is 1-based.
	File string
	Line int
	// SourceHash is SourceHash of the function's source.
	SourceHash string
	// Calls are the in-package callees, sorted, without duplicates.
	Calls []FnRef
	// Verdict is VerdictSubset or VerdictRefused.
	Verdict string
	// Refusals is empty exactly when Verdict is VerdictSubset.
	Refusals []refusal.Refusal
}

// Definition is the engine's GameDefinition (server/static/src/engine/definition.ts),
// derived from a live GameManager. It is the shape
// server/static/src/engine/testdata/gen/vecgame DefinitionFor emits, plus
// the optional DeckDef.GenericValues, DeckDef.ComponentProps, EnumDef.Range
// and PropDef.Constraints. Each addition is omitempty, so a definition with
// none of them serializes byte for byte as DefinitionFor's does.
type Definition struct {
	Name                   string
	Decks                  map[string]DeckDef
	Enums                  map[string]EnumDef
	GroupEnum              string
	Game                   SubStateDef
	Player                 SubStateDef
	DynamicComponentValues map[string]SubStateDef
}

// DeckDef is one deck.
type DeckDef struct {
	Size       int
	Components []json.RawMessage
	// GenericValues is what Deck.SetGenericValues set: the values a
	// sanitized (generic) component shows. Absent when never set.
	GenericValues json.RawMessage `json:",omitempty"`
	// ComponentProps is each component value property's type and enum
	// (what DeckSpec.props carries by hand today). Absent for a deck whose
	// components have no values.
	ComponentProps map[string]ComponentPropDef `json:",omitempty"`
}

// ComponentPropDef is the declared type of one component value property
// (ValuePropDef in rules/binding.ts).
type ComponentPropDef struct {
	// Type is a PropType of definition.ts ("int", "string", "enum", ...).
	Type string
	Enum string `json:",omitempty"`
}

// EnumDef is one enum.
type EnumDef struct {
	// Values maps the decimal enum key to the string a value serializes as.
	Values       map[string]string
	DefaultValue int
	// Parents is a tree enum's parent key per key; the root is its own
	// parent. Absent for other enums.
	Parents map[string]int `json:",omitempty"`
	// Range is a range enum's dimensions (enum.RangeEnum.RangeDimensions),
	// e.g. [8, 8]. Absent for other enums.
	Range []int `json:",omitempty"`
}

// SubStateDef is one sub-state's property schema.
type SubStateDef struct {
	Props map[string]PropDef
}

// PropDef is one property.
type PropDef struct {
	Type string
	// Sanitize maps group name to policy name, exactly as Go parsed the
	// struct tag (StructInflater.PropertySanitizationPolicy).
	Sanitize map[string]string
	Enum     string `json:",omitempty"`
	Deck     string `json:",omitempty"`
	// Size is the slot count of a sized stack, or the space count of a board.
	Size    int `json:",omitempty"`
	MaxSize int `json:",omitempty"`
	// Merged stacks: the sibling properties they view, in order.
	Overlap bool     `json:",omitempty"`
	Stacks  []string `json:",omitempty"`
	// Constraints are the stack's constraint strings from its struct tag
	// (`same(Color)`), in tag order. Absent when there are none.
	Constraints []string `json:",omitempty"`
}
