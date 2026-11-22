// The test every generated game (games/<game>/, written by `boardgame-util
// transpile ./examples/<game>` from Go's own source and exported
// definition) runs: everything Go recorded for the game, the conformance
// corpus (conformance/<game>/) and the fuzz documents
// (conformance/fuzz/<game>/), every field of every version and viewer
// through the shared runner, and each fuzz document's `rejections`: at its
// `atVersion` the refused outside input is proposed again and the engine
// must refuse it with Go's error text, producing no version and changing
// nothing.
//
// The rejection replay is local here (testing/fuzz.ts did not exist on this
// base); the coordinator will switch it to the shared helper. This file
// declares no test of its own; games/<game>/<game>.gen.test.ts calls
// generatedGameSuite.
import assert from 'node:assert/strict';
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import type { CompiledGame, GameSpec, Input, MoveRecord } from '../rules/index.ts';
import { compileGame, createGame, parseInstant, propose } from '../rules/index.ts';
import type { StateJSON } from '../state.ts';
import { cloneJSON } from '../state.ts';
import { type ConformanceDocument, type DocumentResult, FIELDS, passTable, runDocument } from '../testing/conformance.ts';

interface Rejection {
	atVersion: number;
	kind: 'propose' | 'timer' | 'seat';
	name: string;
	input: Record<string, unknown>;
	proposer: number;
	expectedVersion: number;
	error: string;
}

type Document = ConformanceDocument & { rejections?: Rejection[] };

function corpusDir(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'index.json'))) return path.join(dir, 'conformance');
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find conformance/ above ' + import.meta.url);
		dir = parent;
	}
}

/** The input a recorded version answers to, as testing/conformance.ts builds it. */
function inputFor(game: CompiledGame, doc: Document, history: readonly MoveRecord[]): Input {
	const move = doc.versions[history.length + 1].move;
	assert.ok(move !== null, 'version ' + (history.length + 1) + ' records no move');
	if (move.trigger === 'propose') {
		const compiled = game.movesByName[move.name];
		const supplied: Record<string, unknown> = {};
		if (compiled) {
			for (const field of compiled.inputFields) {
				if (field.disposition !== 'required' && field.disposition !== 'server-defaulted') continue;
				if (Object.prototype.hasOwnProperty.call(move.arguments, field.name)) supplied[field.name] = move.arguments[field.name];
			}
		}
		return { kind: 'propose', name: move.name, input: supplied, proposer: move.proposer, expectedVersion: history.length };
	}
	if (move.trigger === 'seat') return { kind: 'seat', player: move.arguments.TargetPlayerIndex as number };
	assert.equal(move.trigger, 'timer', 'version ' + (history.length + 1) + ' answers to ' + move.trigger);
	return { kind: 'timer', force: true };
}

function rejectionInput(r: Rejection): Input {
	switch (r.kind) {
		case 'propose':
			return { kind: 'propose', name: r.name, input: r.input, proposer: r.proposer, expectedVersion: r.expectedVersion };
		case 'seat':
			return { kind: 'seat', player: typeof r.input.TargetPlayerIndex === 'number' ? r.input.TargetPlayerIndex : -1 };
		case 'timer':
			return { kind: 'timer' };
	}
}

