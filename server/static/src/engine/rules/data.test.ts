// The loader (rules/data.ts) on the hand ports. For a game, a GameSpec built
// by gameFromData from the game's exported definition, with the hand port's
// own functions as bodies, must compile to the same header tables, plans and
// state definition as the hand port, and must pass that game's whole
// conformance corpus through the existing runner (testing/conformance.ts).
//
// Two sources of GameData:
//
//  - conformance/definitions/<game>.json, written by the Go exporter
//    (`boardgame-util gamedef export`). Each game whose file is absent is a
//    skipped subtest that says so.
//  - rules/testdata/<game>.gamedata.json: stand-ins written for this test
//    before the exporter existed, for pig and blackjack. Their Definition
//    equals Go's DefinitionFor output (testdata/defs/<game>.def.json) plus
//    the GameData additions, and each move's name, help text, fix-up flag,
//    legal phases, progression, AllowMultipleInProgression and plan equal
//    what the live Go manager reports (rules/testdata/movefacts/<game>.json,
//    from rules/testdata/gen/movefacts); both are checked below. The rest
//    (Stock, Base, Fields, Bodies FnRefs, Traits, Initial, flags) is written
//    by hand from the Go source and is held only by the corpus.
//
// The hand port supplies bodies by role (handPortBodies). A move whose stock
// type rules/stockRegistry.ts cannot build yet gets the hand port's parts
// through the `stock:<name>` hook; a second pass builds every move through
// the `move:<name>` hook instead, so both hooks are exercised.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { policyMapFor } from '../definition.ts';
import type { DeckDef, GameDefinition, SubStateDef } from '../definition.ts';
import { blackjack } from '../games/blackjack.ts';
import { debuganimations } from '../games/debuganimations.ts';
import { memory } from '../games/memory.ts';
import { pig } from '../games/pig.ts';
import { tictactoe } from '../games/tictactoe.ts';
import { werewolf } from '../games/werewolf.ts';
import { bodies as checkersBodies } from '../games/checkers/bodies.gen.ts';
import { GENERIC_COMPONENT } from '../ids.ts';
import { loadEngineState } from '../state.ts';
import { CHECKERS_PREDICATE_REF, spaceIsBlackConstructor } from '../testing/checkers.ts';
import { type ConformanceDocument, FIELDS, firstDifference, runDocument } from '../testing/conformance.ts';
import { compileGame, emptyStateJSON, gameDisplayName, isGameSpecError, variantConfig } from './compile.ts';
import { BASE_DISTRIBUTE_ERROR, type Bodies, type GameData, type MoveData, callStockApply, gameFromData, moveOverrideRef, stockOverrideRef } from './data.ts';
import { type Counter, type ProgressionGroup, evalCounter } from './progression.ts';
import type { CompiledGame, RulesState } from './state.ts';
import { newRulesState } from './state.ts';
import { type StockParts, stockMove } from './stockRegistry.ts';
import type { GameSpec, MoveSpec } from './types.ts';

const HAND_PORTS: Record<string, () => GameSpec> = { memory, blackjack, tictactoe, pig, debuganimations, werewolf };
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

// ----- the hand port as bodies --------------------------------------------

type AnyBody = (...args: never[]) => unknown;

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
	/** Moves given the hand port's parts through `stock:` / whole through `move:`. */
	standIns: string[];
	overrides: string[];
	/** FnRefs served by a TEST_BODIES entry. */
	testBodies: string[];
}

/**
 * Bodies for Go functions a hand port has no separate function for, because
 * the hand port leaves to an engine default what the Go game declares as a
 * method. Each is the Go function, written from the hand port's pieces.
 */
const TEST_BODIES: Record<string, (hand: GameSpec) => AnyBody | undefined> = {
	// `return pState.(*playerState).GameScore()`. The hand port leaves
	// end.playerScore unset, so the engine scores with the gameScore trait
	// (rules/engine.ts checkGameFinished): this is that default.
	'memory.(*gameDelegate).PlayerScore': hand => hand.player.traits?.gameScore,
	// `return false, nil` (never finished). The hand port has no end, which the
	// engine reads as never finished.
	'debuganimations.(*gameDelegate).CheckGameFinished': () => (): { finished: boolean; winners: number[] } => ({ finished: false, winners: [] }),
};

/**
 * The hand port's functions as bodies, by the role the data names each
 * FnRef in. With `wholeMoves`, every move is the hand port's through the
 * `move:` hook; otherwise only moves whose stock type is not yet ported get
 * the hand port's stock-side parts through `stock:`.
 */
