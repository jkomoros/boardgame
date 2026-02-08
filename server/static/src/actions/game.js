import {
  store
} from '../store.js';

import {
  selectGameCurrentState,
  selectGameChest,
  selectGameName
} from '../selectors.js';

import {
  deepCopy,
  getProperty,
  setPropertyInClone
} from '../util.js';

import {
  buildGameUrl,
  apiPost,
  apiGet
} from '../api.ts';

export const UPDATE_GAME_ROUTE = 'UPDATE_GAME_ROUTE';
export const UPDATE_GAME_STATIC_INFO = "UPDATE_GAME_STATIC_INFO";
export const UPDATE_GAME_CURRENT_STATE = "UPDATE_GAME_CURRENT_STATE";
export const CONFIGURE_GAME_REQUEST = 'CONFIGURE_GAME_REQUEST';
export const CONFIGURE_GAME_SUCCESS = 'CONFIGURE_GAME_SUCCESS';
export const CONFIGURE_GAME_FAILURE = 'CONFIGURE_GAME_FAILURE';
export const JOIN_GAME_REQUEST = 'JOIN_GAME_REQUEST';
export const JOIN_GAME_SUCCESS = 'JOIN_GAME_SUCCESS';
export const JOIN_GAME_FAILURE = 'JOIN_GAME_FAILURE';
export const SUBMIT_MOVE_REQUEST = 'SUBMIT_MOVE_REQUEST';
export const SUBMIT_MOVE_SUCCESS = 'SUBMIT_MOVE_SUCCESS';
export const SUBMIT_MOVE_FAILURE = 'SUBMIT_MOVE_FAILURE';
export const FETCH_GAME_INFO_REQUEST = 'FETCH_GAME_INFO_REQUEST';
export const FETCH_GAME_INFO_SUCCESS = 'FETCH_GAME_INFO_SUCCESS';
export const FETCH_GAME_INFO_FAILURE = 'FETCH_GAME_INFO_FAILURE';
export const FETCH_GAME_VERSION_REQUEST = 'FETCH_GAME_VERSION_REQUEST';
export const FETCH_GAME_VERSION_SUCCESS = 'FETCH_GAME_VERSION_SUCCESS';
export const FETCH_GAME_VERSION_FAILURE = 'FETCH_GAME_VERSION_FAILURE';

export const updateGameRoute = (pageExtra) => {
    const pieces = pageExtra.split("/");
    //remove the trailing slash
    if (!pieces[pieces.length - 1]) pieces.pop();
    if (pieces.length != 2) {
      console.warn("URL for game didn't have expected number of pieces");
      return null;
    }
    return {
        type: UPDATE_GAME_ROUTE,
        name: pieces[0],
        id: pieces[1],
    }
}

export const updateGameStaticInfo = (chest, playersInfo, hasEmptySlots, open, visible, isOwner) => {
  return {
    type: UPDATE_GAME_STATIC_INFO,
    chest,
    playersInfo,
    hasEmptySlots,
    open,
    visible,
    isOwner
  }
}

//currentState should be the unexpanded state (as passed in from server). Timer
//infos should be game.ActiveTimers. originalWallClockTime should be the time
//the state was received from the server (so that we can compute how much time
//has elapsed from what the server reported). This will install the currentState
//in, but also set up callbacks to update timer.TimeLeft for any timers in the
//state automatically.
export const installGameState = (currentState, timerInfos, originalWallClockTime) => (dispatch, getState) => {

  const state = getState();
  const chest = selectGameChest(state);
  const gameName = selectGameName(state);

  let [expandedState, pathsToTick] = expandState(currentState, timerInfos, chest, gameName);

  dispatch(updateGameState(expandedState, pathsToTick, originalWallClockTime));

  if (pathsToTick.length) window.requestAnimationFrame(doTick);
}

const updateGameState = (expandedCurrentState, pathsToTick, originalWallClockTime) => {
  return {
    type: UPDATE_GAME_CURRENT_STATE,
    currentState: expandedCurrentState,
    pathsToTick,
    originalWallClockTime
  }
}

