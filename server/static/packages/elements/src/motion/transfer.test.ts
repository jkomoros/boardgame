import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { compileMotionTransferDeclarations } from './transfer.ts';

describe('motion transfer declarations', () => {
  it('copies, normalizes, freezes, and preserves declared order', () => {
    const result = compileMotionTransferDeclarations([
      { key: ' deal:0 ', subjectId: 'card-1', source: 'deck', carrier: 'slot-0' },
      { key: 'deal:1', subjectId: 'card-2', source: 'deck', carrier: 'slot-1', durationMs: 250 },
    ]);
    assert.deepEqual(result, [
      { key: 'deal:0', subjectId: 'card-1', source: 'deck', carrier: 'slot-0', durationMs: 500, timing: 'version' },
      { key: 'deal:1', subjectId: 'card-2', source: 'deck', carrier: 'slot-1', durationMs: 250, timing: 'version' },
    ]);
    assert.ok(Object.isFrozen(result));
    assert.ok(result.every(Object.isFrozen));
  });

  it('rejects malformed or conflicting batches atomically', () => {
    const valid = { key: 'deal:0', subjectId: 'card-1', source: 'deck', carrier: 'slot-0' };
    for (const declarations of [
      [{ ...valid, key: ' ' }],
      [{ ...valid, durationMs: Number.NaN }],
      [valid, { ...valid, subjectId: 'card-2', carrier: 'slot-1' }],
      [valid, { ...valid, key: 'deal:1', carrier: 'slot-1' }],
      [valid, { ...valid, key: 'deal:1', subjectId: 'card-2' }],
    ]) {
      assert.throws(() => compileMotionTransferDeclarations(declarations));
    }
  });
});
