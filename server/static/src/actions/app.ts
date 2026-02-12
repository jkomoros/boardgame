import type { ThunkAction } from 'redux-thunk';
import type { RootState } from '../types/store';

export const UPDATE_PAGE = 'UPDATE_PAGE';
export const UPDATE_OFFLINE = 'UPDATE_OFFLINE';
export const OPEN_SNACKBAR = 'OPEN_SNACKBAR';
export const CLOSE_SNACKBAR = 'CLOSE_SNACKBAR';
export const OPEN_HEADER_PANEL = 'OPEN_HEADER_PANEL';
export const CLOSE_HEADER_PANEL = 'CLOSE_HEADER_PANEL';

export const PAGE_DEFAULT = 'list-games';
export const PAGE_GAME = 'game';
export const PAGE_404 = 'view404';

import {
	gamePath
} from '../util.ts';

export const OFFLINE_DEV_MODE = CONFIG ? CONFIG.offline_dev_mode || false : false;

// Action type definitions
interface UpdatePageAction {
	type: typeof UPDATE_PAGE;
	location: string;
	page: string;
	pageExtra: string;
}

interface UpdateOfflineAction {
	type: typeof UPDATE_OFFLINE;
	offline: boolean;
}

interface OpenSnackbarAction {
	type: typeof OPEN_SNACKBAR;
}

interface CloseSnackbarAction {
	type: typeof CLOSE_SNACKBAR;
}

interface OpenHeaderPanelAction {
	type: typeof OPEN_HEADER_PANEL;
}

interface CloseHeaderPanelAction {
	type: typeof CLOSE_HEADER_PANEL;
}

export type AppAction =
	| UpdatePageAction
	| UpdateOfflineAction
	| OpenSnackbarAction
	| CloseSnackbarAction
	| OpenHeaderPanelAction
	| CloseHeaderPanelAction;

type AppThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, AppAction>;

//if silent is true, then just passively updates the URL to reflect what it should be.
export const navigatePathTo = (path: string, silent: boolean): AppThunk => (dispatch, getState) => {
	const state = getState();
	if (silent) {
		window.history.replaceState({}, '', path);
		return;
	}
	window.history.pushState({}, '', path);
	dispatch(navigated(decodeURIComponent(path), decodeURIComponent(location.search)));
};

export const navigateToGame = (gameName: string, gameID: string): AppThunk => (dispatch) => {
	//Do I need dispatch here, or could I just return?
	dispatch(navigatePathTo(gamePath(gameName, gameID), false));
}

export const navigated = (path: string, query: string): AppThunk => (dispatch) => {

	// Extract the page name from path.
	const page = path === '/' ? PAGE_DEFAULT : path.slice(1);

	// Any other info you might want to extract from the path (like page type),
	// you can do here
	dispatch(loadPage(page, query));

};

const loadPage = (pathname: string, query: string): AppThunk => (dispatch) => {

	//pathname is the whole path minus starting '/', like 'c/VIEW_ID'
	let pieces = pathname.split('/');

	let page = pieces[0];
	let pageExtra = pieces.length < 2 ? '' : pieces.slice(1).join('/');

	if (query) pageExtra += query;

	switch(page) {
	case PAGE_DEFAULT:
		import('../components/boardgame-list-games-view.ts').then(() => {
			// Put code in here that you want to run every time when
			// navigating to view1 after my-view1.js is loaded.
		});
		break;
	case PAGE_GAME:
		import('../components/boardgame-game-view.ts');
        break;
    default:
		page = PAGE_404;
		import('../components/boardgame-404-view.ts');
	}

	dispatch(updatePage(pathname, page, pageExtra));
};

const updatePage = (location: string, page: string, pageExtra: string): UpdatePageAction => {
	return {
		type: UPDATE_PAGE,
		location,
		page,
		pageExtra
	};
};

let snackbarTimer: number;

export const showSnackbar = (): AppThunk => (dispatch) => {
	dispatch({
		type: OPEN_SNACKBAR
	});
	window.clearTimeout(snackbarTimer);
	snackbarTimer = window.setTimeout(() =>
		dispatch({ type: CLOSE_SNACKBAR }), 3000);
};

export const updateOffline = (offline: boolean): AppThunk => (dispatch, getState) => {
	// Show the snackbar only if offline status changes.
	if (offline !== getState().app.offline) {
		dispatch(showSnackbar());
	}
	dispatch({
		type: UPDATE_OFFLINE,
		offline
	});
};

export const openHeaderPanel = (): OpenHeaderPanelAction => {
	return {
		type: OPEN_HEADER_PANEL
	};
};

export const closeHeaderPanel = (): CloseHeaderPanelAction => {
	return {
		type: CLOSE_HEADER_PANEL
	};
};

