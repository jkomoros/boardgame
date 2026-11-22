// Go's slices, nil-aware. A Go `[]T` is `T[] | null`: `var x []T` and
// `x = nil` are null (serialized as `null`), `make([]T, 0)` and `[]T{}` are
// `[]` (serialized as `[]`). `append` is allowed only as
// `x = append(x, ...)`, so a result never aliases another slice.
//
// See fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ".

/**
 * Go: `xs = append(xs, values...)`. `appendTo(null, v)` is `[v]`;
 * `appendTo(null)` (nothing appended) is null, as Go's is nil. Emitted as
 * `xs = appendTo(xs, ...)`.
 */
export function appendTo<T>(xs: T[] | null, ...values: readonly T[]): T[] | null {
	void xs;
	void values;
	throw new Error('go/slice: appendTo not implemented');
}

/** Go: `len(xs)` on a slice. `lenOf(null)` is 0. */
export function lenOf(xs: readonly unknown[] | null): number {
	void xs;
	throw new Error('go/slice: lenOf not implemented');
}

/**
 * Go: `xs[i]` read with a non-constant index. Throws Go's index panic
 * (panic.ts indexOutOfRange) when i is outside 0 .. len(xs)-1; a null
 * slice has length 0.
 */
export function at<T>(xs: readonly T[] | null, i: number): T {
	void xs;
	void i;
	throw new Error('go/slice: at not implemented');
}

/** Go: `xs[i] = v` with a non-constant index. Throws Go's index panic as `at` does. */
export function setAt<T>(xs: T[] | null, i: number, value: T): void {
	void xs;
	void i;
	void value;
	throw new Error('go/slice: setAt not implemented');
}

/** Go: `make([]T, n)`: n copies of T's zero value. Never null. */
export function makeSlice<T>(n: number, zero: T): T[] {
	void n;
	void zero;
	throw new Error('go/slice: makeSlice not implemented');
}

/** Go: `copy(dst, src)`: copies min(len(dst), len(src)) elements and returns that count. */
export function copyInto<T>(dst: T[] | null, src: readonly T[] | null): number {
	void dst;
	void src;
	throw new Error('go/slice: copyInto not implemented');
}

/**
 * Go: `xs[low:high]`, read only (the translator refuses a slice expression
 * that is written or appended to, slice-alias). Throws Go's slice-bounds
 * panic when out of range.
 */
export function subslice<T>(xs: readonly T[] | null, low: number, high: number): readonly T[] | null {
	void xs;
	void low;
	void high;
	throw new Error('go/slice: subslice not implemented');
}
