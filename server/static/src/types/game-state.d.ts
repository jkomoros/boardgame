/**
 * Type definitions for game state structures.
 * These types represent the core game state data structures used throughout the application.
 */

import type { MoveForm } from './api';

/**
 * Raw game state from server (unexpanded).
 * This is stored in Redux and expanded on-read by selectors.
 * Component indices are expanded to full component objects by the selector.
 */
export interface RawGameState {
  /** State version number */
  Version: number;
  /** Global game state */
  Game: RawPlayerState;
  /** Per-player states */
  Players: RawPlayerState[];
  /** Computed values (optional, may include computed player states) */
  Computed?: {
    Players?: any[];
  };
  /** Component values indexed by deck name and index */
  Components?: Record<string, Record<number, any>>;
}

/**
 * Raw player state (before expansion).
 * Properties can contain stacks (with Deck/Indexes) or timers (with IsTimer).
 * The exact properties depend on the game type.
 */
export interface RawPlayerState {
  [key: string]: any;
}

/**
 * Timer metadata for expansion.
 * Stored in Redux state.game.timerInfos and used by selectors to expand timer values.
 */
export interface TimerInfo {
  /** Current time left in milliseconds */
  TimeLeft: number;
  /** Original time left when timer was started (preserved for reference) */
  originalTimeLeft: number;
  /** Timer ID (optional, used for tracking) */
  ID?: string;
}

/**
 * Game object from server containing state and timer info.
 * This is the structure that comes from the server API.
 */
export interface GameFromServer {
  /** Current raw game state */
  CurrentState: RawGameState;
  /** Active timer information */
  ActiveTimers: Record<string, TimerInfo>;
  /** Game version */
  Version: number;
  /** Current player index */
  CurrentPlayerIndex: number;
  /** Whether game is finished */
  Finished: boolean;
  /** Winner indices if game is finished */
  Winners: number[];
  /** Diagram for rendering (optional) */
  Diagram?: string;
  /** Other game-specific properties */
  [key: string]: any;
}

/**
 * State bundle for animation playback.
 * Bundles queue in Redux state.game.animation.pendingBundles and are fired sequentially.
 */
export interface StateBundle {
  /** Wall clock time when bundle was created */
  originalWallClockStartTime: number;
  /** Game object from server (contains CurrentState and ActiveTimers) */
  game: GameFromServer;
  /** Move that triggered this state (null for initial state) */
  move: any | null;
  /** Expanded move forms for this state */
  moveForms: MoveForm[] | null;
  /** Player index viewing this state */
  viewingAsPlayer: number;
}
