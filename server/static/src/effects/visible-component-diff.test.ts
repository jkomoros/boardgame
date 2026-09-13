import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { diffVisibleComponents } from './visible-component-diff.ts';

function visible(ID: string, value = 0) {
  return { ID, Index: 0, Deck: 'cards', GameName: 'demo', Values: { value } };
}

describe('diffVisibleComponents', () => {
  it('reports visible membership changes in stable collection order', () => {
    const result = diffVisibleComponents(
      [visible('a'), visible('b')],
      [visible('b', 99), visible('c')],
    );
    assert.deepEqual(result, {
      status: 'exact', added: ['c'], removed: ['a'], retained: ['b'],
    });
    assert.equal(Object.isFrozen(result), true);
    assert.equal(Object.isFrozen(result.added), true);
  });

  it('treats removed identity as membership only and keeps opaque entries unavailable', () => {
    assert.deepEqual(diffVisibleComponents([visible('gone'), {}], [{}, visible('public')]), {
      status: 'exact', added: ['public'], removed: ['gone'], retained: [],
    });
  });

  it('fails closed rather than guessing through duplicate identity', () => {
    assert.deepEqual(diffVisibleComponents([visible('a'), visible('a')], [visible('a')]), {
      status: 'ambiguous', reason: 'duplicate-before', added: [], removed: [], retained: [],
    });
    assert.deepEqual(diffVisibleComponents([visible('a')], [visible('b'), visible('b')]), {
      status: 'ambiguous', reason: 'duplicate-after', added: [], removed: [], retained: [],
    });
  });
});
