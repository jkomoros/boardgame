import { Dispatch } from 'redux';
import { store } from '../store.js';
import type { RootState, GameChest, PlayerInfo } from '../types/store';
import type { ApiResponse } from '../api';

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
} from '../api';

interface GameRoute {
  name: string;
  id: string;
}

interface TimerInfo {
  TimeLeft: number;
  [key: string]: any;
}

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
export const ENQUEUE_STATE_BUNDLE = 'ENQUEUE_STATE_BUNDLE';
export const DEQUEUE_STATE_BUNDLE = 'DEQUEUE_STATE_BUNDLE';
export const CLEAR_STATE_BUNDLES = 'CLEAR_STATE_BUNDLES';
export const MARK_ANIMATION_STARTED = 'MARK_ANIMATION_STARTED';
export const MARK_ANIMATION_COMPLETED = 'MARK_ANIMATION_COMPLETED';

export const updateGameRoute = (pageExtra: string) => {
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

export const updateGameStaticInfo = (
  chest: GameChest | null,
  playersInfo: PlayerInfo[],
  hasEmptySlots: boolean,
  open: boolean,
  visible: boolean,
  isOwner: boolean
) => {
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
//has elapsed from what the server reported). This will install the RAW currentState
//in Redux, and selectors will expand it on-the-fly. This also sets up callbacks
//to update timer.TimeLeft for any timers in the state automatically.
export const installGameState = (
  currentState: any,
  timerInfos: Record<string, TimerInfo>,
  originalWallClockTime: number
) => (dispatch: Dispatch, getState: () => RootState) => {

  // Extract paths to tick WITHOUT mutating state
  const pathsToTick = extractTimerPaths(currentState, timerInfos);

  // Augment timer infos with originalTimeLeft (preserve the initial value)
  const augmentedTimerInfos: Record<string, TimerInfo> = {};
  if (timerInfos) {
    Object.keys(timerInfos).forEach(timerID => {
      augmentedTimerInfos[timerID] = {
        ...timerInfos[timerID],
        originalTimeLeft: timerInfos[timerID].TimeLeft
      };
    });
  }

  // Store RAW state directly - expansion happens in selectors!
  dispatch(updateGameState(currentState, augmentedTimerInfos, pathsToTick, originalWallClockTime));

  if (pathsToTick.length) window.requestAnimationFrame(doTick);
}

const updateGameState = (
  rawCurrentState: any,
  timerInfos: Record<string, TimerInfo> | null,
  pathsToTick: (string | number)[][],
  originalWallClockTime: number
) => {
  return {
    type: UPDATE_GAME_CURRENT_STATE,
    currentState: rawCurrentState,  // Store RAW state
    timerInfos,                      // Store timer metadata for selectors
    pathsToTick,
    originalWallClockTime
  }
}

/**
 * PURE function to extract timer paths from state without mutation.
 * Walks the state tree and finds all timers that need ticking.
 * Returns array of paths like [["Game", "Timer"], ["Players", "0", "Timer"]].
 */
const extractTimerPaths = (currentState: any, timerInfos: Record<string, TimerInfo> | null): (string | number)[][] => {
  const pathsToTick: (string | number)[][] = [];

  if (!currentState) return pathsToTick;

  // Extract from Game
  extractTimerPathsFromLeaf(currentState.Game, ["Game"], pathsToTick, timerInfos);

  // Extract from Players
  if (currentState.Players) {
    for (let i = 0; i < currentState.Players.length; i++) {
      extractTimerPathsFromLeaf(currentState.Players[i], ["Players", i], pathsToTick, timerInfos);
    }
  }

  return pathsToTick;
}

/**
 * Helper to extract timer paths from a leaf state object.
 */
const extractTimerPathsFromLeaf = (
  leafState: any,
  pathToLeaf: (string | number)[],
  pathsToTick: (string | number)[][],
  timerInfos: Record<string, TimerInfo> | null
): void => {
  if (!leafState) return;

  Object.entries(leafState).forEach(([key, val]) => {
    if (val && typeof val === 'object' && (val as any).IsTimer) {
      // Found a timer - check if it has time remaining
      const timerID = (val as any).ID;
      if (timerInfos?.[timerID]?.TimeLeft > 0) {
        pathsToTick.push([...pathToLeaf, key]);
      }
    }
  });
}

const doTick = (): void => {
  tick();
  const state = store.getState();
  const pathsToTick = state.game ? state.game.pathsToTick : [];
  if (pathsToTick.length > 0) {
    window.requestAnimationFrame(doTick);
  }
}

const tick = (): void => {

  const state = store.getState();
  const rawState = selectGameCurrentState(state);  // This is now raw state
  const timerInfos = state.game?.timerInfos;

  if (!rawState || !timerInfos) return;

  const pathsToTick = state.game ? state.game.pathsToTick : [];
  const originalWallClockStartTime = state.game ? state.game.originalWallClockTime : 0;

  if (pathsToTick.length == 0) return;

  const now = Date.now();
  const elapsed = now - originalWallClockStartTime;

  // Update timer infos (not the state itself!)
  const newTimerInfos = { ...timerInfos };
  const newPaths: (string | number)[][] = [];

  for (let i = 0; i < pathsToTick.length; i++) {
    const currentPath = pathsToTick[i];

    // Get the timer from raw state
    const timer = getProperty(rawState, currentPath);
    if (!timer?.ID) continue;

    const timerID = timer.ID;
    const originalInfo = timerInfos[timerID];
    if (!originalInfo) continue;

    // Calculate new TimeLeft based on elapsed time since original wall clock time
    // originalInfo.TimeLeft is the time left when the state was first received
    const newTimeLeft = Math.max(0, originalInfo.TimeLeft - elapsed);

    // Update timer info (preserve originalTimeLeft, update TimeLeft)
    newTimerInfos[timerID] = {
      ...originalInfo,
      TimeLeft: newTimeLeft
    };

    // Keep in tick list if still has time
    if (newTimeLeft > 0) {
      newPaths.push(currentPath);
    }
  }

  // Optimize: only update if something changed
  const pathsChanged = newPaths.length !== pathsToTick.length;
  const finalPaths = pathsChanged ? newPaths : pathsToTick;

  // Dispatch with updated timer infos (raw state stays the same!)
  store.dispatch(updateGameState(rawState, newTimerInfos, finalPaths, originalWallClockStartTime));
}

/**
 * Configure game properties (open/visible status)
 */
export const configureGame = (
  gameRoute: GameRoute,
  open: boolean,
  visible: boolean,
  admin: boolean
) => async (dispatch: Dispatch): Promise<ApiResponse<any>> => {
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
 */
export const joinGame = (gameRoute: GameRoute) => async (dispatch: Dispatch): Promise<ApiResponse<any>> => {
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
 */
export const submitMove = (
  gameRoute: GameRoute,
  moveData: Record<string, string>
) => async (dispatch: Dispatch): Promise<ApiResponse<any>> => {
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
 */
const expandMoveForms = (moveForms: any[] | null, chest: GameChest | null): any[] | null => {
  if (!moveForms || !chest) return moveForms;

  const expanded = JSON.parse(JSON.stringify(moveForms)); // Deep copy

  for (let i = 0; i < expanded.length; i++) {
    const form = expanded[i];
    // Some forms don't have fields and that's OK.
    if (!form.Fields) continue;
    for (let j = 0; j < form.Fields.length; j++) {
      const field = form.Fields[j];
      if (field.EnumName && (chest as any).Enums) {
        field.Enum = (chest as any).Enums[field.EnumName];
      }
    }
  }
  return expanded;
};

/**
 * Fetch initial game info including static info and first state bundle
 */
export const fetchGameInfo = (
  gameRoute: GameRoute,
  requestedPlayer: number,
  admin: boolean,
  lastFetchedVersion: number
) => async (dispatch: Dispatch): Promise<ApiResponse<any>> => {
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

  const data = response.data as any;

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
 */
export const fetchGameVersion = (
  gameRoute: GameRoute,
  targetVersion: number,
  requestedPlayer: number,
  admin: boolean,
  autoCurrentPlayer: boolean,
  lastFetchedVersion: number,
  gameVersion: number
) => async (dispatch: Dispatch, getState: () => RootState): Promise<ApiResponse<any>> => {
  // Skip if we already have this version
  if (lastFetchedVersion === gameVersion) {
    return { data: null, status: 200 };
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

  const data = response.data as any;

  if (data.Error) {
    console.log('Version getter returned error: ' + data.Error);
    return response;
  }

  // Expand move forms in each bundle
  const state = getState();
  const chest = selectGameChest(state);

  const expandedBundles = data.Bundles.map((serverBundle: any) => ({
    ...serverBundle,
    Forms: expandMoveForms(serverBundle.Forms, chest)
  }));

  dispatch({
    type: FETCH_GAME_VERSION_SUCCESS,
    bundles: expandedBundles
  });

  return response;
};

/**
 * Animation Actions
 */

/**
 * Enqueue a state bundle for animation playback
 */
export const enqueueStateBundle = (bundle: any) => {
  return {
    type: ENQUEUE_STATE_BUNDLE,
    bundle
  };
};

/**
 * Dequeue the next state bundle (after it's been installed)
 */
export const dequeueStateBundle = () => {
  return {
    type: DEQUEUE_STATE_BUNDLE
  };
};

/**
 * Clear all pending state bundles (on reset)
 */
export const clearStateBundles = () => {
  return {
    type: CLEAR_STATE_BUNDLES
  };
};

/**
 * Mark an animation as started (for tracking)
 */
export const markAnimationStarted = (animationId: string) => {
  return {
    type: MARK_ANIMATION_STARTED,
    animationId
  };
};

/**
 * Mark an animation as completed (for tracking)
 */
export const markAnimationCompleted = (animationId: string) => {
  return {
    type: MARK_ANIMATION_COMPLETED,
    animationId
  };
};