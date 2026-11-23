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
//   down as an exact text replacement so it can be rerun. The ten made to
//   hand ports were re-expressed on the generated games when those replaced
//   the hand ports: a translated body's text (bodies.gen.ts) or the
//   exported definition's data (definition.gen.ts: a visibility policy, a
//   move order, an alias, an enum default, a sanitize tag), each with the
//   same meaning and the same expected failure (see each `note`).
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
	/** Must occur exactly `occurs` times in the file (default once); each occurrence is replaced. */
	find: string;
	replace: string;
	/** How many times `find` occurs; every occurrence is replaced. Default 1. */
	occurs?: number;
	/**
	 * Instead of a replacement: the block starting at `find` and the block
	 * starting at `swapWith` (each occurring once, each running to the first
	 * `blockEnd` after its start, inclusive) change places. `replace` is unused.
	 */
	swapWith?: string;
	blockEnd?: string;
	/** Where the table's wording and the replacement differ, why. */
	note?: string;
}

/**
 * The prefix of the note on each mutation that was made to a hand port
 * (games/<game>.ts, removed when the generated games replaced them): it now
 * names the text that carries the same behavior in the generated game, a
 * translated body (bodies.gen.ts) or the exported definition's data
 * (definition.gen.ts), and the note says what the hand-port text was.
 */
const HAND_PORT = 'Re-expressed on the generated game; was ';

