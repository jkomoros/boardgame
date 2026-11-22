// Go's strings: package strings and strconv, and the string semantics
// that differ from JavaScript's. Go strings are UTF-8 bytes: `<` orders by
// bytes and `len` counts them. Indexing or slicing a string is refused by
// the translator (string-bytes), so nothing here does it.
//
// Every answer here is held to Go's by testdata/gen/facadevec. Where Go's
// answer depends on Unicode tables (ToLower, Quote's notion of printable),
// the JavaScript engine's tables are used; the vectors cover the cases games
// write (ASCII, accented letters, Greek, emoji).

import { compareGoStrings } from '../state.ts';

// ----- package strings ----------------------------------------------------

/** Go: `strings.Join(elems, sep)`. A nil slice joins to ''. */
export function join(elems: readonly string[] | null, sep: string): string {
	return elems === null ? '' : elems.join(sep);
}

/** Go's panic text for strings.Repeat with a negative count. */
export const REPEAT_NEGATIVE_TEXT = 'strings: negative Repeat count';

/** Go: `strings.Repeat(s, count)`. Panics with Go's text on a negative count. */
export function repeat(s: string, count: number): string {
	if (count < 0) throw new Error(REPEAT_NEGATIVE_TEXT);
	return s.repeat(count);
}

/** Go: `strings.Split(s, sep)`. An empty sep splits after each UTF-8 sequence. */
export function split(s: string, sep: string): string[] {
	if (sep === '') return Array.from(s);
	return s.split(sep);
}

/**
 * Go: `strings.ToLower(s)`, with Go's Unicode case mapping: rune by rune,
 * with no context (a word-final capital sigma becomes σ, not ς) and one
 * rune for one rune (İ becomes i, not i plus a combining dot).
 */
export function toLower(s: string): string {
	let out = '';
	for (const ch of s) {
		// Go maps one rune to one rune. Where JavaScript's mapping is longer
		// (U+0130 is the only case in Unicode's SpecialCasing for lower case),
		// Go's unicode.ToLower gives the first rune of it.
		out += String.fromCodePoint(ch.toLowerCase().codePointAt(0)!);
	}
	return out;
}

/** unicode.IsSpace: Go's definition of white space. */
function isGoSpace(code: number): boolean {
	switch (code) {
		case 0x09:
		case 0x0a:
		case 0x0b:
		case 0x0c:
		case 0x0d:
		case 0x20:
		case 0x85:
		case 0xa0:
		case 0x1680:
		case 0x2028:
		case 0x2029:
		case 0x202f:
		case 0x205f:
		case 0x3000:
			return true;
	}
	return code >= 0x2000 && code <= 0x200a;
}

/** Go: `strings.TrimSpace(s)`, with Go's definition of white space. */
export function trimSpace(s: string): string {
	let start = 0;
	let end = s.length;
	while (start < end && isGoSpace(s.charCodeAt(start))) start++;
	while (end > start && isGoSpace(s.charCodeAt(end - 1))) end--;
	return s.slice(start, end);
}

/** Go: `strings.HasPrefix(s, prefix)`. */
export function hasPrefix(s: string, prefix: string): boolean {
	return s.startsWith(prefix);
}

/** Go: `strings.HasSuffix(s, suffix)`. */
export function hasSuffix(s: string, suffix: string): boolean {
	return s.endsWith(suffix);
}

/** Go: `strings.Contains(s, substr)`. */
export function contains(s: string, substr: string): boolean {
	return s.includes(substr);
}

// ----- comparison and length ----------------------------------------------

/**
 * Go: `<`, `<=`, `>`, `>=` on strings, and `strings.Compare`. -1, 0 or 1,
 * ordering by code point, which is the order of the UTF-8 bytes.
 * (JavaScript's own `<` compares UTF-16 units and differs outside the basic
 * plane.)
 */
export function strCompare(a: string, b: string): number {
	return Math.sign(compareGoStrings(a, b)) + 0;
}

