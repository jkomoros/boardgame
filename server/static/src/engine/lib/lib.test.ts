// lib/ against real Go (testdata/lib.vectors.json, from testdata/gen/libvec):
// range enums, grid graphs, dice and playing cards answer what the Go
// packages answer. Also that the hand ports that inlined these libraries
// (games/pig.ts's roll and die, games/blackjack.ts's deck and enums) agree
// with them.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import type { EnumDef } from '../definition.ts';
import { newGoRand } from '../rand.ts';
import { blackjack } from '../games/blackjack.ts';
import { pig } from '../games/pig.ts';
import { loadState } from '../rules/engine.ts';
import { type CompiledGame, type GameSnapshot, compileGame, createGame, newRulesState } from '../rules/index.ts';
import { basicDie, defaultDie, dieMax, dieMin, roll, rollWith } from './dice.ts';
import {
	ILLEGAL_VALUE, enumValueFromString, rangeDimensions, rangeEnumDef, rangeToValue, rangeValue, setRangeValueError, valueToRange,
} from './enumrange.ts';
import { connectedByKey, gridConnectedness, neighborsByKey } from './graph.ts';
import { cardEnums, cardString, genericCard, newDeck, newDeckMulti, rankKey, suitKey } from './playingcards.ts';

interface Vectors {
	Ranges: {
		Name: string;
		Dimensions: number[];
		Values: Record<string, string>;
		DefaultValue: number;
		RangeDimensions: number[];
		ValueToRange: { Key: number; Indexes: number[] | null }[];
		RangeToValue: { Indexes: number[]; Key: number; SetError: string }[];
		ValueFromString: { Input: string; Key: number }[];
		MarshalledEnum: { Values: Record<string, string>; DefaultValue: number; Dimensions: number[] };
	}[];
	RangeErrors: { Dimensions: number[]; Error: string }[];
	Graphs: { Name: string; Dimensions: number[]; Directions: string[] | null; Error?: string; Neighbors?: Record<string, number[]>; Connected?: [number, number, number][] }[];
	Dice: { Min: number; Max: number; Faces: number[] | null; DieMin: number; DieMax: number; HasValue: boolean }[];
	Rolls: { Seed: number; Faces: number[]; Steps: { Value: number; SelectedFace: number; RollCount: number }[] }[];
	Decks: { Label: string; Count: number; WithJokers: boolean; Cards: { Suit: string; Rank: string }[]; Strings: string[]; Generic: { Suit: string; Rank: string } }[];
	Enums: Record<string, { Values: Record<string, string>; DefaultValue: number }>;
}

const vectors = JSON.parse(readFileSync(fileURLToPath(new URL('../testdata/lib.vectors.json', import.meta.url)), 'utf8')) as Vectors;

test('range enums: values, dimensions and the default match AddRange', () => {
	assert.ok(vectors.Ranges.length >= 6);
	for (const v of vectors.Ranges) {
		const def = rangeEnumDef(v.Dimensions);
		assert.deepEqual(def.Values, v.Values, v.Name);
		assert.equal(def.DefaultValue, v.DefaultValue, v.Name);
		assert.deepEqual(rangeDimensions(def), v.RangeDimensions, v.Name);
		// The enum as Go marshals it carries the same values and dimensions.
		assert.deepEqual(def.Values, v.MarshalledEnum.Values);
		assert.deepEqual(def.Range, v.MarshalledEnum.Dimensions);
	}
});

test('range enums: ValueToRange, RangeToValue, SetRangeValue, ValueFromString', () => {
	let count = 0;
	for (const v of vectors.Ranges) {
		const def = rangeEnumDef(v.Dimensions);
		for (const c of v.ValueToRange) {
			assert.deepEqual(valueToRange(def, c.Key), c.Indexes, v.Name + ' ValueToRange(' + c.Key + ')');
			assert.deepEqual(rangeValue(def, c.Key), c.Indexes);
			count++;
		}
		for (const c of v.RangeToValue) {
			assert.equal(rangeToValue(def, c.Indexes), c.Key, v.Name + ' RangeToValue(' + c.Indexes.join(',') + ')');
			assert.equal(setRangeValueError(def, c.Indexes) ?? '', c.SetError, v.Name + ' SetRangeValue(' + c.Indexes.join(',') + ')');
			count++;
		}
		for (const c of v.ValueFromString) {
			assert.equal(enumValueFromString(def, c.Input), c.Key, v.Name + ' ValueFromString(' + JSON.stringify(c.Input) + ')');
			count++;
		}
	}
	assert.ok(count >= 500, 'cases: ' + count);
	assert.equal(ILLEGAL_VALUE, 9223372036854775807);
});

test('range enums: AddRange refuses what Go refuses, in its words', () => {
	for (const v of vectors.RangeErrors) assert.throws(() => rangeEnumDef(v.Dimensions), { message: v.Error });
});

