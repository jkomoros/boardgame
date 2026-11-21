// Known-answer tests for the math/rand and FNV ports. Every expected value in
// testdata/rand.vectors.json was produced by running real Go
// (testdata/gen/randvec); nothing here is derived from reading the Go source.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { newGoRand } from './rand.ts';
import type { GoRand } from './rand.ts';
import { fnv1_64 } from './fnv.ts';
import { i64ToDecimal, u64FromDecimal, u64ToDecimal, int64FromNumber, utf8Bytes } from './u64.ts';

interface SeedVec {
	Seed: string;
	Uint64: string[];
	Int63: string[];
	Uint32: number[];
	Int31: number[];
	Int31n: Record<string, number[]>;
	Int63n: Record<string, string[]>;
	Intn: Record<string, string[]>;
	Float64: string[];
	Float32: string[];
	Perm: Record<string, number[]>;
	Shuffle: Record<string, number[]>;
	Mixed: string[];
}
interface FnvVec { Input: string; Sum64: string; Seed: string; Perm5: number[] }

const vectors = JSON.parse(readFileSync(new URL('./testdata/rand.vectors.json', import.meta.url), 'utf8')) as {
	Seeds: SeedVec[];
	FNV: FnvVec[];
};

const fresh = (v: SeedVec) => newGoRand(u64FromDecimal(v.Seed));

test('u64 decimal round trip agrees with BigInt', () => {
	for (const text of ['0', '1', '-1', '18446744073709551615', '9223372036854775807', '-9223372036854775808', '1234567890123456789']) {
		const word = u64FromDecimal(text);
		const expected = BigInt.asUintN(64, BigInt(text));
		assert.equal((BigInt(word.hi) << 32n) | BigInt(word.lo), expected, text);
		assert.equal(u64ToDecimal(word), expected.toString());
		assert.equal(i64ToDecimal(word), BigInt.asIntN(64, expected).toString());
	}
	assert.deepEqual(int64FromNumber(-1), { hi: 0xffffffff, lo: 0xffffffff });
	assert.deepEqual(int64FromNumber(2 ** 40 + 5), { hi: 256, lo: 5 });
});

test('utf8Bytes matches TextEncoder', () => {
	for (const v of vectors.FNV) {
		assert.deepEqual(utf8Bytes(v.Input), Array.from(new TextEncoder().encode(v.Input)), JSON.stringify(v.Input));
	}
});

test('vector file covers the expected seeds', () => {
	assert.equal(vectors.Seeds.length, 60);
	assert.ok(vectors.FNV.length >= 79);
});

test('Uint64 stream matches Go for every seed, past two turns of the register', () => {
	for (const v of vectors.Seeds) {
		const r = fresh(v);
		for (let i = 0; i < v.Uint64.length; i++) {
			assert.equal(u64ToDecimal(r.uint64()), v.Uint64[i], `seed ${v.Seed} draw ${i}`);
		}
	}
});

test('numeric and decomposed seeds agree', () => {
	for (const seed of [0, 1, -1, 42, 2147483647, 2147483648, -2147483647, 2 ** 40 + 12345]) {
		const v = vectors.Seeds.find(s => s.Seed === String(seed));
		assert.ok(v, `vector for seed ${seed}`);
		const r = newGoRand(seed);
		assert.equal(u64ToDecimal(r.uint64()), v.Uint64[0]);
	}
});

test('Int63, Uint32, Int31 match Go', () => {
	for (const v of vectors.Seeds) {
		let r = fresh(v);
		for (const want of v.Int63) assert.equal(i64ToDecimal(r.int63()), want, `Int63 seed ${v.Seed}`);
		r = fresh(v);
		for (const want of v.Uint32) assert.equal(r.uint32(), want, `Uint32 seed ${v.Seed}`);
		r = fresh(v);
		for (const want of v.Int31) assert.equal(r.int31(), want, `Int31 seed ${v.Seed}`);
	}
});

