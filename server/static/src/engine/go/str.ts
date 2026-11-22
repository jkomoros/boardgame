// Go's strings: package strings and strconv, and the string semantics
// that differ from JavaScript's. Go strings are UTF-8 bytes: `<` orders by
// bytes and `len` counts them. Indexing or slicing a string is refused by
// the translator (string-bytes), so nothing here does it.

/** Go: `strings.Join(elems, sep)`. A nil slice joins to ''. */
export function join(elems: readonly string[] | null, sep: string): string {
	void elems;
	void sep;
	throw new Error('go/str: join not implemented');
}

/** Go: `strings.Repeat(s, count)`. Panics with Go's text on a negative count. */
export function repeat(s: string, count: number): string {
	void s;
	void count;
	throw new Error('go/str: repeat not implemented');
}

/** Go: `strings.Split(s, sep)`. An empty sep splits after each UTF-8 sequence. */
export function split(s: string, sep: string): string[] {
	void s;
	void sep;
	throw new Error('go/str: split not implemented');
}

/** Go: `strings.ToLower(s)`, with Go's Unicode case mapping. */
export function toLower(s: string): string {
	void s;
	throw new Error('go/str: toLower not implemented');
}

/** Go: `strings.TrimSpace(s)`, with Go's definition of white space. */
export function trimSpace(s: string): string {
	void s;
	throw new Error('go/str: trimSpace not implemented');
}

/** Go: `strings.HasPrefix(s, prefix)`. */
export function hasPrefix(s: string, prefix: string): boolean {
	void s;
	void prefix;
	throw new Error('go/str: hasPrefix not implemented');
}

/** Go: `strings.HasSuffix(s, suffix)`. */
export function hasSuffix(s: string, suffix: string): boolean {
	void s;
	void suffix;
	throw new Error('go/str: hasSuffix not implemented');
}

/** Go: `strings.Contains(s, substr)`. */
export function contains(s: string, substr: string): boolean {
	void s;
	void substr;
	throw new Error('go/str: contains not implemented');
}

/**
 * Go: `<`, `<=`, `>`, `>=` on strings, and `strings.Compare`. Negative, 0 or
 * positive, ordering by code point, which is the order of the UTF-8 bytes.
 * (JavaScript's own `<` compares UTF-16 units and differs outside the basic
 * plane.)
 */
export function strCompare(a: string, b: string): number {
	void a;
	void b;
	throw new Error('go/str: strCompare not implemented');
}

/** Go: `len(s)` on a string: its length in UTF-8 bytes. */
export function utf8Len(s: string): number {
	void s;
	throw new Error('go/str: utf8Len not implemented');
}

/** Go: `strconv.Itoa(n)`. */
export function itoa(n: number): string {
	void n;
	throw new Error('go/str: itoa not implemented');
}

/**
 * Go: `strconv.Atoi(s)`, as a tuple: the value and null, or 0 and Go's
 * error text (`strconv.Atoi: parsing "x": invalid syntax`, `... value out
 * of range`).
 */
export function atoi(s: string): [number, string | null] {
	void s;
	throw new Error('go/str: atoi not implemented');
}

/**
 * Go: `fmt.Sprintf(format, args...)` and the text of `fmt.Errorf`, for the
 * verbs `%d %s %v %q`, over legal/goformat.ts.
 */
export function sprintf(format: string, ...args: readonly unknown[]): string {
	void format;
	void args;
	throw new Error('go/str: sprintf not implemented');
}
