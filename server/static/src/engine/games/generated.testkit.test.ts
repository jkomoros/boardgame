// The test every generated game (games/<game>/, written by `boardgame-util
// transpile ./examples/<game>` from Go's own source and exported
// definition) runs: everything Go recorded for the game, the conformance
// corpus (conformance/<game>/) and the fuzz documents
// (conformance/fuzz/<game>/), every field of every version and viewer
// through the shared runner, and each fuzz document's `rejections` through
// testing/fuzz.ts's runFuzzDocument: at its `atVersion` the refused outside
// input is proposed again and the engine must refuse it with Go's error
// text, producing no version and changing nothing.
//
// With BOARDGAME_FUZZ_DIR naming a directory (the output of
// `boardgame-util fuzz run --game G --seeds A..B --out DIR`), every
// `<game>/seed-NNNN.json.gz` under it is replayed against the generated game
// too (the transient acceptance run). games/fuzz.test.ts replays the same
// directory through its own registry of the same games. Until the hand ports
// were removed it held the hand ports, so one run checked both against the
// same documents; fabric/transpile/status/replacement.md records the last
// such run.
//
// This file declares no test of its own; games/<game>/<game>.gen.test.ts
// calls generatedGameSuite.
import assert from 'node:assert/strict';
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import type { Bodies } from '../rules/data.ts';
import type { CompiledGame, GameSpec } from '../rules/index.ts';
import { compileGame } from '../rules/index.ts';
import { type DocumentResult, FIELDS, passTable, runDocument } from '../testing/conformance.ts';
import { type FuzzDocument, type FuzzResult, fuzzPassTable, runFuzzDocument } from '../testing/fuzz.ts';

function corpusDir(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'index.json'))) return path.join(dir, 'conformance');
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find conformance/ above ' + import.meta.url);
		dir = parent;
	}
}

/** `seed-NNNN.json.gz` files of one game's directory, in name order, as absolute paths. */
function seedFiles(dir: string): string[] {
	if (!existsSync(dir)) return [];
	return readdirSync(dir)
		.filter(file => /^seed-\d+\.json\.gz$/.test(file))
		.sort()
		.map(file => path.join(dir, file));
}

const readGz = <T>(file: string): T => JSON.parse(gunzipSync(readFileSync(file)).toString('utf8')) as T;

/**
 * `bodies` with the given test-side bodies answered first: how a generated
 * game whose bodies.gen.ts lacks a refused function's body is loaded in a
 * test until that body exists. `pending` maps FnRef to the stand-in.
 */
export function withPendingBodies(bodies: Bodies, pending: Readonly<Record<string, (...args: never[]) => unknown>>): Bodies {
	return ref => (Object.prototype.hasOwnProperty.call(pending, ref) ? pending[ref] : bodies(ref));
}

/** Running totals of one suite, for its pass table. */
export interface SuiteTotals {
	documents: number;
	versions: number;
	viewerVersions: number;
	rejections: number;
}

/** The problems of one fuzz result, empty when it passed and compared everything. */
export function fuzzProblems(result: FuzzResult, doc: FuzzDocument): string[] {
	const problems: string[] = [];
	if (result.document.fatal) problems.push('replay stopped: ' + result.document.fatal);
	for (const field of FIELDS) {
		const f = result.document.fields[field];
		if (f.mismatches > 0) problems.push(field + ': ' + f.mismatches + ' of ' + f.compared + ' differ, e.g.\n    ' + f.examples.join('\n    '));
	}
	if (result.rejections.mismatches > 0) problems.push('rejections: ' + result.rejections.mismatches + ' of ' + result.rejections.compared + ' differ, e.g.\n    ' + result.rejections.examples.join('\n    '));
	if (result.report) problems.push('divergence report: ' + JSON.stringify(result.report, null, 2));
	if (problems.length > 0) return problems;
	if (result.document.fields.state.compared !== doc.versions.length) problems.push('compared ' + result.document.fields.state.compared + ' of ' + doc.versions.length + ' versions');
	if (result.document.fields['view.game'].compared !== doc.versions.length * doc.viewers.length) problems.push('not every viewer of every version was compared');
	if (result.rejections.compared !== (doc.rejections ?? []).length) problems.push('compared ' + result.rejections.compared + ' of ' + (doc.rejections ?? []).length + ' rejections');
	return problems;
}

