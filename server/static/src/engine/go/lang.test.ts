// The language half of the facade (num.ts, str.ts, slice.ts, panic.ts, and
// the enum and graph arithmetic of api.ts) against real Go.
// testdata/facade.vectors.json is written by testdata/gen/facadevec from the
// Go standard library, the runtime's recovered panics, and the framework's
// enum, graph and behavior packages. Nothing here is derived by hand.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { idiv, imod, imul, INT_OVERFLOW_PREFIX } from './num.ts';
import {
	atoi, contains, hasPrefix, hasSuffix, itoa, join, repeat, split, sprintf, strCompare, toLower, trimSpace, utf8Len,
} from './str.ts';
import { appendTo, at, copyInto, lenOf, makeSlice, subslice } from './slice.ts';
import {
	DIVIDE_BY_ZERO_TEXT, MAKESLICE_LEN_TEXT, NIL_DEREF_TEXT, NIL_MAP_WRITE_TEXT, indexOutOfRangeText, sliceBoundsText, divideByZero, nilDeref,
} from './panic.ts';
import {
	ILLEGAL_ENUM_VALUE, legalError, neighborsByKey, player, rangeToValue, rangeValue, setEliminated, setPlayerInactive, setSeatFilled, stockApply,
	valueToRange, variantValue,
} from './api.ts';
import { failT } from '../legal/types.ts';
import { stubRulesState } from './testkit.test.ts';
import type { EnumDef } from '../definition.ts';

const V = JSON.parse(readFileSync(new URL('../testdata/facade.vectors.json', import.meta.url), 'utf8'));

/** The message a call throws, or '' when it returns. */
function thrown(f: () => unknown): string {
	try {
		f();
		return '';
	} catch (e) {
		return (e as Error).message;
	}
}

test('idiv and imod truncate and panic as Go does', () => {
	assert.ok(V.Div.length > 50);
	for (const c of V.Div) {
		const label = `${c.A} / ${c.B}`;
		assert.equal(thrown(() => idiv(c.A, c.B)), c.QuoPanic ?? '', label);
		assert.equal(thrown(() => imod(c.A, c.B)), c.RemPanic ?? '', label);
		if (!c.QuoPanic) assert.ok(Object.is(idiv(c.A, c.B), c.Quo), label);
		if (!c.RemPanic) assert.ok(Object.is(imod(c.A, c.B), c.Rem), label);
	}
});

test('imul agrees with Go wherever the product is exact, and throws where it is not', () => {
	let exact = 0;
	for (const c of V.Mul) {
		const label = `${c.A} * ${c.B}`;
		if (c.Exact) {
			exact++;
			assert.ok(Object.is(imul(c.A, c.B), c.Product), label);
		} else {
			assert.ok(thrown(() => imul(c.A, c.B)).startsWith(INT_OVERFLOW_PREFIX), label);
		}
	}
	assert.ok(exact > 20 && exact < V.Mul.length);
});

test('atoi gives Go value and error text; itoa gives Go text', () => {
	let divergent = 0;
	for (const c of V.Atoi) {
		const [value, err] = atoi(c.S);
		if (!c.Safe) {
			// The documented divergence: Go accepts it, the facade cannot hold it.
			divergent++;
			assert.equal(err, 'strconv.Atoi: parsing ' + JSON.stringify(c.S) + ': value out of range', c.S);
			continue;
		}
		assert.equal(err, c.Err === '' ? null : c.Err, JSON.stringify(c.S));
		assert.ok(Object.is(value, c.Value), JSON.stringify(c.S));
	}
	assert.ok(divergent >= 3);
	for (const c of V.Itoa) assert.equal(itoa(c.N), c.S);
});

test('strCompare and utf8Len order and measure strings as Go does, outside the basic plane too', () => {
	for (const c of V.Compare) {
		assert.equal(strCompare(c.A, c.B), c.Compare, `${JSON.stringify(c.A)} vs ${JSON.stringify(c.B)}`);
		assert.equal(strCompare(c.A, c.B) < 0, c.Less);
	}
	// JavaScript's own < disagrees with Go on some of these pairs; the facade must not.
	assert.ok(V.Compare.some((c: { A: string; B: string; Less: boolean }) => (c.A < c.B) !== c.Less));
	for (const c of V.Len) assert.equal(utf8Len(c.S), c.Len, JSON.stringify(c.S));
});

