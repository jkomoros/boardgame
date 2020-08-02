import {
    UPDATE_GAME_ROUTE
} from '../actions/game.js';

const INITIAL_STATE = {
    id: '',
    name: '',
};

const app = (state = INITIAL_STATE, action) => {
	switch (action.type) {
	case UPDATE_GAME_ROUTE:
		return {
			...state,
            id: action.id,
            name: action.name
        };
	default:
		return state;
	}
};

export default app;
