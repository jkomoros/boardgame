// GameData: a game's exported definition, as data, and the loader that
// builds a GameSpec from it. The source of truth is the Go struct set in
// boardgame-util/lib/gamedef/schema.go; every type here matches its JSON key
// for key. Change both together, and bump SchemaVersion there.
//
// A field that is optional here (`Name?:`) is omitempty in Go; every other
// field is always present, and is null only where the type says so
// (Constants, Variants, Stock, LegalPhases, Progression, GenericValues): the
// exporter writes every other map and slice non-nil, as {} or [].
// See fabric/TRANSPILE-PROPOSAL.md, "The exported definition".
//
// ## Body calling convention
//
// gameFromData resolves every FnRef the data names through `bodies(ref)`
// and calls the function it gets with exactly the arguments below. A
// translated body (bodies.gen.ts), an agent-written one, and a hand port's
// own function all follow it. `state` is always the RulesState of
// rules/state.ts (the facade's functions take it first); `m` is the
// MoveInstance being judged or applied (`m.fields.X` is Go's `m.X`); a
// player is its index (Go's `*playerState`, `ImmutableSubState`,
// `PlayerIndex`).
//
// Go's error results follow the proposal's "An error result, by role":
// where the signature below says THROWS, a Go `return err` with a non-nil
// err is `throw new Error(text)`; for convenience the loader also accepts
// a returned non-null string there and throws it, so `return text` and
// `return null` (Go's `return nil`) work too. Where it says a `string |
// null` result, the string IS the error and nothing is thrown. A body that
// throws anywhere is a Go panic, and the engine treats it as it treats a
// throw from a hand port.
//
// | Role (FunctionEntry.Role) | Where the FnRef is | Signature |
// |---|---|---|
// | apply | MoveData.Bodies.Apply | `(state: RulesState, m: MoveInstance) => void \| string \| null`, THROWS |
// | legal | MoveData.Bodies.Legal | `(state: RulesState, m: MoveInstance, proposer: number, base: () => string \| null) => string \| null` |
// | legalCustom | MoveData.Bodies.LegalCustom | `(state: RulesState, m: MoveInstance, proposer: number) => string \| null \| LegalVerdict` |
// | defaults | MoveData.Bodies.Defaults | `(state: RulesState, m: MoveInstance) => void` |
// | computed (global) | Computed[i].Compute, Scope 'global' | `(state: RulesState) => unknown` (the value) |
// | computed (player) | Computed[i].Compute, Scope 'player' | `(state: RulesState, player: number) => unknown` |
// | setUpBegin | SetUp.Begin | `(state: RulesState, variant: Readonly<Record<string, string>>) => void \| string \| null`, THROWS |
// | distribute | SetUp.Distribute | `(state: RulesState, deck: string, index: number) => Stack \| { state: RulesState; at: StackRef }`, THROWS |
// | setUpFinish | SetUp.Finish | `(state: RulesState) => void \| string \| null`, THROWS |
// | conditionMet | End.ConditionMet | `(state: RulesState) => boolean` |
// | playerScore | End.PlayerScore | `(state: RulesState, player: number) => number` |
// | checkGameFinished | End.Check | `(state: RulesState) => [finished: boolean, winners: number[] \| null] \| { finished: boolean; winners: number[] }` |
// | currentPlayerIndex | CurrentPlayerIndex | `(state: RulesState) => number` |
// | turnDone | Traits.Player.Turn.TurnDone | `(state: RulesState, player: number) => string \| null` |
// | resetForTurnStart | Traits.Player.Turn.ResetForTurnStart | `(state: RulesState, player: number) => void \| string \| null`, THROWS |
// | resetForTurnEnd | Traits.Player.Turn.ResetForTurnEnd | `(state: RulesState, player: number) => void \| string \| null`, THROWS |
// | gameScore | Traits.Player.GameScore (not 'score') | `(state: RulesState, player: number) => number` |
// | playerOrder | Traits.Game.PlayerOrder | `(state: RulesState) => readonly number[] \| null` |
// | counter | CounterData { Kind: 'custom', Count } | `(currentCount: number, length: number) => string \| null` (Go's ValidCounter) |
// | helper | FunctionEntry only | Never resolved by the loader: bodies call helpers directly, with whatever signature their translation has (`state` first) |
// | predicate | Predicates[i].Construct | Not supported in v1: a non-empty Predicates is a GameSpecError |
//
// Details that the table cannot carry:
//
// - legal: `base` is the embedded stock move's Legal (Go's
//   `m.<Stock>.Legal(state, proposer)`), which for moves.Default and the
//   bases that add nothing is the move's base checks (phase, progression,
//   current player, the plan when one is declared). A body that never
//   calls it skips them, as in Go.
// - apply: Go's `m.<Stock>.Apply(state)` is the facade's `stockApply(state,
//   m)`, which returns the stock apply's error text or null. The loader
//   stores the stock parts on the MoveSpec (`MoveSpec.stock`), and
//   `callStockApply(state, m)` below is what the facade calls.
// - defaults: the stock move's DefaultsForState always runs first (the
//   translator drops `m.<Stock>.DefaultsForState(state)` as the first
//   statement), then the body.
// - distribute: the result may be an engine Stack or a facade StackRef
//   (`{ state, at }`, at an engine StackRef location); the loader turns the
//   latter into the stack of the state it was handed.
// - checkGameFinished: Go's two results as a tuple, the translator's shape
//   for multiple results; the object form is accepted too. A null winners
//   list is empty.
// - Bodies are called with `this` undefined; a body must not rely on it.
//
// ## Hooks beside FnRefs
//
// `bodies` may also answer two keys that are not FnRefs, for moves the
// engine cannot build from a stock type yet, and so that a generated game
// can sit beside hand-ported moves:
//
// - `move:<move name>` (moveOverrideRef): `() => MoveSpec`, a whole move,
//   typically a hand port's. The loader takes the move's BEHAVIOR from it
//   (`apply`, `legal`, `defaults`, `legalCustom`, `sourceProperty`,
//   `destinationProperty`, `contributes`) and its CONFIGURATION from the
//   data (name, help text, fix-up, base, fields, legal phases,
//   progression, plan, flags, name sanitization, choices). The move's
//   Bodies FnRefs and Stock are not resolved.
// - `stock:<move name>` (stockOverrideRef): `() => StockParts`, used in
//   place of `stockMove(Stock.Type, Stock.Config)` for that move; its
//   Bodies are resolved as usual. With this hook a move whose stock type is
//   not ported yet still runs the game's own bodies.

