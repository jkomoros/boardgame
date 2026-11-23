// The loader (rules/data.ts) on its own contracts, with the generated games'
// translated Go functions as bodies (games/<game>/bodies.gen.ts).
//
// Two sources of GameData:
//
//  - conformance/definitions/<game>.json, written by the Go exporter
//    (`boardgame-util gamedef export`). The generated game IS that
//    definition with those bodies, and games/<game>/<game>.gen.test.ts holds
//    it to the whole corpus; here the same definition is loaded again with
//    every move supplied whole through the `move:` hook, and must compile to
//    the same header tables, plans and state definition as the generated
//    game and pass the conformance corpus.
//  - rules/testdata/<game>.gamedata.json: stand-ins written for this test
//    before the exporter existed, for pig and blackjack. Their Definition
//    equals Go's DefinitionFor output (testdata/defs/<game>.def.json) plus
//    the GameData additions, and each move's name, help text, fix-up flag,
//    legal phases, progression, AllowMultipleInProgression and plan equal
//    what the live Go manager reports (rules/testdata/movefacts/<game>.json,
//    from rules/testdata/gen/movefacts); both are checked below. The rest
//    (Stock, Base, Fields, Bodies FnRefs, Traits, Initial, flags) is written
//    by hand from the Go source and is held by compiling like the generated
//    game and by the corpus. A fixture move whose stock configuration the
//    registry cannot build, or builds without the Apply that Go's exported
//    configuration yields (the fixtures lack Go's boot-time inferences), gets
//    the registry's parts built from Go's exported configuration of the same
//    move, through the `stock:<name>` hook; a second pass supplies moves
//    whole through `move:` instead, so both hooks are exercised.
//
// Until the six hand ports (games/<game>.ts) were replaced by the generated
// games, the bodies here were the hand ports' functions, adapted by role;
// fabric/transpile/status/replacement.md says what that adapter proved and
// where the record of those functions now is (the pair store).
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { policyMapFor } from '../definition.ts';
import type { DeckDef, GameDefinition, SubStateDef } from '../definition.ts';
import { bodies as blackjackBodies } from '../games/blackjack/bodies.gen.ts';
import { blackjack } from '../games/blackjack/index.ts';
import { bodies as checkersBodies } from '../games/checkers/bodies.gen.ts';
import { bodies as debuganimationsBodies } from '../games/debuganimations/bodies.gen.ts';
import { debuganimations } from '../games/debuganimations/index.ts';
import { bodies as memoryBodies } from '../games/memory/bodies.gen.ts';
import { memory } from '../games/memory/index.ts';
import { bodies as pigBodies } from '../games/pig/bodies.gen.ts';
import { pig } from '../games/pig/index.ts';
import { bodies as tictactoeBodies } from '../games/tictactoe/bodies.gen.ts';
import { tictactoe } from '../games/tictactoe/index.ts';
import { bodies as werewolfBodies } from '../games/werewolf/bodies.gen.ts';
import { werewolf } from '../games/werewolf/index.ts';
import { GENERIC_COMPONENT } from '../ids.ts';
import { loadEngineState } from '../state.ts';
import { type ConformanceDocument, FIELDS, firstDifference, runDocument } from '../testing/conformance.ts';
import { compileGame, emptyStateJSON, gameDisplayName, isGameSpecError, variantConfig } from './compile.ts';
import { BASE_DISTRIBUTE_ERROR, type Bodies, type GameData, type MoveData, callStockApply, gameFromData, moveOverrideRef, stockOverrideRef } from './data.ts';
import { type Counter, type ProgressionGroup, evalCounter } from './progression.ts';
import type { CompiledGame, RulesState } from './state.ts';
import { newRulesState } from './state.ts';
import { type StockParts, stockMove } from './stockRegistry.ts';
import type { GameSpec, MoveSpec } from './types.ts';

/** The six formerly hand-ported games: the generated game and its bodies table. */
const GENERATED: Record<string, { make: () => GameSpec; bodies: Bodies }> = {
	memory: { make: memory, bodies: memoryBodies },
	blackjack: { make: blackjack, bodies: blackjackBodies },
	tictactoe: { make: tictactoe, bodies: tictactoeBodies },
	pig: { make: pig, bodies: pigBodies },
	debuganimations: { make: debuganimations, bodies: debuganimationsBodies },
	werewolf: { make: werewolf, bodies: werewolfBodies },
};
const FIXTURES = ['pig', 'blackjack'];

