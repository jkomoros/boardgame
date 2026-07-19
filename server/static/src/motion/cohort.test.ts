import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { compileMotionCohortSchedule, motion } from './cohort.ts';

const entries = Object.freeze([
  Object.freeze({ subjectId: 'a', legacyDelayMs: 10 }),
  Object.freeze({ subjectId: 'b', legacyDelayMs: 20 }),
  Object.freeze({ subjectId: 'c', legacyDelayMs: 30 }),
]);

describe('motion cohort scheduling', () => {
  it('uses explicit subject order rather than playback order', () => {
    const result = compileMotionCohortSchedule(entries, [
      motion.stagger({ subjects: ['c', 'a'], intervalMs: 45 }),
    ]);
    assert.equal(result.status, 'applied');
    assert.deepEqual(result.entries, [
      { subjectId: 'a', delayMs: 45 },
      { subjectId: 'b', delayMs: 20 },
      { subjectId: 'c', delayMs: 0 },
    ]);
    assert.equal(Object.isFrozen(result), true);
    assert.equal(Object.isFrozen(result.entries), true);
    assert.equal(Object.isFrozen(result.entries[0]), true);
  });

  it('ignores missing subjects and preserves nonmembers', () => {
    const result = compileMotionCohortSchedule(entries, [
      motion.stagger({ subjects: ['missing', 'b'], intervalMs: 12 }),
    ]);
    assert.deepEqual(result.entries, [
      { subjectId: 'a', delayMs: 10 },
      { subjectId: 'b', delayMs: 12 },
      { subjectId: 'c', delayMs: 30 },
    ]);
  });

  it('falls back atomically for duplicate or overlapping declarations', () => {
    const duplicate = compileMotionCohortSchedule(entries, [{
      kind: 'stagger', subjects: ['a', 'a'], intervalMs: 5,
    }]);
    assert.equal(duplicate.status, 'fallback');
    assert.deepEqual(duplicate.entries.map(entry => entry.delayMs), [10, 20, 30]);

    const overlap = compileMotionCohortSchedule(entries, [
      motion.stagger({ subjects: ['a'], intervalMs: 5 }),
      motion.stagger({ subjects: ['a', 'b'], intervalMs: 8 }),
    ]);
    assert.equal(overlap.status, 'fallback');
    assert.deepEqual(overlap.entries.map(entry => entry.delayMs), [10, 20, 30]);
  });

  it('falls back for malformed cadence without throwing from playback', () => {
    const result = compileMotionCohortSchedule(entries, [{
      kind: 'stagger', subjects: ['a'], intervalMs: Number.NaN,
    }]);
    assert.equal(result.status, 'fallback');
    assert.equal(result.reason, 'invalid-cohort');
    assert.doesNotThrow(() => compileMotionCohortSchedule(entries, [null as never]));
    assert.equal(
      compileMotionCohortSchedule(entries, [{ kind: 'stagger' } as never]).status,
      'fallback',
    );
  });

  it('validates and freezes author declarations', () => {
    const subjects = ['a', 'b'];
    const cohort = motion.stagger({ subjects, intervalMs: 40, key: 'deal' });
    subjects.push('c');
    assert.deepEqual(cohort, {
      kind: 'stagger', subjects: ['a', 'b'], intervalMs: 40, key: 'deal',
    });
    assert.equal(Object.isFrozen(cohort), true);
    assert.equal(Object.isFrozen(cohort.subjects), true);
    assert.throws(() => motion.stagger({ subjects: [], intervalMs: 1 }), /subjects/);
    assert.throws(() => motion.stagger({ subjects: ['a', 'a'], intervalMs: 1 }), /unique/);
    assert.throws(() => motion.stagger({ subjects: ['a'], intervalMs: -1 }), /intervalMs/);
  });
});