import type { EnumDef, GameDefinition } from '../definition.ts';
import type { LegalSpec, LegalVerdict } from '../legal/types.ts';
import type { Stack } from '../stack.ts';
import type { StackRef } from '../state.ts';
import { gameSpecError } from './compile.ts';
import type { Counter, CounterData, ProgressionGroup, ValidCounter } from './progression.ts';
import { scoreGameScore } from './stock.ts';
import { type StockParts, stockMove } from './stockRegistry.ts';
import { findMove } from './state.ts';
import type { RulesState } from './state.ts';
import type {
	ChoiceSpec, ComputedSpec, ComputedType, DeckSpec, GameSpec, InputCodec, InputDisposition, LegalResult, MoveFieldSpec, MoveInstance, MoveSpec,
	MoveStockParts, SubStateSpec, VariantKeySpec,
} from './types.ts';

export type { CounterData };

/**
 * A function by its Go identity: "blackjack.(*moveCurrentPlayerHit).Apply",
 * "blackjack.handValue". The package part is the package name.
 */
export type FnRef = string;

export interface GameData {
	SchemaVersion: 1;
	Source: { Package: string; Exporter: string };
	Name: string;
	DisplayName: string;
	Description: string;
	MinNumPlayers: number;
	MaxNumPlayers: number;
	DefaultNumPlayers: number;

	/**
	 * Exactly what testdata/gen/vecgame DefinitionFor emits, plus
	 * Decks[name].GenericValues, Decks[name].ComponentProps, Enums[name].Range
	 * and PropDef.Constraints.
	 */
	Definition: GameDefinition;
	Constants: Readonly<Record<string, unknown>> | null;
	ComponentEnums: Readonly<Record<string, EnumDef>>;
	PhaseEnum: string;
	GroupEnum: string;

	/** Constructor output: non-zero starting values. */
	Initial: InitialData;
	Traits: TraitsData;
	/** The conformance header's variantConfig. */
	Variants: unknown;
	Computed: readonly ComputedData[];
	LegalTemplates: Readonly<Record<string, string>>;
	Predicates: readonly PredicateData[];
	/** In registration order. */
	Moves: readonly MoveData[];
	SetUp: SetUpData;
	End: EndData;
	CurrentPlayerIndex?: FnRef;
	Agents: readonly { Name: string; DisplayName: string }[];
	/** Sorted by Func. */
	Functions: readonly FunctionEntry[];
}

export interface InitialData {
	Game: Readonly<Record<string, unknown>>;
	/** One per player index, 0 .. MaxNumPlayers-1. */
	Players: readonly Readonly<Record<string, unknown>>[];
	/** Deck name to one entry per component, by deck index. */
	Dynamic: Readonly<Record<string, readonly Readonly<Record<string, unknown>>[]>>;
}

export interface TraitsData {
	Game: GameTraitsData;
	Player: PlayerTraitsData;
}

