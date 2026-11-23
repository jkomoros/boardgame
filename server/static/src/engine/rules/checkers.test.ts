// Checkers through the loader: conformance/definitions/checkers.json (the Go
// exporter's output) plus checkers' functions, built by gameFromData and
// compiled by compileGame, held to real Go three ways:
//
//  1. The compiled header (move table, input and choice schemas and their
//     fingerprints, computed descriptors, variant config, chest) equals the
//     header of a Go fuzz-driver document, and every move's compiled plan is
//     the plan Go assembled (MoveData.Plan, from LegalPlanSpecs).
//  2. Three such documents (rules/testdata/checkers/seed-000{0,1,6}.json.gz:
//     unseated; seated; one with a recorded refusal) replay in full: set-up,
//     seating, the 24 Place Token fix-ups, Start Phase Playing and the stuck
//     first playing position, every version, every viewer, legality ledgers
//     included, and the refusal is refused again with Go's text.
//  3. Constructed positions (testdata/checkers.vectors.json.gz, from
//     testdata/gen/checkersvec on a real Go checkers state) reach what no Go
//     game of checkers reaches: Move Token's Legal ledger and Apply
//     (including captures and the finish-minus-start arithmetic), Crown
//     Token's Legal (the token's own Legal method) and Apply, Finish Turn,
//     the end condition and scores.
//
// THE BODIES ARE A STAND-IN for the translator's output: checkers' Go
// transliterated by hand onto the facade (engine/go/), bugs included. The
// FnRefs are the exporter's: checkers.json names every function slot
// (Functions, the move bodies, Traits, SetUp, End, the predicate's
// constructor, ComponentLegal), and patchRefs checks each against
// STAND_IN_REFS, failing with both names on a spelling mismatch and failing
// if checkers.json leaves any slot empty.
//
// The recordings, the replay of the constructed positions and the
// predicate constructor's stand-in (spaceIsBlackConstructor) are in
// ../testing/checkers.ts, shared with games/checkers/checkers.gen.test.ts,
// which holds the TRANSLATED bodies to the same recordings.
//
// To regenerate the Go side:
//
//   ./scripts/go-local run ./boardgame-util fuzz run --game checkers --seeds 0..7 --out DIR
//   cp DIR/checkers/seed-000[016].json.gz server/static/src/engine/rules/testdata/checkers/
//   ./scripts/go-local run ./server/static/src/engine/testdata/gen/checkersvec
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
	type Comp, type GraphRef, ILLEGAL_ENUM_VALUE, componentAt, componentAtKey, componentInstance, containingStack, dynamic, enumEquals, enumKey, first,
	gameStack, legalError, mayMoveToSlot, moveTo, moveToLastSlot, neighborsByKey, player, playerStack, rangeToValue, rangeValue, setEnumKey, stackView,
	swapComponentsByKey, valueToRange,
} from '../go/api.ts';
import type { LegalVerdict } from '../legal/types.ts';
import { loadEngineState } from '../state.ts';
import type { StateJSON } from '../state.ts';
import { type Vectors, CHECKERS_DOCUMENTS, CHECKERS_VECTORS, checkPositions, emptyCounts, readGzipJSON, spaceIsBlackConstructor, stateOf } from '../testing/checkers.ts';
import { type ConformanceDocument, FIELDS, passTable, runDocument } from '../testing/conformance.ts';
import { compileGame, emptyStateJSON, isGameSpecError } from './compile.ts';
import { type Bodies, type GameData, gameFromData } from './data.ts';
import { type Input, createGame, propose } from './engine.ts';
import { parseInstant } from './time.ts';
import { type CompiledGame, type RulesState, newRulesState } from './state.ts';
import type { MoveRecord } from './types.ts';
import type { MoveInstance } from './types.ts';

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '..', '..', '..', '..', '..');
const DEFINITION = path.join(root, 'conformance', 'definitions', 'checkers.json');
const DOCUMENTS = CHECKERS_DOCUMENTS;
const VECTORS = CHECKERS_VECTORS;
/** testdata/gen/facadevec (moves.go): an enum.RangeVal move field of real Go checkers. */
const MOVE_FIELD_VECTORS = path.join(here, '..', 'testdata', 'facade-moves.vectors.json');

