import { createSelector } from 'reselect';
import type {
    CompanionInfo,
    ExpandedGameState,
    FetchedGameInfo,
    FetchedGameVersion,
    GameChest,
    GameListItem,
    ManagerInfo,
    PlayerInfo,
    RootState,
    UserInfo,
} from './types/store';
import type { GameFromServer, RawGameState, TimerInfo, StateBundle } from './types/game-state';
import type { MoveForm, PreconditionEntry } from './types/api';
import { expandGameStateSnapshot } from './state/expand.js';

// The expansion itself is a pure layer-1 function (src/state/expand.ts) so a
// host without Redux can use it; it is re-exported here for existing callers.
export { expandGameStateSnapshot };

// Stable default objects to prevent creating new objects on every selector call
const DEFAULT_ANIMATION_STATE = {
    pendingBundles: [],
    lastFiredBundle: null,
    activeAnimations: []
};

const DEFAULT_VERSION_STATE = {
    current: 0,
    target: -1,
    lastFetched: 0
};

const DEFAULT_SOCKET_STATE = {
    connected: false,
    connectionAttempts: 0,
    lastError: null
};

const DEFAULT_VIEW_STATE = {
    game: null,
    viewingAsPlayer: 0,
    requestedPlayer: 0,
    autoCurrentPlayer: false,
    moveForms: null
};

// App selectors - memoized for performance as they're frequently accessed
export const selectPage = createSelector(
    [(state: RootState) => state.app?.page],
    (page): string => page || ""
);
export const selectPageExtra = (state: RootState): string => state.app ? state.app.pageExtra : "";

// List selectors
export const selectManagers = (state: RootState): ManagerInfo[] => state.list ? state.list.managers : [];
export const selectSelectedManagerIndex = (state: RootState): number => state.list ? state.list.selectedManagerIndex : 0;
export const selectCreateGameNumPlayers = (state: RootState): number => state.list ? state.list.numPlayers : 0;
export const selectCreateGameAgents = (state: RootState): string[] => state.list ? state.list.agents : [];
export const selectCreateGameVariantOptions = (state: RootState): number[] => state.list ? state.list.variantOptions : [];
export const selectCreateGameOpen = (state: RootState): boolean => state.list ? state.list.open : false;
export const selectCreateGameVisible = (state: RootState): boolean => state.list ? state.list.visible : false;
export const selectGameTypeFilter = (state: RootState): string => state.list ? state.list.gameTypeFilter : "";
export const selectParticipatingActiveGames = (state: RootState): GameListItem[] => state.list ? state.list.participatingActiveGames : [];
export const selectParticipatingFinishedGames = (state: RootState): GameListItem[] => state.list ? state.list.participatingFinishedGames : [];
export const selectVisibleActiveGames = (state: RootState): GameListItem[] => state.list ? state.list.visibleActiveGames : [];
export const selectVisibleJoinableGames = (state: RootState): GameListItem[] => state.list ? state.list.visibleJoinableGames : [];
export const selectAllGames = (state: RootState): GameListItem[] => state.list ? state.list.allGames : [];

// Error selectors
export const selectErrorMessage = (state: RootState): string => state.error ? state.error.message : "";
export const selectErrorFriendlyMessage = (state: RootState): string => state.error ? state.error.friendlyMessage : "";
export const selectErrorTitle = (state: RootState): string => state.error ? state.error.title : "";
export const selectErrorShowing = (state: RootState): boolean => state.error ? state.error.showing : false;

// User selectors - memoized for high-frequency checks
export const selectUser = (state: RootState): UserInfo | null => state.user ? state.user.user : null;
export const selectVerifyingAuth = (state: RootState): boolean => state.user ? state.user.verifyingAuth : false;
export const selectLoggedIn = createSelector(
    [(state: RootState) => state.user?.loggedIn],
    (loggedIn): boolean => loggedIn || false
);
export const selectAdminAllowed = (state: RootState): boolean => state.user ? state.user.adminAllowed : false;
export const selectSignInErrorMessage = (state: RootState): string => state.user ? state.user.errorMessage : "";
export const selectAdmin = createSelector(
    [(state: RootState) => state.user?.admin],
    (admin): boolean => admin || false
);
export const selectSignInDialogOpen = (state: RootState): boolean => state.user ? state.user.dialogOpen : false;
export const selectSignInDialogEmail = (state: RootState): string => state.user ? state.user.dialogEmail : "";
export const selectSignInDialogPassword = (state: RootState): string => state.user ? state.user.dialogPassword : "";
export const selectSignInDialogIsCreate = (state: RootState): boolean => state.user ? state.user.dialogIsCreate : false;
export const selectSignInDialogSelectedPage = (state: RootState): number => state.user ? state.user.dialogSelectedPage : 0;

