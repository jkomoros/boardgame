export const UPDATE_MANAGERS = 'UPDATE_MANAGERS';
export const UPDATE_GAMES_LIST = 'UPDATE_GAMES_LIST';
export const UPDATE_GAME_TYPE_FILTER = 'UPDATE_GAME_TYPE_FILTER';

import {
    apiPath
} from '../util.js';

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

export const fetchGamesList = (gameType, isAdmin) => async (dispatch) => {

    //TODO: debounce this

    //TODO: gameType, isAdmin should be recoverable from state and not need to
    //be passed as arguments.

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
        visibleJoinableGames: data.VisibleJoinableGames,
        allGames: data.AllGames,
    })

}


