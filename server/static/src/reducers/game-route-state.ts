import type { GameState } from '../types/store.js';

/** Creates a clean route-owned game session while retaining explicit viewer
 * preferences. No fetched, socket, animation, or error state crosses routes. */
export function gameRouteState(
  name = '',
  id = '',
  preferences?: Pick<GameState['view'], 'requestedPlayer' | 'autoCurrentPlayer'>,
): GameState {
  return {
    id,
    name,
    chest: null,
    playersInfo: [],
    hasEmptySlots: false,
    open: false,
    visible: false,
    isOwner: false,
    companionInfo: null,
    currentState: null,
    timerInfos: null,
    pathsToTick: [],
    originalWallClockTime: 0,
    animation: {
      pendingBundles: [],
      lastFiredBundle: null,
      activeAnimations: [],
    },
    versions: { current: 0, target: -1, lastFetched: 0 },
    socket: { connected: false, connectionAttempts: 0, lastError: null },
    view: {
      game: null,
      viewingAsPlayer: 0,
      requestedPlayer: preferences?.requestedPlayer ?? 0,
      autoCurrentPlayer: preferences?.autoCurrentPlayer ?? false,
      moveForms: null,
    },
    fetchedInfo: null,
    fetchedVersion: null,
    moveSubmitting: false,
    versionFetching: false,
    infoFetching: false,
    infoRequestID: null,
    versionRequestID: null,
    configuring: false,
    error: null,
  };
}