export interface GameTraitsData {
	DrawDiscardPair?: { Draw: string; Discard: string };
	RoundRobin?: true;
	PlayerOrder?: FnRef;
}

export interface PlayerTraitsData {
	Seat?: true;
	Inactive?: true;
	Elimination?: true;
	Role?: true;
	Color?: true;
	/** A declared GameScore method, or 'score' for behaviors.ScoreBehavior. */
	GameScore?: FnRef;
	Turn?: { TurnDone: FnRef; ResetForTurnStart: FnRef; ResetForTurnEnd: FnRef };
}

export interface ComputedData {
	Name: string;
	Scope: 'global' | 'player';
	/** A ComputedType of rules/types.ts. */
	Type: string;
	Compute: FnRef;
}

export interface PredicateData {
	Name: string;
	Construct: FnRef;
}

export interface SetUpData {
	Begin?: FnRef;
	Distribute: FnRef;
	Finish?: FnRef;
}

export interface EndData {
	ConditionMet?: FnRef;
	PlayerScore?: FnRef;
	LowScoreWins?: true;
	Check?: FnRef;
}

export interface MoveData {
	Name: string;
	HelpText: string;
	IsFixUp: boolean;
	/** "blackjack.moveCurrentPlayerHit". */
	GoType: string;
	/** The nearest stock ancestor and its options, or null for moves.Default alone. */
	Stock: StockData | null;
	Base: 'default' | 'currentPlayer' | 'anyPlayer';
	Fields: Readonly<Record<string, FieldData>>;
	NameSanitization?: string;
	HiddenAnimationKey?: string;
	LegalPhases: readonly number[] | null;
	Progression: ProgressionData | null;
	AllowMultipleInProgression: boolean;
	/** The plan Go assembled: base checks, contributions, authored specs, suppressions applied. */
	Plan: PlanData;
	/** The header's projection entry. */
	Choices?: unknown;
	SeatsPlayer?: true;
	GatheringStart?: true;
	EndsPhaseProgression?: true;
	/** Present only for a method a game type declares. */
	Bodies: MoveBodies;
}

export interface StockData {
	/** The stock type's name in package moves: "CurrentPlayer", "StartPhase", ... */
	Type: string;
	Config: Readonly<Record<string, unknown>>;
}

export interface FieldData {
	Type: string;
	Enum?: string;
	Sanitize?: string;
	Input: string;
	Codec: string;
}

export interface PlanData {
	Declared: boolean;
	Specs: readonly LegalSpec[];
	HasCustom: boolean;
}

export interface MoveBodies {
	Apply?: FnRef;
	Legal?: FnRef;
	LegalCustom?: FnRef;
	/** DefaultsForState. */
	Defaults?: FnRef;
}

export type ProgressionData =
	| { readonly Kind: 'move'; readonly Name: string }
	| { readonly Kind: 'serial'; readonly Children: readonly ProgressionData[] }
	| { readonly Kind: 'repeat'; readonly Count: CounterData; readonly Child: ProgressionData }
	| { readonly Kind: 'parallel'; readonly Count: CounterData; readonly Children: readonly ProgressionData[] };

export interface FunctionEntry {
	Func: FnRef;
	/** 'apply', 'legal', 'legalCustom', 'defaults', 'computed', 'helper', ... (gamedef Role constants). */
	Role: string;
	File: string;
	Line: number;
	/** "sha256:" and the hex SHA-256 of the function's gofmt-ed source. */
	SourceHash: string;
	/** In-package callees, sorted. */
	Calls: readonly FnRef[];
	Verdict: 'subset' | 'refused';
	Refusals: readonly Refusal[];
}

/** One reason the translator refused a function (lib/transpile/refusal). */
export interface Refusal {
	Func: FnRef;
	/** One of the closed list in lib/transpile/refusal/refusal.go. */
	Reason: string;
	Detail: string;
	File: string;
	Line: number;
	Column: number;
}

/**
 * The game's functions, by FnRef: a function containing a `switch`, since a
 * pattern module may not hold a record of functions. Undefined for a
 * reference it does not know. See "Body calling convention" above for what
 * each returned function is called with, and "Hooks beside FnRefs" for the
 * two non-FnRef keys.
 */
export type Bodies = (ref: string) => ((...args: never[]) => unknown) | undefined;

/** The `bodies` key of the hand-port override for a move: `() => MoveSpec`. */
export function moveOverrideRef(moveName: string): string {
	return 'move:' + moveName;
}

/** The `bodies` key of the stock-parts stand-in for a move: `() => StockParts`. */
export function stockOverrideRef(moveName: string): string {
	return 'stock:' + moveName;
}

