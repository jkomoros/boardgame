// The game-definition API: what a game hands the engine. Plain data plus
// functions; there are no base classes, decorators or reflection. A GameSpec
// is what a Go game spreads across its delegate, its state structs with
// their struct tags and embedded behaviors, and its ConfigureMoves.
//
// See ../README.md for the walk-through, with memory as the worked example.

import type { EnumDef, PolicyName, PropDef } from '../definition.ts';
import type { LegalCost, LegalRead, LegalSpec, LegalVerdict, PropertyType } from '../legal/types.ts';
import type { Stack } from '../stack.ts';
import type { ValuePropDef } from './binding.ts';
import type { ProgressionGroup } from './progression.ts';
import type { RulesState } from './state.ts';

// ----- records the host persists -----------------------------------------

/**
 * What identifies one game and never changes after creation. The host
 * persists it beside the state. secretSalt must never reach a client.
 */
export interface GameRecord {
	name: string;
	id: string;
	secretSalt: string;
	numPlayers: number;
	/** Game.Variant(): the effective variant; null when it has no entries. */
	variant: Record<string, string> | null;
	/** One per player; '' for a human. Agents are not run by the engine. */
	agents: string[];
	/**
	 * Whether players will ever be seated (Go: the WillSeatPlayer rendezvous).
	 * With it false, seats are never waited for or inactivated as empty.
	 */
	seating: boolean;
}

/** MoveStorageRecord, without the wall-clock Timestamp. */
export interface MoveRecord {
	Name: string;
	Version: number;
	/** The version of the move that began this causal chain. */
	Initiator: number;
	/** The phase enum key before the move applied (0 with no phase). */
	Phase: number;
	Proposer: number;
	/** The move's fields as Go serializes them. */
	Blob: Record<string, unknown>;
}

// ----- state shape --------------------------------------------------------

export interface DeckSpec {
	/** Each component's static values, by deck index; null for a component with none (a plain token). */
	components: readonly (Readonly<Record<string, unknown>> | null)[];
	/** Declared types of those values. Needed for enum-valued ones. */
	props?: Readonly<Record<string, ValuePropDef>>;
	/**
	 * Deck.SetGenericValues: the values a generic (sanitized) component
	 * shows. Absent or null: none were set. compileGame carries it into the
	 * definition as DeckDef.GenericValues, and `state.component(deck,
	 * GENERIC_COMPONENT)` returns it ({} when none were set; Go's Values()
	 * is then nil).
	 */
	generic?: Readonly<Record<string, unknown>> | null;
	/**
	 * interfaces.LegalComponent: the components' Legal(state, legalType)
	 * method, for moves.DefaultComponent. `index` is the deck index;
	 * `legalType` is the configured legal type's key, or null where Go
	 * passes nil. Null when legal, else the reason. Absent: the components
	 * do not implement it.
	 */
	componentLegal?: (state: RulesState, index: number, legalType: number | null) => string | null;
}

export interface SubStateSpec {
	/** Property name to its type and visibility policy (the Go struct tag). */
	props: Readonly<Record<string, PropDef>>;
	/**
	 * Non-zero starting values (what the Go constructor sets). `index` is the
	 * player index, or the deck index for dynamic component values.
	 */
	initial?: (index: number) => Record<string, unknown>;
	/**
	 * The same as `initial`, as data (GameData.Initial): entry `index` is the
	 * non-zero starting values for that index. For the game sub-state it has
	 * one entry, at 0. For player sub-states, one per player index
	 * 0 .. maxNumPlayers-1. For a deck's dynamic component values, one per
	 * deck index. An index past the end, or an entry with no key, means
	 * every property starts at its zero value. When both are given, `initial`
	 * wins and `initialValues` is ignored.
	 */
	initialValues?: readonly Readonly<Record<string, unknown>>[];
}