function repoRoot(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'index.json'))) return dir;
		const parent = path.dirname(dir);
		assert.notEqual(parent, dir, 'could not find conformance/ above ' + import.meta.url);
		dir = parent;
	}
}

const root = repoRoot();
const here = path.dirname(fileURLToPath(import.meta.url));
const corpusIndex = JSON.parse(readFileSync(path.join(root, 'conformance', 'index.json'), 'utf8')) as { file: string; game: string }[];

function readJSON<T>(file: string): T {
	return JSON.parse(readFileSync(file, 'utf8')) as T;
}

const exportedDefinition = (name: string): GameData => readJSON<GameData>(path.join(root, 'conformance', 'definitions', name + '.json'));

// ----- the generated bodies -----------------------------------------------

/** Whether rules/stockRegistry.ts builds this move's stock type yet. */
function stockAvailable(move: MoveData): boolean {
	if (move.Stock === null) return true;
	try {
		return stockMove(move.Stock.Type, move.Stock.Config) !== undefined;
	} catch {
		return false;
	}
}

interface Adapter {
	bodies: Bodies;
	/** Moves given the registry's parts (from Go's exported configuration) through `stock:`. */
	standIns: string[];
	/** Moves supplied whole through `move:`. */
	overrides: string[];
	/** Moves left to their bodies under `wholeMoves`: their apply calls the stock Apply, which `move:` does not carry. */
	notOverridden: string[];
}

/** The registry's parts for a stock configuration, or undefined when it cannot build them. */
function partsFor(stock: NonNullable<MoveData['Stock']>): StockParts | undefined {
	try {
		return stockMove(stock.Type, stock.Config);
	} catch {
		return undefined;
	}
}

/**
 * The generated game's bodies for `data`.
 *
 * A move whose stock configuration in `data` the registry cannot build, or
 * builds without the Apply that Go's exported configuration of the same
 * move yields (a fixture lacks Go's boot-time inferences, such as
 * DrawToPlayer's resolved source and destination), gets the registry's
 * parts built from Go's exported configuration, through `stock:`.
 *
 * With `wholeMoves`, every move whose generated MoveSpec carries no stock
 * Apply is supplied whole through the `move:` hook. A move that does is
 * left to its bodies: the hook takes a move's behavior (`apply`, `legal`,
 * ...) and not its stock parts, and a translated apply calls the stock
 * Apply (`stockApply`), so through `move:` it would find none.
 */
function generatedBodies(data: GameData, name: string, wholeMoves: boolean): Adapter {
	const generated = GENERATED[name];
	const table = new Map<string, () => unknown>();
	const adapter: Adapter = { bodies: ref => table.get(ref) ?? generated.bodies(ref), standIns: [], overrides: [], notOverridden: [] };
	const moves = new Map(generated.make().moves.map(move => [move.name, move] as const));
	const exported = new Map(exportedDefinition(name).Moves.map(move => [move.Name, move] as const));
	for (const move of data.Moves) {
		const whole = moves.get(move.Name);
		assert.ok(whole, 'the generated ' + name + ' has no move ' + move.Name);
		if (wholeMoves && whole.stock?.apply === undefined) {
			table.set(moveOverrideRef(move.Name), (): MoveSpec => whole);
			adapter.overrides.push(move.Name);
			continue;
		}
		if (wholeMoves) adapter.notOverridden.push(move.Name);
		if (move.Stock === null) continue;
		const own = partsFor(move.Stock);
		const go = exported.get(move.Name);
		assert.ok(go?.Stock, name + ': Go exports no stock configuration for ' + move.Name);
		const goParts = partsFor(go.Stock);
		if (own !== undefined && (own.apply !== undefined || goParts?.apply === undefined)) continue;
		assert.ok(goParts !== undefined, name + ': the registry cannot build ' + move.Name + " from Go's exported configuration either");
		table.set(stockOverrideRef(move.Name), (): StockParts => goParts);
		adapter.standIns.push(move.Name);
	}
	return adapter;
}

// ----- comparison with the generated game ---------------------------------

