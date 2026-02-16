import type { Reducer } from 'redux';
import type { AppState } from '../types/store';
import type { AppAction } from '../actions/app.ts';
import {
	UPDATE_PAGE,
	UPDATE_OFFLINE,
	OPEN_SNACKBAR,
	CLOSE_SNACKBAR,
	OPEN_HEADER_PANEL,
	CLOSE_HEADER_PANEL,
} from '../actions/app.ts';


const INITIAL_STATE: AppState = {
	location: '',
	page: '',
	pageExtra: '',
	offline: false,
	snackbarOpened: false,
	headerPanelOpen: false,
};

const app: Reducer<AppState, AppAction> = (state = INITIAL_STATE, action): AppState => {
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