/**
 * Go's `m.<Stock>.Apply(state)` for a move the loader built: runs the
 * embedded stock move's apply on `state` and returns its error text, or
 * null. The facade's `stockApply(state, m)` is this. Throws when the move
 * has no stock apply (Go would not compile).
 */
export function callStockApply(state: RulesState, m: MoveInstance): string | null {
	const compiled = findMove(state.game, m.name);
	const apply = compiled?.spec.stock?.apply;
	if (apply === undefined) throw new Error('move ' + m.name + ' embeds no stock move with an Apply');
	try {
		apply(state, m);
	} catch (thrown) {
		if (thrown instanceof Error) return thrown.message;
		throw thrown;
	}
	return null;
}

type AnyBody = (...args: never[]) => unknown;

const FIELD_TYPES: readonly string[] = ['int', 'bool', 'string', 'playerIndex', 'enum'];
const DISPOSITIONS: readonly string[] = ['required', 'server-defaulted', 'context-owned', 'unsupported'];
const CODECS: readonly string[] = ['integer', 'boolean', 'enum', 'player-index', 'string'];
const COMPUTED_TYPES: readonly string[] = ['bool', 'int', 'string', 'playerIndex', 'intSlice', 'boolSlice', 'stringSlice', 'playerIndexSlice'];
const BASES: readonly string[] = ['default', 'currentPlayer', 'anyPlayer'];

/** A role's result where Go returns an error that the engine expects thrown. */
function throwIfError(result: unknown): void {
	if (typeof result === 'string') throw new Error(result);
}

/** The engine stack a distribute body named: an engine Stack, or a facade StackRef. */
function toStack(state: RulesState, result: unknown, ref: string): Stack {
	if (result !== null && typeof result === 'object') {
		const candidate = result as { moveComponent?: unknown; at?: unknown; state?: unknown };
		if (typeof candidate.moveComponent === 'function') return result as Stack;
		if (candidate.at !== null && typeof candidate.at === 'object' && candidate.state !== undefined) return state.engine.stack(candidate.at as StackRef);
	}
	throw new Error(ref + ' did not return a stack');
}

/** The resolver for one load: every FnRef must have a body. */
function newResolver(bodies: Bodies): <F extends AnyBody>(ref: string, role: string) => F {
	return <F extends AnyBody>(ref: string, role: string): F => {
		const body = bodies(ref);
		if (body === undefined) throw gameSpecError('gameFromData: no body for ' + role + ' ' + ref);
		if (typeof body !== 'function') throw gameSpecError('gameFromData: the body for ' + role + ' ' + ref + ' is not a function');
		return body as F;
	};
}

type Resolve = ReturnType<typeof newResolver>;

function fieldsFromData(move: MoveData): Record<string, MoveFieldSpec> {
	const fields: Record<string, MoveFieldSpec> = {};
	for (const name of Object.keys(move.Fields)) {
		const field = move.Fields[name];
		const where = 'move ' + JSON.stringify(move.Name) + ' field ' + name;
		if (FIELD_TYPES.indexOf(field.Type) < 0) throw gameSpecError(where + ': unknown field type ' + JSON.stringify(field.Type));
		if (DISPOSITIONS.indexOf(field.Input) < 0) throw gameSpecError(where + ': unknown input disposition ' + JSON.stringify(field.Input));
		if (field.Codec !== '' && CODECS.indexOf(field.Codec) < 0) throw gameSpecError(where + ': unknown codec ' + JSON.stringify(field.Codec));
		const out: MoveFieldSpec = { type: field.Type as MoveFieldSpec['type'], input: field.Input as InputDisposition };
		if (field.Enum !== undefined && field.Enum !== '') out.enum = field.Enum;
		if (field.Sanitize !== undefined && field.Sanitize !== '') out.sanitize = field.Sanitize;
		if (field.Codec !== '') out.codec = field.Codec as InputCodec;
		fields[name] = out;
	}
	return fields;
}

function counterFromData(counter: CounterData, resolve: Resolve): Counter {
	if (counter.Kind === 'custom') return resolve<ValidCounter>(counter.Count, 'counter');
	return counter;
}

function progressionFromData(group: ProgressionData, moves: Readonly<Record<string, MoveData>>, resolve: Resolve): ProgressionGroup {
	switch (group.Kind) {
		case 'move': {
			const move = moves[group.Name];
			if (move === undefined) throw gameSpecError('progression names move ' + JSON.stringify(group.Name) + ', which the game does not configure');
			return { kind: 'move', name: group.Name, allowMultiple: move.AllowMultipleInProgression };
		}
		case 'serial':
			return { kind: 'serial', children: group.Children.map(child => progressionFromData(child, moves, resolve)) };
		case 'repeat':
			return { kind: 'repeat', count: counterFromData(group.Count, resolve), child: progressionFromData(group.Child, moves, resolve) };
		case 'parallel':
			return { kind: 'parallel', count: counterFromData(group.Count, resolve), children: group.Children.map(child => progressionFromData(child, moves, resolve)) };
		default:
			throw gameSpecError('unsupported progression kind ' + JSON.stringify((group as { Kind: unknown }).Kind));
	}
}

