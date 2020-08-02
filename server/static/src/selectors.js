import { createSelector } from 'reselect';

export const selectPage = (state) => state.app ? state.app.page : "";
export const selectPageExtra = (state) => state.app ? state.app.pageExtra : "";
export const selectManagers = (state) => state.list ? state.list.managers : [];
export const selectGameTypeFilter = (state) => state.list ? state.list.gameTypeFilter : "";
export const selectParticipatingActiveGames = (state) => state.list ? state.list.participatingActiveGames : [];
export const selectParticipatingFinishedGames = (state) => state.list ? state.list.participatingFinishedGames : [];
export const selectVisibleActiveGames = (state) => state.list ? state.list.visibleActiveGames : [];
export const selectVisibleJoinableGames = (state) => state.list ? state.list.visibleJoinableGames : [];
export const selectAllGames = (state) => state.list ? state.list.allGames : [];
export const selectErrorMessage = (state) => state.error ? state.error.message : "";
export const selectErrorFriendlyMessage = (state) => state.error ? state.error.friendlyMessage : "";
export const selectErrorTitle = (state) => state.error ? state.error.title : "";
export const selectErrorShowing = (state) => state.error ? state.error.showing : "";
export const selectUser = (state) => state.user ? state.user.user : null;
export const selectVerifyingAuth = (state) => state.user ? state.user.verifyingAuth : false;
export const selectLoggedIn = (state) => state.user ? state.user.loggedIn : false;
export const selectAdminAllowed = (state) => state.user ? state.user.adminAllowed : false;
export const selectSignInErrorMessage = (state) => state.user ? state.user.errorMessage : "";

const selectGameID = (state) => state.game ? state.game.id : '';
const selectGameName = (state) => state.game ? state.game.name : '';

export const selectGameRoute = createSelector(
    selectGameID,
    selectGameName,
    (id, name) => id ? {id, name} : null
);