// ----- checkers, transliterated onto the facade (stand-in bodies) ---------
//
// examples/checkers, function by function. Comments name the Go. Go's
// types: a *playerState is its index, a ComponentInstance a Comp, a move
// field `m.X` is `m.fields.X`, and an enum.RangeVal move field is read and
// written through the facade's `{ move: m }` Owner (it holds the value's
// string, "r_c", as the engine and Go's move record store it).

const BOARD_WIDTH = 8;
const NUM_TOKENS = 12;
/** behaviors.ColorBlack and ColorRed: the color enum's keys. */
const COLOR_BLACK = 4;
const COLOR_RED = 0;
void COLOR_RED;
const SPACES = 'spaces';
/** graph.MustNewEnumGridConnectedness(spacesEnum, graph.DirectionDown, graph.DirectionDiagonal) */
const GRAPH_DOWNWARD: GraphRef = { enum: SPACES, dimensions: [BOARD_WIDTH, BOARD_WIDTH], directions: ['DirectionDown', 'DirectionDiagonal'] };
const GRAPH_UPWARD: GraphRef = { enum: SPACES, dimensions: [BOARD_WIDTH, BOARD_WIDTH], directions: ['DirectionUp', 'DirectionDiagonal'] };

/** components.go spaceIsBlack */
function spaceIsBlack(spaceIndex: number): boolean {
	return spaceIndex % 2 === 0;
}

/** components.go (*token).Dynamic: `t.ContainingComponent().ImmutableInstance(state).ImmutableDynamicValues()`. */
function tokenDynamic(state: RulesState, t: Comp): Record<string, unknown> {
	return dynamic(state, componentInstance(state, t.deck, t.index));
}

/** `t.Color.Value() == colorBlack` */
function isBlack(state: RulesState, t: Comp): boolean {
	return enumKey(state, { valuesOf: t }, 'Color') === COLOR_BLACK;
}

/** components.go (*token).AllNextSpaces */
function allNextSpaces(state: RulesState, t: Comp, componentIndex: number): number[] | null {
	const fromBottom = isBlack(state, t);
	let nextSpaces: number[] | null = null;
	const crowned = tokenDynamic(state, t).Crowned === true;
	let g = GRAPH_UPWARD;
	let oppositeG = GRAPH_DOWNWARD;
	if (fromBottom) {
		g = GRAPH_DOWNWARD;
		oppositeG = GRAPH_UPWARD;
	}
	for (const val of neighborsByKey(g, componentIndex) ?? []) nextSpaces = [...(nextSpaces ?? []), val];
	if (crowned) {
		for (const val of neighborsByKey(oppositeG, componentIndex) ?? []) nextSpaces = [...(nextSpaces ?? []), val];
	}
	return nextSpaces;
}

/** components.go (*token).FreeNextSpaces */
function freeNextSpaces(state: RulesState, t: Comp, componentIndex: number): number[] | null {
	const spaces = gameStack(state, 'Spaces');
	let result: number[] | null = null;
	for (const space of allNextSpaces(state, t, componentIndex) ?? []) {
		if (componentAt(spaces, space) === null) result = [...(result ?? []), space];
	}
	return result;
}

/** components.go (*token).LegalCaptureSpaces */
function legalCaptureSpaces(state: RulesState, t: Comp, componentIndex: number): number[] | null {
	const spaces = gameStack(state, 'Spaces');
	const nextSpaces = allNextSpaces(state, t, componentIndex);
	let result: number[] | null = null;
	for (const space of nextSpaces ?? []) {
		const c = componentAt(spaces, space);
		if (c === null) continue;
		// `c.Values() == nil` cannot hold for a token.
		if (enumEquals(state, { valuesOf: c }, 'Color', { valuesOf: t }, 'Color')) continue;
		const startIndexes = valueToRange(state, SPACES, componentIndex);
		const endIndexes = valueToRange(state, SPACES, space);
		if (startIndexes === null || endIndexes === null) throw new Error('runtime error: index out of range [0] with length 0');
		const diff = [endIndexes[0] - startIndexes[0], endIndexes[1] - startIndexes[1]];
		const finalIndexes = [endIndexes[0] + diff[0], endIndexes[1] + diff[1]];
		const finalSpace = rangeToValue(state, SPACES, finalIndexes);
		if (finalSpace === ILLEGAL_ENUM_VALUE) continue;
		if (componentAtKey(spaces, finalSpace) === null) result = [...(result ?? []), finalSpace];
	}
	return result;
}

