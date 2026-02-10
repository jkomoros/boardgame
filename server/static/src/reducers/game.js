import {
	UPDATE_GAME_ROUTE,
	UPDATE_GAME_STATIC_INFO,
	UPDATE_GAME_CURRENT_STATE
} from '../actions/game.js';

const INITIAL_STATE = {
    id: '',
	name: '',
	chest: null,
	playersInfo: [],
	hasEmptySlots: false,
	open: false,
	visible: false,
	isOwner: false,
	// currentState is now RAW state from server (unexpanded)
	// Use selectExpandedGameState selector to get expanded version
	currentState: null,
	// Timer metadata for selector expansion
	timerInfos: null,
	//note that pathsToTick and originalWallClockTime are accessed directly
	//(without selectors) in actions/game.js
	pathsToTick: [],
	originalWallClockTime: 0
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
		};
	case UPDATE_GAME_CURRENT_STATE:
		return {
			...state,
			currentState: action.currentState,
			timerInfos: action.timerInfos,
			pathsToTick: action.pathsToTick,
			originalWallClockTime: action.originalWallClockTime
		};
	default:
		return state;
	}
};

export default app;
