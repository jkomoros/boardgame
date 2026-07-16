import assert from 'node:assert/strict';
import test from 'node:test';
import {
  getAvailableValues,
  getPlayerComputedValue,
  getReadyToStartError,
} from './gathering-shared.ts';

test('gathering accessors consume the expanded computed-state shape', () => {
  const state = {
    Game: {
      Computed: {
        ReadyToStartError: 'Teams are not balanced',
        AvailableTeams: [
          { Key: 1, Name: 'Sun' },
          { Key: 2, Name: 'Moon', CSSColor: 'rebeccapurple' },
        ],
      },
    },
    Players: [
      { Computed: { TeamValue: 'Sun', RoleValue: 'Scout' } },
    ],
  };

  assert.equal(getReadyToStartError(state), 'Teams are not balanced');
  assert.deepEqual(getAvailableValues(state, 'AvailableTeams'), [
    { Key: 1, Name: 'Sun' },
    { Key: 2, Name: 'Moon', CSSColor: 'rebeccapurple' },
  ]);
  assert.equal(getPlayerComputedValue(state, 0, 'TeamValue'), 'Sun');
  assert.equal(getPlayerComputedValue(state, 0, 'RoleValue'), 'Scout');
  assert.equal(getPlayerComputedValue(state, 1, 'TeamValue'), '');
});

test('gathering accessors do not confuse raw and expanded computed state', () => {
  const rawShape = {
    Game: { Computed: { Global: { ReadyToStartError: 'wrong layer' } } },
    Players: [],
  };
  assert.equal(getReadyToStartError(rawShape), '');
  assert.deepEqual(getAvailableValues(null, 'AvailableColors'), []);
  assert.equal(getPlayerComputedValue(undefined, 0, 'ColorValue'), '');
});

test('gathering accessors reject malformed server shapes loudly', () => {
  assert.throws(
    () => getReadyToStartError({ Game: { Computed: { ReadyToStartError: 3 } } }),
    /ReadyToStartError must be a string/,
  );
  assert.throws(
    () => getAvailableValues({ Game: { Computed: { AvailableRoles: {} } } }, 'AvailableRoles'),
    /AvailableRoles must be an array/,
  );
  assert.throws(
    () => getAvailableValues({ Game: { Computed: { AvailableTeams: [{ Key: 1, Name: '' }] } } }, 'AvailableTeams'),
    /AvailableTeams\[0\] is not a valid enum value/,
  );
  assert.throws(
    () => getPlayerComputedValue({ Players: [] }, -1, 'TeamValue'),
    /player index must be a non-negative safe integer/,
  );
  assert.throws(
    () => getPlayerComputedValue({ Players: [{ Computed: { ColorValue: false } }] }, 0, 'ColorValue'),
    /ColorValue must be a string/,
  );
});