/** The header's choice-projection entry, back to the ChoiceSpec that produces it. */
function choiceFromData(raw: unknown, move: MoveData, enums: Readonly<Record<string, EnumDef>>): ChoiceSpec {
	const where = 'move ' + JSON.stringify(move.Name) + ' choices';
	if (raw === null || typeof raw !== 'object') throw gameSpecError(where + ': not an object');
	const entry = raw as { fieldName?: unknown; source?: unknown; stackSource?: { scope?: unknown; property?: unknown }; candidateValues?: unknown };
	if (typeof entry.fieldName !== 'string') throw gameSpecError(where + ': no fieldName');
	const choice: ChoiceSpec = { field: entry.fieldName };
	switch (entry.source) {
		case 'players':
			break;
		case 'stack-slots': {
			const source = entry.stackSource;
			if (!source || (source.scope !== 'game' && source.scope !== 'proposing-player') || typeof source.property !== 'string') {
				throw gameSpecError(where + ': bad stackSource');
			}
			choice.stack = { scope: source.scope, property: source.property };
			break;
		}
		case 'enum-values': {
			const enumName = move.Fields[entry.fieldName]?.Enum;
			const e = enumName === undefined ? undefined : enums[enumName];
			if (e === undefined) throw gameSpecError(where + ': field ' + entry.fieldName + ' names no known enum');
			const candidates = Array.isArray(entry.candidateValues) ? (entry.candidateValues as unknown[]) : [];
			const exclude = Object.keys(e.Values)
				.map(key => e.Values[key])
				.sort()
				.filter(value => candidates.indexOf(value) < 0);
			if (exclude.length > 0) choice.exclude = exclude;
			break;
		}
		default:
			throw gameSpecError(where + ': unknown source ' + JSON.stringify(entry.source));
	}
	return choice;
}

/** The header's variantConfig (GameManager.Variants()), back to the spec. */
function variantsFromData(raw: unknown): Record<string, VariantKeySpec> | undefined {
	if (raw === null || raw === undefined) return undefined;
	if (typeof raw !== 'object') throw gameSpecError('Variants is not an object');
	type Display = { DisplayName?: string; Description?: string };
	const config = raw as Record<string, Display & { Default?: string; Values?: Record<string, Display> }>;
	const out: Record<string, VariantKeySpec> = {};
	for (const key of Object.keys(config)) {
		const item = config[key];
		const values: Record<string, { displayName?: string; description?: string }> = {};
		for (const name of Object.keys(item.Values ?? {})) {
			const value = (item.Values as Record<string, Display>)[name];
			values[name] = { displayName: value.DisplayName ?? '', description: value.Description ?? '' };
		}
		out[key] = { displayName: item.DisplayName ?? '', description: item.Description ?? '', default: item.Default ?? '', values };
	}
	return out;
}

/** The stock parts a move's behavior starts from. */
function stockPartsFor(move: MoveData, bodies: Bodies): StockParts {
	const standIn = bodies(stockOverrideRef(move.Name));
	if (standIn !== undefined) return (standIn as () => StockParts)();
	if (move.Stock === null) return {};
	const parts = stockMove(move.Stock.Type, move.Stock.Config);
	if (parts === undefined) throw gameSpecError('move ' + JSON.stringify(move.Name) + ': stock move ' + move.Stock.Type + ' is not ported (engine gap)');
	return parts;
}

/** What the data says about a move, whatever its behavior. */
function configuredMove(move: MoveData, moves: Readonly<Record<string, MoveData>>, enums: Readonly<Record<string, EnumDef>>, resolve: Resolve): Omit<MoveSpec, 'apply'> {
	if (BASES.indexOf(move.Base) < 0) throw gameSpecError('move ' + JSON.stringify(move.Name) + ': unknown base ' + JSON.stringify(move.Base));
	const spec: Omit<MoveSpec, 'apply'> = {
		name: move.Name,
		helpText: move.HelpText,
		fixUp: move.IsFixUp,
		base: move.Base,
		fields: fieldsFromData(move),
		allowMultipleInProgression: move.AllowMultipleInProgression,
		planEnabled: move.Plan.Declared,
	};
	if (move.Plan.Declared) spec.plan = move.Plan.Specs;
	if (move.LegalPhases !== null && move.LegalPhases.length > 0) spec.legalPhases = move.LegalPhases;
	if (move.Progression !== null) spec.progression = progressionFromData(move.Progression, moves, resolve);
	if (move.NameSanitization !== undefined) spec.nameSanitization = move.NameSanitization;
	if (move.HiddenAnimationKey !== undefined) spec.hiddenAnimationKey = move.HiddenAnimationKey;
	if (move.Choices !== undefined && move.Choices !== null) spec.choices = choiceFromData(move.Choices, move, enums);
	if (move.SeatsPlayer === true) spec.seatsPlayer = true;
	if (move.GatheringStart === true) spec.gatheringStart = true;
	if (move.EndsPhaseProgression === true) spec.endsPhaseProgression = true;
	return spec;
}

