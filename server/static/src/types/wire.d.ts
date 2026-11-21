/**
 * Wire types: the JSON shapes a renderer host hands to a renderer and its
 * elements, whatever produced them. They mirror what the Go server emits, and
 * a host that is not the webapp (a Fabric guest, a fixture) supplies the same
 * shapes.
 *
 * This file is layer 1 (see scripts/layers.json). It must not import the Redux
 * store's types (./store) or the HTTP response envelopes (./api); those import
 * from here instead.
 */
import type {
  CatalogComponent,
  FullGameState,
  GameChest as RendererGameChest,
} from './boardgame-types';

// ---- Game chest and expanded state ----

/**
 * Game chest containing component configurations and deck defaults.
 */
export type GameChest = RendererGameChest<
  Readonly<Record<string, readonly CatalogComponent[]>>
>;

/**
 * Expanded game state with all component references resolved.
 * State from server has component indices that get expanded to full objects.
 *
 * For type-safe access in game renderers, use the generated _types.ts
 * in each game's client/ directory (via boardgame-util emit-types).
 * See also FullGameState in types/boardgame-types.ts.
 */
export type ExpandedGameState = FullGameState<
  object,
  object,
  object,
  object,
  Readonly<Record<string, readonly unknown[]>>
>;

/**
 * Information about a player in the game.
 */
export interface PlayerInfo {
  /** Whether this player slot is currently unoccupied. */
  IsEmpty: boolean;
  /** Whether the occupied slot is controlled by an agent. */
  IsAgent: boolean;
  /** Player display name supplied by the server. */
  DisplayName: string;
  /** Optional player photo URL supplied by the server. */
  PhotoURL?: string;
}

// ---- Companion mode ----

/**
 * Companion-mode bundle from the doGameInfo response. Matches the JSON
 * shape emitted by server/api/main.go doGameInfo(). Keys are PascalCase
 * because Go's JSON marshaling uses struct field names.
 */
export interface CompanionSeatPresentation {
  playerIndex: number;
  displayName: string;
  avatarSlug: string;
}

export interface CompanionTableSession {
  Status: 'active' | 'available';
  IsThisTable: boolean;
  CanTakeOver: boolean;
  /** Server-authoritative delay before checking whether an active Table died. */
  RetryAfterMs: number;
  /** This exact former Table was intentionally moved to another screen. */
  DisplacedByTransfer: boolean;
}

export interface CompanionInfo {
  CompanionMode: boolean;
  RoomCode: string;
  RoomLocked: boolean;
  SeatPresentations: CompanionSeatPresentation[];
  Absent: number[];
  /** Ownership and recovery state for the single shared Table display. */
  TableSession: CompanionTableSession;
  /** Server-computed authority verdict for host actions on this request. */
  IsHost: boolean;
  /** This request may create the finished game's one durable successor. */
  CanRematch: boolean;
  /** Published only after the successor is fully configured and safe to enter. */
  RematchGameID: string;
}

// ---- Moves, forms and legality ----

export type ProjectedMoveChoiceSource = 'players' | 'enum-values' | 'stack-slots';

export interface ProjectedMoveChoiceCandidateWire {
  readonly Value: string | number;
  readonly Available: boolean;
}

export interface ProjectedMoveChoiceSetWire {
  readonly MoveName: string;
  readonly FieldName: string;
  readonly Source: ProjectedMoveChoiceSource;
  readonly Candidates: readonly ProjectedMoveChoiceCandidateWire[];
}

/** Decoded but still untrusted until checked against the generated game schema. */
export type ProjectedMoveChoicesWire =
  | {
    readonly StateVersion: number;
    readonly MoveChoiceProjectionSchemaFingerprint: string;
    readonly ProjectionSchemaVersion: number;
    readonly Status: 'ready';
    readonly Sets: readonly ProjectedMoveChoiceSetWire[];
  }
  | {
    readonly StateVersion: number;
    readonly MoveChoiceProjectionSchemaFingerprint: string;
    readonly ProjectionSchemaVersion: number;
    readonly Status: 'failed';
    readonly Sets: readonly [];
  };

/**
 * A move form describing an available move and its parameters.
 */
