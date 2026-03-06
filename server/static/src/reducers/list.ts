import type { Reducer } from 'redux';
import type { ListState } from '../types/store';
import type { ListAction } from '../actions/list.js';
import {
    UPDATE_MANAGERS,
    UPDATE_GAMES_LIST,
    UPDATE_GAME_TYPE_FILTER,
    UPDATE_SELECTED_MANAGER_INDEX,
    UPDATE_NUM_PLAYERS,
    UPDATE_AGENT_NAME,
    UPDATE_VARIANT_OPTION,
    UPDATE_CREATE_GAME_VISIBLE,
    UPDATE_CREATE_GAME_OPEN
} from '../actions/list.js';

const INITIAL_STATE: ListState = {
    gameTypeFilter: "",
    selectedManagerIndex: -1,
    numPlayers: 0,
    agents: [],
    variantOptions: [],
    visible: false,
    open: false,
    managers: [],
    allGames: [],
    participatingActiveGames: [],
    participatingFinishedGames: [],
    visibleActiveGames: [],
    visibleJoinableGames: [],
};

const list: Reducer<ListState, ListAction> = (state = INITIAL_STATE, action): ListState => {
	switch (action.type) {
	case UPDATE_MANAGERS:
        const newNumPlayers = action.managers[0].DefaultNumPlayers || 0
        const newNumVariantOptions = (action.managers[0].Variant || []).length;
		return {
			...state,
            managers: action.managers,
            selectedManagerIndex: 0,
            numPlayers: newNumPlayers,
            agents: Array(newNumPlayers).fill(""),
            variantOptions: Array(newNumVariantOptions).fill(0)
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
        const newManager = state.managers[action.index]
        const updatedNumPlayers =  newManager ? newManager.DefaultNumPlayers || 0 : 0;
        const updatedNumVariantOptions = newManager ? (newManager.Variant || []).length : 0;
        return {
            ...state,
            selectedManagerIndex: action.index,
            numPlayers: updatedNumPlayers,
            agents: Array(updatedNumPlayers).fill(""),
            variantOptions: Array(updatedNumVariantOptions).fill(0)
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
    case UPDATE_AGENT_NAME:
        const modifiedAgents = [...state.agents];
        modifiedAgents[action.index] = action.name;
        return {
            ...state,
            agents: modifiedAgents
        }
    case UPDATE_VARIANT_OPTION:
        const modifiedVariantOptions = [...state.variantOptions];
        modifiedVariantOptions[action.variantIndex] = action.optionIndex;
        return {
            ...state,
            variantOptions: modifiedVariantOptions
        }
    case UPDATE_CREATE_GAME_OPEN:
        return {
            ...state,
            open: action.open
        }
    case UPDATE_CREATE_GAME_VISIBLE:
        return {
            ...state,
            visible: action.visible
        }
	default:
		return state;
	}
};

export default list;