function moveFromData(move: MoveData, moves: Readonly<Record<string, MoveData>>, enums: Readonly<Record<string, EnumDef>>, bodies: Bodies, resolve: Resolve): MoveSpec {
	const configured = configuredMove(move, moves, enums, resolve);
	const where = 'move ' + JSON.stringify(move.Name);
	let spec: MoveSpec;
	const override = bodies(moveOverrideRef(move.Name));
	if (override !== undefined) {
		const hand = (override as () => MoveSpec)();
		spec = { ...configured, apply: hand.apply };
		if (hand.legal) spec.legal = hand.legal;
		if (hand.defaults) spec.defaults = hand.defaults;
		if (hand.legalCustom) spec.legalCustom = hand.legalCustom;
		if (hand.contributes) spec.contributes = hand.contributes;
		if (hand.sourceProperty !== undefined) spec.sourceProperty = hand.sourceProperty;
		if (hand.destinationProperty !== undefined) spec.destinationProperty = hand.destinationProperty;
	} else {
		const parts = stockPartsFor(move, bodies);
		const stockSpec: Partial<MoveSpec> = parts.spec ?? {};
		const stock: MoveStockParts = {};
		// Only an explicit part is the stock Apply. A base type (Default,
		// CurrentPlayer, FixUp, ...) has none in Go, and the registry drops
		// it; its spec may still carry a placeholder, which must not stand in
		// for the game's missing Apply.
		const stockApply = parts.apply;
		const stockLegal = parts.legal ?? stockSpec.legal;
		const stockDefaults = parts.defaults ?? stockSpec.defaults;
		if (stockApply) stock.apply = stockApply;
		if (stockLegal) stock.legal = stockLegal;
		if (stockDefaults) stock.defaults = stockDefaults;

		let apply: MoveSpec['apply'];
		if (move.Bodies.Apply !== undefined) {
			const body = resolve<(state: RulesState, m: MoveInstance) => unknown>(move.Bodies.Apply, 'apply');
			apply = (state, m) => throwIfError(body(state, m));
		} else if (stockApply) {
			apply = stockApply;
		} else {
			throw gameSpecError(where + ' has no Apply: neither a body nor a stock move supplies one');
		}
		spec = { ...configured, apply, stock };
		if (stockSpec.sourceProperty !== undefined) spec.sourceProperty = stockSpec.sourceProperty;
		if (stockSpec.destinationProperty !== undefined) spec.destinationProperty = stockSpec.destinationProperty;
		if (stockSpec.contributes) spec.contributes = stockSpec.contributes;

		if (move.Bodies.Legal !== undefined) {
			const body = resolve<(state: RulesState, m: MoveInstance, proposer: number, base: () => LegalResult) => LegalResult>(move.Bodies.Legal, 'legal');
			spec.legal = (state, m, proposer, base) => body(state, m, proposer, () => (stockLegal ? stockLegal(state, m, proposer, base) : base()));
		} else if (stockLegal) {
			spec.legal = stockLegal;
		}
		if (move.Bodies.Defaults !== undefined) {
			const body = resolve<(state: RulesState, m: MoveInstance) => void>(move.Bodies.Defaults, 'defaults');
			spec.defaults = (state, m) => {
				if (stockDefaults) stockDefaults(state, m);
				body(state, m);
			};
		} else if (stockDefaults) {
			spec.defaults = stockDefaults;
		}
		if (move.Bodies.LegalCustom !== undefined) {
			spec.legalCustom = resolve<(state: RulesState, m: MoveInstance, proposer: number) => LegalResult | LegalVerdict>(move.Bodies.LegalCustom, 'legalCustom');
		}
	}
	if (move.Plan.HasCustom !== (spec.legalCustom !== undefined)) {
		throw gameSpecError(where + ': Plan.HasCustom is ' + move.Plan.HasCustom + ' but the move ' + (spec.legalCustom ? 'has' : 'has no') + ' LegalCustom');
	}
	if (spec.legalCustom !== undefined && !move.Plan.Declared) throw gameSpecError(where + ': a LegalCustom check needs a declared plan');
	return spec;
}

