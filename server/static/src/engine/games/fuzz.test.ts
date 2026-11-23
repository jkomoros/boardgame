// The ported games against the fuzz corpus (conformance/fuzz/ at the
// repository root): every seeded run the Go fuzz driver checked in for
// memory, blackjack, tictactoe, pig, debuganimations and werewolf is
// replayed on the TypeScript engine. Versions and views are compared as in
// conformance.test.ts; in addition every input Go refused is proposed at the
// version Go refused it and must be refused with Go's exact text, returning
// no versions and leaving the state and history untouched.
//
// With BOARDGAME_FUZZ_DIR naming a directory (the output of
// `boardgame-util fuzz run --seeds A..B --out DIR`), every document under it
// whose game the registry knows is replayed too: the transient acceptance run.
//
// See ../testing/fuzz.ts for what is compared and the divergence report.
import assert from 'node:assert/strict';
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import type { CompiledGame, GameSpec } from '../rules/index.ts';
import { compileGame } from '../rules/index.ts';
import { FIELDS } from '../testing/conformance.ts';
import { type FuzzDocument, type FuzzResult, fuzzPassTable, runFuzzDocument } from '../testing/fuzz.ts';
import { blackjack } from './blackjack.ts';
import { debuganimations } from './debuganimations.ts';
import { memory } from './memory.ts';
import { pig } from './pig.ts';
import { tictactoe } from './tictactoe.ts';
import { werewolf } from './werewolf.ts';
import { checkersGenerated } from './checkers/index.ts';

/** Game name to the factory of the game under test. */
type Registry = (game: string) => (() => GameSpec) | undefined;

const HAND_PORTS: readonly string[] = ['memory', 'blackjack', 'tictactoe', 'pig', 'debuganimations', 'werewolf'];

/** The six hand ports. A generated game is tested by putting its own registry in front of this one. */
const handPorts: Registry = game => {
	switch (game) {
		case 'memory': return memory;
		case 'blackjack': return blackjack;
		case 'tictactoe': return tictactoe;
		case 'pig': return pig;
		case 'debuganimations': return debuganimations;
		case 'werewolf': return werewolf;
		default: return undefined;
	}
};

/** Games with no hand port: the generated game is the one under test. */
const generatedOnly: Registry = game => (game === 'checkers' ? checkersGenerated : undefined);

const registries: readonly Registry[] = [handPorts, generatedOnly];
const specFor = (game: string): (() => GameSpec) | undefined => {
	for (const registry of registries) {
		const found = registry(game);
		if (found !== undefined) return found;
	}
	return undefined;
};

/**
 * Documents known to differ, by name, each with its entry in
 * fabric/transpile/HAND-PORT-BUGS.md. Such a document must still fail, and
 * on the listed field: when it starts passing, the entry has to go.
 */
const EXPECTED_FAILURES: ReadonlyMap<string, string> = new Map();

function fuzzDir(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'fuzz', 'index.json'))) return path.join(dir, 'conformance', 'fuzz');
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find conformance/fuzz/ above ' + import.meta.url);
		dir = parent;
	}
}

/** Every `<game>/seed-NNNN.json.gz` under a directory, in name order. */
function documentsUnder(root: string): { file: string; game: string; name: string }[] {
	const out: { file: string; game: string; name: string }[] = [];
	for (const game of readdirSync(root).sort()) {
		const dir = path.join(root, game);
		if (!statSync(dir).isDirectory()) continue;
		for (const file of readdirSync(dir).sort()) {
			if (!/^seed-\d+\.json\.gz$/.test(file)) continue;
			out.push({ file: path.join(dir, file), game, name: game + '/' + file.replace(/\.json\.gz$/, '') });
		}
	}
	return out;
}

const compiled = new Map<string, CompiledGame>();
const gameFor = (name: string): CompiledGame => {
	let game = compiled.get(name);
	if (game === undefined) {
		const spec = specFor(name);
		assert.ok(spec !== undefined, 'no game registered for ' + name);
		game = compileGame(spec());
		compiled.set(name, game);
	}
	return game;
};

function check(label: string, entry: { file: string; game: string; name: string }, results: FuzzResult[]): void {
	test(label + ': ' + entry.name, () => {
		const doc = JSON.parse(gunzipSync(readFileSync(entry.file)).toString('utf8')) as FuzzDocument;
		const spec = specFor(entry.game);
		const result = runFuzzDocument(gameFor(entry.game), doc, spec === undefined ? {} : { spec });
		result.name = entry.name;
		results.push(result);

		// A document that compared nothing proves nothing.
		const expectedFailure = EXPECTED_FAILURES.get(entry.name);
		if (expectedFailure !== undefined) {
			assert.ok(!result.passed, entry.name + ' is listed as an expected failure (' + expectedFailure + ') but passed; remove it from EXPECTED_FAILURES and HAND-PORT-BUGS.md');
			assert.equal(result.report?.field, expectedFailure, entry.name + ': failed on another field than listed: ' + JSON.stringify(result.report));
			return;
		}
		const problems: string[] = [];
		if (result.document.fatal) problems.push('replay stopped: ' + result.document.fatal);
		for (const field of FIELDS) {
			const f = result.document.fields[field];
			if (f.mismatches > 0) problems.push(field + ': ' + f.mismatches + ' of ' + f.compared + ' differ, e.g.\n    ' + f.examples.join('\n    '));
		}
		if (result.rejections.mismatches > 0) problems.push('rejections: ' + result.rejections.mismatches + ' of ' + result.rejections.compared + ' differ, e.g.\n    ' + result.rejections.examples.join('\n    '));
		if (result.report) problems.push('divergence report: ' + JSON.stringify(result.report, null, 2));
		assert.equal(problems.length, 0, '\n  ' + problems.join('\n  '));
		assert.equal(result.document.fields.state.compared, doc.versions.length, 'every version was compared');
		assert.equal(result.document.fields['view.game'].compared, doc.versions.length * doc.viewers.length, 'every viewer of every version was compared');
		assert.equal(result.rejections.compared, (doc.rejections ?? []).length, 'every rejection was compared');
	});
}