/** Go: `len(s)` on a string: its length in UTF-8 bytes. */
export function utf8Len(s: string): number {
	let bytes = 0;
	for (const ch of s) {
		const code = ch.codePointAt(0)!;
		if (code < 0x80) bytes += 1;
		else if (code < 0x800) bytes += 2;
		else if (code < 0x10000) bytes += 3;
		else bytes += 4;
	}
	return bytes;
}

// ----- package strconv ----------------------------------------------------

/** Go: `strconv.Itoa(n)`. */
export function itoa(n: number): string {
	return String(n + 0);
}

/** Go's uint64 cutoff in ParseUint: (1<<64-1)/10 + 1, in decimal. */
const UINT64_CUTOFF = '1844674407370955162';
const UINT64_MAX = '18446744073709551615';
const INT64_MIN_MAGNITUDE = '9223372036854775808';
/** math.MaxInt64 and math.MinInt64, as numbers (both read as 2^63 in magnitude). */
const INT64_MAX = 9223372036854775807;
const INT64_MIN = -9223372036854775808;

/** Compares two decimal digit strings without leading zeros, as numbers. */
function compareDigits(a: string, b: string): number {
	if (a.length !== b.length) return a.length - b.length;
	return a < b ? -1 : a > b ? 1 : 0;
}

/**
 * Go: `strconv.Atoi(s)`, as a tuple: the value and null, or 0 and Go's
 * error text (`strconv.Atoi: parsing "x": invalid syntax`, `... value out
 * of range`). Syntax and range are decided in Go's order: the digits are
 * scanned left to right and the first failure wins.
 *
 * DIVERGENCE (as legal/goformat.ts goAtoi): Go accepts any int64. A number
 * is exact only to 2^53, so an integer that fits int64 but is not a safe
 * integer is reported as out of range rather than rounded. On a range
 * error the value is Go's: math.MaxInt64 or math.MinInt64, which as numbers
 * are 2^63 and -2^63.
 */
export function atoi(s: string): [number, string | null] {
	const syntax: [number, string | null] = [0, 'strconv.Atoi: parsing ' + quote(s) + ': invalid syntax'];
	let body = s;
	let negative = false;
	if (body.length > 0 && (body[0] === '+' || body[0] === '-')) {
		negative = body[0] === '-';
		body = body.slice(1);
	}
	// On a range error Go returns the int64 bound of the input's sign.
	const range: [number, string | null] = [negative ? INT64_MIN : INT64_MAX, 'strconv.Atoi: parsing ' + quote(s) + ': value out of range'];
	if (body.length === 0) return syntax;
	// The fast path (fewer than 19 bytes) cannot overflow; the slow path
	// (ParseUint) reports overflow as soon as the value passes uint64.
	let digits = '';
	for (let i = 0; i < body.length; i++) {
		const c = body.charCodeAt(i);
		if (c < 0x30 || c > 0x39) return syntax;
		if (compareDigits(digits, UINT64_CUTOFF) >= 0) return range;
		digits = digits === '' || digits === '0' ? String(c - 0x30) : digits + String(c - 0x30);
		if (compareDigits(digits, UINT64_MAX) > 0) return range;
	}
	if (digits === '0') digits = '';
	if (digits !== '') {
		const limit = compareDigits(digits, INT64_MIN_MAGNITUDE);
		if (negative ? limit > 0 : limit >= 0) return range;
	}
	const value = digits === '' ? 0 : Number(digits);
	if (!Number.isSafeInteger(value)) return range;
	return [negative ? -value + 0 : value, null];
}

/**
 * unicode.IsPrint / strconv.IsPrint: letters, marks, numbers, punctuation,
 * symbols and the ASCII space.
 */
function isPrint(ch: string): boolean {
	return ch === ' ' || /^[\p{L}\p{M}\p{N}\p{P}\p{S}]$/u.test(ch);
}

function hex(code: number, width: number): string {
	return code.toString(16).padStart(width, '0');
}

