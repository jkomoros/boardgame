// A pure SHA-1 (FIPS 180-4). The Go core derives every component id from
// crypto/sha1, and the engine cannot use a platform crypto API: it runs in a
// sandboxed handler with none.
//
// SHA-1 is used here as a fingerprint, exactly as the core uses it. It is not
// relied on for collision resistance.

import { utf8Bytes } from './u64.ts';

const HEX = '0123456789abcdef';

/** The SHA-1 digest of bytes, as 40 lowercase hex characters. */
export function sha1HexBytes(bytes: ArrayLike<number>): string {
	const length = bytes.length;
	// Message + 0x80 + zero pad + 64-bit big-endian bit length.
	const total = ((length + 8) >> 6) * 64 + 64;
	const msg = new Array<number>(total);
	for (let i = 0; i < length; i++) msg[i] = bytes[i] & 0xff;
	msg[length] = 0x80;
	for (let i = length + 1; i < total; i++) msg[i] = 0;
	const bitsHi = Math.floor(length / 536870912); // length * 8 / 2^32
	const bitsLo = (length << 3) >>> 0;
	msg[total - 8] = (bitsHi >>> 24) & 0xff;
	msg[total - 7] = (bitsHi >>> 16) & 0xff;
	msg[total - 6] = (bitsHi >>> 8) & 0xff;
	msg[total - 5] = bitsHi & 0xff;
	msg[total - 4] = (bitsLo >>> 24) & 0xff;
	msg[total - 3] = (bitsLo >>> 16) & 0xff;
	msg[total - 2] = (bitsLo >>> 8) & 0xff;
	msg[total - 1] = bitsLo & 0xff;

	let h0 = 0x67452301;
	let h1 = 0xefcdab89;
	let h2 = 0x98badcfe;
	let h3 = 0x10325476;
	let h4 = 0xc3d2e1f0;
	const w = new Array<number>(80);

	for (let block = 0; block < total; block += 64) {
		for (let t = 0; t < 16; t++) {
			const o = block + t * 4;
			w[t] = ((msg[o] << 24) | (msg[o + 1] << 16) | (msg[o + 2] << 8) | msg[o + 3]) | 0;
		}
		for (let t = 16; t < 80; t++) {
			const x = w[t - 3] ^ w[t - 8] ^ w[t - 14] ^ w[t - 16];
			w[t] = (x << 1) | (x >>> 31);
		}
		let a = h0;
		let b = h1;
		let c = h2;
		let d = h3;
		let e = h4;
		for (let t = 0; t < 80; t++) {
			let f: number;
			let k: number;
			if (t < 20) {
				f = (b & c) | (~b & d);
				k = 0x5a827999;
			} else if (t < 40) {
				f = b ^ c ^ d;
				k = 0x6ed9eba1;
			} else if (t < 60) {
				f = (b & c) | (b & d) | (c & d);
				k = 0x8f1bbcdc;
			} else {
				f = b ^ c ^ d;
				k = 0xca62c1d6;
			}
			const temp = (((a << 5) | (a >>> 27)) + f + e + k + w[t]) | 0;
			e = d;
			d = c;
			c = (b << 30) | (b >>> 2);
			b = a;
			a = temp;
		}
		h0 = (h0 + a) | 0;
		h1 = (h1 + b) | 0;
		h2 = (h2 + c) | 0;
		h3 = (h3 + d) | 0;
		h4 = (h4 + e) | 0;
	}

	let out = '';
	const words = [h0, h1, h2, h3, h4];
	for (let i = 0; i < 5; i++) {
		for (let shift = 28; shift >= 0; shift -= 4) {
			out += HEX.charAt((words[i] >>> shift) & 0xf);
		}
	}
	return out;
}

/** The SHA-1 digest of the UTF-8 bytes of text, as lowercase hex. */
export function sha1Hex(text: string): string {
	return sha1HexBytes(utf8Bytes(text));
}