function handPortBodies(data: GameData, hand: GameSpec, wholeMoves: boolean): Adapter {
	const table = new Map<string, AnyBody>();
	const adapter: Adapter = { bodies: ref => table.get(ref), standIns: [], overrides: [], testBodies: [] };
	const add = (ref: string | undefined, handBody: AnyBody | undefined, role: string): void => {
		if (ref === undefined) return;
		let body = handBody;
		if (body === undefined && TEST_BODIES[ref] !== undefined) {
			body = TEST_BODIES[ref](hand);
			if (body !== undefined && !adapter.testBodies.includes(ref)) adapter.testBodies.push(ref);
		}
		assert.ok(body !== undefined, 'the hand port has no ' + role + ' for ' + ref);
		const existing = table.get(ref);
		// One Go method shared by several move configs (werewolf's moveCastVote
		// for Cast Vote and Cast Night Vote) is one closure per config in a hand
		// port built by a factory. The same function expression is accepted and
		// the first kept: the corpus then runs every config sharing it on that one.
		assert.ok(existing === undefined || existing === body || existing.toString() === body.toString(), 'two different hand-port functions for ' + ref);
		if (existing === undefined) table.set(ref, body);
	};
	const handMoves = new Map(hand.moves.map(move => [move.name, move] as const));
	for (const move of data.Moves) {
		const hm = handMoves.get(move.Name);
		assert.ok(hm, 'the hand port has no move ' + move.Name);
		if (wholeMoves) {
			table.set(moveOverrideRef(move.Name), (): MoveSpec => hm);
			adapter.overrides.push(move.Name);
			continue;
		}
		add(move.Bodies.Apply, hm.apply, 'apply');
		add(move.Bodies.Legal, hm.legal, 'legal');
		add(move.Bodies.LegalCustom, hm.legalCustom, 'legalCustom');
		add(move.Bodies.Defaults, hm.defaults, 'defaults');
		if (!stockAvailable(move)) {
			// What the stock ancestor contributes, as the hand port has it: the
			// methods the game does not override, and the stack properties.
			const parts: StockParts = { spec: {} };
			if (move.Bodies.Apply === undefined) parts.apply = hm.apply;
			if (move.Bodies.Legal === undefined && hm.legal) parts.legal = hm.legal;
			if (move.Bodies.Defaults === undefined && hm.defaults) parts.defaults = hm.defaults;
			const spec: Partial<MoveSpec> = {};
			if (hm.sourceProperty !== undefined) spec.sourceProperty = hm.sourceProperty;
			if (hm.destinationProperty !== undefined) spec.destinationProperty = hm.destinationProperty;
			if (hm.contributes) spec.contributes = hm.contributes;
			parts.spec = spec;
			table.set(stockOverrideRef(move.Name), (): StockParts => parts);
			adapter.standIns.push(move.Name);
		}
	}
	const computed = new Map((hand.computed ?? []).map(item => [item.name, item.compute as AnyBody] as const));
	for (const item of data.Computed) add(item.Compute, computed.get(item.Name), 'computed ' + item.Name);
	const playerTraits = hand.player.traits ?? {};
	if (data.Traits.Player.GameScore !== undefined && data.Traits.Player.GameScore !== 'score') add(data.Traits.Player.GameScore, playerTraits.gameScore, 'gameScore');
	const turn = data.Traits.Player.Turn;
	if (turn) {
		add(turn.TurnDone, playerTraits.turn?.turnDone, 'turnDone');
		add(turn.ResetForTurnStart, playerTraits.turn?.resetForTurnStart, 'resetForTurnStart');
		add(turn.ResetForTurnEnd, playerTraits.turn?.resetForTurnEnd, 'resetForTurnEnd');
	}
	add(data.Traits.Game.PlayerOrder, hand.game.traits?.playerOrder, 'playerOrder');
	add(data.SetUp.Begin, hand.setUp.begin, 'setUp.begin');
	add(data.SetUp.Distribute, hand.setUp.distribute, 'setUp.distribute');
	add(data.SetUp.Finish, hand.setUp.finish, 'setUp.finish');
	add(data.End.ConditionMet, hand.end?.conditionMet, 'end.conditionMet');
	add(data.End.PlayerScore, hand.end?.playerScore, 'end.playerScore');
	add(data.End.Check, hand.end?.check, 'end.check');
	add(data.CurrentPlayerIndex, hand.currentPlayerIndex, 'currentPlayerIndex');
	return adapter;
}

