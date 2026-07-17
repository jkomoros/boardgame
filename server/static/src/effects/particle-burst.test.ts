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

  it('normalizes non-finite public inputs', () => {
    const particles = burstParticles(Number.NaN, Number.POSITIVE_INFINITY, 3);
    assert.equal(particles.length, 1);
    assert.ok(Number.isFinite(particles[0].distancePx));
  });
});
