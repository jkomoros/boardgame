// lib/ runs inside the same sandboxed handler as the rest of the engine, so
// it is held to rules/purity.test.ts's rule: no DOM, no Node, no crypto, no
// timers, no clock, no ambient randomness, no asynchrony, and imports only
// of engine modules. (../sandbox.test.ts covers the module-scope rules.)
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const LIB = fileURLToPath(new URL('.', import.meta.url));
const files = readdirSync(LIB).filter(name => name.endsWith('.ts') && !name.endsWith('.test.ts'));

/** Source with comments and string literals blanked, so prose cannot trip a rule. */
function code(text: string): string {
	return text
		.replace(/\/\*[\s\S]*?\*\//g, ' ')
		.replace(/\/\/[^\n]*/g, ' ')
		.replace(/'(?:[^'\\\n]|\\.)*'/g, "''")
		.replace(/"(?:[^"\\\n]|\\.)*"/g, '""');
}

test('the four libraries are present', () => {
	for (const expected of ['enumrange.ts', 'graph.ts', 'dice.ts', 'playingcards.ts']) assert.ok(files.includes(expected), expected + ' is missing');
});

test('lib imports only engine modules', () => {
	for (const name of files) {
		const text = readFileSync(join(LIB, name), 'utf8');
		for (const match of text.matchAll(/(?:import|export)[^;]*?from\s+'([^']+)'/g)) {
			assert.match(match[1], /^(\.\/|\.\.\/|\.\.\/rules\/)[A-Za-z0-9-]+\.ts$/, `${name} imports ${match[1]}`);
		}
		assert.doesNotMatch(code(text), /\brequire\s*\(|\bimport\s*\(/, `${name} loads a module dynamically`);
	}
});

test('lib uses no platform API, clock, timer, or ambient randomness', () => {
	const forbidden: [RegExp, string][] = [
		[/\bBigInt\b|\b\d+n\b/, 'BigInt'],
		[/\bcrypto\b/, 'crypto'],
		[/\bsetTimeout\b|\bsetInterval\b|\bsetImmediate\b|\bqueueMicrotask\b|\brequestAnimationFrame\b/, 'a timer'],
		[/\bDate\b|\bperformance\b/, 'a clock'],
		[/Math\.random/, 'Math.random'],
		[/\bdocument\b|\bwindow\b|\bnavigator\b|\blocalStorage\b|\bfetch\b|\bglobalThis\b/, 'a browser global'],
		[/\bprocess\b|\bBuffer\b|\b__dirname\b/, 'a Node global'],
		[/\bTextEncoder\b|\bTextDecoder\b|\bstructuredClone\b|\bAtomics\b|\bSharedArrayBuffer\b|\bWebAssembly\b/, 'a host API'],
		[/\basync\b|\bawait\b|\bPromise\b/, 'asynchrony'],
		[/\beval\b|\bnew Function\b/, 'dynamic code'],
		[/\bclass\s/, 'a class'],
	];
	for (const name of files) {
		const text = code(readFileSync(join(LIB, name), 'utf8'));
		for (const [pattern, what] of forbidden) {
			const found = pattern.exec(text);
			assert.equal(found, null, `${name} uses ${what}: ${found ? found[0] : ''}`);
		}
	}
});