// ----- comparison with the hand port --------------------------------------

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
 * A progression group with each counter (a function in a hand port, data
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

/** The whole proof for one GameData: compiles like the hand port and passes the corpus, in both hook modes. */
function proveLoader(t: { diagnostic(message: string): void }, data: GameData, name: string, noStandIns = false): void {
	const hand = HAND_PORTS[name]();
	const handGame = compileGame(hand);
	for (const wholeMoves of [false, true]) {
		const adapter = handPortBodies(data, hand, wholeMoves);
		// An exported definition carries everything stockMove needs, Go's
		// boot-time inferences included (gamedef addResolvedConfig): no move may
		// fall back to the hand port's stock parts.
		if (noStandIns) assert.deepEqual(adapter.standIns, [], name + ': moves the stock registry cannot build from the exported configuration');
		const loaded = compileGame(gameFromData(data, adapter.bodies));
		assert.equal(firstDifference(compiledSummary(loaded), compiledSummary(handGame)), null, name + ': the loaded game compiles as the hand port does');
		const run = runCorpus(loaded, name);
		t.diagnostic(
			name + (wholeMoves ? ' (every move through move:)' : ' (bodies)') + ': ' + run.documents + ' documents, ' + run.versions + ' versions pass; ' +
				(wholeMoves ? adapter.overrides.length + ' move: overrides' : adapter.standIns.length + ' of ' + data.Moves.length + ' moves had stock parts through stock: [' + adapter.standIns.join(', ') + ']') +
				'; test bodies [' + adapter.testBodies.join(', ') + ']',
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

	test(`${name} fixture: gameFromData with the hand port's functions passes the corpus`, t => {
		proveLoader(t, data, name);
	});
}

// ----- the exported definitions -------------------------------------------

for (const name of Object.keys(HAND_PORTS)) {
	const file = path.join(root, 'conformance', 'definitions', name + '.json');
	test(`${name} exported definition: gameFromData with the hand port's functions passes the corpus`, t => {
		if (!existsSync(file)) {
			t.skip('conformance/definitions/' + name + '.json is absent: the Go exporter has not written it yet');
			return;
		}
		const data = readJSON<GameData>(file);
		if (data.Functions.length === 0 && data.SetUp.Distribute === '') {
			// The exporter's static half fills every function reference. A
			// definition written without it cannot be loaded; say so, loudly.
			t.skip('conformance/definitions/' + name + '.json holds run-time parts only (no function references): the static half has not been joined yet');
			return;
		}
		proveLoader(t, data, name, true);
	});
}

// ----- the loader's own contracts -----------------------------------------

const pigData = (): GameData => readJSON<GameData>(path.join(here, 'testdata', 'pig.gamedata.json'));
const blackjackData = (): GameData => readJSON<GameData>(path.join(here, 'testdata', 'blackjack.gamedata.json'));

test('a FnRef with no body is a GameSpecError naming it', () => {
	const data = pigData();
	const full = handPortBodies(data, pig(), false).bodies;
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
	const full = handPortBodies(data, pig(), false).bodies;
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
		return wanted === CHECKERS_PREDICATE_REF ? spaceIsBlackConstructor : checkersBodies(wanted);
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
	const bodies = handPortBodies(data, pig(), false).bodies;
	const withoutStandIn: Bodies = ref => (ref === stockOverrideRef('Finish Turn') ? undefined : bodies(ref));
	assert.throws(() => gameFromData(data, withoutStandIn));
});

test('Plan.HasCustom must agree with a LegalCustom body', () => {
	const data = blackjackData();
	const hit = data.Moves.find(item => item.Name === 'Current Player Hit') as MoveData;
	const edited: GameData = { ...data, Moves: data.Moves.map(item => (item === hit ? { ...hit, Plan: { ...hit.Plan, HasCustom: false } } : item)) };
	assert.throws(() => gameFromData(edited, handPortBodies(data, blackjack(), false).bodies), (thrown: unknown) => isGameSpecError(thrown));
});

test('initial values, generic values and component props come through', () => {
	const game = compileGame(gameFromData(blackjackData(), handPortBodies(blackjackData(), blackjack(), false).bodies));
	assert.deepEqual(game.def.Decks.cards.GenericValues, { Suit: '�', Rank: 'Unknown' });
	assert.deepEqual(game.components.cards, { Rank: { Type: 'enum', Enum: 'Rank' }, Suit: { Type: 'enum', Enum: 'Suit' } });
	const engine = loadEngineState(game.def, { gameId: '', secretSalt: '' }, 0, emptyStateJSON(game, 2), { skipConservation: true });
	const state = newRulesState(game, engine, { name: 'blackjack', id: '', secretSalt: '', numPlayers: 2, variant: null, agents: [], seating: false }, [], { pendingSeat: null });
	assert.deepEqual(state.component('cards', GENERIC_COMPONENT), { Suit: '�', Rank: 'Unknown' });
	assert.deepEqual(state.component('cards', 0), { Suit: '♠', Rank: 'Ace' });

	const pigGame = compileGame(gameFromData(pigData(), handPortBodies(pigData(), pig(), false).bodies));
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
	const bodies = handPortBodies(data, pig(), false).bodies;
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
	const bodies = handPortBodies(data, pig(), false).bodies;
	const withString: Bodies = ref => (ref === 'pig.(*moveDoneTurn).Apply' ? (): string => 'not now' : bodies(ref));
	const spec = gameFromData(data, withString);
	const doneTurn = spec.moves.find(move => move.name === 'Done Turn') as MoveSpec;
	assert.throws(() => doneTurn.apply({} as RulesState, { name: 'Done Turn', fields: {} }), /not now/);

	const withCheck: GameData = { ...data, End: { Check: 'pig.check' } };
	const tuple: Bodies = ref => (ref === 'pig.check' ? (): [boolean, number[] | null] => [true, null] : bodies(ref));
	const checked = gameFromData(withCheck, tuple);
	assert.deepEqual(checked.end?.check?.({} as RulesState), { finished: true, winners: [] });
});
