
export const UPDATE_PAGE = 'UPDATE_PAGE';
export const UPDATE_OFFLINE = 'UPDATE_OFFLINE';
export const OPEN_SNACKBAR = 'OPEN_SNACKBAR';
export const CLOSE_SNACKBAR = 'CLOSE_SNACKBAR';
export const OPEN_HEADER_PANEL = 'OPEN_HEADER_PANEL';
export const CLOSE_HEADER_PANEL = 'CLOSE_HEADER_PANEL';

export const PAGE_DEFAULT = 'list-games';
export const PAGE_COMMENT = 'game';
export const PAGE_404 = 'view404';

//if silent is true, then just passively updates the URL to reflect what it should be.
export const navigatePathTo = (path, silent) => (dispatch, getState) => {
	const state = getState();
	if (silent) {
		window.history.replaceState({}, '', path);
		return;
	}
	window.history.pushState({}, '', path);
	dispatch(navigated(decodeURIComponent(path), decodeURIComponent(location.search)));
};

export const navigated = (path, query) => (dispatch) => {

	// Extract the page name from path.
	const page = path === '/' ? PAGE_DEFAULT : path.slice(1);

	// Any other info you might want to extract from the path (like page type),
	// you can do here
	dispatch(loadPage(page, query));

};

const loadPage = (pathname, query) => (dispatch) => {

	//pathname is the whole path minus starting '/', like 'c/VIEW_ID'
	let pieces = pathname.split('/');

	let page = pieces[0];
	let pageExtra = pieces.length < 2 ? '' : pieces.slice(1).join('/');

	if (query) pageExtra += query;

	switch(page) {
	case PAGE_DEFAULT:
		import('../boardgame-list-games-view.js').then(() => {
			// Put code in here that you want to run every time when
			// navigating to view1 after my-view1.js is loaded.
		});
		break;
	case PAGE_GAME:
		import('../boardgame-game-view.js');
        break;
    default:
		page = PAGE_404;
		import('../boardgame-404-view.js');
	}

	dispatch(updatePage(pathname, page, pageExtra));
};

const updatePage = (location, page, pageExtra) => {
	return {
		type: UPDATE_PAGE,
		location,
		page,
		pageExtra
	};
};

let snackbarTimer;

export const showSnackbar = () => (dispatch) => {
	dispatch({
		type: OPEN_SNACKBAR
	});
	window.clearTimeout(snackbarTimer);
	snackbarTimer = window.setTimeout(() =>
		dispatch({ type: CLOSE_SNACKBAR }), 3000);
};

export const updateOffline = (offline) => (dispatch, getState) => {
	// Show the snackbar only if offline status changes.
	if (offline !== getState().app.offline) {
		dispatch(showSnackbar());
	}
	dispatch({
		type: UPDATE_OFFLINE,
		offline
	});
};

export const openHeaderPanel = () => {
	return {
		type: OPEN_HEADER_PANEL
	};
};

export const closeHeaderPanel = () => {
	return {
		type: CLOSE_HEADER_PANEL
	};
};