/** Replays the document's inputs and, at each rejection's version, proposes the refused input again. */
function checkRejections(game: CompiledGame, file: string, doc: Document): number {
	const rejections = doc.rejections ?? [];
	const now = parseInstant(doc.engine.clock);
	assert.ok(now !== null);
	const created = createGame(game, {
		id: doc.game.id, secretSalt: doc.game.secretSalt, numPlayers: doc.game.numPlayers, variant: doc.game.variant,
		agents: doc.game.agents, seating: doc.game.seating, now,
	});
	assert.ok(created.record !== null && created.error === undefined, file + ': set-up failed');
	const record = created.record;
	const history: MoveRecord[] = [];
	let state = created.versions[created.versions.length - 1].state;
	for (const version of created.versions) if (version.move) history.push(version.move);
	let compared = 0;
	const at = (version: number): void => {
		for (const r of rejections.filter(r => r.atVersion === version)) {
			const before = cloneJSON({ state, history });
			const result = propose(game, { record, history, state }, rejectionInput(r), { now });
			assert.equal(result.error, r.error, file + ': rejection at version ' + version + ' (' + r.kind + ' ' + r.name + ' by ' + r.proposer + ')');
			assert.equal(result.versions.length, 0, file + ': a refused input produced versions');
			assert.deepEqual({ state, history }, before, file + ': a refused input changed the state');
			compared++;
		}
	};
	at(history.length);
	while (history.length + 1 < doc.versions.length) {
		const step = propose(game, { record, history, state: state as StateJSON }, inputFor(game, doc, history), { now });
		assert.equal(step.error, undefined, file + ': version ' + (history.length + 1) + ' was refused: ' + step.error);
		for (const version of step.versions) {
			if (version.move) history.push(version.move);
			state = version.state;
		}
		at(history.length);
	}
	assert.equal(compared, rejections.length, file + ': every rejection was at a version the replay reached');
	return compared;
}

/**
 * Registers the tests of one generated game: it compiles, then each corpus
 * and fuzz document passes, then a pass table with the totals compared.
 */
export function generatedGameSuite(name: string, make: () => GameSpec, expected: { corpus: number; fuzz: number }): void {
	const dir = corpusDir();
	const read = (file: string): Document => JSON.parse(gunzipSync(readFileSync(path.join(dir, file))).toString('utf8')) as Document;
	const index = JSON.parse(readFileSync(path.join(dir, 'index.json'), 'utf8')) as { file: string; game: string }[];
	const corpus = index.filter(entry => entry.game === name).map(entry => entry.file);
	const fuzzDir = path.join(dir, 'fuzz', name);
	const fuzz = existsSync(fuzzDir)
		? readdirSync(fuzzDir)
				.filter(file => file.endsWith('.json.gz'))
				.sort()
				.map(file => path.join('fuzz', name, file))
		: [];

	let game: CompiledGame;
	const results: DocumentResult[] = [];
	const totals = { documents: 0, versions: 0, viewerVersions: 0, rejections: 0 };

	test('the generated ' + name + ' compiles', () => {
		game = compileGame(make());
	});

	test(name + ' has ' + expected.corpus + ' corpus documents and ' + expected.fuzz + ' fuzz documents', () => {
		assert.equal(corpus.length, expected.corpus);
		assert.equal(fuzz.length, expected.fuzz);
	});

	const check = (file: string, doc: Document): void => {
		const result = runDocument(game, doc);
		result.name = file.replace(/\.json\.gz$/, '');
		results.push(result);
		const problems: string[] = [];
		if (result.fatal) problems.push('replay stopped: ' + result.fatal);
		for (const field of FIELDS) {
			const r = result.fields[field];
			if (r.mismatches > 0) problems.push(field + ': ' + r.mismatches + ' of ' + r.compared + ' differ, e.g.\n    ' + r.examples.join('\n    '));
		}
		assert.equal(problems.length, 0, '\n  ' + problems.join('\n  '));
		assert.equal(result.fields.state.compared, doc.versions.length, 'every version was compared');
		assert.equal(result.fields['view.game'].compared, doc.versions.length * doc.viewers.length, 'every viewer of every version was compared');
		totals.documents++;
		totals.versions += doc.versions.length;
		totals.viewerVersions += doc.versions.length * doc.viewers.length;
	};

	for (const file of corpus) {
		test('generated ' + name + ': ' + file, () => check(file, read(file)));
	}
	for (const file of fuzz) {
		test('generated ' + name + ': ' + file, () => {
			const doc = read(file);
			check(file, doc);
			totals.rejections += checkRejections(game, file, doc);
		});
	}

	test('generated ' + name + ' pass table', t => {
		t.diagnostic('\n' + passTable(results));
		t.diagnostic(`documents ${totals.documents}, versions ${totals.versions}, viewer-versions ${totals.viewerVersions}, rejections ${totals.rejections}`);
		assert.equal(totals.documents, corpus.length + fuzz.length);
	});
}
