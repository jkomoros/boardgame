// Go's integer arithmetic where JavaScript's differs. `int` is a number;
// see fabric/TRANSPILE-PROPOSAL.md, "Semantics that differ". `+`, `-`,
// comparisons and `%` (which takes the dividend's sign in both languages)
// are emitted as themselves; these are the operations that are not.
//
// Every result is normalized so that JavaScript's -0 never escapes: Go has
// no negative zero, and -0 would serialize as 0 but compare unequal under
// Object.is.

import { divideByZero } from './panic.ts';

/** The text imul throws when a product leaves the safe-integer range. */
export const INT_OVERFLOW_PREFIX = 'go/num: integer overflow: ';

/**
 * Go: `a / b` on ints. Truncates toward zero (Math.trunc(a / b)); throws
 * DIVIDE_BY_ZERO_TEXT (panic.ts) when b is 0.
 */
export function idiv(a: number, b: number): number {
	if (b === 0) divideByZero();
	return Math.trunc(a / b) + 0;
}

/** Go: `a % b` on ints. The same as JavaScript's `%`, plus the zero check of idiv. */
export function imod(a: number, b: number): number {
	if (b === 0) divideByZero();
	return (a % b) + 0;
}

/**
 * Go: `a * b` on ints. Throws at once when the product is not a safe
 * integer (Number.isSafeInteger): 64-bit wraparound is not emulated, and a
 * game that relies on it is out of scope.
 */
export function imul(a: number, b: number): number {
	const product = a * b + 0;
	if (!Number.isSafeInteger(product)) throw new Error(INT_OVERFLOW_PREFIX + a + ' * ' + b + ' is not a safe integer');
	return product;
}
