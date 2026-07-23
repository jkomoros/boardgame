import assert from 'node:assert/strict';
import test from 'node:test';
import { compileLegacyAnimationOverlap, hasLegacyAnimationOverlap } from './legacy-overlap.ts';

type Move = Readonly<{ name: string }>;
const defaultHook = (_from: Move | null, _to: Move | null) => 0;

test('legacy overlap preserves successor-aware fraction timing', () => {
  const from = Object.freeze({ name: 'reveal' });
  const to = Object.freeze({ name: 'deal' });
  let received: readonly [Move | null, Move | null] | null = null;
  const renderer = {
    animationOverlap(actualFrom: Move | null, actualTo: Move | null) {
      received = [actualFrom, actualTo];
      return actualTo?.name === 'deal' ? 0.3 : 0;
    },
  };
  assert.deepEqual(
    compileLegacyAnimationOverlap(renderer, defaultHook, from, to, 500),
    { configured: true, delayMs: 150 },
  );
  assert.deepEqual(received, [from, to]);
  assert.deepEqual(
    compileLegacyAnimationOverlap(renderer, defaultHook, from, { name: 'score' }, 500),
    { configured: true, delayMs: null },
  );
});

test('legacy overlap retains clamping and default-hook detection', () => {
  assert.equal(hasLegacyAnimationOverlap({ animationOverlap: defaultHook }, defaultHook), false);
  assert.equal(hasLegacyAnimationOverlap({ animationOverlap: () => 0 }, defaultHook), true);
  assert.deepEqual(
    compileLegacyAnimationOverlap({ animationOverlap: defaultHook }, defaultHook, null, null, 500),
    { configured: false, delayMs: null },
  );
  for (const fraction of [Number.NEGATIVE_INFINITY, -1, 0, 1, 2, Number.POSITIVE_INFINITY, Number.NaN]) {
    assert.deepEqual(
      compileLegacyAnimationOverlap(
        { animationOverlap: () => fraction }, defaultHook, null, null, 500,
      ),
      { configured: true, delayMs: null },
    );
  }
  assert.deepEqual(
    compileLegacyAnimationOverlap(
      { animationOverlap: () => 0.25 }, defaultHook, null, null, 800,
    ),
    { configured: true, delayMs: 200 },
  );
});