/** strconv's appendEscapedRune. */
function escapeRune(ch: string, quoteChar: string): string {
	if (ch === quoteChar || ch === '\\') return '\\' + ch;
	if (isPrint(ch)) return ch;
	switch (ch) {
		case '\u0007':
			return '\\a';
		case '\b':
			return '\\b';
		case '\f':
			return '\\f';
		case '\n':
			return '\\n';
		case '\r':
			return '\\r';
		case '\t':
			return '\\t';
		case '\v':
			return '\\v';
	}
	let code = ch.codePointAt(0)!;
	if (code < 0x20 || code === 0x7f) return '\\x' + hex(code, 2);
	// A lone UTF-16 surrogate has no UTF-8 encoding; Go would see U+FFFD.
	if (code >= 0xd800 && code <= 0xdfff) code = 0xfffd;
	if (code < 0x10000) return '\\u' + hex(code, 4);
	return '\\U' + hex(code, 8);
}

/** Go: `strconv.Quote(s)`, which is also `%q` of a string. */
export function quote(s: string): string {
	let out = '"';
	for (const ch of s) out += escapeRune(ch, '"');
	return out + '"';
}

/** Go: `strconv.QuoteRune(r)`, which is also `%q` of an integer. */
export function quoteRune(r: number): string {
	const valid = Number.isInteger(r) && r >= 0 && r <= 0x10ffff && !(r >= 0xd800 && r <= 0xdfff);
	return "'" + escapeRune(String.fromCodePoint(valid ? r : 0xfffd), "'") + "'";
}

// ----- package fmt --------------------------------------------------------

/** One parsed verb: its flags, width, and letter. */
interface Directive {
	minus: boolean;
	plus: boolean;
	zero: boolean;
	width: number;
	verb: string;
}

/** Go's type name for a value, as fmt prints it in %!verb(type=value). */
function goTypeName(arg: unknown): string {
	if (typeof arg === 'string') return 'string';
	if (typeof arg === 'boolean') return 'bool';
	if (typeof arg === 'number') return Number.isInteger(arg) ? 'int' : 'float64';
	if (Array.isArray(arg)) return arg.length === 0 ? '[]interface {}' : '[]' + goTypeName(arg[0]);
	throw new Error('go/str: sprintf: unsupported argument ' + String(arg));
}

/** strconv.FormatFloat(x, 'g', -1, 64): the shortest form fmt's %v prints. */
function formatFloat(x: number): string {
	if (Number.isNaN(x)) return 'NaN';
	if (x === Infinity) return '+Inf';
	if (x === -Infinity) return '-Inf';
	const sign = x < 0 ? '-' : '';
	const [mantissa, exponent] = Math.abs(x).toExponential().split('e');
	const digits = mantissa.replace('.', '');
	const exp = Number(exponent);
	if (exp < -4 || exp >= 6) {
		const e = Math.abs(exp);
		return sign + digits[0] + (digits.length > 1 ? '.' + digits.slice(1) : '') + 'e' + (exp < 0 ? '-' : '+') + (e < 10 ? '0' + e : String(e));
	}
	const point = exp + 1;
	if (point <= 0) return sign + '0.' + '0'.repeat(-point) + digits;
	if (point >= digits.length) return sign + digits + '0'.repeat(point - digits.length);
	return sign + digits.slice(0, point) + '.' + digits.slice(point);
}

/**
 * fmt's padding: width counts runes. The '0' flag pads with zeros whatever
 * the verb (Go pads `%06s` to `0000ab`), after the sign of a number.
 */
function pad(text: string, d: Directive, numeric: boolean): string {
	const length = Array.from(text).length;
	if (length >= d.width) return text;
	const fill = d.width - length;
	if (d.minus) return text + ' '.repeat(fill);
	if (!d.zero) return ' '.repeat(fill) + text;
	const signed = numeric && (text[0] === '-' || text[0] === '+');
	return signed ? text[0] + '0'.repeat(fill) + text.slice(1) : '0'.repeat(fill) + text;
}

/** %v of a value: the text fmt prints for it with no verb-specific form. */
function plainValue(arg: unknown): string {
	if (arg === null) return '<nil>';
	if (typeof arg === 'number') return Number.isInteger(arg) ? String(arg + 0) : formatFloat(arg);
	if (typeof arg === 'string') return arg;
	if (typeof arg === 'boolean') return arg ? 'true' : 'false';
	if (Array.isArray(arg)) return '[' + arg.map(plainValue).join(' ') + ']';
	throw new Error('go/str: sprintf: unsupported argument ' + String(arg));
}

