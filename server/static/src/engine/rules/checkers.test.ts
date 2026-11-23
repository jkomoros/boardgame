// Checkers through the loader: conformance/definitions/checkers.json (the Go
// exporter's output) plus checkers' generated bodies
// (games/checkers/bodies.gen.ts, `boardgame-util transpile
// ./examples/checkers`), built by gameFromData and compiled by compileGame,
// held to real Go at the loader level:
//
//  1. Every function slot checkers.json names has a body; a predicate with no
//     body is a GameSpecError naming it; every move's compiled plan is the
//     plan Go assembled (MoveData.Plan, from LegalPlanSpecs); a game
//     predicate constructor's error fails compileGame with its text.
//  2. Three Go fuzz-driver documents (rules/testdata/checkers/seed-000{0,1,6}
//     .json.gz: seed 0 at default bounds, unseated, played to the end; seed 1
//     seated; seed 6; the last two cut at 60 versions) replay in full through
//     the conformance runner: header, every version, every viewer, legality
//     ledgers included, and each recorded refusal is refused again with Go's
//     text.
//  3. The facade's { move } owner and componentInstance, against Go.
//
// The constructed positions (testdata/checkers.vectors.json.gz) are replayed
// by games/checkers/checkers.gen.test.ts against the same bodies. The
// recordings and the replay code are ../testing/checkers.ts.
//
// To regenerate the Go side:
//
//   ./scripts/go-local run ./boardgame-util fuzz run --game checkers --seed 0 --out DIR
//   ./scripts/go-local run ./boardgame-util fuzz run --game checkers --seeds 1..6 --max-proposals 40 --max-versions 60 --out DIR2
//   cp DIR/checkers/seed-0000.json.gz DIR2/checkers/seed-000[16].json.gz server/static/src/engine/rules/testdata/checkers/
//   ./scripts/go-local run ./server/static/src/engine/testdata/gen/checkersvec
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { bodies as checkersBodies } from '../games/checkers/bodies.gen.ts';
import { componentAt, componentInstance, containingStack, enumEquals, enumKey, gameStack, rangeValue, setEnumKey, stackView } from '../go/api.ts';
import { loadEngineState } from '../state.ts';
import type { StateJSON } from '../state.ts';
import { type Vectors, CHECKERS_DOCUMENTS, CHECKERS_PREDICATE_REF, CHECKERS_VECTORS, readGzipJSON, stateOf } from '../testing/checkers.ts';
import { type ConformanceDocument, FIELDS, passTable, proposedInput, runDocument } from '../testing/conformance.ts';
import { compileGame, emptyStateJSON, isGameSpecError } from './compile.ts';
import { type GameData, gameFromData } from './data.ts';
import { type Input, createGame, propose } from './engine.ts';
import { parseInstant } from './time.ts';
import { type CompiledGame, newRulesState } from './state.ts';
import type { MoveRecord } from './types.ts';
import type { MoveInstance } from './types.ts';

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '..', '..', '..', '..', '..');
const DEFINITION = path.join(root, 'conformance', 'definitions', 'checkers.json');
const DOCUMENTS = CHECKERS_DOCUMENTS;
const VECTORS = CHECKERS_VECTORS;
/** testdata/gen/facadevec (moves.go): an enum.RangeVal move field of real Go checkers. */
const MOVE_FIELD_VECTORS = path.join(here, '..', 'testdata', 'facade-moves.vectors.json');

function loadDefinition(): GameData {
	return JSON.parse(readFileSync(DEFINITION, 'utf8')) as GameData;
}

/** Every FnRef checkers.json names, in any slot. */
function namedRefs(data: GameData): string[] {
	const named = new Set<string>();
	for (const move of data.Moves) for (const ref of Object.values(move.Bodies)) if (ref) named.add(ref);
	for (const ref of [data.SetUp.Begin, data.SetUp.Distribute, data.SetUp.Finish, data.End.ConditionMet, data.End.PlayerScore, data.End.Check, data.Traits.Player.GameScore]) if (ref) named.add(ref);
	for (const ref of Object.values(data.Traits.Player.Turn ?? {})) if (ref) named.add(ref);
	for (const ref of Object.values(data.ComponentLegal ?? {})) named.add(ref);
	for (const p of data.Predicates) named.add(p.Construct);
	return [...named].sort();
}