//return [expandedState, pathsToTick]
const expandState = (currentState, timerInfos, chest, gameName) => {
  //Takes the currentState and returns an object where all of the Stacks are replaced by actual references to the component they reference.

  var pathsToTick = [];

  let newState = deepCopy(currentState);

  expandLeafState(newState, newState.Game, ["Game"], pathsToTick, timerInfos, chest, gameName)
  for (var i = 0; i < newState.Players.length; i++) {
    expandLeafState(newState, newState.Players[i], ["Players", i], pathsToTick, timerInfos, chest, gameName)
  }

  return [newState, pathsToTick];

}

const expandLeafState = (wholeState, leafState, pathToLeaf, pathsToTick, timerInfos, chest, gameName) => {
  //Returns an expanded version of leafState. leafState should have keys that are either bools, floats, strings, or Stacks.

  var entries = Object.entries(leafState);
  for (var i = 0; i < entries.length; i++) {
    let item = entries[i];
    let key = item[0];
    let val = item[1];
    //Note: null is typeof "object"
    if (val && typeof val == "object") {
      if (val.Deck) {
        expandStack(val, wholeState, chest, gameName);
      } else if (val.IsTimer) {
        expandTimer(val, pathToLeaf.concat([key]), pathsToTick, timerInfos);
      }   
    }
  }

  //Copy in Player computed state if it exists, for convenience. Do it after expanding properties
  if (pathToLeaf && pathToLeaf.length == 2 && pathToLeaf[0] == "Players") {
    if (wholeState.Computed && wholeState.Computed.Players && wholeState.Computed.Players.length) {
      leafState.Computed = wholeState.Computed.Players[pathToLeaf[1]];
    }
  }
}

const expandStack = (stack, wholeState, chest, gameName) => {
  if (!stack.Deck) {
    //Meh, I guess it's not a stack
    return;
  }

  let components = Array(stack.Indexes.length).fill(null);

  for (var i = 0; i < stack.Indexes.length; i++) {
    let index = stack.Indexes[i];
    if (index == -1) {
      components[i] = null;
      continue;
    }

    //TODO: this should be a constant
    if(index == -2) {
      //TODO: to handle this appropriately we'd need to know how to
      //produce a GenericComponent for each Deck clientside.
      components[i] = {};
    } else {
      components[i] = componentForDeckAndIndex(stack.Deck, index, wholeState, chest);
    }

    if (stack.IDs) {
      components[i].ID = stack.IDs[i];
    }
    components[i].Deck = stack.Deck;
    components[i].GameName = gameName;
  }

  stack.GameName = gameName;
  stack.Components = components;

}

const expandTimer = (timer, pathToLeaf, pathsToTick, timerInfo) => {

  //Always make sure these default to a number so databinding can use them.
  timer.TimeLeft = 0;
  timer.originalTimeLeft = 0;

  if (!timerInfo) return;

  let info = timerInfo[timer.ID];

  if (!info) return;
  timer.TimeLeft = info.TimeLeft;
  timer.originalTimeLeft = timer.TimeLeft;
  pathsToTick.push(pathToLeaf);
}


const componentForDeckAndIndex = (deckName, index, wholeState, chest) => {
  let deck = chest.Decks[deckName];

  if (!deck) return null;

  let result = {...deck[index]};

  if (wholeState && wholeState.Components) {
    if (wholeState.Components[deckName]) {
      result.DynamicValues = wholeState.Components[deckName][index];
    }
  }

  return result

}

const doTick = () => {
  tick();
  const state = store.getState();
  const pathsToTick = state.game ? state.game.pathsToTick : [];
  if (pathsToTick.length > 0) {
    window.requestAnimationFrame(doTick);
  }
}

