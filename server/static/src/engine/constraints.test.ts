// Stack constraints and by-key stack access against real Go.
// testdata/constraints.vectors.json (testdata/gen/stackvec, constraints.go)
// holds scripted operations on live states of three games whose stack struct
// tags carry constraints: the constraint vector game (every constraint kind,
// testdata/gen/stackvec/cgame), tictactoe (`same(Value)`) and checkers
// (`same(Color)`). Each game's definition, with its real constraint strings
// in PropDef.Constraints, was derived from its live Go manager. The engine
// replays every operation: Go's error text or result, the stacks that
// changed, no change on error. The file also holds Go's error for each
// malformed constraint tag.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { cloneJSON, loadEngineState } from './state.ts';
import type { EngineState, StateJSON, StackRef } from './state.ts';
import { isStackError } from './stack.ts';
import type { StackJSON } from './stack.ts';
import type { GameDefinition, PropDef } from './definition.ts';
import { constraintsFromStackTag, goAtoi, parseConstraintExpr, parseStackConstraints, splitTagFields } from './constraints.ts';

interface Op {
	Op: string;
	Src?: StackRef;
	Dst?: StackRef;
	Index: number;
	Slot: number;
	NoComponent?: boolean;
	Err: string;
	Result?: number;
	Constraint?: string;
	Count?: number;
	OrderDependent?: boolean;
	Changed?: Record<string, StackJSON>;
	SecretMoveCount?: Record<string, number[]>;
}
interface Scenario {
	Game: string;
	GameID: string;
	SecretSalt: string;
	Version: number;
	NumPlayers: number;
	Before: StateJSON;
	Ops: Op[];
	After: StateJSON;
}
interface Vectors {
	Games: Record<string, GameDefinition>;
	Scenarios: Scenario[];
	TagErrors: { Tag: string; Err: string }[];
}

const vectors = JSON.parse(readFileSync(new URL('./testdata/constraints.vectors.json', import.meta.url), 'utf8')) as Vectors;
const { Games, Scenarios, TagErrors } = vectors;

function load(def: GameDefinition, sc: Scenario, json: StateJSON = sc.Before): EngineState {
	return loadEngineState(def, { gameId: sc.GameID, secretSalt: sc.SecretSalt }, sc.Version, json);
}

/** null for Go's nil error, else its text. */
const goErr = (op: Op): string | null => (op.Err === '' ? null : op.Err);

/**
 * Runs one operation. Returns the error a May... form reported (or null),
 * and undefined for an operation that reports by throwing.
 */
function apply(state: EngineState, op: Op): string | null | undefined {
	const src = op.Src && op.Op !== 'componentAtKey' ? state.stack(op.Src) : undefined;
	const dst = op.Dst ? state.stack(op.Dst) : undefined;
	switch (op.Op) {
		case 'move': return void src!.moveComponent(op.Index, dst!, op.Slot);
		case 'secretMove': return void src!.secretMoveComponent(op.Index, dst!, op.Slot);
		case 'moveFirst': return void src!.moveComponentToFirstSlot(op.Index, dst!);
		case 'moveLast': return void src!.moveComponentToLastSlot(op.Index, dst!);
		case 'moveNext': return void src!.moveComponentToNextSlot(op.Index, dst!);
		case 'slideFirst': return void src!.slideToFirstSlot(op.Index);
		case 'slideLast': return void src!.slideToLastSlot(op.Index);
		case 'moveAll': return void src!.moveAllTo(dst!);
		case 'moveCount': return void src!.moveCountTo(dst!, op.Index);
		case 'shuffle': return void src!.shuffle();
		case 'swap': return void src!.swapComponents(op.Index, op.Slot);
		case 'swapByKey': return void src!.swapComponentsByKey(op.Index, op.Slot);
		case 'mayMoveTo': return src!.mayMoveComponentTo(op.Index, dst!);
		case 'mayMoveToSlot': return src!.mayMoveComponentToSlot(op.Index, dst!, op.Slot);
		case 'mayMoveAllTo': return src!.mayMoveAllTo(dst!);
		case 'mayMoveCountTo': return src!.mayMoveCountTo(dst!, op.Index);
		case 'maySwapByKey': return src!.maySwapComponentsByKey(op.Index, op.Slot);
		case 'componentAtKey': {
			const found = state.view(op.Src!).componentAtKey(op.Index);
			assert.equal(found, op.Result ?? null, 'componentAtKey ' + op.Index + ' of ' + JSON.stringify(op.Src));
			return null;
		}
	}
	throw new Error('unknown op ' + op.Op);
}