const corpus = fuzzDir();
const index = JSON.parse(readFileSync(path.join(corpus, 'index.json'), 'utf8')) as { files: { file: string; game: string }[] };
const checkedIn = documentsUnder(corpus);
const corpusResults: FuzzResult[] = [];

test('the fuzz corpus holds eight documents for each hand port, as its index lists', () => {
	for (const game of [...HAND_PORTS, 'checkers']) assert.equal(checkedIn.filter(entry => entry.game === game).length, 8, game);
	assert.equal(checkedIn.length, index.files.length, 'documents on disk against index.json');
	for (const entry of checkedIn) assert.ok(specFor(entry.game) !== undefined, 'no game registered for ' + entry.game);
});

for (const entry of checkedIn) check('fuzz', entry, corpusResults);

test('fuzz pass table', t => {
	t.diagnostic('\n' + fuzzPassTable(corpusResults));
	assert.equal(corpusResults.length, checkedIn.length);
	const rejections = corpusResults.reduce((sum, result) => sum + result.rejections.compared, 0);
	t.diagnostic('rejections compared: ' + rejections);
	assert.ok(rejections > 0, 'no rejection was compared');
});

const transient = process.env.BOARDGAME_FUZZ_DIR;
if (transient !== undefined && transient !== '') {
	const entries = documentsUnder(transient).filter(entry => specFor(entry.game) !== undefined);
	const transientResults: FuzzResult[] = [];
	test('BOARDGAME_FUZZ_DIR holds documents', () => {
		assert.ok(entries.length > 0, 'no document for a registered game under ' + transient);
	});
	for (const entry of entries) check('fuzz (BOARDGAME_FUZZ_DIR)', entry, transientResults);
	test('BOARDGAME_FUZZ_DIR pass table', t => {
		t.diagnostic('\n' + fuzzPassTable(transientResults));
		assert.equal(transientResults.length, entries.length);
	});
}

// The report itself, on a recording altered in memory: the runner must point
// at the altered value, and at a rejection whose text was changed.
test('the divergence report names the first difference', () => {
	const file = path.join(corpus, 'tictactoe', 'seed-0001.json.gz');
	const load = (): FuzzDocument => JSON.parse(gunzipSync(readFileSync(file)).toString('utf8')) as FuzzDocument;

	const altered = load();
	const game = (altered.versions[5].state as { Game: { Slots: { Indexes: number[] } } }).Game;
	assert.equal(game.Slots.Indexes[2], 0);
	game.Slots.Indexes[2] = 7;
	const result = runFuzzDocument(gameFor('tictactoe'), altered, { spec: tictactoe });
	assert.equal(result.passed, false);
	const report = result.report;
	assert.ok(report !== undefined);
	assert.deepEqual(
		{ document: report.document, version: report.version, viewer: report.viewer, field: report.field, path: report.path, want: report.want, got: report.got, move: report.move, repro: report.repro },
		{
			document: 'tictactoe/seed-0001', version: 5, viewer: undefined, field: 'state', path: 'Game.Slots.Indexes[2]', want: 7, got: 0,
			move: { name: 'Place Token', proposer: 0, trigger: 'propose' },
			repro: 'boardgame-util fuzz run --game tictactoe --seed 1 --max-proposals 60 --max-versions 5 --out DIR',
		},
	);
	assert.deepEqual(report.lastOutsideInput, { atVersion: 4, name: 'Place Token', input: altered.versions[5].move?.arguments, trigger: 'propose' });
	assert.ok(report.bodiesExecuted !== undefined && report.bodiesExecuted.some(name => name.startsWith('moves[Place Token]')), JSON.stringify(report.bodiesExecuted));

	const refused = load();
	const rejection = refused.rejections?.[0];
	assert.ok(rejection !== undefined);
	rejection.error += '!';
	const second = runFuzzDocument(gameFor('tictactoe'), refused, { spec: tictactoe });
	assert.equal(second.document.passed, true);
	assert.equal(second.rejections.mismatches, 1);
	assert.equal(second.report?.field, 'rejection');
	assert.equal(second.report?.version, rejection.atVersion);
	assert.deepEqual(second.report?.got, { error: "it's not your turn", versions: 0 });
	assert.deepEqual(second.report?.want, { error: "it's not your turn!", versions: 0 });
	assert.equal(second.report?.repro, 'boardgame-util fuzz run --game tictactoe --seed 1 --max-proposals 60 --max-versions ' + (rejection.atVersion + 1) + ' --out DIR');
	assert.ok(second.report?.bodiesExecuted !== undefined && second.report.bodiesExecuted.length > 0, JSON.stringify(second.report?.bodiesExecuted));
});