function computedFromData(item: ComputedData, resolve: Resolve): ComputedSpec {
	if (COMPUTED_TYPES.indexOf(item.Type) < 0) throw gameSpecError('computed property ' + item.Name + ': unknown type ' + JSON.stringify(item.Type));
	const type = item.Type as ComputedType;
	if (item.Scope === 'global') return { name: item.Name, scope: 'global', type, compute: resolve<(state: RulesState) => unknown>(item.Compute, 'computed') };
	if (item.Scope === 'player') {
		return { name: item.Name, scope: 'player', type, compute: resolve<(state: RulesState, player: number) => unknown>(item.Compute, 'computed') };
	}
	throw gameSpecError('computed property ' + item.Name + ': unknown scope ' + JSON.stringify(item.Scope));
}

function decksFromData(def: GameDefinition): Record<string, DeckSpec> {
	const decks: Record<string, DeckSpec> = {};
	for (const name of Object.keys(def.Decks)) {
		const deck = def.Decks[name];
		const components: (Readonly<Record<string, unknown>> | null)[] = [];
		for (let i = 0; i < deck.Size; i++) {
			const values = deck.Components ? deck.Components[i] : null;
			components.push(values === undefined || values === null ? null : (values as Readonly<Record<string, unknown>>));
		}
		const spec: DeckSpec = { components, generic: deck.GenericValues ?? null };
		if (deck.ComponentProps !== undefined) {
			const props: Record<string, { Type: (typeof deck.ComponentProps)[string]['Type']; Enum?: string }> = {};
			for (const prop of Object.keys(deck.ComponentProps)) {
				const item = deck.ComponentProps[prop];
				props[prop] = item.Enum !== undefined && item.Enum !== '' ? { Type: item.Type, Enum: item.Enum } : { Type: item.Type };
			}
			spec.props = props;
		}
		decks[name] = spec;
	}
	return decks;
}

/**
 * Builds the GameSpec a hand port would write from a game's exported
 * definition and its functions. Every FnRef the data names is resolved
 * through `bodies` (a missing one is a GameSpecError naming it); a move's
 * stock ancestor through `stockMove`, unless a hook answers for it. See
 * "Body calling convention" and "Hooks beside FnRefs" at the top of this
 * file. compileGame then derives everything else, as for a hand port.
 */
