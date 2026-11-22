// The engine runs inside a sandboxed JavaScript handler: no DOM, no Node, no
// crypto, no timers, no clock, no ambient randomness, no dependencies. This
// test reads the engine's own sources and fails by name if any of that creeps
// in. It covers the files this directory owns directly; subdirectories
// (testing/ is Node-only test support, legal/ belongs to another module) are
// not scanned.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import * as engine from './index.ts';

const HERE = fileURLToPath(new URL('.', import.meta.url));
const sources = readdirSync(HERE).filter(name => name.endsWith('.ts') && !name.endsWith('.test.ts'));

/** Source with comments and string literals blanked, so prose cannot trip a rule. */
function code(text: string): string {
	return text
		.replace(/\/\*[\s\S]*?\*\//g, ' ')
		.replace(/\/\/[^\n]*/g, ' ')
		.replace(/'(?:[^'\\\n]|\\.)*'/g, "''")
		.replace(/"(?:[^"\\\n]|\\.)*"/g, '""');
}

test('the engine sources are the expected set', () => {
	assert.deepEqual(sources.sort(), [
		'constraints.ts', 'definition.ts', 'fnv.ts', 'gosort.ts', 'ids.ts', 'index.ts', 'rand-cooked.ts', 'rand.ts', 'sanitize.ts', 'sha1.ts', 'stack.ts', 'state.ts', 'u64.ts',
	]);
});

test('engine modules import only each other', () => {
	for (const name of sources) {
		const text = readFileSync(join(HERE, name), 'utf8');
		for (const match of text.matchAll(/(?:import|export)[^;]*?from\s+'([^']+)'/g)) {
			assert.match(match[1], /^\.\/[a-z0-9-]+\.ts$/, `${name} imports ${match[1]}`);
			assert.ok(sources.includes(match[1].slice(2)), `${name} imports ${match[1]}, which is not an engine module`);
		}
		assert.doesNotMatch(code(text), /\brequire\s*\(|\bimport\s*\(/, `${name} loads a module dynamically`);
	}
});

test('engine modules use no platform API, clock, timer, or ambient randomness', () => {
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
	for (const name of sources) {
		const text = code(readFileSync(join(HERE, name), 'utf8'));
		for (const [pattern, what] of forbidden) {
			const found = pattern.exec(text);
			assert.equal(found, null, `${name} uses ${what}: ${found ? found[0] : ''}`);
		}
	}
});

test('the public surface is what README.md documents', () => {
	assert.deepEqual(Object.keys(engine).sort(), [
		'ADMIN_PLAYER_INDEX', 'ANY_PLAYER_INDEX', 'EMPTY_SLOT', 'GENERIC_COMPONENT', 'GROUP_ALL', 'GROUP_OTHER', 'GROUP_SELF',
		'OBSERVER_PLAYER_INDEX', 'POLICY_ORDER', 'UNLIMITED_SLOTS', 'cloneJSON', 'componentId', 'facetSurvives', 'fnv1_64', 'fnv1_64Bytes',
		'goSort', 'indexesOf', 'isMergedStackView', 'isSanitizeError', 'isStack', 'isStackError', 'isStateError', 'loadEngineState', 'newGoRand',
		'parseSanitizeTag', 'policyFromString', 'policyMapFor', 'resolveSanitizationPolicy', 'sanitizationTransformation', 'sanitizeError',
		'sanitizedForPlayer', 'secretMoveCountFor', 'sha1Hex', 'sha1HexBytes', 'stackError', 'stateError', 'stateRand', 'stateRandSeed',
		'stateRandSeedInput', 'viewFor', 'visibleFacets',
	]);
});
