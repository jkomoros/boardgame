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

export const UPDATE_GAME_ROUTE = 'UPDATE_GAME_ROUTE';
export const UPDATE_GAME_STATIC_INFO = "UPDATE_GAME_STATIC_INFO";
export const UPDATE_GAME_CURRENT_STATE = "UPDATE_GAME_CURRENT_STATE";

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