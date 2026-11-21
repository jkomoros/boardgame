// An exact port of Go's math/rand (v1): the additive lagged Fibonacci source
// in rng.go and the *Rand methods in rand.go that the boardgame core, its
// components and the example games call.
//
// Everything here must match Go bit for bit, because Go never records a
// draw: every shuffle and roll in a recorded game is recomputed from the
// game id, the secret salt and the state version. See stateRandSeed.

import { RNG_COOKED } from './rand-cooked.ts';
import { fnv1_64 } from './fnv.ts';
import { add64, int64FromNumber, mul32, negate64 } from './u64.ts';
import type { U64 } from './u64.ts';

const RNG_LEN = 607;
const RNG_TAP = 273;
const INT32_MAX = 2147483647;
const TWO32 = 4294967296;
const TWO63 = 9223372036854775808;

function seedrand(x: number): number {
	const hi = Math.trunc(x / 44488);
	const lo = x % 44488;
	x = 48271 * lo - 3399 * hi;
	if (x < 0) x += INT32_MAX;
	return x;
}

/** seed % (2^31 - 1) with Go's truncating sign rule, then Seed's fix-ups. */
function reduceSeed(seed: U64): number {
	const negative = seed.hi >= 0x80000000;
	const mag = negative ? negate64(seed) : seed;
	// 2^32 mod (2^31 - 1) is 2.
	let r = ((mag.hi % INT32_MAX) * 2 + (mag.lo % INT32_MAX)) % INT32_MAX;
	if (negative && r !== 0) r = INT32_MAX - r;
	if (r === 0) r = 89482311;
	return r;
}

/** Go's `*rand.Rand` over `rand.NewSource(seed)`. */
export class GoRand {
	private tap = 0;
	private feed = RNG_LEN - RNG_TAP;
	private readonly vecHi = new Array<number>(RNG_LEN);
	private readonly vecLo = new Array<number>(RNG_LEN);

	/**
	 * @param seed an int64, as a safe JavaScript integer or as a 64-bit
	 * two's-complement word.
	 */
	constructor(seed: number | U64) {
		this.seed(seed);
	}

	/** rand.Seed / rngSource.Seed. */
	seed(seed: number | U64): void {
		const word = typeof seed === 'number' ? int64FromNumber(seed) : seed;
		this.tap = 0;
		this.feed = RNG_LEN - RNG_TAP;
		let x = reduceSeed(word);
		for (let i = -20; i < RNG_LEN; i++) {
			x = seedrand(x);
			if (i >= 0) {
				// u = int64(x) << 40
				let hi = (x << 8) >>> 0;
				let lo = 0;
				x = seedrand(x);
				// u ^= int64(x) << 20
				hi = (hi ^ (x >>> 12)) >>> 0;
				lo = (lo ^ (x << 20)) >>> 0;
				x = seedrand(x);
				// u ^= int64(x)
				lo = (lo ^ x) >>> 0;
				this.vecHi[i] = (hi ^ RNG_COOKED[2 * i]) >>> 0;
				this.vecLo[i] = (lo ^ RNG_COOKED[2 * i + 1]) >>> 0;
			}
		}
	}

	/** rngSource.Uint64: the raw 64-bit output. */
	uint64(): U64 {
		this.tap--;
		if (this.tap < 0) this.tap += RNG_LEN;
		this.feed--;
		if (this.feed < 0) this.feed += RNG_LEN;
		const x = add64(
			{ hi: this.vecHi[this.feed], lo: this.vecLo[this.feed] },
			{ hi: this.vecHi[this.tap], lo: this.vecLo[this.tap] },
		);
		this.vecHi[this.feed] = x.hi;
		this.vecLo[this.feed] = x.lo;
		return x;
	}

	/** Int63, as a 64-bit word whose top bit is clear. */
	int63(): U64 {
		const x = this.uint64();
		return { hi: x.hi & 0x7fffffff, lo: x.lo };
	}

	/** Uint32: uint32(Int63() >> 31). */
	uint32(): number {
		const x = this.int63();
		return ((x.hi << 1) | (x.lo >>> 31)) >>> 0;
	}

	/** Int31: int32(Int63() >> 32). */
	int31(): number {
		return this.int63().hi;
	}

	/** Int31n. `n` must be in [1, 2^31 - 1]. */
	int31n(n: number): number {
		if (!(n > 0) || n > INT32_MAX || Math.floor(n) !== n) {
			throw new Error('invalid argument to Int31n');
		}
		if ((n & (n - 1)) === 0) {
			return this.int31() & (n - 1);
		}
		const max = INT32_MAX - (2147483648 % n);
		let v = this.int31();
		while (v > max) v = this.int31();
		return v % n;
	}

