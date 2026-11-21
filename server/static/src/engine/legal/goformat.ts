// The few pieces of Go's standard library whose exact output shows up in
// legality text: strconv.Atoi (every numeric predicate argument), %q quoting
// (every error that names a path or an argument), and %v of a string slice.
//
// These exist so the TypeScript reproduces Go's Unknown reasons and
// construction errors byte for byte. They are not general-purpose ports.

/** A successful or failed strconv.Atoi. `error` is Go's err.Error() text. */
export type AtoiResult = { ok: true; value: number } | { ok: false; error: string };

const INT64_MAX = 9223372036854775807n;
const INT64_MIN = -9223372036854775808n;

/**
 * Go: strconv.Atoi. Accepts an optional sign and decimal digits, nothing else
 * (no whitespace, no underscores, no base prefix). The error text is the
 * *strconv.NumError Go returns.
 *
 * DIVERGENCE (documented in fabric/status/engine-legal.md): Go accepts any
 * int64. A JavaScript number is exact only to 2^53, so an integer that fits
 * int64 but not a safe integer is reported as out of range here rather than
 * silently rounded.
 */
export function goAtoi(s: string): AtoiResult {
  if (!/^[+-]?[0-9]+$/.test(s)) {
    return { ok: false, error: `strconv.Atoi: parsing ${goQuote(s)}: invalid syntax` };
  }
  const big = BigInt(s);
  if (big > INT64_MAX || big < INT64_MIN || big > BigInt(Number.MAX_SAFE_INTEGER) || big < BigInt(Number.MIN_SAFE_INTEGER)) {
    return { ok: false, error: `strconv.Atoi: parsing ${goQuote(s)}: value out of range` };
  }
  return { ok: true, value: Number(big) };
}

const QUOTE_ESCAPES: Readonly<Record<string, string>> = {
  '\u0007': '\\a',
  '\b': '\\b',
  '\f': '\\f',
  '\n': '\\n',
  '\r': '\\r',
  '\t': '\\t',
  '\v': '\\v',
  '\\': '\\\\',
  '"': '\\"',
};

/**
 * Go: strconv.Quote, the %q verb for a string. Printable text passes through;
 * quotes, backslashes and control characters are escaped the way Go escapes
 * them. Non-printable characters above U+007F are passed through unescaped,
 * which differs from Go only for text no legal path or argument contains.
 */
export function goQuote(s: string): string {
  let out = '"';
  for (const ch of s) {
    const escaped = QUOTE_ESCAPES[ch];
    if (escaped !== undefined) {
      out += escaped;
      continue;
    }
    const code = ch.codePointAt(0) as number;
    if (code < 0x20 || code === 0x7f) {
      out += '\\x' + code.toString(16).padStart(2, '0');
      continue;
    }
    out += ch;
  }
  return out + '"';
}

/** Go: fmt's %v of a []string, e.g. `[a b]`; a nil or empty slice is `[]`. */
export function goStringSlice(items: readonly string[] | undefined): string {
  return '[' + (items ?? []).join(' ') + ']';
}
