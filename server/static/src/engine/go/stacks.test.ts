// The stack half of the facade (api.ts) against real Go.
// testdata/facade-stacks.vectors.json is written by testdata/gen/facadevec:
// sequences of component and stack calls in Go's own shapes on live vecgame
// states, with what each returned (an error text or nil, a component, a
// stack and slot), the text of each panic, and every top-level property of
// the state that changed. The facade replays each sequence through the same
// calls a translation would make and must agree after every operation,
// failures included: a failed operation must change nothing.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import type { GameDefinition } from '../definition.ts';
import type { StateJSON, StackRef as EngineStackRef } from '../state.ts';
import type { RulesState } from '../rules/state.ts';
import {
	componentAt, componentAtKey, components, containingStack, contractSize, dynamic, enumEquals, enumKey, expandSize, first, last,
	mayMoveAllTo, mayMoveCountTo, mayMoveTo, mayMoveToSlot, moveAllTo, moveCountTo, moveTo, moveToFirstSlot, moveToLastSlot, moveToNextSlot,
	publicShuffle, secretMoveTo, setEnumKey, setSize, shuffle, sizeToFit, slideToFirstSlot, slideToLastSlot, sortComponents, swapComponents,
	swapComponentsByKey, values,
} from './api.ts';
import type { Comp, StackRef } from './api.ts';
import { rulesStateFor } from './testkit.test.ts';

const read = (name: string) => JSON.parse(readFileSync(new URL('../testdata/' + name, import.meta.url), 'utf8'));
const def = read('defs/vecgame.def.json') as GameDefinition;

interface Op {
	Op: string;
	Src?: EngineStackRef;
	Dst?: EngineStackRef;
	Snap?: number;
	Pos?: number;
	Index?: number;
	Slot?: number;
	N?: number;
	Owner?: string;
	Player?: number;
	Key?: number;
	Result: unknown;
	Panic?: string;
	Changed?: Record<string, unknown>;
}
interface Scenario {
	GameID: string;
	SecretSalt: string;
	Version: number;
	NumPlayers: number;
	Before: StateJSON;
	Ops: Op[];
	After: StateJSON;
}
const { Scenarios } = read('facade-stacks.vectors.json') as { Scenarios: Scenario[] };

/** The state blob's top-level properties, as the generator flattens them. */
function flatten(blob: StateJSON): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const key of Object.keys(blob.Game)) out['Game.' + key] = blob.Game[key];
	blob.Players.forEach((p, i) => {
		for (const key of Object.keys(p)) out[`Players[${i}].${key}`] = p[key];
	});
	for (const deck of Object.keys(blob.Components ?? {})) {
		blob.Components[deck].forEach((p, i) => {
			for (const key of Object.keys(p)) out[`Components.${deck}[${i}].${key}`] = p[key];
		});
	}
	if (blob.SecretMoveCount && Object.keys(blob.SecretMoveCount).length > 0) out.SecretMoveCount = blob.SecretMoveCount;
	return out;
}

/** vecgame's StackRef.Path for a facade stack. */
function pathOf(s: StackRef): string {
	const at = s.at;
	const base = at.Group === 'game' ? 'Game' : at.Group === 'player' ? `Players[${at.Player ?? 0}]` : `DynamicComponentValues[${at.Deck}][${at.Index ?? 0}]`;
	return base + '.' + at.Prop + (at.Space === undefined ? '' : `[${at.Space}]`);
}

function compJSON(c: Comp | null): unknown {
	return c === null ? null : { Deck: c.deck, Index: c.index };
}

function rank(state: RulesState, c: Comp): number {
	return c.deck === 'cards' ? (values(state, c).Rank as number) : 0;
}

