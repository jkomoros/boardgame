// The Go core's own sanitization fixtures (testdata/sanitize and
// testdata/after_dynamic_component_move at the repository root), consumed the
// way sanitization_test.go's TestSanitization consumes them: compose the
// input and expected blobs with patchtree, load the input as a state of the
// test game with the fixed test id and salt, install the case's policies,
// sanitize for the case's viewer, and compare JSON.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { loadEngineState } from './state.ts';
import type { EngineState } from './state.ts';
import type { StateJSON } from './state.ts';
import type { GameDefinition, PropDef } from './definition.ts';
import { sanitizedForPlayer, ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX } from './sanitize.ts';
import { patchTreeJSON, parseJdPatch, applyJdPatch } from './testing/patchtree.ts';

const TESTDATA = join(fileURLToPath(new URL('.', import.meta.url)), '..', '..', '..', '..', 'testdata');

// testGame(t, stableIds=true) in main_test.go.
const IDENTITY = { gameId: 'FAKEIDFORTESTING', secretSalt: 'FAKESALTFORTESTING' };

// The test game of package boardgame's own tests (main_test.go:
// testGameState, testPlayerState, testingComponentDynamic; enums in
// game_manager_test.go; deck in game_delegate_test.go). It lives in _test.go
// files, so no generator can import it; this is transcribed by hand.
function testGameDefinition(config: SanitizationTestConfig): GameDefinition {
	const tagged = (props: Record<string, PropDef>, tags: Record<string, string> | undefined) => {
		const out: Record<string, PropDef> = {};
		// sanitizationTestConfig.policyForSubObject: every property gets
		// policyFromStructTag(config[propName], defaultGroup).
		for (const name of Object.keys(props)) out[name] = { ...props[name], Sanitize: (tags ?? {})[name] ?? '' };
		return { Props: out };
	};
	return {
		Name: 'testgame',
		Decks: { test: { Size: 5 } },
		Enums: {
			color: { Values: { 0: 'Red', 1: 'Blue', 2: 'Green' }, DefaultValue: 0 },
			phase: {
				Values: { 0: '', 1: 'Set Up', 2: 'Normal', 3: 'Normal > Player Start', 4: 'Normal > Activate Card', 5: 'Scoring' },
				DefaultValue: 1,
			},
		},
		Game: tagged({
			Phase: { Type: 'enum', Enum: 'phase' },
			CurrentPlayer: { Type: 'playerIndex' },
			DrawDeck: { Type: 'stack', Deck: 'test' },
			Timer: { Type: 'timer' },
			MyIntSlice: { Type: 'intSlice' },
			MyStringSlice: { Type: 'stringSlice' },
			MyBoolSlice: { Type: 'boolSlice' },
			MyPlayerIndexSlice: { Type: 'playerIndexSlice' },
			MyEnumValue: { Type: 'enum', Enum: 'color' },
			MyEnumConst: { Type: 'enum', Enum: 'color' },
			DownSizeStack: { Type: 'sizedStack', Deck: 'test', Size: 4 },
			OtherStack: { Type: 'sizedStack', Deck: 'test', Size: 2 },
			MyMergedStack: { Type: 'mergedStack', Deck: 'test', Stacks: ['DownSizeStack', 'OtherStack'] },
			MyBoard: { Type: 'board', Deck: 'test', Size: 3 },
		}, config.Game),
		Player: tagged({
			Score: { Type: 'int' },
			MovesLeftThisTurn: { Type: 'int' },
			Hand: { Type: 'sizedStack', Deck: 'test', Size: 2 },
			IsFoo: { Type: 'bool' },
			EnumVal: { Type: 'enum', Enum: 'color' },
		}, config.Player),
		DynamicComponentValues: {
			test: tagged({
				IntVar: { Type: 'int' },
				Stack: { Type: 'sizedStack', Deck: 'test', Size: 1 },
				Enum: { Type: 'enum', Enum: 'color' },
			}, (config.DynamicComponentValues ?? {}).test),
		},
	};
}

// testGameDelegate.ConfigureComputedProperties (game_delegate_test.go).
function computed(state: EngineState): unknown {
	let sum = 0;
	for (const player of state.players) sum += player.Score as number;
	return {
		Global: { SumAllScores: sum },
		Players: state.players.map(player => ({
			EffectiveMovesLeftThisTurn: (player.MovesLeftThisTurn as number) + (player.IsFoo ? 5 : 0),
		})),
	};
}

interface SanitizationTestConfig {
	Game?: Record<string, string>;
	Player?: Record<string, string>;
	DynamicComponentValues?: Record<string, Record<string, string>>;
}
interface Case {
	policy: SanitizationTestConfig;
	playerIndex: number;
	inputPatchTree: string;
	expectedPatchTree: string;
}

const gameSlices = (policy: string) => ({
	DrawDeck: policy, MyIntSlice: policy, MyBoolSlice: policy, MyStringSlice: policy, MyPlayerIndexSlice: policy, MyBoard: policy,
});

