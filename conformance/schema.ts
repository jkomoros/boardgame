// The format of a conformance document: conformance/<game>/<name>.json.gz,
// gunzipped and parsed. Types only; nothing here is imported or executed.
//
// The source of truth is the Go struct set in
// boardgame-util/lib/conformance/document.go. Bump SchemaVersion there and
// update this file together.
//
// A document records one game run on the Go engine. For every state version
// it holds the unsanitized state, the move that produced it, and what every
// viewer is sent. Compare structurally after parsing: object key order and
// whitespace carry no meaning. Everything omitted or rewritten relative to
// the live wire is listed in `normalization`.

/** A seat index, or -1 for the observer, or -2 for the admin. */
export type PlayerIndex = number;

/** Any JSON value. Used where the shape is the game's own. */
export type Json = null | boolean | number | string | Json[] | { [key: string]: Json };

export interface ConformanceDocument {
	schemaVersion: 1;
	name: string;
	source: Source;
	normalization: NormalizationRule[];
	game: GameHeader;
	engine: EngineHeader;
	/** The order of `views` in every version: [-2, -1, 0, 1, ...]. */
	viewers: PlayerIndex[];
	/** The static move table, in the engine's registration order. */
	moves: MoveType[];
	/** boardgame.BuildMoveInputSchema, as the engine encodes it. */
	moveInputSchema: Json;
	/** boardgame.BuildMoveChoiceProjectionSchema; null when the game declares none. */
	moveChoiceProjectionSchema: Json;
	/** GameManager.ComputedPropertyDescriptors. */
	computedProperties: Json;
	/** The game type's variant configuration: every key, its values and default. */
	variantConfig: Json;
	chest: Chest;
	/** One entry per state version; versions[i].version === i. */
	versions: Version[];
}

export interface Source {
	kind: 'scenario' | 'golden';
	/** The golden history file, relative to the repository root. */
	path?: string;
	description?: string;
	/** What the run exercises. */
	covers?: string;
}

export interface NormalizationRule {
	/** A path in this document, or the wire/storage name of an absent field. */
	field: string;
	action: 'omitted' | 'fixed' | 'derived' | 'rewritten' | 'kept';
	reason: string;
}

/** Everything needed to recreate the game before version 0. */
export interface GameHeader {
	name: string;
	displayName: string;
	/** Component ids and all randomness are functions of id, secretSalt and version. */
	id: string;
	secretSalt: string;
	numPlayers: number;
	minNumPlayers: number;
	maxNumPlayers: number;
	defaultNumPlayers: number;
	/** The effective variant, with the configuration's defaults filled in. */
	variant: { [key: string]: string };
	/** One entry per player; always empty strings (agent games are not exported). */
	agents: string[];
	/**
	 * True when the run seats players (it contains a move with trigger
	 * "seat"). The game is told before version 0, and behaves differently:
	 * with seating off, seats are never waited for or inactivated as empty.
	 */
	seating: boolean;
	/** The final values. Per-version values are on each Version. */
	finished: boolean;
	winners: PlayerIndex[];
	finalVersion: number;
}

export interface EngineHeader {
	/**
	 * The fixed timer clock, RFC 3339. It never advances: a timer started at
	 * any version has Deadline = clock + its duration, and fires only when a
	 * move with trigger "timer" says so.
	 */
	clock: string;
	legalCatalogVersion: number;
	moveInputSchemaFingerprint: string;
	moveChoiceProjectionSchemaFingerprint: string;
	moveChoiceProjectionSchemaVersion: number;
	/** Names of the chest enums used for phases and sanitization groups; "" if none. */
	phaseEnum: string;
	groupEnum: string;
}

export interface MoveType {
	name: string;
	helpText: string;
	/** Fix-up moves are proposed by the engine, never offered to players. */
	isFixUp: boolean;
	/** False when some viewer may not see this move's name. */
	namePublic: boolean;
	/** The canonical name, plus the hidden-name alias if one is configured. */
	animationKeys: string[];
	isGatheringStart: boolean;
	/** Sorted by name. */
	fields: MoveField[];
}

