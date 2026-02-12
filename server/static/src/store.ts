import {
	createStore,
	compose,
	applyMiddleware,
	combineReducers,
	Store
} from 'redux';
import { thunk, type ThunkDispatch } from 'redux-thunk';
import { lazyReducerEnhancer } from 'pwa-helpers/lazy-reducer-enhancer.js';
import type { RootState } from './types/store';

import app from './reducers/app.ts';
import error from './reducers/error.ts';
import user from './reducers/user.ts';

// Extend window for Redux DevTools
declare global {
	interface Window {
		__REDUX_DEVTOOLS_EXTENSION_COMPOSE__?: typeof compose;
	}
}

// Sets up a Chrome extension for time travel debugging.
// See https://github.com/zalmoxisus/redux-devtools-extension for more information.
const devCompose = window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ || compose;

// Extend Store type to include addReducers method from lazyReducerEnhancer
interface LazyStore extends Store<RootState, any> {
	addReducers: (reducers: Record<string, any>) => void;
}

// Initializes the Redux store with a lazyReducerEnhancer (so that you can
// lazily add reducers after the store has been created) and redux-thunk (so
// that you can dispatch async actions). See the "Redux and state management"
// section of the wiki for more details:
// https://github.com/Polymer/pwa-starter-kit/wiki/4.-Redux-and-state-management
export const store = createStore(
	(state: any) => state,
	devCompose(
		lazyReducerEnhancer(combineReducers),
		applyMiddleware(thunk))
) as LazyStore;

// Initially loaded reducers.
store.addReducers({
	app,
	error,
	user
});

//Connect it up so the reselect-tools extension will show the selector graph.
//https://github.com/skortchmark9/reselect-tools for how to install the
//extension
import * as selectors from './selectors.ts';
//TODO: why can I not use the basic import? The es build output doesn't have any
//export keywords, maybe it's configured wrong?
import {
	getStateWith,
	registerSelectors,
} from 'reselect-tools/src';
getStateWith(() => store.getState());  // allows you to get selector inputs and outputs
registerSelectors(selectors); // register string names for selectors
