// Go's run-time panics, as the facade throws them. A translated function
// that would panic in Go throws an Error whose message is exactly Go's text,
// so a recorded panic (a legality ledger's unknown entry, say) compares
// equal. The texts are the Go runtime's own (runtime/error.go, panic.go),
// checked against recovered panics by testdata/gen/facadevec.
//
// See fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ".

/** A method call or field read through a nil pointer. */
export const NIL_DEREF_TEXT = 'runtime error: invalid memory address or nil pointer dereference';

/** Integer division or remainder by zero. */
export const DIVIDE_BY_ZERO_TEXT = 'runtime error: integer divide by zero';

/**
 * The start of an index panic. The whole text is
 * `runtime error: index out of range [i] with length n`, or, for a negative
 * index, `runtime error: index out of range [i]`.
 */
export const INDEX_OUT_OF_RANGE_PREFIX = 'runtime error: index out of range [';

/**
 * The start of a slice-expression panic:
 * `runtime error: slice bounds out of range [:j] with capacity n`, and the
 * other forms Go's runtime spells.
 */
export const SLICE_BOUNDS_PREFIX = 'runtime error: slice bounds out of range [';

/** An assignment to an entry of a nil map. */
export const NIL_MAP_WRITE_TEXT = 'assignment to entry in nil map';

/** `make([]T, n)` with n negative (runtime.panicmakeslicelen). */
export const MAKESLICE_LEN_TEXT = 'runtime error: makeslice: len out of range';

/** Go: the panic of `*p` or `p.M()` with p nil. Throws NIL_DEREF_TEXT. */
export function nilDeref(): never {
	throw new Error(NIL_DEREF_TEXT);
}

/** Go: the panic of `a / 0` or `a % 0`. Throws DIVIDE_BY_ZERO_TEXT. */
export function divideByZero(): never {
	throw new Error(DIVIDE_BY_ZERO_TEXT);
}

/** Go's text for indexing a slice of `length` at `index`. */
export function indexOutOfRangeText(index: number, length: number): string {
	if (index < 0) return INDEX_OUT_OF_RANGE_PREFIX + index + ']';
	return INDEX_OUT_OF_RANGE_PREFIX + index + '] with length ' + length;
}

/** Go: the panic of `xs[index]` out of range. Throws indexOutOfRangeText(index, length). */
export function indexOutOfRange(index: number, length: number): never {
	throw new Error(indexOutOfRangeText(index, length));
}

/**
 * Go's text for `xs[low:high]` out of range on a slice of capacity
 * `capacity`, or '' when the bounds are legal. The runtime checks the high
 * bound against the capacity first, then the low bound against the high
 * one, both as unsigned comparisons (so a negative bound fails its check);
 * a negative bound is reported without the other number.
 */
export function sliceBoundsText(low: number, high: number, capacity: number): string {
	if (high < 0) return SLICE_BOUNDS_PREFIX + ':' + high + ']';
	if (high > capacity) return SLICE_BOUNDS_PREFIX + ':' + high + '] with capacity ' + capacity;
	if (low < 0) return SLICE_BOUNDS_PREFIX + low + ':]';
	if (low > high) return SLICE_BOUNDS_PREFIX + low + ':' + high + ']';
	return '';
}

/** Go: the panic of `xs[low:high]` out of range. Throws sliceBoundsText(low, high, capacity). */
export function sliceBoundsOutOfRange(low: number, high: number, capacity: number): never {
	throw new Error(sliceBoundsText(low, high, capacity));
}
