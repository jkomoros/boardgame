// Sanitization against real Go. testdata/sanitize.vectors.json holds, for
// many states x viewers x policy configurations of the vector game, the view
// Go's Game.JSONForPlayer produced (testdata/gen/sanitizevec).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { EngineState, cloneJSON } from './state.ts';
import type { StateJSON } from './state.ts';
import type { GameDefinition, PolicyName, SubStateDef } from './definition.ts';
import { parseSanitizeTag, policyFromString } from './definition.ts';
import {
	sanitizedForPlayer, viewFor, resolveSanitizationPolicy, sanitizationTransformation, facetSurvives,
	SanitizeError, OBSERVER_PLAYER_INDEX, ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX,
} from './sanitize.ts';

const read = (name: string) => JSON.parse(readFileSync(new URL('./testdata/' + name, import.meta.url), 'utf8'));
const baseDef = read('defs/vecgame.def.json') as GameDefinition;

type PolicyMaps = Record<string, Record<string, PolicyName>>;
interface Outcome { Err?: string; Panic?: boolean; View?: StateJSON }
interface Case {
	Label: string;
	State: number;
	Viewer: number;
	Policies: { Game: PolicyMaps; Player: PolicyMaps; Dynamic: Record<string, PolicyMaps> };
	Outcomes: Outcome[];
}
interface VecState { GameID: string; SecretSalt: string; Version: number; NumPlayers: number; State: StateJSON }
const vectors = read('sanitize.vectors.json') as { States: VecState[]; Cases: Case[] };

function overlay(sub: SubStateDef, maps: PolicyMaps | undefined): SubStateDef {
	const props = { ...sub.Props };
	for (const name of Object.keys(maps ?? {})) props[name] = { ...props[name], Sanitize: maps![name] };
	return { Props: props };
}

function defFor(c: Case): GameDefinition {
	return {
		...baseDef,
		Game: overlay(baseDef.Game, c.Policies.Game),
		Player: overlay(baseDef.Player, c.Policies.Player),
		DynamicComponentValues: { cards: overlay(baseDef.DynamicComponentValues.cards, (c.Policies.Dynamic ?? {}).cards) },
	};
}

function stateFor(c: Case): EngineState {
	const vs = vectors.States[c.State];
	return EngineState.load(defFor(c), { gameId: vs.GameID, secretSalt: vs.SecretSalt }, vs.Version, vs.State);
}

const withoutComponents = (view: StateJSON) => {
	const copy = cloneJSON(view) as Partial<StateJSON>;
	delete copy.Components;
	return copy;
};
const sameJSON = (a: unknown, b: unknown) => {
	try {
		assert.deepStrictEqual(a, b);
		return true;
	} catch {
		return false;
	}
};

test('vector file has the expected breadth', () => {
	assert.equal(vectors.States.length, 3);
	assert.ok(vectors.Cases.length > 150, String(vectors.Cases.length));
	const viewers = new Set(vectors.Cases.map(c => c.Viewer));
	for (const v of [ANY_PLAYER_INDEX, ADMIN_PLAYER_INDEX, OBSERVER_PLAYER_INDEX, 0, 1, 2]) assert.ok(viewers.has(v), 'viewer ' + v);
});

test('every case where Go has one answer: viewFor matches it exactly', () => {
	let checked = 0;
	for (const c of vectors.Cases) {
		if (c.Outcomes.length !== 1 || !c.Outcomes[0].View) continue;
		const label = `${c.Label} / state ${c.State} / viewer ${c.Viewer}`;
		assert.deepEqual(viewFor(stateFor(c), c.Viewer), c.Outcomes[0].View, label);
		assert.deepEqual(
			viewFor(stateFor(c), c.Viewer, { mergedStacksRevealDynamicValues: true }),
			c.Outcomes[0].View,
			label + ' (reveal mode agrees whenever Go is deterministic)',
		);
		checked++;
	}
	assert.ok(checked > 100, String(checked));
});

