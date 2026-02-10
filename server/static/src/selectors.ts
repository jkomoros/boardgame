import { createSelector } from 'reselect';
import type { RootState, GameChest, PlayerInfo, ExpandedGameState, UserInfo } from './types/store';

// App selectors
export const selectPage = (state: RootState): string => state.app ? state.app.page : "";
export const selectPageExtra = (state: RootState): string => state.app ? state.app.pageExtra : "";

// List selectors
export const selectManagers = (state: RootState): any[] => state.list ? state.list.managers : [];
export const selectSelectedManagerIndex = (state: RootState): number => state.list ? state.list.selectedManagerIndex : 0;
export const selectCreateGameNumPlayers = (state: RootState): number => state.list ? state.list.numPlayers : 0;
export const selectCreateGameAgents = (state: RootState): any[] => state.list ? state.list.agents : [];
export const selectCreateGameVariantOptions = (state: RootState): any[] => state.list ? state.list.variantOptions : [];
export const selectCreateGameOpen = (state: RootState): boolean => state.list ? state.list.open : false;
export const selectCreateGameVisible = (state: RootState): boolean => state.list ? state.list.visible : false;
export const selectGameTypeFilter = (state: RootState): string => state.list ? state.list.gameTypeFilter : "";
export const selectParticipatingActiveGames = (state: RootState): any[] => state.list ? state.list.participatingActiveGames : [];
export const selectParticipatingFinishedGames = (state: RootState): any[] => state.list ? state.list.participatingFinishedGames : [];
export const selectVisibleActiveGames = (state: RootState): any[] => state.list ? state.list.visibleActiveGames : [];
export const selectVisibleJoinableGames = (state: RootState): any[] => state.list ? state.list.visibleJoinableGames : [];
export const selectAllGames = (state: RootState): any[] => state.list ? state.list.allGames : [];

// Error selectors
export const selectErrorMessage = (state: RootState): string => state.error ? state.error.message : "";
export const selectErrorFriendlyMessage = (state: RootState): string => state.error ? state.error.friendlyMessage : "";
export const selectErrorTitle = (state: RootState): string => state.error ? state.error.title : "";
export const selectErrorShowing = (state: RootState): boolean => state.error ? state.error.showing : false;

// User selectors
export const selectUser = (state: RootState): UserInfo | null => state.user ? state.user.user : null;
export const selectVerifyingAuth = (state: RootState): boolean => state.user ? state.user.verifyingAuth : false;
export const selectLoggedIn = (state: RootState): boolean => state.user ? state.user.loggedIn : false;
export const selectAdminAllowed = (state: RootState): boolean => state.user ? state.user.adminAllowed : false;
export const selectSignInErrorMessage = (state: RootState): string => state.user ? state.user.errorMessage : "";
export const selectAdmin = (state: RootState): boolean => state.user ? state.user.admin : false;
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
export const selectGameCurrentState = (state: RootState): ExpandedGameState | null => state.game ? state.game.currentState : null;
export const selectGameLoading = (state: RootState): boolean => state.game ? state.game.loading : false;
export const selectGameError = (state: RootState): string | null => state.game ? state.game.error : null;

const selectGameID = (state: RootState): string => state.game ? state.game.id : '';
export const selectGameName = (state: RootState): string => state.game ? state.game.name : '';

export const selectGameRoute = createSelector(
    selectGameID,
    selectGameName,
    (id, name): { id: string; name: string } | null => id ? {id, name} : null
);

// Animation selectors
export const selectAnimationState = (state: RootState) => state.game?.animation || {
    pendingBundles: [],
    lastFiredBundle: null,
    activeAnimations: []
};
export const selectPendingBundles = (state: RootState): any[] => selectAnimationState(state).pendingBundles;
export const selectLastFiredBundle = (state: RootState): any | null => selectAnimationState(state).lastFiredBundle;
export const selectActiveAnimations = (state: RootState): string[] => selectAnimationState(state).activeAnimations;
export const selectHasPendingBundles = (state: RootState): boolean => selectPendingBundles(state).length > 0;
export const selectNextBundle = (state: RootState): any | null => {
    const bundles = selectPendingBundles(state);
    return bundles.length > 0 ? bundles[0] : null;
};

// Version selectors
export const selectVersionState = (state: RootState) => state.game?.versions || {
    current: 0,
    target: -1,
    lastFetched: 0
};
export const selectCurrentVersion = (state: RootState): number => selectVersionState(state).current;
export const selectTargetVersion = (state: RootState): number => selectVersionState(state).target;
export const selectLastFetchedVersion = (state: RootState): number => selectVersionState(state).lastFetched;

// Socket selectors
export const selectSocketState = (state: RootState) => state.game?.socket || {
    connected: false,
    connectionAttempts: 0,
    lastError: null
};
export const selectSocketConnected = (state: RootState): boolean => selectSocketState(state).connected;
export const selectSocketConnectionAttempts = (state: RootState): number => selectSocketState(state).connectionAttempts;
export const selectSocketError = (state: RootState): string | null => selectSocketState(state).lastError;

