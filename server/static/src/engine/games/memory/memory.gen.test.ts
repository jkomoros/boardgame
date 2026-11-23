// The generated memory (games/memory/, `boardgame-util transpile
// ./examples/memory`) against everything Go recorded for memory: its 6
// conformance documents and its 8 fuzz documents, including their
// rejections. See ../generated.testkit.test.ts for what is compared.
//
// PENDING: one function is refused by the translator and has no body in
// bodies.gen.ts, so the generated index.ts cannot load memory on its own:
//
//   memory.(*moveStartHideCardsTimer).Apply
//     package-var-read memory.hideCardMoveName
//
// It is bound for the agent loop (fabric/TRANSPILE-PROPOSAL.md, "The agent
// loop"). Until its body exists, this test supplies a TEST-SIDE STAND-IN:
// the hand port's own Start Hide Cards Timer apply (games/memory.ts), wired
// in front of the generated `bodies`. Everything else under test is
// generated.
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { isGameSpecError } from '../../rules/compile.ts';
import { gameFromData } from '../../rules/data.ts';
import type { GameSpec } from '../../rules/types.ts';
import { generatedGameSuite, withPendingBodies } from '../generated.testkit.test.ts';
import { memory } from '../memory.ts';
import { bodies } from './bodies.gen.ts';
import { definition } from './definition.gen.ts';
import { memoryGenerated } from './index.ts';

/** The FnRefs the translator refused, which bodies.gen.ts therefore lacks. */
/** None now: the one refused function was translated by a coding agent from its brief and accepted (see bodies.gen.ts, `@generated agent:`). */
const PENDING: string[] = [];
const AGENT_WRITTEN = ['memory.(*moveStartHideCardsTimer).Apply'];

test('memory: nothing is pending; index.ts loads; the agent-written function has a body', () => {
	assert.deepEqual(PENDING, []);
	for (const ref of AGENT_WRITTEN) assert.notEqual(bodies(ref), undefined, ref + ' has no body');
	assert.equal(memoryGenerated().name, 'memory');
	void definition;
});

generatedGameSuite('memory', memoryGenerated, { corpus: 6, fuzz: 8 });
