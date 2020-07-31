export const selectManagers = (state) => state.list ? state.list.managers : [];
export const selectParticipatingActiveGames = (state) => state.list ? state.list.participatingActiveGames : [];
export const selectParticipatingFinishedGames = (state) => state.list ? state.list.participatingFinishedGames : [];
export const selectVisibleActiveGames = (state) => state.list ? state.list.visibleActiveGames : [];
export const selectVisibleJoinableGames = (state) => state.list ? state.list.visibleJoinableGames : [];
export const selectAllGames = (state) => state.list ? state.list.allGames : [];
