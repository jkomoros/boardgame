import {
	UPDATE_GAME_ROUTE,
	UPDATE_GAME_STATIC_INFO,
	UPDATE_GAME_CURRENT_STATE,
	ENQUEUE_STATE_BUNDLE,
	DEQUEUE_STATE_BUNDLE,
	CLEAR_STATE_BUNDLES,
	MARK_ANIMATION_STARTED,
	MARK_ANIMATION_COMPLETED,
	SET_CURRENT_VERSION,
	SET_TARGET_VERSION,
	SET_LAST_FETCHED_VERSION,
	SOCKET_CONNECTED,
	SOCKET_DISCONNECTED,
	SOCKET_ERROR,
	UPDATE_VIEW_STATE,
	SET_VIEWING_AS_PLAYER,
	SET_REQUESTED_PLAYER,
	SET_AUTO_CURRENT_PLAYER,
	UPDATE_MOVE_FORMS
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
	},
	// Version tracking state
	versions: {
		current: 0,
		target: -1,
		lastFetched: 0
	},
	// WebSocket connection state
	socket: {
		connected: false,
		connectionAttempts: 0,
		lastError: null
	},
	// View state
	view: {
		game: null,
		viewingAsPlayer: 0,
		requestedPlayer: 0,
		autoCurrentPlayer: false,
		moveForms: null
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
	case SET_CURRENT_VERSION:
		return {
			...state,
			versions: {
				...state.versions,
				current: action.version
			}
		};
	case SET_TARGET_VERSION:
		return {
			...state,
			versions: {
				...state.versions,
				target: action.version
			}
		};
	case SET_LAST_FETCHED_VERSION:
		return {
			...state,
			versions: {
				...state.versions,
				lastFetched: action.version
			}
		};
	case SOCKET_CONNECTED:
		return {
			...state,
			socket: {
				...state.socket,
				connected: true,
				connectionAttempts: 0,
				lastError: null
			}
		};
	case SOCKET_DISCONNECTED:
		return {
			...state,
			socket: {
				...state.socket,
				connected: false,
				connectionAttempts: state.socket.connectionAttempts + 1
			}
		};
	case SOCKET_ERROR:
		return {
			...state,
			socket: {
				...state.socket,
				lastError: action.error
			}
		};
	case UPDATE_VIEW_STATE:
		return {
			...state,
			view: {
				...state.view,
				game: action.game,
				viewingAsPlayer: action.viewingAsPlayer,
				moveForms: action.moveForms
			}
		};
	case SET_VIEWING_AS_PLAYER:
		return {
			...state,
			view: {
				...state.view,
				viewingAsPlayer: action.playerIndex
			}
		};
	case SET_REQUESTED_PLAYER:
		return {
			...state,
			view: {
				...state.view,
				requestedPlayer: action.playerIndex
			}
		};
	case SET_AUTO_CURRENT_PLAYER:
		return {
			...state,
			view: {
				...state.view,
				autoCurrentPlayer: action.autoFollow
			}
		};
	case UPDATE_MOVE_FORMS:
		return {
			...state,
			view: {
				...state.view,
				moveForms: action.moveForms
			}
		};
	default:
		return state;
	}
};

export default app;
