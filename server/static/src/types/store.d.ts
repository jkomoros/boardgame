/**
 * Type definitions for the Redux store structure.
 *
 * The store uses lazy reducer loading - app, error, and user are always loaded,
 * while game and list are loaded on-demand when navigating to those views.
 */

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
  /** Current expanded game state (null if not loaded) */
  currentState: ExpandedGameState | null;
  /** Paths in state that need timer ticks */
  pathsToTick: string[];
  /** Original wall clock time when state was loaded */
  originalWallClockTime: number;
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
  /** List of games */
  games: GameListItem[];
  /** Whether list is currently loading */
  loading: boolean;
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
