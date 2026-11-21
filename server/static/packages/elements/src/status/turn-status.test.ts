import assert from 'node:assert/strict';
import test from 'node:test';
import {
  AdminPlayerIndex,
  AnyPlayerIndex,
  ObserverPlayerIndex,
  turnStatusPresentation,
} from './turn-status.ts';

const context = (currentPlayerIndex: number, viewerPlayerIndex: number, extras = {}) => ({
  currentPlayerIndex,
  viewerPlayerIndex,
  finished: false,
  animating: false,
  ...extras,
});

test('turn status distinguishes players, observers, admins, and simultaneous turns', () => {
  assert.deepEqual(turnStatusPresentation(context(0, 0)), { kind: 'active', message: 'Your turn' });
  assert.deepEqual(turnStatusPresentation(context(1, 0), ['Ada', 'Grace']), {
    kind: 'waiting',
    message: "Grace's turn",
  });
  assert.deepEqual(turnStatusPresentation(context(1, ObserverPlayerIndex)), {
    kind: 'waiting',
    message: "Player 2's turn",
  });
  assert.deepEqual(turnStatusPresentation(context(1, AdminPlayerIndex)), {
    kind: 'waiting',
    message: "Player 2's turn",
  });
  assert.deepEqual(turnStatusPresentation(context(AnyPlayerIndex, 0)), {
    kind: 'active',
    message: 'Your turn',
  });
  assert.deepEqual(turnStatusPresentation(context(AnyPlayerIndex, ObserverPlayerIndex)), {
    kind: 'simultaneous',
    message: 'All players may act',
  });
});

test('turn status gates unstable or finished state and rejects sentinel foot guns', () => {
  assert.equal(turnStatusPresentation(context(0, 0, { animating: true })), null);
  assert.equal(turnStatusPresentation(context(0, 0, { finished: true })), null);
  assert.equal(turnStatusPresentation(context(ObserverPlayerIndex, 0)), null);
  assert.throws(() => turnStatusPresentation(context(0, AnyPlayerIndex)), /viewerPlayerIndex must be a concrete player/);
  assert.throws(
    () => turnStatusPresentation({ ...context(0, 0), extra: true } as never),
    /must contain exactly/,
  );
  assert.throws(() => turnStatusPresentation(context(0, 0), ['']), /only non-empty strings/);
  assert.throws(
    () => turnStatusPresentation(context(0, 0), [], false as never),
    /activeLabel must be a non-empty string/,
  );
});
