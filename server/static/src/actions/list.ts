import type { ThunkAction } from 'redux-thunk';
import type { RootState, GameListItem, ManagerInfo } from '../types/store';
import type { UserAction } from './user.js';
import type { AppAction } from './app.js';
import type { ErrorAction } from './error.js';
import { apiGet, apiPost, buildApiUrl } from '../api.js';
import {
    decodeCreateGameResponse,
    decodeGamesListResponse,
    decodeManagersResponse,
} from '../types/list-response.js';

export const UPDATE_MANAGERS = 'UPDATE_MANAGERS';
export const UPDATE_GAMES_LIST = 'UPDATE_GAMES_LIST';
export const UPDATE_GAME_TYPE_FILTER = 'UPDATE_GAME_TYPE_FILTER';
export const UPDATE_SELECTED_MANAGER_INDEX = "UPDATE_SELECTED_MANAGER_INDEX";
export const UPDATE_NUM_PLAYERS = "UPDATE_NUM_PLAYERS";
export const UPDATE_AGENT_NAME = "UPDATE_AGENT_NAME";
export const UPDATE_VARIANT_OPTION = "UPDATE_VARIANT_OPTION";
export const UPDATE_CREATE_GAME_OPEN = "UPDATE_CREATE_GAME_OPEN";
export const UPDATE_CREATE_GAME_VISIBLE = "UPDATE_CREATE_GAME_VISIBLE";

import {
    selectGameTypeFilter,
    selectAdmin,
    selectLoggedIn,
    selectCreateGameAgents,
    selectCreateGameVariantOptions
} from '../selectors.js';

import {
    setSignedInAction,
    showSignInDialog
} from './user.js';

import {
    navigateToGame
} from './app.js';

import {
    updateAndShowError
} from './error.js';
import { rememberSurfaceForGame } from '../utils/companion-surface.js';

// Action type definitions
interface UpdateManagersAction {
    type: typeof UPDATE_MANAGERS;
    managers: ManagerInfo[];
}

interface UpdateGamesListAction {
    type: typeof UPDATE_GAMES_LIST;
    participatingActiveGames: GameListItem[];
    participatingFinishedGames: GameListItem[];
    visibleActiveGames: GameListItem[];
    visibleJoinableGames: GameListItem[];
    allGames: GameListItem[];
}

interface UpdateGameTypeFilterAction {
    type: typeof UPDATE_GAME_TYPE_FILTER;
    name: string;
}

interface UpdateSelectedManagerIndexAction {
    type: typeof UPDATE_SELECTED_MANAGER_INDEX;
    index: number;
}

interface UpdateNumPlayersAction {
    type: typeof UPDATE_NUM_PLAYERS;
    numPlayers: number;
}

interface UpdateAgentNameAction {
    type: typeof UPDATE_AGENT_NAME;
    index: number;
    name: string;
}

interface UpdateVariantOptionAction {
    type: typeof UPDATE_VARIANT_OPTION;
    variantIndex: number;
    optionIndex: number;
}

interface UpdateCreateGameOpenAction {
    type: typeof UPDATE_CREATE_GAME_OPEN;
    open: boolean;
}

interface UpdateCreateGameVisibleAction {
    type: typeof UPDATE_CREATE_GAME_VISIBLE;
    visible: boolean;
}

export type ListAction =
    | UpdateManagersAction
    | UpdateGamesListAction
    | UpdateGameTypeFilterAction
    | UpdateSelectedManagerIndexAction
    | UpdateNumPlayersAction
    | UpdateAgentNameAction
    | UpdateVariantOptionAction
    | UpdateCreateGameOpenAction
    | UpdateCreateGameVisibleAction;

// ListThunk can dispatch list, user, app, and error actions since they interact
type ListThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, ListAction | UserAction | AppAction | ErrorAction>;

export const fetchManagers = (): ListThunk<Promise<void>> => async (dispatch) => {
    const response = await apiGet<unknown>(buildApiUrl('list/manager'));
    if (!response.data) {
        dispatch(updateAndShowError('', response.error || 'Manager list response was missing', response.friendlyError || 'Could not load game types'));
        return;
    }
    let managers: ManagerInfo[];
    try {
        managers = decodeManagersResponse(response.data);
    } catch (error) {
        console.error('[manager-list] rejected server payload:', error);
        dispatch(updateAndShowError('', error instanceof Error ? error.message : 'Invalid manager list response', 'The server returned an invalid game-type list'));
        return;
    }

    dispatch({
        type: UPDATE_MANAGERS,
        managers
    })
}

