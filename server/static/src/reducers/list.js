import {
    UPDATE_MANAGERS,
    UPDATE_GAMES_LIST,
    UPDATE_GAME_TYPE_FILTER,
    UPDATE_SELECTED_MANAGER_INDEX,
    UPDATE_NUM_PLAYERS
} from '../actions/list.js';

const INITIAL_STATE = {
    gameTypeFilter: "",
    selectedManagerIndex: -1,
    numPlayers: 0,
    agents: [],
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
        const newNumPlayers = action.managers[0].DefaultNumPlayers || 0
		return {
			...state,
            managers: action.managers,
            selectedManagerIndex: 0,
            numPlayers: newNumPlayers,
            agents: Array(newNumPlayers).fill("")
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
        const updatedNumPlayers = state.managers[action.index] ? state.managers[action.index].DefaultNumPlayers || 0 : 0;
        return {
            ...state,
            selectedManagerIndex: action.index,
            numPlayers: updatedNumPlayers,
            agents: Array(updatedNumPlayers).fill("")
        };
    case UPDATE_NUM_PLAYERS:
        const newAgents = [...state.agents];
        while (action.numPlayers > newAgents.length) {
            newAgents.push("");
        }
        while (action.numPlayers < newAgents.length) {
            newAgents.pop();
        }
        return {
            ...state,
            numPlayers: action.numPlayers,
            agents: newAgents,
        }
	default:
		return state;
	}
};

export default app;
