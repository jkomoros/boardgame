// The mutation harness's own pieces (scripts/fuzz-mutations.mjs runs them):
// every historical mutation still names exactly one place in the current
// sources, and the body generator produces the mutants it promises.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import ts from 'typescript';
import { HISTORICAL_MUTATIONS, LOOP_GUARD_COUNTER, applyMutation, bodyMutants, occurrences } from './mutate.ts';

const engineFile = (relative: string): string => readFileSync(fileURLToPath(new URL('../' + relative, import.meta.url)), 'utf8');

test('the 20 historical mutations each apply to exactly one place', () => {
	assert.equal(HISTORICAL_MUTATIONS.length, 20);
	assert.equal(new Set(HISTORICAL_MUTATIONS.map(m => m.id)).size, 20, 'ids are unique');
	for (const mutation of HISTORICAL_MUTATIONS) {
		const source = engineFile(mutation.file);
		assert.equal(occurrences(source, mutation.find), 1, mutation.id + ': ' + JSON.stringify(mutation.find) + ' in ' + mutation.file);
		const mutated = applyMutation(source, mutation);
		assert.notEqual(mutated, source, mutation.id);
		assert.equal(mutated.length, source.length - mutation.find.length + mutation.replace.length, mutation.id);
	}
});

test('applyMutation refuses a text that does not occur exactly once', () => {
	const base = { id: 'x', description: '', failedOn: '', file: 'f.ts', replace: 'y' };
	assert.throws(() => applyMutation('a a', { ...base, find: 'a' }), /occurs 2 times/);
	assert.throws(() => applyMutation('b', { ...base, find: 'a' }), /occurs 0 times/);
});

const SAMPLE = `import { x } from './x.ts';
export function target(n: number): number {
	let total = 0;
	for (let i = 0; i < n; i++) total += 2;
	if (total >= 10) return total;
	return -1;
}
export function other(): number { return 1 < 2 ? 3 : 4; }
`;

test('bodyMutants: every comparison, integer literal and statement of the function, one at a time', () => {
	const result = bodyMutants(ts, 'sample.ts', SAMPLE, 'target');
	assert.equal(result.startLine, 2);
	assert.equal(result.endLine, 7);
	const summary = result.mutants.map(m => m.kind + ' ' + m.line + ':' + m.column + ' ' + m.before.split(' ')[0] + ' -> ' + m.after);
	assert.deepEqual(summary, [
		'shift-literal 3:14 0 -> 1',
		'delete-statement 4:2 for -> (deleted)',
		'shift-literal 4:15 0 -> 1',
		'flip-comparison 4:20 < -> >=',
		'delete-statement 4:30 total -> (deleted)',
		'shift-literal 4:39 2 -> 3',
		'delete-statement 5:2 if -> (deleted)',
		'flip-comparison 5:12 >= -> <',
		'shift-literal 5:15 10 -> 11',
		'delete-statement 5:19 return -> (deleted)',
		'delete-statement 6:2 return -> (deleted)',
		'shift-literal 6:10 1 -> 2',
	]);
	assert.equal(result.skippedDeclarations, 1, 'let total is not deleted');
	assert.equal(result.total, result.mutants.length);

	// Every mutant is the guarded file with one change, and parses.
	for (const mutant of result.mutants) {
		const parsed = ts.createSourceFile('m.ts', mutant.source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
		const diagnostics = (parsed as unknown as { parseDiagnostics: unknown[] }).parseDiagnostics;
		assert.equal(diagnostics.length, 0, 'mutant ' + mutant.id + ' does not parse:\n' + mutant.source);
		assert.ok(mutant.source.includes(LOOP_GUARD_COUNTER) || mutant.kind === 'delete-statement', 'mutant ' + mutant.id + ' lost its loop guard');
	}
	// The loop guard wraps an unbraced body, and other functions are untouched.
	assert.ok(result.guarded.includes('{ if (++globalThis.' + LOOP_GUARD_COUNTER));
	assert.ok(result.guarded.includes('export function other(): number { return 1 < 2 ? 3 : 4; }'));
	// Deleting the unbraced loop body keeps the guard around the empty statement.
	const deletedBody = result.mutants.find(m => m.kind === 'delete-statement' && m.line === 4 && m.column === 30);
	assert.ok(deletedBody?.source.includes(LOOP_GUARD_COUNTER), 'the guard survives deleting the loop body');
});

test('bodyMutants: a real hand-port function', () => {
	const source = engineFile('games/pig.ts');
	const result = bodyMutants(ts, 'games/pig.ts', source, 'pig');
	assert.ok(result.total > 10, 'mutants: ' + result.total);
	assert.ok(result.mutants.some(m => m.kind === 'flip-comparison'));
	assert.ok(result.mutants.some(m => m.kind === 'shift-literal'));
	assert.ok(result.mutants.some(m => m.kind === 'delete-statement'));
	assert.throws(() => bodyMutants(ts, 'games/pig.ts', source, 'noSuchFunction'), /no top-level function/);
});
