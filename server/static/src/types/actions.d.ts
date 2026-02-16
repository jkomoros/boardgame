/**
 * Type definitions for Redux actions.
 * All actions use discriminated unions for type safety in the reducer.
 */

import type { GameChest, PlayerInfo } from './store';
import type { MoveForm } from './api';
import type { RawGameState, TimerInfo, StateBundle } from './game-state';

// Import action type constants for typeof usage
import {
  UPDATE_GAME_ROUTE,
  UPDATE_GAME_STATIC_INFO,
  UPDATE_GAME_CURRENT_STATE,
  CONFIGURE_GAME_REQUEST,
  CONFIGURE_GAME_SUCCESS,
  CONFIGURE_GAME_FAILURE,
  JOIN_GAME_REQUEST,
  JOIN_GAME_SUCCESS,
  JOIN_GAME_FAILURE,
  SUBMIT_MOVE_REQUEST,
  SUBMIT_MOVE_SUCCESS,
  SUBMIT_MOVE_FAILURE,
  FETCH_GAME_INFO_REQUEST,
  FETCH_GAME_INFO_SUCCESS,
  FETCH_GAME_INFO_FAILURE,
  FETCH_GAME_VERSION_REQUEST,
  FETCH_GAME_VERSION_SUCCESS,
  FETCH_GAME_VERSION_FAILURE,
  ENQUEUE_STATE_BUNDLE,
  DEQUEUE_STATE_BUNDLE,
  CLEAR_STATE_BUNDLES,
  MARK_ANIMATION_STARTED,
  MARK_ANIMATION_COMPLETED,
  SET_CURRENT_VERSION,
  SET_TARGET_VERSION,
  SET_LAST_FETCHED_VERSION,
  SOCKET_CONNECTED,
  SOCKET_DISCONNECTED,
  SOCKET_ERROR,
  UPDATE_VIEW_STATE,
  SET_VIEWING_AS_PLAYER,
  SET_REQUESTED_PLAYER,
  SET_AUTO_CURRENT_PLAYER,
  UPDATE_MOVE_FORMS,
  CLEAR_FETCHED_INFO,
  CLEAR_FETCHED_VERSION
} from '../actions/game';

// ============================================================================
// Game Route Actions
// ============================================================================

export interface UpdateGameRouteAction {
  type: typeof UPDATE_GAME_ROUTE;
  name: string;
  id: string;
}

// ============================================================================
// Game State Actions
// ============================================================================

export interface UpdateGameStaticInfoAction {
  type: typeof UPDATE_GAME_STATIC_INFO;
  chest: GameChest | null;
  playersInfo: PlayerInfo[];
  hasEmptySlots: boolean;
  open: boolean;
  visible: boolean;
  isOwner: boolean;
}

export interface UpdateGameCurrentStateAction {
  type: typeof UPDATE_GAME_CURRENT_STATE;
  currentState: RawGameState;
  timerInfos: Record<string, TimerInfo> | null;
  pathsToTick: (string | number)[][];
  originalWallClockTime: number;
}

// ============================================================================
// Async Operation Actions (15 total: 5 groups × 3 each)
// ============================================================================

// Configure Game
export interface ConfigureGameRequestAction {
  type: typeof CONFIGURE_GAME_REQUEST;
}

export interface ConfigureGameSuccessAction {
  type: typeof CONFIGURE_GAME_SUCCESS;
}

export interface ConfigureGameFailureAction {
  type: typeof CONFIGURE_GAME_FAILURE;
  error: string;
  friendlyError: string;
}

// Join Game
export interface JoinGameRequestAction {
  type: typeof JOIN_GAME_REQUEST;
}

export interface JoinGameSuccessAction {
  type: typeof JOIN_GAME_SUCCESS;
}

export interface JoinGameFailureAction {
  type: typeof JOIN_GAME_FAILURE;
  error: string;
  friendlyError: string;
}

// Submit Move
export interface SubmitMoveRequestAction {
  type: typeof SUBMIT_MOVE_REQUEST;
}

export interface SubmitMoveSuccessAction {
  type: typeof SUBMIT_MOVE_SUCCESS;
}

export interface SubmitMoveFailureAction {
  type: typeof SUBMIT_MOVE_FAILURE;
  error: string;
  friendlyError: string;
}

// Fetch Game Info
export interface FetchGameInfoRequestAction {
  type: typeof FETCH_GAME_INFO_REQUEST;
}

export interface FetchGameInfoSuccessAction {
  type: typeof FETCH_GAME_INFO_SUCCESS;
  chest: GameChest | null;
  playersInfo: PlayerInfo[];
  hasEmptySlots: boolean;
  open: boolean;
  visible: boolean;
  isOwner: boolean;
  game: RawGameState;
  forms: MoveForm[] | null;
  viewingAsPlayer: number;
  stateVersion: number;
}

