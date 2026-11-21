// Refused proposals, which the conformance corpus cannot show (every
// proposal recorded there succeeded).
//
// testdata/rejections.vectors.json is produced by the real Go engine
// (testdata/gen/rejectvec): a wrong proposer, a stale version, illegal
// arguments, input the binder refuses, a move after game over, an unknown
// move, another game's move, a move out of phase or out of its progression,
// and the fix-up cap. For each, the TypeScript engine must report Go's
// error text and, like Go, leave the state exactly as it was. At the fix-up
// cap Go has already committed versions when it fails; the engine must
// return the same number of versions ending in the same state.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { blackjack } from '../games/blackjack.ts';
import { memory } from '../games/memory.ts';
import { pig } from '../games/pig.ts';
import { tictactoe } from '../games/tictactoe.ts';
import type { StateJSON } from '../state.ts';
import { cloneJSON } from '../state.ts';
import {
	type CompiledGame, ERR_TOO_MANY_FIX_UPS, type GameSnapshot, type GameSpec, type Input, MAX_FIX_UP_RECURSION, type MoveRecord,
	addForPhase, compileGame, createGame, noOp, propose,
} from './index.ts';

interface Step {
	name: string;
	proposer: number;
	fields?: Record<string, unknown>;
	input?: Record<string, unknown>;
	useInput?: boolean;
	expectedVersion: number;
	from?: string;
}

interface Case {
	name: string;
	game: string;
	id: string;
	secretSalt: string;
	numPlayers: number;
	variant?: Record<string, string>;
	seating?: boolean;
	setup: Step[];
	attempt?: Step;
	error: string;
	versionBefore: number;
	versionAfter: number;
	finalState?: StateJSON;
	finalMoves?: string[];
}

const vectors = JSON.parse(readFileSync(fileURLToPath(new URL('../testdata/rejections.vectors.json', import.meta.url)), 'utf8')) as { cases: Case[] };

// The generator's wrapped tictactoe: an always-legal fix-up named Spin,
// everywhere or only once the first move has been made.
const spin = noOp({ name: 'Spin', helpText: 'A fix-up that is always legal' });
const specs: Record<string, GameSpec> = {
	memory,
	blackjack,
	tictactoe,
	pig,
	'spin-always': { ...tictactoe, name: 'main', moves: [...tictactoe.moves, spin] },
	'spin-after-first-move': { ...tictactoe, name: 'main', moves: [...tictactoe.moves, ...addForPhase(1, spin)] },
};
const compiled: Record<string, CompiledGame> = {};
const gameFor = (name: string): CompiledGame => (compiled[name] ??= compileGame(specs[name]));

function inputOf(step: Step): Input {
	const expected = step.expectedVersion >= 0 ? { expectedVersion: step.expectedVersion } : {};
	if (step.useInput) return { kind: 'propose', name: step.name, input: step.input ?? {}, proposer: step.proposer, ...expected };
	return { kind: 'record', name: step.name, fields: step.fields ?? {}, proposer: step.proposer, ...expected };
}

const groups = new Map<string, Case[]>();
for (const c of vectors.cases) {
	const group = c.name.includes('fix-up cap') ? 'the fix-up cap' : c.game;
	groups.set(group, [...(groups.get(group) ?? []), c]);
}

test('the rejection vectors cover every kind of refusal', () => {
	const errors = vectors.cases.map(c => c.error);
	for (const fragment of [
		"it's not your turn", 'ObserverPlayerIndex', 'The proposer was not valid.', 'stale game version', 'already occupied', 'out of range',
		'is not an occupied slot', 'requires field', 'does not accept creator field', 'cannot unmarshal', 'must not be null', 'unknown move',
		'engine-owned', 'not configured for this game', 'Move is not legal in phase', 'Move name does not match', 'Game was already finished',
		ERR_TOO_MANY_FIX_UPS,
	]) {
		assert.ok(errors.some(error => error.includes(fragment)), 'no vector whose error mentions ' + JSON.stringify(fragment));
	}
});

for (const [group, cases] of groups) {
	test(`Go rejections: ${group} (${cases.length} cases)`, () => {
		for (const c of cases) {
			const game = gameFor(c.game);
			const created = createGame(game, { id: c.id, secretSalt: c.secretSalt, numPlayers: c.numPlayers, variant: c.variant ?? null, seating: c.seating === true, now: 0 });
			assert.ok(created.record !== null, c.name + ': createGame refused: ' + created.error);

			if (!c.attempt) {
				// The cap is reached while the game is being created.
				assert.equal(created.error, c.error, c.name);
				assert.equal(created.versions.length - 1, c.versionAfter, c.name + ': versions committed before the cap');
				assert.deepEqual(created.versions[created.versions.length - 1].state, c.finalState, c.name + ': final state');
				assert.deepEqual(created.versions.slice(1).map(version => version.move?.Name), c.finalMoves, c.name + ': moves applied');
				continue;
			}
			assert.equal(created.error, undefined, c.name + ': set-up failed');

			const history: MoveRecord[] = [];
			let state = created.versions[created.versions.length - 1].state;
			for (const version of created.versions) if (version.move) history.push(version.move);
			const snapshot = (): GameSnapshot => ({ record: created.record!, history, state });

			c.setup.forEach((step, i) => {
				const result = propose(game, snapshot(), inputOf(step), { now: 0 });
				assert.equal(result.error, undefined, `${c.name}: setup step ${i} (${step.name}) was refused`);
				for (const version of result.versions) history.push(version.move as MoveRecord);
				state = result.versions[result.versions.length - 1].state;
			});
			assert.equal(history.length, c.versionBefore, c.name + ': version before the attempt');

			const before = cloneJSON({ state, history });
			const result = propose(game, snapshot(), inputOf(c.attempt), { now: 0 });
			assert.equal(result.error, c.error, c.name);
			// Whatever happened, the caller's own state and history were not touched.
			assert.deepEqual({ state, history }, before, c.name + ': the engine modified its input');

			if (c.versionAfter === c.versionBefore) {
				assert.equal(result.versions.length, 0, c.name + ': a refused proposal produced versions');
				continue;
			}
			assert.equal(result.versions.length, c.versionAfter - c.versionBefore, c.name + ': versions committed before the cap');
			assert.equal(result.versions.length, MAX_FIX_UP_RECURSION + 2, c.name);
			assert.deepEqual(result.versions[result.versions.length - 1].state, c.finalState, c.name + ': final state');
			assert.deepEqual(result.versions.map(version => version.move?.Name), c.finalMoves, c.name + ': moves applied');
		}
	});
}
