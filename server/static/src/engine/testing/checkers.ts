// Test support shared by rules/checkers.test.ts (checkers through the loader
// with hand-written stand-in bodies) and games/checkers/checkers.gen.test.ts
// (the translated bodies): the Go recordings of checkers and the code that
// replays them against a compiled game.
//
//  - CHECKERS_DOCUMENTS: three Go fuzz-driver documents
//    (rules/testdata/checkers/seed-000{0,1,6}.json.gz: unseated; seated; one
//    with a recorded refusal).
//  - CHECKERS_VECTORS and checkPositions: constructed positions
//    (testdata/checkers.vectors.json.gz, from testdata/gen/checkersvec on a
//    real Go checkers state): Move Token's Legal texts and ledgers, Crown
//    Token's defaults and Legal, Finish Turn's Legal, every recorded Apply
//    (error, panic, or the whole state after), the end condition and scores.
//  - spaceIsBlackConstructor: a STAND-IN for the one checkers function the
//    translator refuses, `checkers.(*gameDelegate).ConfigurePredicateConstructors`
//    (predicate-constructor): main.go's "checkers.spaceIsBlack" constructor
//    transliterated by hand onto PredicateContext. It is pending the agent
//    loop; until then both tests load it from here.
//
// Node-only (reads files); not engine code.
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { sprintf } from '../go/str.ts';
import { failT, passVerdict, unknownVerdict } from '../legal/types.ts';
import type { LegalVerdict } from '../legal/types.ts';
import type { CompiledGame, RulesState } from '../rules/state.ts';
import { newRulesState } from '../rules/state.ts';
import { syncTimerRecords } from '../rules/timers.ts';
import type { GamePredicate, MoveInstance, PredicateSpec } from '../rules/types.ts';
import { loadEngineState } from '../state.ts';
import type { StateJSON } from '../state.ts';
import { firstDifference } from './conformance.ts';

const here = path.dirname(fileURLToPath(import.meta.url));

/** Seed 0 (unseated), seed 1 (seated: Seat Player and Activate Inactive Player), seed 6 (one refusal). */
export const CHECKERS_DOCUMENTS: readonly string[] = [0, 1, 6].map(seed => path.join(here, '..', 'rules', 'testdata', 'checkers', 'seed-000' + seed + '.json.gz'));
export const CHECKERS_VECTORS = path.join(here, '..', 'testdata', 'checkers.vectors.json.gz');

export function readGzipJSON<T>(file: string): T {
	return JSON.parse(gunzipSync(readFileSync(file)).toString('utf8')) as T;
}

// ----- the predicate stand-in ----------------------------------------------

/** The FnRef of checkers' predicate constructor (GameData.Predicates[0].Construct). */
export const CHECKERS_PREDICATE_REF = 'checkers.(*gameDelegate).ConfigurePredicateConstructors';

/** components.go spaceIsBlack */
function spaceIsBlack(spaceIndex: number): boolean {
	return spaceIndex % 2 === 0;
}

/** main.go ConfigurePredicateConstructors: "checkers.spaceIsBlack"'s Constructor (STAND-IN, see above). */
export function spaceIsBlackConstructor(spec: PredicateSpec): GamePredicate {
	if (spec.args.length !== 1) throw new Error(sprintf('checkers.spaceIsBlack: requires exactly 1 arg (the space-index field), got %d', spec.args.length));
	const field = spec.args[0];
	let template = spec.message;
	if (template === '') template = 'checkers.black_spaces_only';
	return {
		name: 'checkers.spaceIsBlack',
		args: spec.args,
		reads: [{ path: field, facet: 'values' }],
		cost: 0,
		emittedTemplates: [template],
		emittedBindings: { [template]: [] },
		evaluate(ctx) {
			const [val, propType, err] = ctx.resolvePath(field);
			if (err !== null) return unknownVerdict(err);
			if (propType !== 'TypeEnum') return unknownVerdict('checkers.spaceIsBlack: path ' + field + ' is not an enum property');
			const ev = val as { value: number } | null;
			if (ev === null) return unknownVerdict('checkers.spaceIsBlack: path ' + field + ' resolved to a nil or non-enum value');
			if (spaceIsBlack(ev.value)) return passVerdict();
			return failT(template);
		},
	};
}

// ----- constructed positions ----------------------------------------------

export interface VectorEntry {
	name: string;
	args: string[];
	verdict: { outcome: string; message?: { template: string; bindings?: Record<string, unknown> }; reason?: string };
	serializable: boolean;
	clientEvaluable: boolean;
	fieldDependent: boolean;
	reads: { path: string; facet: string }[];
}

