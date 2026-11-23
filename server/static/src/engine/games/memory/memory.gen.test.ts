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
const PENDING = ['memory.(*moveStartHideCardsTimer).Apply'];

function handPortStandIn(): (...args: never[]) => unknown {
	const move = memory().moves.find(m => m.name === 'Start Hide Cards Timer');
	assert.ok(move?.apply !== undefined, "the hand port's Start Hide Cards Timer has no apply");
	return move.apply;
}

const memoryWithStandIn = (): GameSpec => gameFromData(definition(), withPendingBodies(bodies, { [PENDING[0]]: handPortStandIn() }));

test('memory: the pending FnRefs are exactly the refused functions, and index.ts names them when it cannot load', () => {
	for (const ref of PENDING) assert.equal(bodies(ref), undefined, ref + ' has a generated body now; remove the stand-in');
	assert.throws(memoryGenerated, (thrown: unknown) => isGameSpecError(thrown) && PENDING.every(ref => (thrown as Error).message.includes(ref)));
});

generatedGameSuite('memory', memoryWithStandIn, { corpus: 6, fuzz: 8 });