function normalizedSubState(def: SubStateDef, defaultGroup: string): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const name of Object.keys(def.Props)) {
		const prop = def.Props[name];
		out[name] = {
			Type: prop.Type, Sanitize: policyMapFor(prop, defaultGroup), Enum: prop.Enum ?? '', Deck: prop.Deck ?? '', Size: prop.Size ?? 0,
			MaxSize: prop.MaxSize ?? 0, Overlap: prop.Overlap === true, Stacks: prop.Stacks ?? [],
		};
	}
	return out;
}

/** A definition without the GameData additions, in comparable form. */
function normalizedDefinition(def: GameDefinition): unknown {
	const decks: Record<string, DeckDef> = {};
	for (const name of Object.keys(def.Decks)) decks[name] = { Size: def.Decks[name].Size, Components: def.Decks[name].Components ?? [] };
	const dynamic: Record<string, unknown> = {};
	for (const deck of Object.keys(def.DynamicComponentValues)) dynamic[deck] = normalizedSubState(def.DynamicComponentValues[deck], 'all');
	return JSON.parse(JSON.stringify({
		Name: def.Name, Decks: decks, Enums: def.Enums, GroupEnum: def.GroupEnum ?? '',
		Game: normalizedSubState(def.Game, 'all'), Player: normalizedSubState(def.Player, 'other'), Dynamic: dynamic,
	}));
}

/**
 * A progression group with each counter (a function in a hand-written GameSpec, data
 * from the loader) replaced by its results on currentCount 0..16 at the
 * length the group evaluates it at (1 for repeat, the child count for
 * parallel).
 */
function comparableGroup(group: ProgressionGroup): unknown {
	const grid = (counter: Counter, length: number): (string | null)[] => {
		const out: (string | null)[] = [];
		for (let i = 0; i <= 16; i++) out.push(evalCounter(counter, i, length));
		return out;
	};
	switch (group.kind) {
		case 'move':
			return group;
		case 'serial':
			return { kind: 'serial', children: group.children.map(comparableGroup) };
		case 'repeat':
			return { kind: 'repeat', count: grid(group.count, 1), child: comparableGroup(group.child) };
		case 'parallel':
			return { kind: 'parallel', count: grid(group.count, group.children.length), children: group.children.map(comparableGroup) };
	}
}

/** What compileGame derived that a client or the corpus sees, per move and for the game. */
function compiledSummary(game: CompiledGame): unknown {
	const spec = game.spec;
	return {
		displayName: gameDisplayName(spec),
		players: [spec.minNumPlayers, spec.maxNumPlayers, spec.defaultNumPlayers],
		definition: normalizedDefinition(game.def),
		phaseEnum: game.phaseEnum?.name ?? '',
		templates: game.templates,
		constants: spec.constants ?? null,
		variants: variantConfig(spec),
		computed: (spec.computed ?? []).map(item => [item.name, item.scope, item.type]),
		moveInputSchema: game.moveInputSchema,
		moveInputSchemaFingerprint: game.moveInputSchemaFingerprint,
		choiceSchema: game.choiceSchema,
		choiceSchemaFingerprint: game.choiceSchemaFingerprint,
		moves: game.moves.map(move => ({
			name: move.spec.name,
			helpText: move.spec.helpText,
			fixUp: move.spec.fixUp,
			base: move.spec.base ?? 'default',
			fields: move.fieldNames.map(name => [name, move.fields[name].type, move.fields[name].enum ?? '']),
			fieldPolicies: move.fieldPolicies,
			namePolicy: move.namePolicy,
			declared: move.declared,
			plan: move.plan.specs,
			hasCustom: move.plan.custom !== null,
			legalPhases: move.spec.legalPhases ?? [],
			progression: move.spec.progression ? comparableGroup(move.spec.progression) : null,
			seatsPlayer: move.spec.seatsPlayer === true,
			gatheringStart: move.spec.gatheringStart === true,
		})),
	};
}

function runCorpus(game: CompiledGame, name: string): { documents: number; versions: number } {
	const entries = corpusIndex.filter(entry => entry.game === name);
	assert.ok(entries.length > 0, 'no conformance documents for ' + name);
	let versions = 0;
	for (const entry of entries) {
		const doc = JSON.parse(gunzipSync(readFileSync(path.join(root, 'conformance', entry.file))).toString('utf8')) as ConformanceDocument;
		const result = runDocument(game, doc);
		const problems: string[] = [];
		if (result.fatal) problems.push('replay stopped: ' + result.fatal);
		for (const field of FIELDS) {
			const item = result.fields[field];
			if (item.mismatches > 0) problems.push(field + ': ' + item.mismatches + ' of ' + item.compared + ' differ, e.g.\n    ' + item.examples.join('\n    '));
		}
		assert.equal(problems.length, 0, entry.file + ':\n  ' + problems.join('\n  '));
		assert.equal(result.fields.state.compared, doc.versions.length, entry.file + ': every version was compared');
		versions += doc.versions.length;
	}
	return { documents: entries.length, versions };
}

