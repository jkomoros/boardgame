// Stack operations against real Go. testdata/stacks.vectors.json holds
// scripted sequences run on live boardgame states (testdata/gen/stackvec):
// the blob before, every stack that changed after each operation, each error
// Go returned, and the blob after. The engine replays each sequence and must
// agree at every step.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { EngineState, cloneJSON } from './state.ts';
import type { StateJSON, StackRef } from './state.ts';
import { Stack, StackError, MergedStackView } from './stack.ts';
import type { StackJSON } from './stack.ts';
import type { GameDefinition } from './definition.ts';
import { GoRand } from './rand.ts';
import { EMPTY_SLOT } from './ids.ts';

const read = (name: string) => JSON.parse(readFileSync(new URL('./testdata/' + name, import.meta.url), 'utf8'));
const def = read('defs/vecgame.def.json') as GameDefinition;

interface Op {
	Op: string;
	Src?: StackRef;
	Dst?: StackRef;
	Index: number;
	Slot: number;
	NoComponent?: boolean;
	Err: string;
	Changed?: Record<string, StackJSON>;
	SecretMoveCount?: Record<string, number[]>;
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
const { Scenarios } = read('stacks.vectors.json') as { Scenarios: Scenario[] };

function loadScenario(sc: Scenario): EngineState {
	return EngineState.load(def, { gameId: sc.GameID, secretSalt: sc.SecretSalt }, sc.Version, sc.Before);
}

function apply(state: EngineState, op: Op): void {
	const src = op.Src ? state.stack(op.Src) : undefined;
	const dst = op.Dst ? state.stack(op.Dst) : undefined;
	switch (op.Op) {
		case 'move': return src!.moveComponent(op.Index, dst!, op.Slot);
		case 'secretMove': return src!.secretMoveComponent(op.Index, dst!, op.Slot);
		case 'moveFirst': return src!.moveComponentToFirstSlot(op.Index, dst!);
		case 'moveLast': return src!.moveComponentToLastSlot(op.Index, dst!);
		case 'moveNext': return src!.moveComponentToNextSlot(op.Index, dst!);
		case 'slideFirst': return src!.slideToFirstSlot(op.Index);
		case 'slideLast': return src!.slideToLastSlot(op.Index);
		case 'moveAll': return src!.moveAllTo(dst!);
		case 'moveCount': return src!.moveCountTo(dst!, op.Index);
		case 'shuffle': return src!.shuffle();
		case 'publicShuffle': return src!.publicShuffle();
		case 'swap': return src!.swapComponents(op.Index, op.Slot);
		case 'expand': return src!.expandSize(op.Index);
		case 'contract': return src!.contractSize(op.Index);
		case 'setSize': return src!.setSize(op.Index);
		case 'sizeToFit': return src!.sizeToFit();
		case 'sortAsc': return src!.sortComponents((a, b) => a !== EMPTY_SLOT && (b === EMPTY_SLOT || a < b));
		case 'sortTies': return src!.sortComponents((a, b) => a !== EMPTY_SLOT && (b === EMPTY_SLOT || a % 3 < b % 3));
		case 'sortDesc': return src!.sortComponents((a, b) => a !== EMPTY_SLOT && (b === EMPTY_SLOT || a > b));
	}
	throw new Error('unknown op ' + op.Op);
}

/** Every stack's JSON by Go owner path, plus the two merged views. */
function stacksByPath(state: EngineState): Record<string, StackJSON> {
	const out: Record<string, StackJSON> = {};
	for (const stack of state.physicalStacks()) out[stack.path] = stack.toJSON();
	out['Game.All'] = state.view({ Group: 'game', Prop: 'All' }).toJSON();
	out['Game.Overlap'] = state.view({ Group: 'game', Prop: 'Overlap' }).toJSON();
	return out;
}

test('vector file has the expected breadth', () => {
	assert.equal(Scenarios.length, 20);
	const kinds = new Set<string>();
	let errors = 0;
	for (const sc of Scenarios) for (const op of sc.Ops) {
		kinds.add(op.Op);
		if (op.Err) errors++;
	}
	assert.equal(kinds.size, 19);
	assert.ok(errors > 200, 'error paths are exercised');
});

test('a recorded Go state loads and serializes back to the same JSON', () => {
	for (const sc of Scenarios) {
		assert.deepEqual(loadScenario(sc).toJSON(), sc.Before, sc.GameID);
		const after = EngineState.load(def, { gameId: sc.GameID, secretSalt: sc.SecretSalt }, sc.Version, sc.After);
		assert.deepEqual(after.toJSON(), sc.After, sc.GameID + ' after');
	}
});

test('every operation matches Go: changed stacks, counts, errors, and no change on error', () => {
	let steps = 0;
	for (const sc of Scenarios) {
		const state = loadScenario(sc);
		let expectedStacks = stacksByPath(state);
		let expectedCounts = cloneJSON(state.data.SecretMoveCount ?? {});
		sc.Ops.forEach((op, i) => {
			const label = `${sc.GameID} op ${i} ${op.Op} ${JSON.stringify(op.Src)} -> ${JSON.stringify(op.Dst)} [${op.Index}, ${op.Slot}]`;
			let thrown: unknown = null;
			try {
				apply(state, op);
			} catch (e) {
				thrown = e;
			}
			if (op.NoComponent) {
				assert.ok(thrown instanceof StackError, label + ': expected a refusal for a missing component');
			} else if (op.Err !== '') {
				assert.ok(thrown instanceof StackError, label + ': expected error ' + op.Err);
				assert.equal((thrown as Error).message, op.Err, label);
			} else {
				assert.equal(thrown, null, label + ': unexpected error ' + (thrown as Error | null)?.message);
			}
			expectedStacks = { ...expectedStacks, ...(op.Changed ?? {}) };
			if (op.SecretMoveCount) expectedCounts = op.SecretMoveCount;
			assert.deepEqual(stacksByPath(state), expectedStacks, label);
			assert.deepEqual(state.data.SecretMoveCount ?? {}, expectedCounts ?? {}, label);
			assert.equal(state.conservationError(), null, label);
			steps++;
		});
		assert.deepEqual(state.toJSON(), sc.After, sc.GameID + ' final state');
	}
	assert.equal(steps, 20 * 50);
});

test('shuffles within one state continue one generator sequence', () => {
	// Two shuffles at one version must differ from two states each shuffled
	// once: Go memoizes state.Rand() per state.
	const sc = Scenarios[0];
	const a = loadScenario(sc);
	const before = a.gameStack('Draw').indexes();
	assert.ok(a.rand() instanceof GoRand);
	assert.equal(a.rand(), a.rand(), 'state.Rand() is memoized');
	a.gameStack('Draw').publicShuffle();
	const once = a.gameStack('Draw').indexes();
	a.gameStack('Draw').publicShuffle();
	const twice = a.gameStack('Draw').indexes();
	const b = loadScenario(sc);
	b.gameStack('Draw').publicShuffle();
	assert.deepEqual(b.gameStack('Draw').indexes(), once);
	const c = b.copy();
	c.gameStack('Draw').publicShuffle();
	// The copy has a fresh generator, so it repeats the first permutation
	// rather than continuing to the second.
	assert.deepEqual(c.gameStack('Draw').indexes(), once.map((_, i) => once[permOf(before, once)[i]]));
	assert.notDeepEqual(twice, c.gameStack('Draw').indexes());
});

function permOf(before: number[], after: number[]): number[] {
	return after.map(v => before.indexOf(v));
}

test('copy is independent, and copy(version) reseeds the generator', () => {
	const sc = Scenarios[1];
	const a = loadScenario(sc);
	const b = a.copy(sc.Version + 1);
	b.gameStack('Draw').shuffle();
	assert.deepEqual(a.toJSON(), sc.Before);
	assert.notDeepEqual(b.toJSON(), sc.Before);
	assert.equal(b.version, sc.Version + 1);
	const ids = b.gameStack('Draw').ids();
	const seen = b.gameStack('Draw').idsLastSeen();
	for (const id of ids) assert.equal(seen[id], sc.Version + 1, 'scrambled ids are seen at the new version');
});

test('a sanitized state refuses modification with Go\'s message', () => {
	const sc = Scenarios[0];
	const state = loadScenario(sc).copySanitized();
	assert.throws(() => state.gameStack('Draw').shuffle(), {
		name: 'StackError',
		message: "Modifications not allowed: stack's state is sanitized",
	});
	assert.throws(() => state.gameStack('Draw').moveComponent(0, state.gameStack('Limited'), 0), {
		message: "Source doesn't allow modifications: Modifications not allowed: stack's state is sanitized",
	});
});

test('stacks from another state are rejected', () => {
	const a = loadScenario(Scenarios[0]);
	const b = loadScenario(Scenarios[0]);
	assert.throws(() => a.gameStack('Draw').moveComponent(0, b.gameStack('Limited'), 0), {
		message: 'Source and destination belong to different states',
	});
	assert.throws(() => a.gameStack('Draw').moveAllTo(b.gameStack('Limited')), {
		message: 'source and destination belong to different states',
	});
});

test('merged views are views, and reject being used as physical stacks', () => {
	const state = loadScenario(Scenarios[0]);
	assert.throws(() => state.gameStack('All'), /is a view, not a physical stack/);
	const all = state.view({ Group: 'game', Prop: 'All' });
	assert.ok(all instanceof MergedStackView);
	assert.ok(state.view({ Group: 'game', Prop: 'Draw' }) instanceof Stack);
	const draw = state.gameStack('Draw');
	const limited = state.gameStack('Limited');
	draw.moveComponent(0, limited, 0);
	assert.equal(all.len(), draw.len() + limited.len());
	assert.deepEqual(all.indexes(), [...draw.indexes(), ...limited.indexes()]);
	assert.equal(all.maxSize(), 3);
});

test('an overlap view shows the first part that has a component in each slot', () => {
	const state = loadScenario(Scenarios[0]);
	const draw = state.gameStack('Draw');
	const row = state.gameStack('Row');
	const under = state.gameStack('Under');
	const first = draw.indexes()[0];
	const second = draw.indexes()[1];
	const third = draw.indexes()[2];
	draw.moveComponent(0, under, 1);
	draw.moveComponent(0, row, 1);
	draw.moveComponent(0, under, 3);
	const overlap = state.view({ Group: 'game', Prop: 'Overlap' });
	assert.deepEqual(overlap.indexes(), [-1, second, -1, third]);
	assert.equal(overlap.numComponents(), 2);
	assert.equal(overlap.slotsRemaining(), 2);
	assert.equal(overlap.isSized(), true);
	assert.equal(overlap.toJSON().Size, 4);
	assert.equal(under.indexes()[1], first);
});

test('conservation: Go\'s wording for each violation', () => {
	const sc = Scenarios[0];
	const identity = { gameId: sc.GameID, secretSalt: sc.SecretSalt };
	const mutate = (f: (json: StateJSON) => void): string | null => {
		const json = cloneJSON(sc.Before);
		f(json);
		return EngineState.load(def, identity, sc.Version, json, { skipConservation: true }).conservationError();
	};
	const draw = (json: StateJSON) => json.Game.Draw as StackJSON;
	assert.equal(mutate(() => undefined), null);
	const firstIndex = draw(sc.Before).Indexes![0];
	assert.equal(
		mutate(json => { draw(json).Indexes!.shift(); }),
		`deck "cards" component ${firstIndex} is missing from every attached stack`,
	);
	assert.equal(
		mutate(json => { (json.Game.Limited as StackJSON).Indexes!.push(firstIndex); }),
		`deck "cards" component ${firstIndex} appears at both Game.Draw[0] and Game.Limited[0]`,
	);
	assert.equal(
		mutate(json => { draw(json).Indexes![0] = -1; }),
		'Game.Draw[0] contains the empty-slot sentinel in a growable stack',
	);
	assert.equal(
		mutate(json => { draw(json).Indexes![0] = -2; }),
		'Game.Draw[0] contains a generic component in an unsanitized state',
	);
	assert.equal(
		mutate(json => { draw(json).Indexes![0] = 99; }),
		'Game.Draw[0] contains invalid component index 99 for deck "cards"',
	);
	assert.equal(
		mutate(json => { (json.Game.Tokens as StackJSON).Deck = 'nope'; }),
		'Game.Tokens names unknown deck "nope"',
	);
	assert.throws(
		() => EngineState.load(def, identity, sc.Version, (() => { const j = cloneJSON(sc.Before); draw(j).Indexes!.shift(); return j; })()),
		/Loaded state violated component conservation: deck "cards" component/,
	);
});

test('a sized stack whose Indexes disagree with Size does not load', () => {
	const sc = Scenarios[0];
	const json = cloneJSON(sc.Before);
	(json.Game.Row as StackJSON).Indexes!.push(-1);
	assert.throws(
		() => EngineState.load(def, { gameId: sc.GameID, secretSalt: sc.SecretSalt }, sc.Version, json),
		/didn't agree with size/,
	);
});

test('containingStack finds each component exactly where conservation puts it', () => {
	const state = loadScenario(Scenarios[2]);
	for (let i = 0; i < def.Decks.cards.Size; i++) {
		const found = state.containingStack('cards', i);
		assert.ok(found);
		assert.equal(found.stack.indexes()[found.slotIndex], i);
	}
	assert.equal(state.containingStack('cards', -2), null);
});
