// Source mutations for the fuzz corpus's adequacy test
// (fabric/TRANSPILE-PROPOSAL.md, "Mutations").
//
// Nothing here touches the file system: a mutation is a function from a
// file's text to new text, and the runner (scripts/fuzz-mutations.mjs)
// substitutes it while the module loads, in a child process, so no mutated
// file is ever written.
//
// Two kinds:
//
// - HISTORICAL_MUTATIONS: the 20 single-line mutations the engine-rules
//   session made by hand to the hand ports and the engine
//   (fabric/status/engine-rules.md, "Is the runner honest?"), each written
//   down as an exact text replacement so it can be rerun.
// - bodyMutants: the routine generator. For one function of one file, every
//   comparison flipped, every integer literal shifted by one, and every
//   statement deleted, one at a time.

// The compiler module's types, by a type query: this file never imports
// TypeScript at run time (the caller passes the module in), and a static
// import would also count as a client dependency (src/vite-optimize-deps.test.ts).
type TSModule = typeof import('typescript');
type TSNode = import('typescript').Node;
type TSSourceFile = import('typescript').SourceFile;
type TSIterationStatement = import('typescript').IterationStatement;

/** One exact text replacement in one engine source file. */
export interface SourceMutation {
	id: string;
	/** The mutation as fabric/status/engine-rules.md names it. */
	description: string;
	/** What it failed on against the 22-document conformance corpus, from that table. */
	failedOn: string;
	/** Relative to src/engine/. */
	file: string;
	/** Must occur exactly once in the file. */
	find: string;
	replace: string;
	/** Where the table's wording and the replacement differ, why. */
	note?: string;
}