export function gameFromData(data: GameData, bodies: Bodies): GameSpec {
	if (data.SchemaVersion !== 1) throw gameSpecError('gameFromData: unsupported SchemaVersion ' + String(data.SchemaVersion));
	if (data.Predicates.length > 0) {
		throw gameSpecError('gameFromData: custom legal predicate constructors are not supported (' + data.Predicates.map(p => p.Construct).join(', ') + ')');
	}
	if (data.PhaseEnum !== '' && data.PhaseEnum !== 'phase' && data.PhaseEnum !== 'Phase') {
		throw gameSpecError('gameFromData: the phase enum must be named "phase", not ' + JSON.stringify(data.PhaseEnum));
	}
	if (data.GroupEnum !== '' && data.GroupEnum !== 'group') {
		throw gameSpecError('gameFromData: the group enum must be named "group", not ' + JSON.stringify(data.GroupEnum));
	}
	const resolve = newResolver(bodies);
	const def = data.Definition;
	const enums: Record<string, EnumDef> = { ...data.ComponentEnums, ...def.Enums };

	const movesByName: Record<string, MoveData> = {};
	for (const move of data.Moves) movesByName[move.Name] = move;

	// ----- sub-states and traits
	const gameTraits: NonNullable<GameSpec['game']['traits']> = {};
	const gameTraitData = data.Traits.Game;
	if (gameTraitData.DrawDiscardPair) gameTraits.drawDiscardPair = { draw: gameTraitData.DrawDiscardPair.Draw, discard: gameTraitData.DrawDiscardPair.Discard };
	if (gameTraitData.PlayerOrder !== undefined) gameTraits.playerOrder = resolve<(state: RulesState) => readonly number[] | null>(gameTraitData.PlayerOrder, 'playerOrder');

	const playerTraits: NonNullable<GameSpec['player']['traits']> = {};
	const playerTraitData = data.Traits.Player;
	if (playerTraitData.Seat) playerTraits.seat = true;
	if (playerTraitData.Inactive) playerTraits.inactive = true;
	if (playerTraitData.Elimination) playerTraits.elimination = true;
	if (playerTraitData.Role) playerTraits.role = true;
	if (playerTraitData.GameScore !== undefined) {
		playerTraits.gameScore = playerTraitData.GameScore === 'score' ? scoreGameScore : resolve<(state: RulesState, player: number) => number>(playerTraitData.GameScore, 'gameScore');
	}
	if (playerTraitData.Turn) {
		const turn = playerTraitData.Turn;
		const turnDone = resolve<(state: RulesState, player: number) => string | null>(turn.TurnDone, 'turnDone');
		const start = resolve<(state: RulesState, player: number) => unknown>(turn.ResetForTurnStart, 'resetForTurnStart');
		const end = resolve<(state: RulesState, player: number) => unknown>(turn.ResetForTurnEnd, 'resetForTurnEnd');
		playerTraits.turn = {
			turnDone: (state, player) => turnDone(state, player),
			resetForTurnStart: (state, player) => throwIfError(start(state, player)),
			resetForTurnEnd: (state, player) => throwIfError(end(state, player)),
		};
	}

	const game: SubStateSpec & { traits?: NonNullable<GameSpec['game']['traits']> } = { props: def.Game.Props, initialValues: [data.Initial.Game], traits: gameTraits };
	const player: SubStateSpec & { traits?: NonNullable<GameSpec['player']['traits']> } = { props: def.Player.Props, initialValues: data.Initial.Players, traits: playerTraits };
	const dynamic: Record<string, SubStateSpec> = {};
	for (const deck of Object.keys(def.DynamicComponentValues)) {
		dynamic[deck] = { props: def.DynamicComponentValues[deck].Props, initialValues: data.Initial.Dynamic[deck] ?? [] };
	}
	for (const deck of Object.keys(data.Initial.Dynamic)) {
		if (dynamic[deck] === undefined) throw gameSpecError('gameFromData: Initial.Dynamic names deck ' + deck + ', which has no dynamic values');
	}

	// ----- set-up and the end
	const setUp: GameSpec['setUp'] = {
		distribute: (() => {
			const body = resolve<(state: RulesState, deck: string, index: number) => unknown>(data.SetUp.Distribute, 'distribute');
			return (state: RulesState, deck: string, index: number): Stack => toStack(state, body(state, deck, index), data.SetUp.Distribute);
		})(),
	};
	if (data.SetUp.Begin !== undefined) {
		const body = resolve<(state: RulesState, variant: Readonly<Record<string, string>>) => unknown>(data.SetUp.Begin, 'setUpBegin');
		setUp.begin = (state, variant) => throwIfError(body(state, variant));
	}
	if (data.SetUp.Finish !== undefined) {
		const body = resolve<(state: RulesState) => unknown>(data.SetUp.Finish, 'setUpFinish');
		setUp.finish = state => throwIfError(body(state));
	}

	const end: NonNullable<GameSpec['end']> = {};
	if (data.End.ConditionMet !== undefined) end.conditionMet = resolve<(state: RulesState) => boolean>(data.End.ConditionMet, 'conditionMet');
	if (data.End.PlayerScore !== undefined) end.playerScore = resolve<(state: RulesState, player: number) => number>(data.End.PlayerScore, 'playerScore');
	if (data.End.LowScoreWins) end.lowScoreWins = true;
	if (data.End.Check !== undefined) {
		const body = resolve<(state: RulesState) => unknown>(data.End.Check, 'checkGameFinished');
		const ref = data.End.Check;
		end.check = state => {
			const result = body(state);
			if (Array.isArray(result)) return { finished: result[0] === true, winners: Array.isArray(result[1]) ? (result[1] as number[]) : [] };
			if (result !== null && typeof result === 'object') {
				const item = result as { finished?: unknown; winners?: unknown };
				return { finished: item.finished === true, winners: Array.isArray(item.winners) ? (item.winners as number[]) : [] };
			}
			throw new Error(ref + ' did not return [finished, winners]');
		};
	}

	const spec: GameSpec = {
		name: data.Name,
		displayName: data.DisplayName,
		minNumPlayers: data.MinNumPlayers,
		maxNumPlayers: data.MaxNumPlayers,
		defaultNumPlayers: data.DefaultNumPlayers,
		decks: decksFromData(def),
		enums: def.Enums,
		componentEnums: data.ComponentEnums,
		constants: data.Constants,
		game,
		player,
		dynamicComponentValues: dynamic,
		computed: data.Computed.map(item => computedFromData(item, resolve)),
		legalTemplates: data.LegalTemplates,
		moves: data.Moves.map(move => moveFromData(move, movesByName, enums, bodies, resolve)),
		setUp,
	};
	const variants = variantsFromData(data.Variants);
	if (variants !== undefined) spec.variants = variants;
	if (end.conditionMet || end.playerScore || end.check || end.lowScoreWins) spec.end = end;
	if (data.CurrentPlayerIndex !== undefined) spec.currentPlayerIndex = resolve<(state: RulesState) => number>(data.CurrentPlayerIndex, 'currentPlayerIndex');
	return spec;
}