const tick = () => {

  const state = store.getState();
  const currentState = selectGameCurrentState(state);

  if (!currentState) return;

  const pathsToTick = state.game ? state.game.pathsToTick : [];
  const originalWallClockStartTime = state.game ? state.game.originalWallClockTime : 0;

  if (pathsToTick.length == 0) return;

  let newPaths = [];

  //We'll use util.setPropertyInClone, so the newState will diverge from
  //currentState as we write to it, but can start out the same.
  let newState = currentState;


  for (let i = 0; i < pathsToTick.length; i++) {
    let currentPath = pathsToTick[i];

    let timer = getProperty(newState, currentPath);

    let now = Date.now();
    let difference = now - originalWallClockStartTime;

    let result = Math.max(0, timer.originalTimeLeft - difference);

    newState = setPropertyInClone(newState, currentPath.concat(["TimeLeft"]), result);

    //If we still have time to tick on this, then make sure it's still
    //in the list of things to tick.
    if (timer.TimeLeft > 0) {
      newPaths.push(currentPath);
    }
  }

  if (newPaths.length == pathsToTick.length) {
    //If the length of pathsToTick didn't change, don't change it, so that
    //strict equality matches in the new state will work.
    newPaths = pathsToTick;
  }

  store.dispatch(updateGameState(newState, newPaths, originalWallClockStartTime));
}

/**
 * Configure game properties (open/visible status)
 * @param {Object} gameRoute - Game route with name and id
 * @param {boolean} open - Whether game is open for new players
 * @param {boolean} visible - Whether game is publicly visible
 * @param {boolean} admin - Whether request is from admin
 */
export const configureGame = (gameRoute, open, visible, admin) => async (dispatch) => {
  dispatch({ type: CONFIGURE_GAME_REQUEST });

  const url = buildGameUrl(gameRoute.name, gameRoute.id, 'configure');
  const response = await apiPost(url, {
    open: open ? 1 : 0,
    visible: visible ? 1 : 0,
    admin: admin ? 1 : 0
  }, 'application/x-www-form-urlencoded');

  if (response.error) {
    dispatch({
      type: CONFIGURE_GAME_FAILURE,
      error: response.error,
      friendlyError: response.friendlyError
    });
    return response; // Return response for component to handle
  } else {
    dispatch({ type: CONFIGURE_GAME_SUCCESS });
    return response; // Return success response
  }
};

/**
 * Join a game as a player
 * @param {Object} gameRoute - Game route with name and id
 */
export const joinGame = (gameRoute) => async (dispatch) => {
  dispatch({ type: JOIN_GAME_REQUEST });

  const url = buildGameUrl(gameRoute.name, gameRoute.id, 'join');
  const response = await apiPost(url, {}, 'application/x-www-form-urlencoded');

  if (response.error) {
    dispatch({
      type: JOIN_GAME_FAILURE,
      error: response.error,
      friendlyError: response.friendlyError
    });
    return response; // Return response for component to handle
  } else {
    dispatch({ type: JOIN_GAME_SUCCESS });
    return response; // Return success response
  }
};

/**
 * Submit a move to the game
 * @param {Object} gameRoute - Game route with name and id
 * @param {Object} moveData - Move data including MoveType, fields, admin, player
 */
export const submitMove = (gameRoute, moveData) => async (dispatch) => {
  dispatch({ type: SUBMIT_MOVE_REQUEST });

  const url = buildGameUrl(gameRoute.name, gameRoute.id, 'move');
  const response = await apiPost(url, moveData, 'application/x-www-form-urlencoded');

  if (response.error) {
    dispatch({
      type: SUBMIT_MOVE_FAILURE,
      error: response.error,
      friendlyError: response.friendlyError
    });
    return response; // Return response for component to handle
  } else {
    dispatch({ type: SUBMIT_MOVE_SUCCESS });
    return response; // Return success response
  }
};

/**
 * Expand move form fields with enum values from chest
 * @param {Array} moveForms - Move forms to expand
 * @param {Object} chest - Game chest containing enums
 * @returns {Array} Expanded move forms
 */
