// What the engine step does beyond what the conformance corpus can show.
// The corpus fixes the clock, seats every player at a named index, and
// never refuses a tick; these tests cover timers against a moving clock,
// the host contract for ticks and seats, and that a step is a pure function
// of what it is handed.
//
// Also here: each game's state definition is compared with the definition
// Go derives by reflection from the real game (testdata/defs, produced by
// testdata/gen/defs), and the small pure helpers (instants, SHA-256,
// progression matching) are checked against known answers.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { policyMapFor } from '../definition.ts';
import type { GameDefinition, SubStateDef } from '../definition.ts';
import { blackjack } from '../games/blackjack.ts';
import { memory } from '../games/memory.ts';
import { pig } from '../games/pig.ts';
import { tictactoe } from '../games/tictactoe.ts';
import type { StateJSON } from '../state.ts';
import { cloneJSON } from '../state.ts';
import {
	type CompiledGame, type GameRecord, type Input, type MoveRecord, type StepResult,
	activeTimers, compileGame, countBetween, countExactly, createGame, formatInstant, gameStatus, matchTape, optionalGroup, parallelGroup,
	parseInstant, propose, repeatGroup, serialGroup, viewerOutput,
} from './index.ts';
import { sha256Hex } from './sha256.ts';

const T0 = parseInstant('2026-03-04T05:06:07Z') as number;

/** A game held the way a host holds one: record, history, current state. */
class Table {
	readonly game: CompiledGame;
	readonly record: GameRecord;
	readonly history: MoveRecord[] = [];
	state: StateJSON;

	constructor(game: CompiledGame, options: { numPlayers?: number; variant?: Record<string, string>; seating?: boolean; salt?: string }) {
		this.game = game;
		const created = createGame(game, { id: '0123456789ABCDEF', secretSalt: options.salt ?? 'engine-test-salt', now: T0, ...options });
		assert.equal(created.error, undefined);
		this.record = created.record as GameRecord;
		this.state = created.versions[created.versions.length - 1].state;
		for (const version of created.versions) if (version.move) this.history.push(version.move);
	}

	step(input: Input, now: number): StepResult {
		const result = propose(this.game, { record: this.record, history: this.history, state: this.state }, input, { now });
		for (const version of result.versions) this.history.push(version.move as MoveRecord);
		if (result.versions.length > 0) this.state = result.versions[result.versions.length - 1].state;
		return result;
	}

	get version(): number {
		return this.history.length;
	}
}

/** Two hidden slots of a memory layout whose cards differ (or match). */
function memorySlots(table: Table, wantMatch: boolean): [number, number] {
	const hidden = (table.state.Game.HiddenCards as { Indexes: number[] }).Indexes;
	const type = (index: number): string => memory().decks.cards.components[index]?.Type as string;
	for (let i = 0; i < hidden.length; i++) {
		for (let j = i + 1; j < hidden.length; j++) {
			if (hidden[i] < 0 || hidden[j] < 0) continue;
			if ((type(hidden[i]) === type(hidden[j])) === wantMatch) return [i, j];
		}
	}
	throw new Error('no such pair of slots');
}

const games = { memory: compileGame(memory()), blackjack: compileGame(blackjack()), tictactoe: compileGame(tictactoe()), pig: compileGame(pig()) };

// ----- timers -------------------------------------------------------------

function mismatch(): { table: Table; timerId: string } {
	const table = new Table(games.memory, { variant: { numcards: 'small' } });
	const [a, b] = memorySlots(table, false);
	assert.equal(table.step({ kind: 'propose', name: 'Reveal Card', input: { CardIndex: a }, proposer: 0 }, T0 + 1000).error, undefined);
	// The second reveal is followed by the fix-up that starts the timer.
	const second = table.step({ kind: 'propose', name: 'Reveal Card', input: { CardIndex: b }, proposer: 0 }, T0 + 2000);
	assert.deepEqual(second.versions.map(version => version.move?.Name), ['Reveal Card', 'Start Hide Cards Timer']);
	const timers = activeTimers(table.state);
	assert.equal(timers.length, 1);
	return { table, timerId: timers[0].guard.ID };
}