/** What Go's embedded game-state behaviors provide. */
export interface GameTraits {
	/** behaviors.DrawDiscardPair: the names of the two stack properties. */
	drawDiscardPair?: { draw: string; discard: string };
	/** moves/interfaces.PlayerOrderer: a custom turn order, or null. */
	playerOrder?: (state: RulesState) => readonly number[] | null;
}

/** interfaces.PlayerTurnFinisher. */
export interface TurnHooks {
	/** Null when the player's turn is done, else why not. */
	turnDone(state: RulesState, player: number): string | null;
	resetForTurnStart(state: RulesState, player: number): void;
	resetForTurnEnd(state: RulesState, player: number): void;
}

/** What Go's embedded player-state behaviors and methods provide. */
export interface PlayerTraits {
	/** behaviors.Seat: bool properties SeatFilled and SeatClosed. */
	seat?: boolean;
	/** behaviors.InactivePlayer: bool property PlayerInactive. */
	inactive?: boolean;
	/** behaviors.PlayerElimination: bool property Eliminated. */
	elimination?: boolean;
	/**
	 * behaviors.PlayerRole: enum property Role (enum "role"), hidden from
	 * other players. Viewers get a computed RoleValue where they may see it.
	 */
	role?: boolean;
	/**
	 * behaviors.PlayerColor: enum property Color (enum "color"). Viewers get
	 * the computed ColorValue where they may see it, and AvailableColors.
	 */
	color?: boolean;
	/** A GameScore() method (behaviors.ScoreBehavior reads property Score). */
	gameScore?: (state: RulesState, player: number) => number;
	turn?: TurnHooks;
}

// ----- variants -----------------------------------------------------------

export interface VariantDisplay {
	displayName?: string;
	description?: string;
}

export interface VariantKeySpec extends VariantDisplay {
	/** The value used when none is given; '' for no default. */
	default?: string;
	values: Readonly<Record<string, VariantDisplay>>;
}

// ----- computed properties ------------------------------------------------

export type ComputedType = 'bool' | 'int' | 'string' | 'playerIndex' | 'intSlice' | 'boolSlice' | 'stringSlice' | 'playerIndexSlice';

export type ComputedSpec =
	| { name: string; scope: 'global'; type: ComputedType; compute: (state: RulesState) => unknown }
	| { name: string; scope: 'player'; type: ComputedType; compute: (state: RulesState, player: number) => unknown };

// ----- moves --------------------------------------------------------------

/** MoveInputDisposition. */
export type InputDisposition = 'required' | 'server-defaulted' | 'context-owned' | 'unsupported';
/** MoveInputCodec. */
export type InputCodec = 'integer' | 'boolean' | 'enum' | 'player-index' | 'string';

export interface MoveFieldSpec {
	type: 'int' | 'bool' | 'string' | 'playerIndex' | 'enum';
	/** For type 'enum': the enum's name. */
	enum?: string;
	/**
	 * Who may see the argument in the move record: a Go `sanitize` tag
	 * ("self:visible"). Absent means hidden from everyone but the admin.
	 */
	sanitize?: string;
	/** How a proposer supplies it. Default 'required'. */
	input?: InputDisposition;
	codec?: InputCodec;
}

/** One move, with its field values. */
export interface MoveInstance {
	readonly name: string;
	fields: Record<string, unknown>;
}

/** MoveChoiceProjection. */
export interface ChoiceSpec {
	field: string;
	/** A stack whose occupied slots are the choices. */
	stack?: { scope: 'game' | 'proposing-player'; property: string };
	/** Enum values to leave out (enum-valued fields only). */
	exclude?: readonly string[];
}

/** What a legality check returns: null for legal, or the reason. */
export type LegalResult = string | null;

