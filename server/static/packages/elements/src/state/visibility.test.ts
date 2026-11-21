import assert from 'node:assert/strict';
import { test } from 'node:test';
import { facetAvailable, viewGameProp, viewPlayerProp } from './visibility.ts';
import type { StateVisibility } from './visibility.ts';

test('known zero/default values and hidden values remain distinct', () => {
  const state = {
    Game: { Score: 0, Secret: 0 },
    Players: [{ Role: 'Villager', Vote: 0 }, { Role: 'Villager', Vote: 0 }],
    Visibility: {
      Game: { Score: ['values'], Secret: [] },
      Players: [{ Role: ['values'], Vote: ['values'] }, { Role: [], Vote: [] }],
    } satisfies StateVisibility,
  };
  assert.deepEqual(viewGameProp(state, 'Score'), { known: true, value: 0 });
  assert.deepEqual(viewGameProp(state, 'Secret'), { known: false });
  assert.deepEqual(viewPlayerProp(state, 0, 'Role'), { known: true, value: 'Villager' });
  assert.deepEqual(viewPlayerProp(state, 1, 'Role'), { known: false });
  assert.deepEqual(viewPlayerProp(state, 1, 'Vote'), { known: false });
  assert.deepEqual(viewPlayerProp(state, 0.5, 'Role'), { known: false });
  assert.deepEqual(viewPlayerProp(state, 10, 'Role'), { known: false });
});

test('partial stack disclosure does not establish value knowledge', () => {
  assert.equal(facetAvailable(['count', 'nonempty'], 'count'), true);
  assert.equal(facetAvailable(['count', 'nonempty'], 'values'), false);
  assert.equal(facetAvailable(undefined, 'values'), false);
  assert.deepEqual(viewGameProp({ Game: { Value: 1 } }, 'Value'), { known: false });
});

test('a new snapshot recomputes knowledge without caching old values', () => {
  const visible = { Game: { Value: 'secret' }, Visibility: {
    Game: { Value: ['values'] }, Players: [],
  } satisfies StateVisibility };
  assert.deepEqual(viewGameProp(visible, 'Value'), { known: true, value: 'secret' });
  const hidden = { ...visible, Game: { Value: '' }, Visibility: {
    Game: { Value: [] }, Players: [],
  } satisfies StateVisibility };
  assert.deepEqual(viewGameProp(hidden, 'Value'), { known: false });
});
