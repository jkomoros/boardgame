import assert from 'node:assert/strict';
import test from 'node:test';
import {
  animationReachedMotionProgress,
  compileMotionRelease,
  selectMotionReleaseParticipants,
} from './release.ts';

test('motion release compilation is strict, normalized, and immutable', () => {
  const compiled = compileMotionRelease({
    key: ' deal ', progress: 0.3, subjects: [' a ', 'b'],
  });
  assert.deepEqual(compiled, { key: 'deal', progress: 0.3, subjects: ['a', 'b'] });
  assert.ok(Object.isFrozen(compiled));
  assert.ok(Object.isFrozen(compiled.subjects));
  for (const declaration of [
    { progress: 0 }, { progress: 1 }, { progress: Number.NaN },
    { progress: 0.3, subjects: [] },
    { progress: 0.3, subjects: ['a', ' a '] },
  ]) assert.throws(() => compileMotionRelease(declaration));
});

test('subject selection fails closed on missing or ambiguous primaries', () => {
  const animation = {} as Animation;
  const participants = [
    { subjectId: 'a', animation },
    { subjectId: 'b', animation },
    { subjectId: 'b', animation },
  ];
  assert.deepEqual(
    selectMotionReleaseParticipants(compileMotionRelease({ progress: 0.2 }), participants),
    participants,
  );
  assert.equal(selectMotionReleaseParticipants(
    compileMotionRelease({ progress: 0.2, subjects: ['missing'] }), participants,
  ), null);
  assert.equal(selectMotionReleaseParticipants(
    compileMotionRelease({ progress: 0.2, subjects: ['b'] }), participants,
  ), null);
  assert.deepEqual(selectMotionReleaseParticipants(
    compileMotionRelease({ progress: 0.2, subjects: ['a'] }), participants,
  ), [participants[0]]);
});

test('motion progress uses computed active duration and excludes end delay', () => {
  const animation = {
    currentTime: 699.4,
    playState: 'running',
    effect: { getComputedTiming: () => ({ delay: 100, activeDuration: 2000, endDelay: 9000 }) },
  } as unknown as Animation;
  assert.equal(animationReachedMotionProgress(animation, 0.3), false);
  (animation as unknown as { currentTime: number }).currentTime = 700;
  assert.equal(animationReachedMotionProgress(animation, 0.3), true);
  (animation as unknown as { playState: AnimationPlayState }).playState = 'idle';
  assert.equal(animationReachedMotionProgress(animation, 0.3), false);
  (animation as unknown as { playState: AnimationPlayState }).playState = 'finished';
  assert.equal(animationReachedMotionProgress(animation, 0.3), true);
});