function badVerb(d: Directive, arg: unknown): string {
	if (arg === null) return '%!' + d.verb + '(<nil>)';
	return '%!' + d.verb + '(' + goTypeName(arg) + '=' + plainValue(arg) + ')';
}

/** One argument under one verb. A slice applies the verb to each element. */
function formatArg(d: Directive, arg: unknown): string {
	if (Array.isArray(arg)) return '[' + arg.map(element => formatArg(d, element)).join(' ') + ']';
	if (arg === null) return d.verb === 'v' ? pad('<nil>', d, false) : badVerb(d, arg);
	switch (d.verb) {
		case 'v':
			// %+v means struct field names, not a sign: Go clears plus for v.
			if (typeof arg === 'number') return pad(plainValue(arg), d, true);
			return pad(plainValue(arg), d, false);
		case 'd':
			if (typeof arg === 'number' && Number.isInteger(arg)) return pad(signed(String(arg + 0), d, arg), d, true);
			return badVerb(d, arg);
		case 's':
			if (typeof arg === 'string') return pad(arg, d, false);
			return badVerb(d, arg);
		case 'q':
			if (typeof arg === 'string') return pad(quote(arg), d, false);
			if (typeof arg === 'number' && Number.isInteger(arg)) return pad(quoteRune(arg), d, false);
			return badVerb(d, arg);
	}
	// A verb fmt knows but this port does not (%x, %f, ...) would print
	// something real in Go; refuse it. Any other letter is a bad verb in Go
	// too, and prints as one.
	if ('bcdeEfFgGoOpqstTUvxX'.includes(d.verb)) throw new Error('go/str: sprintf: verb %' + d.verb + ' is not supported');
	return badVerb(d, arg);
}

function signed(text: string, d: Directive, value: number): string {
	return d.plus && value >= 0 && !Number.isNaN(value) && value !== Infinity ? '+' + text : text;
}

/**
 * Go: `fmt.Sprintf(format, args...)` and the text of `fmt.Errorf`, for the
 * verbs `%d %s %v %q` with the flags `-`, `+`, `0` and a width, and `%%`.
 * Go's own markers for a bad call are reproduced: `%!d(string=x)` for a
 * verb that does not fit, `%!d(MISSING)`, `%!(EXTRA int=1)`, `%!(NOVERB)`.
 *
 * A number is an int when it is an integer and a float64 otherwise (Go's
 * static type is not visible here). null prints as Go's nil interface,
 * `<nil>`; a nil slice argument would print `[]` in Go.
 */
export function sprintf(format: string, ...args: readonly unknown[]): string {
	let out = '';
	let next = 0;
	let i = 0;
	while (i < format.length) {
		const ch = format[i];
		if (ch !== '%') {
			out += ch;
			i++;
			continue;
		}
		i++;
		const d: Directive = { minus: false, plus: false, zero: false, width: 0, verb: '' };
		for (; i < format.length; i++) {
			const flag = format[i];
			if (flag === '-') {
				d.minus = true;
				d.zero = false;
			}
			else if (flag === '+') d.plus = true;
			else if (flag === '0') d.zero = !d.minus;
			else break;
		}
		while (i < format.length && format[i] >= '0' && format[i] <= '9') {
			d.width = d.width * 10 + (format.charCodeAt(i) - 0x30);
			i++;
		}
		if (i >= format.length) {
			out += '%!(NOVERB)';
			break;
		}
		const verb = String.fromCodePoint(format.codePointAt(i)!);
		i += verb.length;
		if (verb === '%') {
			out += '%';
			continue;
		}
		d.verb = verb;
		if (next >= args.length) {
			out += '%!' + verb + '(MISSING)';
			continue;
		}
		out += formatArg(d, args[next++]);
	}
	if (next < args.length) {
		const extra = args.slice(next).map(arg => (arg === null ? '<nil>' : goTypeName(arg) + '=' + plainValue(arg)));
		out += '%!(EXTRA ' + extra.join(', ') + ')';
	}
	return out;
}
