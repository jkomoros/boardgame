import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeGameInfoResponse, decodeGameVersionResponse } from './server-response.ts';

function game(version = 2) {
  return {
    CurrentState: { Version: version, Game: {}, Players: [{}] },
    ActiveTimers: {},
    Version: version,
    CurrentPlayerIndex: 0,
    Finished: false,
    Winners: [],
  };
}

function info() {
  return {
    Status: 'Success',
    Chest: {},
    Players: [{ IsEmpty: false, IsAgent: false, DisplayName: 'Ada' }],
    HasEmptySlots: true,
    GameOpen: true,
    GameVisible: false,
    IsOwner: true,
    Game: game(),
    ViewingAsPlayer: 0,
    StateVersion: 2,
    LegalCatalogVersion: 1,
    MoveInputSchemaFingerprint: 'sha256:test',
    CompanionInfo: {
      CompanionMode: false,
      RoomCode: '',
      RoomLocked: false,
      SeatPresentations: null,
      Absent: null,
    },
  };
}

test('game-info decoder validates and normalizes optional collections', () => {
  const decoded = decodeGameInfoResponse(info());
  assert.equal(decoded.Status, 'Success');
  assert.equal(decoded.Forms, null);
  assert.deepEqual(decoded.CompanionInfo?.SeatPresentations, []);
  assert.deepEqual(decoded.CompanionInfo?.Absent, []);
  assert.equal(decoded.Game.Version, 2);
  assert.equal(decoded.Players[0].DisplayName, 'Ada');
});

test('game-info decoder normalizes Go nil move slices', () => {
  const decoded = decodeGameInfoResponse({
    ...info(),
    Forms: [{ Name: 'Done', HelpText: '', Fields: null, Preconditions: null }],
  });
  assert.deepEqual(decoded.Forms, [{ Name: 'Done', HelpText: '' }]);
});

test('game-info decoder names malformed nested server fields', () => {
  const malformedState = info();
  malformedState.Game.CurrentState.Players = {} as unknown as object[];
  assert.throws(
    () => decodeGameInfoResponse(malformedState),
    /Game info response\.Game\.CurrentState\.Players must be an array/,
  );

  const malformedForm = {
    ...info(),
    Forms: [{
      Name: 'Move', HelpText: '',
      Preconditions: [{
        name: 'test', verdict: 'fail', evaluable: true,
        message: { template: 'bad', bindings: { secret: {} } },
      }],
    }],
  };
  assert.throws(
    () => decodeGameInfoResponse(malformedForm),
    /Preconditions\[0\]\.message\.bindings\.secret must be a string, number, or boolean/,
  );
});

test('game-version decoder validates bundles and bounds untrusted collections', () => {
  const decoded = decodeGameVersionResponse({
    Status: 'Success',
    Bundles: [{ Game: game(3), Forms: null, ViewingAsPlayer: -1, Move: { Name: 'Roll', Version: 3 } }],
  });
  assert.equal(decoded.Bundles[0].Game.Version, 3);
  assert.equal(decoded.Bundles[0].Forms, null);

  assert.throws(
    () => decodeGameVersionResponse({ Status: 'Success', Bundles: new Array(10_001).fill(null) }),
    /Bundles exceeds the maximum of 10000 entries/,
  );
  assert.throws(
    () => decodeGameVersionResponse({
      Status: 'Success',
      Bundles: [{ Game: { ...game(), Version: 1.5 }, Forms: [], ViewingAsPlayer: 0, Move: null }],
    }),
    /Bundles\[0\]\.Game\.Version must be a non-negative safe integer/,
  );
});
