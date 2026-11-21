// hash/fnv's New64: 64-bit FNV-1 (multiply, then xor), which is what the Go
// core seeds its generator with. Note this is FNV-1, not FNV-1a.

import { utf8Bytes } from './u64.ts';
import type { U64 } from './u64.ts';

// offset64 = 14695981039346656037 = 0xcbf29ce484222325
// prime64  = 1099511628211        = 0x00000100000001b3 = 2^40 + 0x1b3

/** FNV-1 64-bit over raw bytes. */
export function fnv1_64Bytes(bytes: ArrayLike<number>): U64 {
	// 16-bit limbs, least significant first.
	let h0 = 0x2325;
	let h1 = 0x8422;
	let h2 = 0x9ce4;
	let h3 = 0xcbf2;
	for (let i = 0; i < bytes.length; i++) {
		// h * (2^40 + 0x1b3) mod 2^64. The 2^40 term is h << 40: limb k of h
		// lands at bit 16k+40, so h0 contributes (h0 << 8) to limb 2 upward
		// and h1 contributes its low byte to the top of limb 3.
		const m0 = h0 * 0x1b3;
		const m1 = h1 * 0x1b3 + (m0 >>> 16);
		const m2 = h2 * 0x1b3 + (m1 >>> 16) + ((h0 & 0xff) << 8);
		const m3 = h3 * 0x1b3 + (m2 >>> 16) + (h0 >>> 8) + ((h1 & 0xff) << 8);
		h0 = (m0 & 0xffff) ^ (bytes[i] & 0xff);
		h1 = m1 & 0xffff;
		h2 = m2 & 0xffff;
		h3 = m3 & 0xffff;
	}
	return { hi: (h3 * 65536 + h2) >>> 0, lo: (h1 * 65536 + h0) >>> 0 };
}

/** FNV-1 64-bit over the UTF-8 bytes of a string. */
export function fnv1_64(text: string): U64 {
	return fnv1_64Bytes(utf8Bytes(text));
}
