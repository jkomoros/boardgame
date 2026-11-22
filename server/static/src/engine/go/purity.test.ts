// The facade runs inside the same sandboxed handler as the rest of the
// engine, and translated game code imports it, so it is held to the same
// rule as rules/ and games/ (../rules/purity.test.ts): no DOM, no Node, no
// crypto, no timers, no clock, no ambient randomness, no BigInt, no
// asynchrony, and no imports from outside the engine. ../sandbox.test.ts
// separately checks what it holds at module scope.
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const HERE = fileURLToPath(new URL('.', import.meta.url));
const files = readdirSync(HERE).filter(name => name.endsWith('.ts') && !name.endsWith('.test.ts'));

/** Source with comments and string literals blanked, so prose cannot trip a rule. */
function code(text: string): string {
	return text
		.replace(/\/\*[\s\S]*?\*\//g, ' ')
		.replace(/\/\/[^\n]*/g, ' ')
		.replace(/'(?:[^'\\\n]|\\.)*'/g, "''")
		.replace(/"(?:[^"\\\n]|\\.)*"/g, '""');
}

test('the facade is the five frozen modules', () => {
	assert.deepEqual(files.sort(), ['api.ts', 'num.ts', 'panic.ts', 'slice.ts', 'str.ts']);
});

test('facade modules import only engine modules', () => {
	for (const name of files) {
		const text = readFileSync(join(HERE, name), 'utf8');
		for (const match of text.matchAll(/(?:import|export)[^;]*?from\s+'([^']+)'/g)) {
			assert.match(match[1], /^(\.\/|\.\.\/|\.\.\/legal\/|\.\.\/rules\/)[A-Za-z0-9-]+\.ts$/, `${name} imports ${match[1]}`);
		}
		assert.doesNotMatch(code(text), /\brequire\s*\(|\bimport\s*\(/, `${name} loads a module dynamically`);
	}
});

test('facade modules use no platform API, clock, timer, BigInt or ambient randomness', () => {
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
		[/\bclass\b/, 'a class'],
	];
	for (const name of files) {
		const text = code(readFileSync(join(HERE, name), 'utf8'));
		for (const [pattern, what] of forbidden) {
			const found = pattern.exec(text);
			assert.equal(found, null, `${name} uses ${what}: ${found ? found[0] : ''}`);
		}
	}
});

test('every facade function is implemented, or throws a message naming its engine gap', () => {
	for (const name of files) {
		const text = readFileSync(join(HERE, name), 'utf8');
		assert.doesNotMatch(text, /not implemented/, `${name} still has a frozen stub`);
		for (const match of text.matchAll(/throw new Error\(\s*'(engine-gap:[^']*)'/g)) assert.match(match[1], /^engine-gap: \w+ needs /);
	}
});
