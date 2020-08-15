import {
	UPDATE_GAME_ROUTE,
	UPDATE_GAME_STATIC_INFO
} from '../actions/game.js';

const INITIAL_STATE = {
    id: '',
	name: '',
	chest: null,
	playersInfo: [],
	hasEmptySlots: false,
	open: false,
	visible: false,
	isOwner: false
};

const app = (state = INITIAL_STATE, action) => {
	switch (action.type) {
	case UPDATE_GAME_ROUTE:
		return {
			...state,
            id: action.id,
            name: action.name
		};
	case UPDATE_GAME_STATIC_INFO:
		return {
			...state,
			chest: action.chest,
			playersInfo: action.playersInfo,
			hasEmptySlots: action.hasEmptySlots,
			open: action.open,
			visible: action.visible,
			isOwner: action.isOwner
		}
	default:
		return state;
	}
};

export default app;
