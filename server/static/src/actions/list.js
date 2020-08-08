export const UPDATE_MANAGERS = 'UPDATE_MANAGERS';
export const UPDATE_GAMES_LIST = 'UPDATE_GAMES_LIST';
export const UPDATE_GAME_TYPE_FILTER = 'UPDATE_GAME_TYPE_FILTER';

import {
    apiPath
} from '../util.js';

import {
    selectGameTypeFilter,
    selectAdmin,
    selectLoggedIn
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

    let response = await fetch(apiPath('new/game'), {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        mode: 'cors',
        body: body,
    });

    let responseJSON = await response.json();

    if (responseJSON.Status == "Success") {
        dispatch(navigateToGame(responseJSON.GameName, responseJSON.GameID));
    } else {
        dispatch(updateAndShowError("", responseJSON.Error, responseJSON.FriendlyError));
    }
};


