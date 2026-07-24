import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeGameInfoResponse, decodeGameVersionResponse } from './server-response.ts';

function game(version = 2) {
  return {
    Name: 'pig',
    ID: 'GAME',
    NumPlayers: 2,
    Agents: ['', ''],
    Variant: null,
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
    Chest: { Decks: null, Enums: {}, Constants: null },
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
      IsHost: false,
      CanRematch: false,
      RematchGameID: '',
      TableSession: {
        Status: 'available',
        IsThisTable: false,
        CanTakeOver: false,
        RetryAfterMs: 0,
        DisplacedByTransfer: false,
      },
    },
  };
}

test('game-info decoder validates and normalizes optional collections', () => {
  const decoded = decodeGameInfoResponse(info());
  assert.equal(decoded.Status, 'Success');
  assert.deepEqual(decoded.Chest.Decks, {});
  assert.deepEqual(decoded.Chest.Constants, {});
  assert.equal(decoded.Forms, null);
  assert.deepEqual(decoded.CompanionInfo?.SeatPresentations, []);
  assert.deepEqual(decoded.CompanionInfo?.Absent, []);
  assert.equal(decoded.Game.Version, 2);
  assert.equal(decoded.Players[0].DisplayName, 'Ada');
});

test('game-info decoder preserves optional typed projected choices', () => {
  const absent = decodeGameInfoResponse(info());
  assert.equal(absent.ProjectedMoveChoices, undefined);

  const source = info() as ReturnType<typeof info> & Record<string, unknown>;
  source.ProjectedMoveChoices = {
    StateVersion: 2,
    MoveChoiceProjectionSchemaFingerprint: 'sha256:choices',
    ProjectionSchemaVersion: 1,
    Status: 'ready',
    Sets: [
      {
        MoveName: 'Choose Player', FieldName: 'TargetPlayer', Source: 'players',
        Candidates: [{ Value: 0, Available: true }, { Value: 1, Available: false }],
      },
      {
        MoveName: 'Guess Card', FieldName: 'GuessedCard', Source: 'enum-values',
        Candidates: [{ Value: 'Guard', Available: true }],
      },
      {
        MoveName: 'Choose Card', FieldName: 'TargetCard', Source: 'stack-slots',
        Candidates: [{ Value: 0, Available: true }, { Value: 3, Available: false }],
      },
    ],
  };
  const decoded = decodeGameInfoResponse(source);
  assert.deepEqual(decoded.ProjectedMoveChoices, source.ProjectedMoveChoices);
});

test('projected-choice decoder rejects malformed status, values, and failed payloads', () => {
  const source = info() as ReturnType<typeof info> & Record<string, unknown>;
  const envelope = {
    StateVersion: 2,
    MoveChoiceProjectionSchemaFingerprint: 'sha256:choices',
    ProjectionSchemaVersion: 1,
    Status: 'ready',
    Sets: [{
      MoveName: 'Choose', FieldName: 'Target', Source: 'players',
      Candidates: [{ Value: {}, Available: true }],
    }],
  };
  source.ProjectedMoveChoices = envelope;
  assert.throws(() => decodeGameInfoResponse(source), /Value must be a string or safe integer/);
  source.ProjectedMoveChoices = { ...envelope, Status: 'unknown', Sets: [] };
  assert.throws(() => decodeGameInfoResponse(source), /Status must be "ready" or "failed"/);
  source.ProjectedMoveChoices = {
    ...envelope,
    Status: 'failed',
    Sets: [{
      MoveName: 'Choose', FieldName: 'Target', Source: 'players',
      Candidates: [{ Value: 0, Available: true }],
    }],
  };
  assert.throws(() => decodeGameInfoResponse(source), /Sets must be empty/);
});

