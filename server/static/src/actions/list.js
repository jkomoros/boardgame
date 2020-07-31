export const UPDATE_MANAGERS = 'UPDATE_MANAGERS';

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