test('cases where Go\'s answer depends on map order: both documented modes', () => {
	// A visible merged stack over non-visible stacks of a deck with dynamic
	// component values. Go either reveals those values (merged stack examined
	// first) or reveals nothing / panics (examined last).
	let unstable = 0;
	let defaultAmongGo = 0;
	for (const c of vectors.Cases) {
		if (c.Outcomes.length < 2) continue;
		unstable++;
		const label = `${c.Label} / state ${c.State} / viewer ${c.Viewer}`;
		const goViews = c.Outcomes.filter(o => o.View).map(o => o.View!);
		assert.ok(goViews.length >= 1, label);

		const revealing = viewFor(stateFor(c), c.Viewer, { mergedStacksRevealDynamicValues: true });
		assert.ok(goViews.some(v => sameJSON(v, revealing)), label + ': reveal mode equals one of Go\'s answers');

		const conservative = viewFor(stateFor(c), c.Viewer);
		for (const v of goViews) {
			assert.deepEqual(withoutComponents(conservative), withoutComponents(v), label + ': identical outside dynamic values');
		}
		if (goViews.some(v => sameJSON(v, conservative))) defaultAmongGo++;

		// The default never shows a dynamic value that reveal mode hides.
		const cons = conservative.Components.cards;
		const rev = revealing.Components.cards;
		for (let i = 0; i < cons.length; i++) {
			if (!sameJSON(cons[i], rev[i])) {
				assert.equal(cons[i].Counter, 0, label + `: card ${i} is hidden in the default mode`);
			}
		}
	}
	assert.ok(unstable > 20, String(unstable));
	assert.ok(defaultAmongGo > 0, 'the default mode is itself one of Go\'s answers where Go does not panic');
});

test('errors and panics: the engine refuses where Go fails every time', () => {
	let checked = 0;
	for (const c of vectors.Cases) {
		if (c.Outcomes.some(o => o.View)) continue;
		checked++;
		const outcome = c.Outcomes[0];
		assert.throws(() => viewFor(stateFor(c), c.Viewer), (err: unknown) => {
			assert.ok(err instanceof SanitizeError, c.Label);
			if (!outcome.Panic) {
				// Game.JSONForPlayer prefixes state.SanitizedForPlayer's error.
				assert.equal("Couldn't sanitize state: " + err.message, outcome.Err, c.Label);
			}
			return true;
		});
	}
	assert.equal(checked, 4);
});

test('the admin and out-of-range viewers get the unsanitized state, with SecretMoveCount', () => {
	const c = vectors.Cases.find(x => x.Viewer === ADMIN_PLAYER_INDEX)!;
	const state = stateFor(c);
	assert.equal(sanitizedForPlayer(state, ADMIN_PLAYER_INDEX), state);
	assert.equal(sanitizedForPlayer(state, state.numPlayers()), state);
	const view = viewFor(state, ADMIN_PLAYER_INDEX);
	assert.ok(view.SecretMoveCount);
	assert.deepEqual(view.Visibility!.Game.Draw, ['values', 'count', 'occupancy', 'order', 'nonempty']);
	const observer = viewFor(state, OBSERVER_PLAYER_INDEX);
	assert.equal(observer.SecretMoveCount, undefined);
	assert.deepEqual(viewFor(state, ANY_PLAYER_INDEX), observer);
});

test('sanitizing does not change the state it is given', () => {
	const c = vectors.Cases.find(x => x.Label === 'game all:hidden')!;
	const state = stateFor(c);
	const before = state.toJSON();
	viewFor(state, 0);
	assert.deepEqual(state.toJSON(), before);
});

