Translate one Go function to TypeScript. Its behavior must be identical,
including behavior that looks like a bug: do not fix, simplify or guard
anything the Go does not guard. If the Go would panic, the TypeScript must
throw the same text (see "Panics").

FUNCTION
  memory.(*moveStartHideCardsTimer).Apply, role apply, from examples/memory/moves.go:111 at Go source hash sha256:639e2cc4e8c5299134062a27115a7eb15d4cfc693e6d32ff7e1399c7ce5b66d6

GO SOURCE
  func (m *moveStartHideCardsTimer) Apply(state boardgame.State) error {
  	game, _ := concreteStates(state)

  	move := state.Game().MoveByName(hideCardMoveName)

  	game.HideCardsTimer.Start(hideCardsDuration, move)

  	return nil
  }

WHY THE TRANSLATOR REFUSED IT
  package-var-read memory.hideCardMoveName, at moves.go:114:34
    move := state.Game().MoveByName(hideCardMoveName)

TYPES IT TOUCHES
  The game's types it names, verbatim:

    //boardgame:codegen
    type gameState struct {
    	base.SubState
    	behaviors.CurrentPlayerBehavior
    	CardSet        string
    	NumCards       int
    	HiddenCards    boardgame.SizedStack  `sizedstack:"cards,40" sanitize:"order"`
    	VisibleCards   boardgame.SizedStack  `sizedstack:"cards,40"`
    	Cards          boardgame.MergedStack `overlap:"VisibleCards,HiddenCards"`
    	HideCardsTimer boardgame.Timer
    	//Where cards not in use reside most of the time
    	UnusedCards boardgame.Stack `stack:"cards"`
    }

    // VisibleCards==2 and the static Type mismatch are declarative. Only the
    // timer's runtime state remains in LegalCustom.
    //
    //boardgame:codegen
    type moveStartHideCardsTimer struct {
    	moves.FixUp
    }

  Framework types it names:

    boardgame.Game (github.com/jkomoros/boardgame): its methods are the API table's rows (THE TYPESCRIPT API YOU MAY USE).
    boardgame.Move (github.com/jkomoros/boardgame): its TypeScript shape is in the Types table under SEMANTICS.
    boardgame.State (github.com/jkomoros/boardgame): its TypeScript shape is in the Types table under SEMANTICS.
    boardgame.Timer (github.com/jkomoros/boardgame): its methods are the API table's rows (THE TYPESCRIPT API YOU MAY USE).

  State properties it reads or writes, with their PropDef from the exported definition:

    game.HideCardsTimer: {"Type":"timer","Sanitize":{"all":"visible"}}

  Package variables it reads, verbatim, and every assignment to each:

    var hideCardMoveName string
      assigned in memory.(*gameDelegate).ConfigureMoves at main.go:364:
        hideCardMoveName = hideCardConfig.Name()
        where, in that function:
          hideCardConfig := auto.MustConfig(
          		new(moveHideCards),
          		moves.WithHelpText("After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."),
          		// Declarative migration (Workstream 9 re-migration): Legal() is
          		// deleted (see moves.go); its two gates are now these preconditions,
          		// in the same order the old imperative chain ran (CurrentPlayer's
          		// proposer check is contributed automatically first). PropCompare
          		// with "<=",0 (not PropEquals ==,0: there is no PropAtMost, and == would
          		// wrongly reject a negative value the imperative `>0` treats as legal)
          		// mirrors `if CardsLeftToReveal > 0 { fail }`; StackNotEmpty mirrors
          		// `if VisibleCards.NumComponents() < 1 { fail }`.
          		moves.WithLegalPreconditions(
          			legal.PropCompare("player.CardsLeftToReveal", "<=", 0).WithMessage("hide.cards_still_to_reveal"),
          			legal.StackNotEmpty("game.VisibleCards").WithMessage("hide.no_cards_to_hide"),
          		),
          	)


  Package constants it reads, verbatim (the translator folds a constant to its value):

    const hideCardsDuration = 4 * time.Second
      = 4000000000

  The game's moves, from the exported definition (name: Go type, fix-up, stock type):

    "Seat Player": moves.SeatPlayer, fix-up true, stock moves.SeatPlayer
    "Activate Inactive Player": moves.ActivateInactivePlayer, fix-up true, stock moves.ActivateInactivePlayer
    "Reveal Card": memory.moveRevealCard, fix-up false, stock moves.MoveComponentToSlot
    "Hide Cards": memory.moveHideCards, fix-up false, stock moves.CurrentPlayer
    "Finish Turn": moves.FinishTurn, fix-up true, stock moves.FinishTurn
    "Capture Cards": memory.moveCaptureCards, fix-up true, stock moves.FixUp
    "Start Hide Cards Timer": memory.moveStartHideCardsTimer, fix-up true, stock moves.FixUp

