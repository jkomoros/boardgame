import {
	UPDATE_PAGE,
	UPDATE_OFFLINE,
	OPEN_SNACKBAR,
	CLOSE_SNACKBAR,
	OPEN_HEADER_PANEL,
	CLOSE_HEADER_PANEL,
} from '../actions/app.js';


const INITIAL_STATE = {
	location: '',
	page: '',
	pageExtra: '',
	offline: false,
	snackbarOpened: false,
	headerPanelOpen: false,
};

const app = (state = INITIAL_STATE, action) => {
	switch (action.type) {
	case UPDATE_PAGE:
		return {
			...state,
			location: action.location,
			page: action.page,
			pageExtra: action.pageExtra
		};
	case UPDATE_OFFLINE:
		return {
			...state,
			offline: action.offline
		};
	case OPEN_SNACKBAR:
		return {
			...state,
			snackbarOpened: true
		};
	case CLOSE_SNACKBAR:
		return {
			...state,
			snackbarOpened: false
		};
	case OPEN_HEADER_PANEL:
		return {
			...state,
			headerPanelOpen: true
		};
	case CLOSE_HEADER_PANEL: 
		return {
			...state,
			headerPanelOpen: false
		};
	default:
		return state;
	}
};

export default app;