// The table in TestSanitization, in order.
const cases: Case[] = [
	{ policy: {}, playerIndex: ADMIN_PLAYER_INDEX, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize' },
	{ policy: { Game: gameSlices('len') }, playerIndex: ANY_PLAYER_INDEX, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/len' },
	{ policy: { Game: gameSlices('len') }, playerIndex: 0, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/len' },
	{ policy: { Player: { Hand: 'len' } }, playerIndex: 0, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/len_player' },
	{ policy: { Game: gameSlices('order') }, playerIndex: 0, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/order' },
	{ policy: { Player: { Hand: 'order' } }, playerIndex: 0, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/order_player' },
	{
		policy: {
			Game: { ...gameSlices('hidden'), MyEnumValue: 'hidden' },
			Player: { Hand: 'all:hidden,self:visible', MovesLeftThisTurn: 'all:hidden', IsFoo: 'all:hidden', EnumVal: 'all:hidden' },
		},
		playerIndex: 0, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/hidden',
	},
	{
		policy: { Game: gameSlices('nonempty'), Player: { Hand: 'all:nonempty' } },
		playerIndex: 0, inputPatchTree: 'sanitize', expectedPatchTree: 'sanitize/nonempty',
	},
	{
		policy: { DynamicComponentValues: { test: { IntVar: 'hidden' } } },
		playerIndex: 0, inputPatchTree: 'after_dynamic_component_move',
		expectedPatchTree: 'after_dynamic_component_move/sanitization_with_dynamic_state',
	},
	{
		policy: { DynamicComponentValues: { test: { IntVar: 'hidden' } }, Player: { Hand: 'hidden' } },
		playerIndex: 1, inputPatchTree: 'after_dynamic_component_move',
		expectedPatchTree: 'after_dynamic_component_move/sanitization_with_dynamic_state_sanitized',
	},
	{
		policy: { Game: { DrawDeck: 'len' }, Player: { Hand: 'len' } },
		playerIndex: 0, inputPatchTree: 'after_dynamic_component_move',
		expectedPatchTree: 'after_dynamic_component_move/sanitization_with_dynamic_state_transitive',
	},
];

/** json.MarshalIndent(sanitizedState): state.MarshalJSON, i.e. with Computed. */
function marshal(state: EngineState): StateJSON {
	const out = state.toJSON();
	delete out.Timers;
	out.Computed = computed(state);
	return out;
}

test('patchtree: jd patches apply as the Go library applies them', () => {
	const doc = { a: [1, 2, 3, 4], b: { c: 'x' } };
	const patch = parseJdPatch('@ ["a",3]\n- 4\n@ ["a",1]\n- 2\n+ 9\n@ ["a",0]\n+ 7\n@ ["b","c"]\n- "x"\n+ "y"\n@ ["b","d"]\n+ {"e":1}\n');
	assert.deepEqual(applyJdPatch(doc, patch), { a: [7, 1, 9, 3], b: { c: 'y', d: { e: 1 } } });
	assert.throws(() => applyJdPatch({ a: 1 }, parseJdPatch('@ ["a"]\n- 2\n+ 3\n')), /expected to remove/);
});

test('every fixture directory under testdata/sanitize is exercised', () => {
	const used = new Set(cases.map(c => c.expectedPatchTree));
	for (const dir of ['sanitize', 'sanitize/hidden', 'sanitize/len', 'sanitize/len_player', 'sanitize/nonempty', 'sanitize/order', 'sanitize/order_player']) {
		assert.ok(used.has(dir), dir);
	}
	assert.equal(cases.length, 11);
});

for (const [i, c] of cases.entries()) {
	test(`TestSanitization ${i}: ${c.inputPatchTree} -> ${c.expectedPatchTree} for player ${c.playerIndex}`, () => {
		const input = patchTreeJSON(join(TESTDATA, c.inputPatchTree)) as unknown as StateJSON;
		const expected = patchTreeJSON(join(TESTDATA, c.expectedPatchTree)) as unknown as StateJSON;
		// "version doesn't matter in this context": the Go test loads at 0.
		const state = loadEngineState(testGameDefinition(c.policy), IDENTITY, 0, input);
		const sanitized = sanitizedForPlayer(state, c.playerIndex);
		assert.deepEqual(marshal(sanitized), expected);
	});
}

test('the fixtures\' recorded ids recompute from the test id and salt', () => {
	// Loading discards recorded ids and recomputes them; serializing an
	// unsanitized fixture back therefore checks every id in it.
	for (const tree of ['sanitize', 'after_dynamic_component_move', 'after_move', 'no_op']) {
		const input = patchTreeJSON(join(TESTDATA, tree)) as unknown as StateJSON;
		const state = loadEngineState(testGameDefinition({}), IDENTITY, 0, input);
		assert.deepEqual(marshal(state), input, tree);
	}
});