function buildCheckers(): { game: CompiledGame; data: GameData } {
	const data = loadDefinition();
	return { game: compileGame(gameFromData(data, checkersBodies)), data };
}

// ----- 1 and 2: the definition, the plans, the documents --------------------

test('checkers.json loads and compiles, and every function it names has a generated body', () => {
	const { game, data } = buildCheckers();
	assert.equal(game.moves.length, 7);
	const named = namedRefs(data);
	assert.ok(named.includes(CHECKERS_PREDICATE_REF));
	assert.ok(named.includes('checkers.(*gameDelegate).BeginSetUp'));
	assert.deepEqual(named.filter(ref => checkersBodies(ref) === undefined), [], 'functions with no body');
	// A predicate the data registers with no body is a GameSpecError naming it.
	assert.throws(
		() => gameFromData(data, ref => (ref === CHECKERS_PREDICATE_REF ? undefined : checkersBodies(ref))),
		thrown => isGameSpecError(thrown) && thrown.message === 'gameFromData: no body for predicate ' + CHECKERS_PREDICATE_REF,
	);
});

test('every compiled plan is the plan Go assembled', () => {
	const { game, data } = buildCheckers();
	for (const move of data.Moves) {
		const compiled = game.movesByName[move.Name];
		assert.equal(compiled.declared, move.Plan.Declared, move.Name + ' declared');
		if (!move.Plan.Declared) continue;
		assert.deepEqual(
			compiled.plan.steps.map(step => ({ name: step.predicate.name, args: [...step.predicate.args] })),
			move.Plan.Specs.map(spec => ({ name: spec.name, args: [...(spec.args ?? [])] })),
			move.Name + ' steps',
		);
		assert.equal(compiled.plan.custom !== null, move.Plan.HasCustom, move.Name + ' custom tail');
	}
	// Move Token's game predicate is checkers' own: built from the spec, Go's metadata.
	const step = game.movesByName['Move Token'].plan.steps.find(s => s.predicate.name === 'checkers.spaceIsBlack');
	assert.ok(step);
	assert.deepEqual(step.predicate.reads, [{ path: 'move.SpaceIndex', facet: 'values' }]);
	assert.equal(step.fieldDependent, true);
	assert.equal(step.predicate.clientEvaluable, false);
	assert.deepEqual(step.predicate.emittedTemplates, ['checkers.black_spaces_only']);
});

// The text after "constructor failed: " is the constructor's own (Go's
// fmt.Errorf, transliterated); the prefix is the engine's plan assembly.
// No Go vector pins the whole boot error for a game predicate.
test('a game predicate constructor error fails compileGame with the constructor text', () => {
	const data = loadDefinition();
	const moves = data.Moves.map(move =>
		move.Name !== 'Move Token'
			? move
			: { ...move, Plan: { ...move.Plan, Specs: move.Plan.Specs.map(spec => (spec.name === 'checkers.spaceIsBlack' ? { ...spec, args: ['move.SpaceIndex', 'extra'] } : spec)) } },
	);
	assert.throws(
		() => compileGame(gameFromData({ ...data, Moves: moves }, checkersBodies)),
		thrown =>
			thrown instanceof Error &&
			thrown.message.endsWith('boardgame: legal spec "checkers.spaceIsBlack": constructor failed: checkers.spaceIsBlack: requires exactly 1 arg (the space-index field), got 2'),
	);
});

function asState(value: unknown): StateJSON {
	return value as StateJSON;
}

interface Rejection {
	atVersion: number;
	kind: 'propose' | 'timer' | 'seat';
	name: string;
	input: Record<string, unknown>;
	proposer: number;
	expectedVersion: number;
	error: string;
}

/**
 * The recorded refusals of a version-2 document, proposed again at their
 * version (the snapshot is the document's own state and move records, which
 * the replay has just shown the engine reproduces): each must be refused
 * with Go's text. Returns one line per mismatch.
 */