/** components.go (*token).Legal (interfaces.LegalComponent), for Crown Token's moves.DefaultComponent. */
function tokenLegal(state: RulesState, receiver: { deck: string; index: number }, legalType: number | null): string | null {
	void legalType;
	// The receiver: `t.ContainingComponent().ImmutableInstance(state)`.
	const t = componentInstance(state, receiver.deck, receiver.index);
	let targetRow = BOARD_WIDTH - 1;
	if (isBlack(state, t)) targetRow = 0;
	const [, slotIndex, err] = containingStack(state, t);
	if (err !== null) return "Component's position could not be found: " + err;
	const indexes = valueToRange(state, SPACES, slotIndex);
	if (indexes === null) throw new Error('runtime error: index out of range [0] with length 0');
	if (indexes[0] !== targetRow) return 'Not in the target row';
	if (tokenDynamic(state, t).Crowned === true) return 'Already crowned';
	return null;
}

/** main.go DistributeComponentToStarterStack */
function distribute(state: RulesState, deck: string): unknown {
	if (deck === 'Tokens') return gameStack(state, 'UnusedTokens');
	throw new Error('Unknown deck');
}

/** main.go GameEndConditionMet */
function gameEndConditionMet(state: RulesState): boolean {
	for (let p = 0; p < state.numPlayers(); p++) {
		if (stackView(playerStack(state, p, 'CapturedTokens')).numComponents() >= NUM_TOKENS) return true;
	}
	return false;
}

/** state.go (*playerState).GameScore */
function gameScore(state: RulesState, p: number): number {
	return stackView(playerStack(state, p, 'CapturedTokens')).numComponents();
}

/** main.go PlayerScore */
function playerScore(state: RulesState, p: number): number {
	return gameScore(state, p);
}

/** state.go (*playerState).TurnDone */
function turnDone(state: RulesState, p: number): string | null {
	if (player(state, p).FinishedTurn !== true) return 'the player has not yet finished their turn';
	return null;
}

/** state.go (*playerState).ResetForTurnStart */
function resetForTurnStart(state: RulesState, p: number): string | null {
	player(state, p).FinishedTurn = false;
	return null;
}

/** state.go (*playerState).ResetForTurnEnd */
function resetForTurnEnd(state: RulesState, p: number): string | null {
	void state;
	void p;
	return null;
}

/** moves.go (*movePlaceToken).DefaultsForState */
function placeTokenDefaults(state: RulesState, m: MoveInstance): void {
	const unused = gameStack(state, 'UnusedTokens');
	const spaces = gameStack(state, 'Spaces');
	if (stackView(unused).numComponents() <= 0) return;
	const nextToken = componentAt(unused, 0);
	if (nextToken === null) throw new Error('runtime error: invalid memory address or nil pointer dereference');
	let fromBottom = false;
	if (enumKey(state, { valuesOf: nextToken }, 'Color') === COLOR_BLACK) fromBottom = true;
	let startIndex = 0;
	let increment = 1;
	let endCondition = stackView(spaces).len();
	if (fromBottom) {
		startIndex = stackView(spaces).len() - 1;
		increment = -1;
		endCondition = 0;
	}
	for (let i = startIndex; i !== endCondition; i += increment) {
		if (!spaceIsBlack(i)) continue;
		if (componentAt(spaces, i) === null) {
			// `m.TargetIndex.SetValue(enum.EnumKey(i))`, its error discarded.
			setEnumKey(state, { move: m }, 'TargetIndex', i);
			return;
		}
	}
}

/** moves.go (*movePlaceToken).LegalCustom */
function placeTokenLegalCustom(state: RulesState, m: MoveInstance): string | null {
	const firstToken = first(gameStack(state, 'UnusedTokens'));
	const err = mayMoveToSlot(firstToken, gameStack(state, 'Spaces'), enumKey(state, { move: m }, 'TargetIndex'));
	if (err !== null) return err;
	if (!spaceIsBlack(enumKey(state, { move: m }, 'TargetIndex'))) return 'The proposed space is not black';
	return null;
}

/** moves.go (*movePlaceToken).Apply */
function placeTokenApply(state: RulesState, m: MoveInstance): string | null {
	return moveTo(first(gameStack(state, 'UnusedTokens')), gameStack(state, 'Spaces'), enumKey(state, { move: m }, 'TargetIndex'));
}

