// The engine has to load inside a Common Fabric pattern, whose sandbox
// verifies every module-scope statement before a module may run (labs
// docs/specs/sandboxing/SES_SANDBOXING_SPEC.md, section 4.2). This test
// reads every engine source, the rules layer, the legal evaluator and the
// games included, and fails by name on each form that sandbox refuses:
//
//   - a class at module scope (it would hold mutable prototype state);
//   - `let` or `var` at module scope;
//   - a module-scope constant that is not a function and not plain data: a
//     record holding functions, a `new` of anything, a stateful regular
//     expression, or arithmetic over other constants (`DIM * DIM`). A call is allowed, since its result is checked when the
//     module loads; games/*.ts are loaded by sandbox tests for that;
//   - a generator at module scope;
//   - any statement at module scope that is not a declaration: a loop that
//     fills a table, an assignment, a call;
//   - a double cast through `unknown`, which the pattern compiler rejects;
//   - an import that leaves the engine directory.
//
// It is the cheap, labs-free half of the check. The other half is loading
// the engine in a real pattern: fabric/patterns/*/parity.test.tsx.
//
// The rules themselves are testing/sourceRules.ts sandboxProblems, which
// `boardgame-util port check` also runs on an agent's translation.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join, relative, resolve, dirname } from 'node:path';
import ts from 'typescript';
import { sandboxProblems } from './testing/sourceRules.ts';

const HERE = fileURLToPath(new URL('.', import.meta.url));

function sources(dir: string, out: string[] = []): string[] {
	for (const entry of readdirSync(dir, { withFileTypes: true })) {
		const path = join(dir, entry.name);
		if (entry.isDirectory()) {
			if (entry.name === 'testing' || entry.name === 'testdata') continue;
			sources(path, out);
		} else if (entry.name.endsWith('.ts') && !entry.name.endsWith('.test.ts')) {
			out.push(path);
		}
	}
	return out;
}

const files = sources(HERE);

test('every engine source is scanned', () => {
	assert.ok(files.length >= 49, `only ${files.length} sources found`);
	for (const dir of ['rules', 'legal', 'games']) {
		assert.ok(files.some(file => relative(HERE, file).startsWith(dir + '/')), `${dir}/ is not scanned`);
	}
});

test('no engine module holds anything at module scope that the pattern sandbox refuses', () => {
	const problems: string[] = [];
	for (const file of files) {
		const name = relative(HERE, file);
		const text = readFileSync(file, 'utf8');
		const escapes = (specifier: string): boolean => relative(HERE, resolve(dirname(file), specifier)).startsWith('..');
		problems.push(...sandboxProblems(ts, name, text, escapes));
	}
	assert.deepEqual(problems, []);
});
