/**
 * Type definitions for the Redux store structure.
 *
 * The store uses lazy reducer loading - app, error, and user are always loaded,
 * while game and list are loaded on-demand when navigating to those views.
 */

import type { RawGameState, TimerInfo, StateBundle } from './game-state';
import type { MoveForm } from './api';

/**
 * Root Redux state containing all slices.
 * Game and list are optional as they're lazily loaded.
 */
export interface RootState {
  app: AppState;
  error: ErrorState;
  user: UserState;
  game?: GameState;
  list?: ListState;
}

/**
 * App-wide state for routing, offline status, and UI.
 */
export interface AppState {
  /** Current window.location */
  location: string;
  /** Current page route (e.g., 'game', 'list') */
  page: string;
  /** Additional routing info (e.g., game ID) */
  pageExtra: string;
  /** Whether the app is offline */
  offline: boolean;
  /** Whether the snackbar notification is visible */
  snackbarOpened: boolean;
  /** Whether the header side panel is open */
  headerPanelOpen: boolean;
}

/**
 * Error state for displaying error messages to users.
 */
export interface ErrorState {
  message: string;
  friendlyMessage: string;
  title: string;
  showing: boolean;
}

/**
 * User authentication and profile state.
 */
export interface UserState {
  /** Whether user has admin privileges active */
  admin: boolean;
  /** Whether user is allowed to activate admin mode */
  adminAllowed: boolean;
  /** Whether user is logged in */
  loggedIn: boolean;
  /** Whether authentication is being verified */
  verifyingAuth: boolean;
  /** User object from backend server (null if not logged in) */
  user: UserInfo | null;
  /** Sign-in error message */
  errorMessage: string;
  /** Whether sign-in dialog is open */
  dialogOpen: boolean;
  /** Email field in sign-in dialog */
  dialogEmail: string;
  /** Password field in sign-in dialog */
  dialogPassword: string;
  /** Whether dialog is in create account mode (vs sign in) */
  dialogIsCreate: boolean;
  /** Currently selected tab in sign-in dialog */
  dialogSelectedPage: number;
}

/**
 * User information returned from backend.
 */
export interface UserInfo {
  id: string;
  displayName?: string;
  email?: string;
  photoURL?: string;
}

// StateBundle is now imported from './game-state'

/**
 * Animation state for managing state bundle playback queue.
 */
export interface AnimationState {
  /** Queue of state bundles waiting to be played */
  pendingBundles: StateBundle[];
  /** Last bundle that was fired/installed */
  lastFiredBundle: StateBundle | null;
  /** IDs of currently active animations */
  activeAnimations: string[];
}

/**
 * Version tracking state for managing game state versions and fetching.
 */
export interface VersionState {
  /** Current game version (from latest installed state) */
  current: number;
  /** Target version to fetch (set by WebSocket or manual navigation) */
  target: number;
  /** Last version successfully fetched from server */
  lastFetched: number;
}

/**
 * WebSocket connection state.
 */
export interface SocketState {
  /** Whether WebSocket is currently connected */
  connected: boolean;
  /** Number of connection attempts (for backoff) */
  connectionAttempts: number;
  /** Last error message (null if no error) */
  lastError: string | null;
}

/**
 * View state for the current game view.
 */
export interface ViewState {
  /** Full game object from server */
  game: any | null;
  /** Player index currently viewing as */
  viewingAsPlayer: number;
  /** Player index requested (before applying auto-current-player) */
  requestedPlayer: number;
  /** Whether to automatically follow the current player */
  autoCurrentPlayer: boolean;
  /** Move forms for the current state */
  moveForms: MoveForm[] | null;
}

/**
 * Game state containing game data, players, and current state.
 */
export interface GameState {
  /** Game ID */
  id: string;
  /** Game type name (e.g., 'blackjack', 'memory') */
  name: string;
  /** Game chest containing deck configurations and components */
  chest: GameChest | null;
  /** Information about all players in the game */
  playersInfo: PlayerInfo[];
  /** Whether game has empty player slots */
  hasEmptySlots: boolean;
  /** Whether game is open for new players */
  open: boolean;
  /** Whether game is visible to non-players */
  visible: boolean;
  /** Whether current user is the game owner */
  isOwner: boolean;
  /** Current RAW game state from server (unexpanded - use selectExpandedGameState selector to get expanded version) */
  currentState: RawGameState | null;
  /** Timer metadata for expansion (maps timer ID to TimerInfo) */
  timerInfos: Record<string, TimerInfo> | null;
  /** Paths in state that need timer ticks (can contain strings and numbers for array indices) */
  pathsToTick: (string | number)[][];
  /** Original wall clock time when state was loaded */
  originalWallClockTime: number;
  /** Animation system state (bundle queue and playback) */
  animation: AnimationState;
  /** Version tracking state */
  versions: VersionState;
  /** WebSocket connection state */
  socket: SocketState;
  /** View state (game object, viewing player, move forms) */
  view: ViewState;
  /** Fetched info data from fetchGameInfo (null when not available or after processing) */
  fetchedInfo: any | null;
  /** Fetched version data from fetchGameVersion (null when not available or after processing) */
  fetchedVersion: any | null;
  /** Whether a fetch operation is in progress */
  loading: boolean;
  /** Last error message from fetch operations (null if no error) */
  error: string | null;
}

/**
 * Game chest containing component configurations and deck defaults.
 */
export interface GameChest {
  [key: string]: any;
  // TODO: Define specific chest structure based on game architecture
}

/**
 * Expanded game state with all component references resolved.
 * State from server has component indices that get expanded to full objects.
 */
export interface ExpandedGameState {
  [key: string]: any;
  // TODO: Define specific state structure based on game architecture
}

/**
 * Information about a player in the game.
 */
export interface PlayerInfo {
  /** Player index in game */
  index: number;
  /** Player display name */
  name: string;
  /** Player photo URL */
  photoURL?: string;
  /** Whether this slot is empty */
  isEmpty?: boolean;
}

/**
 * List view state containing games list and filters.
 */
export interface ListState {
  /** Game type filter */
  gameTypeFilter: string;
  /** Index of selected game manager */
  selectedManagerIndex: number;
  /** Number of players for new game */
  numPlayers: number;
  /** Agent names for each player slot */
  agents: string[];
  /** Variant options for new game */
  variantOptions: number[];
  /** Whether new game should be visible */
  visible: boolean;
  /** Whether new game should be open */
  open: boolean;
  /** Available game managers */
  managers: any[]; // TODO: Define GameManager type
  /** All games */
  allGames: GameListItem[];
  /** Games user is participating in (active) */
  participatingActiveGames: GameListItem[];
  /** Games user participated in (finished) */
  participatingFinishedGames: GameListItem[];
  /** Visible active games */
  visibleActiveGames: GameListItem[];
  /** Visible games that can be joined */
  visibleJoinableGames: GameListItem[];
}

/**
 * Summary of a game shown in the games list.
 */
export interface GameListItem {
  /** Game ID */
  id: string;
  /** Game type name */
  name: string;
  /** Number of players */
  numPlayers: number;
  /** Whether game is open */
  open: boolean;
  /** Whether game is visible */
  visible: boolean;
  /** Game owner */
  owner: string;
  /** Creation timestamp */
  created: number;
  /** Last modified timestamp */
  modified: number;
}
