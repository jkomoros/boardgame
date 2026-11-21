// The path grammar and resolution on hand-written fixtures. The Go vectors
// (vectors.test.ts) cover every grammar error and every non-board path kind
// against real states; this file covers what no example game has - indexed
// board paths - plus boot-time path validation.
import assert from 'node:assert/strict';
import { test } from 'node:test';
import type { LegalContext } from './access.ts';
import {
  legalPathReadsMove,
  parseLegalPath,
  resolveEnumPath,
  resolveIntPath,
  resolveLegalPath,
  resolveStackPath,
  validateLegalPath,
  validateLegalPathTypes,
} from './path.ts';
import * as fx from './testing/fixtures.ts';

const zones = fx.enumOf('zones', { 0: 'North', 1: 'South' });
const zoneNames = { 0: 'North', 1: 'South' };
const colors = fx.enumOf('colors', { 0: 'Red', 1: 'Blue' });
const cards = fx.deck('cards', [{ Suit: 'Hearts' }, { Suit: 'Spades' }, { Suit: 'Clubs' }]);
const north = new fx.FixtureStack('cards', 'growable', [cards[0]]);
const south = new fx.FixtureStack('cards', 'growable', [cards[1], cards[2]]);
const hands = [new fx.FixtureStack('cards', 'growable', []), new fx.FixtureStack('cards', 'growable', [])];

const state = fx.state({
  game: {
    Round: 3,
    Board: new fx.FixtureBoard([north, south], zones),
    Plain: new fx.FixtureBoard([north, south]),
    Flaky: fx.typed('TypeInt', () => {
      throw new Error('storage offline');
    }),
  },
  players: [
    { Score: 7, Areas: new fx.FixtureBoard([hands[0]]) },
    { Score: 9, Areas: new fx.FixtureBoard([hands[1]]) },
  ],
  currentPlayerIndex: 1,
});

function ctx(moveFields: Record<string, unknown> | null, proposer = 0): LegalContext {
  return { state, move: moveFields === null ? null : fx.move(moveFields), proposer };
}

test('parseLegalPath recognises each kind', () => {
  const parse = (p: string) => {
    const r = parseLegalPath(p);
    assert.ok(r.ok, p);
    return r.value;
  };
  assert.deepEqual(parse('game.Round'), { kind: 'game', prop: 'Round', moveField: '', boardIndexField: '' });
  assert.equal(parse('player.Score').kind, 'player');
  assert.equal(parse('proposer.Score').kind, 'proposer');
  assert.equal(parse('players[*].Score').kind, 'playersAll');
  assert.equal(parse('move.Index').kind, 'move');
  assert.deepEqual(parse('players[move.Target].Score'), { kind: 'playersMoveField', prop: 'Score', moveField: 'Target', boardIndexField: '' });
  assert.deepEqual(parse('game.Board[move.Zone]'), { kind: 'game', prop: 'Board', moveField: '', boardIndexField: 'Zone' });
  assert.deepEqual(parse('players[move.Target].Areas[move.Slot]'), {
    kind: 'playersMoveField',
    prop: 'Areas',
    moveField: 'Target',
    boardIndexField: 'Slot',
  });
  // Only the first dot separates kind from property.
  assert.equal(parse('game.A.B').prop, 'A.B');
});

test('a read of move.X or players[move.F].X makes a predicate field dependent; an indexed board alone does not', () => {
  assert.equal(legalPathReadsMove('move.Index'), true);
  assert.equal(legalPathReadsMove('players[move.Target].Score'), true);
  assert.equal(legalPathReadsMove('game.Round'), false);
  assert.equal(legalPathReadsMove('player.Score'), false);
  assert.equal(legalPathReadsMove('not a path'), false);
  // Go classifies by path kind only: game.Board[move.Zone] is a game path.
  assert.equal(legalPathReadsMove('game.Board[move.Zone]'), false);
});

test('an int move field selects a board space by index', () => {
  const stack = resolveStackPath('game.Board[move.Zone]', ctx({ Zone: 1 }));
  assert.deepEqual(stack, { ok: true, value: south });
  assert.deepEqual(resolveStackPath('game.Plain[move.Zone]', ctx({ Zone: 0 })), { ok: true, value: north });
  assert.deepEqual(resolveStackPath('game.Board[move.Zone]', ctx({ Zone: 2 })), {
    ok: false,
    error: 'boardgame: legal path "game.Board[move.Zone]": board selector "Zone" was out of range',
  });
});

test('an enum move field selects a board space by key, and must be the board\'s own enum', () => {
  assert.deepEqual(resolveStackPath('game.Board[move.Zone]', ctx({ Zone: fx.enumValue(zones, 0, zoneNames) })), { ok: true, value: north });
  assert.deepEqual(resolveStackPath('game.Board[move.Zone]', ctx({ Zone: fx.enumValue(colors, 0, { 0: 'Red' }) })), {
    ok: false,
    error: 'boardgame: legal path "game.Board[move.Zone]": enum move field "Zone" does not match board enum',
  });
  // A board with no enum cannot be keyed by one.
  assert.equal(resolveStackPath('game.Plain[move.Zone]', ctx({ Zone: fx.enumValue(zones, 0, zoneNames) })).ok, false);
});