function checkRejections(game: CompiledGame, doc: ConformanceDocument & { rejections?: Rejection[] }): { checked: number; mismatches: string[] } {
	const mismatches: string[] = [];
	const rejections = doc.rejections ?? [];
	const now = parseInstant(doc.engine.clock);
	assert.ok(now !== null);
	const created = createGame(game, { id: doc.game.id, secretSalt: doc.game.secretSalt, numPlayers: doc.game.numPlayers, variant: doc.game.variant, agents: doc.game.agents, seating: doc.game.seating, now });
	assert.ok(created.record);
	for (const rejection of rejections) {
		const history: MoveRecord[] = [];
		for (let v = 1; v <= rejection.atVersion; v++) {
			const m = doc.versions[v].move;
			assert.ok(m);
			history.push({ Name: m.name, Version: m.version, Initiator: m.initiator, Phase: m.phase, Proposer: m.proposer, Blob: m.arguments as Record<string, unknown> });
		}
		const snapshot = { record: created.record, history, state: asState(doc.versions[rejection.atVersion].state) };
		let input: Input;
		if (rejection.kind === 'propose') input = { kind: 'propose', name: rejection.name, input: rejection.input, proposer: rejection.proposer, expectedVersion: rejection.expectedVersion };
		else if (rejection.kind === 'seat') input = { kind: 'seat', player: rejection.input.TargetPlayerIndex as number };
		else input = { kind: 'timer', force: true };
		const step = propose(game, snapshot, input, { now });
		if (step.error !== rejection.error || step.versions.length !== 0) {
			mismatches.push(doc.name + ' rejection at v' + rejection.atVersion + ' ' + JSON.stringify(rejection.input) + ': got ' + JSON.stringify(step.error) + ' (' + step.versions.length + ' versions), want ' + JSON.stringify(rejection.error));
		}
	}
	return { checked: rejections.length, mismatches };
}

test('the Go fuzz documents of checkers replay in full: header, every version, every viewer, every refusal', () => {
	const { game } = buildCheckers();
	const results = [];
	let rejections = 0;
	for (const file of DOCUMENTS) {
		assert.ok(existsSync(file), file + ' is missing');
		const doc = readGzipJSON<ConformanceDocument & { rejections?: Rejection[] }>(file);
		const result = runDocument(game, doc);
		results.push(result);
		const table = passTable([result]);
		for (const field of FIELDS) {
			assert.equal(result.fields[field].mismatches, 0, field + ': ' + result.fields[field].examples.join('\n') + '\n' + table);
		}
		assert.equal(result.fatal, undefined, table);
		assert.ok(result.passed, table);
		assert.equal(result.fields.state.compared, doc.versions.length);
		assert.equal(result.fields['view.legality'].compared, doc.versions.length * doc.viewers.length);
		// Set-up, then real play.
		assert.ok(doc.versions.some(v => v.move?.name === 'Start Phase Playing'));
		assert.equal(doc.versions.filter(v => v.move?.name === 'Place Token').length, 24);
		assert.ok(doc.versions.some(v => v.move?.name === 'Move Token'), doc.name + ' applies no Move Token');
		const refused = checkRejections(game, doc);
		assert.deepEqual(refused.mismatches, []);
		rejections += refused.checked;
		if (path.basename(file) === 'seed-0000.json.gz') {
			assert.equal(doc.versions[doc.versions.length - 1].finished, true, 'seed 0 plays to the end');
			assert.ok(doc.versions.some(v => v.move?.name === 'Crown Token'), 'seed 0 crowns a token');
		}
		if (path.basename(file) === 'seed-0001.json.gz') assert.ok(doc.versions.some(v => v.move?.name === 'Seat Player'), 'seed 1 is seated');
	}
	assert.ok(rejections > 0, 'the documents record refusals');
	console.log(passTable(results));
});

// ----- the facade's { move } Owner, against Go (testdata/gen/facadevec) ----

interface MoveFieldRead {
	Stored: unknown;
	Value: number;
	Range: number[] | null;
}

interface MoveFieldVectors {
	Game: string;
	Move: string;
	Field: string;
	Enum: string;
	Initial: MoveFieldRead;
	Sets: (MoveFieldRead & { Key: number; Error: string })[];
	Records: (MoveFieldRead & { Blob: Record<string, unknown>; Error: string })[];
}

