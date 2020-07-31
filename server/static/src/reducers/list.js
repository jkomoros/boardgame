import {
    UPDATE_MANAGERS
} from '../actions/list.js';


const INITIAL_STATE = {
    managers: [],
};

const app = (state = INITIAL_STATE, action) => {
	switch (action.type) {
	case UPDATE_MANAGERS:
		return {
			...state,
            managers: action.managers,
		};
	default:
		return state;
	}
};

export default app;
