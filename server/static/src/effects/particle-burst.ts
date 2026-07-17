export const MAX_BURST_PARTICLES = 24;

export interface BurstParticle {
  angleDeg: number;
  distancePx: number;
  delayMs: number;
  sizePx: number;
  rotationDeg: number;
}

function hashSeed(value: string | number): number {
  if (typeof value === 'number' && Number.isFinite(value)) return value >>> 0;
  const text = String(value);
  let hash = 2166136261;
  for (let index = 0; index < text.length; index++) {
    hash ^= text.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return hash >>> 0;
}

function seededRandom(seed: number): () => number {
  let state = seed || 0x6d2b79f5;
  return () => {
    state += 0x6d2b79f5;
    let value = state;
    value = Math.imul(value ^ (value >>> 15), value | 1);
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
  };
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
  const random = seededRandom(hashSeed(seed));
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
