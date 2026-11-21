// The ported games against the conformance corpus (conformance/ at the
// repository root): every document for memory, blackjack, tictactoe, pig,
// debuganimations and werewolf is replayed on the TypeScript engine from its recorded id, salt, players,
// variant, seating and outside inputs alone, and at every version the stored
// state, the move record, and every viewer's game object, sanitized move,
// legality and projected choices must equal what Go recorded.
//
// See ../testing/conformance.ts for exactly what is compared and how.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import type { CompiledGame, GameSpec } from '../rules/index.ts';
import { compileGame } from '../rules/index.ts';
import { type ConformanceDocument, type DocumentResult, FIELDS, passTable, runDocument } from '../testing/conformance.ts';
import { blackjack } from './blackjack.ts';
import { debuganimations } from './debuganimations.ts';
import { memory } from './memory.ts';
import { pig } from './pig.ts';
import { tictactoe } from './tictactoe.ts';
import { werewolf } from './werewolf.ts';

function corpusDir(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'index.json'))) return path.join(dir, 'conformance');
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find conformance/ above ' + import.meta.url);
		dir = parent;
	}
}

const dir = corpusDir();
const index = JSON.parse(readFileSync(path.join(dir, 'index.json'), 'utf8')) as { file: string; game: string; name: string }[];

// The four games this layer set out to port, and the two stretch games.
const specs: Record<string, GameSpec> = { memory: memory(), blackjack: blackjack(), tictactoe: tictactoe(), pig: pig(), debuganimations: debuganimations(), werewolf: werewolf() };
const compiled: Record<string, CompiledGame> = {};
const results: DocumentResult[] = [];

test('every ported game compiles', () => {
	for (const name of Object.keys(specs)) compiled[name] = compileGame(specs[name]);
});

const entries = index.filter(entry => specs[entry.game] !== undefined);

test('the corpus holds documents for every ported game', () => {
	for (const name of Object.keys(specs)) assert.ok(entries.some(entry => entry.game === name), 'no document for ' + name);
	assert.ok(entries.length >= 19, 'documents found: ' + entries.length);
});

for (const entry of entries) {
	test('conformance: ' + entry.file, () => {
		const doc = JSON.parse(gunzipSync(readFileSync(path.join(dir, entry.file))).toString('utf8')) as ConformanceDocument;
		const result = runDocument(compiled[entry.game], doc);
		result.name = entry.file.replace(/\.json\.gz$/, '');
		results.push(result);
		const problems: string[] = [];
		if (result.fatal) problems.push('replay stopped: ' + result.fatal);
		for (const field of FIELDS) {
			const field_ = result.fields[field];
			if (field_.mismatches > 0) problems.push(field + ': ' + field_.mismatches + ' of ' + field_.compared + ' differ, e.g.\n    ' + field_.examples.join('\n    '));
		}
		assert.equal(problems.length, 0, '\n  ' + problems.join('\n  '));
		// A document that compared nothing proves nothing.
		assert.ok(result.fields.state.compared === doc.versions.length, 'every version was compared');
		assert.ok(result.fields['view.game'].compared === doc.versions.length * doc.viewers.length, 'every viewer of every version was compared');
	});
}

test('conformance pass table', t => {
	t.diagnostic('\n' + passTable(results));
	assert.equal(results.length, entries.length);
});