test('Int31n matches Go, including power-of-two and rejection paths', () => {
	for (const v of vectors.Seeds) {
		for (const [n, wants] of Object.entries(v.Int31n)) {
			const r = fresh(v);
			for (const want of wants) assert.equal(r.int31n(Number(n)), want, `Int31n(${n}) seed ${v.Seed}`);
		}
	}
});

test('Int63n matches Go up to 2^52', () => {
	for (const v of vectors.Seeds) {
		for (const [n, wants] of Object.entries(v.Int63n)) {
			const r = fresh(v);
			for (const want of wants) assert.equal(String(r.int63n(Number(n))), want, `Int63n(${n}) seed ${v.Seed}`);
		}
	}
});

test('Intn matches Go on both sides of 2^31', () => {
	for (const v of vectors.Seeds) {
		for (const [n, wants] of Object.entries(v.Intn)) {
			const r = fresh(v);
			for (const want of wants) assert.equal(String(r.intn(Number(n))), want, `Intn(${n}) seed ${v.Seed}`);
		}
	}
});

test('Float64 and Float32 match Go exactly', () => {
	for (const v of vectors.Seeds) {
		let r = fresh(v);
		for (const want of v.Float64) assert.equal(r.float64(), Number(want), `Float64 seed ${v.Seed}`);
		r = fresh(v);
		for (const want of v.Float32) assert.equal(r.float32(), Number(want), `Float32 seed ${v.Seed}`);
	}
});

test('Perm matches Go for every size', () => {
	let checked = 0;
	for (const v of vectors.Seeds) {
		for (const [n, want] of Object.entries(v.Perm)) {
			assert.deepEqual(fresh(v).perm(Number(n)), want ?? [], `Perm(${n}) seed ${v.Seed}`);
			checked++;
		}
	}
	assert.ok(checked > 500);
});

test('Shuffle matches Go for every size', () => {
	for (const v of vectors.Seeds) {
		for (const [n, want] of Object.entries(v.Shuffle)) {
			const arr = Array.from({ length: Number(n) }, (_, i) => i);
			fresh(v).shuffle(arr.length, (i, j) => { const t = arr[i]; arr[i] = arr[j]; arr[j] = t; });
			assert.deepEqual(arr, want ?? [], `Shuffle(${n}) seed ${v.Seed}`);
		}
	}
});

test('interleaved calls on one generator continue one sequence', () => {
	for (const v of vectors.Seeds) {
		const r = fresh(v);
		let k = 0;
		for (let i = 0; i < 6; i++) {
			assert.equal('perm:' + JSON.stringify(r.perm(5 + i)), v.Mixed[k++]);
			assert.equal('intn6:' + r.intn(6), v.Mixed[k++]);
			assert.equal(r.float64(), Number(v.Mixed[k++].slice('float64:'.length)));
			assert.equal('int63:' + i64ToDecimal(r.int63()), v.Mixed[k++]);
		}
	}
});

test('invalid arguments throw, as Go panics', () => {
	const r = newGoRand(1);
	assert.throws(() => r.int31n(0));
	assert.throws(() => r.intn(0));
	assert.throws(() => r.intn(-3));
	assert.throws(() => r.int63n(0));
	assert.throws(() => r.int63n(2 ** 52 + 2), /not supported/);
	assert.throws(() => r.shuffle(-1, () => undefined));
});

test('FNV-1 64 matches hash/fnv New64, and seeds the generator as state.Rand does', () => {
	for (const v of vectors.FNV) {
		const sum = fnv1_64(v.Input);
		assert.equal(u64ToDecimal(sum), v.Sum64, JSON.stringify(v.Input));
		assert.equal(i64ToDecimal(sum), v.Seed, JSON.stringify(v.Input));
		assert.deepEqual(newGoRand(sum).perm(5), v.Perm5, JSON.stringify(v.Input));
	}
});