export const updateGameTypeFilter = (name: string): UpdateGameTypeFilterAction => {
    return {
        type: UPDATE_GAME_TYPE_FILTER,
        name,
    }
}

export const fetchGamesList = (): ListThunk<Promise<void>> => async (dispatch, getState) => {

    //TODO: debounce this

    const state = getState();
    const gameType = selectGameTypeFilter(state);
    const isAdmin = selectAdmin(state);

    const response = await apiGet<unknown>(buildApiUrl('list/game', {
        name: gameType,
        admin: isAdmin ? 1 : 0
    }));
    if (!response.data) {
        dispatch(updateAndShowError('', response.error || 'Games list response was missing', response.friendlyError || 'Could not load games'));
        return;
    }
    let data;
    try {
        data = decodeGamesListResponse(response.data);
    } catch (error) {
        console.error('[games-list] rejected server payload:', error);
        dispatch(updateAndShowError('', error instanceof Error ? error.message : 'Invalid games list response', 'The server returned an invalid games list'));
        return;
    }

    dispatch({
        type: UPDATE_GAMES_LIST,
        participatingActiveGames: data.ParticipatingActiveGames,
        participatingFinishedGames: data.ParticipatingFinishedGames,
        visibleActiveGames: data.VisibleActiveGames,
        //TODO: it's weird that we rename this variable from the server here
        visibleJoinableGames: data.VisibleJoinableActiveGames,
        allGames: data.AllGames,
    })

}

export const createGame = (propertyDict: Record<string, string | number | boolean>): ListThunk<Promise<void>> => async (dispatch, getState) => {

    //TODO: we should probably have this signature take something different,
    //like manager, numPlayers, open, visible separately, then a bundle of
    //game-specific variant properties

    const state = getState();
    const loggedIn = selectLoggedIn(state);

    if (!loggedIn) {
        setSignedInAction(() => dispatch(createGame(propertyDict)));
        dispatch(showSignInDialog());
        return;
    }

    const response = await apiPost<unknown>(buildApiUrl('new/game'), propertyDict, 'application/x-www-form-urlencoded');
    if (!response.data) {
        dispatch(updateAndShowError('', response.error || 'Create game response was missing', response.friendlyError || 'Could not create the game'));
        return;
    }
    try {
        const created = decodeCreateGameResponse(response.data);
        if (propertyDict['companionMode'] === '1' || propertyDict['companionMode'] === true) {
            rememberSurfaceForGame(created.GameID, 'table');
        }
        dispatch(navigateToGame(
            created.GameName,
            created.GameID,
            propertyDict['companionMode'] === '1' || propertyDict['companionMode'] === true
                ? 'table'
                : undefined,
        ));
    } catch (error) {
        console.error('[create-game] rejected server payload:', error);
        dispatch(updateAndShowError('', error instanceof Error ? error.message : 'Invalid create game response', 'The server returned an invalid create-game response'));
    }
};

export const updateSelectedMangerIndex = (index: number): UpdateSelectedManagerIndexAction => {
    return{
        type: UPDATE_SELECTED_MANAGER_INDEX,
        index
    }
}

export const updateNumPlayers = (numPlayers: number): UpdateNumPlayersAction => {
    return {
        type: UPDATE_NUM_PLAYERS,
        numPlayers
    }
}

export const updateAgentName = (index: number, name: string): ListThunk => (dispatch, getState) => {
    const agents = selectCreateGameAgents(getState());
    if (index < 0 || index >= agents.length) return;
    if (agents[index] == name) return;
    dispatch({
        type: UPDATE_AGENT_NAME,
        index,
        name
    })
}

export const updateVariantOption = (variantIndex: number, optionIndex: number): ListThunk => (dispatch, getState) => {
    const variantOptions = selectCreateGameVariantOptions(getState());
    if (variantIndex < 0 || variantIndex >= variantOptions.length) return;
    if (variantOptions[variantIndex] == optionIndex) return;
    dispatch({
        type: UPDATE_VARIANT_OPTION,
        variantIndex,
        optionIndex
    })
}

export const updateOpen = (open: boolean): UpdateCreateGameOpenAction => {
    return {
        type: UPDATE_CREATE_GAME_OPEN,
        open
    }
}

export const updateVisible = (visible: boolean): UpdateCreateGameVisibleAction => {
    return {
        type: UPDATE_CREATE_GAME_VISIBLE,
        visible
    }
}
