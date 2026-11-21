// The rules layer and the game definitions run inside the same sandboxed
// handler as the engine core, so they are held to the same rule: no DOM, no
// Node, no crypto, no timers, no clock, no ambient randomness, no
// asynchrony, and no imports from outside the engine. This reads the sources
// of rules/ and games/ and fails by name if any of that creeps in.
// (../purity.test.ts does the same for the core; legal/ has its own rules.)
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const ENGINE = fileURLToPath(new URL('..', import.meta.url));
const files: string[] = [];
for (const dir of ['rules', 'games']) {
	for (const name of readdirSync(join(ENGINE, dir))) {
		if (name.endsWith('.ts') && !name.endsWith('.test.ts')) files.push(dir + '/' + name);
	}
}

/** Source with comments and string literals blanked, so prose cannot trip a rule. */
function code(text: string): string {
	return text
		.replace(/\/\*[\s\S]*?\*\//g, ' ')
		.replace(/\/\/[^\n]*/g, ' ')
		.replace(/'(?:[^'\\\n]|\\.)*'/g, "''")
		.replace(/"(?:[^"\\\n]|\\.)*"/g, '""');
}

test('the rules layer and the four games are present', () => {
	for (const expected of ['rules/engine.ts', 'rules/view.ts', 'rules/types.ts', 'games/memory.ts', 'games/blackjack.ts', 'games/tictactoe.ts', 'games/pig.ts']) {
		assert.ok(files.includes(expected), expected + ' is missing');
	}
});

test('rules and games import only engine modules', () => {
	for (const name of files) {
		const text = readFileSync(join(ENGINE, name), 'utf8');
		for (const match of text.matchAll(/(?:import|export)[^;]*?from\s+'([^']+)'/g)) {
			assert.match(match[1], /^(\.\/|\.\.\/|\.\.\/legal\/|\.\.\/rules\/|\.\.\/games\/)[A-Za-z0-9-]+\.ts$/, `${name} imports ${match[1]}`);
		}
		assert.doesNotMatch(code(text), /\brequire\s*\(|\bimport\s*\(/, `${name} loads a module dynamically`);
	}
});

test('rules and games use no platform API, clock, timer, or ambient randomness', () => {
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
	];
	for (const name of files) {
		const text = code(readFileSync(join(ENGINE, name), 'utf8'));
		for (const [pattern, what] of forbidden) {
			const found = pattern.exec(text);
			assert.equal(found, null, `${name} uses ${what}: ${found ? found[0] : ''}`);
		}
	}
});