// View selectors
export const selectViewState = (state: RootState) => state.game?.view || {
    game: null,
    viewingAsPlayer: 0,
    requestedPlayer: 0,
    autoCurrentPlayer: false,
    moveForms: null
};
export const selectGame = (state: RootState): any | null => selectViewState(state).game;
export const selectViewingAsPlayer = (state: RootState): number => selectViewState(state).viewingAsPlayer;
export const selectRequestedPlayer = (state: RootState): number => selectViewState(state).requestedPlayer;
export const selectAutoCurrentPlayer = (state: RootState): boolean => selectViewState(state).autoCurrentPlayer;
export const selectMoveForms = (state: RootState): any[] | null => selectViewState(state).moveForms;

// Internal selector for timer infos (will be added to state)
const selectGameTimerInfos = (state: RootState): Record<string, any> | null =>
    state.game?.timerInfos || null;

/**
 * Memoized selector that expands raw game state.
 * Expansion adds synthetic properties (Components, GameName, TimeLeft) without mutating.
 * This is the PURE version - returns new objects, never mutates inputs.
 */
export const selectExpandedGameState = createSelector(
    [selectGameCurrentState, selectGameChest, selectGameName, selectGameTimerInfos],
    (rawState, chest, gameName, timerInfos): ExpandedGameState | null => {
        if (!rawState || !chest) return null;

        // Pure expansion - returns new object tree
        const expandedGame = expandLeafState(rawState, rawState.Game, chest, gameName, timerInfos);
        const expandedPlayers = rawState.Players.map((player: any) =>
            expandLeafState(rawState, player, chest, gameName, timerInfos)
        );

        return {
            ...rawState,
            Game: expandedGame,
            Players: expandedPlayers,
        };
    }
);

/**
 * Pure function to expand a leaf state object.
 * Walks properties and expands stacks and timers inline.
 */
const expandLeafState = (
    wholeState: any,
    leafState: any,
    chest: GameChest,
    gameName: string,
    timerInfos: Record<string, any> | null
): any => {
    const result = { ...leafState };

    Object.entries(leafState).forEach(([key, val]) => {
        // Skip null and non-objects
        if (!val || typeof val !== 'object') return;

        // Expand stacks (objects with Deck property)
        if ((val as any).Deck) {
            result[key] = expandStack(val, wholeState, chest, gameName);
        }
        // Expand timers (objects with IsTimer property)
        else if ((val as any).IsTimer) {
            result[key] = expandTimer(val, timerInfos);
        }
    });

    // Copy in Player computed state if it exists
    const pathToLeaf = getPathToLeaf(wholeState, leafState);
    if (pathToLeaf?.length === 2 && pathToLeaf[0] === 'Players') {
        const playerIndex = pathToLeaf[1];
        if (wholeState.Computed?.Players?.[playerIndex]) {
            result.Computed = wholeState.Computed.Players[playerIndex];
        }
    }

    return result;
};

/**
 * Pure function to expand a stack (deck of components).
 * Returns new object with Components and GameName added.
 */
const expandStack = (
    stack: any,
    wholeState: any,
    chest: GameChest,
    gameName: string
): any => {
    if (!stack.Deck) return stack;

    const components = stack.Indexes.map((index: number, i: number) => {
        if (index === -1) return null;

        // Generic component (index -2)
        if (index === -2) return {};

        // Resolve actual component
        const component = componentForDeckAndIndex(stack.Deck, index, wholeState, chest);
        if (!component) return null;

        // Add ID if available
        const result = { ...component };
        if (stack.IDs?.[i]) {
            result.ID = stack.IDs[i];
        }
        result.Deck = stack.Deck;
        result.GameName = gameName;

        return result;
    });

    return {
        ...stack,
        GameName: gameName,
        Components: components,
    };
};

/**
 * Pure function to expand a timer.
 * Returns new object with TimeLeft and originalTimeLeft added.
 * originalTimeLeft is the value when first received; TimeLeft is updated by tick.
 */
const expandTimer = (
    timer: any,
    timerInfos: Record<string, any> | null
): any => {
    const result = {
        ...timer,
        TimeLeft: 0,
        originalTimeLeft: 0,
    };

    if (timerInfos?.[timer.ID]) {
        const info = timerInfos[timer.ID];
        result.TimeLeft = info.TimeLeft;
        // originalTimeLeft should be preserved from when timer was first installed
        result.originalTimeLeft = info.originalTimeLeft ?? info.TimeLeft;
    }

    return result;
};

/**
 * Helper to get component for a deck and index.
 */
const componentForDeckAndIndex = (
    deckName: string,
    index: number,
    wholeState: any,
    chest: GameChest
): any | null => {
    const deck = (chest as any).Decks?.[deckName];
    if (!deck) return null;

    const result = { ...deck[index] };

    // Add dynamic values if available
    if (wholeState.Components?.[deckName]?.[index]) {
        result.DynamicValues = wholeState.Components[deckName][index];
    }

    return result;
};

/**
 * Helper to determine the path to a leaf state within the whole state.
 * Used to identify Player states for Computed property copying.
 */
const getPathToLeaf = (wholeState: any, leafState: any): string[] | null => {
    // Check if it's the Game
    if (wholeState.Game === leafState) {
        return ['Game'];
    }

    // Check if it's a Player
    if (wholeState.Players) {
        const index = wholeState.Players.indexOf(leafState);
        if (index !== -1) {
            return ['Players', index];
        }
    }

    return null;
};