const expandMoveForms = (moveForms, chest) => {
  if (!moveForms) return null;

  const expanded = JSON.parse(JSON.stringify(moveForms)); // Deep copy

  for (let i = 0; i < expanded.length; i++) {
    const form = expanded[i];
    // Some forms don't have fields and that's OK.
    if (!form.Fields) continue;
    for (let j = 0; j < form.Fields.length; j++) {
      const field = form.Fields[j];
      if (field.EnumName) {
        field.Enum = chest.Enums[field.EnumName];
      }
    }
  }
  return expanded;
};

/**
 * Fetch initial game info including static info and first state bundle
 * @param {Object} gameRoute - Game route with name and id
 * @param {number} requestedPlayer - Player index to view as
 * @param {boolean} admin - Whether viewing as admin
 * @param {number} lastFetchedVersion - Last fetched version (for from= param)
 */
export const fetchGameInfo = (gameRoute, requestedPlayer, admin, lastFetchedVersion) => async (dispatch) => {
  dispatch({ type: FETCH_GAME_INFO_REQUEST });

  const url = buildGameUrl(
    gameRoute.name,
    gameRoute.id,
    'info',
    {
      player: requestedPlayer,
      admin: admin ? 1 : 0,
      from: lastFetchedVersion
    }
  );

  const response = await apiGet(url);

  if (response.error) {
    dispatch({
      type: FETCH_GAME_INFO_FAILURE,
      error: response.error,
      friendlyError: response.friendlyError
    });
    return response;
  }

  const data = response.data;

  // Expand move forms with enum values
  const expandedForms = expandMoveForms(data.Forms, data.Chest);

  dispatch({
    type: FETCH_GAME_INFO_SUCCESS,
    chest: data.Chest,
    playersInfo: data.Players,
    hasEmptySlots: data.HasEmptySlots,
    open: data.GameOpen,
    visible: data.GameVisible,
    isOwner: data.IsOwner,
    game: data.Game,
    forms: expandedForms,
    viewingAsPlayer: data.ViewingAsPlayer,
    stateVersion: data.StateVersion
  });

  return response;
};

/**
 * Fetch game version bundles for animation playback
 * @param {Object} gameRoute - Game route with name and id
 * @param {number} targetVersion - Version to fetch
 * @param {number} requestedPlayer - Player index to view as
 * @param {boolean} admin - Whether viewing as admin
 * @param {boolean} autoCurrentPlayer - Whether to auto-select current player
 * @param {number} lastFetchedVersion - Last fetched version
 * @param {number} gameVersion - Current game version
 */
export const fetchGameVersion = (
  gameRoute,
  targetVersion,
  requestedPlayer,
  admin,
  autoCurrentPlayer,
  lastFetchedVersion,
  gameVersion
) => async (dispatch, getState) => {
  // Skip if we already have this version
  if (lastFetchedVersion === gameVersion) {
    return { data: null };
  }

  dispatch({ type: FETCH_GAME_VERSION_REQUEST });

  const url = buildGameUrl(
    gameRoute.name,
    gameRoute.id,
    `version/${targetVersion}`,
    {
      player: requestedPlayer,
      admin: admin ? 1 : 0,
      current: autoCurrentPlayer ? 1 : 0,
      from: lastFetchedVersion
    }
  );

  const response = await apiGet(url);

  if (response.error) {
    dispatch({
      type: FETCH_GAME_VERSION_FAILURE,
      error: response.error,
      friendlyError: response.friendlyError
    });
    return response;
  }

  const data = response.data;

  if (data.Error) {
    console.log('Version getter returned error: ' + data.Error);
    return response;
  }

  // Expand move forms in each bundle
  const state = getState();
  const chest = selectGameChest(state);

  const expandedBundles = data.Bundles.map(serverBundle => ({
    ...serverBundle,
    Forms: expandMoveForms(serverBundle.Forms, chest)
  }));

  dispatch({
    type: FETCH_GAME_VERSION_SUCCESS,
    bundles: expandedBundles
  });

  return response;
};