test('a started timer is a deadline in the state: the save time plus its duration', () => {
	const { table, timerId } = mismatch();
	const [timer] = activeTimers(table.state);
	assert.equal(timer.deadlineMs, T0 + 2000 + 4000);
	assert.deepEqual(timer.move, { Name: 'Hide Cards', Version: 3, Initiator: 0, Timestamp: '0001-01-01T00:00:00Z', Phase: 0, Proposer: -2, Blob: { TargetPlayerIndex: 0 } });
	assert.match(timerId, /^[A-F0-9]{16}$/);
	assert.equal((table.state.Timers as { Deadline: string }[])[0].Deadline, formatInstant(T0 + 6000));
	// The public half of the timer is only its id; viewers get the time left.
	assert.deepEqual(table.state.Game.HideCardsTimer, { ID: timerId, IsTimer: true });
	const at = { record: table.record, history: table.history, state: table.state, settled: true, now: T0 + 3500 };
	assert.deepEqual(viewerOutput(table.game, at, 1).game.ActiveTimers, { [timerId]: { TimeLeft: 2500 } });
	assert.deepEqual(viewerOutput(table.game, { ...at, now: T0 + 99000 }, 1).game.ActiveTimers, { [timerId]: { TimeLeft: 0 } });
});

test('a tick is refused until the deadline has passed, then fires once', () => {
	const { table, timerId } = mismatch();
	const before = table.version;
	assert.deepEqual(table.step({ kind: 'timer', id: timerId }, T0 + 5999), { versions: [], error: 'timer deadline has not passed' });
	assert.equal(table.version, before);

	const fired = table.step({ kind: 'timer', id: timerId }, T0 + 6000);
	assert.equal(fired.error, undefined);
	assert.deepEqual(fired.versions.map(version => [version.move?.Name, version.move?.Proposer, version.move?.Initiator]), [
		['Hide Cards', -2, before + 1],
		['Finish Turn', -2, before + 1],
	]);
	assert.equal((table.state.Timers as { Status: string }[])[0].Status, 'fired');
	assert.deepEqual(activeTimers(table.state), []);

	// The same tick again, or any other, has nothing to complete.
	assert.deepEqual(table.step({ kind: 'timer', id: timerId }, T0 + 9000), { versions: [], error: 'timer completion was stale or already handled' });
	assert.deepEqual(table.step({ kind: 'timer' }, T0 + 9000), { versions: [], error: 'timer completion was stale or already handled' });
});

test('a tick must name the timer that is active', () => {
	const { table } = mismatch();
	assert.deepEqual(table.step({ kind: 'timer', id: 'AAAAAAAAAAAAAAAA' }, T0 + 60000), { versions: [], error: 'timer completion was stale or already handled' });
});

test('a timer the player pre-empted cannot fire', () => {
	const { table, timerId } = mismatch();
	assert.equal(table.step({ kind: 'propose', name: 'Hide Cards', proposer: 0 }, T0 + 3000).error, undefined);
	assert.equal((table.state.Timers as { Status: string }[])[0].Status, 'canceled');
	assert.deepEqual(table.step({ kind: 'timer', id: timerId }, T0 + 60000), { versions: [], error: 'timer completion was stale or already handled' });
});

test('a forced tick ignores the clock (the Go test hook)', () => {
	const { table } = mismatch();
	assert.equal(table.step({ kind: 'timer', force: true }, T0 + 2001).error, undefined);
});

// ----- seating ------------------------------------------------------------

test('a seat input fills the named seat, or the next open one', () => {
	const table = new Table(games.tictactoe, { seating: true });
	const first = table.step({ kind: 'seat', player: -1 }, T0);
	assert.deepEqual(first.versions.map(version => [version.move?.Name, version.move?.Blob.TargetPlayerIndex]), [
		['Seat Player', 0],
		['Activate Inactive Player', 0],
	]);
	const second = table.step({ kind: 'seat', player: -1 }, T0);
	assert.deepEqual(second.versions.map(version => version.move?.Blob.TargetPlayerIndex), [1, 1]);
	assert.deepEqual(table.state.Players.map(player => [player.SeatFilled, player.PlayerInactive]), [[true, false], [true, false]]);
	// No seat is left.
	const third = table.step({ kind: 'seat', player: -1 }, T0);
	assert.equal(third.versions.length, 0);
	assert.ok(third.error);
	// A filled seat cannot be taken again.
	assert.equal(table.step({ kind: 'seat', player: 0 }, T0).versions.length, 0);
});

test('whether seats will be filled is an input to creation, not part of the state', () => {
	// Blackjack with seating off deals at once; with it on, it waits.
	const dealt = new Table(games.blackjack, { numPlayers: 3 });
	assert.equal(dealt.state.Game.Phase, 'Normal Play');
	const waiting = new Table(games.blackjack, { numPlayers: 3, seating: true });
	assert.equal(waiting.version, 0);
	assert.equal(waiting.state.Game.Phase, 'Gathering');
	assert.equal(waiting.record.seating, true);
});