/** moves.go (*moveMoveToken).LegalCustom */
function moveTokenLegalCustom(state: RulesState, m: MoveInstance): string | null | LegalVerdict {
	const g = gameStack(state, 'Spaces');
	const tokenIndex = enumKey(state, { move: m }, 'TokenIndexToMove');
	const c = componentAtKey(g, tokenIndex);
	// `t := c.Values().(*token)` panics on a nil c.
	if (c === null) throw new Error('runtime error: invalid memory address or nil pointer dereference');
	for (const space of freeNextSpaces(state, c, tokenIndex) ?? []) {
		if (enumKey(state, { move: m }, 'SpaceIndex') === space) return null;
	}
	for (const space of legalCaptureSpaces(state, c, tokenIndex) ?? []) {
		if (enumKey(state, { move: m }, 'SpaceIndex') === space) return null;
	}
	return legalError('checkers.illegal_dest');
}

/** moves.go (*moveMoveToken).Apply */
function moveTokenApply(state: RulesState, m: MoveInstance): string | null {
	const g = gameStack(state, 'Spaces');
	// `state.CurrentPlayer().(*playerState)`
	const p = state.currentPlayerIndex();
	player(state, p);
	const swapErr = swapComponentsByKey(g, enumKey(state, { move: m }, 'TokenIndexToMove'), enumKey(state, { move: m }, 'SpaceIndex'));
	if (swapErr !== null) return "Couldn't move token: " + swapErr;
	const startIndexes = rangeValue(state, { move: m }, 'TokenIndexToMove');
	if (startIndexes === null || startIndexes.length !== 2) return "Couldn't get indexes for token space";
	const finishIndexes = rangeValue(state, { move: m }, 'SpaceIndex');
	if (finishIndexes === null || finishIndexes.length !== 2) return "Couldn't get indexes for finish space";
	const middleIndexes = [finishIndexes[0] - startIndexes[0], finishIndexes[1] - startIndexes[1]];
	const middleSpace = rangeToValue(state, SPACES, middleIndexes);
	if (middleSpace === ILLEGAL_ENUM_VALUE) return 'Invalid result from range to value';
	const c = componentAtKey(g, middleSpace);
	let tokenCaptured = false;
	if (c !== null) {
		if (!enumEquals(state, { valuesOf: c }, 'Color', p, 'Color')) {
			tokenCaptured = true;
			const err = moveToLastSlot(componentAtKey(g, middleSpace), playerStack(state, p, 'CapturedTokens'));
			if (err !== null) return "Couldn't capture token: " + err;
		}
	}
	if (!tokenCaptured) {
		player(state, p).FinishedTurn = true;
	} else {
		const t = componentAtKey(g, enumKey(state, { move: m }, 'SpaceIndex'));
		if (t === null) throw new Error('runtime error: invalid memory address or nil pointer dereference');
		if ((legalCaptureSpaces(state, t, enumKey(state, { move: m }, 'SpaceIndex')) ?? []).length === 0) player(state, p).FinishedTurn = true;
	}
	return null;
}

/** moves.go (*moveCrownToken).Apply */
function crownTokenApply(state: RulesState, m: MoveInstance): string | null {
	const c = componentAt(gameStack(state, 'Spaces'), m.fields.ComponentIndex as number);
	if (c === null) return 'No token at that space';
	dynamic(state, c).Crowned = true;
	return null;
}

// ----- the FnRefs, and patching them into the data -------------------------

/** What the static half is expected to call each function (gamedef FnRef spelling). */
const STAND_IN_REFS = {
	distribute: 'checkers.(*gameDelegate).DistributeComponentToStarterStack',
	conditionMet: 'checkers.(*gameDelegate).GameEndConditionMet',
	playerScore: 'checkers.(*gameDelegate).PlayerScore',
	gameScore: 'checkers.(*playerState).GameScore',
	turnDone: 'checkers.(*playerState).TurnDone',
	resetForTurnStart: 'checkers.(*playerState).ResetForTurnStart',
	resetForTurnEnd: 'checkers.(*playerState).ResetForTurnEnd',
	component: 'checkers.(*token).Legal',
	predicate: 'checkers.(*gameDelegate).ConfigurePredicateConstructors',
	placeTokenDefaults: 'checkers.(*movePlaceToken).DefaultsForState',
	placeTokenLegalCustom: 'checkers.(*movePlaceToken).LegalCustom',
	placeTokenApply: 'checkers.(*movePlaceToken).Apply',
	moveTokenLegalCustom: 'checkers.(*moveMoveToken).LegalCustom',
	moveTokenApply: 'checkers.(*moveMoveToken).Apply',
	crownTokenApply: 'checkers.(*moveCrownToken).Apply',
} as const;

