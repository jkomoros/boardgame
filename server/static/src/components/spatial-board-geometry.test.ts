import assert from 'node:assert/strict';
import test from 'node:test';
import type { ExpandedStack } from '../types/boardgame-types.ts';
import { piecesFromSizedStacks } from './spatial-board-geometry.ts';

function stack(ids: readonly string[], components: ExpandedStack<object, object>['Components']): ExpandedStack<object, object> {
  return {
    Deck: 'tokens',
    Indexes: components.map((_, index) => index),
    IDs: ids,
    IDsLastSeen: {},
    ShuffleCount: 0,
    GameName: 'fixture',
    Components: components,
  };
}

test('piecesFromSizedStacks creates explicit stable piece-to-space projections', () => {
  const token = { Index: 2, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'token-2' };
  const source = stack(['', 'token-2', ''], [null, token, null]);
  const pieces = piecesFromSizedStacks([source], ['hall', 'library', 'study'] as const);
  assert.deepEqual(pieces, [{
    id: 'token-2',
    space: 'library',
    stack: source,
    slot: 1,
    component: token,
  }]);
  assert.ok(Object.isFrozen(pieces));
  assert.ok(Object.isFrozen(pieces[0]));
});

test('piecesFromSizedStacks rejects cardinality and stable-ID mismatches loudly', () => {
  assert.throws(
    () => piecesFromSizedStacks([stack([''], [null])], ['hall', 'study']),
    /1 slots but 2 space keys/,
  );
  const token = { Index: 0, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'catalog-id' };
  assert.throws(
    () => piecesFromSizedStacks([stack([''], [token])], ['hall']),
    /occupied slot 0 has no stable ID/,
  );
});