// Game selectors
export const selectGameChest = (state: RootState): GameChest | null => state.game ? state.game.chest : null;
export const selectGamePlayersInfo = (state: RootState): PlayerInfo[] => state.game ? state.game.playersInfo : [];
export const selectGameHasEmptySlots = (state: RootState): boolean => state.game ? state.game.hasEmptySlots : false;
export const selectGameOpen = (state: RootState): boolean => state.game ? state.game.open : false;
export const selectGameVisible = (state: RootState): boolean => state.game ? state.game.visible : false;
export const selectGameIsOwner = (state: RootState): boolean => state.game ? state.game.isOwner : false;
// selectGameCompanionInfo returns the doGameInfo CompanionInfo bundle for
// the active game, or null for solo-mode games / before the first
// gameInfo fetch.
export const selectGameCompanionInfo = (state: RootState): CompanionInfo | null => state.game ? state.game.companionInfo : null;
// Returns raw game state (not expanded) - use selectExpandedGameState for expanded version
export const selectGameCurrentState = (state: RootState): RawGameState | null => state.game ? state.game.currentState : null;
// Per-operation loading selectors
export const selectMoveSubmitting = (state: RootState): boolean => state.game ? state.game.moveSubmitting : false;
export const selectVersionFetching = (state: RootState): boolean => state.game ? state.game.versionFetching : false;
export const selectInfoFetching = (state: RootState): boolean => state.game ? state.game.infoFetching : false;
export const selectConfiguring = (state: RootState): boolean => state.game ? state.game.configuring : false;
// Combined loading selector for backward compatibility
export const selectGameLoading = (state: RootState): boolean =>
    state.game ? (state.game.moveSubmitting || state.game.versionFetching || state.game.infoFetching || state.game.configuring) : false;
export const selectGameError = (state: RootState): string | null => state.game ? state.game.error : null;

// Fetched data selectors (for async thunk responses)
export const selectFetchedInfo = (state: RootState): FetchedGameInfo | null => state.game?.fetchedInfo || null;
export const selectFetchedVersion = (state: RootState): FetchedGameVersion | null => state.game?.fetchedVersion || null;

const selectGameID = (state: RootState): string => state.game ? state.game.id : '';
export const selectGameName = (state: RootState): string => state.game ? state.game.name : '';

export const selectGameRoute = createSelector(
    selectGameID,
    selectGameName,
    (id, name): { id: string; name: string } | null => id ? {id, name} : null
);

// Animation selectors - memoized for performance as they're checked frequently
export const selectAnimationState = (state: RootState) =>
    state.game?.animation || DEFAULT_ANIMATION_STATE;

// Memoized to prevent creating new array references on every call
export const selectPendingBundles = createSelector(
    [selectAnimationState],
    (animationState): StateBundle[] => animationState.pendingBundles
);

export const selectLastFiredBundle = (state: RootState): StateBundle | null => selectAnimationState(state).lastFiredBundle;
export const selectActiveAnimations = (state: RootState): string[] => selectAnimationState(state).activeAnimations;

// Memoized to prevent recomputation on every render
export const selectHasPendingBundles = createSelector(
    [selectPendingBundles],
    (bundles): boolean => bundles.length > 0
);

// Memoized to prevent unnecessary recalculations
export const selectNextBundle = createSelector(
    [selectPendingBundles],
    (bundles): StateBundle | null => bundles[0] ?? null
);

// Version selectors
export const selectVersionState = (state: RootState) =>
    state.game?.versions || DEFAULT_VERSION_STATE;
export const selectCurrentVersion = (state: RootState): number => selectVersionState(state).current;
export const selectTargetVersion = (state: RootState): number => selectVersionState(state).target;
export const selectLastFetchedVersion = (state: RootState): number => selectVersionState(state).lastFetched;

// Socket selectors
export const selectSocketState = (state: RootState) =>
    state.game?.socket || DEFAULT_SOCKET_STATE;
export const selectSocketConnected = (state: RootState): boolean => selectSocketState(state).connected;
export const selectSocketConnectionAttempts = (state: RootState): number => selectSocketState(state).connectionAttempts;
export const selectSocketError = (state: RootState): string | null => selectSocketState(state).lastError;

// View selectors
export const selectViewState = (state: RootState) =>
    state.game?.view || DEFAULT_VIEW_STATE;
export const selectGame = (state: RootState): GameFromServer | null => selectViewState(state).game;
export const selectViewingAsPlayer = (state: RootState): number => selectViewState(state).viewingAsPlayer;
export const selectRequestedPlayer = (state: RootState): number => selectViewState(state).requestedPlayer;
export const selectAutoCurrentPlayer = (state: RootState): boolean => selectViewState(state).autoCurrentPlayer;
export const selectMoveForms = (state: RootState): MoveForm[] | null => selectViewState(state).moveForms;