export interface MoveForm {
  /** Move type name */
  Name: string;
  /** Help text explaining the move */
  HelpText: string;
  /** Form fields for move parameters */
  Fields?: MoveFormField[];
  /** Whether this move is legal for the viewing player right now */
  LegalForPlayer?: boolean;
  /** Error message from Legal() if the move is not legal for this player */
  LegalForPlayerError?: string;
  /** Whether this move is structurally legal (legal for any player / admin) */
  LegalForAnyone?: boolean;
  /** Whether this move is a gathering "start game" move (e.g., CloseAllSeats) */
  IsGatheringStart?: boolean;
  /**
   * Per-predicate declarative-legality ledger (design spec §6). Present
   * only for a move type that opted in to declarative legality
   * (WithLegalPreconditions); absent for an opaque (un-migrated) move, whose
   * legality is fully described by LegalForPlayer/LegalForPlayerError/
   * LegalForAnyone alone. No UI reads this yet (Task 10 ships the wire
   * format only) -- it is plumbed through selectMoveLegality
   * (selectors.ts) as MoveLegalityInfo.preconditions for future use.
   */
  Preconditions?: PreconditionEntry[];
}

/**
 * One predicate's line in a MoveForm's Preconditions ledger (design spec
 * §6). Mirrors server/api/main.go's preconditionEntry.
 */
export interface PreconditionEntry {
  /** Predicate's registry name ("any" for a compositor, "custom" for the CustomLegaler escape hatch). */
  name: string;
  /** Predicate's string args; absent for compositors/custom. */
  args?: string[];
  /** Three-valued outcome. */
  verdict: 'pass' | 'fail' | 'unknown';
  /** Present only for a non-Pass verdict. */
  message?: PreconditionMessage;
  /**
   * Whether a client could reproduce this entry's verdict itself:
   * has-a-wire-form AND every declared Read survives the viewer's
   * sanitization. When false, message.bindings (if any) has been stripped
   * server-side (the #693 guard) -- only the template key ships.
   */
  evaluable: boolean;
  /**
   * True for a field-dependent verdict: computed against a server-chosen
   * (DefaultsForState-bound) move, so a different choice of move field
   * values could evaluate differently.
   */
  provisional?: boolean;
}

/** The Message half of a PreconditionEntry -- template key plus (subject to the #693 guard) bindings. */
export interface PreconditionMessage {
  /** Template key to look up in the chest's LegalTemplates table. */
  template: string;
  /**
   * Named values substituted into the template body. Absent whenever the
   * owning entry's evaluable is false (design spec §6, #693 guard):
   * bindings are derived from state a less-privileged viewer may not be
   * allowed to see.
   */
  bindings?: Record<string, string | number | boolean>;
}

/**
 * A field in a move form.
 */
export interface MoveFormField {
  /** Field name */
  Name: string;
  /** Numeric boardgame.PropertyType value from the Go API. */
  Type: number;
  /** Current/default wire value for the move property. */
  DefaultValue: JsonValue;
  /** Name of the enum (used for expansion) */
  EnumName?: string;
  /** Expanded enum values (populated during expansion) */
  Enum?: EnumDefinition;
}

/** A value that can cross the server's JSON transport boundary. */
export type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };

/** Serialized enum metadata supplied in the game chest. */
export interface EnumDefinition {
  Values?: Record<string, string>;
  [key: string]: unknown;
}

/**
 * A move that was made in the game.
 */
export type ClientMove = Readonly<{
  /** Viewer-specific public key used only to select animation policy. */
  AnimationKey: string;
  /** State version produced by the move. */
  Version: number;
  /** Explicitly disclosed, viewer-sanitized move properties. */
  Properties?: Readonly<Record<string, JsonValue>>;
}>;

/**
 * One move's legality for the viewer, derived from its MoveForm. Hosts build a
 * record of these keyed by move name and hand it to the renderer.
 */
export interface MoveLegalityInfo {
    legalForPlayer: boolean;
    legalForAnyone: boolean;
    error?: string;
    /**
     * Per-predicate declarative-legality ledger (design spec §6), passed
     * through verbatim from MoveForm.Preconditions. Present only for a
     * move type that opted in to declarative legality; undefined for an
     * opaque move, where legalForPlayer/legalForAnyone/error alone
     * describe legality. No UI reads this yet -- Task 10 ships the plumbing
     * only, ahead of a future client evaluator/explainer feature.
     */
    preconditions?: PreconditionEntry[];
}