type AnyBody = (...args: never[]) => unknown;

/** The stand-in bodies, by FnRef: a switch, as a generated bodies table is. */
function checkersBodies(ref: string): AnyBody | undefined {
	switch (ref) {
		case STAND_IN_REFS.distribute:
			return distribute;
		case STAND_IN_REFS.conditionMet:
			return gameEndConditionMet;
		case STAND_IN_REFS.playerScore:
			return playerScore;
		case STAND_IN_REFS.gameScore:
			return gameScore;
		case STAND_IN_REFS.turnDone:
			return turnDone;
		case STAND_IN_REFS.resetForTurnStart:
			return resetForTurnStart;
		case STAND_IN_REFS.resetForTurnEnd:
			return resetForTurnEnd;
		case STAND_IN_REFS.component:
			return tokenLegal;
		case STAND_IN_REFS.predicate:
			return spaceIsBlackConstructor;
		case STAND_IN_REFS.placeTokenDefaults:
			return placeTokenDefaults;
		case STAND_IN_REFS.placeTokenLegalCustom:
			return placeTokenLegalCustom;
		case STAND_IN_REFS.placeTokenApply:
			return placeTokenApply;
		case STAND_IN_REFS.moveTokenLegalCustom:
			return moveTokenLegalCustom;
		case STAND_IN_REFS.moveTokenApply:
			return moveTokenApply;
		case STAND_IN_REFS.crownTokenApply:
			return crownTokenApply;
	}
	return undefined;
}

/**
 * The function a slot must name: the data's when it names one (which must
 * then be the stand-in's spelling), else the stand-in's, recorded in
 * `patched`. An empty or absent FnRef is "not exported yet".
 */
function slot(where: string, current: string | undefined, expected: string, patched: string[]): string {
	if (current === undefined || current === '') {
		patched.push(where);
		return expected;
	}
	assert.equal(current, expected, 'checkers.json names ' + JSON.stringify(current) + ' for ' + where + ', but the stand-in body is ' + JSON.stringify(expected) + ': update STAND_IN_REFS');
	return current;
}

/** Move name to the bodies the Go type declares itself. */
const MOVE_BODIES: Record<string, { Apply?: string; LegalCustom?: string; Defaults?: string }> = {
	'Place Token': { Defaults: STAND_IN_REFS.placeTokenDefaults, LegalCustom: STAND_IN_REFS.placeTokenLegalCustom, Apply: STAND_IN_REFS.placeTokenApply },
	'Move Token': { LegalCustom: STAND_IN_REFS.moveTokenLegalCustom, Apply: STAND_IN_REFS.moveTokenApply },
	'Crown Token': { Apply: STAND_IN_REFS.crownTokenApply },
};

/** checkers.json with every function reference the static half will write filled in. */
function patchRefs(data: GameData): { data: GameData; patched: string[] } {
	const patched: string[] = [];
	const player = data.Traits.Player;
	const turn = player.Turn;
	const moves = data.Moves.map(move => {
		const want = MOVE_BODIES[move.Name] ?? {};
		const bodies: Record<string, string> = {};
		for (const key of ['Apply', 'Legal', 'LegalCustom', 'Defaults'] as const) {
			const have = move.Bodies[key];
			const expected = want[key as keyof typeof want];
			if (expected === undefined) {
				assert.ok(have === undefined || have === '', 'checkers.json gives move ' + move.Name + ' a ' + key + ' body (' + String(have) + ') the stand-in does not have');
				continue;
			}
			bodies[key] = slot('Moves[' + move.Name + '].Bodies.' + key, have, expected, patched);
		}
		return { ...move, Bodies: bodies };
	});
	const componentLegal = data.ComponentLegal ?? {};
	for (const deck of Object.keys(componentLegal)) assert.equal(deck, 'Tokens', 'checkers.json has a component Legal for deck ' + deck);
	const out: GameData = {
		...data,
		Traits: {
			...data.Traits,
			Player: {
				...player,
				GameScore: slot('Traits.Player.GameScore', player.GameScore, STAND_IN_REFS.gameScore, patched),
				Turn: {
					TurnDone: slot('Traits.Player.Turn.TurnDone', turn?.TurnDone, STAND_IN_REFS.turnDone, patched),
					ResetForTurnStart: slot('Traits.Player.Turn.ResetForTurnStart', turn?.ResetForTurnStart, STAND_IN_REFS.resetForTurnStart, patched),
					ResetForTurnEnd: slot('Traits.Player.Turn.ResetForTurnEnd', turn?.ResetForTurnEnd, STAND_IN_REFS.resetForTurnEnd, patched),
				},
			},
		},
		Predicates: data.Predicates.map(item => {
			assert.equal(item.Name, 'checkers.spaceIsBlack', 'checkers.json registers an unexpected predicate');
			return { ...item, Construct: slot('Predicates[' + item.Name + '].Construct', item.Construct, STAND_IN_REFS.predicate, patched) };
		}),
		ComponentLegal: { Tokens: slot('ComponentLegal.Tokens', componentLegal.Tokens, STAND_IN_REFS.component, patched) },
		Moves: moves,
		SetUp: { ...data.SetUp, Distribute: slot('SetUp.Distribute', data.SetUp.Distribute, STAND_IN_REFS.distribute, patched) },
		End: {
			...data.End,
			ConditionMet: slot('End.ConditionMet', data.End.ConditionMet, STAND_IN_REFS.conditionMet, patched),
			PlayerScore: slot('End.PlayerScore', data.End.PlayerScore, STAND_IN_REFS.playerScore, patched),
		},
	};
	return { data: out, patched };
}

