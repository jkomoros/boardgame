// Go's integer arithmetic where JavaScript's differs. `int` is a number;
// see fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ". `+`, `-`,
// comparisons and `%` (which takes the dividend's sign in both languages)
// are emitted as themselves; these are the operations that are not.

/**
 * Go: `a / b` on ints. Truncates toward zero (Math.trunc(a / b)); throws
 * DIVIDE_BY_ZERO_TEXT (panic.ts) when b is 0.
 */
export function idiv(a: number, b: number): number {
	void a;
	void b;
	throw new Error('go/num: idiv not implemented');
}

/** Go: `a % b` on ints. The same as JavaScript's `%`, plus the zero check of idiv. */
export function imod(a: number, b: number): number {
	void a;
	void b;
	throw new Error('go/num: imod not implemented');
}

/**
 * Go: `a * b` on ints. Throws at once when the product is not a safe
 * integer (Number.isSafeInteger): 64-bit wraparound is not emulated, and a
 * game that relies on it is out of scope.
 */
export function imul(a: number, b: number): number {
	void a;
	void b;
	throw new Error('go/num: imul not implemented');
}
