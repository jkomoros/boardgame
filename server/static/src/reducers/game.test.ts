import assert from 'node:assert/strict';
import test from 'node:test';
import { gameRouteState } from './game-route-state.ts';

test('route state starts clean and preserves only explicit viewer preferences', () => {
  const state = gameRouteState('beta', 'B', {
    requestedPlayer: 2,
    autoCurrentPlayer: true,
  });

  assert.equal(state.name, 'beta');
  assert.equal(state.id, 'B');
  assert.equal(state.chest, null);
  assert.deepEqual(state.playersInfo, []);
  assert.equal(state.currentState, null);
  assert.equal(state.infoFetching, false);
  assert.equal(state.versionFetching, false);
  assert.equal(state.infoRequestID, null);
  assert.equal(state.versionRequestID, null);
  assert.equal(state.error, null);
  assert.deepEqual(state.socket, { connected: false, connectionAttempts: 0, lastError: null });
  assert.deepEqual(state.versions, { current: 0, target: -1, lastFetched: 0 });
  assert.deepEqual(state.animation.activeAnimations, []);
  assert.equal(state.view.viewingAsPlayer, 0);
  assert.equal(state.view.requestedPlayer, 2);
  assert.equal(state.view.autoCurrentPlayer, true);
});