/**
 * The proof for one GameData: loaded with the generated bodies in each of
 * `modes`, it compiles as the generated game does and passes the corpus.
 */
/** Per exported game, how many moves its last move: pass supplied whole. */
const overridden: Record<string, number> = {};

function proveLoader(t: { diagnostic(message: string): void }, data: GameData, name: string, modes: readonly boolean[], noStandIns = false): void {
	const reference = compileGame(GENERATED[name].make());
	for (const wholeMoves of modes) {
		const adapter = generatedBodies(data, name, wholeMoves);
		// An exported definition carries everything stockMove needs, Go's
		// boot-time inferences included (gamedef addResolvedConfig): no move may
		// need a stand-in.
		if (noStandIns) assert.deepEqual(adapter.standIns, [], name + ': moves the stock registry cannot build from the exported configuration');
		if (wholeMoves) {
			overridden[name] = adapter.overrides.length;
			// tictactoe's every move carries a stock Apply, so none can go through move:.
			assert.equal(adapter.overrides.length + adapter.notOverridden.length, data.Moves.length);
		}
		const loaded = compileGame(gameFromData(data, adapter.bodies));
		assert.equal(firstDifference(compiledSummary(loaded), compiledSummary(reference)), null, name + ': the loaded game compiles as the generated game does');
		const run = runCorpus(loaded, name);
		t.diagnostic(
			name + (wholeMoves ? ' (moves through move:)' : ' (bodies)') + ': ' + run.documents + ' documents, ' + run.versions + ' versions pass; ' +
				(wholeMoves ? adapter.overrides.length + ' move: overrides, ' + adapter.notOverridden.length + ' left to bodies (stock Apply) [' + adapter.notOverridden.join(', ') + ']; ' : '') +
				adapter.standIns.length + ' of ' + data.Moves.length + ' moves had stock parts through stock: [' + adapter.standIns.join(', ') + ']',
		);
	}
}

// ----- the stand-in fixtures ----------------------------------------------

interface MoveFacts {
	Name: string;
	HelpText: string;
	IsFixUp: boolean;
	LegalPhases: number[] | null;
	Progression: unknown;
	AllowMultipleInProgression: boolean;
	Plan: MoveData['Plan'];
}

for (const name of FIXTURES) {
	const data = readJSON<GameData>(path.join(here, 'testdata', name + '.gamedata.json'));

	test(`${name} fixture: the Go-derivable parts are what Go reports`, () => {
		const facts = readJSON<MoveFacts[]>(path.join(here, 'testdata', 'movefacts', name + '.json'));
		assert.deepEqual(
			data.Moves.map(move => ({
				Name: move.Name, HelpText: move.HelpText, IsFixUp: move.IsFixUp, LegalPhases: move.LegalPhases, Progression: move.Progression,
				AllowMultipleInProgression: move.AllowMultipleInProgression, Plan: move.Plan,
			})),
			facts,
		);
		const go = readJSON<GameDefinition>(path.join(here, '..', 'testdata', 'defs', name + '.def.json'));
		const decks: Record<string, unknown> = {};
		for (const deck of Object.keys(data.Definition.Decks)) {
			const { GenericValues: _generic, ComponentProps: _props, ...rest } = data.Definition.Decks[deck];
			decks[deck] = rest;
		}
		assert.deepEqual({ ...data.Definition, Decks: decks }, go);
	});

	test(`${name} fixture: gameFromData with the generated game's functions passes the corpus`, t => {
		proveLoader(t, data, name, [false, true]);
	});
}

// ----- the exported definitions -------------------------------------------

for (const name of Object.keys(GENERATED)) {
	test(`${name} exported definition: moves through move: compile as the generated game and pass the corpus`, t => {
		const data = exportedDefinition(name);
		// The bodies pass needs no stand-in; it is the generated game itself,
		// which games/<game>/<game>.gen.test.ts replays.
		assert.deepEqual(generatedBodies(data, name, false).standIns, [], name + ': moves the stock registry cannot build from the exported configuration');
		proveLoader(t, data, name, [true], true);
	});
}

