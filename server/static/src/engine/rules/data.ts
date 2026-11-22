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

import type { EnumDef, GameDefinition } from '../definition.ts';
import type { LegalSpec } from '../legal/types.ts';
import type { CounterData } from './progression.ts';
import type { GameSpec } from './types.ts';

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
 * reference it does not know.
 */
export type Bodies = (ref: string) => ((...args: never[]) => unknown) | undefined;

/**
 * Builds the GameSpec a hand port would write from a game's exported
 * definition and its translated functions.
 *
 * Not implemented: workstream 6 owns it.
 */
export function gameFromData(data: GameData, bodies: Bodies): GameSpec {
	void data;
	void bodies;
	throw new Error('gameFromData: not implemented');
}
