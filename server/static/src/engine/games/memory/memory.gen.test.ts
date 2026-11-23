// The generated memory (games/memory/, `boardgame-util transpile
// ./examples/memory`) against everything Go recorded for memory: its 6
// conformance documents and its 8 fuzz documents, including their
// rejections. See ../generated.testkit.test.ts for what is compared.
//
// One function was refused by the translator:
//
//   memory.(*moveStartHideCardsTimer).Apply
//     package-var-read memory.hideCardMoveName
//
// A coding agent translated it from its brief (fabric/TRANSPILE-PROPOSAL.md,
// "The agent loop") and `port accept` stored it in bodies.gen.ts under an
// `@generated agent:` header, so index.ts loads memory on its own and
// everything under test is generated.
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { generatedGameSuite } from '../generated.testkit.test.ts';
import { bodies } from './bodies.gen.ts';
import { memory } from './index.ts';

/** The FnRefs the translator refused, which bodies.gen.ts therefore lacks: none. */
const PENDING: string[] = [];
const AGENT_WRITTEN = ['memory.(*moveStartHideCardsTimer).Apply'];

test('memory: nothing is pending; index.ts loads; the agent-written function has a body', () => {
	assert.deepEqual(PENDING, []);
	for (const ref of AGENT_WRITTEN) assert.notEqual(bodies(ref), undefined, ref + ' has no body');
	assert.equal(memory().name, 'memory');
});

generatedGameSuite('memory', memory, { corpus: 6, fuzz: 8 });