/** The value an encoded Go argument stands for. */
function argValue(a: { T: string; V: unknown }): unknown {
	if (a.T === 'float64') {
		const text = a.V as string;
		return text === '+Inf' ? Infinity : text === '-Inf' ? -Infinity : Number(text);
	}
	return a.V;
}

test('sprintf prints %d %s %v %q, flags, widths and Go bad-call markers as fmt does', () => {
	for (const c of V.Sprintf) {
		const args = (c.Args ?? []).map(argValue);
		assert.equal(sprintf(c.Format, ...args), c.Out, `${JSON.stringify(c.Format)} ${JSON.stringify(c.Args)}`);
	}
	assert.match(thrown(() => sprintf('%x', 1)), /not supported/);
});

test('the strings functions agree with package strings', () => {
	for (const c of V.Strings) {
		const a = c.Args as unknown[];
		const call = (): unknown => {
			switch (c.Fn) {
				case 'ToLower': return toLower(a[0] as string);
				case 'TrimSpace': return trimSpace(a[0] as string);
				case 'Split': return split(a[0] as string, a[1] as string);
				case 'Join': return join(a[0] as string[] | null, a[1] as string);
				case 'Repeat': return repeat(a[0] as string, a[1] as number);
				case 'HasPrefix': return hasPrefix(a[0] as string, a[1] as string);
				case 'HasSuffix': return hasSuffix(a[0] as string, a[1] as string);
				case 'Contains': return contains(a[0] as string, a[1] as string);
			}
			throw new Error('unknown ' + c.Fn);
		};
		const label = `${c.Fn}(${JSON.stringify(a)})`;
		if (c.Panic) {
			assert.equal(thrown(call), c.Panic, label);
			continue;
		}
		assert.deepEqual(call(), c.Result, label);
	}
});

test('panic texts are the Go runtime texts', () => {
	for (const c of V.Panics) {
		const [x, y, z] = c.Args ?? [];
		switch (c.Kind) {
			case 'index': {
				const xs = makeSlice(y, 0);
				assert.equal(thrown(() => at(xs, x)), c.Text, `index ${x} of ${y}`);
				if (c.Text !== '') assert.equal(indexOutOfRangeText(x, y), c.Text);
				break;
			}
			case 'slice': {
				const xs = makeSlice(z, 0);
				assert.equal(thrown(() => subslice(xs, x, y)), c.Text, `[${x}:${y}] of ${z}`);
				assert.equal(sliceBoundsText(x, y, z), c.Text);
				break;
			}
			case 'nilDeref':
			case 'nilInterface':
				assert.equal(NIL_DEREF_TEXT, c.Text);
				assert.equal(thrown(nilDeref), c.Text);
				break;
			case 'nilMap':
				assert.equal(NIL_MAP_WRITE_TEXT, c.Text);
				break;
			case 'divide':
			case 'modulo':
				assert.equal(DIVIDE_BY_ZERO_TEXT, c.Text);
				assert.equal(thrown(divideByZero), c.Text);
				break;
			case 'makeslice':
				assert.equal(MAKESLICE_LEN_TEXT, c.Text);
				assert.equal(thrown(() => makeSlice(x, 0)), c.Text);
				break;
			default:
				assert.fail('unknown panic kind ' + c.Kind);
		}
	}
});

test('nil-aware slices serialize as Go does: null for nil, [] for empty', () => {
	const programs: Record<string, () => number[] | readonly number[] | null> = {
		var: () => null,
		appendNothingToNil: () => appendTo<number>(null),
		appendToNil: () => appendTo<number>(null, 1, 2),
		literalEmpty: () => [],
		makeZero: () => makeSlice(0, 0),
		makeThree: () => makeSlice(3, 0),
		appendNothingToEmpty: () => appendTo<number>([]),
		assignNil: () => null,
		subsliceNil: () => subslice<number>(null, 0, 0),
		subsliceEmpty: () => subslice([1, 2, 3], 1, 1),
		subsliceMiddle: () => subslice([1, 2, 3], 1, 3),
		lenNil: () => [lenOf(null)],
		copyShort: () => {
			const dst = makeSlice(2, 0);
			const n = copyInto(dst, [7, 8, 9]);
			return appendTo(dst, n);
		},
		copyLong: () => {
			const dst = [1, 2, 3, 4];
			const n = copyInto(dst, [7, 8]);
			return appendTo(dst, n);
		},
		copyNil: () => {
			const dst = [1];
			const n = copyInto(dst, null);
			return appendTo(dst, n);
		},
		copyOverlap: () => {
			// copy(x[1:], x): the facade's translation writes through a view.
			const x = [1, 2, 3, 4];
			const tail = x.slice(1);
			copyInto(tail, x);
			for (let i = 0; i < tail.length; i++) x[i + 1] = tail[i];
			return x;
		},
		aliasAfterAppend: () => {
			let a: number[] | null = [1, 2];
			const b = a;
			a = appendTo(a, 3);
			return [lenOf(a), lenOf(b)];
		},
	};
	for (const c of V.Slices) {
		const program = programs[c.Name];
		assert.ok(program, 'no program for ' + c.Name);
		assert.equal(JSON.stringify({ X: program() }), JSON.stringify(c.JSON), c.Name);
	}
	assert.equal(Object.keys(programs).length, V.Slices.length);
});

