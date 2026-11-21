/**
 * The one seeded-random primitive behind every deterministic visual.
 *
 * A deterministic visual (a dice tumble, a particle burst) is built the same
 * way everywhere in this codebase: turn a stable IDENTITY into a uint32, then
 * draw every quantity from a reproducible stream seeded with it. That recipe
 * had been written out three times -- `dice-sim.ts`'s `createRandom`,
 * `boardgame-die.ts`'s `dieRollSeed`, and `particle-burst.ts`'s own
 * `hashSeed`/`seededRandom` -- and the copies had already drifted apart on the
 * one detail that matters: whether the seed survives the trip intact.
 *
 * Nothing here touches the DOM or the clock. Same arguments, same stream,
 * forever, on any conforming engine -- which is the entire reason for having a
 * PRNG here instead of `Math.random`.
 */

/**
 * splitmix32's finaliser: the avalanche step, on its own. Every output bit
 * depends on every input bit, so two seeds that differ in one bit produce
 * unrelated streams.
 */
export function mix32(value: number): number {
  let x = value >>> 0;
  x = Math.imul(x ^ (x >>> 16), 0x21f0aaad) >>> 0;
  x = Math.imul(x ^ (x >>> 15), 0x735a2d97) >>> 0;
  return (x ^ (x >>> 15)) >>> 0;
}

/**
 * FNV-1a over the string's UTF-16 code units. The point is only that distinct
 * identities land on distinct uint32s without structure; `hashSeedBits` below
 * then spreads that uint32 across the stream.
 *
 * (`0x811c9dc5`/`0x01000193` are the same constants `particle-burst.ts` spelled
 * in decimal as `2166136261`/`16777619`.)
 */
export function fnv1a32(text: string): number {
  let hash = 0x811c9dc5;
  for (let index = 0; index < text.length; index++) {
    hash ^= text.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return hash >>> 0;
}

/**
 * Scratch for reading a double's bits. A `DataView` rather than a
 * `Uint32Array` over a `Float64Array` because the latter is host-endian and
 * this module's whole contract is that a seed reproduces on any engine;
 * `getUint32(_, true)` pins little-endian regardless of the machine.
 */
const SEED_BITS = new DataView(new ArrayBuffer(8));

/**
 * ALL 64 bits of a seed, avalanched down to one uint32 of PRNG state.
 *
 * Reading all 64 bits rather than an int32 of the seed is the load-bearing
 * part. Every truncating variant of this -- `Math.trunc(seed) ^ ...`, or
 * `seed >>> 0` -- aliases whole FAMILIES of distinct seeds onto one stream:
 * `3` and `3.7` collapse together, so do `1` and `2**32 + 1`, and so do `-1`
 * and `2**32 - 1`. Callers derive seeds from things like (component id, state
 * version) or a reserved-particle count, and an aliased seed replays one
 * visual for a different event with no signal anywhere.
 *
 * `-0` is folded onto `0` (`seed + 0`) because the two are `===` and callers
 * have no way to tell which one they produced.
 *
 * What is NOT promised: mulberry32's state is 32 bits, so there are only 2^32
 * possible streams and distinct seeds must sometimes collide. The guarantee is
 * that collisions are unstructured -- a pseudorandom ~2^-32 per pair -- rather
 * than the systematic families truncation created.
 */
export function hashSeedBits(seed: number): number {
  SEED_BITS.setFloat64(0, seed + 0, true);
  return mix32(mix32(SEED_BITS.getUint32(0, true) ^ 0x9e3779b9) ^ SEED_BITS.getUint32(4, true));
}

/**
 * A seed from whatever stable identity a caller has: a string identity is
 * hashed, a number is passed through UNTRUNCATED (`hashSeedBits` reads all of
 * it). Non-finite numbers go through their string form instead, because a
 * NaN's bit pattern is not portable and `Infinity` has no distinguishable
 * halves worth avalanching -- neither would honour the cross-engine promise.
 */
export function seedFromIdentity(identity: string | number): number {
  if (typeof identity === 'number' && Number.isFinite(identity)) return identity;
  return fnv1a32(String(identity));
}

/**
 * mulberry32, seeded through `hashSeedBits`. Chosen because its whole state is
 * one uint32 and every operation is integer, so it reproduces exactly on any
 * conforming engine.
 *
 * The seed is avalanched rather than used raw because consecutive seeds
 * otherwise differ by one bit of state and mulberry32's FIRST few outputs stay
 * correlated -- which means seed 1 and seed 2 throw their dice in nearly the
 * same direction, or burst their particles at nearly the same angle. Callers
 * here draw every seed-dependent quantity in the first dozen outputs, so that
 * is exactly the regime that matters.
 */
export function createSeededRandom(seed: number): () => number {
  let state = hashSeedBits(seed);
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let t = state;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
