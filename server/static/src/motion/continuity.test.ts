import assert from 'node:assert/strict';
import { describe, test } from 'node:test';
import { resolveStructuralContinuity } from './continuity.ts';

const exact = (subjectId: string, collectionId: string) => ({ subjectId, collectionId });
const history = (collectionId: string, lastSeen: Record<string, number>) => ({ collectionId, lastSeen });

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
    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], histories), {
      status: 'resolved', subjectId: 'card', presence: 'appearing',
      from: { kind: 'collection', collectionId: 'other' },
      to: { kind: 'subject', phase: 'after', collectionId: 'current' },
      evidence: 'history',
    });
    assert.deepEqual(resolveStructuralContinuity('card', [exact('card', 'current')], [], histories), {
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
    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], histories), expected);
    assert.deepEqual(resolveStructuralContinuity('card', [], [exact('card', 'current')], [...histories].reverse()), expected);
  });

  test('fails closed for duplicate identity, same-stack-only, malformed, and absent evidence', () => {
    assert.equal(resolveStructuralContinuity('card', [exact('card', 'a'), exact('card', 'b')], [], []).reason,
      'duplicate-exact-sighting');
    assert.equal(resolveStructuralContinuity('card', [], [exact('card', 'a')], [history('a', { card: 1 })]).reason,
      'missing-history');
    assert.equal(resolveStructuralContinuity('card', [], [exact('card', 'a')], [history('b', { card: NaN })]).reason,
      'invalid-history');
    assert.equal(resolveStructuralContinuity('card', [], [], []).reason, 'absent-both-sides');
  });

  test('does not expose history versions or candidate sets', () => {
    const result = resolveStructuralContinuity('card', [], [exact('card', 'a')], [
      history('a', { card: 4 }), history('b', { card: 3 }),
    ]);
    assert.equal(JSON.stringify(result).includes('3'), false);
    assert.equal(JSON.stringify(result).includes('4'), false);
  });
});
