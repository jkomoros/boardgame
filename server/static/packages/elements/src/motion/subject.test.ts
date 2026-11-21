import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { motionSilhouette, sanitizeMotionSubjectSnapshot } from './subject.ts';

describe('motion subject snapshots', () => {
  it('creates a minimal immutable silhouette capability', () => {
    const snapshot = motionSilhouette('rounded-rectangle');
    assert.deepEqual(snapshot, { kind: 'silhouette', shape: 'rounded-rectangle' });
    assert.equal(Object.isFrozen(snapshot), true);
  });

  it('copies exact safe values and rejects content-bearing extensions', () => {
    const source = { kind: 'silhouette', shape: 'circle' };
    const snapshot = sanitizeMotionSubjectSnapshot(source);
    source.shape = 'rectangle';
    assert.deepEqual(snapshot, { kind: 'silhouette', shape: 'circle' });
    assert.equal(sanitizeMotionSubjectSnapshot({
      kind: 'silhouette', shape: 'circle', text: 'hidden face',
    }), null);
    assert.equal(sanitizeMotionSubjectSnapshot({ kind: 'artwork', src: 'secret.png' }), null);
    assert.equal(sanitizeMotionSubjectSnapshot(null), null);
  });

  it('rejects unsupported shapes at the public constructor', () => {
    assert.throws(() => motionSilhouette('star' as never), /shape/);
  });
});
