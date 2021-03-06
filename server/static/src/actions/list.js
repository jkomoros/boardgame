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

export const fetchManagers = () => async (dispatch) => {

    let response = await fetch(apiPath('list/manager'));

    let data = await response.json();

    let managers = data.Managers;

    dispatch({
        type: UPDATE_MANAGERS,
        managers
    })
}

export const updateGameTypeFilter = (name) => {
    return {
        type: UPDATE_GAME_TYPE_FILTER,
        name,
    }
}

export const fetchGamesList = () => async (dispatch, getState) => {

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

    let data = await response.json();

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

export const createGame = (propertyDict) => async (dispatch, getState) => {

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

    let responseJSON = await response.json();

    if (responseJSON.Status == "Success") {
        dispatch(navigateToGame(responseJSON.GameName, responseJSON.GameID));
    } else {
        dispatch(updateAndShowError("", responseJSON.Error, responseJSON.FriendlyError));
    }
};

export const updateSelectedMangerIndex = (index) => {
    return{
        type: UPDATE_SELECTED_MANAGER_INDEX,
        index
    }
}

export const updateNumPlayers = (numPlayers) => {
    return {
        type: UPDATE_NUM_PLAYERS,
        numPlayers
    }
}

export const updateAgentName = (index, name) => (dispatch, getState) => {
    const agents = selectCreateGameAgents(getState());
    if (index < 0 || index >= agents.length) return;
    if (agents[index] == name) return;
    dispatch({
        type: UPDATE_AGENT_NAME,
        index,
        name
    })
}

export const updateVariantOption = (variantIndex, optionIndex) => (dispatch, getState) => {
    const variantOptions = selectCreateGameVariantOptions(getState());
    if (variantIndex < 0 || variantIndex >= variantOptions.length) return;
    if (variantOptions[variantIndex] == optionIndex) return;
    dispatch({
        type: UPDATE_VARIANT_OPTION,
        variantIndex,
        optionIndex
    })
}

export const updateOpen = (open) => {
    return {
        type: UPDATE_CREATE_GAME_OPEN,
        open
    }
}

export const updateVisible = (visible) => {
    return {
        type: UPDATE_CREATE_GAME_VISIBLE,
        visible
    }
}

