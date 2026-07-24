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

test('AllGames entries omit Players entirely and must decode to an empty roster, not throw', () => {
  // Regression test: server/api/main.go's doListGames populates AllGames
  // straight from storage.ListGames's CombinedStorageRecord (no per-player
  // enrichment -- only Participating/Visible lists get that), so every real
  // AllGames entry over the wire has no "Players" key at all. Before this
  // fix, list-response.ts's game() decoder required Players unconditionally
  // and threw "AllGames[0].Players must be an array" on every single
  // /list/game response that included any game at all, which surfaced as an
  // "invalid games list" error dialog blocking every offline-dev-mode e2e
  // test that creates a game (confirmed against the live dev server).
  const bareGame = {
    ID: 'ABC123',
    Name: 'pig',
    ReadableLastActivity: 'a moment ago',
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