test('an enum.RangeVal move field through the facade: Value, SetValue, RangeValue and the stored record, as Go', () => {
	const { game } = buildCheckers();
	const vectors = JSON.parse(readFileSync(MOVE_FIELD_VECTORS, 'utf8')) as MoveFieldVectors;
	assert.equal(vectors.Game, 'checkers');
	const engine = loadEngineState(game.def, { gameId: '', secretSalt: '' }, 0, emptyStateJSON(game, 2), { skipConservation: true });
	const state = newRulesState(game, engine, { name: 'checkers', id: '', secretSalt: '', numPlayers: 2, variant: null, agents: ['', ''], seating: false }, [], { pendingSeat: null });
	const read = (m: MoveInstance): MoveFieldRead => ({ Stored: m.fields[vectors.Field], Value: enumKey(state, { move: m }, vectors.Field), Range: rangeValue(state, { move: m }, vectors.Field) });
	const move = state.newMove(vectors.Move);
	assert.ok(move);
	assert.deepEqual(read(move), vectors.Initial, 'a new move');
	for (const set of vectors.Sets) {
		const err = setEnumKey(state, { move }, vectors.Field, set.Key);
		assert.equal(err ?? '', set.Error, 'SetValue(' + set.Key + ')');
		assert.deepEqual(read(move), { Stored: set.Stored, Value: set.Value, Range: set.Range }, 'after SetValue(' + set.Key + ')');
	}
	for (const record of vectors.Records) {
		assert.equal(record.Error, '');
		// A move record's blob is the move's fields, as the engine holds them.
		const fresh = state.newMove(vectors.Move);
		assert.ok(fresh);
		Object.assign(fresh.fields, record.Blob);
		assert.deepEqual(read(fresh), { Stored: record.Stored, Value: record.Value, Range: record.Range }, 'record ' + JSON.stringify(record.Blob));
	}
	// Equals across owners: a move field and the enum it is drawn from.
	const other = state.newMove(vectors.Move);
	assert.ok(other);
	assert.equal(setEnumKey(state, { move: other }, vectors.Field, vectors.Sets[vectors.Sets.length - 1].Key), null);
	assert.equal(enumEquals(state, { move }, vectors.Field, { move: other }, vectors.Field), true);
	// A field that is not an enum is refused, not read as one.
	assert.throws(() => enumKey(state, { move }, 'NoSuchField'), /is not an enum property/);
});

// The replay harness re-proposes a recorded move from its record, which
// stores a range-enum field as its string ("4_6"); Go's driver proposed the
// enum's key (input codec 'integer'). Found replaying the first checkers
// documents in which a Move Token is applied: the engine refused every
// recorded Move Token with Go's "cannot unmarshal string" text.
test('proposedInput turns a recorded range-enum move field back into the key Go proposed', () => {
	const { game } = buildCheckers();
	const vectors = JSON.parse(readFileSync(MOVE_FIELD_VECTORS, 'utf8')) as MoveFieldVectors;
	assert.ok(vectors.Records.length > 0);
	for (const record of vectors.Records) {
		const input = proposedInput(game, vectors.Move, { ...record.Blob, TargetPlayerIndex: 0 });
		assert.equal(input[vectors.Field], record.Value, 'record ' + JSON.stringify(record.Blob));
		assert.equal(Object.prototype.hasOwnProperty.call(input, 'TargetPlayerIndex'), false, 'a context-owned field is not proposed');
	}
});

test('componentInstance: the handle a component method has for its own component', { skip: existsSync(VECTORS) ? false : VECTORS + ' is absent' }, () => {
	const { game } = buildCheckers();
	const vectors = readGzipJSON<Vectors>(VECTORS);
	const state = stateOf(game, vectors, vectors.positions[0]);
	const spaces = gameStack(state, 'Spaces');
	let checked = 0;
	for (let slot = 0; slot < stackView(spaces).len(); slot++) {
		const c = componentAt(spaces, slot);
		if (c === null) continue;
		const handle = componentInstance(state, c.deck, c.index);
		assert.deepEqual({ deck: handle.deck, index: handle.index, slot: handle.slot, at: handle.stack.at }, { deck: c.deck, index: c.index, slot, at: spaces.at });
		const [stack, at, err] = containingStack(state, handle);
		assert.equal(err, null);
		assert.equal(at, slot);
		assert.deepEqual(stack?.at, spaces.at);
		checked++;
	}
	assert.ok(checked > 0);
	assert.throws(() => componentInstance(state, 'NoSuchDeck', 0), /no deck NoSuchDeck/);
	assert.throws(() => componentInstance(state, 'Tokens', 24), /has no component 24/);
});