export interface MoveField {
	name: string;
	/** A boardgame.PropertyType name, e.g. "TypeInt", "TypePlayerIndex", "TypeEnum". */
	type: string;
	enumName?: string;
}

/** The chest as the server sends it. */
export interface Chest {
	/** Deck name to its components, in index order. */
	Decks: { [deck: string]: ChestComponent[] };
	Enums: Json;
	Constants: { [name: string]: Json };
	/** Legal message template key to template text. Absent with no declarative moves. */
	LegalTemplates?: { [key: string]: string };
}

export interface ChestComponent {
	Index: number;
	Values: Json;
}

export interface Version {
	version: number;
	/**
	 * True when the engine was waiting for outside input at this version:
	 * setup finished, or a proposal and all of its fix-ups applied. The live
	 * server reports legality and projected choices only for these versions.
	 */
	settled: boolean;
	/** The game's values as of this version. */
	finished: boolean;
	winners: PlayerIndex[];
	/** The move that produced this version. Null for version 0. */
	move: MoveRecord | null;
	/** The unsanitized state exactly as storage holds it. */
	state: StoredState;
	/** One per `viewers`, in that order. */
	views: View[];
}

export interface MoveRecord {
	name: string;
	version: number;
	/** The version of the move that started this causal chain. */
	initiator: number;
	/** The phase enum value before the move applied, and its string value ("" with no phase enum). */
	phase: number;
	phaseName: string;
	/** -2 for moves the engine or a timer proposed. */
	proposer: PlayerIndex;
	/** The move's serialized properties. */
	arguments: { [field: string]: Json };
	/** The move type's isFixUp. */
	isFixUp: boolean;
	/**
	 * What made the engine apply the move. Only "propose", "timer" and "seat"
	 * are outside input a replaying engine must be given; "setup" and "fixup"
	 * moves must be produced by the engine itself.
	 *
	 * - "setup": starts the chain the engine runs when the game is created.
	 * - "propose": `proposer` proposed it with `arguments`.
	 * - "timer": the timer whose intended move this is fired.
	 * - "seat": a player was seated at arguments.TargetPlayerIndex and a
	 *   fix-up check was forced.
	 * - "fixup": it followed the previous move in the same chain.
	 */
	trigger: 'setup' | 'propose' | 'timer' | 'seat' | 'fixup';
}

/** The state storage record. Sub-state shapes are the game's own. */
export interface StoredState {
	Game: { [property: string]: Json };
	Players: { [property: string]: Json }[];
	/** Deck name to per-component dynamic values; {} when the game has none. */
	Components: { [deck: string]: Json[] };
	/** Every Timer property in the state, active or not. */
	Timers: StoredTimer[];
	/**
	 * Deck name to a per-component count of secret moves. A component's id is
	 * a hash over the game id, salt, deck, index and this count. Absent until
	 * any count is non-zero.
	 */
	SecretMoveCount?: { [deck: string]: number[] };
}

export interface StoredTimer {
	Ref: { Group: number; PropName: string; PlayerIndex: number; DeckName: string; DeckIndex: number };
	/** "" when not active. Drawn from the state's seeded randomness. */
	ID: string;
	Generation: number;
	/** RFC 3339. Go's zero time until the timer first starts. */
	Deadline: string;
	Status: 'inactive' | 'active' | 'canceled' | 'fired';
	/** The move the timer will propose as admin. Present while active. */
	Move?: { Name: string; Version: number; Initiator: number; Phase: number; Proposer: PlayerIndex; Timestamp: string; Blob: Json };
}