/** Every stack's JSON by Go owner path: physical stacks, board spaces and merged views. */
function stacksByPath(state: EngineState): Record<string, StackJSON> {
	const out: Record<string, StackJSON> = {};
	for (const stack of state.physicalStacks()) out[stack.path] = stack.toJSON();
	for (const sub of state.subStates) {
		for (const name of Object.keys(sub.def.Props)) {
			if (sub.def.Props[name].Type === 'mergedStack') out[sub.path + '.' + name] = state.mergedView(sub, name, []).toJSON();
		}
	}
	return out;
}

function replay(def: GameDefinition, sc: Scenario): number {
	const state = load(def, sc);
	let expectedStacks = stacksByPath(state);
	let expectedCounts = cloneJSON(state.data.SecretMoveCount ?? {});
	sc.Ops.forEach((op, i) => {
		const label = `${sc.Game} ${sc.GameID} op ${i} ${op.Op} ${JSON.stringify(op.Src)} -> ${JSON.stringify(op.Dst)} [${op.Index}, ${op.Slot}]`;
		let thrown: unknown = null;
		let reported: string | null | undefined;
		try {
			reported = apply(state, op);
		} catch (e) {
			thrown = e;
		}
		if (reported !== undefined) {
			// A May... form or a lookup: it reports, it never throws.
			assert.equal(thrown, null, label + ': unexpected throw ' + (thrown as Error | null)?.message);
			if (op.NoComponent) assert.notEqual(reported, null, label + ': expected a refusal for a missing component');
			else assert.equal(reported, goErr(op), label);
		} else if (op.NoComponent) {
			assert.ok(isStackError(thrown), label + ': expected a refusal for a missing component');
		} else if (op.Err !== '') {
			assert.ok(isStackError(thrown), label + ': expected error ' + op.Err);
			assert.equal((thrown as Error).message, op.Err, label);
		} else {
			assert.equal(thrown, null, label + ': unexpected error ' + (thrown as Error | null)?.message);
		}
		expectedStacks = { ...expectedStacks, ...(op.Changed ?? {}) };
		if (op.SecretMoveCount) expectedCounts = op.SecretMoveCount;
		assert.deepEqual(stacksByPath(state), expectedStacks, label);
		assert.deepEqual(state.data.SecretMoveCount ?? {}, expectedCounts ?? {}, label);
		assert.equal(state.conservationError(), null, label);
	});
	assert.deepEqual(state.toJSON(), sc.After, sc.GameID + ' final state');
	return sc.Ops.length;
}

const isConstraintError = (text: string): boolean =>
	['which exceeds max of', 'has mixed values', 'duplicate value', 'distinct values, which exceeds'].some(marker => text.includes(marker));