test('game-info decoder isolates opaque creator payloads and drops unknown envelope fields', () => {
  const source = info() as ReturnType<typeof info> & Record<string, unknown>;
  source.Chest = { Enums: { Phase: { Values: { Play: 'Play' } } } };
  source.Game.CurrentState.Game = { Nested: { Value: 1 } };
  Object.assign(source.Chest, { InjectedChestField: 'drop me' });
  Object.assign(source.Game.CurrentState, { InjectedStateField: 'drop me' });
  Object.assign(source.Game, { InjectedGameField: 'drop me' });
  Object.assign(source.CompanionInfo, { InjectedCompanionField: 'drop me' });
  source.Game.ActiveTimers = {
    timer: { TimeLeft: 500, InjectedTimerField: 'drop me' },
  } as unknown as typeof source.Game.ActiveTimers;

  const decoded = decodeGameInfoResponse(source);
  const decodedGame = decoded.Game as unknown as Record<string, unknown>;
  const decodedCompanion = decoded.CompanionInfo as unknown as Record<string, unknown>;
  const decodedTimer = decoded.Game.ActiveTimers.timer as unknown as Record<string, unknown>;
  assert.equal((decoded.Chest as unknown as Record<string, unknown>).InjectedChestField, undefined);
  assert.equal(
    (decoded.Game.CurrentState as unknown as Record<string, unknown>).InjectedStateField,
    undefined,
  );
  assert.equal(decodedGame.InjectedGameField, undefined);
  assert.equal(decodedCompanion.InjectedCompanionField, undefined);
  assert.equal(decodedTimer.InjectedTimerField, undefined);

  (source.Game.CurrentState.Game.Nested as { Value: number }).Value = 2;
  ((source.Chest.Enums as Record<string, unknown>).Phase as {
    Values: Record<string, string>;
  }).Values.Play = 'mutated';
  assert.deepEqual(decoded.Game.CurrentState.Game, { Nested: { Value: 1 } });
  assert.equal(decoded.Chest.Enums?.Phase.Values?.Play, 'Play');
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

  assert.throws(
    () => decodeGameInfoResponse({
      ...info(),
      Chest: { Decks: { cards: [{ Index: 1, Values: {} }] } },
    }),
    /Chest\.Decks\.cards\[0\]\.Index must equal its canonical index 0/,
  );

  assert.throws(
    () => decodeGameInfoResponse({
      ...info(),
      Game: {
        ...game(),
        CurrentState: { Game: {}, Players: [{}], Components: { cards: {} } },
      },
    }),
    /CurrentState\.Components\.cards must be an array/,
  );
});

test('deck components without ComponentValues decode with null Values, not throw', () => {
  // Regression test: a deck built with AddComponent(nil) (debuganimations'
  // tokens deck, examples/debuganimations/main.go) has no ComponentValues,
  // and Deck.MarshalJSON (deck.go) serializes that as `Values: null`. The
  // decoder previously rejected it ("Chest.Decks.tokens[0].Values must be an
  // object"), which surfaced as a blocking "Couldn't toggle" dialog the
  // moment admin mode fetched game info for debuganimations.
  const decoded = decodeGameInfoResponse({
    ...info(),
    Chest: { Decks: { tokens: [{ Index: 0, Values: null }] }, Enums: {}, Constants: null },
  });
  assert.deepEqual(decoded.Chest.Decks, { tokens: [{ Index: 0, Values: null }] });
});

test('game-info decoder validates Table session state combinations', () => {
  const active = info();
  active.CompanionInfo.TableSession = {
    Status: 'active', IsThisTable: true, CanTakeOver: false, RetryAfterMs: 12_000, DisplacedByTransfer: false,
  };
  assert.deepEqual(decodeGameInfoResponse(active).CompanionInfo?.TableSession, {
    Status: 'active', IsThisTable: true, CanTakeOver: false, RetryAfterMs: 12_000, DisplacedByTransfer: false,
  });

  const malformedStatus = info();
  malformedStatus.CompanionInfo.TableSession.Status = 'missing' as 'active';
  assert.throws(() => decodeGameInfoResponse(malformedStatus), /Status must be "active" or "available"/);

  const contradictoryActive = info();
  contradictoryActive.CompanionInfo.TableSession = {
    Status: 'active', IsThisTable: false, CanTakeOver: true, RetryAfterMs: 1, DisplacedByTransfer: false,
  };
  assert.throws(() => decodeGameInfoResponse(contradictoryActive), /cannot be true while the Table is active/);

  const contradictoryAvailable = info();
  contradictoryAvailable.CompanionInfo.TableSession = {
    Status: 'available', IsThisTable: true, CanTakeOver: false, RetryAfterMs: 0, DisplacedByTransfer: false,
  };
  assert.throws(() => decodeGameInfoResponse(contradictoryAvailable), /cannot be true when the Table is available/);
});

test('game-version decoder validates bundles and bounds untrusted collections', () => {
  const decoded = decodeGameVersionResponse({
    Status: 'Success',
    Bundles: [{ Game: game(3), Forms: null, ViewingAsPlayer: -1, Move: { AnimationKey: 'Roll', Version: 3 } }],
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