export const HISTORICAL_MUTATIONS: readonly SourceMutation[] = [
	{
		id: 'blackjack-ace',
		description: 'Blackjack ace rule `>=` to `>`',
		failedOn: 'view.game, view.legality',
		file: 'games/blackjack/bodies.gen.ts',
		find: 'if (currentValue >= 11) {',
		replace: 'if (currentValue > 11) {',
		note: HAND_PORT + 'games/blackjack.ts `if (value >= TARGET_SCORE - 10) break;`; the generated handValue has the constant folded (TargetScore - 10 is 11).',
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
		file: 'games/memory/definition.gen.ts',
		// The move order is data: the two moves' literals in Moves change places.
		find: "\t\t\t{\n\t\t\t\tName: 'Capture Cards',",
		swapWith: "\t\t\t{\n\t\t\t\tName: 'Start Hide Cards Timer',",
		blockEnd: '\n\t\t\t},\n',
		replace: '',
		note: HAND_PORT + 'games/memory.ts, the two fix-up factories swapped in its moves list.',
	},
	{
		id: 'hole-card-policy',
		description: 'Hole card policy `len` to `order`',
		failedOn: 'view.game',
		file: 'games/blackjack/definition.gen.ts',
		find: "HiddenHand: {\n\t\t\t\t\t\tType: 'stack',\n\t\t\t\t\t\tSanitize: {\n\t\t\t\t\t\t\tother: 'len',",
		replace: "HiddenHand: {\n\t\t\t\t\t\tType: 'stack',\n\t\t\t\t\t\tSanitize: {\n\t\t\t\t\t\t\tother: 'order',",
		note: HAND_PORT + "games/blackjack.ts, the HiddenHand prop's `Sanitize: 'len'`.",
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
		file: 'games/tictactoe/definition.gen.ts',
		find: "Slot: {\n\t\t\t\t\t\tType: 'int',\n\t\t\t\t\t\tSanitize: 'self:visible',\n",
		replace: "Slot: {\n\t\t\t\t\t\tType: 'int',\n",
		note: HAND_PORT + "games/tictactoe.ts, Place Token's `fields: { Slot: { type: 'int', sanitize: 'self:visible' } }`.",
	},
	{
		id: 'hidden-name-alias',
		description: 'The hidden-name alias dropped',
		failedOn: 'header, view.move',
		file: 'games/tictactoe/definition.gen.ts',
		// The move's own key (rules/data.ts reads move.HiddenAnimationKey), not the copy in its stock configuration.
		find: "\t\t\t\tHiddenAnimationKey: 'Hidden Action',\n",
		replace: '',
		note: HAND_PORT + "games/tictactoe.ts, Place Token's `hiddenAnimationKey: 'Hidden Action',`.",
	},
	{
		id: 'pig-extra-draw',
		description: "One extra draw before pig's starting player",
		failedOn: "replay stops (wrong player's turn)",
		file: 'games/pig/bodies.gen.ts',
		find: 'let startingPlayer: number = state.rand().intn(state.numPlayers());',
		replace: 'state.rand().intn(state.numPlayers()); let startingPlayer: number = state.rand().intn(state.numPlayers());',
		note: HAND_PORT + 'games/pig.ts `state.g.CurrentPlayer = state.rand().intn(state.numPlayers());`.',
	},
	{
		id: 'error-string',
		description: 'An error string changed by one word',
		failedOn: 'view.legality',
		file: 'games/blackjack/bodies.gen.ts',
		find: "'Current player is already at target scores'",
		replace: "'Current player is already at target score'",
		note: HAND_PORT + 'games/blackjack.ts, the same string.',
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
		// roundRobinParts (apply(state, move)): the round robin of the
		// registry's dealComponents, which the generated blackjack's two deals run.
		find: 'apply(state, move) {\n\t\t\tif (state.g.RRHasStarted !== true) {\n\t\t\t\tconst first = starter(state);\n\t\t\t\tstate.g.RRLastPlayer = state.previous(first);',
		replace: 'apply(state, move) {\n\t\t\tif (state.g.RRHasStarted !== true) {\n\t\t\t\tconst first = starter(state);\n\t\t\t\tstate.g.RRLastPlayer = first;',
		note: 'Moved with the generated games: it was the same line in dealCountComponents\'s own round robin (apply(state)), which only the blackjack hand port ran; against that copy the corpus no longer fails, since no shipped game runs it.',
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
		file: 'rules/data.ts',
		// The loader builds every generated game's progressions from the definition.
		find: 'allowMultiple: move.AllowMultipleInProgression }',
		replace: 'allowMultiple: false }',
		note: "Moved with the generated games: it was rules/stock.ts's installer (`allowMultiple: item.allowMultipleInProgression === true`, addOrderedForPhase), which only hand-written specs use; against that line the corpus no longer fails, since no shipped game runs it.",
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
		file: 'games/werewolf/definition.gen.ts',
		find: "VoteTarget: {\n\t\t\t\t\t\tType: 'playerIndex',\n\t\t\t\t\t\tSanitize: 'self:visible,same-role:visible',",
		replace: "VoteTarget: {\n\t\t\t\t\t\tType: 'playerIndex',\n\t\t\t\t\t\tSanitize: 'self:visible',",
		// The hand port's one vote factory served both votes; the definition writes each move's field.
		occurs: 2,
		note: HAND_PORT + "games/werewolf.ts, the vote factory's `fields: { VoteTarget: { type: 'playerIndex', sanitize: 'self:visible,same-role:visible' } }`, which Cast Vote and Cast Night Vote shared.",
	},
	{
		id: 'tree-enum-default',
		description: 'Tree enum default changed',
		failedOn: 'header, state',
		file: 'games/debuganimations/definition.gen.ts',
		find: "'1': 'Normal',\n\t\t\t\t\t},\n\t\t\t\t\tDefaultValue: 1,",
		replace: "'1': 'Normal',\n\t\t\t\t\t},\n\t\t\t\t\tDefaultValue: 0,",
		note: HAND_PORT + "games/debuganimations.ts, the phase enum's `DefaultValue: 1`.",
	},
	{
		id: 'public-shuffle-secret',
		description: 'Public shuffle made secret',
		failedOn: 'state, view.game',
		file: 'games/debuganimations/bodies.gen.ts',
		find: "return publicShuffle(gameStack(state, 'FanStack'));",
		replace: "return shuffle(gameStack(state, 'FanStack'));",
		note: HAND_PORT + "games/debuganimations.ts `apply: state => state.gameStack('FanStack').publicShuffle(),`.",
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

/** `source` with the mutation applied; throws unless `find` occurs exactly as many times as it must. */
export function applyMutation(source: string, mutation: SourceMutation): string {
	const want = mutation.occurs ?? 1;
	const count = occurrences(source, mutation.find);
	if (count !== want) throw new Error(mutation.id + ': ' + JSON.stringify(mutation.find) + ' occurs ' + count + ' times in ' + mutation.file + ', not ' + (want === 1 ? 'once' : want + ' times'));
	if (mutation.swapWith === undefined) return source.split(mutation.find).join(mutation.replace);
	const other = occurrences(source, mutation.swapWith);
	if (other !== 1) throw new Error(mutation.id + ': ' + JSON.stringify(mutation.swapWith) + ' occurs ' + other + ' times in ' + mutation.file + ', not once');
	const end = mutation.blockEnd;
	if (end === undefined || end === '') throw new Error(mutation.id + ': a swap needs blockEnd');
	const block = (start: string): { at: number; to: number } => {
		const at = source.indexOf(start);
		const close = source.indexOf(end, at + start.length);
		if (close < 0) throw new Error(mutation.id + ': no ' + JSON.stringify(end) + ' after ' + JSON.stringify(start));
		return { at, to: close + end.length };
	};
	const [first, second] = [block(mutation.find), block(mutation.swapWith)].sort((a, b) => a.at - b.at);
	if (first.to > second.at) throw new Error(mutation.id + ': the two blocks overlap');
	return source.slice(0, first.at) + source.slice(second.at, second.to) + source.slice(first.to, second.at) + source.slice(first.at, first.to) + source.slice(second.to);
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
