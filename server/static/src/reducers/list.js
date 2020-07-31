import {
    UPDATE_MANAGERS,
    UPDATE_GAMES_LIST
} from '../actions/list.js';

const INITIAL_STATE = {
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
	default:
		return state;
	}
};

export default app;
