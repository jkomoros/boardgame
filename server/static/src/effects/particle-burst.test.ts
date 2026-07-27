import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { burstParticles, MAX_BURST_PARTICLES } from './particle-burst.ts';

describe('burstParticles', () => {
  it('is deterministic for a seed and varies across seeds', () => {
    assert.deepEqual(burstParticles(12, 80, 'game-7-v4'), burstParticles(12, 80, 'game-7-v4'));
    assert.notDeepEqual(burstParticles(12, 80, 'game-7-v4'), burstParticles(12, 80, 'game-7-v5'));
  });

  it('clamps particle count and unsafe spread values', () => {
    const particles = burstParticles(10_000, 10_000, 1);
    assert.equal(particles.length, MAX_BURST_PARTICLES);
    assert.ok(particles.every(particle => particle.distancePx >= 132 && particle.distancePx <= 240));
  });

  it('always returns at least one bounded particle', () => {
    const [particle] = burstParticles(-3, -10, 2);
    assert.ok(particle);
    assert.ok(particle.distancePx >= 4.4 && particle.distancePx <= 8);
    assert.ok(particle.sizePx >= 4 && particle.sizePx <= 9);
  });

  // The seed used to be truncated with `value >>> 0` (a ToInt32/ToUint32
  // conversion) before it reached the PRNG, so whole FAMILIES of distinct
  // seeds aliased onto one stream: a fraction lost its fractional part, a
  // value past 2^32 wrapped, and a negative wrapped onto its unsigned twin.
  // Two bursts that should look different played the identical burst, with
  // no signal anywhere. Same bug class the dice seed hash fixed.
  it('does not alias distinct numeric seeds onto one burst', () => {
    const burst = (seed: number) => burstParticles(6, 60, seed);
    // fraction dropped by truncation
    assert.notDeepEqual(burst(3), burst(3.7));
    // wrapped past 2^32
    assert.notDeepEqual(burst(1), burst(2 ** 32 + 1));
    // negative wrapped onto its unsigned twin
    assert.notDeepEqual(burst(-1), burst(2 ** 32 - 1));
    // a whole aliasing family, all of which truncated to 7
    assert.notDeepEqual(burst(7), burst(7.5));
    assert.notDeepEqual(burst(7), burst(2 ** 32 + 7));
  });

  it('normalizes non-finite public inputs', () => {
    const particles = burstParticles(Number.NaN, Number.POSITIVE_INFINITY, 3);
    assert.equal(particles.length, 1);
    assert.ok(Number.isFinite(particles[0].distancePx));
  });
});
