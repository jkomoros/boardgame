import {
	UPDATE_GAME_ROUTE,
	UPDATE_GAME_STATIC_INFO,
	UPDATE_GAME_CURRENT_STATE,
	ENQUEUE_STATE_BUNDLE,
	DEQUEUE_STATE_BUNDLE,
	CLEAR_STATE_BUNDLES,
	MARK_ANIMATION_STARTED,
	MARK_ANIMATION_COMPLETED
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
	originalWallClockTime: 0,
	// Animation system state
	animation: {
		pendingBundles: [],
		lastFiredBundle: null,
		activeAnimations: []
	}
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
	case ENQUEUE_STATE_BUNDLE:
		return {
			...state,
			animation: {
				...state.animation,
				pendingBundles: [...state.animation.pendingBundles, action.bundle]
			}
		};
	case DEQUEUE_STATE_BUNDLE:
		const [firedBundle, ...remainingBundles] = state.animation.pendingBundles;
		return {
			...state,
			animation: {
				...state.animation,
				pendingBundles: remainingBundles,
				lastFiredBundle: firedBundle || state.animation.lastFiredBundle
			}
		};
	case CLEAR_STATE_BUNDLES:
		return {
			...state,
			animation: {
				...state.animation,
				pendingBundles: []
			}
		};
	case MARK_ANIMATION_STARTED:
		return {
			...state,
			animation: {
				...state.animation,
				activeAnimations: [...state.animation.activeAnimations, action.animationId]
			}
		};
	case MARK_ANIMATION_COMPLETED:
		return {
			...state,
			animation: {
				...state.animation,
				activeAnimations: state.animation.activeAnimations.filter(id => id !== action.animationId)
			}
		};
	default:
		return state;
	}
};

export default app;
