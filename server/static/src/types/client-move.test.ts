import assert from 'node:assert/strict';
import test from 'node:test';
import { clientMoveFromWire } from './client-move.ts';

test('client move boundary copies only safe immutable animation metadata', () => {
  const move = clientMoveFromWire({
    AnimationKey: 'ChooseSecretCard', Version: 17, Properties: { DeckName: 'Roles', Card: 4 },
    Blob: { Target: 4 }, Proposer: 2, Timestamp: 'secret-ish',
  });
  assert.deepEqual(move, {
    AnimationKey: 'ChooseSecretCard', Version: 17, Properties: { DeckName: 'Roles', Card: 4 },
  });
  assert.equal(Object.isFrozen(move), true);
  assert.equal(clientMoveFromWire(null), null);
});

test('client move boundary rejects malformed metadata loudly', () => {
  for (const invalid of [undefined, 3, [], {}, { AnimationKey: ' ', Version: 1 },
    { AnimationKey: 'Move', Version: -1 }, { AnimationKey: 'Move', Version: 1.5 },
    { AnimationKey: 'Move', Version: 1, Properties: new Date() },
    { AnimationKey: 'Move', Version: Number.MAX_SAFE_INTEGER + 1 }]) {
    assert.throws(() => clientMoveFromWire(invalid));
  }
  assert.throws(() => clientMoveFromWire({ AnimationKey: 'x'.repeat(257), Version: 1 }), /exceeds 256/);
});