// ----- purity of the step -------------------------------------------------

test('a step is a function of its inputs: same snapshot and input, same versions', () => {
	const table = new Table(games.pig, {});
	const current = table.state.Game.CurrentPlayer as number;
	const snapshot = { record: table.record, history: table.history.slice(), state: cloneJSON(table.state) };
	const frozen = cloneJSON(snapshot);
	const input: Input = { kind: 'propose', name: 'Roll Dice', proposer: current };
	const a = propose(table.game, snapshot, input, { now: T0 });
	const b = propose(table.game, snapshot, input, { now: T0 + 123456 });
	assert.equal(a.error, undefined);
	assert.deepEqual(a, b, 'the clock does not reach a move that starts no timer');
	assert.deepEqual(snapshot, frozen, 'the snapshot was modified');
	// A different salt rolls a different game.
	const other = new Table(games.pig, { salt: 'another-salt' });
	assert.notDeepEqual(other.state, table.state);
});

test('creation refuses what Go refuses', () => {
	const refused = (options: object): string | undefined =>
		createGame(games.memory, { id: 'X', secretSalt: 'Y', now: 0, ...options }).error;
	assert.equal(refused({ numPlayers: 1 }), 'The number of players, 1 was not legal.');
	assert.equal(refused({ numPlayers: 7 }), 'The number of players, 7 was not legal.');
	assert.equal(refused({ numPlayers: -1 }), 'The number of players, -1 is not legal. There must be one or more players.');
	assert.equal(refused({ variant: { color: 'red' } }), "That variation is not legal for this game: configuration had a property called color that isn't expected");
	assert.equal(refused({ variant: { numcards: 'huge' } }), "That variation is not legal for this game: configuration's Number of Cards property had a value that wasn't allowed: huge");
	assert.equal(refused({ agents: ['a'] }), 'If agentNames is not nil, it must have length equivalent to numPlayers.');
	const created = createGame(games.memory, { id: 'X', secretSalt: 'Y', now: 0 });
	assert.deepEqual(created.record?.variant, { cardset: 'all', numcards: 'medium' });
	assert.equal(created.record?.numPlayers, 2);
});

test('a finished game is recognized from its state alone', () => {
	const table = new Table(games.tictactoe, {});
	for (const [player, slot] of [[0, 0], [1, 3], [0, 1], [1, 4]] as const) {
		assert.equal(table.step({ kind: 'propose', name: 'Place Token', input: { Slot: slot }, proposer: player }, T0).error, undefined);
	}
	assert.deepEqual(gameStatus(table.game, table), { finished: false, winners: [] });
	const last = table.step({ kind: 'propose', name: 'Place Token', input: { Slot: 2 }, proposer: 0 }, T0);
	// The winning move ends the chain: no fix-up follows a finished game.
	assert.deepEqual(last.versions.map(version => [version.move?.Name, version.finished, version.winners]), [['Place Token', true, [0]]]);
	assert.deepEqual(gameStatus(table.game, table), { finished: true, winners: [0] });
});

// ----- definitions against Go's reflection --------------------------------

function normalizedSubState(def: SubStateDef, defaultGroup: string): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const name of Object.keys(def.Props)) {
		const prop = def.Props[name];
		out[name] = {
			Type: prop.Type, Sanitize: policyMapFor(prop, defaultGroup), Enum: prop.Enum ?? '', Deck: prop.Deck ?? '', Size: prop.Size ?? 0,
			MaxSize: prop.MaxSize ?? 0, Overlap: prop.Overlap === true, Stacks: prop.Stacks ?? [],
		};
	}
	return out;
}

