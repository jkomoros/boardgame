/**
 * Type definitions for API responses from the boardgame server.
 * These match the JSON structures returned by the Go backend.
 */

import type { GameChest, PlayerInfo, ExpandedGameState } from './store';

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
  Game: any; // Raw game state from server (not yet expanded)
  /** Available move forms for current state */
  Forms: MoveForm[];
  /** Which player index is viewing */
  ViewingAsPlayer: number;
  /** Version of the state being returned (may differ from Game.Version) */
  StateVersion: number;
}

/**
 * Response from /api/game/{name}/{id}/version/{version} endpoint.
 * Returns state bundles for animation playback.
 */
export interface GameVersionResponse extends BaseApiResponse {
  /** Array of state bundles to animate through */
  Bundles: StateBundle[];
  Error?: string;
}

/**
 * A state bundle containing game state, forms, and move information.
 * Used for animating state transitions.
 */
export interface StateBundle {
  /** Game state snapshot */
  Game: any; // Raw game state from server
  /** Available move forms for this state */
  Forms: MoveForm[];
  /** Which player is viewing */
  ViewingAsPlayer: number;
  /** The move that led to this state (null for initial state) */
  Move: Move | null;
}

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
}

/**
 * A field in a move form.
 */
export interface MoveFormField {
  /** Field name */
  Name: string;
  /** Field type (e.g., 'string', 'int', 'enum') */
  Type: string;
  /** Help text for the field */
  HelpText?: string;
  /** For enum fields, the available options */
  EnumOptions?: any[];
  /** Default value */
  Default?: any;
  /** Name of the enum (used for expansion) */
  EnumName?: string;
  /** Expanded enum values (populated during expansion) */
  Enum?: any;
}

/**
 * A move that was made in the game.
 */
export interface Move {
  /** Move type name */
  Name: string;
  /** Player who made the move */
  Player: number;
  /** Move parameters */
  [key: string]: any;
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
