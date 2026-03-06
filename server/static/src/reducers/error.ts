import type { Reducer } from 'redux';
import type { ErrorState } from '../types/store';
import type { ErrorAction } from '../actions/error.js';
import {
    SHOW_ERROR,
    UPDATE_ERROR,
    HIDE_ERROR
} from '../actions/error.js';

const INITIAL_STATE: ErrorState = {
    title: '',
    message: '',
    friendlyMessage: '',
    showing: false,
};

const error: Reducer<ErrorState, ErrorAction> = (state = INITIAL_STATE, action): ErrorState => {
	switch (action.type) {
	case SHOW_ERROR:
		return {
			...state,
            showing: true,
        };
    case HIDE_ERROR:
        return {
            ...state,
            showing: false,
        };
    case UPDATE_ERROR:
        return {
            ...state,
            title: action.title,
            message: action.message,
            friendlyMessage: action.friendlyMessage,
        }
	default:
		return state;
	}
};

export default error;