test('the vectors cover every constraint kind, refused and accepted, for every operation', () => {
	const cells = new Set<string>();
	for (const sc of Scenarios) {
		if (sc.Game !== 'cgame') continue;
		for (const op of sc.Ops) {
			if (!op.Constraint || op.NoComponent) continue;
			if (op.Err === '') {
				cells.add(op.Op + '|' + op.Constraint + '|accept');
				if ((op.Count ?? 0) > 1) cells.add(op.Op + '|' + op.Constraint + '|acceptMany');
			} else if (isConstraintError(op.Err)) {
				cells.add(op.Op + '|' + op.Constraint + '|refuse');
				if (op.OrderDependent) cells.add(op.Op + '|' + op.Constraint + '|order');
			}
		}
	}
	const moves = ['move', 'secretMove', 'moveFirst', 'moveLast', 'moveNext', 'mayMoveTo', 'mayMoveToSlot'];
	const transfers = ['moveAll', 'moveCount', 'mayMoveAllTo', 'mayMoveCountTo'];
	for (const kind of ['max', 'same', 'unique', 'maxdistinct']) {
		for (const op of [...moves, ...transfers]) {
			assert.ok(cells.has(op + '|' + kind + '|accept'), op + ' accepted into ' + kind);
			assert.ok(cells.has(op + '|' + kind + '|refuse'), op + ' refused by ' + kind);
		}
		for (const op of transfers) {
			// Refused although each component passes alone: only the
			// copy-and-simulate path can find these.
			assert.ok(cells.has(op + '|' + kind + '|order'), op + ' refused by ' + kind + ' only in order');
			assert.ok(cells.has(op + '|' + kind + '|acceptMany'), op + ' of several accepted by ' + kind);
		}
	}
	assert.ok(cells.has('move|combo|refuse'), 'a stack with several constraints');
	for (const game of ['tictactoe', 'checkers']) {
		const ops = Scenarios.filter(sc => sc.Game === game).flatMap(sc => sc.Ops);
		assert.ok(ops.some(op => op.Constraint === 'same' && op.Err === '' && !op.NoComponent && op.Op === 'move'), game + ': an accepted move');
		assert.ok(ops.some(op => op.Constraint === 'same' && isConstraintError(op.Err)), game + ': a refused move');
	}
	const byKey = Scenarios.flatMap(sc => sc.Ops).filter(op => op.Op.includes('ByKey') || op.Op === 'componentAtKey');
	assert.ok(byKey.some(op => op.Op === 'swapByKey' && op.Err === ''));
	assert.ok(byKey.some(op => op.Op === 'swapByKey' && op.Err !== ''));
	assert.ok(byKey.some(op => op.Op === 'maySwapByKey' && op.Err === ''));
	assert.ok(byKey.some(op => op.Op === 'maySwapByKey' && op.Err !== ''));
	assert.ok(byKey.some(op => op.Op === 'componentAtKey' && op.Result !== undefined && op.Src!.Prop === 'Both'), 'on a merged stack');
	assert.ok(byKey.some(op => op.Op === 'componentAtKey' && op.Result === undefined));
});

test('the definitions carry the games\' real constraint strings', () => {
	assert.deepEqual(Games.tictactoe.Player.Props.UnusedTokens.Constraints, ['same(Value)']);
	assert.deepEqual(Games.checkers.Player.Props.CapturedTokens.Constraints, ['same(Color)']);
	assert.equal(Games.checkers.Game.Props.Spaces.Constraints, undefined);
	assert.deepEqual(Games.cgame.Game.Props.TwoSuits.Constraints, ['maxdistinct(Suit, 2)']);
	assert.deepEqual(Games.cgame.Game.Props.Combo.Constraints, ['max(3)', 'unique(component.Color)', 'same(Component.Face)']);
});

test('checkers and tictactoe constraints load into an EngineState and are attached to the right stacks', () => {
	const checkers = Scenarios.find(sc => sc.Game === 'checkers')!;
	const state = load(Games.checkers, checkers);
	for (let p = 0; p < checkers.NumPlayers; p++) {
		const captured = state.playerStack(p, 'CapturedTokens');
		assert.equal(captured.constraints.length, 1);
		assert.deepEqual(
			{ kind: captured.constraints[0].kind, propPath: captured.constraints[0].propPath, prefix: captured.constraints[0].prefix },
			{ kind: 'same', propPath: 'Color', prefix: '' },
		);
	}
	assert.equal(state.gameStack('Spaces').constraints.length, 0);
	assert.equal(state.gameStack('UnusedTokens').constraints.length, 0);
	const ttt = Scenarios.find(sc => sc.Game === 'tictactoe')!;
	const tstate = load(Games.tictactoe, ttt);
	assert.equal(tstate.playerStack(0, 'UnusedTokens').constraints[0].text, 'same(Value)');
	// Copies, and sanitized copies, carry them too.
	assert.equal(state.copy().playerStack(0, 'CapturedTokens').constraints.length, 1);
	assert.equal(state.copySanitized().playerStack(1, 'CapturedTokens').constraints.length, 1);
});

test('every operation matches Go: errors, results, changed stacks, and no change on error', () => {
	let steps = 0;
	for (const sc of Scenarios) steps += replay(Games[sc.Game], sc);
	assert.equal(steps, Scenarios.reduce((n, sc) => n + sc.Ops.length, 0));
	assert.ok(steps > 2000);
});

test('without DeckDef.ComponentProps, component values are read from the deck JSON alone, with the same results', () => {
	const def = cloneJSON(Games.cgame) as { Decks: Record<string, { ComponentProps?: unknown }> } & GameDefinition;
	for (const deck of Object.keys(def.Decks)) delete def.Decks[deck].ComponentProps;
	for (const sc of Scenarios) if (sc.Game === 'cgame') replay(def, sc);
});