export interface MoveSpec {
	name: string;
	helpText: string;
	/** A fix-up move is proposed by the engine, never by a player. */
	fixUp: boolean;
	/**
	 * 'currentPlayer' is moves.CurrentPlayer: the move has a TargetPlayerIndex
	 * field defaulting to the current player, and (for a declared-legality
	 * move) the base check that the proposer is that player.
	 *
	 * 'anyPlayer' is moves.AnyPlayer: a move any seated player makes for
	 * themself. It has a TargetPlayerIndex field that the engine sets to the
	 * proposer; call anyPlayerLegal first in the move's `legal`.
	 */
	base?: 'default' | 'currentPlayer' | 'anyPlayer';
	/** Fields beyond what `base` supplies. */
	fields?: Readonly<Record<string, MoveFieldSpec>>;
	/** DefaultsForState, run after the base's defaults. */
	defaults?: (state: RulesState, move: MoveInstance) => void;

	// Legality. A move has DECLARED legality when it lists preconditions,
	// has a legalCustom, or sets planEnabled; its checks are then a plan
	// whose per-predicate verdicts every viewer receives.
	/** moves.WithLegalPreconditions. */
	preconditions?: readonly LegalSpec[];
	/** moves.WithoutLegalPrecondition: names of base checks to drop. */
	suppress?: readonly string[];
	/** CustomLegaler.LegalCustom: runs last, only if everything passed. */
	legalCustom?: (state: RulesState, move: MoveInstance, proposer: number) => LegalResult | LegalVerdict;
	planEnabled?: boolean;
	/**
	 * Base checks contributed beyond phase, progression, stack constraints
	 * and proposer (what ContributedPreconditions overrides add).
	 */
	contributes?: (base: LegalSpec[]) => LegalSpec[];
	/** moves.WithSourceProperty / WithDestinationProperty. */
	sourceProperty?: string;
	destinationProperty?: string;
	/**
	 * A Legal() override. `base` runs the checks above (moves.Default.Legal);
	 * a move that does not call it skips them, as in Go. Absent: just base.
	 */
	legal?: (state: RulesState, move: MoveInstance, proposer: number, base: () => LegalResult) => LegalResult;

	/** Apply. Throws to refuse; the state is then discarded. */
	apply: (state: RulesState, move: MoveInstance) => void;

	/** moves.WithMoveNameSanitization: who may see the move's name. */
	nameSanitization?: string;
	/** The animation key viewers who may not see the name get instead. */
	hiddenAnimationKey?: string;

	/** interfaces.AllowMultipleInProgression (moves.FixUpMulti). */
	allowMultipleInProgression?: boolean;
	/** moves.WithChoices. */
	choices?: ChoiceSpec;
	/** interfaces.SeatPlayerMover: the move a seat injection applies. */
	seatsPlayer?: boolean;
	/** interfaces.GatheringStartMover. */
	gatheringStart?: boolean;
	/** Ends an ordered phase (a StartPhase move or a NoOp). */
	endsPhaseProgression?: boolean;

	// Set by the installers (addForPhase, addOrderedForPhase), not by hand.
	/** moves.WithLegalPhases: phase enum keys. */
	legalPhases?: readonly number[];
	/** moves.WithLegalMoveProgression. */
	progression?: ProgressionGroup;

	/**
	 * The legal plan Go assembled ((*GameManager).LegalPlanSpecs, carried in
	 * GameData's MoveData.Plan.Specs): base checks, contributions and
	 * authored specs with suppressions applied. When present it is the
	 * move's plan, verbatim: `preconditions`, `contributes` and `suppress`
	 * are ignored, and so are the base checks compileGame would otherwise
	 * derive from `legalPhases`, `progression`, the stack properties and
	 * `base`. Whether the plan's ledger is published still follows
	 * `planEnabled`/`legalCustom`/`preconditions`. The loader sets it only
	 * for a move whose Plan.Declared is true (Go assembles no plan for the
	 * others; their base checks run imperatively in Default.Legal).
	 */
	plan?: readonly LegalSpec[];

