import assert from 'node:assert/strict';
import test from 'node:test';
import { readinessPresentation } from './readiness.ts';

test('readiness summarizes required participants without treating an empty phase as complete', () => {
  const presentation = readinessPresentation([
    { key: 0, label: 'Ada', state: 'ready' },
    { key: 1, label: 'Grace', state: 'waiting' },
    { key: 2, label: 'Observer', state: 'not-required' },
  ]);
  assert.deepEqual(presentation, {
    participants: [
      { key: 0, label: 'Ada', state: 'ready' },
      { key: 1, label: 'Grace', state: 'waiting' },
      { key: 2, label: 'Observer', state: 'not-required' },
    ],
    readyCount: 1,
    requiredCount: 2,
    complete: false,
    empty: false,
    message: '1 of 2 ready',
  });
  assert.equal(Object.isFrozen(presentation), true);
  assert.equal(Object.isFrozen(presentation.participants), true);
  assert.equal(Object.isFrozen(presentation.participants[0]), true);
  assert.deepEqual(readinessPresentation([]), {
    participants: [], readyCount: 0, requiredCount: 0,
    complete: false, empty: true, message: 'No participants are required',
  });
});

test('readiness supports closed labels and rejects malformed authoring loudly', () => {
  assert.equal(readinessPresentation([
    { key: 'a', label: 'Ada', state: 'ready' },
  ], { complete: 'Choices locked' }).message, 'Choices locked');
  assert.equal(readinessPresentation([
    { key: 'a', label: 'Ada', state: 'waiting' },
  ], { progress: 'votes cast' }).message, '0 of 1 votes cast');
  assert.throws(() => readinessPresentation(null as never), /participants must be an array/);
  assert.throws(() => readinessPresentation([
    { key: 0, label: 'Ada', state: 'ready' },
    { key: 0, label: 'Grace', state: 'waiting' },
  ]), /duplicate participant key/);
  assert.throws(() => readinessPresentation([
    { key: Number.NaN, label: 'Ada', state: 'ready' },
  ]), /invalid key/);
  assert.throws(() => readinessPresentation([
    { key: ' ', label: 'Ada', state: 'ready' },
  ]), /invalid key/);
  assert.throws(() => readinessPresentation([
    { key: 0, label: ' ', state: 'ready' },
  ]), /non-empty label/);
  assert.throws(() => readinessPresentation([
    { key: 0, label: 'Ada', state: 'maybe' as never },
  ]), /unknown state/);
  assert.throws(() => readinessPresentation([], { empty: ' ' }), /emptyLabel/);
  assert.throws(() => readinessPresentation([], { progress: ' ' }), /progressLabel/);
  assert.throws(() => readinessPresentation(Array.from({ length: 257 }, (_, key) => ({
    key, label: `Player ${key}`, state: 'waiting' as const,
  }))), /maximum of 256/);
});
