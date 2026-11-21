import assert from 'node:assert/strict';
import { it } from 'node:test';
import { effectBudgetSnapshot, reserveEffectBudget } from './effect-budget.ts';

it('shares one degrading budget across effect-layer callers in a document', () => {
  const document = {} as Document;
  const first = reserveEffectBudget(document, 24);
  const second = reserveEffectBudget(document, 24);
  const third = reserveEffectBudget(document, 24);
  assert.equal(first?.particles, 24);
  assert.equal(second?.particles, 24);
  assert.equal(third?.particles, 12);
  assert.deepEqual(effectBudgetSnapshot(document), { effects: 3, particles: 60 });
  assert.equal(reserveEffectBudget(document, 1), null);
  second?.release();
  assert.deepEqual(effectBudgetSnapshot(document), { effects: 2, particles: 36 });
  const replacement = reserveEffectBudget(document, 24);
  assert.equal(replacement?.particles, 24);
  first?.release();
  third?.release();
  replacement?.release();
  replacement?.release();
  assert.deepEqual(effectBudgetSnapshot(document), { effects: 0, particles: 0 });
});