export interface ApplyVector {
	move: string;
	fields: Record<string, unknown>;
	/** Go's Apply error, or "" when it returned nil. */
	error: string;
	/** "panic: <text>" when Apply panicked. */
	panic?: string;
	/** The state after a nil-error Apply. */
	after?: StateJSON;
}

export interface PositionVector {
	name: string;
	version: number;
	state: StateJSON;
	/** [token key, space key, proposer, text index] for Move Token; text -1 is nil. */
	moveLegal: [number, number, number, number][];
	/** [token key, space key, proposer, verdict (as an entry's), entry indexes] */
	moveLedgers: [number, number, number, number, number[]][];
	/** Crown Token: the default ComponentIndex, and [slot, text index] for the admin. */
	crownDefault: number;
	crownLegal: [number, number][];
	finishTurnLegal: number;
	applies: ApplyVector[];
	conditionMet: boolean;
	scores: number[];
}

export interface Vectors {
	game: { id: string; secretSalt: string; numPlayers: number };
	texts: string[];
	entries: VectorEntry[];
	positions: PositionVector[];
}

export function stateOf(game: CompiledGame, vectors: Vectors, position: PositionVector): RulesState {
	const record = { name: 'checkers', id: vectors.game.id, secretSalt: vectors.game.secretSalt, numPlayers: vectors.game.numPlayers, variant: null, agents: ['', ''], seating: false };
	const engine = loadEngineState(game.def, { gameId: record.id, secretSalt: record.secretSalt }, position.version, position.state);
	syncTimerRecords(engine);
	return newRulesState(game, engine, record, [], { pendingSeat: null });
}

function moveWith(state: RulesState, name: string, fields: Readonly<Record<string, unknown>>): MoveInstance {
	const move = state.newMove(name);
	if (!move) throw new Error('no move ' + name);
	for (const key of Object.keys(fields)) move.fields[key] = fields[key];
	return move;
}

function spacesString(state: RulesState, key: number): string {
	const e = state.game.enums.get('spaces');
	if (!e) throw new Error('no enum spaces');
	return e.stringOf(key);
}

function guarded(f: () => string | null): string {
	try {
		const text = f();
		return text ?? '';
	} catch (thrown) {
		return 'panic: ' + (thrown instanceof Error ? thrown.message : String(thrown));
	}
}

/** A verdict in the generator's JSON shape (Go omits an empty reason and empty bindings). */
function verdictOf(v: LegalVerdict): VectorEntry['verdict'] {
	const verdict: VectorEntry['verdict'] = { outcome: v.outcome };
	if (v.message) {
		verdict.message = v.message.bindings && Object.keys(v.message.bindings).length > 0
			? { template: v.message.template, bindings: { ...v.message.bindings } }
			: { template: v.message.template };
	}
	if (v.reason) verdict.reason = v.reason;
	return verdict;
}

function entryOf(entry: { name: string; args: readonly string[]; verdict: LegalVerdict; serializable: boolean; clientEvaluable: boolean; fieldDependent: boolean; reads: readonly { path: string; facet: string }[] }): VectorEntry {
	return { name: entry.name, args: [...entry.args], verdict: verdictOf(entry.verdict), serializable: entry.serializable, clientEvaluable: entry.clientEvaluable, fieldDependent: entry.fieldDependent, reads: entry.reads.map(r => ({ path: r.path, facet: r.facet })) };
}

export interface PositionCounts {
	positions: number;
	moveLegal: number;
	legalMoves: number;
	ledgers: number;
	crownLegal: number;
	finishTurn: number;
	applies: number;
	appliedOk: number;
	captures: number;
	ends: number;
}

export function emptyCounts(): PositionCounts {
	return { positions: 0, moveLegal: 0, legalMoves: 0, ledgers: 0, crownLegal: 0, finishTurn: 0, applies: 0, appliedOk: 0, captures: 0, ends: 0 };
}

/**
 * Replays the given positions of the vectors against `game`, adding to
 * `counts`; returns the first (up to 20) differences from Go.
 */