	/**
	 * Int63n. Go accepts any positive int64; this port accepts n up to 2^52,
	 * which keeps every intermediate exact without BigInt, and throws beyond
	 * that. The result is returned as a JavaScript number.
	 */
	int63n(n: number): number {
		if (!(n > 0) || Math.floor(n) !== n) {
			throw new Error('invalid argument to Int63n');
		}
		if (n > 4503599627370496) {
			throw new Error('Int63n: n above 2^52 is not supported by this port');
		}
		const nHi = Math.floor(n / TWO32);
		const nLo = n >>> 0;
		const m = n - 1;
		const mHi = Math.floor(m / TWO32);
		const mLo = m >>> 0;
		if ((nHi & mHi) === 0 && (nLo & mLo) === 0) {
			// n is a power of two: Int63() & (n - 1).
			const v = this.int63();
			return (v.hi & mHi) * TWO32 + ((v.lo & mLo) >>> 0);
		}
		// max = 2^63 - 1 - (2^63 % n). v > max exactly when
		// (2^63 - 1 - v) < (2^63 % n).
		const rem = mod63(0x80000000, 0, n);
		let v = this.int63();
		for (;;) {
			const wHi = 0x7fffffff - v.hi;
			const wLo = 0xffffffff - v.lo;
			// rem < n <= 2^52, so w can only be below it when wHi < 2^20.
			if (wHi < 1048576 && wHi * TWO32 + wLo < rem) {
				v = this.int63();
				continue;
			}
			break;
		}
		return mod63(v.hi, v.lo, n);
	}

	/** Intn. Supports n up to 2^52 (see int63n). */
	intn(n: number): number {
		if (!(n > 0) || Math.floor(n) !== n) {
			throw new Error('invalid argument to Intn');
		}
		if (n <= INT32_MAX) return this.int31n(n);
		return this.int63n(n);
	}

	/** Float64, in [0, 1). */
	float64(): number {
		for (;;) {
			const v = this.int63();
			// float64(int63): one correctly rounded conversion. hi * 2^32 is
			// exact, so the single rounding happens in the addition.
			const f = (v.hi * TWO32 + v.lo) / TWO63;
			if (f === 1) continue;
			return f;
		}
	}

	/** Float32, in [0, 1), returned as the float32 value widened to a number. */
	float32(): number {
		for (;;) {
			const f = Math.fround(this.float64());
			if (f === 1) continue;
			return f;
		}
	}

	/** Perm: a pseudo-random permutation of [0, n). */
	perm(n: number): number[] {
		const m = new Array<number>(n);
		for (let i = 0; i < n; i++) m[i] = 0;
		for (let i = 0; i < n; i++) {
			const j = this.intn(i + 1);
			m[i] = m[j];
			m[j] = i;
		}
		return m;
	}

	/** Shuffle: Fisher-Yates with Go's unexported int31n (Lemire) draws. */
	shuffle(n: number, swap: (i: number, j: number) => void): void {
		if (n < 0) throw new Error('invalid argument to Shuffle');
		let i = n - 1;
		for (; i > INT32_MAX - 1; i--) {
			swap(i, this.int63n(i + 1));
		}
		for (; i > 0; i--) {
			swap(i, this.lemire31n(i + 1));
		}
	}

	/** rand.go's unexported int31n. */
	private lemire31n(n: number): number {
		let v = this.uint32();
		let prod = mul32(v, n);
		if (prod.lo < n) {
			const thresh = (TWO32 - n) % n;
			while (prod.lo < thresh) {
				v = this.uint32();
				prod = mul32(v, n);
			}
		}
		return prod.hi;
	}
}

/** (hi * 2^32 + lo) % n for a 63-bit value and 0 < n <= 2^52. */
function mod63(hi: number, lo: number, n: number): number {
	let r = 0;
	for (let bit = 31; bit >= 0; bit--) {
		r = (r * 2 + ((hi >>> bit) & 1)) % n;
	}
	for (let bit = 31; bit >= 0; bit--) {
		r = (r * 2 + ((lo >>> bit) & 1)) % n;
	}
	return r;
}

/**
 * The string the Go core hashes to seed state.Rand(): game id, secret salt
 * and the decimal state version, concatenated (state.go, `(*state).Rand`).
 */
export function stateRandSeedInput(gameId: string, secretSalt: string, version: number): string {
	return gameId + secretSalt + String(version);
}

/** int64(fnv.New64().Sum64()) of the seed input, as a 64-bit word. */
export function stateRandSeed(gameId: string, secretSalt: string, version: number): U64 {
	return fnv1_64(stateRandSeedInput(gameId, secretSalt, version));
}

/**
 * The generator state.Rand() returns for a state at `version`. The Go core
 * memoizes it per state, so successive draws while applying one move continue
 * one sequence; callers must do the same.
 */
export function stateRand(gameId: string, secretSalt: string, version: number): GoRand {
	return new GoRand(stateRandSeed(gameId, secretSalt, version));
}