	/**
	 * The embedded stock move's callable parts, for a move built by the
	 * loader (rules/data.ts) whose Go type embeds a stock move and
	 * overrides one of its methods. The facade's `stockApply(state, m)`
	 * (Go: `m.<Stock>.Apply(state)`) calls `stock.apply` through
	 * `callStockApply` in rules/data.ts; `stock.legal` is what a Legal
	 * body's `base` argument runs. The engine itself never reads it.
	 */
	stock?: MoveStockParts;
}

/** The callable parts of a stock move (rules/stockRegistry.ts StockParts, without its spec). */
export interface MoveStockParts {
	apply?: (state: RulesState, move: MoveInstance) => void;
	legal?: (state: RulesState, move: MoveInstance, proposer: number, base: () => LegalResult) => LegalResult;
	defaults?: (state: RulesState, move: MoveInstance) => void;
}

// ----- game-registered legal predicates -----------------------------------

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

/**
 * Go: legal.PredicateConstructor.Constructor, for a game's own predicate
 * (ConfigurePredicateConstructors). Returns the predicate, or Go's error:
 * a throw or a returned string, which plan assembly reports as Go does
 * (`boardgame: legal spec "<name>": constructor failed: <text>`). Go's
 * chest and resolve arguments are not passed; no game uses them.
 */
export type GamePredicateConstructor = (spec: PredicateSpec) => GamePredicate | string;

// ----- the game -----------------------------------------------------------

export interface GameSpec {
	name: string;
	/** Default: the name, title-cased. */
	displayName?: string;
	minNumPlayers: number;
	maxNumPlayers: number;
	defaultNumPlayers: number;

	decks: Readonly<Record<string, DeckSpec>>;
	/** The chest's enums. The one named "phase" is the phase enum. */
	enums?: Readonly<Record<string, EnumDef>>;
	/** Enums component values use that the chest does not ship. */
	componentEnums?: Readonly<Record<string, EnumDef>>;
	constants?: Readonly<Record<string, unknown>> | null;

	game: SubStateSpec & { traits?: GameTraits };
	player: SubStateSpec & { traits?: PlayerTraits };
	/** Deck name to its components' dynamic values. */
	dynamicComponentValues?: Readonly<Record<string, SubStateSpec>>;

	variants?: Readonly<Record<string, VariantKeySpec>>;
	computed?: readonly ComputedSpec[];
	/** ConfigureLegalTemplates: message template key to text. */
	legalTemplates?: Readonly<Record<string, string>>;
	/**
	 * ConfigurePredicateConstructors (legal.ConstructorConfigurer): the
	 * game's own predicates, overlaid by name on the built-in catalog (a
	 * name matching a built-in overrides it, as in Go). A plan spec naming
	 * one is built by its constructor once, at compile time.
	 */
	predicates?: readonly { name: string; construct: GamePredicateConstructor }[];

	/** In registration order; fix-ups are tried in this order. */
	moves: readonly MoveSpec[];

	setUp: {
		/** BeginSetUp: before components are distributed. */
		begin?: (state: RulesState, variant: Readonly<Record<string, string>>) => void;
		/** DistributeComponentToStarterStack. */
		distribute: (state: RulesState, deck: string, index: number) => Stack;
		/** FinishSetUp: after every component is in its starter stack. */
		finish?: (state: RulesState) => void;
	};

	/** The end of the game. With none of these, the game never ends. */
	end?: {
		/** GameEndConditionMet; winners are then the top scorers still in play. */
		conditionMet?: (state: RulesState) => boolean;
		/** PlayerScore. Default: the player trait's gameScore, else 0. */
		playerScore?: (state: RulesState, player: number) => number;
		lowScoreWins?: boolean;
		/** A CheckGameFinished override; replaces all of the above. */
		check?: (state: RulesState) => { finished: boolean; winners: number[] };
	};

	/** CurrentPlayerIndex. Default: game.CurrentPlayer, made valid. */
	currentPlayerIndex?: (state: RulesState) => number;
	/** ReadyToStart: null when the gathered players may begin. */
	readyToStart?: (state: RulesState) => string | null;
}

export type { PolicyName };
