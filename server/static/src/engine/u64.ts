// 64-bit integer helpers over pairs of uint32 halves.
//
// The engine runs inside a sandboxed handler and must reproduce Go's 64-bit
// arithmetic bit for bit. It deliberately avoids BigInt so it depends on
// nothing beyond ES5-era number semantics: every intermediate value below
// stays an exact integer under 2^53.

/** A 64-bit word. `hi` and `lo` are each an integer in [0, 2^32). */
export interface U64 {
	readonly hi: number;
	readonly lo: number;
}

const TWO32 = 4294967296;

/** Wrapping 64-bit addition. */
export function add64(a: U64, b: U64): U64 {
	const lo = a.lo + b.lo;
	const carry = lo >= TWO32 ? 1 : 0;
	return { hi: (a.hi + b.hi + carry) >>> 0, lo: lo >>> 0 };
}

/** The exact 64-bit product of two uint32 values. */
export function mul32(a: number, b: number): U64 {
	const a0 = a & 0xffff;
	const a1 = a >>> 16;
	const b0 = b & 0xffff;
	const b1 = b >>> 16;
	const p00 = a0 * b0;
	const p01 = a0 * b1;
	const p10 = a1 * b0;
	const p11 = a1 * b1;
	// mid < 2^34, exact.
	const mid = p01 + p10 + Math.floor(p00 / 65536);
	const lo = ((mid & 0xffff) * 65536 + (p00 & 0xffff)) >>> 0;
	const hi = (p11 + Math.floor(mid / 65536)) >>> 0;
	return { hi, lo };
}

/**
 * A two's-complement int64 from a JavaScript integer. The integer must be
 * safe (|n| <= 2^53 - 1).
 */
export function int64FromNumber(n: number): U64 {
	if (!Number.isSafeInteger(n)) {
		throw new Error('int64FromNumber: not a safe integer: ' + n);
	}
	if (n >= 0) {
		return { hi: Math.floor(n / TWO32) >>> 0, lo: n >>> 0 };
	}
	const mag = -n;
	return negate64({ hi: Math.floor(mag / TWO32) >>> 0, lo: mag >>> 0 });
}

/** Two's-complement negation. */
export function negate64(a: U64): U64 {
	const lo = (~a.lo + 1) >>> 0;
	const hi = (~a.hi + (lo === 0 ? 1 : 0)) >>> 0;
	return { hi, lo };
}

/**
 * Parses a decimal integer string (optionally negative) into a 64-bit word,
 * wrapping modulo 2^64. Accepts both uint64 and int64 renderings.
 */
export function u64FromDecimal(text: string): U64 {
	let negative = false;
	let i = 0;
	if (text.charAt(0) === '-') {
		negative = true;
		i = 1;
	}
	if (i >= text.length) throw new Error('u64FromDecimal: empty');
	// Four 16-bit limbs, least significant first.
	const limbs = [0, 0, 0, 0];
	for (; i < text.length; i++) {
		const digit = text.charCodeAt(i) - 48;
		if (digit < 0 || digit > 9) throw new Error('u64FromDecimal: bad digit in ' + text);
		let carry = digit;
		for (let k = 0; k < 4; k++) {
			const v = limbs[k] * 10 + carry;
			limbs[k] = v & 0xffff;
			carry = v >>> 16;
		}
	}
	const word: U64 = {
		hi: (limbs[3] * 65536 + limbs[2]) >>> 0,
		lo: (limbs[1] * 65536 + limbs[0]) >>> 0,
	};
	return negative ? negate64(word) : word;
}

/** Renders a 64-bit word as an unsigned decimal string. */
export function u64ToDecimal(a: U64): string {
	const limbs = [a.hi >>> 16, a.hi & 0xffff, a.lo >>> 16, a.lo & 0xffff];
	let out = '';
	for (;;) {
		let rem = 0;
		let any = false;
		for (let k = 0; k < 4; k++) {
			const v = rem * 65536 + limbs[k];
			limbs[k] = Math.floor(v / 10);
			rem = v % 10;
			if (limbs[k] !== 0) any = true;
		}
		out = String(rem) + out;
		if (!any) break;
	}
	return out;
}

/** Renders a 64-bit word as a signed (two's-complement) decimal string. */
export function i64ToDecimal(a: U64): string {
	if (a.hi >= 0x80000000) return '-' + u64ToDecimal(negate64(a));
	return u64ToDecimal(a);
}

/** The UTF-8 encoding of a string, matching Go's []byte(string). */
export function utf8Bytes(text: string): number[] {
	const out: number[] = [];
	for (let i = 0; i < text.length; i++) {
		let code = text.charCodeAt(i);
		if (code >= 0xd800 && code <= 0xdbff && i + 1 < text.length) {
			const next = text.charCodeAt(i + 1);
			if (next >= 0xdc00 && next <= 0xdfff) {
				code = 0x10000 + ((code - 0xd800) << 10) + (next - 0xdc00);
				i++;
			}
		}
		if (code >= 0xd800 && code <= 0xdfff) {
			// A lone surrogate. Go strings cannot hold one; it would have
			// arrived as U+FFFD.
			code = 0xfffd;
		}
		if (code < 0x80) {
			out.push(code);
		} else if (code < 0x800) {
			out.push(0xc0 | (code >> 6), 0x80 | (code & 0x3f));
		} else if (code < 0x10000) {
			out.push(0xe0 | (code >> 12), 0x80 | ((code >> 6) & 0x3f), 0x80 | (code & 0x3f));
		} else {
			out.push(0xf0 | (code >> 18), 0x80 | ((code >> 12) & 0x3f), 0x80 | ((code >> 6) & 0x3f), 0x80 | (code & 0x3f));
		}
	}
	return out;
}