test('least restrictive applicable policy wins', () => {
	const map = { all: 'hidden', self: 'len', 'same-color': 'order' } as const;
	assert.equal(resolveSanitizationPolicy(map, { all: true }), 'hidden');
	assert.equal(resolveSanitizationPolicy(map, { all: true, self: true }), 'len');
	assert.equal(resolveSanitizationPolicy(map, { all: true, self: true, 'same-color': true }), 'order');
	assert.equal(resolveSanitizationPolicy(map, { all: true, self: false, 'same-color': false }), 'hidden');
	assert.equal(resolveSanitizationPolicy(map, { other: true }), 'visible', 'no applicable group: fallback');
	assert.equal(resolveSanitizationPolicy(map, { other: true }, 'hidden'), 'hidden');
	assert.equal(resolveSanitizationPolicy({ all: 'invalid', self: 'hidden' }, { all: true, self: true }), 'hidden');
	assert.equal(resolveSanitizationPolicy({ all: 'invalid' }, { all: true }), 'invalid');
});

test('struct-tag parsing matches Go (the exported definition is Go\'s own parse)', () => {
	// testdata/defs/vecgame.def.json records what Go's StructInflater made
	// of each tag in testdata/gen/vecgame/main.go.
	const tags: [SubStateDef, string, string, string][] = [
		[baseDef.Game, 'Draw', '', 'all'],
		[baseDef.Game, 'TagHidden', 'hidden', 'all'],
		[baseDef.Game, 'TagLen', 'len', 'all'],
		[baseDef.Game, 'TagRedOnly', 'all:hidden,Red:visible', 'all'],
		[baseDef.Game, 'TagSpaced', ' Hidden ', 'all'],
		[baseDef.Player, 'Hand', 'len', 'other'],
		[baseDef.Player, 'Slots', 'self:visible,other:order', 'other'],
		[baseDef.Player, 'Secret', 'hidden', 'other'],
		[baseDef.Player, 'TeamSecret', 'all:hidden,same-color:visible', 'other'],
		[baseDef.Player, 'OwnOnly', 'all:nonempty,self:visible', 'other'],
		[baseDef.Player, 'Score', '', 'other'],
	];
	for (const [sub, prop, tag, defaultGroup] of tags) {
		assert.deepEqual(parseSanitizeTag(tag, defaultGroup), sub.Props[prop].Sanitize, prop);
	}
	assert.deepEqual(parseSanitizeTag('a:b:c', 'all'), { all: 'invalid' });
	assert.deepEqual(parseSanitizeTag('self:nonsense', 'other'), { self: 'invalid' });
	// sanitization_test.go TestPolicyFromString
	assert.equal(policyFromString('VisibLE'), 'visible');
	assert.equal(policyFromString('order '), 'order');
	assert.equal(policyFromString('Hidden'), 'hidden');
	assert.equal(policyFromString('visibl'), 'invalid');
});

test('a tag string and its parsed map sanitize identically', () => {
	const c = vectors.Cases.find(x => x.Label === 'struct tags' && x.Viewer === 1)!;
	const vs = vectors.States[c.State];
	const tagged: GameDefinition = cloneJSON(baseDef);
	(tagged.Player.Props.Slots as { Sanitize: unknown }).Sanitize = 'self:visible,other:order';
	(tagged.Player.Props.Secret as { Sanitize: unknown }).Sanitize = 'hidden';
	(tagged.Game.Props.TagRedOnly as { Sanitize: unknown }).Sanitize = 'all:hidden,Red:visible';
	delete (tagged.Game.Props.Draw as { Sanitize?: unknown }).Sanitize;
	const state = EngineState.load(tagged, { gameId: vs.GameID, secretSalt: vs.SecretSalt }, vs.Version, vs.State);
	assert.deepEqual(viewFor(state, 1), c.Outcomes[0].View);
});