export const HISTORICAL_MUTATIONS: readonly SourceMutation[] = [
	{
		id: 'blackjack-ace',
		description: 'Blackjack ace rule `>=` to `>`',
		failedOn: 'view.game, view.legality',
		file: 'games/blackjack.ts',
		find: 'if (value >= TARGET_SCORE - 10) break;',
		replace: 'if (value > TARGET_SCORE - 10) break;',
	},
	{
		id: 'null-winners',
		description: 'The null-winners rule removed',
		failedOn: 'status, view.game',
		file: 'testing/conformance.ts',
		find: 'return value === null ? [] : value;',
		replace: 'return value as T;',
	},
	{
		id: 'memory-fixup-order',
		description: "Memory's two fix-ups swapped",
		failedOn: 'header, then the fix-up chain',
		file: 'games/memory.ts',
		find: 'captureCards(),\n\t\t\tstartHideCardsTimer(),',
		replace: 'startHideCardsTimer(),\n\t\t\tcaptureCards(),',
	},
	{
		id: 'hole-card-policy',
		description: 'Hole card policy `len` to `order`',
		failedOn: 'view.game',
		file: 'games/blackjack.ts',
		find: "HiddenHand: { Type: 'stack', Deck: 'cards', MaxSize: 1, Sanitize: 'len' },",
		replace: "HiddenHand: { Type: 'stack', Deck: 'cards', MaxSize: 1, Sanitize: 'order' },",
	},
	{
		id: 'timer-id-draws',
		description: 'Timer id drawn with 15 draws, not 16',
		failedOn: 'state, view.game',
		file: 'rules/state.ts',
		find: 'while (id.length < TIMER_ID_LENGTH) id +=',
		replace: 'while (id.length < TIMER_ID_LENGTH - 1) id +=',
	},
	{
		id: 'computed-unsanitized',
		description: 'Computed properties from the unsanitized state',
		failedOn: 'view.game',
		file: 'rules/view.ts',
		find: 'const rules = newRulesState(game, sanitized, at.record, at.history, { pendingSeat: null });',
		replace: 'const rules = newRulesState(game, state.engine, at.record, at.history, { pendingSeat: null });',
	},
	{
		id: 'move-field-sanitize',
		description: "A move field's `sanitize` tag dropped",
		failedOn: 'view.move',
		file: 'games/tictactoe.ts',
		find: "fields: { Slot: { type: 'int', sanitize: 'self:visible' } },",
		replace: "fields: { Slot: { type: 'int' } },",
	},
	{
		id: 'hidden-name-alias',
		description: 'The hidden-name alias dropped',
		failedOn: 'header, view.move',
		file: 'games/tictactoe.ts',
		find: "hiddenAnimationKey: 'Hidden Action',",
		replace: '',
	},
	{
		id: 'pig-extra-draw',
		description: "One extra draw before pig's starting player",
		failedOn: "replay stops (wrong player's turn)",
		file: 'games/pig.ts',
		find: 'state.g.CurrentPlayer = state.rand().intn(state.numPlayers());',
		replace: 'state.rand().intn(state.numPlayers()); state.g.CurrentPlayer = state.rand().intn(state.numPlayers());',
	},
	{
		id: 'error-string',
		description: 'An error string changed by one word',
		failedOn: 'view.legality',
		file: 'games/blackjack.ts',
		find: "'Current player is already at target scores'",
		replace: "'Current player is already at target score'",
	},
	{
		id: 'empty-choice-sets',
		description: 'Empty choice sets kept',
		failedOn: 'view.choices',
		file: 'rules/view.ts',
		find: '\t\tif (available === 0) continue;\n',
		replace: '',
	},
	{
		id: 'round-robin-start',
		description: "Round-robin's starting point off by one",
		failedOn: 'the deal order',
		file: 'rules/stock.ts',
		find: 'state.g.RRLastPlayer = state.previous(first);',
		replace: 'state.g.RRLastPlayer = first;',
	},
	{
		id: 'seated-player-active',
		description: 'A seated player left active',
		failedOn: 'the gathering chain',
		file: 'rules/stock.ts',
		find: 'if (state.game.spec.player.traits?.inactive) player.PlayerInactive = true;',
		replace: '',
	},
	{
		id: 'progression-repeatable',
		description: 'Repeatable moves not repeatable in a progression',
		failedOn: 'the deal chain',
		file: 'rules/stock.ts',
		find: 'allowMultiple: item.allowMultipleInProgression === true',
		replace: 'allowMultiple: false',
	},
	{
		id: 'any-player-target',
		description: "`moves.AnyPlayer`'s target not set to the proposer",
		failedOn: 'replay stops',
		file: 'rules/stock.ts',
		find: '\t\tmove.fields.TargetPlayerIndex = proposer;\n',
		replace: '',
	},
	{
		id: 'role-visibility',
		description: 'Role value shown regardless of visibility',
		failedOn: 'view.game',
		file: 'rules/view.ts',
		find: "if (hasRole && valueVisible(i, 'Role')) values.RoleValue",
		replace: 'if (hasRole) values.RoleValue',
	},
	{
		id: 'vote-same-role',
		description: "`same-role` dropped from a vote's visibility",
		failedOn: 'view.move',
		file: 'games/werewolf.ts',
		find: "fields: { VoteTarget: { type: 'playerIndex', sanitize: 'self:visible,same-role:visible' } },",
		replace: "fields: { VoteTarget: { type: 'playerIndex', sanitize: 'self:visible' } },",
	},
	{
		id: 'tree-enum-default',
		description: 'Tree enum default changed',
		failedOn: 'header, state',
		file: 'games/debuganimations.ts',
		find: "enums: { phase: { Values: { 0: '', 1: 'Normal' }, DefaultValue: 1, Parents: { 0: 0, 1: 0 } } },",
		replace: "enums: { phase: { Values: { 0: '', 1: 'Normal' }, DefaultValue: 0, Parents: { 0: 0, 1: 0 } } },",
	},
	{
		id: 'public-shuffle-secret',
		description: 'Public shuffle made secret',
		failedOn: 'state, view.game',
		file: 'games/debuganimations.ts',
		find: "apply: state => state.gameStack('FanStack').publicShuffle(),",
		replace: "apply: state => state.gameStack('FanStack').shuffle(),",
	},
	{
		id: 'timer-move-state',
		description: "Timer's move bound to the committed state, not the one being built",
		failedOn: 'state',
		file: 'rules/state.ts',
		find: 'const move = this.committed.newMove(name);',
		replace: 'const move = this.newMove(name);',
		note: 'The engine binds MoveByName to the committed state, as Go does; the mutation binds it to the state being built (the table names the correct binding).',
	},
];