test('a player board is indexed on the selected player', () => {
  assert.deepEqual(resolveStackPath('player.Areas[move.Slot]', ctx({ Slot: 0 })), { ok: true, value: hands[1] });
  assert.deepEqual(resolveStackPath('proposer.Areas[move.Slot]', ctx({ Slot: 0 }, 0)), { ok: true, value: hands[0] });
  assert.deepEqual(resolveStackPath('players[move.Who].Areas[move.Slot]', ctx({ Slot: 0, Who: fx.playerIndex(0) })), {
    ok: true,
    value: hands[0],
  });
});

test('indexed board failures are Go\'s', () => {
  const error = (p: string, c: LegalContext) => {
    const r = resolveLegalPath(p, c);
    assert.equal(r.ok, false, p);
    return r.ok ? '' : r.error;
  };
  assert.equal(error('game.Board[move.Zone]', ctx(null)), 'boardgame: legal path "game.Board[move.Zone]": indexed board requires a move');
  assert.equal(error('game.Board[move.Zone]', ctx({})), 'boardgame: legal path "game.Board[move.Zone]": move field "Zone" does not exist');
  assert.equal(
    error('game.Board[move.Zone]', ctx({ Zone: true })),
    'boardgame: legal path "game.Board[move.Zone]": move field "Zone" has PropertyType TypeBool, expected TypeInt or TypeEnum',
  );
  assert.equal(
    error('game.Round[move.Zone]', ctx({ Zone: 0 })),
    'boardgame: legal path "game.Round[move.Zone]": board property "Round" was unavailable: No such Board prop: Round',
  );
});

test('typed resolvers reject the wrong type and report a failing read', () => {
  assert.deepEqual(resolveIntPath('game.Round', ctx(null)), { ok: true, value: 3 });
  assert.deepEqual(resolveIntPath('player.Score', ctx(null)), { ok: true, value: 9 });
  assert.deepEqual(resolveEnumPath('game.Round', ctx(null)), { ok: false, error: 'legal: path "game.Round" is not an enum property' });
  assert.deepEqual(resolveIntPath('game.Flaky', ctx(null)), {
    ok: false,
    error: 'boardgame: legal path "game.Flaky": error reading property "Flaky": storage offline',
  });
  assert.deepEqual(resolveIntPath('game.Round', { state: null, move: null, proposer: 0 }), {
    ok: false,
    error: 'boardgame: legal path "game.Round": state was nil',
  });
});

test('validateLegalPath checks existence against an example state and move reader', () => {
  const moveReader = fx.move({ Zone: 0, Target: fx.playerIndex(0), Count: 1, Tag: fx.enumValue(colors, 0, { 0: 'Red' }) }).reader();
  assert.equal(validateLegalPath('game.Round', state, null), null);
  assert.equal(validateLegalPath('player.Score', state, null), null);
  assert.equal(validateLegalPath('players[*].Score', state, null), null);
  assert.equal(validateLegalPath('move.Zone', null, moveReader), null);
  assert.equal(validateLegalPath('players[move.Target].Score', state, moveReader), null);
  assert.equal(validateLegalPath('game.Board[move.Zone]', state, moveReader), null);

  assert.equal(validateLegalPath('game.Nope', state, null), 'boardgame: legal path "game.Nope": property "Nope" does not exist');
  assert.equal(validateLegalPath('game.Round', null, null), 'boardgame: legal path "game.Round": no example state provided to validate against');
  assert.equal(
    validateLegalPath('move.Zone', state, null),
    'boardgame: legal path "move.Zone": is a move.* path but no move reader was provided to validate against',
  );
  assert.match(validateLegalPath('players[move.Count].Score', state, moveReader) ?? '', /move field "Count" has PropertyType TypeInt, expected TypePlayerIndex/);
  assert.equal(
    validateLegalPath('game.Round[move.Zone]', state, moveReader),
    'boardgame: legal path "game.Round[move.Zone]": property "Round" has PropertyType TypeInt, expected TypeBoard before [move.Zone]',
  );
  assert.equal(
    validateLegalPath('game.Board[move.Target]', state, moveReader),
    'boardgame: legal path "game.Board[move.Target]": move field "Target" has PropertyType TypePlayerIndex, expected TypeInt or TypeEnum',
  );
  assert.equal(
    validateLegalPath('game.Board[move.Tag]', state, moveReader),
    'boardgame: legal path "game.Board[move.Tag]": enum move field "Tag" does not match board enum',
  );
  assert.equal(
    validateLegalPath('player.Score', fx.state({ game: {} }), null),
    'boardgame: legal path "player.Score": example state has no player states to validate against',
  );
});

test('validateLegalPathTypes treats an indexed board as a stack and skips what it cannot check', () => {
  const moveReader = fx.move({ Zone: 0 }).reader();
  assert.equal(validateLegalPathTypes('game.Board[move.Zone]', ['TypeStack'], state, moveReader), null);
  assert.equal(
    validateLegalPathTypes('game.Round', ['TypeStack'], state, null),
    'boardgame: legal path "game.Round": property "Round" has PropertyType TypeInt, expected TypeStack',
  );
  assert.equal(
    validateLegalPathTypes('game.Round', ['TypeBool', 'TypeEnum'], state, null),
    'boardgame: legal path "game.Round": property "Round" has PropertyType TypeInt, expected one of [TypeBool TypeEnum]',
  );
  // No move reader: a move path's type cannot be checked, so it is not an error.
  assert.equal(validateLegalPathTypes('move.Zone', ['TypeBool'], state, null), null);
  assert.equal(validateLegalPathTypes('game.Round', ['TypeBool'], null, moveReader), null);
});