test('the behavior setters write what Go writes', () => {
	const setters: Record<string, (state: ReturnType<typeof stubRulesState>, i: number) => void> = {
		SetSeatFilled: setSeatFilled,
		SetEliminated: setEliminated,
		SetPlayerInactive: setPlayerInactive,
	};
	for (const name of Object.keys(V.Behaviors)) {
		const state = stubRulesState();
		const before = { ...state.players[1] };
		setters[name](state, 1);
		const changed: Record<string, unknown> = {};
		for (const key of Object.keys(state.players[1])) if (state.players[1][key] !== before[key]) changed[key] = state.players[1][key];
		assert.deepEqual(changed, V.Behaviors[name], name);
	}
});

test('range enums: RangeToValue, ValueToRange and RangeValue as enum.RangeEnum', () => {
	assert.equal(ILLEGAL_ENUM_VALUE, Number(V.IllegalValue));
	for (const e of V.RangeEnums) {
		const def: EnumDef = { Values: e.Values, DefaultValue: e.DefaultValue, Range: e.Dimensions };
		const state = stubRulesState({ [e.Name]: def }, { Spot: { Type: 'enum', Enum: e.Name } });
		for (const c of e.ValueToRange) assert.deepEqual(valueToRange(state, e.Name, c.Key), c.Indexes, `${e.Name} ${c.Key}`);
		for (const c of e.RangeToValue) {
			assert.equal(rangeToValue(state, e.Name, c.Indexes), c.Illegal ? ILLEGAL_ENUM_VALUE : c.Key, `${e.Name} ${JSON.stringify(c.Indexes)}`);
		}
		// RangeValue reads the stored string of a property holding each key.
		for (const c of e.ValueToRange) {
			if (c.Indexes === null) continue;
			state.g.Spot = e.Values[String(c.Key)];
			assert.deepEqual(rangeValue(state, 'game', 'Spot'), c.Indexes);
		}
	}
});

test('grid graphs: NeighborsByKey as graph.MustNewEnumGridConnectedness builds them', () => {
	for (const g of V.Graphs) {
		const ref = { enum: g.Enum, dimensions: g.Dimensions, directions: g.Directions };
		for (const key of Object.keys(g.Neighbors)) {
			const got = neighborsByKey(ref, Number(key));
			assert.deepEqual(got === null ? null : got.slice().sort((a, b) => a - b), g.Neighbors[key], `${g.Enum} ${JSON.stringify(g.Directions)} ${key}`);
		}
	}
	assert.match(thrown(() => neighborsByKey({ enum: 'grid', dimensions: [3, 4], directions: ['Sideways'] }, 0)), /Sideways/);
});

test('variantValue, legalError, player and stockApply', () => {
	const state = stubRulesState({}, {}, { numcards: 'small' });
	assert.equal(variantValue(state, 'numcards'), 'small');
	// A Go map lookup of a missing key is the zero value, including for keys
	// that name properties of a JavaScript object.
	assert.equal(variantValue(state, 'missing'), '');
	assert.equal(variantValue(state, 'constructor'), '');
	assert.equal(variantValue(stubRulesState(), 'numcards'), '');
	assert.deepEqual(legalError('memory.no_cards'), failT('memory.no_cards'));
	assert.deepEqual(legalError('memory.no_cards', { n: 2 }), failT('memory.no_cards', { n: 2 }));
	assert.equal(player(state, 2), state.players[2]);
	const index = V.Panics.find((c: { Kind: string; Args: number[] }) => c.Kind === 'index' && c.Args[0] === 5 && c.Args[1] === 3).Text;
	assert.equal(thrown(() => player(state, 5)), index);
	assert.match(thrown(() => stockApply(state, { name: 'x', fields: {} })), /^engine-gap: stockApply needs /);
});