/** The number of times `find` occurs in `source`. */
export function occurrences(source: string, find: string): number {
	if (find === '') return 0;
	let count = 0;
	for (let at = source.indexOf(find); at >= 0; at = source.indexOf(find, at + find.length)) count++;
	return count;
}

/** `source` with the mutation applied; throws unless `find` occurs exactly once. */
export function applyMutation(source: string, mutation: SourceMutation): string {
	const count = occurrences(source, mutation.find);
	if (count !== 1) throw new Error(mutation.id + ': ' + JSON.stringify(mutation.find) + ' occurs ' + count + ' times in ' + mutation.file + ', not once');
	return source.replace(mutation.find, () => mutation.replace);
}

// ---------------------------------------------------------------------------
// The routine generator

export type MutantKind = 'flip-comparison' | 'shift-literal' | 'delete-statement';

export interface BodyMutant {
	/** 1-based, in source order within the function. */
	id: number;
	kind: MutantKind;
	/** 1-based line and column of the mutated text. */
	line: number;
	column: number;
	before: string;
	after: string;
	/** The whole file, mutated, with the loop guards. */
	source: string;
}

export interface BodyMutants {
	file: string;
	functionName: string;
	/** The function's line span, 1-based. */
	startLine: number;
	endLine: number;
	/** Every mutant, in source order, before any count bound. */
	total: number;
	mutants: BodyMutant[];
	/** Declarations not deleted: removing one only makes a later use throw. */
	skippedDeclarations: number;
	/** The file with only the loop guards: the baseline every mutant is compared with. */
	guarded: string;
}

/**
 * The guard put at the top of every loop body of the target function: a
 * mutant that makes a loop run forever throws after this many iterations
 * (counted across the whole process) instead. A count, not a clock.
 */
export const LOOP_GUARD_LIMIT = 100_000_000;
export const LOOP_GUARD_COUNTER = '__fuzzMutantLoopSteps';
const LOOP_GUARD = 'if (++globalThis.' + LOOP_GUARD_COUNTER + ' > ' + LOOP_GUARD_LIMIT + ") throw new Error('mutant loop bound reached');";

const FLIPS: Readonly<Record<string, string>> = {
	'<': '>=', '<=': '>', '>': '<=', '>=': '<', '===': '!==', '!==': '===', '==': '!=', '!=': '==',
};

interface Edit {
	start: number;
	end: number;
	text: string;
}

function applyEdits(source: string, edits: readonly Edit[]): string {
	const sorted = [...edits].sort((a, b) => b.start - a.start || b.end - a.end);
	let out = source;
	for (const edit of sorted) out = out.slice(0, edit.start) + edit.text + out.slice(edit.end);
	return out;
}

/** The top-level function or `const name = (...) =>` named `name`. */
function findFunction(ts: TSModule, file: TSSourceFile, name: string): TSNode | null {
	for (const statement of file.statements) {
		if (ts.isFunctionDeclaration(statement) && statement.name?.text === name && statement.body) return statement;
		if (ts.isVariableStatement(statement)) {
			for (const decl of statement.declarationList.declarations) {
				if (ts.isIdentifier(decl.name) && decl.name.text === name && decl.initializer && (ts.isArrowFunction(decl.initializer) || ts.isFunctionExpression(decl.initializer))) return decl.initializer;
			}
		}
	}
	return null;
}

function isLoop(ts: TSModule, node: TSNode): node is TSIterationStatement {
	return ts.isForStatement(node) || ts.isForInStatement(node) || ts.isForOfStatement(node) || ts.isWhileStatement(node) || ts.isDoStatement(node);
}

function isInType(ts: TSModule, node: TSNode): boolean {
	for (let at: TSNode | undefined = node.parent; at !== undefined; at = at.parent) {
		if (ts.isTypeNode(at)) return true;
		if (ts.isStatement(at) || ts.isExpressionStatement(at)) return false;
	}
	return false;
}

