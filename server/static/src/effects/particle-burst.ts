import { createSeededRandom, seedFromIdentity } from '../utils/seeded-random.ts';

export const MAX_BURST_PARTICLES = 24;

export interface BurstParticle {
  angleDeg: number;
  distancePx: number;
  delayMs: number;
  sizePx: number;
  rotationDeg: number;
}

export function burstParticles(
  requestedCount: number,
  spreadPx: number,
  seed: string | number,
): readonly BurstParticle[] {
  const finiteCount = Number.isFinite(requestedCount) ? requestedCount : 1;
  const finiteSpread = Number.isFinite(spreadPx) ? spreadPx : 8;
  const count = Math.max(1, Math.min(MAX_BURST_PARTICLES, Math.round(finiteCount)));
  const spread = Math.max(8, Math.min(240, finiteSpread));
  // The shared seeded-random primitive, composed the same way the dice path
  // composes it (identity -> uint32 -> avalanched mulberry32 stream). The
  // local copy this replaced truncated a numeric seed with `>>> 0` first, so
  // e.g. 3 and 3.7, or 1 and 2**32 + 1, played the identical burst.
  const random = createSeededRandom(seedFromIdentity(seed));
  const phase = random() * 360;

  return Object.freeze(Array.from({ length: count }, (_, index) => {
    // Even angular spacing prevents visibly empty wedges; seeded jitter keeps
    // the result lively while remaining stable in replays and screenshots.
    const angleStep = 360 / count;
    return Object.freeze({
      angleDeg: phase + index * angleStep + (random() - 0.5) * angleStep * 0.7,
      distancePx: spread * (0.55 + random() * 0.45),
      delayMs: Math.round(random() * 55),
      sizePx: Math.round(4 + random() * 5),
      rotationDeg: Math.round((random() - 0.5) * 240),
    });
  }));
}
