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