export interface View {
	viewer: PlayerIndex;
	game: ViewerGame;
	/**
	 * The sanitized move record. Null for version 0, and null when this
	 * viewer may see nothing of the move (its name is hidden from them and no
	 * alias is configured); the version still advances.
	 */
	move: ViewerMove | null;
	/**
	 * One entry per move offered to this viewer, in `moves` order. Fix-up
	 * moves are never listed. A move whose name is not public is listed only
	 * for a viewer who may see the name and for whom it is currently legal
	 * (and for the admin).
	 */
	legality: MoveLegality[];
	/**
	 * The actor-only choice snapshot. Present only at a settled version, for
	 * a seated player, in a game that declares choice projections.
	 */
	projectedMoveChoices?: ProjectedMoveChoices;
}

/** Game.JSONForPlayer for one viewer and state, normalized. No Diagram. */
export interface ViewerGame {
	Name: string;
	ID: string;
	NumPlayers: number;
	Agents: string[];
	Variant: { [key: string]: string };
	/** The version being viewed (on the live wire: the head version). */
	Version: number;
	/** As of this version (on the live wire: the game's current values). */
	Finished: boolean;
	Winners: PlayerIndex[];
	/** The delegate's current player for the sanitized state. */
	CurrentPlayerIndex: PlayerIndex;
	/** Timer id to milliseconds left at engine.clock, for each active timer. */
	ActiveTimers: { [timerID: string]: { TimeLeft: number } };
	CurrentState: ViewerState;
}

/** The sanitized state as a client receives it. */
export interface ViewerState {
	Game: { [property: string]: Json };
	Players: { [property: string]: Json }[];
	Components: { [deck: string]: Json[] };
	/** Computed from the SANITIZED state, so values can differ by viewer. */
	Computed: {
		Global: { [name: string]: Json };
		Players: { [name: string]: Json }[];
	};
	/** For each property, the facets this viewer can trust. */
	Visibility: {
		Game: { [property: string]: VisibilityFacet[] };
		Players: { [property: string]: VisibilityFacet[] }[];
	};
	/** Present only for the admin, whose state is not sanitized. */
	SecretMoveCount?: { [deck: string]: number[] };
}

export type VisibilityFacet = 'values' | 'count' | 'occupancy' | 'order' | 'nonempty';

/** Game.MoveJSONForPlayer. */
export interface ViewerMove {
	/** The move's name, or its hidden-name alias for a viewer who may not see the name. */
	AnimationKey: string;
	Version: number;
	/** Only the move properties this viewer may see. Absent when there are none. */
	Properties?: { [field: string]: Json };
}

export interface MoveLegality {
	name: string;
	/** Move.Legal(state, viewer) succeeded for the default-bound move. Always false for the observer. */
	legalForPlayer: boolean;
	/** That call's error text, when it failed. */
	legalForPlayerError?: string;
	/** Move.Legal(state, admin) succeeded: the move is structurally possible. */
	legalForAnyone: boolean;
	/** The default value the engine bound to each move field for this state. */
	defaults: { [field: string]: Json };
	/** The declarative-legality ledger. Absent for moves that have not opted in. */
	preconditions?: Precondition[];
}

export interface Precondition {
	/** The predicate's registry name; "any" for a compositor, "custom" for the escape hatch. */
	name: string;
	args?: string[];
	verdict: 'pass' | 'fail' | 'unknown';
	/** Present on a non-pass verdict. Bindings ship only when `evaluable`. */
	message?: { template: string; bindings?: { [name: string]: Json } };
	/** Whether this viewer's client could reproduce the verdict from what it can see. */
	evaluable: boolean;
	/** True when the verdict depends on move fields, so other inputs could differ. */
	provisional?: boolean;
}

export interface ProjectedMoveChoices {
	StateVersion: number;
	MoveChoiceProjectionSchemaFingerprint: string;
	ProjectionSchemaVersion: number;
	Status: string;
	Sets?: {
		MoveName: string;
		FieldName: string;
		Source: string;
		Candidates: { Value: Json; Available: boolean }[];
	}[];
}

/** conformance/index.json: one row per document. */
export interface CorpusIndexEntry {
	/** Path relative to conformance/. */
	file: string;
	game: string;
	name: string;
	source: string;
	/** Number of state versions, counting version 0. */
	versions: number;
	viewers: number;
	finished: boolean;
	covers: string;
}
