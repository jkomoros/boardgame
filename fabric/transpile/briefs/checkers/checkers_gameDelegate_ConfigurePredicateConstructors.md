Translate one Go function to TypeScript. Its behavior must be identical,
including behavior that looks like a bug: do not fix, simplify or guard
anything the Go does not guard. If the Go would panic, the TypeScript must
throw the same text (see "Panics").

FUNCTION
  checkers.(*gameDelegate).ConfigurePredicateConstructors, role predicate, from examples/checkers/main.go:146 at Go source hash sha256:399a2423a08a16b314ef1c3ad2b721c5cd52e11c690486d2d8620ff1e3239c4a

GO SOURCE
  func (g *gameDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {
  	return []*legal.PredicateConstructor{{
  		Name: "checkers.spaceIsBlack",
  		Constructor: func(spec legal.Spec, _ *boardgame.ComponentChest,
  			_ func(legal.Spec) (*legal.Predicate, error)) (*legal.Predicate, error) {
  			if len(spec.Args) != 1 {
  				return nil, fmt.Errorf("checkers.spaceIsBlack: requires exactly 1 arg (the space-index field), got %d", len(spec.Args))
  			}
  			field := spec.Args[0]

  			template := spec.Message
  			if template == "" {
  				template = "checkers.black_spaces_only"
  			}

  			return &legal.Predicate{
  				Name: "checkers.spaceIsBlack",
  				Args: spec.Args,
  				Reads: []legal.Read{
  					{Path: legal.PropPath(field), Facet: boardgame.LegalFacetValues},
  				},
  				Cost:             boardgame.LegalCostTrivial,
  				EmittedTemplates: []string{template},
  				// Recommended (not required) game-registered metadata: the
  				// FailT below attaches no bindings, so the template body —
  				// default or WithMessage retarget — may not reference any
  				// {placeholder}; declaring that here gets the mismatch
  				// caught at boot instead of rendering a bare placeholder
  				// name mid-game (see legal/doc.go's game-registered
  				// predicates section).
  				EmittedBindings: map[string][]string{template: nil},
  				Evaluate: func(ctx legal.Context) legal.Verdict {
  					val, propType, err := ctx.ResolvePath(legal.PropPath(field))
  					if err != nil {
  						return legal.UnknownVerdict(err.Error())
  					}
  					if propType != boardgame.TypeEnum {
  						return legal.UnknownVerdict("checkers.spaceIsBlack: path " + field + " is not an enum property")
  					}
  					ev, ok := val.(enum.ImmutableVal)
  					if !ok || ev == nil {
  						return legal.UnknownVerdict("checkers.spaceIsBlack: path " + field + " resolved to a nil or non-enum value")
  					}
  					if spaceIsBlack(ev.Value().Int()) {
  						return legal.PassVerdict()
  					}
  					return legal.FailT(template)
  				},
  			}, nil
  		},
  	}}
  }

WHY THE TRANSLATOR REFUSED IT
  predicate-constructor checkers.spaceIsBlack, at main.go:146:1
    func (g *gameDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {

TYPES IT TOUCHES
  The game's types it names, verbatim:

    type gameDelegate struct {
    	base.GameDelegate
    }

  Framework types it names:

    boardgame.ComponentChest (github.com/jkomoros/boardgame), verbatim:

      // ComponentChest is a list of all decks, enums, and constants, for this game
      // type. Each game type has one ComponentChest, which is an immutable set of
      // all components in this game type. Every game of a given game type uses the
      // same ComponentChest. NewGameManager creates a new ComponentChest, filling it
      // with values returned from your GameDelegate's ConfigureDecks(),
      // ConfigureConstants(), and ConfigureEnums().
      type ComponentChest struct {
      	initialized bool
      	deckNames   []string
      	decks       map[string]*Deck
      	constants   map[string]interface{}
      	enums       *enum.Set

      	manager *GameManager
      }

    boardgame.LegalContext (github.com/jkomoros/boardgame), verbatim:

      // LegalContext is the entire vocabulary a LegalPredicate's Evaluate func may
      // reference — the wall before the Turing tarpit: no I/O, no mutation,
      // nothing beyond these four fields. See the design spec §1 "Predicates are
      // structs-with-a-func, not interfaces".
      type LegalContext struct {
      	// State is the state being evaluated against.
      	State ImmutableState
      	// Move is the move being evaluated, or nil during field-independent
      	// evaluation (a predicate whose declared Reads include no move.* path
      	// may be evaluated with a nil Move). See evalLegalPredicate for the
      	// runtime guard that protects a predicate that touches ctx.Move despite
      	// not declaring a move.* read.
      	Move Move
      	// ProposerPlayerIndex is the runtime PlayerIndex proposing the move. It is
      	// a concrete player during ordinary play and AdminPlayerIndex for engine
      	// actions; ObserverPlayerIndex and AnyPlayerIndex are never actors.
      	ProposerPlayerIndex PlayerIndex
      	// Chest is the game's component chest.
      	Chest *ComponentChest
      }

    boardgame.LegalPredicate (github.com/jkomoros/boardgame), verbatim:

      // LegalPredicate is one resolved legality question: either a leaf, built by
      // a LegalPredicateConstructor from a LegalSpec, or a compositor (only "any"
      // in v1), whose Sub holds its children. See the design spec §1.
      type LegalPredicate struct {
      	// Name is the registry identity of this predicate, e.g.
      	// "playerPropAtLeast", or "any" for the any-compositor.
      	Name string
      	// Args are the string arguments this predicate was constructed with.
      	// With Name, this round-trips the registry. Empty for compositors.
      	Args []string
      	// Reads is the declared read-set: a conservative over-approximation of
      	// the state this predicate's Evaluate depends on, used for client
      	// evaluability and (in the future) dirty-tracking. For the "any"
      	// compositor, Reads is the union of its children's Reads.
      	Reads []LegalRead
      	// Cost is a coarse cost tier, metadata only in v1. For "any", Cost is
      	// the max of its children's Cost.
      	Cost LegalCost
      	// Evaluate is the pure function that computes this predicate's verdict.
      	// Callers should invoke it via evalLegalPredicate rather than directly,
      	// so that a zero-value (invalid) verdict, a panic, or an undeclared
      	// touch of a nil ctx.Move are all converted to a fail-closed
      	// LegalUnknown instead of propagating.
      	Evaluate func(ctx LegalContext) LegalVerdict
      	// RequiredReadTypes optionally pins the PropertyType expected at each
      	// declared path. Boot validation rejects a present-but-wrongly-typed
      	// property instead of allowing every runtime evaluation to become Unknown.
      	RequiredReadTypes map[LegalPropPath]PropertyType
      	// AllowedReadTypes is the honest polymorphic counterpart to
      	// RequiredReadTypes: the property at each declared path must have one of
      	// the listed types. A path may appear in exactly one of these maps. This is
      	// intentionally uncommon; catalog predicates should prefer the stronger
      	// single-type contract whenever their evaluator does.
      	AllowedReadTypes map[LegalPropPath][]PropertyType
      	// RequiredComponentFields declares static component-value fields read
      	// through a stack. Boot validation checks every concrete component in the
      	// stack's deck exposes the named field with one of AllowedTypes. This is a
      	// deliberately narrow companion to RequiredReadTypes: it supports closed
      	// component selectors without adding arbitrary property traversal to the
      	// legal path grammar.
      	RequiredComponentFields []LegalComponentFieldRequirement
      	// AdminPolicy is copied from the originating spec. AdminBypass is legal
      	// only for proposer-dependent predicates and prevents them from trying to
      	// resolve AdminPlayerIndex as an ordinary player.
      	AdminPolicy LegalAdminPolicy
      	// UsesProposer declares that Evaluate depends on the proposing actor even
      	// when that dependency is not represented by a proposer.* Read (for
      	// example, comparing the proposer with a move's PlayerIndex field).
      	UsesProposer bool
      	// ClientEvaluable means the generic client catalog knows this predicate's
      	// semantics. Serialization alone is insufficient: game-defined predicates
      	// and compositors without serialized children are deliberately false.
      	ClientEvaluable bool
      	// EmittedTemplates is the conservative set of template keys this
      	// predicate's Evaluate may emit in a Fail (or Message-carrying Unknown)
      	// verdict — the effective keys AFTER any Spec.Message override was
      	// applied at construction time. It is populated by the constructor that
      	// built the predicate (each catalog constructor lists the key(s) its
      	// Evaluate can FailT with; a game-registered constructor should do the
      	// same). Boot-time validation (validateLegalEmittedTemplates) checks
      	// every key here against the game's merged template table, so a
      	// predicate that can emit an unregistered template key fails at
      	// NewGameManager naming the owning move rather than rendering a bare key
      	// mid-game — closing the spec §3 "unregistered keys are a boot error"
      	// invariant for game-registered predicates and implicit catalog
      	// defaults alike. Empty/nil is allowed (an opaque escape-hatch wrapper,
      	// or a predicate that never fails with a template, declares none).
      	EmittedTemplates []string
      	// EmittedBindings maps each template key in EmittedTemplates to the set
      	// of binding names Evaluate is GUARANTEED to attach whenever it emits a
      	// Message with that key (footgun-batch F4). Boot-time validation
      	// (validateLegalEmittedBindings) extracts the {placeholders} from each
      	// key's resolved template body — including WithMessage retargets and
      	// ConfigureLegalTemplates body overrides — and requires them to be a
      	// subset of that key's entry here, so a template referencing a binding
      	// its predicate never emits fails at NewGameManager (naming the owning
      	// move, the key, and the missing binding) rather than rendering the bare
      	// placeholder name mid-game. A key whose Evaluate emission carries no
      	// bindings maps to nil/empty. When two distinct emission branches are
      	// collapsed onto ONE key by a Spec.Message override (e.g. mayMoveTo's
      	// no-component and may-not-move branches), the entry must be the
      	// INTERSECTION of the branches' binding sets — only a binding present on
      	// every emission with that key is guaranteed renderable. Like
      	// EmittedTemplates, this is server-side construction metadata, never
      	// serialized (LegalPredicate has no wire form; LegalSpec is the wire
      	// shape). A nil map means "no metadata declared": validation is skipped
      	// for the whole predicate. Every catalog constructor populates it;
      	// game-registered constructors are strongly encouraged to (see
      	// legal/doc.go's game-registered predicates section) but a nil map stays
      	// legal so existing registrations keep booting.
      	EmittedBindings map[string][]string
      	// Sub holds child predicates; only set for compositors ("any" in v1).
      	Sub []*LegalPredicate
      	// opaque marks an escape-hatch wrapper predicate (e.g. LegalCustom) that
      	// has no serialized form: it did not come from resolving a LegalSpec
      	// and cannot be turned back into one. Unexported: only this package's
      	// escape-hatch wrapper (added in a later task) sets it.
      	opaque bool
      }

    boardgame.LegalPredicateConstructor (github.com/jkomoros/boardgame), verbatim:

      // LegalPredicateConstructor is a named factory that turns a LegalSpec into a
      // *LegalPredicate. Constructors mirror constraints.StackConstraintConstructor:
      // games register their own predicates through the same registry mechanism
      // resolveLegalSpecs consumes, which is how a game-specific check (e.g.
      // checkers' board-geometry predicate) stays serializable. See the design
      // spec §1.
      type LegalPredicateConstructor struct {
      	// Name is the registry key this constructor is registered under. It
      	// should match the Name every LegalSpec it constructs from carries.
      	Name string
      	// Constructor builds a *LegalPredicate from spec. resolve is provided so
      	// a constructor whose spec has its own notion of sub-specs (distinct
      	// from the "any" compositor) can recursively resolve them through the
      	// same registry and validation path as top-level specs.
      	Constructor func(spec LegalSpec, chest *ComponentChest,
      		resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error)
      }

    boardgame.LegalPropPath (github.com/jkomoros/boardgame), verbatim:

      // LegalPropPath is a property path in the state resolver's path grammar,
      // e.g. "game.HiddenCards" or "move.CardIndex".
      type LegalPropPath string

    boardgame.LegalRead (github.com/jkomoros/boardgame), verbatim:

      // LegalRead declares one property a predicate depends on, along with the
      // facet of that property it actually needs. Read-sets are a conservative
      // over-approximation used for client evaluability and, in the future,
      // dirty-tracking.
      type LegalRead struct {
      	// Path is the property path being read.
      	Path LegalPropPath
      	// Facet is the aspect of that property being read.
      	Facet LegalFacet
      }

    boardgame.LegalSpec (github.com/jkomoros/boardgame), verbatim:

      // LegalSpec is the serializable, registry-resolvable form of a predicate or
      // compositor. A leaf spec has Name/Args; a compositor spec (only "any" in
      // v1) has Name/Sub. Message optionally overrides the predicate's default
      // template key.
      type LegalSpec struct {
      	// Name is the registry identity of the predicate or compositor, e.g.
      	// "playerPropAtLeast" or "any".
      	Name string `json:"name"`
      	// Args are the string arguments to the named predicate.
      	Args []string `json:"args,omitempty"`
      	// Sub holds child specs for a compositor.
      	Sub []LegalSpec `json:"sub,omitempty"`
      	// Message, if set, overrides the predicate's default template key.
      	Message string `json:"message,omitempty"`
      	// AdminPolicy controls how this predicate behaves when the engine/admin
      	// is the proposer. The zero value evaluates normally. AdminBypass is only
      	// valid for proposer-dependent predicates reading proposer or move fields.
      	AdminPolicy LegalAdminPolicy `json:"adminPolicy,omitempty"`
      }

    boardgame.LegalVerdict (github.com/jkomoros/boardgame), verbatim:

      // LegalVerdict is the result of evaluating one predicate. Its zero value is
      // intentionally invalid; use the constructors in the legal package
      // (legal.PassVerdict, legal.FailT, legal.UnknownVerdict) to build one.
      type LegalVerdict struct {
      	// Outcome is the three-valued result of evaluation.
      	Outcome LegalOutcome
      	// Message is set on LegalFail (and optionally LegalUnknown); nil on
      	// LegalPass.
      	Message *LegalMessage
      	// Reason explains why the outcome is LegalUnknown, e.g. "reads hidden
      	// property HiddenCards".
      	Reason string
      }

    boardgame.PropertyType (github.com/jkomoros/boardgame), verbatim:

      // PropertyType is an enumeration of the types that are legal to have on an
      // underyling object that can return a Reader. This ensures that State objects
      // are not overly complex and can be reasoned about cleanly. See
      // PropertyReadSetConfigurer and ConfigurableSubState for more.
      type PropertyType int

    enum.EnumKey (github.com/jkomoros/boardgame/enum): its TypeScript shape is in the Types table under SEMANTICS.
    enum.ImmutableVal (github.com/jkomoros/boardgame/enum): its methods are the API table's rows (THE TYPESCRIPT API YOU MAY USE).

  State properties it reads or writes, with their PropDef from the exported definition:

    (none)

  Framework constants it reads, with their values:

    boardgame.LegalCostTrivial (boardgame.LegalCost) = 0
    boardgame.LegalFacetValues (boardgame.LegalFacet) = 0
    boardgame.TypeEnum (boardgame.PropertyType) = 5

  The exported definition's predicates this function constructs (GameData.Predicates[i].Construct names it):

    "checkers.spaceIsBlack": the loader calls the translation with a spec whose name is "checkers.spaceIsBlack", once per plan spec naming it
    Go returns a list of constructors; the translation is the Constructor of the entry with that Name.

  The game's moves, from the exported definition (name: Go type, fix-up, stock type):

    "Seat Player": moves.SeatPlayer, fix-up true, stock moves.SeatPlayer
    "Activate Inactive Player": moves.ActivateInactivePlayer, fix-up true, stock moves.ActivateInactivePlayer
    "Place Token": checkers.movePlaceToken, fix-up true, stock moves.FixUpMulti
    "Start Phase Playing": moves.StartPhase, fix-up true, stock moves.StartPhase
    "Crown Token": checkers.moveCrownToken, fix-up true, stock moves.DefaultComponent
    "Finish Turn": moves.FinishTurn, fix-up true, stock moves.FinishTurn
    "Move Token": checkers.moveMoveToken, fix-up false, stock moves.CurrentPlayer

FUNCTIONS IT CALLS (already translated; call them, do not rewrite them)
  checkers.spaceIsBlack
    Go:         func spaceIsBlack(spaceIndex int) bool
    TypeScript: export function checkers_spaceIsBlack(state: RulesState, spaceIndex: number): boolean
    import { checkers_spaceIsBlack } from './bodies.gen.ts';

THE TYPESCRIPT API YOU MAY USE
  Names a translation may import (module relative to src/engine/; "type" marks a type-only export):

    go/api.ts: componentAt, componentAtKey, components, containingStack, contractSize, diceRoll, dynamic, dynamicOf, enumEquals, enumKey, expandSize, first, gameStack, last, legalError, mayMoveAllTo, mayMoveCountTo, mayMoveTo, mayMoveToSlot, moveAllTo, moveCountTo, moveTo, moveToFirstSlot, moveToLastSlot, moveToNextSlot, neighborsByKey, player, playerIndexes, playerStack, publicShuffle, rangeToValue, rangeValue, secretMoveTo, setEliminated, setEnumKey, setPlayerInactive, setSeatFilled, setSize, shuffle, sizeToFit, slideToFirstSlot, slideToLastSlot, sortComponents, stackView, stockApply, swapComponents, swapComponentsByKey, valueToRange, values, valuesAreNil, valuesOf, variantValue
    go/num.ts: idiv, imod, imul
    go/panic.ts: nonNil
    go/slice.ts: appendTo, at, copyInto, lenOf, makeSlice, setAt, subslice
    go/str.ts: atoi, contains, hasPrefix, hasSuffix, itoa, join, repeat, split, sprintf, strCompare, toLower, trimSpace, utf8Len
    legal/types.ts: failT, passVerdict, unknownVerdict
    go/api.ts (type): Comp, DynamicRef, GraphRef, StackRef
    legal/types.ts (type): LegalCost, LegalRead, LegalVerdict, PropertyType
    rand.ts (type): GoRand
    rules/state.ts (type): RulesState, TimerHandle
    rules/types.ts (type): GamePredicate, MoveInstance, PredicateContext, PredicateSpec

  The API table's rows for each Go call this function makes outside its package (lib/transpile/subset):

    (boardgame.LegalContext).ResolvePath: not in the API table
    (enum.EnumKey).Int: enum key as int (call; added: EnumKey and int are both number; Int() is the identity conversion the Expressions section allows)
    (enum.ImmutableVal).Value: enumKey (call; proposal)
    fmt.Errorf: sprintf (format; proposal)
    legal.FailT: not in the API table
    legal.PassVerdict: not in the API table
    legal.UnknownVerdict: not in the API table

  The types the signatures name, verbatim:

    // rules/state.ts
    export interface RulesState {
    	readonly game: CompiledGame;
    	readonly engine: EngineState;
    	readonly record: GameRecord;
    	/** The move records of versions 1 through the committed version. */
    	readonly history: readonly MoveRecord[];
    	readonly env: Environment;
    	/**
    	 * The committed state this one is a working copy of (itself, for a
    	 * committed state). Go's state.Game().MoveByName binds a new move to the
    	 * game's current state, not to the copy a move is being applied to.
    	 */
    	readonly committed: RulesState;
    	/** Timers started on this copy whose deadline is set when it is saved. */
    	readonly pendingDeadlines: { record: TimerRecord; durationMs: number }[];
    	legalState: LegalState | null;
    	readonly framework: LegalFramework;
    	// ----- the state itself ----------------------------------------------
    	readonly version: number;
    	/** The game sub-state's JSON; scalars are read and written here. */
    	readonly g: SubStateJSON;
    	/** Each player sub-state's JSON. */
    	readonly players: SubStateJSON[];
    	numPlayers(): number;
    	gameStack(prop: string): Stack;
    	playerStack(player: number, prop: string): Stack;
    	/** Any game stack property, including a merged one, read-only. */
    	gameView(prop: string): StackView;
    	/** Any player stack property, including a merged one, read-only. */
    	playerView(player: number, prop: string): StackView;
    	/**
    	 * A component's static values ({} for a component that has none). For
    	 * GENERIC_COMPONENT (-2), the deck's generic values ({} when none were set).
    	 */
    	component(deck: string, deckIndex: number): Readonly<Record<string, unknown>>;
    	/** A component's dynamic values. */
    	dynamic(deck: string, deckIndex: number): SubStateJSON;
    	/** state.Rand(): draw only while applying a move or setting up. */
    	rand(): GoRand;
    	/** The variant the game was created with (never null: {} when empty). */
    	variant(): Readonly<Record<string, string>>;
    	// ----- what Go asks the delegate -------------------------------------
    	/** The evaluator's view of this state. */
    	readonly legal: LegalState;
    	/** delegate.CurrentPlayerIndex(state). */
    	currentPlayerIndex(): number;
    	/** behaviors.PlayerIsInactive. */
    	playerIsInactive(player: number): boolean;
    	/** behaviors.PlayerIsEliminated. */
    	playerIsEliminated(player: number): boolean;
    	/** delegate.PlayerMayBeActive. */
    	playerMayBeActive(player: number): boolean;
    	/** delegate.CustomPlayerOrder. */
    	customPlayerOrder(): readonly number[] | null;
    	/** delegate.NumSeatedActivePlayers. */
    	numSeatedActivePlayers(): number;
    	/** PlayerIndex.Valid. */
    	valid(player: number): boolean;
    	/** PlayerIndex.Next. */
    	next(player: number): number;
    	/** PlayerIndex.EnsureValid. */
    	ensureValid(player: number): number;
    	/** PlayerIndex.Previous. */
    	previous(player: number): number;
    	/** PlayerIndex.WithinBounds. */
    	withinBounds(player: number): boolean;
    	// ----- phases --------------------------------------------------------
    	/** delegate.CurrentPhase: null when the game state has no Phase enum. */
    	currentPhase(): { key: number; name: string; ancestors: readonly number[] | null } | null;
    	/** The current phase's key, or 0 with no phase (as Go records it). */
    	currentPhaseKey(): number;
    	/** behaviors.PhaseBehavior.SetCurrentPhase. */
    	setPhase(key: number): void;
    	/**
    	 * Default.legalMoveInProgression: would this move, appended to the moves
    	 * applied since the current phase was entered, still match its phase's
    	 * progression? Null when it would (or the move is in no progression).
    	 */
    	moveInProgression(moveName: string): string | null;
    	// ----- moves ---------------------------------------------------------
    	/**
    	 * A new move of the named type with its defaults for this state (Go:
    	 * MoveByNameForState). Null for an unknown name.
    	 */
    	newMove(name: string): (MoveInstance & { version: number; initiator: number }) | null;
    	/** state.Game().MoveByName(name): bound to the game's committed state. */
    	moveByName(name: string): MoveInstance & { version: number; initiator: number };
    	legalContext(compiled: CompiledMove, move: MoveInstance, proposer: number): LegalContext;
    	/**
    	 * move.Legal(state, proposer): null when legal, otherwise the error text.
    	 * A check that throws is an error, as a Go panic there would be.
    	 */
    	legalError(move: MoveInstance, proposer: number): string | null;
    	/** GameManager.LegalEvaluateLedger: every predicate's verdict. Null for a move whose legality is not declared. */
    	legalLedger(move: MoveInstance, proposer: number): LegalPlanResult | null;
    	/** delegate.ProposeFixUpMove: the first fix-up move legal for the admin. */
    	proposeFixUpMove(): (MoveInstance & { version: number; initiator: number }) | null;
    	// ----- timers --------------------------------------------------------
    	/** A Timer property of the game state. */
    	timer(prop: string): TimerHandle;
    	/** A Timer property of a player state. */
    	playerTimer(player: number, prop: string): TimerHandle;
    }

    // rules/state.ts
    /** A Timer property of the state (timer.go). */
    export interface TimerHandle {
    	readonly state: RulesState;
    	readonly record: TimerRecord;
    	readonly node: { ID: string; IsTimer: true };
    	/** Timer.Active. */
    	active(): boolean;
    	readonly id: string;
    	/**
    	 * Timer.Start: when `durationMs` has passed (counted from the moment the
    	 * state is saved), `move` is proposed as the admin. Draws the timer's id
    	 * from the state's generator: 16 draws.
    	 */
    	start(durationMs: number, move: MoveInstance & { version?: number; initiator?: number }): void;
    	/** Timer.Cancel: whether the timer was active. */
    	cancel(): boolean;
    }

    // rules/types.ts
    /** One move, with its field values. */
    export interface MoveInstance {
    	readonly name: string;
    	fields: Record<string, unknown>;
    }

    // legal/types.ts
    /** Go: LegalVerdict. */
    export interface LegalVerdict {
      outcome: LegalOutcome;
      /** Set on fail, and optionally on unknown (the `any` compositor). */
      message?: LegalMessage;
      /** Why the outcome is unknown. Diagnostic text; matches Go byte for byte where Go produces it. */
      reason?: string;
    }

  A predicate constructor's types and verdicts (role predicate), verbatim:

    // rules/types.ts
    /**
     * Go: the legal.Spec a game's predicate constructor receives. `args` is the
     * spec's Args ([] where Go's is nil); `message` is its Message, '' when the
     * spec does not override the template key.
     */
    export interface PredicateSpec {
    	readonly name: string;
    	readonly args: readonly string[];
    	readonly message: string;
    }

    // rules/types.ts
    /**
     * Go: legal.Context, as a game-registered predicate's Evaluate sees it: all
     * it may read. Nothing else is reachable, and Evaluate must not write.
     */
    export interface PredicateContext {
    	/** Go: ctx.State (a read-only view; the facade's readers take it). */
    	readonly state: RulesState;
    	/** Go: ctx.Move. Null where Go's is nil (an evaluation with no move). */
    	readonly move: MoveInstance | null;
    	/** Go: ctx.ProposerPlayerIndex. */
    	readonly proposer: number;
    	/**
    	 * Go: `ctx.ResolvePath(legal.PropPath(path))`: the property a legal path
    	 * names, its PropertyType (Go's String() form, 'TypeEnum', 'TypeInt',
    	 * ...), and Go's error text (then the value is null and the type
    	 * 'TypeIllegal'). The value is, by type: a number for TypeInt and
    	 * TypePlayerIndex, a boolean, a string, `{ enum: { name }, value: key,
    	 * string }` or null for TypeEnum (Go's enum.ImmutableVal, nil), and a
    	 * legal-layer stack or board handle for TypeStack/TypeBoard.
    	 */
    	resolvePath(path: string): [value: unknown, type: PropertyType, err: string | null];
    }

    // rules/types.ts
    /**
     * Go: *legal.Predicate, what a constructor returns. Only `name`, `reads` and
     * `evaluate` are required; everything else is Go's zero value when absent.
     * `reads` must cover every path `evaluate` resolves (Go's honor system); a
     * read of `move.*` makes the ledger entry provisional.
     */
    export interface GamePredicate {
    	name: string;
    	/** Go: Args. Absent or empty where Go leaves them nil. */
    	args?: readonly string[];
    	reads: readonly LegalRead[];
    	requiredReadTypes?: Readonly<Record<string, PropertyType>>;
    	allowedReadTypes?: Readonly<Record<string, readonly PropertyType[]>>;
    	/** Go: Cost (metadata only). Default 0, LegalCostTrivial. */
    	cost?: LegalCost;
    	/** Go: UsesProposer. */
    	usesProposer?: boolean;
    	/** Go: EmittedTemplates: every template key evaluate may emit. */
    	emittedTemplates?: readonly string[];
    	/** Go: EmittedBindings. Absent is Go's nil (no metadata). */
    	emittedBindings?: Readonly<Record<string, readonly string[]>>;
    	/**
    	 * Go: Evaluate. Returns Go's verdict shape (legal/types.ts passVerdict,
    	 * failT, unknownVerdict). A throw is Go's panic: the plan turns it into
    	 * an unknown verdict, `predicate "<name>" panicked: <message>`.
    	 */
    	evaluate: (ctx: PredicateContext) => LegalVerdict;
    }

    // rules/types.ts
    /**
     * Go: legal.PredicateConstructor.Constructor, for a game's own predicate
     * (ConfigurePredicateConstructors). Returns the predicate, or Go's error:
     * a throw or a returned string, which plan assembly reports as Go does
     * (`boardgame: legal spec "<name>": constructor failed: <text>`). Go's
     * chest and resolve arguments are not passed; no game uses them.
     */
    export type GamePredicateConstructor = (spec: PredicateSpec) => GamePredicate | string;

    // legal/types.ts
    /** Go: LegalRead. One property a predicate depends on, and the facet it needs. */
    export interface LegalRead {
      path: string;
      facet: LegalFacet;
    }

    // legal/types.ts
    /**
     * Go: LegalFacet. Which aspect of a property a predicate reads. The names are
     * the wire spelling (Go visibilityFacetName), the same strings a viewer's
     * state carries in its Visibility block.
     */
    export type LegalFacet = 'values' | 'count' | 'occupancy' | 'order' | 'nonempty';

    // legal/types.ts
    /** Go: LegalCost. Metadata only; never affects evaluation order. */
    export type LegalCost = 0 | 1 | 2 | 3;

    // legal/types.ts
    /**
     * Go: PropertyType, spelled as its String() form so error text and vectors
     * compare directly.
     */
    export type PropertyType =
      | 'TypeIllegal'
      | 'TypeInt'
      | 'TypeBool'
      | 'TypeString'
      | 'TypePlayerIndex'
      | 'TypeEnum'
      | 'TypeIntSlice'
      | 'TypeBoolSlice'
      | 'TypeStringSlice'
      | 'TypePlayerIndexSlice'
      | 'TypeStack'
      | 'TypeBoard'
      | 'TypeTimer'
      | 'TypeEnumSlice';

    // legal/types.ts
    /** Go: LegalBindingValue. The closed scalar union a template binding may hold. */
    export type LegalBindingValue = string | number | boolean;

    // legal/types.ts
    /** Go: legal.PassVerdict. */
    export function passVerdict(): LegalVerdict

    // legal/types.ts
    /**
     * Go: legal.FailT. Bindings are omitted from the message entirely when none
     * are given, exactly as Go leaves Bindings nil.
     */
    export function failT(template: string, bindings?: Record<string, LegalBindingValue>): LegalVerdict

    // legal/types.ts
    /** Go: legal.UnknownVerdict. */
    export function unknownVerdict(reason: string): LegalVerdict

  go/api.ts

    // The Go-shaped facade the translator emits calls to: the third column of
    // fabric/TRANSPILE-PROPOSAL.md's API table, written once over the engine
    // (rules/state.ts, stack.ts). It has Go's shapes, not the engine's: a stack
    // operation takes a component and returns an error (`string | null`)
    // instead of taking a slot and throwing, so translated code keeps Go's
    // control flow, including discarded errors. A failed operation changes
    // nothing (the engine validates before it mutates, as Go does).
    //
    // Calls the table maps to "the same" engine method (state.numPlayers(),
    // state.currentPlayerIndex(), state.next(p), state.rand().intn(n),
    // state.moveByName(n), state.newMove(n), state.timer('P').start(ms, m),
    // t.active(), t.cancel(), stack length queries through `stackView`, ...) are
    // emitted against the engine directly and have no function here.
    //
    // Types, from the proposal's "Types" table:
    //
    //   Go *playerState                     number, the player index
    //   Go Stack, SizedStack, MergedStack    StackRef
    //   Go ComponentInstance (nil-able)      Comp | null
    //   Go *dynamicValues struct             DynamicRef
    //   Go enum.EnumKey                      number (the key)
    //   Go error                             string | null
    //
    // A Go ComponentInstance is the component's identity (deck and index) and
    // the state it was read from, not a slot: every operation on it finds where
    // the component is now (state.containingStack). So is a Comp. A handle from
    // a `components(s)` snapshot whose component has since moved still names
    // that component, and operating on it acts on it where it now is, exactly
    // as in Go. What fails is what fails in Go: a component the state's index
    // has lost track of (after sizedStack.ContractSize, "that component's
    // location is not public information"), or a generic one.
    //
    // Every answer is held to Go's by testdata/gen/facadevec (go/stacks.test.ts,
    // go/lang.test.ts).

    /**
     * A stack, as Go code holds one: any stack property (growable, sized or
     * merged) or one space of a board, of the state it was read from. A plain
     * location, so a handle may be held across stack operations.
     */
    export interface StackRef {
    	readonly state: RulesState;
    	/** Where the stack lives: sub-state, property, and board space. */
    	readonly at: EngineStackRef;
    }

    /**
     * A component instance, as Go's ComponentInstance: its identity in its deck,
     * and the stack and slot it was found at. Operations use the identity (and
     * the state of `stack`), never `slot`, as Go's do. `slot` is -1 for a
     * component handed to a `sortComponents` less function, which has no slot.
     * `index` is GENERIC_COMPONENT (-2) for a sanitized component.
     */
    export interface Comp {
    	readonly stack: StackRef;
    	readonly slot: number;
    	readonly deck: string;
    	readonly index: number;
    }

    /** A component's dynamic values struct, as Go code holds one. */
    export interface DynamicRef {
    	readonly deck: string;
    	readonly index: number;
    }

    /**
     * Who holds an enum property: the game sub-state, a player (by index), a
     * component's dynamic values, or a component's static values. The component
     * may be null (Go's nil), which panics on use as Go's does.
     */
    export type Owner =
    	| 'game'
    	| number
    	| { readonly dynamicOf: Comp | DynamicRef | null }
    	| { readonly valuesOf: Comp | DynamicRef | null }
    	| { readonly fieldsOf: MoveInstance };

    /**
     * A grid graph over a range enum (enum/graph: MustNewEnumGridConnectedness),
     * rebuilt from its constructor's constant arguments: the enum's name, its
     * dimensions (EnumDef.Range) and the Go direction names given
     * (`DirectionUp`, `DirectionDown`, `DirectionLeft`, `DirectionRight`,
     * `DirectionPerpendicular`, `DirectionDiagonal`), in order.
     */
    export interface GraphRef {
    	readonly enum: string;
    	readonly dimensions: readonly number[];
    	readonly directions: readonly string[];
    }

    /**
     * Go: enum.IllegalValue (math.MaxInt64), as a number. It is not a safe
     * integer (it reads as 2^63), but it is the same number the translator's
     * folded constant `enum.IllegalValue` becomes, so comparisons agree. It is
     * never stored: storing it fails in Go too.
     */
    export const ILLEGAL_ENUM_VALUE = 9223372036854775807;

    /**
     * Go: `players[i]`, a player sub-state, for reading and writing its scalar
     * properties (`player(state, i).X`). Throws Go's index panic when i is not
     * a player index of this game.
     */
    export function player(state: RulesState, i: number): SubStateJSON

    /**
     * Go: the `players` slice concreteStates builds (and `state.PlayerStates()`
     * ranged over): every player, as its index, in order. Added for the
     * translator (workstream 10).
     */
    export function playerIndexes(state: RulesState): number[]

    /** Go: `game.X` where X is a stack property of the game state (merged stacks included). */
    export function gameStack(state: RulesState, prop: string): StackRef

    /** Go: `p.X` where X is a stack property of player i's state (merged stacks included). */
    export function playerStack(state: RulesState, i: number, prop: string): StackRef

    /**
     * The engine's read-only view of a stack, for the queries the table maps to
     * the same name: `s.Len()`, `NumComponents()`, `SlotsRemaining()`,
     * `MaxSize()` become `stackView(s).len()`, and so on.
     */
    export function stackView(s: StackRef): StackView

    /** Go: `s.ComponentAt(i)`, `s.ImmutableComponentAt(i)`. Null where Go returns nil. */
    export function componentAt(s: StackRef, i: number): Comp | null

    /** Go: `s.First()`, `s.ImmutableFirst()`. Null for an empty stack. */
    export function first(s: StackRef): Comp | null

    /** Go: `s.Last()`, `s.ImmutableLast()`. Null for an empty stack. */
    export function last(s: StackRef): Comp | null

    /**
     * Go: `s.Components()`, `s.ImmutableComponents()`: a snapshot taken now, one
     * entry per slot, null for an empty slot of a sized stack.
     */
    export function components(s: StackRef): (Comp | null)[]

    /**
     * Go: `s.ComponentAtKey(k)`, `s.ImmutableComponentAtKey(k)` on a stack
     * indexed by an enum: Go's is `ComponentAt(int(k))`.
     */
    export function componentAtKey(s: StackRef, key: number): Comp | null

    /**
     * Go: `c.Values().(*T)`: the component's static values (the deck's generic
     * values for a generic component). Throws NIL_DEREF_TEXT when c is null.
     * A generic component of a deck with no generic values has nil values in
     * Go, and the type assertion panics; that throws here too, with its own text.
     */
    export function values(state: RulesState, c: Comp | DynamicRef | null): Readonly<Record<string, unknown>>

    /**
     * Go: `c.DynamicValues().(*T)`, or a dynamic values struct Go code already
     * holds: its properties, writable. Throws NIL_DEREF_TEXT when c is null.
     * For a Comp, the values are those of the state the component was read
     * from, as in Go. A generic component gets a fresh zero struct, as Go's
     * DynamicValues does (writes to it are lost); its enum and stack
     * properties are absent, Go's nil.
     */
    export function dynamic(state: RulesState, c: Comp | DynamicRef | null): SubStateJSON

    /**
     * Go: `c.Values().(*T)` as a value Go code holds: the handle a translated
     * values struct is (its fields are then read with `values(state, h)`),
     * after the checks Go's call and type assertion make: Go's nil-dereference
     * panic for a nil c, and values()'s throw for a generic component with no
     * values. Added for the translator (workstream 10).
     */
    export function valuesOf(state: RulesState, c: Comp | DynamicRef | null): DynamicRef

    /**
     * Go: `c.DynamicValues().(*T)` as a value Go code holds: the handle a
     * translated dynamic values struct is (`dynamic(state, h)` reads and writes
     * it), after dynamic()'s checks. Added for the translator (workstream 10).
     */
    export function dynamicOf(state: RulesState, c: Comp | DynamicRef | null): DynamicRef

    /**
     * Go: `c.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())`: one roll
     * of the die c from the state's generator, as components/dice does it (and
     * lib/dice.ts `rollWith`, which the facade may not import): one draw of
     * Intn(len(Faces)) becomes SelectedFace, Value becomes that face, RollCount
     * goes up by one; a die whose static values have no Faces does not roll.
     * Go's nil-dereference panic for a nil c. Added for the translator
     * (workstream 10).
     */
    export function diceRoll(state: RulesState, c: Comp | DynamicRef | null): void

    /**
     * Go: `c.ContainingImmutableStack()`: where the component is now, as Go's
     * three results (stack, slot, error). The stack is null when err is not.
     * The answer is from the state the component was read from. A DynamicRef
     * (Go's `c.ContainingComponent().ImmutableInstance(state)` in a values
     * method, added for the translator) is located in `state`.
     */
    export function containingStack(state: RulesState, c: Comp | DynamicRef | null): [StackRef | null, number, string | null]

    /**
     * Go: `c.Values() == nil`: a generic component of a deck with no generic
     * values (Go's Values() is then nil; values() throws for it). Go's
     * nil-dereference panic for a nil c. Added for the translator (workstream 10).
     */
    export function valuesAreNil(state: RulesState, c: Comp | DynamicRef | null): boolean

    /** Go: `c.MoveTo(dest, slot)`. Null on success, else Go's error text; nothing changes on failure. */
    export function moveTo(c: Comp | null, dest: StackRef, slot: number): string | null

    /** Go: `c.SecretMoveTo(dest, slot)`. */
    export function secretMoveTo(c: Comp | null, dest: StackRef, slot: number): string | null

    /** Go: `c.MoveToFirstSlot(dest)`. */
    export function moveToFirstSlot(c: Comp | null, dest: StackRef): string | null

    /** Go: `c.MoveToLastSlot(dest)`. */
    export function moveToLastSlot(c: Comp | null, dest: StackRef): string | null

    /** Go: `c.MoveToNextSlot(dest)`. */
    export function moveToNextSlot(c: Comp | null, dest: StackRef): string | null

    /** Go: `c.SlideToFirstSlot()`. */
    export function slideToFirstSlot(c: Comp | null): string | null

    /** Go: `c.SlideToLastSlot()`. */
    export function slideToLastSlot(c: Comp | null): string | null

    /** Go: `c.MayMoveTo(dest)`: null when the move would succeed, else why not. */
    export function mayMoveTo(c: Comp | null, dest: StackRef): string | null

    /** Go: `c.MayMoveToSlot(dest, slot)`. */
    export function mayMoveToSlot(c: Comp | null, dest: StackRef, slot: number): string | null

    /** Go: `s.MoveAllTo(dest)`. */
    export function moveAllTo(s: StackRef, dest: StackRef): string | null

    /** Go: `s.MoveCountTo(dest, n)`. */
    export function moveCountTo(s: StackRef, dest: StackRef, n: number): string | null

    /** Go: `s.MayMoveAllTo(dest)`. */
    export function mayMoveAllTo(s: StackRef, dest: StackRef): string | null

    /** Go: `s.MayMoveCountTo(dest, n)`. */
    export function mayMoveCountTo(s: StackRef, dest: StackRef, n: number): string | null

    /** Go: `s.Shuffle()`. Draws from state.rand() as Go does. */
    export function shuffle(s: StackRef): string | null

    /** Go: `s.PublicShuffle()`. */
    export function publicShuffle(s: StackRef): string | null

    /** Go: `s.SwapComponents(i, j)`. */
    export function swapComponents(s: StackRef, i: number, j: number): string | null

    /** Go: `s.SwapComponentsByKey(i, j)` on a stack indexed by an enum: Go's is `SwapComponents(int(i), int(j))`. */
    export function swapComponentsByKey(s: StackRef, i: number, j: number): string | null

    /**
     * Go: `s.SortComponents(less)`, with Go's sort.Sort order among equal
     * elements (gosort.ts). `less` gets components, null for an empty slot;
     * their `slot` is -1. A `less` that throws (a Go panic) leaves the stack
     * as it was, as Go's sort of a copy does.
     */
    export function sortComponents(s: StackRef, less: (a: Comp | null, b: Comp | null) => boolean): string | null

    /** Go: `s.SetSize(n)` on a sized stack. */
    export function setSize(s: StackRef, n: number): string | null

    /** Go: `s.ExpandSize(n)` on a sized stack. */
    export function expandSize(s: StackRef, n: number): string | null

    /** Go: `s.ContractSize(n)` on a sized stack. */
    export function contractSize(s: StackRef, n: number): string | null

    /** Go: `s.SizeToFit()` on a sized stack. */
    export function sizeToFit(s: StackRef): string | null

    /** Go: `state.Game().Variant()[k]`. A missing key gives ''. */
    export function variantValue(state: RulesState, key: string): string

    /** Go: `p.SetEliminated()` (behaviors.PlayerElimination). */
    export function setEliminated(state: RulesState, i: number): void

    /** Go: `p.SetPlayerInactive()` (behaviors.InactivePlayer). */
    export function setPlayerInactive(state: RulesState, i: number): void

    /** Go: `p.SetSeatFilled()` (behaviors.Seat): a filled seat is also closed. */
    export function setSeatFilled(state: RulesState, i: number): void

    /** Go: `v.Value()` on an enum property: its key, converted from the stored string. */
    export function enumKey(state: RulesState, owner: Owner, prop: string): number

    /** Go: `v.SetValue(k)` on an enum property: stores k's string. Null, or Go's error for a key the enum lacks. */
    export function setEnumKey(state: RulesState, owner: Owner, prop: string, key: number): string | null

    /** Go: `v.Equals(o)` on two enum properties: the same enum and the same key. */
    export function enumEquals(state: RulesState, owner: Owner, prop: string, otherOwner: Owner, otherProp: string): boolean

    /**
     * Go: `v.RangeValue()` on a range-enum property: its indexes, one per
     * dimension, parsed from the stored string. Null where Go returns nil.
     */
    export function rangeValue(state: RulesState, owner: Owner, prop: string): number[] | null

    /**
     * Go: `e.RangeToValue(indexes...)`: the key whose string is the indexes
     * joined with `_` (as `e.ValueFromString` finds it), or ILLEGAL_ENUM_VALUE.
     */
    export function rangeToValue(state: RulesState, enumName: string, indexes: readonly number[]): number

    /** Go: `e.ValueToRange(k)`: the indexes of key k. Null where Go returns nil (a key the enum lacks). */
    export function valueToRange(state: RulesState, enumName: string, key: number): number[] | null

    /**
     * Go: `g.NeighborsByKey(k)` on a graph from MustNewEnumGridConnectedness:
     * the keys of the (up to eight) adjacent cells that pass every direction
     * filter, or null where Go returns nil (no neighbors, or a key the enum
     * lacks). Go returns them in a map's iteration order, which differs run to
     * run; they are returned ascending here, and Go code whose result depends on
     * the order has no single answer to match.
     */
    export function neighborsByKey(graph: GraphRef, key: number): number[] | null

    /**
     * Go: `m.<Stock>.Apply(state)` inside a move's own Apply: the embedded
     * stock move's apply, which the loader resolves from MoveData.Stock.
     */
    export function stockApply(state: RulesState, m: MoveInstance): string | null

    /** Go: `legal.Errorf(key, bindings)`: a failing verdict with that template and bindings. */
    export function legalError(key: string, bindings?: Readonly<Record<string, LegalBindingValue>>): LegalVerdict

  go/num.ts

    // Go's integer arithmetic where JavaScript's differs. `int` is a number;
    // see fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ". `+`, `-`,
    // comparisons and `%` (which takes the dividend's sign in both languages)
    // are emitted as themselves; these are the operations that are not.
    //
    // Every result is normalized so that JavaScript's -0 never escapes: Go has
    // no negative zero, and -0 would serialize as 0 but compare unequal under
    // Object.is.

    /** The text imul throws when a product leaves the safe-integer range. */
    export const INT_OVERFLOW_PREFIX = 'go/num: integer overflow: ';

    /**
     * Go: `a / b` on ints. Truncates toward zero (Math.trunc(a / b)); throws
     * DIVIDE_BY_ZERO_TEXT (panic.ts) when b is 0.
     */
    export function idiv(a: number, b: number): number

    /** Go: `a % b` on ints. The same as JavaScript's `%`, plus the zero check of idiv. */
    export function imod(a: number, b: number): number

    /**
     * Go: `a * b` on ints. Throws at once when the product is not a safe
     * integer (Number.isSafeInteger): 64-bit wraparound is not emulated, and a
     * game that relies on it is out of scope.
     */
    export function imul(a: number, b: number): number

  go/str.ts

    // Go's strings: package strings and strconv, and the string semantics
    // that differ from JavaScript's. Go strings are UTF-8 bytes: `<` orders by
    // bytes and `len` counts them. Indexing or slicing a string is refused by
    // the translator (string-bytes), so nothing here does it.
    //
    // Every answer here is held to Go's by testdata/gen/facadevec. Where Go's
    // answer depends on Unicode tables (ToLower, Quote's notion of printable),
    // the JavaScript engine's tables are used; the vectors cover the cases games
    // write (ASCII, accented letters, Greek, emoji).

    /** Go: `strings.Join(elems, sep)`. A nil slice joins to ''. */
    export function join(elems: readonly string[] | null, sep: string): string

    /** Go's panic text for strings.Repeat with a negative count. */
    export const REPEAT_NEGATIVE_TEXT = 'strings: negative Repeat count';

    /** Go: `strings.Repeat(s, count)`. Panics with Go's text on a negative count. */
    export function repeat(s: string, count: number): string

    /** Go: `strings.Split(s, sep)`. An empty sep splits after each UTF-8 sequence. */
    export function split(s: string, sep: string): string[]

    /**
     * Go: `strings.ToLower(s)`, with Go's Unicode case mapping: rune by rune,
     * with no context (a word-final capital sigma becomes σ, not ς) and one
     * rune for one rune (İ becomes i, not i plus a combining dot).
     */
    export function toLower(s: string): string

    /** Go: `strings.TrimSpace(s)`, with Go's definition of white space. */
    export function trimSpace(s: string): string

    /** Go: `strings.HasPrefix(s, prefix)`. */
    export function hasPrefix(s: string, prefix: string): boolean

    /** Go: `strings.HasSuffix(s, suffix)`. */
    export function hasSuffix(s: string, suffix: string): boolean

    /** Go: `strings.Contains(s, substr)`. */
    export function contains(s: string, substr: string): boolean

    /**
     * Go: `<`, `<=`, `>`, `>=` on strings, and `strings.Compare`. -1, 0 or 1,
     * ordering by code point, which is the order of the UTF-8 bytes.
     * (JavaScript's own `<` compares UTF-16 units and differs outside the basic
     * plane.)
     */
    export function strCompare(a: string, b: string): number

    /** Go: `len(s)` on a string: its length in UTF-8 bytes. */
    export function utf8Len(s: string): number

    /** Go: `strconv.Itoa(n)`. */
    export function itoa(n: number): string

    /**
     * Go: `strconv.Atoi(s)`, as a tuple: the value and null, or 0 and Go's
     * error text (`strconv.Atoi: parsing "x": invalid syntax`, `... value out
     * of range`). Syntax and range are decided in Go's order: the digits are
     * scanned left to right and the first failure wins.
     *
     * DIVERGENCE (as legal/goformat.ts goAtoi): Go accepts any int64. A number
     * is exact only to 2^53, so an integer that fits int64 but is not a safe
     * integer is reported as out of range rather than rounded. On a range
     * error the value is Go's: math.MaxInt64 or math.MinInt64, which as numbers
     * are 2^63 and -2^63.
     */
    export function atoi(s: string): [number, string | null]

    /** Go: `strconv.Quote(s)`, which is also `%q` of a string. */
    export function quote(s: string): string

    /** Go: `strconv.QuoteRune(r)`, which is also `%q` of an integer. */
    export function quoteRune(r: number): string

    /**
     * Go: `fmt.Sprintf(format, args...)` and the text of `fmt.Errorf`, for the
     * verbs `%d %s %v %q` with the flags `-`, `+`, `0` and a width, and `%%`.
     * Go's own markers for a bad call are reproduced: `%!d(string=x)` for a
     * verb that does not fit, `%!d(MISSING)`, `%!(EXTRA int=1)`, `%!(NOVERB)`.
     *
     * A number is an int when it is an integer and a float64 otherwise (Go's
     * static type is not visible here). null prints as Go's nil interface,
     * `<nil>`; a nil slice argument would print `[]` in Go.
     */
    export function sprintf(format: string, ...args: readonly unknown[]): string

  go/slice.ts

    // Go's slices, nil-aware. A Go `[]T` is `T[] | null`: `var x []T` and
    // `x = nil` are null (serialized as `null`), `make([]T, 0)` and `[]T{}` are
    // `[]` (serialized as `[]`). `append` is allowed only as
    // `x = append(x, ...)`, so a result never aliases another slice.
    //
    // Capacity is not modeled: a slice's capacity is its length. Go code whose
    // result depends on spare capacity (re-slicing past len) is refused by the
    // translator (slice-alias).
    //
    // See fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ".

    /**
     * Go: `xs = append(xs, values...)`. `appendTo(null, v)` is `[v]`;
     * `appendTo(null)` (nothing appended) is null, as Go's is nil. Emitted as
     * `xs = appendTo(xs, ...)`. The result is a new array, so another variable
     * still holding the old slice keeps its old length, as in Go.
     */
    export function appendTo<T>(xs: T[] | null, ...values: readonly T[]): T[] | null

    /** Go: `len(xs)` on a slice. `lenOf(null)` is 0. */
    export function lenOf(xs: readonly unknown[] | null): number

    /**
     * Go: `xs[i]` read with a non-constant index. Throws Go's index panic
     * (panic.ts indexOutOfRange) when i is outside 0 .. len(xs)-1; a null
     * slice has length 0.
     */
    export function at<T>(xs: readonly T[] | null, i: number): T

    /** Go: `xs[i] = v` with a non-constant index. Throws Go's index panic as `at` does. */
    export function setAt<T>(xs: T[] | null, i: number, value: T): void

    /**
     * Go: `make([]T, n)`: n copies of T's zero value. Never null. Throws Go's
     * `makeslice: len out of range` for a negative n.
     */
    export function makeSlice<T>(n: number, zero: T): T[]

    /**
     * Go: `copy(dst, src)`: copies min(len(dst), len(src)) elements and returns
     * that count. dst and src may be the same array, as in Go.
     */
    export function copyInto<T>(dst: T[] | null, src: readonly T[] | null): number

    /**
     * Go: `xs[low:high]`, read only (the translator refuses a slice expression
     * that is written or appended to, slice-alias). Throws Go's slice-bounds
     * panic when out of range. A nil slice sliced to [0:0] stays nil, as in Go.
     */
    export function subslice<T>(xs: readonly T[] | null, low: number, high: number): readonly T[] | null

  go/panic.ts

    // Go's run-time panics, as the facade throws them. A translated function
    // that would panic in Go throws an Error whose message is exactly Go's text,
    // so a recorded panic (a legality ledger's unknown entry, say) compares
    // equal. The texts are the Go runtime's own (runtime/error.go, panic.go),
    // checked against recovered panics by testdata/gen/facadevec.
    //
    // See fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ".

    /** A method call or field read through a nil pointer. */
    export const NIL_DEREF_TEXT = 'runtime error: invalid memory address or nil pointer dereference';

    /** Integer division or remainder by zero. */
    export const DIVIDE_BY_ZERO_TEXT = 'runtime error: integer divide by zero';

    /**
     * The start of an index panic. The whole text is
     * `runtime error: index out of range [i] with length n`, or, for a negative
     * index, `runtime error: index out of range [i]`.
     */
    export const INDEX_OUT_OF_RANGE_PREFIX = 'runtime error: index out of range [';

    /**
     * The start of a slice-expression panic:
     * `runtime error: slice bounds out of range [:j] with capacity n`, and the
     * other forms Go's runtime spells.
     */
    export const SLICE_BOUNDS_PREFIX = 'runtime error: slice bounds out of range [';

    /** An assignment to an entry of a nil map. */
    export const NIL_MAP_WRITE_TEXT = 'assignment to entry in nil map';

    /** `make([]T, n)` with n negative (runtime.panicmakeslicelen). */
    export const MAKESLICE_LEN_TEXT = 'runtime error: makeslice: len out of range';

    /** Go: the panic of `*p` or `p.M()` with p nil. Throws NIL_DEREF_TEXT. */
    export function nilDeref(): never

    /**
     * Go: using an interface or pointer value that may be nil where Go would
     * dereference it (a nil Stack's method, a nil Move passed to a timer, a nil
     * error's Error()). x itself, or Go's nil-dereference panic. Added for the
     * translator (workstream 10).
     */
    export function nonNil<T>(x: T | null): T

    /** Go: the panic of `a / 0` or `a % 0`. Throws DIVIDE_BY_ZERO_TEXT. */
    export function divideByZero(): never

    /** Go's text for indexing a slice of `length` at `index`. */
    export function indexOutOfRangeText(index: number, length: number): string

    /** Go: the panic of `xs[index]` out of range. Throws indexOutOfRangeText(index, length). */
    export function indexOutOfRange(index: number, length: number): never

    /**
     * Go's text for `xs[low:high]` out of range on a slice of capacity
     * `capacity`, or '' when the bounds are legal. The runtime checks the high
     * bound against the capacity first, then the low bound against the high
     * one, both as unsigned comparisons (so a negative bound fails its check);
     * a negative bound is reported without the other number.
     */
    export function sliceBoundsText(low: number, high: number, capacity: number): string

    /** Go: the panic of `xs[low:high]` out of range. Throws sliceBoundsText(low, high, capacity). */
    export function sliceBoundsOutOfRange(low: number, high: number, capacity: number): never

  The engine's libraries (lib/*.ts). A translation reaches them through the
  facade above (diceRoll, rangeValue, neighborsByKey, ...); their own names
  are not importable (see OUTPUT). For reference:

  lib/dice.ts

    // components/dice: die components with any faces, and the roll that draws
    // from the state's generator. A die is a component whose static values are
    // `{ Faces }` and whose dynamic values are `{ Value, SelectedFace,
    // RollCount }` (dice.Value and dice.DynamicValue).

    /** dice.Value: a die's static values. */
    export interface DieValues {
    	Faces: number[];
    }

    /** The dynamic value properties dice.DynamicValue declares, to spread into a deck's dynamicComponentValues props. */
    export const dieDynamicProps: Readonly<Record<string, PropDef>> = {
    	Value: { Type: 'int' },
    	SelectedFace: { Type: 'int' },
    	RollCount: { Type: 'int' },
    };

    /** dice.BasicDie(min, max): a face for each of min through max; null (Go's nil) unless min < max. */
    export function basicDie(min: number, max: number): DieValues | null

    /** dice.DefaultDie(): a six-sided die. */
    export function defaultDie(): DieValues

    /** (*dice.Value).Min(): the lowest face; math.MaxInt64 for a die with none. */
    export function dieMin(values: { readonly Faces: readonly number[] }): number

    /** (*dice.Value).Max(): the highest face; math.MinInt64 for a die with none. */
    export function dieMax(values: { readonly Faces: readonly number[] }): number

    /**
     * (*dice.DynamicValue).Roll(r): one draw of r.Intn(len(Faces)), which
     * becomes SelectedFace; Value becomes that face; RollCount goes up by one
     * even when the face is the one already showing.
     */
    export function rollWith(rand: GoRand, values: { readonly Faces: readonly number[] }, dynamic: Record<string, unknown>): void

    /**
     * `c.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())` for the die
     * at `index` of `deck`: draws from the state's generator, as Go does.
     */
    export function roll(state: RulesState, deck: string, index: number): void

  lib/enumrange.ts

    // Range enums (enum.RangeEnum, enum/range.go): an enum whose values are the
    // cells of a multi-dimensional grid, each value's string being its indexes
    // joined by "_" ("0_0", "0_1", ... for AddRange("spaces", 8, 8)). The core
    // and the rules layer store such an enum as an ordinary EnumDef with
    // `Range` set to its dimensions; these functions are the Go methods over it.
    //
    // Every function is a port of the Go method of the same name and keeps its
    // odd corners: RangeToValue goes through ValueFromString (so it also accepts
    // the decimal key of a one-dimensional enum), ValueToRange of a key the enum
    // lacks is nil (null here), and an index outside the grid is IllegalValue.

    /**
     * enum.IllegalValue, math.MaxInt64. A JavaScript number cannot hold it
     * exactly; this literal is the nearest double (2^63), which is also what a
     * Go vector's 9223372036854775807 parses to, so comparisons agree.
     */
    export const ILLEGAL_VALUE = 9223372036854775807;

    /**
     * enum.Set.AddRange: the EnumDef of a range enum with these dimensions.
     * Throws Go's error text for no dimensions or a dimension below 1.
     */
    export function rangeEnumDef(dimensions: readonly number[]): EnumDef

    /** Whether the enum has this key (Enum.Valid). */
    export function enumValid(def: EnumDef, key: number): boolean

    /** Enum.String(key): '' for a key the enum lacks. */
    export function enumString(def: EnumDef, key: number): string

    /**
     * Enum.ValueFromString: the key whose string matches after trimming and
     * lower-casing, else the key whose decimal form `text` is, else
     * IllegalValue. Go ranges over a map for the first rule, so two strings
     * that normalize alike have no fixed answer there; the lowest key is taken.
     */
    export function enumValueFromString(def: EnumDef, text: string): number

    /** RangeEnum.RangeDimensions(); null for an enum that is not a range enum. */
    export function rangeDimensions(def: EnumDef): number[] | null

    /**
     * RangeEnum.ValueToRange(key): the indexes of the cell, one per dimension.
     * Null (Go's nil) for a key the enum lacks.
     */
    export function valueToRange(def: EnumDef, key: number): number[] | null

    /** RangeEnum.RangeToValue(indexes...): the key of that cell, or ILLEGAL_VALUE. */
    export function rangeToValue(def: EnumDef, indexes: readonly number[]): number

    /**
     * RangeVal.SetRangeValue(indexes...)'s outcome: Go's error text, or null.
     * The key it sets on success is rangeToValue(def, indexes).
     */
    export function setRangeValueError(def: EnumDef, indexes: readonly number[]): string | null

    /** ImmutableRangeVal.RangeValue() of a value set to `key`. */
    export function rangeValue(def: EnumDef, key: number): number[] | null

  lib/graph.ts

    // Graphs over enum values (enum/graph): the grid connectedness a board game
    // builds once at package level (graph.MustNewEnumGridConnectedness) and the
    // adjacency queries on it. A graph here is data: each node's neighbors,
    // sorted, so it can be built when asked for and never needs to be kept.
    //
    // Go stores a node's edges in a map and Neighbors ranges over it, so Go's
    // order of neighbors is unspecified and changes from run to run. The port
    // returns them ascending; any Go caller whose result depends on the order
    // has no fixed answer to reproduce. Go returns nil for a node with no edges;
    // the port returns [].

    /** A finished directed graph whose nodes are an enum's keys. */
    export interface EnumGraph {
    	/** Node key (decimal) to its neighbors, ascending. A node with no edges is absent. */
    	readonly edges: Readonly<Record<string, readonly number[]>>;
    }

    /** The Direction* edge filters of enum/graph/grid.go. */
    export type Direction = 'up' | 'down' | 'left' | 'right' | 'perpendicular' | 'diagonal';

    /**
     * A direction by any of the names a caller may carry: the Go identifier
     * ("graph.DirectionDown", "DirectionDown") or the short form ("down").
     * Throws for anything else.
     */
    export function parseDirection(name: string): Direction

    /**
     * graph.NewGridConnectedness(e, filters...): every cell connected to each of
     * its up to 8 neighbors that passes every filter, in order. `def` must be a
     * two-dimensional range enum; otherwise Go's error is thrown.
     */
    export function gridConnectedness(def: EnumDef, directions: readonly (Direction | string)[] = []): EnumGraph

    /** EnumGraph.NeighborsByKey(start): ascending; [] where Go returns nil. */
    export function neighborsByKey(graph: EnumGraph, start: number): number[]

    /** EnumGraph.ConnectedByKey(from, to). */
    export function connectedByKey(graph: EnumGraph, from: number, to: number): boolean

    /** Whether every key of a graph's edges is a value of `def` (a sanity check for loaded data). */
    export function graphMatchesEnum(graph: EnumGraph, def: EnumDef): boolean

  lib/playingcards.ts

    // components/playingcards: a deck of American playing cards. A card's static
    // values are `{ Suit, Rank }`, each an enum value stored as its string, with
    // the enums "Suit" and "Rank" (playingcards' own enum set, auto_enum.go).
    // Deck order is Go's deckCanonicalOrder: suits spades, hearts, clubs,
    // diamonds; within each, ace through king; then two jokers if asked for.

    export const SUIT_UNKNOWN = 0;

    export const SUIT_SPADES = 1;

    export const SUIT_HEARTS = 2;

    export const SUIT_CLUBS = 3;

    export const SUIT_DIAMONDS = 4;

    export const SUIT_JOKERS = 5;

    export const RANK_UNKNOWN = 0;

    export const RANK_ACE = 1;

    export const RANK_2 = 2;

    export const RANK_3 = 3;

    export const RANK_4 = 4;

    export const RANK_5 = 5;

    export const RANK_6 = 6;

    export const RANK_7 = 7;

    export const RANK_8 = 8;

    export const RANK_9 = 9;

    export const RANK_10 = 10;

    export const RANK_JACK = 11;

    export const RANK_QUEEN = 12;

    export const RANK_KING = 13;

    export const RANK_JOKER = 14;

    /** playingcards.SuitEnum, named "Suit". */
    export function suitEnum(): EnumDef

    /** playingcards.RankEnum, named "Rank". */
    export function rankEnum(): EnumDef

    /** Both enums under their Go names, for a GameSpec's componentEnums. */
    export function cardEnums(): Record<string, EnumDef>

    /** A card's static values: the enums' strings. */
    export interface Card {
    	Suit: string;
    	Rank: string;
    }

    /** The static values for a card of this suit and rank key. */
    export function card(suit: number, rank: number): Card

    /** playingcards.NewDeck(withJokers): the components, in deck order. */
    export function newDeck(withJokers: boolean): Card[]

    /** playingcards.NewDeckMulti(count, withJokers): `count` decks (at least one) end to end. */
    export function newDeckMulti(count: number, withJokers: boolean): Card[]

    /** The deck's generic values (Deck.SetGenericValues): unknown suit and rank. */
    export function genericCard(): Card

    /** The key of a rank string, or RANK_UNKNOWN for one the enum lacks. */
    export function rankKey(rank: string): number

    /** The key of a suit string, or SUIT_UNKNOWN for one the enum lacks. */
    export function suitKey(suit: string): number

    /** (*playingcards.Card).String(): "<suit> <rank>". */
    export function cardString(c: Card): string


SEMANTICS
  From fabric/TRANSPILE-PROPOSAL.md, "Types" and "Semantics that differ", verbatim.

  | Go | TypeScript |
  |---|---|
  | `int` | `number`, checked to be a safe integer when stored |
  | `bool`, `string`, `float64` | `boolean`, `string`, `number` |
  | `boardgame.PlayerIndex` | `number` |
  | `enum.EnumKey`, and an enum value read from state | `number` (the key); the facade converts to and from the stored string |
  | `*gameState` (the game's game sub-state) | no value: property access goes through `state` |
  | `*playerState` | `number`, the player's index |
  | A dynamic component values struct | a handle `{ deck, index }` |
  | `boardgame.Stack`, `SizedStack`, `ImmutableStack`, `MergedStack` | the facade's `StackRef` |
  | `boardgame.ComponentInstance`, immutable or not | `Comp \| null`: `{ stack, slot, deck, index }` |
  | `boardgame.Timer` | the engine's `TimerHandle` |
  | `boardgame.Move` | `MoveInstance` |
  | `error` | `string \| null` |
  | `[]T` for the scalar types above and `[]*playerState` | `T[] \| null`, nil-aware |
  | `map[K]V` with K an integer or string type | `Map<K, V>`, index and assign only |
  | Multiple results | a tuple array, destructured at the call |
  | `time.Duration` constant | whole milliseconds |

  `concreteStates(state)` and any function whose body only type-asserts
  sub-states is erased: its results are the game handle and the list of
  player indexes.

  Refused types: fixed arrays, struct values that are copied or mutated,
  pointers to locals, sized integers (`int8`, `uint32`, ...), `float32`,
  `complex`, channels, function values other than a `less` callback or a
  literal passed straight to a known call, interfaces outside the table,
  type parameters.

  | Go | Handling |
  |---|---|
  | Integer division truncates toward zero; `%` takes the dividend's sign; division by zero panics | `idiv(a, b)` is `Math.trunc(a / b)` and throws Go's `runtime error: integer divide by zero`. `%` is the same in both languages; `imod` adds the zero check |
  | `int` is 64 bits and wraps | Not emulated. Every integer written to state or to a move field is checked with `Number.isSafeInteger` when the version is saved, and the engine refuses the move with its own text. `*` goes through `imul`, which checks at once. A game that relies on wraparound is out of scope and the lint says so |
  | Map iteration order is random | `range` over a map is refused (`map-range`). Go code whose result depends on the order has no single answer for a port to match |
  | `nil` | A nil component is `null`. Calling a method on it goes through the facade, which throws Go's nil-dereference text. Go code that compares to `nil` first translates to a `null` comparison. A panic inside a declared plan becomes an unknown ledger entry with that text, as in Go; outside one it is a refusal, where Go takes the process down (engine-rules divergence 5) |
  | Index out of range panics | Slice and player indexing by a non-constant goes through `at(xs, i)` or `player(state, i)`, which throw `runtime error: index out of range [i] with length n` |
  | A nil slice serializes as `null`, an empty one as `[]` | `var x []T` and `x = nil` are `null`; `append(null, v)` is `[v]`; `make([]T, 0)` and `[]T{}` are `[]`. `len(null)` is 0 and ranging over it does nothing |
  | Slices alias their backing array | `append` is allowed only as `x = append(x, ...)`. A slice expression `a[i:j]` is allowed only where the result is read and never written or appended to. Anything else is refused (`slice-alias`) |
  | `Components()` is a snapshot | `components(s)` returns an array of handles taken before the loop. A handle whose component has since moved fails as Go's does, with the engine's "not public information" text (engine-core divergence 3 is the same mechanism) |
  | A stack operation returns an error and changes nothing on failure | Facade operations return `string \| null`. A discarded result is discarded; `if err := ...; err != nil { return ... }` is the same shape in TypeScript |
  | An `error` result, by role | In `Apply` and set-up, returning a non-nil error becomes `throw`; the engine discards the state. In `Legal`, `LegalCustom` and `TurnDone` it is the returned string. In a helper it stays a returned `string \| null` |
  | `sort.Sort`, `sort.Slice` are unstable but deterministic | `gosort.ts` ports `sort.Sort`, including the order among equal elements. `sort.Slice` is generated from the same template in the standard library; that its order among equals matches was not checked and needs a vector before it is mapped. `sort.Stable` and `sort.SliceStable` map to `Array.prototype.sort`, which is stable. `less` must return what Go's returns for every pair |
  | Strings compare by UTF-8 bytes; `len` counts bytes | `<` on strings goes through `strCompare`, which orders by code point (the same order as UTF-8 bytes; JavaScript's own `<` differs for characters outside the basic plane, which memory's cards are). `len(s)` is `utf8Len(s)`. Indexing or slicing a string is refused (`string-bytes`) |
  | `state.Rand()` is the only legal source | `state.rand()`, in Go's call order. Go evaluates call arguments left to right, as JavaScript does. Any other `*rand.Rand` or the package-level functions are refused (`foreign-rand`); the existing lint already flags the import. Starting a timer draws 16 values (engine-core divergence 13); the facade's `start` does |
  | Enum values are integers | Stored as strings; the facade converts. `ValueFromString` on names that normalize alike has no single answer (engine-legal finding 9) and is refused |
  | `time.Duration` is nanoseconds | Folded to milliseconds; a constant that is not a whole number of milliseconds is refused (`sub-ms-duration`) |
  | Loop variables are per iteration (the module is on Go 1.26) | `let` in a `for` head has the same meaning |
  | Wall clock, I/O, logging | `time.Now`, `log`, `os`, `encoding/json` are refused (`io`) |


EXAMPLES
  From the checked-in generated games (server/static/src/engine/games/<game>/bodies.gen.ts), until boardgame-util/lib/transpile/testdata/pairs holds pairs, never from checkers itself; nearest refusal reason first, then role, then Go calls shared.

  Example 1: blackjack.handValue (game blackjack, role helper, producer transpile@1, refusals none (translated mechanically))

    Go:

      func handValue(components []boardgame.ImmutableComponentInstance) int {
      	var numUnconvertedAces int
      	var currentValue int

      	for _, c := range components {
      		card := c.Values().(*playingcards.Card)
      		switch card.Rank.Value() {
      		case playingcards.RankAce:
      			numUnconvertedAces++
      			//We count the ace as 1 now. Later we'll check to see if we can
      			//expand any aces.
      			currentValue++
      		case playingcards.RankJack, playingcards.RankQueen, playingcards.RankKing:
      			currentValue += 10
      		default:
      			currentValue += card.Rank.Value().Int()
      		}
      	}

      	for numUnconvertedAces > 0 {

      		if currentValue >= (targetScore - 10) {
      			break
      		}

      		numUnconvertedAces--
      		currentValue += 10
      	}

      	return currentValue
      }

    TypeScript:

      export function blackjack_handValue(state: RulesState, components_: (Comp | null)[] | null): number {
      	let numUnconvertedAces: number = 0;
      	let currentValue: number = 0;
      	{
      		const $s2 = components_;
      		for (let $i1 = 0; $i1 < lenOf($s2); $i1++) {
      			let c: Comp | null = $s2![$i1];
      			let card: DynamicRef = valuesOf(state, c);
      			switch (enumKey(state, { valuesOf: card }, 'Rank')) {
      				case 1: {
      					numUnconvertedAces++;
      					currentValue++;
      					break;
      				}
      				case 11:
      				case 12:
      				case 13: {
      					currentValue = currentValue + 10;
      					break;
      				}
      				default: {
      					currentValue = currentValue + enumKey(state, { valuesOf: card }, 'Rank');
      					break;
      				}
      			}
      		}
      	}
      	for (; numUnconvertedAces > 0;) {
      		if (currentValue >= 11) {
      			break;
      		}
      		numUnconvertedAces--;
      		currentValue = currentValue + 10;
      	}
      	return currentValue;
      }

  Example 2: blackjack.(*gameDelegate).DistributeComponentToStarterStack (game blackjack, role distribute, producer transpile@1, refusals none (translated mechanically))

    Go:

      func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

      	game, _ := concreteStates(state)

      	card := c.Values().(*playingcards.Card)

      	if card.Rank.Value() == playingcards.RankJoker {
      		return game.UnusedCards, nil
      	}

      	return game.DrawStack, nil

      }

    TypeScript:

      export function blackjack_gameDelegate_DistributeComponentToStarterStack(state: RulesState, deck: string, index: number): StackRef {
      	const c: DynamicRef = { deck, index };
      	let card: DynamicRef = valuesOf(state, c);
      	if (enumKey(state, { valuesOf: card }, 'Rank') === 14) {
      		return gameStack(state, 'UnusedCards');
      	}
      	return gameStack(state, 'DrawStack');
      }

  Example 3: blackjack.(*moveIncrementRoundsCompleted).Legal (game blackjack, role legal, producer transpile@1, refusals none (translated mechanically))

    Go:

      func (m *moveIncrementRoundsCompleted) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
      	if err := m.FixUp.Legal(state, proposer); err != nil {
      		return err
      	}
      	game, _ := concreteStates(state)
      	// Legal only if no player has cards and flags are reset (cleanup already done)
      	for _, p := range state.ImmutablePlayerStates() {
      		player := p.(*playerState)
      		if player.HiddenHand.NumComponents() > 0 || player.VisibleHand.NumComponents() > 0 {
      			return errors.New("cards haven't been collected yet")
      		}
      	}
      	// Check that we haven't already incremented (by looking at whether we're
      	// still in the cleanup phase — the StartPhase move after us will change it)
      	if game.Phase.Value() != phaseRoundCleanup {
      		return errors.New("not in cleanup phase")
      	}
      	return nil
      }

    TypeScript:

      export function blackjack_moveIncrementRoundsCompleted_Legal(state: RulesState, m: MoveInstance, proposer: number, base: () => string | null): string | null {
      	{
      		let err: string | null = base();
      		if (err !== null) {
      			return err;
      		}
      	}
      	{
      		const $s2 = playerIndexes(state);
      		for (let $i1 = 0; $i1 < lenOf($s2); $i1++) {
      			let p: number = $s2![$i1];
      			let player_: number = p;
      			if ((stackView(playerStack(state, player_, 'HiddenHand')).numComponents() > 0) || (stackView(playerStack(state, player_, 'VisibleHand')).numComponents() > 0)) {
      				return 'cards haven\'t been collected yet';
      			}
      		}
      	}
      	if (enumKey(state, 'game', 'Phase') !== 3) {
      		return 'not in cleanup phase';
      	}
      	return null;
      }

OUTPUT
  One file, server/static/src/engine/games/checkers/checkers_gameDelegate_ConfigurePredicateConstructors.ts, containing exactly:
    export function checkers_gameDelegate_ConfigurePredicateConstructors(spec: PredicateSpec): GamePredicate | string ... }
  That is the body calling convention of rules/data.ts for role predicate:
    | predicate | Predicates[i].Construct | `(spec: PredicateSpec) => GamePredicate \| string`, THROWS (rules/types.ts; details below) |
    - predicate: Go's ConfigurePredicateConstructors entry's Constructor.
      `spec` is `{ name, args, message }` (Go's legal.Spec; args [] for nil,
      message '' for none). It returns Go's *legal.Predicate as
      `{ name, args?, reads, requiredReadTypes?, allowedReadTypes?, cost?,
      usesProposer?, emittedTemplates?, emittedBindings?, evaluate }`, or
      Go's constructor error (thrown, or returned as a string). `evaluate`
      is `(ctx: PredicateContext) => LegalVerdict`, and the context is all it
      may read: `ctx.state` (the RulesState, for the facade's readers),
      `ctx.move` (the MoveInstance, or null where Go's ctx.Move is nil),
      `ctx.proposer`, and `ctx.resolvePath(path)`, Go's `ctx.ResolvePath`
      as the tuple `[value, PropertyType, err]` (an enum value is
      `{ enum: { name }, value: key, string }` or null). It returns Go's
      verdict shape: passVerdict(), failT(template, bindings?) or
      unknownVerdict(reason) from legal/types.ts. It must not write. A
      throw is Go's panic: the verdict is unknown with Go's
      `predicate "<name>" panicked: <text>` reason. The constructor runs
      once per plan spec naming it, at compile time; Go's `chest` and
      `resolve` arguments are not passed.
  Imports only from ../../go/*.ts, with the .ts extension, of the names
  listed under "Names a translation may import" above (the types and the
  verdict functions there come from ../../rules/*.ts and ../../legal/types.ts),
  plus, from ./bodies.gen.ts, the functions listed under "FUNCTIONS IT
  CALLS". Nothing else: no default or namespace import, no renaming.
  No classes, no module-scope let, no objects of functions, no double casts,
  no Math.random, Date, BigInt, async, or anything outside the engine.

HOW TO CHECK
  boardgame-util port check checkers.(*gameDelegate).ConfigurePredicateConstructors --file server/static/src/engine/games/checkers/checkers_gameDelegate_ConfigurePredicateConstructors.ts
  It type-checks your file, runs the sandbox and purity checks on it, and
  replays the game's fuzz corpus with your function in place. It prints
  PASS, or the first difference: version, viewer, field, the JSON path, what
  Go recorded and what you produced, and the arguments your function was
  last called with.