test('the move: hook served moves of every exported game but tictactoe', () => {
	for (const name of Object.keys(GENERATED)) {
		if (name === 'tictactoe') assert.equal(overridden[name], 0, name);
		else assert.ok(overridden[name] > 0, name + ': no move went through move:');
	}
});

// ----- the loader's own contracts -----------------------------------------

const pigData = (): GameData => readJSON<GameData>(path.join(here, 'testdata', 'pig.gamedata.json'));
const blackjackData = (): GameData => readJSON<GameData>(path.join(here, 'testdata', 'blackjack.gamedata.json'));

test('a FnRef with no body is a GameSpecError naming it', () => {
	const data = pigData();
	const full = generatedBodies(data, 'pig', false).bodies;
	const missing: Bodies = ref => (ref === 'pig.(*moveRollDice).Apply' ? undefined : full(ref));
	assert.throws(
		() => gameFromData(data, missing),
		(thrown: unknown) => isGameSpecError(thrown) && /apply pig\.\(\*moveRollDice\)\.Apply/.test((thrown as Error).message),
	);
});

test("an empty SetUp.Distribute is base.GameDelegate's stub, not a FnRef", () => {
	// Werewolf declares no DistributeComponentToStarterStack (it has no
	// deck); the exporter writes ''. Go's stub errors if it is ever called.
	const want = readJSON<{ distributeComponentToStarterStackError: string }>(path.join(here, '..', 'testdata', 'base-delegate.vectors.json'));
	const data = pigData();
	const full = generatedBodies(data, 'pig', false).bodies;
	const asked: string[] = [];
	const spec = gameFromData({ ...data, SetUp: { ...data.SetUp, Distribute: '' } }, ref => {
		asked.push(ref);
		return full(ref);
	});
	assert.ok(!asked.includes(''), 'the loader asked bodies for the empty FnRef');
	assert.equal(want.distributeComponentToStarterStackError, BASE_DISTRIBUTE_ERROR);
	const game = compileGame(spec);
	const engine = loadEngineState(game.def, { gameId: '', secretSalt: '' }, 0, emptyStateJSON(game, 2), { skipConservation: true });
	const state = newRulesState(game, engine, { name: 'pig', id: '', secretSalt: '', numPlayers: 2, variant: null, agents: [], seating: false }, [], { pendingSeat: null });
	assert.throws(() => spec.setUp.distribute(state, 'dice', 0), { message: want.distributeComponentToStarterStackError });
});

test("a ComponentLegal body gets its receiver as the values handle { deck, index }, the translator's shape", () => {
	// checkers' (*token).Legal is translated as (state, t: DynamicRef,
	// legalType); DeckSpec.componentLegal stays (state, index, legalType).
	const data = readJSON<GameData>(path.join(root, 'conformance', 'definitions', 'checkers.json'));
	const ref = data.ComponentLegal?.Tokens;
	assert.equal(ref, 'checkers.(*token).Legal');
	const calls: unknown[][] = [];
	const spec = gameFromData(data, wanted => {
		if (wanted === ref) {
			return (...args: unknown[]): string | null => {
				calls.push(args.slice(1));
				return 'recorded';
			};
		}
		return checkersBodies(wanted);
	});
	const legal = spec.decks.Tokens.componentLegal;
	assert.ok(legal !== undefined);
	const state = {} as RulesState;
	assert.equal(legal(state, 5, null), 'recorded');
	assert.equal(legal(state, 0, 3), 'recorded');
	assert.deepEqual(calls, [[{ deck: 'Tokens', index: 5 }, null], [{ deck: 'Tokens', index: 0 }, 3]]);
});

test('a stock type the engine cannot build is a GameSpecError, not a silent default', () => {
	const data = pigData();
	const move = data.Moves.find(item => item.Name === 'Finish Turn') as MoveData;
	if (stockAvailable(move)) return; // stockMove builds it now; nothing to refuse.
	const bodies = generatedBodies(data, 'pig', false).bodies;
	const withoutStandIn: Bodies = ref => (ref === stockOverrideRef('Finish Turn') ? undefined : bodies(ref));
	assert.throws(() => gameFromData(data, withoutStandIn));
});