/**
 * Registers the tests of one generated game: it compiles, then each corpus
 * and fuzz document passes, then a pass table with the totals compared.
 * `make` builds the game under test (normally the generated index.ts's
 * `<game>()`).
 */
export function generatedGameSuite(name: string, make: () => GameSpec, expected: { corpus: number; fuzz: number }): void {
	const dir = corpusDir();
	const index = JSON.parse(readFileSync(path.join(dir, 'index.json'), 'utf8')) as { file: string; game: string }[];
	const corpus = index.filter(entry => entry.game === name).map(entry => entry.file);
	const fuzz = seedFiles(path.join(dir, 'fuzz', name));

	let compiled: CompiledGame | undefined;
	const game = (): CompiledGame => {
		compiled ??= compileGame(make());
		return compiled;
	};
	const results: DocumentResult[] = [];
	const fuzzResults: FuzzResult[] = [];
	const totals: SuiteTotals = { documents: 0, versions: 0, viewerVersions: 0, rejections: 0 };
	const count = (doc: FuzzDocument, rejections: number): void => {
		totals.documents++;
		totals.versions += doc.versions.length;
		totals.viewerVersions += doc.versions.length * doc.viewers.length;
		totals.rejections += rejections;
	};

	test('the generated ' + name + ' compiles', () => {
		game();
	});

	test(name + ' has ' + expected.corpus + ' corpus documents and ' + expected.fuzz + ' fuzz documents', () => {
		assert.equal(corpus.length, expected.corpus);
		assert.equal(fuzz.length, expected.fuzz);
	});

	for (const file of corpus) {
		test('generated ' + name + ': ' + file, () => {
			const doc = readGz<FuzzDocument>(path.join(dir, file));
			const result = runDocument(game(), doc);
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
			count(doc, 0);
		});
	}

	const fuzzTest = (label: string, file: string, into: FuzzResult[], add: boolean): void => {
		test(label + ': ' + name + '/' + path.basename(file, '.json.gz'), () => {
			const doc = readGz<FuzzDocument>(file);
			const result = runFuzzDocument(game(), doc, { spec: make });
			result.name = name + '/' + path.basename(file, '.json.gz');
			into.push(result);
			const problems = fuzzProblems(result, doc);
			assert.equal(problems.length, 0, '\n  ' + problems.join('\n  '));
			if (add) count(doc, result.rejections.compared);
		});
	};
	for (const file of fuzz) fuzzTest('generated ' + name + ' fuzz', file, fuzzResults, true);

	test('generated ' + name + ' pass table', t => {
		t.diagnostic('\n' + passTable(results));
		t.diagnostic('\n' + fuzzPassTable(fuzzResults));
		t.diagnostic(`documents ${totals.documents}, versions ${totals.versions}, viewer-versions ${totals.viewerVersions}, rejections ${totals.rejections}`);
		assert.equal(totals.documents, corpus.length + fuzz.length);
	});

	const transient = process.env.BOARDGAME_FUZZ_DIR;
	if (transient !== undefined && transient !== '') {
		const files = seedFiles(path.join(transient, name));
		const transientResults: FuzzResult[] = [];
		for (const file of files) fuzzTest('generated ' + name + ' fuzz (BOARDGAME_FUZZ_DIR)', file, transientResults, false);
		if (files.length > 0) {
			test('generated ' + name + ' BOARDGAME_FUZZ_DIR pass table', t => {
				t.diagnostic('\n' + fuzzPassTable(transientResults));
				const sum = { versions: 0, viewerVersions: 0, rejections: 0 };
				for (const r of transientResults) {
					sum.versions += r.document.versions;
					sum.viewerVersions += r.document.versions * r.document.viewers;
					sum.rejections += r.rejections.compared;
				}
				t.diagnostic(`BOARDGAME_FUZZ_DIR ${name}: documents ${transientResults.length}, versions ${sum.versions}, viewer-versions ${sum.viewerVersions}, rejections ${sum.rejections}, passed ${transientResults.filter(r => r.passed).length}`);
				assert.equal(transientResults.length, files.length);
			});
		}
	}
}
