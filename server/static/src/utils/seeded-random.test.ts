import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  createSeededRandom,
  fnv1a32,
  hashSeedBits,
  mix32,
  seedFromIdentity,
} from './seeded-random.ts';

const draw = (seed: number, count = 8): number[] => {
  const random = createSeededRandom(seed);
  return Array.from({ length: count }, () => random());
};

describe('seeded-random', () => {
  it('produces the same stream for the same seed and stays in [0, 1)', () => {
    assert.deepEqual(draw(12345), draw(12345));
    assert.ok(draw(9, 256).every(value => value >= 0 && value < 1));
  });

  // The reason this module exists. A truncating seed hash (`seed >>> 0`, or
  // `Math.trunc`) aliases whole families of distinct seeds onto one stream;
  // these are the exact pairs the truncating copy in particle-burst.ts
  // collided, verified against it before the migration:
  //   hashSeed(3)  === hashSeed(3.7)         === 3
  //   hashSeed(1)  === hashSeed(2**32 + 1)   === 1
  //   hashSeed(-1) === hashSeed(2**32 - 1)   === 4294967295
  //   hashSeed(7)  === hashSeed(7.5) === hashSeed(2**32 + 7) === 7
  it('does not alias the seed families a ToUint32 truncation collapsed', () => {
    const collidedUnderTruncation: readonly (readonly [number, number])[] = [
      [3, 3.7],
      [1, 2 ** 32 + 1],
      [-1, 2 ** 32 - 1],
      [7, 7.5],
      [7, 2 ** 32 + 7],
      [0, -0.9],
    ];
    for (const [left, right] of collidedUnderTruncation) {
      assert.equal(left >>> 0, right >>> 0, `${left} and ${right} must truncate alike`);
      assert.notEqual(
        hashSeedBits(left),
        hashSeedBits(right),
        `${left} and ${right} must not share PRNG state`,
      );
      assert.notDeepEqual(draw(left), draw(right), `${left} and ${right} must not share a stream`);
    }
  });

  it('reads all 64 bits of the seed, and folds -0 onto 0', () => {
    // Two doubles that differ only in the low half, i.e. only in bits a
    // 32-bit read would never see.
    assert.notEqual(hashSeedBits(1.0000000000000002), hashSeedBits(1));
    assert.equal(hashSeedBits(-0), hashSeedBits(0));
    assert.deepEqual(draw(-0), draw(0));
  });

  it('decorrelates adjacent seeds in the first outputs', () => {
    // Every caller draws its seed-dependent quantities from the first dozen
    // outputs, so this -- not asymptotic quality -- is the regime that
    // matters. Un-avalanched mulberry32 keeps adjacent seeds visibly close.
    for (let seed = 0; seed < 64; seed++) {
      const [a] = draw(seed, 1);
      const [b] = draw(seed + 1, 1);
      assert.ok(Math.abs(a - b) > 0.001, `seeds ${seed}/${seed + 1} start too close: ${a} vs ${b}`);
    }
  });

  it('hashes string identities to distinct, order-sensitive uint32s', () => {
    assert.equal(fnv1a32('game-7-v4'), fnv1a32('game-7-v4'));
    assert.notEqual(fnv1a32('game-7-v4'), fnv1a32('game-7-v5'));
    assert.notEqual(fnv1a32('ab'), fnv1a32('ba'));
    assert.equal(fnv1a32(''), 0x811c9dc5);
    assert.ok(Number.isInteger(fnv1a32('x')) && fnv1a32('x') >= 0 && fnv1a32('x') <= 0xffffffff);
  });

  it('takes a finite number identity untruncated and a string identity hashed', () => {
    assert.equal(seedFromIdentity(3.7), 3.7);
    assert.equal(seedFromIdentity(2 ** 32 + 1), 2 ** 32 + 1);
    assert.equal(seedFromIdentity('roll#4'), fnv1a32('roll#4'));
    // Non-finite numbers have no portable bit pattern, so they route through
    // their string form rather than through the double reader.
    assert.equal(seedFromIdentity(Number.NaN), fnv1a32('NaN'));
    assert.equal(seedFromIdentity(Number.POSITIVE_INFINITY), fnv1a32('Infinity'));
    assert.notEqual(seedFromIdentity(Number.POSITIVE_INFINITY), seedFromIdentity(Number.NEGATIVE_INFINITY));
  });

  it('avalanches: a one-bit input change flips about half the output bits', () => {
    let total = 0;
    for (let bit = 0; bit < 32; bit++) {
      const flipped = mix32(0x12345678 ^ (1 << bit)) ^ mix32(0x12345678);
      let ones = 0;
      for (let index = 0; index < 32; index++) ones += (flipped >>> index) & 1;
      total += ones;
    }
    const average = total / 32;
    assert.ok(average > 12 && average < 20, `expected ~16 flipped bits, saw ${average}`);
  });
});
