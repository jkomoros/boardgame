import {
    SHOW_ERROR,
    UPDATE_ERROR,
    HIDE_ERROR
} from '../actions/error.js';

const INITIAL_STATE = {
    title: '',
    message: '',
    friendlyMessage: '',
    showing: false,
};

const app = (state = INITIAL_STATE, action) => {
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

export default app;
