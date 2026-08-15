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

// Four game renderers -- valentine, murdermrmonroe, pass and sequenceforge --
// each branched their whole effectsForTransition body on `context.move?.Name`.
// There is no `Name`: the move name crosses the boundary as `AnimationKey`,
// and even the server's own wire payload has no such field. So every one of
// those branches compared `undefined` to a move name, was never once true, and
// the games' travel/pulse effects had never played. The type checker says
// "use ['Name']" rather than "no such property", which is exactly the advice
// that would have preserved the bug, so the contract is pinned here too.
test('the move name crosses the boundary as AnimationKey, and only as that', () => {
  const move = clientMoveFromWire({ AnimationKey: 'Draw Card', Version: 3, Name: 'Draw Card' });
  assert.equal(move?.AnimationKey, 'Draw Card');
  assert.equal(Object.prototype.hasOwnProperty.call(move, 'Name'), false);
  assert.deepEqual(Object.keys(move as object).sort(), ['AnimationKey', 'Version']);
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
