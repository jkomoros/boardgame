import assert from 'node:assert/strict';
import test from 'node:test';
import { clientMoveFromWire } from './client-move.ts';

test('client move boundary copies only safe immutable animation metadata', () => {
  const move = clientMoveFromWire({
    Name: 'Choose Secret Card', Version: 17,
    Blob: { Target: 4 }, Proposer: 2, Timestamp: 'secret-ish',
  });
  assert.deepEqual(move, { Name: 'Choose Secret Card', Version: 17 });
  assert.equal(Object.isFrozen(move), true);
  assert.equal(clientMoveFromWire(null), null);
});

test('client move boundary rejects malformed metadata loudly', () => {
  for (const invalid of [undefined, 3, [], {}, { Name: ' ', Version: 1 },
    { Name: 'Move', Version: -1 }, { Name: 'Move', Version: 1.5 },
    { Name: 'Move', Version: Number.MAX_SAFE_INTEGER + 1 }]) {
    assert.throws(() => clientMoveFromWire(invalid));
  }
  assert.throws(() => clientMoveFromWire({ Name: 'x'.repeat(257), Version: 1 }), /exceeds 256/);
});