test('recorded states load and serialize back unchanged, including one set-up left violating a constraint', () => {
	for (const sc of Scenarios) {
		const def = Games[sc.Game];
		assert.deepEqual(load(def, sc).toJSON(), sc.Before, sc.GameID);
		assert.deepEqual(load(def, sc, sc.After).toJSON(), sc.After, sc.GameID + ' after');
	}
	// Go's set-up put a Red and a Green card into SameColor (same(Color)):
	// distribution to starter stacks is not checked, and neither is loading.
	const sc = Scenarios.find(s => s.Game === 'cgame' && s.Version === 0)!;
	const state = load(Games.cgame, sc);
	const sameColor = state.gameStack('SameColor');
	assert.deepEqual(sameColor.indexes(), [0, 1]);
	// The starter-stack insert the rules layer's set-up uses does not check either.
	const draw = state.gameStack('Draw');
	const moved = draw.removeComponentAt(0);
	sameColor.insertStarterComponent(moved);
	assert.equal(sameColor.numComponents(), 3);
});

test('a malformed constraint fails to load with Go\'s error', () => {
	assert.ok(TagErrors.length >= 14);
	for (const { Tag, Err } of TagErrors) {
		// PropDef.Constraints holds the tag's fields after the deck and the
		// size: the strings struct_inflater.go hands to its constraint parser.
		const fields = splitTagFields(Tag).slice(1).map(f => f.trim()).filter(f => f !== '');
		if (fields.length > 0 && parseConstraintExpr(fields[0]) === null) fields.shift();
		const prop: PropDef = { Type: 'stack', Deck: 'cards', Constraints: fields };
		assert.equal(parseStackConstraints('S', prop, Games.cgame), Err, Tag);
		const def: GameDefinition = { ...Games.cgame, Game: { Props: { ...Games.cgame.Game.Props, S: prop } } };
		const sc = Scenarios[0];
		const json = cloneJSON(sc.Before);
		json.Game.S = { Deck: 'cards', Indexes: [], IDs: [], IDsLastSeen: {}, ShuffleCount: 0 };
		assert.throws(() => load(def, sc, json), { name: 'StateError', message: Err }, Tag);
	}
});

test('parsing helpers', () => {
	assert.deepEqual(splitTagFields('cards,5,maxdistinct(color,2),max(3)'), ['cards', '5', 'maxdistinct(color,2)', 'max(3)']);
	assert.deepEqual(constraintsFromStackTag('Tokens, same(Color) ,BOARD_SIZE'), ['same(Color)']);
	assert.deepEqual(parseConstraintExpr('maxdistinct( Suit , 2 )'), { name: 'maxdistinct', args: ['Suit', '2'] });
	assert.deepEqual(parseConstraintExpr('max()'), { name: 'max', args: [] });
	assert.equal(parseConstraintExpr('(Color)'), null);
	assert.equal(parseConstraintExpr('same(Color'), null);
	assert.equal(goAtoi('+7'), 7);
	assert.equal(goAtoi('-9223372036854775808'), -9223372036854775808);
	assert.equal(goAtoi('9223372036854775808'), 'strconv.Atoi: parsing "9223372036854775808": value out of range');
	assert.equal(goAtoi(''), 'strconv.Atoi: parsing "": invalid syntax');
	assert.equal(goAtoi('1_000'), 'strconv.Atoi: parsing "1_000": invalid syntax');
	// An argument naming a chest constant resolves when the definition
	// carries constants.
	const withConstants = { ...Games.cgame, Constants: { LIMIT: 2, NAME: 'x' } };
	const limited = parseStackConstraints('S', { Type: 'stack', Deck: 'cards', Constraints: ['max(LIMIT)'] }, withConstants);
	assert.ok(typeof limited !== 'string' && limited[0].max === 2);
	assert.equal(
		parseStackConstraints('S', { Type: 'stack', Deck: 'cards', Constraints: ['max(NAME)'] }, withConstants),
		'S was a nil SizedStack and its struct tag was not valid: constraint max error: max constraint argument is not a valid int: NAME is a constant, but not of type int',
	);
	// A property that is a stack cannot be compared the way Go's %v prints it.
	assert.match(
		parseStackConstraints('S', { Type: 'stack', Deck: 'cards', Constraints: ['same(Nest)'] }, Games.cgame) as string,
		/is a stack, board or timer/,
	);
});
