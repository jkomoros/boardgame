import assert from 'node:assert/strict';
import test from 'node:test';
import {
  decodeCreateGameResponse,
  decodeGamesListResponse,
  decodeManagersResponse,
} from './list-response.ts';

test('manager decoder normalizes nil Go slices and enforces player bounds', () => {
  const managers = decodeManagersResponse({
    Status: 'Success',
    Managers: [{
      Name: 'pig',
      DisplayName: 'Pig',
      Description: '',
      DefaultNumPlayers: 2,
      MinNumPlayers: 2,
      MaxNumPlayers: 8,
      Agents: null,
      Variant: null,
      SupportsTableHandMode: false,
    }],
  });
  assert.deepEqual(managers[0].Agents, []);
  assert.deepEqual(managers[0].Variant, []);
  assert.throws(
    () => decodeManagersResponse({
      Status: 'Success',
      Managers: [{ ...managers[0], DefaultNumPlayers: 9 }],
    }),
    /DefaultNumPlayers must be within its player bounds/,
  );
});

test('games decoder uses the actual PascalCase wire shape and normalizes omitted admin games', () => {
  const game = {
    ID: 'ABC123',
    Name: 'pig',
    Players: [{ IsEmpty: false, IsAgent: false, DisplayName: 'Ada', PhotoURL: 'ada.png' }],
    ReadableLastActivity: 'a moment ago',
    Open: true,
    Visible: false,
  };
  const decoded = decodeGamesListResponse({
    Status: 'Success',
    ParticipatingActiveGames: [game],
    ParticipatingFinishedGames: null,
    VisibleActiveGames: [],
    VisibleJoinableActiveGames: [],
  });
  assert.equal(decoded.ParticipatingActiveGames[0].Players[0].PhotoURL, 'ada.png');
  assert.deepEqual(decoded.ParticipatingFinishedGames, []);
  assert.deepEqual(decoded.AllGames, []);
  assert.throws(
    () => decodeGamesListResponse({
      Status: 'Success',
      ParticipatingActiveGames: [{ ...game, Players: [{ ...game.Players[0], IsEmpty: 'no' }] }],
    }),
    /ParticipatingActiveGames\[0\]\.Players\[0\]\.IsEmpty must be a boolean/,
  );
});

test('bare AllGames entries decode without the server-enriched fields, not throw', () => {
  // Regression test: server/api/main.go's doListGames populates AllGames
  // straight from storage.ListGames's CombinedStorageRecord. Only the
  // Participating/Visible lists pass through gameStorageRecordWithUsers,
  // which adds exactly two fields: Players and ReadableLastActivity. A real
  // AllGames entry over the wire therefore has NEITHER key. Before this fix
  // the game() decoder required both unconditionally and threw
  // ("AllGames[0].Players must be an array", then
  // "AllGames[0].ReadableLastActivity must be a string") on every
  // /list/game?admin=1 response containing any game, surfacing as an
  // "invalid games list" error dialog that blocked every offline-dev-mode
  // e2e test that toggles Admin Mode (confirmed against the live server).
  const bareGame = {
    ID: 'ABC123',
    Name: 'pig',
    Open: false,
    Visible: false,
  };
  const decoded = decodeGamesListResponse({
    Status: 'Success',
    ParticipatingActiveGames: [],
    ParticipatingFinishedGames: [],
    VisibleActiveGames: [],
    VisibleJoinableActiveGames: [],
    AllGames: [bareGame],
  });
  assert.deepEqual(decoded.AllGames[0].Players, []);
  assert.equal(decoded.AllGames[0].ReadableLastActivity, '');
});

test('create-game decoder requires complete navigation identity', () => {
  assert.deepEqual(
    decodeCreateGameResponse({ Status: 'Success', GameName: 'pig', GameID: 'ABC123' }),
    { GameName: 'pig', GameID: 'ABC123' },
  );
  assert.throws(
    () => decodeCreateGameResponse({ Status: 'Success', GameName: 'pig' }),
    /GameID must be a non-empty string/,
  );
});