/**
 * Memoized selector deriving a legality map from move forms.
 * Returns a record keyed by move name with legality info for each move.
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

export const selectMoveLegality = createSelector(
    [selectMoveForms],
    (moveForms): Record<string, MoveLegalityInfo> => {
        const result: Record<string, MoveLegalityInfo> = {};
        if (!moveForms) return result;
        for (const form of moveForms) {
            result[form.Name] = {
                legalForPlayer: form.LegalForPlayer ?? false,
                legalForAnyone: form.LegalForAnyone ?? false,
                ...(form.LegalForPlayerError !== undefined ? { error: form.LegalForPlayerError } : {}),
                ...(form.Preconditions !== undefined ? { preconditions: form.Preconditions } : {}),
            };
        }
        return result;
    }
);

// Internal selector for timer infos (will be added to state)
export const selectGameTimerInfos = (state: RootState): Record<string, TimerInfo> | null =>
    state.game?.timerInfos || null;

/**
 * PERFORMANCE CRITICAL: Base expanded game state WITHOUT timer expansion.
 * This selector does NOT depend on timerInfos, which updates 60+ times/second.
 * By excluding timer expansion here, we prevent full game state re-expansion
 * on every timer tick, which would cause massive performance overhead.
 *
 * Use this for most game state access where live timer updates aren't needed.
 */
export const selectExpandedGameStateWithoutTimers = createSelector(
    [selectGameCurrentState, selectGameChest, selectGameName],
    (rawState, chest, gameName): ExpandedGameState | null => {
        if (!rawState || !chest) return null;

        return expandGameStateSnapshot(rawState, chest, gameName);

    }
);

/**
 * PERFORMANCE CRITICAL: Expanded game state WITH timer updates.
 * This selector depends on timerInfos (updates 60+ times/second) and should
 * ONLY be used by components that need live timer countdowns.
 *
 * Most components should use selectExpandedGameStateWithoutTimers instead
 * to avoid unnecessary re-renders on every timer tick.
 */
export const selectTimerExpandedGameState = createSelector(
    [selectGameCurrentState, selectGameChest, selectGameName, selectGameTimerInfos],
    (rawState, chest, gameName, timerInfos): ExpandedGameState | null => {
        if (!rawState || !chest) return null;

        // Pure expansion with timer info included
        return expandGameStateSnapshot(rawState, chest, gameName, timerInfos);

    }
);

/**
 * Main game state selector - uses the non-timer version by default.
 * This prevents the common case (displaying game state) from re-rendering
 * 60+ times/second due to timer ticks.
 */
export const selectExpandedGameState = selectExpandedGameStateWithoutTimers;

/**
 * Returns a wrapper that deduplicates array results by shallow comparison.
 * Prevents unnecessary re-renders when the array contents haven't changed
 * but createSelector returns a new reference due to input object identity.
 */
function stableArray<T>(fn: (state: any) => T[]): (state: any) => T[] {
    let prev: T[] = [];
    return (state: any): T[] => {
        const next = fn(state);
        if (next.length === prev.length && next.every((v, i) => v === prev[i])) {
            return prev;
        }
        prev = next;
        return next;
    };
}

function stableNullableArray<T>(fn: (state: any) => T[] | null): (state: any) => T[] | null {
    let prev: T[] | null = null;
    return (state: any): T[] | null => {
        const next = fn(state);
        if (next === null && prev === null) return prev;
        if (next === null || prev === null) { prev = next; return next; }
        if (next.length === prev!.length && next.every((v, i) => v === prev![i])) {
            return prev;
        }
        prev = next;
        return next;
    };
}

/**
 * Selector for player colors from computed properties.
 * Returns an array of CSS color strings, one per player.
 * Uses stable reference to avoid unnecessary re-renders.
 */
const _selectPlayerColors = createSelector(
    [selectExpandedGameStateWithoutTimers],
    (state): string[] => state?.Players?.map(
        (p: any) => p?.Computed?.Color || ''
    ) || []
);
export const selectPlayerColors = stableArray(_selectPlayerColors);

/**
 * Selector for player activity from computed properties.
 * Returns an array of booleans, true if the player may be active.
 * Uses stable reference to avoid unnecessary re-renders.
 */
const _selectPlayerActivity = createSelector(
    [selectExpandedGameStateWithoutTimers],
    (state): boolean[] => state?.Players?.map(
        (p: any) => p?.Computed?.MayBeActive !== false
    ) || []
);
export const selectPlayerActivity = stableArray(_selectPlayerActivity);

/**
 * Selector for custom player order from global computed properties.
 * Returns an array of player indices in display order, or null if default.
 * Uses stable reference to avoid unnecessary re-renders.
 */
const _selectPlayerOrder = createSelector(
    [selectExpandedGameStateWithoutTimers],
    (state): number[] | null => {
        const game = state?.Game as Readonly<Record<string, unknown>> | undefined;
        const computed = game?.['Computed'];
        if (!computed || typeof computed !== 'object' || Array.isArray(computed)) return null;
        const order = (computed as Readonly<Record<string, unknown>>)['PlayerOrder'];
        if (!Array.isArray(order)
            || !order.every(playerIndex => Number.isSafeInteger(playerIndex) && playerIndex >= 0)) {
            return null;
        }
        return order;
    }
);
export const selectPlayerOrder = stableNullableArray(_selectPlayerOrder);