test('Plan.HasCustom must agree with a LegalCustom body', () => {
	const data = blackjackData();
	const hit = data.Moves.find(item => item.Name === 'Current Player Hit') as MoveData;
	const edited: GameData = { ...data, Moves: data.Moves.map(item => (item === hit ? { ...hit, Plan: { ...hit.Plan, HasCustom: false } } : item)) };
	assert.throws(() => gameFromData(edited, generatedBodies(data, 'blackjack', false).bodies), (thrown: unknown) => isGameSpecError(thrown));
});

test('initial values, generic values and component props come through', () => {
	const game = compileGame(gameFromData(blackjackData(), generatedBodies(blackjackData(), 'blackjack', false).bodies));
	assert.deepEqual(game.def.Decks.cards.GenericValues, { Suit: '�', Rank: 'Unknown' });
	assert.deepEqual(game.components.cards, { Rank: { Type: 'enum', Enum: 'Rank' }, Suit: { Type: 'enum', Enum: 'Suit' } });
	const engine = loadEngineState(game.def, { gameId: '', secretSalt: '' }, 0, emptyStateJSON(game, 2), { skipConservation: true });
	const state = newRulesState(game, engine, { name: 'blackjack', id: '', secretSalt: '', numPlayers: 2, variant: null, agents: [], seating: false }, [], { pendingSeat: null });
	assert.deepEqual(state.component('cards', GENERIC_COMPONENT), { Suit: '�', Rank: 'Unknown' });
	assert.deepEqual(state.component('cards', 0), { Suit: '♠', Rank: 'Ace' });

	const pigGame = compileGame(gameFromData(pigData(), generatedBodies(pigData(), 'pig', false).bodies));
	const empty = emptyStateJSON(pigGame, 3);
	assert.equal(empty.Game.TargetScore, 100);
	assert.equal(empty.Components.dice[0].Value, 1);
	// A deck with no generic values: the generic component has none.
	const pigEngine = loadEngineState(pigGame.def, { gameId: '', secretSalt: '' }, 0, empty, { skipConservation: true });
	const pigState = newRulesState(pigGame, pigEngine, { name: 'pig', id: '', secretSalt: '', numPlayers: 3, variant: null, agents: [], seating: false }, [], { pendingSeat: null });
	assert.deepEqual(pigState.component('dice', GENERIC_COMPONENT), {});
	assert.equal(pigGame.def.Decks.dice.GenericValues, undefined);
});

test('callStockApply runs the stock part and returns its error text', () => {
	const data = pigData();
	const bodies = generatedBodies(data, 'pig', false).bodies;
	const calls: string[] = [];
	const withStock: Bodies = ref => {
		if (ref === stockOverrideRef('Roll Dice')) {
			return (): StockParts => ({
				apply: (_state, m) => {
					calls.push(m.name);
					if (m.fields.fail === true) throw new Error('stock refused');
				},
			});
		}
		return bodies(ref);
	};
	const game = compileGame(gameFromData(data, withStock));
	const state = { game } as RulesState;
	assert.equal(callStockApply(state, { name: 'roll dice', fields: {} }), null);
	assert.equal(callStockApply(state, { name: 'Roll Dice', fields: { fail: true } }), 'stock refused');
	assert.deepEqual(calls, ['roll dice', 'Roll Dice']);
	assert.throws(() => callStockApply(state, { name: 'Done Turn', fields: {} }), /no stock move with an Apply/);
});

test('error results: a returned string from an apply-like body is thrown; a check tuple is read', () => {
	const data = pigData();
	const bodies = generatedBodies(data, 'pig', false).bodies;
	const withString: Bodies = ref => (ref === 'pig.(*moveDoneTurn).Apply' ? (): string => 'not now' : bodies(ref));
	const spec = gameFromData(data, withString);
	const doneTurn = spec.moves.find(move => move.name === 'Done Turn') as MoveSpec;
	assert.throws(() => doneTurn.apply({} as RulesState, { name: 'Done Turn', fields: {} }), /not now/);

	const withCheck: GameData = { ...data, End: { Check: 'pig.check' } };
	const tuple: Bodies = ref => (ref === 'pig.check' ? (): [boolean, number[] | null] => [true, null] : bodies(ref));
	const checked = gameFromData(withCheck, tuple);
	assert.deepEqual(checked.end?.check?.({} as RulesState), { finished: true, winners: [] });
});