export function checkPositions(game: CompiledGame, vectors: Vectors, positions: readonly PositionVector[], counts: PositionCounts): string[] {
	const failures: string[] = [];
	const fail = (text: string): void => {
		if (failures.length < 20) failures.push(text);
	};
	for (const position of positions) {
		counts.positions++;
		const at = position.name;
		const fresh = (): RulesState => stateOf(game, vectors, position);
		// Move Token: Legal's text for every recorded (token, space, proposer).
		const state = fresh();
		for (const [token, space, proposer, textIndex] of position.moveLegal) {
			counts.moveLegal++;
			const want = textIndex < 0 ? '' : vectors.texts[textIndex];
			if (want === '') counts.legalMoves++;
			const move = moveWith(state, 'Move Token', { TokenIndexToMove: spacesString(state, token), SpaceIndex: spacesString(state, space) });
			const got = guarded(() => state.legalError(move, proposer));
			if (got !== want) fail(at + ' Move Token ' + token + '->' + space + ' by ' + proposer + ': got ' + JSON.stringify(got) + ', want ' + JSON.stringify(want));
		}
		// Ledgers, entry by entry.
		for (const [token, space, proposer, verdictIndex, entryIndexes] of position.moveLedgers) {
			counts.ledgers++;
			const move = moveWith(state, 'Move Token', { TokenIndexToMove: spacesString(state, token), SpaceIndex: spacesString(state, space) });
			const result = state.legalLedger(move, proposer);
			if (!result || !result.entries) {
				fail(at + ' ledger ' + token + '->' + space + ' by ' + proposer + ': no ledger');
				continue;
			}
			const got = { verdict: verdictOf(result.verdict), entries: result.entries.map(entryOf) };
			const want = { verdict: vectors.entries[verdictIndex].verdict, entries: entryIndexes.map(i => vectors.entries[i]) };
			const diff = firstDifference(got, want);
			if (diff !== null) fail(at + ' ledger ' + token + '->' + space + ' by ' + proposer + ': ' + diff);
		}
		// Crown Token: its defaults, and Legal for every slot.
		const crown = state.newMove('Crown Token');
		if (!crown) throw new Error('no move Crown Token');
		if (crown.fields.ComponentIndex !== position.crownDefault) fail(at + ' Crown Token default: got ' + String(crown.fields.ComponentIndex) + ', want ' + position.crownDefault);
		for (const [componentIndex, textIndex] of position.crownLegal) {
			counts.crownLegal++;
			const want = textIndex < 0 ? '' : vectors.texts[textIndex];
			const move = moveWith(state, 'Crown Token', { ComponentIndex: componentIndex });
			const got = guarded(() => state.legalError(move, -2));
			if (got !== want) fail(at + ' Crown Token ' + componentIndex + ': got ' + JSON.stringify(got) + ', want ' + JSON.stringify(want));
		}
		counts.finishTurn++;
		const finishWant = position.finishTurnLegal < 0 ? '' : vectors.texts[position.finishTurnLegal];
		const finishGot = guarded(() => state.legalError(moveWith(state, 'Finish Turn', {}), -2));
		if (finishGot !== finishWant) fail(at + ' Finish Turn: got ' + JSON.stringify(finishGot) + ', want ' + JSON.stringify(finishWant));
		// The end.
		counts.ends++;
		const end = game.spec.end;
		if (!end?.conditionMet || !end.playerScore) throw new Error('the game has no GameEndConditionMet or PlayerScore');
		const playerScore = end.playerScore;
		if (end.conditionMet(state) !== position.conditionMet) fail(at + ' GameEndConditionMet');
		const scores = position.scores.map((_, p) => playerScore(state, p));
		if (firstDifference(scores, position.scores) !== null) fail(at + ' scores: got ' + JSON.stringify(scores) + ', want ' + JSON.stringify(position.scores));
		// Apply, each on a fresh copy of the position.
		for (const apply of position.applies) {
			counts.applies++;
			const working = fresh();
			const move = moveWith(working, apply.move, apply.fields);
			let got: string;
			try {
				working.game.movesByName[apply.move].spec.apply(working, move);
				got = '';
			} catch (thrown) {
				got = thrown instanceof Error ? thrown.message : String(thrown);
			}
			const want = apply.panic !== undefined ? apply.panic.replace(/^panic: /, '') : apply.error;
			const where = at + ' Apply ' + apply.move + ' ' + JSON.stringify(apply.fields);
			if (got !== want) {
				fail(where + ': got ' + JSON.stringify(got) + ', want ' + JSON.stringify(want));
				continue;
			}
			if (apply.after !== undefined) {
				counts.appliedOk++;
				const after = working.engine.toJSON();
				const diff = firstDifference(after, apply.after);
				if (diff !== null) fail(where + ': state after: ' + diff);
				const capturedBefore = (position.state.Players as { CapturedTokens: { Indexes: number[] } }[]).reduce((n, p) => n + p.CapturedTokens.Indexes.length, 0);
				const capturedAfter = (apply.after.Players as { CapturedTokens: { Indexes: number[] } }[]).reduce((n, p) => n + p.CapturedTokens.Indexes.length, 0);
				if (capturedAfter > capturedBefore) counts.captures++;
			}
		}
	}
	return failures;
}
