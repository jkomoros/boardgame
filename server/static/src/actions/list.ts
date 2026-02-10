import type { ThunkAction } from 'redux-thunk';
import type { RootState, GameListItem } from '../types/store';
import type { UserAction } from './user.js';
import type { AppAction } from './app.js';
import type { ErrorAction } from './error.js';

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
    apiPath,
    postFetchParams
} from '../util.js';

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

// Action type definitions
interface UpdateManagersAction {
    type: typeof UPDATE_MANAGERS;
    managers: any[];
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

    let response = await fetch(apiPath('list/manager'));

    let data = await response.json() as { Managers: any[] };

    let managers = data.Managers;

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

    let response = await fetch(apiPath('list/game', {
        name: gameType,
        admin: isAdmin ? 1 : 0
    }),{
        credentials: 'include',
    });

    let data = await response.json() as {
        ParticipatingActiveGames: GameListItem[];
        ParticipatingFinishedGames: GameListItem[];
        VisibleActiveGames: GameListItem[];
        VisibleJoinableActiveGames: GameListItem[];
        AllGames: GameListItem[];
    };

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

    const body = Object.entries(propertyDict).map((entry) => '' + entry[0] + '=' + entry[1]).join('&');

    let response = await fetch(apiPath('new/game'), postFetchParams(body));

    let responseJSON = await response.json() as {
        Status: string;
        GameName?: string;
        GameID?: string;
        Error?: string;
        FriendlyError?: string;
    };

    if (responseJSON.Status == "Success") {
        dispatch(navigateToGame(responseJSON.GameName!, responseJSON.GameID!));
    } else {
        dispatch(updateAndShowError("", responseJSON.Error || "", responseJSON.FriendlyError || ""));
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