FUNCTIONS IT CALLS (already translated; call them, do not rewrite them)
  (none)

  Erased (no TypeScript function; see the Types table under SEMANTICS):
    memory.concreteStates: its results are the game handle (property access through state.g, gameStack(state, ...)) and the players (playerIndexes(state), a player is its index)

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

    (*boardgame.Game).MoveByName: moveByName (call; proposal)
    (boardgame.ImmutableState).Game: state.Game() (chain; proposal: only as state.Game().Variant(), MoveByName or MoveByNameForState)
    (boardgame.Timer).Start: timer.start (call; proposal)

    The move embeds moves.FixUp. Its parts (the API table's m.<Stock>.X rows):
      m.FixUp.Apply: stockApply
      m.FixUp.DefaultsForState: dropped when it is the first statement
      m.FixUp.Legal: base()

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
  From the checked-in generated games (server/static/src/engine/games/<game>/bodies.gen.ts), until boardgame-util/lib/transpile/testdata/pairs holds pairs, never from memory itself; nearest refusal reason first, then role, then Go calls shared.

  Example 1: debuganimations.(*moveShuffleCards).Apply (game debuganimations, role apply, producer transpile@1, refusals none (translated mechanically))

    Go:

      func (m *moveShuffleCards) Apply(state boardgame.State) error {

      	game, _ := concreteStates(state)

      	return game.FanStack.Shuffle()

      }

    TypeScript:

      export function debuganimations_moveShuffleCards_Apply(state: RulesState, m: MoveInstance): string | null {
      	return shuffle(gameStack(state, 'FanStack'));
      }

  Example 2: blackjack.(*moveIncrementRoundsCompleted).Apply (game blackjack, role apply, producer transpile@1, refusals none (translated mechanically))

    Go:

      func (m *moveIncrementRoundsCompleted) Apply(state boardgame.State) error {
      	game, _ := concreteStates(state)
      	game.RoundsCompleted++
      	return nil
      }

    TypeScript:

      export function blackjack_moveIncrementRoundsCompleted_Apply(state: RulesState, m: MoveInstance): string | null {
      	state.g.RoundsCompleted = (state.g.RoundsCompleted as number) + 1;
      	return null;
      }

  Example 3: debuganimations.(*moveVisibleShuffleCards).Apply (game debuganimations, role apply, producer transpile@1, refusals none (translated mechanically))

    Go:

      func (m *moveVisibleShuffleCards) Apply(state boardgame.State) error {

      	game, _ := concreteStates(state)

      	return game.FanStack.PublicShuffle()

      }

    TypeScript:

      export function debuganimations_moveVisibleShuffleCards_Apply(state: RulesState, m: MoveInstance): string | null {
      	return publicShuffle(gameStack(state, 'FanStack'));
      }

OUTPUT
  One file, server/static/src/engine/games/memory/memory_moveStartHideCardsTimer_Apply.ts, containing exactly:
    export function memory_moveStartHideCardsTimer_Apply(state: RulesState, m: MoveInstance): string | null ... }
  That is the body calling convention of rules/data.ts for role apply:
    | apply | MoveData.Bodies.Apply | `(state: RulesState, m: MoveInstance) => void \| string \| null`, THROWS |
    - apply: Go's `m.<Stock>.Apply(state)` is the facade's `stockApply(state,
      m)`, which returns the stock apply's error text or null. The loader
      stores the stock parts on the MoveSpec (`MoveSpec.stock`), and
      `callStockApply(state, m)` below is what the facade calls.
  Imports only from ../../go/*.ts, with the .ts extension, of the names
  listed under "Names a translation may import" above (the types and the
  verdict functions there come from ../../rules/*.ts and ../../legal/types.ts),
  plus, from ./bodies.gen.ts, the functions listed under "FUNCTIONS IT
  CALLS". Nothing else: no default or namespace import, no renaming.
  No classes, no module-scope let, no objects of functions, no double casts,
  no Math.random, Date, BigInt, async, or anything outside the engine.

HOW TO CHECK
  boardgame-util port check memory.(*moveStartHideCardsTimer).Apply --file server/static/src/engine/games/memory/memory_moveStartHideCardsTimer_Apply.ts
  It type-checks your file, runs the sandbox and purity checks on it, and
  replays the game's fuzz corpus with your function in place. It prints
  PASS, or the first difference: version, viewer, field, the JSON path, what
  Go recorded and what you produced, and the arguments your function was
  last called with.