test('grid connectedness: every node has the neighbors Go gives it', () => {
	let nodes = 0;
	for (const v of vectors.Graphs) {
		const def: EnumDef = rangeEnumDef(v.Dimensions);
		const label = v.Name + ' ' + v.Dimensions.join('x') + ' ' + (v.Directions ?? []).join('+');
		if (v.Error !== undefined) {
			assert.throws(() => gridConnectedness(def, v.Directions ?? []), { message: v.Error }, label);
			continue;
		}
		const graph = gridConnectedness(def, v.Directions ?? []);
		for (const key of Object.keys(v.Neighbors ?? {})) {
			assert.deepEqual(neighborsByKey(graph, Number(key)), v.Neighbors![key], label + ' node ' + key);
			nodes++;
		}
		for (const [from, to, connected] of v.Connected ?? []) assert.equal(connectedByKey(graph, from, to), connected === 1, label + ' ' + from + '->' + to);
	}
	assert.ok(nodes > 300, 'nodes: ' + nodes);
});

test('grid connectedness: short and Go direction names are the same filter', () => {
	const def = rangeEnumDef([8, 8]);
	assert.deepEqual(gridConnectedness(def, ['graph.DirectionDown', 'DirectionDiagonal']), gridConnectedness(def, ['down', 'diagonal']));
	assert.throws(() => gridConnectedness(def, ['sideways']), /unknown direction/);
});

test('dice: BasicDie, Min and Max', () => {
	for (const v of vectors.Dice) {
		if (v.Min === 0 && v.Max === 0) {
			assert.equal(dieMin({ Faces: [] }), v.DieMin);
			assert.equal(dieMax({ Faces: [] }), v.DieMax);
			continue;
		}
		const die = v.Min === 1 && v.Max === 6 && v.HasValue && v.Faces !== null ? basicDie(1, 6) : basicDie(v.Min, v.Max);
		if (!v.HasValue) {
			assert.equal(die, null, 'BasicDie(' + v.Min + ', ' + v.Max + ')');
			continue;
		}
		assert.deepEqual(die?.Faces, v.Faces);
		assert.equal(dieMin(die!), v.DieMin);
		assert.equal(dieMax(die!), v.DieMax);
	}
	assert.deepEqual(defaultDie(), basicDie(1, 6));
});

test('dice: Roll draws from the generator as Go does', () => {
	for (const v of vectors.Rolls) {
		const rand = newGoRand(v.Seed);
		const dynamic: Record<string, unknown> = { Value: 0, SelectedFace: 0, RollCount: 0 };
		for (const [i, step] of v.Steps.entries()) {
			rollWith(rand, { Faces: v.Faces }, dynamic);
			assert.deepEqual(dynamic, step, 'seed ' + v.Seed + ' roll ' + (i + 1));
		}
	}
});

test('playing cards: every deck, its strings and its generic card', () => {
	for (const v of vectors.Decks) {
		const cards = v.Label.startsWith('NewDeckMulti') ? newDeckMulti(v.Count, v.WithJokers) : newDeck(v.WithJokers);
		assert.deepEqual(cards, v.Cards, v.Label);
		assert.deepEqual(cards.map(cardString), v.Strings, v.Label);
		assert.deepEqual(genericCard(), v.Generic, v.Label);
		for (const card of cards) {
			assert.equal(cardEnums().Suit.Values[String(suitKey(card.Suit))], card.Suit);
			assert.equal(cardEnums().Rank.Values[String(rankKey(card.Rank))], card.Rank);
		}
	}
	assert.deepEqual(cardEnums(), vectors.Enums);
});

test("blackjack's inlined deck and enums are playingcards'", () => {
	const spec = blackjack();
	assert.deepEqual(spec.decks.cards.components, newDeck(false));
	const enums = { ...(spec.enums ?? {}), ...(spec.componentEnums ?? {}) };
	assert.deepEqual(enums.Suit, cardEnums().Suit);
	assert.deepEqual(enums.Rank, cardEnums().Rank);
});

test("pig's inlined die and roll are dice's", () => {
	const spec = pig();
	assert.deepEqual(spec.decks.dice.components, [defaultDie()]);
	// Apply pig's own Roll Dice and dice.roll to two copies of one state:
	// the same draw from the same generator gives the same die.
	const game = compileGame(spec);
	for (const seed of ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']) {
		const created = createGame(game, { id: 'libtest-' + seed, secretSalt: 'salt-' + seed, numPlayers: 2, variant: {}, seating: false, now: 0 });
		const snapshot: GameSnapshot = {
			record: created.record!,
			history: created.versions.flatMap(v => (v.move ? [v.move] : [])),
			state: created.versions[created.versions.length - 1].state,
		};
		const viaPig = runRoll(game, snapshot, 'pig');
		const viaLib = runRoll(game, snapshot, 'lib');
		assert.deepEqual(viaLib, viaPig, 'seed ' + seed);
		assert.equal((viaLib as { RollCount: number }).RollCount, 1);
	}
});

function runRoll(game: CompiledGame, snapshot: GameSnapshot, how: 'pig' | 'lib'): unknown {
	const loaded = loadState(game, snapshot);
	const rules = newRulesState(game, loaded.engine.copy(snapshot.history.length + 1), snapshot.record, snapshot.history, { pendingSeat: null }, loaded);
	if (how === 'pig') {
		const move = game.movesByName['Roll Dice'];
		move.spec.apply(rules, { name: 'Roll Dice', fields: { TargetPlayerIndex: rules.currentPlayerIndex() } });
	} else {
		roll(rules, 'dice', 0);
	}
	return { ...rules.dynamic('dice', 0) };
}
