/**
 * Type definitions for API responses from the boardgame server.
 * These match the JSON structures returned by the Go backend.
 */

import type { CompanionInfo, GameChest, MoveForm, PlayerInfo, ProjectedMoveChoicesWire } from '@boardgame/elements/types/wire.js';
import type { GameFromServer } from '@boardgame/elements/types/game-state.js';

// The shapes a renderer and its elements take are wire types (./wire), shared
// with hosts that never touch this HTTP API. They are re-exported here so
// existing host imports keep working.
export type {
  ClientMove,
  EnumDefinition,
  JsonValue,
  MoveForm,
  MoveFormField,
  PreconditionEntry,
  PreconditionMessage,
  ProjectedMoveChoiceCandidateWire,
  ProjectedMoveChoiceSetWire,
  ProjectedMoveChoiceSource,
  ProjectedMoveChoicesWire,
} from '@boardgame/elements/types/wire.js';

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