function loadDefinition(): GameData {
	return JSON.parse(readFileSync(DEFINITION, 'utf8')) as GameData;
}

function buildCheckers(): { game: CompiledGame; data: GameData; patched: string[] } {
	const { data, patched } = patchRefs(loadDefinition());
	const bodies: Bodies = ref => checkersBodies(ref);
	return { game: compileGame(gameFromData(data, bodies)), data, patched };
}

// ----- 1 and 2: the definition, the header, the plans, the document --------

test('checkers.json loads, compiles, and names functions only where the stand-in does', () => {
	const { game, patched } = buildCheckers();
	assert.equal(game.moves.length, 7);
	// The exporter fills every slot: nothing is patched in.
	assert.deepEqual(patched, [], 'slots checkers.json leaves empty');
	// A predicate the data registers with no body is a GameSpecError naming it.
	const raw = loadDefinition();
	const { data } = patchRefs(raw);
	assert.throws(
		() => gameFromData(data, ref => (ref === STAND_IN_REFS.predicate ? undefined : checkersBodies(ref))),
		thrown => isGameSpecError(thrown) && thrown.message === 'gameFromData: no body for predicate ' + STAND_IN_REFS.predicate,
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
	const { data } = patchRefs(loadDefinition());
	const moves = data.Moves.map(move =>
		move.Name !== 'Move Token'
			? move
			: { ...move, Plan: { ...move.Plan, Specs: move.Plan.Specs.map(spec => (spec.name === 'checkers.spaceIsBlack' ? { ...spec, args: ['move.SpaceIndex', 'extra'] } : spec)) } },
	);
	assert.throws(
		() => compileGame(gameFromData({ ...data, Moves: moves }, ref => checkersBodies(ref))),
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
		// The game reaches the playing phase, and then nothing is legal.
		assert.ok(doc.versions.some(v => v.move?.name === 'Start Phase Playing'));
		assert.equal(doc.versions.filter(v => v.move?.name === 'Place Token').length, 24);
		const refused = checkRejections(game, doc);
		assert.deepEqual(refused.mismatches, []);
		rejections += refused.checked;
	}
	assert.equal(rejections, 1, 'seed 6 records one refusal');
	console.log(passTable(results));
});

// ----- 3: constructed positions from testdata/gen/checkersvec -------------

const vectorsPresent = existsSync(VECTORS);


test('constructed positions from real Go: Move Token, Crown Token, Finish Turn, the end', { skip: vectorsPresent ? false : VECTORS + ' is absent' }, () => {
	const { game } = buildCheckers();
	const vectors = readGzipJSON<Vectors>(VECTORS);
	const counts = emptyCounts();
	const failures = checkPositions(game, vectors, vectors.positions, counts);
	console.log('checkers vectors:', JSON.stringify(counts));
	assert.deepEqual(failures, []);
	// The vectors must actually reach what the fuzzer cannot.
	assert.ok(counts.legalMoves > 0, 'no legal Move Token in the vectors');
	assert.ok(counts.appliedOk > 0, 'no Apply succeeded');
	assert.ok(counts.captures > 0, 'no capture');
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

test('componentInstance: the handle a component method has for its own component', () => {
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