export interface FetchGameInfoFailureAction {
  type: typeof FETCH_GAME_INFO_FAILURE;
  error: string;
  friendlyError: string;
}

// Fetch Game Version
export interface FetchGameVersionRequestAction {
  type: typeof FETCH_GAME_VERSION_REQUEST;
}

export interface FetchGameVersionSuccessAction {
  type: typeof FETCH_GAME_VERSION_SUCCESS;
  bundles: StateBundle[];
}

export interface FetchGameVersionFailureAction {
  type: typeof FETCH_GAME_VERSION_FAILURE;
  error: string;
  friendlyError: string;
}

// ============================================================================
// Animation Actions
// ============================================================================

export interface EnqueueStateBundleAction {
  type: typeof ENQUEUE_STATE_BUNDLE;
  bundle: StateBundle;
}

export interface DequeueStateBundleAction {
  type: typeof DEQUEUE_STATE_BUNDLE;
}

export interface ClearStateBundlesAction {
  type: typeof CLEAR_STATE_BUNDLES;
}

export interface MarkAnimationStartedAction {
  type: typeof MARK_ANIMATION_STARTED;
  animationId: string;
}

export interface MarkAnimationCompletedAction {
  type: typeof MARK_ANIMATION_COMPLETED;
  animationId: string;
}

// ============================================================================
// Version Actions
// ============================================================================

export interface SetCurrentVersionAction {
  type: typeof SET_CURRENT_VERSION;
  version: number;
}

export interface SetTargetVersionAction {
  type: typeof SET_TARGET_VERSION;
  version: number;
}

export interface SetLastFetchedVersionAction {
  type: typeof SET_LAST_FETCHED_VERSION;
  version: number;
}

// ============================================================================
// WebSocket Actions
// ============================================================================

export interface SocketConnectedAction {
  type: typeof SOCKET_CONNECTED;
}

export interface SocketDisconnectedAction {
  type: typeof SOCKET_DISCONNECTED;
}

export interface SocketErrorAction {
  type: typeof SOCKET_ERROR;
  error: string;
}

// ============================================================================
// View State Actions
// ============================================================================

export interface UpdateViewStateAction {
  type: typeof UPDATE_VIEW_STATE;
  game: any; // Keep as any - full game object structure varies by game type
  viewingAsPlayer: number;
  moveForms: MoveForm[] | null;
}

export interface SetViewingAsPlayerAction {
  type: typeof SET_VIEWING_AS_PLAYER;
  playerIndex: number;
}

export interface SetRequestedPlayerAction {
  type: typeof SET_REQUESTED_PLAYER;
  playerIndex: number;
}

export interface SetAutoCurrentPlayerAction {
  type: typeof SET_AUTO_CURRENT_PLAYER;
  autoFollow: boolean;
}

export interface UpdateMoveFormsAction {
  type: typeof UPDATE_MOVE_FORMS;
  moveForms: MoveForm[] | null;
}

export interface ClearFetchedInfoAction {
  type: typeof CLEAR_FETCHED_INFO;
}

export interface ClearFetchedVersionAction {
  type: typeof CLEAR_FETCHED_VERSION;
}

// ============================================================================
// Discriminated Union of All Action Types
// ============================================================================

export type GameAction =
  | UpdateGameRouteAction
  | UpdateGameStaticInfoAction
  | UpdateGameCurrentStateAction
  | ConfigureGameRequestAction
  | ConfigureGameSuccessAction
  | ConfigureGameFailureAction
  | JoinGameRequestAction
  | JoinGameSuccessAction
  | JoinGameFailureAction
  | SubmitMoveRequestAction
  | SubmitMoveSuccessAction
  | SubmitMoveFailureAction
  | FetchGameInfoRequestAction
  | FetchGameInfoSuccessAction
  | FetchGameInfoFailureAction
  | FetchGameVersionRequestAction
  | FetchGameVersionSuccessAction
  | FetchGameVersionFailureAction
  | EnqueueStateBundleAction
  | DequeueStateBundleAction
  | ClearStateBundlesAction
  | MarkAnimationStartedAction
  | MarkAnimationCompletedAction
  | SetCurrentVersionAction
  | SetTargetVersionAction
  | SetLastFetchedVersionAction
  | SocketConnectedAction
  | SocketDisconnectedAction
  | SocketErrorAction
  | UpdateViewStateAction
  | SetViewingAsPlayerAction
  | SetRequestedPlayerAction
  | SetAutoCurrentPlayerAction
  | UpdateMoveFormsAction
  | ClearFetchedInfoAction
  | ClearFetchedVersionAction;