for (const name of ['memory', 'blackjack', 'tictactoe', 'pig'] as const) {
	test(`${name}: the state definition is the one Go derives from the real game`, () => {
		const go = JSON.parse(readFileSync(fileURLToPath(new URL(`../testdata/defs/${name}.def.json`, import.meta.url)), 'utf8')) as GameDefinition;
		const def = games[name].def;
		assert.equal(def.Name, go.Name);
		assert.deepEqual(normalizedSubState(def.Game, 'all'), normalizedSubState(go.Game, 'all'));
		assert.deepEqual(normalizedSubState(def.Player, 'other'), normalizedSubState(go.Player, 'other'));
		assert.deepEqual(Object.keys(def.DynamicComponentValues).sort(), Object.keys(go.DynamicComponentValues).sort());
		for (const deck of Object.keys(go.DynamicComponentValues)) {
			assert.deepEqual(normalizedSubState(def.DynamicComponentValues[deck], 'all'), normalizedSubState(go.DynamicComponentValues[deck], 'all'));
		}
		assert.deepEqual(JSON.parse(JSON.stringify(def.Decks)), go.Decks);
		// Go's definition lists only the chest's enums.
		for (const enumName of Object.keys(go.Enums)) assert.deepEqual(def.Enums[enumName], go.Enums[enumName]);
		assert.equal(def.GroupEnum ?? '', go.GroupEnum ?? '');
	});
}

// ----- small pure helpers -------------------------------------------------

test('instants format and parse as Go marshals time.Time', () => {
	assert.equal(formatInstant(0), '1970-01-01T00:00:00Z');
	assert.equal(formatInstant(1577836804000), '2020-01-01T00:00:04Z');
	assert.equal(formatInstant(1577836804250), '2020-01-01T00:00:04.25Z');
	assert.equal(formatInstant(951782400000), '2000-02-29T00:00:00Z');
	assert.equal(formatInstant(-1), '1969-12-31T23:59:59.999Z');
	assert.equal(parseInstant('0001-01-01T00:00:00Z'), -62135596800000);
	assert.equal(formatInstant(-62135596800000), '0001-01-01T00:00:00Z');
	assert.equal(parseInstant('2020-01-01T02:00:04+02:00'), 1577836804000);
	assert.equal(parseInstant('2020-01-01T00:00:04.123456789Z'), 1577836804123);
	assert.equal(parseInstant('not a time'), null);
	// Every day of four centuries round-trips.
	for (let day = -200000; day < 200000; day += 97) {
		const ms = day * 86400000 + 12345678;
		assert.equal(parseInstant(formatInstant(ms)), ms);
	}
});

test('SHA-256 agrees with the platform on inputs of every block alignment', () => {
	assert.equal(sha256Hex(''), 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855');
	assert.equal(sha256Hex('abc'), 'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad');
	let text = '';
	for (let length = 0; length < 200; length++) {
		assert.equal(sha256Hex(text), createHash('sha256').update(text, 'utf8').digest('hex'), 'length ' + length);
		text += length % 7 === 0 ? 'é' : String.fromCharCode(97 + (length % 26));
	}
});

test('progression matching follows moves.matchTape', () => {
	// Hand-derived from moves/group_*.go and moves/auto_config.go; the
	// blackjack documents of the conformance corpus exercise the same code
	// against real Go for Serial, Optional and repeatable moves.
	const move = (name: string, allowMultiple = false) => ({ kind: 'move' as const, name, allowMultiple });
	const deal = serialGroup(optionalGroup(move('activate', true)), move('wait'), optionalGroup(move('inactivate', true)), move('deal', true), move('start'));
	assert.equal(matchTape(deal, ['wait']), null);
	assert.equal(matchTape(deal, ['activate', 'activate', 'wait', 'deal', 'deal', 'deal', 'start']), null);
	assert.equal(matchTape(deal, ['deal']), 'Move name does not match: deal is not wait');
	assert.equal(matchTape(deal, ['wait', 'start']), 'Move name does not match: start is not deal');
	assert.equal(matchTape(deal, ['wait', 'deal', 'start', 'start']), 'The progression only matched some of the proposed move history');
	// A prefix of the progression is always acceptable: the tape ran out.
	assert.equal(matchTape(deal, []), null);
	assert.equal(matchTape(deal, ['wait', 'inactivate']), null);

	const twice = repeatGroup(countExactly(2), move('a'));
	assert.equal(matchTape(twice, ['a', 'a']), null);
	assert.equal(matchTape(twice, ['a', 'a', 'a']), 'The progression only matched some of the proposed move history');
	assert.equal(matchTape(repeatGroup(countBetween(1, 2), move('a')), ['b']), 'Move name does not match: b is not a');

	const either = serialGroup(parallelGroup(move('a'), serialGroup(move('b'), move('c'))), move('d'));
	assert.equal(matchTape(either, ['b', 'c', 'a', 'd']), null);
	assert.equal(matchTape(either, ['a', 'b', 'c', 'd']), null);
	assert.equal(matchTape(either, ['a', 'd']), 'no more items match, but tape still left and count not yet reached');
});
