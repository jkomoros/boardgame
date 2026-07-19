/**
 * Type definitions for game state structures.
 * These types represent the core game state data structures used throughout the application.
 */

import type { ClientMove, JsonValue, MoveForm, ProjectedMoveChoicesWire } from './api';
import type { VersionAnimationContext } from '../components/companion-sync';

/**
 * Raw game state from server (unexpanded).
 * This is stored in Redux and expanded on-read by selectors.
 * Component indices are expanded to full component objects by the selector.
 */
export interface RawGameState {
  /** Optional legacy state version; authoritative version lives on GameFromServer. */
  Version?: number;
  /** Global game state */
  Game: RawPlayerState;
  /** Per-player states */
  Players: RawPlayerState[];
  /** Computed values (optional, may include computed global and player states) */
  Computed?: {
    Global?: RawPlayerState;
    Players?: RawPlayerState[];
  };
  /** Component values indexed by deck name and index */
  Components?: Record<string, JsonValue[]>;
}

/**
 * Raw player state (before expansion).
 * Properties can contain stacks (with Deck/Indexes) or timers (with IsTimer).
 * The exact properties depend on the game type.
 */
export type RawPlayerState = object;

/**
 * Timer metadata for expansion.
 * Stored in Redux state.game.timerInfos and used by selectors to expand timer values.
 */
export interface TimerInfo {
  /** Current time left in milliseconds */
  TimeLeft: number;
  /** Original time left when timer was started (preserved for reference) */
  originalTimeLeft?: number;
  /** Timer ID (optional, used for tracking) */
  ID?: string;
}

/**
 * Game object from server containing state and timer info.
 * This is the structure that comes from the server API.
 */
export interface GameFromServer {
  /** Registered game package name. */
  Name: string;
  /** Stable game instance identifier. */
  ID: string;
  /** Number of player slots configured for this game. */
  NumPlayers: number;
  /** Agent name for each configured automated seat. */
  Agents: string[];
  /** Selected variant value keyed by variant property name. */
  Variant: Record<string, string>;
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
  move: ClientMove | null;
  /** Expanded move forms for this state */
  moveForms: MoveForm[] | null;
  /** Player index viewing this state */
  viewingAsPlayer: number;
  /** Candidate legality projected for this bundle's exact state/viewer. */
  projectedMoveChoices: ProjectedMoveChoicesWire | null;
  /** Scoped animation policy reserved for this exact companion version. */
  animationContext?: VersionAnimationContext | null;
  /** Opaque local identity for exactly one installed animation cycle. */
  motionCycleId?: number;
}
