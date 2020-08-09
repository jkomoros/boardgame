import {
    UPDATE_MANAGERS,
    UPDATE_GAMES_LIST,
    UPDATE_GAME_TYPE_FILTER,
    UPDATE_SELECTED_MANAGER_INDEX
} from '../actions/list.js';

const INITIAL_STATE = {
    gameTypeFilter: "",
    selectedManagerIndex: -1,
    managers: [],
    allGames: [],
    participatingActiveGames: [],
    participatingFinishedGames: [],
    visibleActiveGames: [],
    visibleJoinableGames: [],
};

const app = (state = INITIAL_STATE, action) => {
	switch (action.type) {
	case UPDATE_MANAGERS:
		return {
			...state,
            managers: action.managers,
            selectedManagerIndex: 0,
        };
    case UPDATE_GAME_TYPE_FILTER:
        return {
            ...state,
            gameTypeFilter: action.name
        };
    case UPDATE_GAMES_LIST:
        return {
            ...state,
            participatingActiveGames: action.participatingActiveGames,
            participatingFinishedGames: action.participatingFinishedGames,
            visibleActiveGames: action.visibleActiveGames,
            visibleJoinableGames: action.visibleJoinableGames,
            allGames: action.allGames || [],
        };
    case UPDATE_SELECTED_MANAGER_INDEX:
        return {
            ...state,
            selectedManagerIndex: action.index
        };
	default:
		return state;
	}
};

export default app;
