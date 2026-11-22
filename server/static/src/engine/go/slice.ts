// Go's slices, nil-aware. A Go `[]T` is `T[] | null`: `var x []T` and
// `x = nil` are null (serialized as `null`), `make([]T, 0)` and `[]T{}` are
// `[]` (serialized as `[]`). `append` is allowed only as
// `x = append(x, ...)`, so a result never aliases another slice.
//
// Capacity is not modeled: a slice's capacity is its length. Go code whose
// result depends on spare capacity (re-slicing past len) is refused by the
// translator (slice-alias).
//
// See fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ".

import { MAKESLICE_LEN_TEXT, indexOutOfRange, sliceBoundsText } from './panic.ts';

/**
 * Go: `xs = append(xs, values...)`. `appendTo(null, v)` is `[v]`;
 * `appendTo(null)` (nothing appended) is null, as Go's is nil. Emitted as
 * `xs = appendTo(xs, ...)`. The result is a new array, so another variable
 * still holding the old slice keeps its old length, as in Go.
 */
export function appendTo<T>(xs: T[] | null, ...values: readonly T[]): T[] | null {
	if (values.length === 0) return xs;
	const out = xs === null ? [] : xs.slice();
	for (const value of values) out.push(value);
	return out;
}

/** Go: `len(xs)` on a slice. `lenOf(null)` is 0. */
export function lenOf(xs: readonly unknown[] | null): number {
	return xs === null ? 0 : xs.length;
}

/**
 * Go: `xs[i]` read with a non-constant index. Throws Go's index panic
 * (panic.ts indexOutOfRange) when i is outside 0 .. len(xs)-1; a null
 * slice has length 0.
 */
export function at<T>(xs: readonly T[] | null, i: number): T {
	const length = lenOf(xs);
	if (i < 0 || i >= length || xs === null) indexOutOfRange(i, length);
	return xs[i];
}

/** Go: `xs[i] = v` with a non-constant index. Throws Go's index panic as `at` does. */
export function setAt<T>(xs: T[] | null, i: number, value: T): void {
	const length = lenOf(xs);
	if (i < 0 || i >= length || xs === null) indexOutOfRange(i, length);
	xs[i] = value;
}

/**
 * Go: `make([]T, n)`: n copies of T's zero value. Never null. Throws Go's
 * `makeslice: len out of range` for a negative n.
 */
export function makeSlice<T>(n: number, zero: T): T[] {
	if (n < 0) throw new Error(MAKESLICE_LEN_TEXT);
	const out = new Array<T>(n);
	for (let i = 0; i < n; i++) out[i] = zero;
	return out;
}

/**
 * Go: `copy(dst, src)`: copies min(len(dst), len(src)) elements and returns
 * that count. dst and src may be the same array, as in Go.
 */
export function copyInto<T>(dst: T[] | null, src: readonly T[] | null): number {
	if (dst === null || src === null) return 0;
	const count = Math.min(dst.length, src.length);
	const from = src.slice(0, count);
	for (let i = 0; i < count; i++) dst[i] = from[i];
	return count;
}

/**
 * Go: `xs[low:high]`, read only (the translator refuses a slice expression
 * that is written or appended to, slice-alias). Throws Go's slice-bounds
 * panic when out of range. A nil slice sliced to [0:0] stays nil, as in Go.
 */
export function subslice<T>(xs: readonly T[] | null, low: number, high: number): readonly T[] | null {
	const text = sliceBoundsText(low, high, lenOf(xs));
	if (text !== '') throw new Error(text);
	if (xs === null) return null;
	return xs.slice(low, high);
}