/** Runs one recorded operation through the facade; returns its result as the generator encodes it. */
function perform(state: RulesState, op: Op, snaps: (Comp | null)[][]): unknown {
	const ref = (at: EngineStackRef | undefined): StackRef => ({ state, at: at! });
	const src = ref(op.Src);
	const dst = ref(op.Dst);
	const handle = (): Comp | null => (op.Snap !== undefined ? snaps[op.Snap][op.Pos ?? 0] : componentAt(src, op.Index ?? 0));
	switch (op.Op) {
		case 'move': return moveTo(handle(), dst, op.Slot ?? 0);
		case 'secretMove': return secretMoveTo(handle(), dst, op.Slot ?? 0);
		case 'moveFirst': return moveToFirstSlot(handle(), dst);
		case 'moveLast': return moveToLastSlot(handle(), dst);
		case 'moveNext': return moveToNextSlot(handle(), dst);
		case 'slideFirst': return slideToFirstSlot(handle());
		case 'slideLast': return slideToLastSlot(handle());
		case 'mayMove': return mayMoveTo(handle(), dst);
		case 'mayMoveSlot': return mayMoveToSlot(handle(), dst, op.Slot ?? 0);
		case 'containing': {
			const [stack, slot, err] = containingStack(state, handle());
			return { Path: stack === null ? null : pathOf(stack), Slot: slot, Err: err };
		}
		case 'values': return values(state, handle());
		case 'dynamic': return dynamic(state, handle()).Counter;
		case 'dynamicBump': {
			const d = dynamic(state, handle());
			(d.Counter as number)++;
			return d.Counter;
		}
		case 'markGet': return enumKey(state, { dynamicOf: handle() }, 'Mark');
		case 'markSet': return setEnumKey(state, { dynamicOf: handle() }, 'Mark', op.Key ?? 0);
		case 'colorGet': return enumKey(state, op.Owner === 'game' ? 'game' : (op.Player ?? 0), 'Color');
		case 'colorSet': return setEnumKey(state, op.Owner === 'game' ? 'game' : (op.Player ?? 0), 'Color', op.Key ?? 0);
		case 'colorEquals': return enumEquals(state, 'game', 'Color', op.Player ?? 0, 'Color');
		case 'moveAll': return moveAllTo(src, dst);
		case 'moveCount': return moveCountTo(src, dst, op.N ?? 0);
		case 'mayMoveAll': return mayMoveAllTo(src, dst);
		case 'mayMoveCount': return mayMoveCountTo(src, dst, op.N ?? 0);
		case 'shuffle': return shuffle(src);
		case 'publicShuffle': return publicShuffle(src);
		case 'swap': return swapComponents(src, op.Index ?? 0, op.Slot ?? 0);
		case 'swapByKey': return swapComponentsByKey(src, op.Index ?? 0, op.Slot ?? 0);
		case 'sortRank':
			return sortComponents(src, (a, b) => {
				if (a === null) return false;
				if (b === null) return true;
				return rank(state, a) < rank(state, b);
			});
		case 'expand': return expandSize(src, op.N ?? 0);
		case 'contract': return contractSize(src, op.N ?? 0);
		case 'setSize': return setSize(src, op.N ?? 0);
		case 'sizeToFit': return sizeToFit(src);
		case 'first': return compJSON(first(src));
		case 'last': return compJSON(last(src));
		case 'componentAt': return compJSON(componentAt(src, op.Index ?? 0));
		case 'componentAtKey': return compJSON(componentAtKey(src, op.Index ?? 0));
		case 'components': {
			const snap = components(src);
			snaps.push(snap);
			return snap.map(compJSON);
		}
	}
	throw new Error('unknown op ' + op.Op);
}

test('the vectors exercise every facade stack call, stale handles and panics', () => {
	const kinds = new Set<string>();
	let panics = 0;
	let errors = 0;
	let stale = 0;
	for (const sc of Scenarios) {
		for (const op of sc.Ops) {
			kinds.add(op.Op);
			if (op.Panic) panics++;
			if (typeof op.Result === 'string') errors++;
			if (typeof op.Result === 'string' && op.Result.includes('not public information')) stale++;
		}
	}
	assert.equal(kinds.size, 36);
	assert.ok(panics > 50 && errors > 100 && stale >= 2, `${panics} panics, ${errors} errors, ${stale} stale`);
});

for (const [n, sc] of Scenarios.entries()) {
	test(`scenario ${n}: ${sc.Ops.length} facade calls on a ${sc.NumPlayers}-player vecgame state at version ${sc.Version}`, () => {
		const state = rulesStateFor(def, { gameId: sc.GameID, secretSalt: sc.SecretSalt }, sc.Version, sc.Before);
		const expected = flatten(sc.Before);
		assert.deepEqual(flatten(state.engine.toJSON()), expected, 'before');
		const snaps: (Comp | null)[][] = [];
		sc.Ops.forEach((op, i) => {
			const label = `op ${i} ${op.Op} ${JSON.stringify({ ...op, Result: undefined, Changed: undefined })}`;
			let result: unknown = null;
			let panic = '';
			try {
				result = perform(state, op, snaps);
			} catch (e) {
				panic = (e as Error).message;
			}
			assert.equal(panic, op.Panic ?? '', label);
			assert.deepEqual(result === undefined ? null : result, op.Result, label);
			Object.assign(expected, op.Changed ?? {});
			assert.deepEqual(flatten(state.engine.toJSON()), expected, label + ': state');
		});
		assert.deepEqual(state.engine.toJSON(), sc.After, 'after');
	});
}
