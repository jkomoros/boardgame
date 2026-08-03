import assert from 'node:assert/strict';
import { describe, test } from 'node:test';
import { resolveStructuralContinuity } from './continuity.ts';
import type { MotionContinuityResolution } from './continuity.ts';

const exact = (subjectId: string, collectionId: string) => ({ subjectId, collectionId });
const history = (collectionId: string, lastSeen: Record<string, number>) => ({ collectionId, lastSeen });

// `reason` lives only on the unresolved arm of the union -- a resolved
// continuity has no reason to give. Reading it therefore has to assert the arm
// first, which also makes each failure-mode test say out loud that it expects a
// refusal rather than merely that some `reason` field happens to match.
function unresolved(
  result: MotionContinuityResolution,
): Extract<MotionContinuityResolution, { status: 'unresolved' }> {
  assert.ok(result.status === 'unresolved',
    `expected an unresolved continuity, got ${JSON.stringify(result)}`);
  return result;
}

describe('structural continuity', () => {
  test('exact identity dominates contradictory history', () => {
    const result = resolveStructuralContinuity('card', [exact('card', 'a')], [exact('card', 'b')], [
      history('c', { card: 99 }),
    ]);
    assert.deepEqual(result, {
      status: 'resolved', subjectId: 'card', presence: 'retained',
      from: { kind: 'subject', phase: 'before', collectionId: 'a' },
      to: { kind: 'subject', phase: 'after', collectionId: 'b' },
      evidence: 'identity',
    });
    assert.ok(Object.isFrozen(result));
    assert.ok(result.status !== 'resolved' || Object.isFrozen(result.from));
  });

  test('resolves appearing and departing symmetrically from unique external history', () => {
    const histories = [history('current', { card: 8 }), history('other', { card: 7 })];
    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], histories, 'strict'), {
      status: 'resolved', subjectId: 'card', presence: 'appearing',
      from: { kind: 'collection', collectionId: 'other' },
      to: { kind: 'subject', phase: 'after', collectionId: 'current' },
      evidence: 'history',
    });
    assert.deepEqual(resolveStructuralContinuity('card', [exact('card', 'current')], [], histories, 'strict'), {
      status: 'resolved', subjectId: 'card', presence: 'departing',
      from: { kind: 'subject', phase: 'before', collectionId: 'current' },
      to: { kind: 'collection', collectionId: 'other' },
      evidence: 'history',
    });
  });

  test('is permutation invariant and never chooses among tied candidates', () => {
    const histories = [
      history('current', { card: 9 }),
      history('b', { card: 7 }),
      history('a', { card: 7 }),
    ];
    const expected = {
      status: 'unresolved', subjectId: 'card', endpoint: 'source', reason: 'ambiguous-history',
    };
    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], histories, 'strict'), expected);
    assert.deepEqual(resolveStructuralContinuity(
      'card', [], [exact('card', 'current')], [...histories].reverse(), 'strict',
    ), expected);
  });

  test('fails closed for duplicate identity, same-stack-only, malformed, and absent evidence', () => {
    assert.equal(unresolved(
      resolveStructuralContinuity('card', [exact('card', 'a'), exact('card', 'b')], [], []),
    ).reason, 'duplicate-exact-sighting');
    assert.equal(unresolved(resolveStructuralContinuity(
      'card', [], [exact('card', 'a')], [history('a', { card: 1 })], 'strict',
    )).reason, 'missing-history');
    assert.equal(unresolved(
      resolveStructuralContinuity('card', [], [exact('card', 'a')], [history('b', { card: NaN })]),
    ).reason, 'invalid-history');
    assert.equal(unresolved(resolveStructuralContinuity('card', [], [], [])).reason, 'absent-both-sides');
  });

  test('does not expose history versions or candidate sets', () => {
    const result = resolveStructuralContinuity('card', [], [exact('card', 'a')], [
      history('a', { card: 4 }), history('b', { card: 3 }),
    ]);
    assert.equal(JSON.stringify(result).includes('3'), false);
    assert.equal(JSON.stringify(result).includes('4'), false);
  });

  test('defaults to the historical ordered winner/runner-up behavior', () => {
    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], [
      history('current', { card: 9 }),
      history('previous', { card: 8 }),
    ]), {
      status: 'resolved', subjectId: 'card', presence: 'appearing',
      from: { kind: 'collection', collectionId: 'previous' },
      to: { kind: 'subject', phase: 'after', collectionId: 'current' },
      evidence: 'history',
    });

    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], [
      history('first', { card: 7 }),
      history('second', { card: 7 }),
      history('third', { card: 7 }),
    ]), {
      status: 'resolved', subjectId: 'card', presence: 'appearing',
      from: { kind: 'collection', collectionId: 'second' },
      to: { kind: 'subject', phase: 'after', collectionId: 'current' },
      evidence: 'history',
    });
  });

  test('preserves the legacy same-collection fallback when no runner-up exists', () => {
    const result = resolveStructuralContinuity('card', [], [exact('card', 'current')], [
      history('current', { card: 3 }),
    ]);
    assert.equal(result.status, 'resolved');
    if (result.status === 'resolved') {
      assert.deepEqual(result.from, { kind: 'collection', collectionId: 'current' });
    }
  });
});
