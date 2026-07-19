/**
 * Type definitions for API responses from the boardgame server.
 * These match the JSON structures returned by the Go backend.
 */

import type { CompanionInfo, GameChest, PlayerInfo, ExpandedGameState } from './store';
import type { GameFromServer } from './game-state';

/**
 * Base response structure from the server.
 * All API responses include Status, and optionally Error/FriendlyError.
 */
export interface BaseApiResponse {
  Status: 'Success' | 'Failure';
  Error?: string;
  FriendlyError?: string;
}

/**
 * Response from /api/game/{name}/{id}/info endpoint.
 * Includes complete game information and the current state bundle.
 */
export interface GameInfoResponse extends BaseApiResponse {
  /** Game chest with deck configurations and components */
  Chest: GameChest;
  /** Information about all players */
  Players: PlayerInfo[];
  /** Whether the game has empty player slots */
  HasEmptySlots: boolean;
  /** Whether the game is open to new players */
  GameOpen: boolean;
  /** Whether the game is publicly visible */
  GameVisible: boolean;
  /** Whether the current user owns the game */
  IsOwner: boolean;
  /** Current game state */
  Game: GameFromServer;
  /** Available move forms for current state */
  Forms: MoveForm[] | null;
  /** Which player index is viewing */
  ViewingAsPlayer: number;
  /** Version of the state being returned (may differ from Game.Version) */
  StateVersion: number;
  /**
   * Stamps the shape of the declarative-legality wire format (design spec
   * §6). A client with an older bundled catalog should treat unknown
   * predicate names as evaluable:false and defer to the server's own
   * verdicts rather than mis-evaluating them itself.
   */
  LegalCatalogVersion: number;
  /** Canonical server move-input contract; generated clients must match it. */
  MoveInputSchemaFingerprint: string;
  /** Actor-scoped, finite candidate legality for this exact state. */
  ProjectedMoveChoices?: ProjectedMoveChoicesWire;
  /** Companion-mode metadata, or null for an ordinary solo game. */
  CompanionInfo: CompanionInfo | null;
}

/**
 * Response from /api/game/{name}/{id}/version/{version} endpoint.
 * Returns state bundles for animation playback.
 */
export interface GameVersionResponse extends BaseApiResponse {
  /** Array of state bundles to animate through */
  Bundles: ServerStateBundle[];
  Error?: string;
}

/**
 * A state bundle containing game state, forms, and move information.
 * Used for animating state transitions.
 */
export interface ServerStateBundle {
  /** Game state snapshot */
  Game: GameFromServer;
  /** Available move forms for this state */
  Forms: MoveForm[] | null;
  /** Which player is viewing */
  ViewingAsPlayer: number;
  /** The move that led to this state (null for initial state) */
  Move: unknown;
  /** Actor-scoped, finite candidate legality for this exact state. */
  ProjectedMoveChoices?: ProjectedMoveChoicesWire;
}

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
 * Response from /api/game/{name}/{id}/move endpoint.
 * Result of submitting a move.
 */
export interface MoveResponse extends BaseApiResponse {
  // Server returns Status, Error, FriendlyError
  // No additional fields on success
}

/**
 * Response from /api/game/{name}/{id}/configure endpoint.
 * Result of configuring game properties.
 */
export interface ConfigureResponse extends BaseApiResponse {
  // Server returns Status, Error, FriendlyError
  // No additional fields on success
}

/**
 * Response from /api/game/{name}/{id}/join endpoint.
 * Result of joining a game.
 */
export interface JoinResponse extends BaseApiResponse {
  // Server returns Status, Error, FriendlyError
  // No additional fields on success
}
