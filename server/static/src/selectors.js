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

const selectGameID = (state) => state.game ? state.game.id : '';
const selectGameName = (state) => state.game ? state.game.name : '';

export const selectGameRoute = createSelector(
    selectGameID,
    selectGameName,
    (id, name) => id ? {id, name} : null
);