/**
 * Every mutant of the function `functionName` in `source`. `ts` is the
 * TypeScript compiler module (the caller imports it; the engine does not).
 */
export function bodyMutants(ts: TSModule, fileName: string, source: string, functionName: string): BodyMutants {
	const file = ts.createSourceFile(fileName, source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
	const target = findFunction(ts, file, functionName);
	if (target === null) throw new Error('no top-level function ' + functionName + ' in ' + fileName);

	// Each guard edit remembers the loop it belongs to.
	const guards: (Edit & { loopStart: number; loopEnd: number })[] = [];
	const candidates: { kind: MutantKind; edit: Edit; before: string }[] = [];
	let skippedDeclarations = 0;

	const visit = (node: TSNode): void => {
		if (isLoop(ts, node)) {
			const body = node.statement;
			const loop = { loopStart: node.getStart(file), loopEnd: node.getEnd() };
			if (ts.isBlock(body)) guards.push({ ...loop, start: body.getStart(file) + 1, end: body.getStart(file) + 1, text: ' ' + LOOP_GUARD + ' ' });
			else {
				guards.push({ ...loop, start: body.getStart(file), end: body.getStart(file), text: '{ ' + LOOP_GUARD + ' ' });
				guards.push({ ...loop, start: body.getEnd(), end: body.getEnd(), text: ' }' });
			}
		}
		if (ts.isBinaryExpression(node) && FLIPS[node.operatorToken.getText(file)] !== undefined) {
			const op = node.operatorToken;
			candidates.push({ kind: 'flip-comparison', edit: { start: op.getStart(file), end: op.getEnd(), text: FLIPS[op.getText(file)] }, before: op.getText(file) });
		}
		if (ts.isNumericLiteral(node) && /^\d+$/.test(node.getText(file)) && !isInType(ts, node)) {
			const text = node.getText(file);
			candidates.push({ kind: 'shift-literal', edit: { start: node.getStart(file), end: node.getEnd(), text: String(Number(text) + 1) }, before: text });
		}
		if (ts.isStatement(node) && node !== target && !ts.isBlock(node) && !ts.isEmptyStatement(node)) {
			if (ts.isVariableStatement(node) || ts.isFunctionDeclaration(node) || ts.isClassDeclaration(node) || ts.isTypeAliasDeclaration(node) || ts.isInterfaceDeclaration(node)) skippedDeclarations++;
			else candidates.push({ kind: 'delete-statement', edit: { start: node.getStart(file), end: node.getEnd(), text: ';' }, before: node.getText(file) });
		}
		ts.forEachChild(node, visit);
	};
	ts.forEachChild(target, visit);

	candidates.sort((a, b) => a.edit.start - b.edit.start || a.edit.end - b.edit.end);
	const mutants = candidates.map((candidate, i): BodyMutant => {
		const { line, character } = file.getLineAndCharacterOfPosition(candidate.edit.start);
		// The guards of a loop inside a deleted statement go with it; those of
		// the loop whose unbraced body is the deleted statement stay.
		const kept = guards.filter(g => !(g.loopStart >= candidate.edit.start && g.loopEnd <= candidate.edit.end));
		const after = candidate.kind === 'delete-statement' ? '(deleted)' : candidate.edit.text;
		const before = candidate.before.length > 80 ? candidate.before.slice(0, 77).replace(/\s+/g, ' ') + '...' : candidate.before.replace(/\s+/g, ' ');
		return { id: i + 1, kind: candidate.kind, line: line + 1, column: character + 1, before, after, source: applyEdits(source, [...kept, candidate.edit]) };
	});
	const start = file.getLineAndCharacterOfPosition(target.getStart(file)).line + 1;
	const end = file.getLineAndCharacterOfPosition(target.getEnd()).line + 1;
	return { file: fileName, functionName, startLine: start, endLine: end, total: mutants.length, mutants, skippedDeclarations, guarded: applyEdits(source, guards) };
}
