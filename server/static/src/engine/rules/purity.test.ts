// The rules layer and the game definitions run inside the same sandboxed
// handler as the engine core, so they are held to the same rule: no DOM, no
// Node, no crypto, no timers, no clock, no ambient randomness, no
// asynchrony, and no imports from outside the engine. This reads the sources
// of rules/ and games/ and fails by name if any of that creeps in.
// (../purity.test.ts does the same for the core; legal/ has its own rules.)
// The forbidden list is ../testing/sourceRules.ts, which `boardgame-util
// port check` also runs on an agent's translation.
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { loadsDynamically, platformUses } from '../testing/sourceRules.ts';

const ENGINE = fileURLToPath(new URL('..', import.meta.url));
const files: string[] = [];
for (const dir of ['rules', 'games']) {
	for (const name of readdirSync(join(ENGINE, dir))) {
		if (name.endsWith('.ts') && !name.endsWith('.test.ts')) files.push(dir + '/' + name);
	}
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
		assert.equal(loadsDynamically(text), false, `${name} loads a module dynamically`);
	}
});

test('rules and games use no platform API, clock, timer, or ambient randomness', () => {
	for (const name of files) {
		const text = readFileSync(join(ENGINE, name), 'utf8');
		const found = platformUses(text)[0];
		assert.equal(found, undefined, found ? `${name} uses ${found[0]}: ${found[1]}` : '');
	}
});