test('groups: self, other, enum values and same-/different- are computed per player state', () => {
	const c = vectors.Cases.find(x => x.Label === 'struct tags' && x.Viewer === 0 && vectors.States[x.State].NumPlayers >= 3)!;
	const state = stateFor(c);
	const t = sanitizationTransformation(state, 0);
	const colors = state.players.map(p => p.Color);
	assert.equal(t.Players[0].Slots, 'visible', 'self');
	assert.equal(t.Players[1].Slots, 'order', 'other');
	assert.equal(t.Players[0].Secret, 'visible', 'an untagged group on a player state means "other"');
	assert.equal(t.Players[1].Secret, 'hidden');
	for (let i = 0; i < colors.length; i++) {
		assert.equal(t.Players[i].TeamSecret, colors[i] === colors[0] ? 'visible' : 'hidden', 'same-color ' + i);
		assert.equal(t.Players[i].RivalNote, colors[i] !== colors[0] ? 'visible' : 'hidden', 'different-color ' + i);
	}
	assert.equal(t.Game.TagRedOnly, colors[0] === 'Red' ? 'visible' : 'hidden', 'a group-enum value names the viewer\'s own group');
	const observer = sanitizationTransformation(state, OBSERVER_PLAYER_INDEX);
	assert.equal(observer.Game.TagRedOnly, 'hidden');
	assert.equal(observer.Players[0].Slots, 'order');
});

test('an unknown computed group is an error, as in Go', () => {
	const c = vectors.Cases.find(x => x.Label === 'struct tags')!;
	const vs = vectors.States[c.State];
	const def: GameDefinition = cloneJSON(baseDef);
	(def.Player.Props.Score as { Sanitize: unknown }).Sanitize = { 'sideways-color': 'hidden' };
	const state = EngineState.load(def, { gameId: vs.GameID, secretSalt: vs.SecretSalt }, vs.Version, vs.State);
	assert.throws(() => viewFor(state, 0), /sideways was used as a computed group name function/);
	(def.Player.Props.Score as { Sanitize: unknown }).Sanitize = { nonsense: 'hidden' };
	const state2 = EngineState.load(def, { gameId: vs.GameID, secretSalt: vs.SecretSalt }, vs.Version, vs.State);
	assert.throws(() => viewFor(state2, 0), /Unsupported group name: nonsense/);
});

test('facet truth table (legal_path.go facetSurvives)', () => {
	const row = (p: PolicyName) => (['values', 'count', 'occupancy', 'order', 'nonempty'] as const).filter(f => facetSurvives(p, f));
	assert.deepEqual(row('visible'), ['values', 'count', 'occupancy', 'order', 'nonempty']);
	assert.deepEqual(row('order'), ['count', 'occupancy', 'order', 'nonempty']);
	assert.deepEqual(row('len'), ['count', 'nonempty']);
	assert.deepEqual(row('nonempty'), ['nonempty']);
	assert.deepEqual(row('hidden'), []);
});

test('a sanitized state cannot be sanitized again or modified', () => {
	const c = vectors.Cases.find(x => x.Label === 'struct tags' && x.Viewer === 0)!;
	const sanitized = sanitizedForPlayer(stateFor(c), 0);
	assert.equal(sanitized.sanitized, true);
	assert.throws(() => sanitizedForPlayer(sanitized, 0), /already sanitized/);
	assert.throws(() => sanitized.gameStack('Draw').shuffle(), /state is sanitized/);
	assert.equal(sanitized.conservationError(), null, 'sanitized states are exempt from conservation');
});

test('a view Go produced loads as a sanitized state and serializes back unchanged', () => {
	// What a client holds. Merged stacks are recomputed from the sanitized
	// parts, so this also checks the views agree with Go after sanitization.
	let checked = 0;
	for (const c of vectors.Cases) {
		if (c.Outcomes.length !== 1 || !c.Outcomes[0].View || c.Viewer < -1) continue;
		const vs = vectors.States[c.State];
		const view = c.Outcomes[0].View;
		if (c.Viewer >= vs.NumPlayers) continue;
		const state = EngineState.load(defFor(c), { gameId: vs.GameID, secretSalt: '' }, vs.Version, view, { sanitized: true });
		assert.equal(state.sanitized, true);
		assert.deepEqual(state.visibility, view.Visibility);
		assert.deepEqual({ ...state.toJSON(), Visibility: state.visibility }, view, c.Label);
		checked++;
	}
	assert.ok(checked > 80, String(checked));
});
