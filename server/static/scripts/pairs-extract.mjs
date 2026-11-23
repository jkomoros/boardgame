#!/usr/bin/env node
/**
 * Splits the six hand ports (src/engine/games/<game>.ts) into stored pairs
 * (fabric/TRANSPILE-PROPOSAL.md, "Stored pairs"): for each Go function of a
 * game's manifest (conformance/definitions/<game>.json `Functions`) that the
 * hand port has a counterpart for, the text of a self-contained
 * `accepted.ts` that exports one function, named as the translator names
 * the Go function (lib/transpile/mangle.go), with the role's calling
 * convention (the table at the top of src/engine/rules/data.ts; a helper has
 * the signature its translation has).
 *
 *   node scripts/pairs-extract.mjs [--game G ...]
 *
 * prints `{ pairs: [...], none: [...] }` as JSON. It writes nothing: the
 * store is written by the Go generator
 * (boardgame-util/lib/transpile/pairs/testdata/gen), which runs this script,
 * and games/pairs.test.ts calls extractPairs() to check the store is what it
 * would write today.
 *
 * How a counterpart is found. For every FnRef the loader resolves, the hand
 * port's function for that role is the one rules/data.test.ts's adapter
 * uses (move apply/legal/legalCustom/defaults, computed, traits, set-up,
 * end, current player). That function object is located in the hand port's
 * source: Node strips types by replacing them with spaces, so the
 * function's toString() is found verbatim in stripTypeScriptTypes(source).
 * A helper (never resolved by the loader) has no runtime slot; SPECIAL
 * names its counterpart, or says there is none.
 *
 * Extraction kinds (meta.json `extraction`):
 *
 *  - mechanical: the hand port's function, its parameters renamed into the
 *    role's signature, its body copied; every module-scope declaration it
 *    reaches (constants, helpers, types) is copied above it and every import
 *    it uses is re-pointed at the engine. Nothing is written by hand.
 *  - wrapped: the hand port's function is copied as above, under its own
 *    name, and the exported function is a minimal wrapper: either a factory
 *    call bound to the move's configuration (the function captures a
 *    factory's options) or a SPECIAL entry's hand-written body (the hand
 *    port's shape differs from Go's: helpers shared, defaults left to the
 *    engine).
 *  - none: the hand port has no code for it (it leaves the Go function to
 *    an engine default); the body is SPECIAL's, the default written out.
 */
import { readFileSync } from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import ts from 'typescript';

const scriptPath = fileURLToPath(import.meta.url);
const staticDir = path.resolve(path.dirname(scriptPath), '..');
const engineDir = path.join(staticDir, 'src', 'engine');
const repoRoot = path.resolve(staticDir, '..', '..');
export const STORE_DIR = path.join(repoRoot, 'boardgame-util', 'lib', 'transpile', 'testdata', 'pairs');
export const GAMES = ['memory', 'blackjack', 'tictactoe', 'pig', 'debuganimations', 'werewolf'];

// ---------------------------------------------------------------------------
// lib/transpile/mangle.go, ported. The Go store test checks every pair's
// directory name against transpile.Mangle.

function goIdentifier(s) {
	return /^[A-Za-z][A-Za-z0-9_]*$/.test(s) && !s.endsWith('_');
}

const escapeUnderscores = s => s.replace(/_/g, '__');

/** The TypeScript name the translator gives the function a FnRef names. */
export function mangle(ref) {
	let base = ref;
	let literal = 0;
	const lit = /\.func([1-9][0-9]*)$/.exec(ref);
	if (lit) {
		literal = Number(lit[1]);
		base = ref.slice(0, lit.index);
	}
	const dot = base.indexOf('.');
	if (dot <= 0) throw new Error('mangle: FnRef ' + ref + ' has no package');
	const pkg = base.slice(0, dot);
	const rest = base.slice(dot + 1);
	let parts;
	if (rest.startsWith('(')) {
		const closing = rest.indexOf(').');
		if (closing < 0) throw new Error('mangle: unterminated receiver in ' + ref);
		let recv = rest.slice(1, closing);
		const method = rest.slice(closing + 2);
		const pointer = recv.startsWith('*');
		recv = recv.replace(/^\*/, '');
		if (!goIdentifier(recv) || !goIdentifier(method)) throw new Error('mangle: ' + ref + ' is not pkg.(*T).M');
		parts = [escapeUnderscores(pkg), escapeUnderscores(recv) + (pointer ? '' : '$v'), escapeUnderscores(method)];
	} else {
		if (!goIdentifier(rest)) throw new Error('mangle: ' + ref + ' is not pkg.fn');
		parts = [escapeUnderscores(pkg), escapeUnderscores(rest)];
	}
	if (!goIdentifier(pkg)) throw new Error('mangle: bad package in ' + ref);
	return parts.join('_') + (literal > 0 ? '$' + literal : '');
}

// ---------------------------------------------------------------------------
// The calling convention (src/engine/rules/data.ts), by role.

const STATE = ['state', 'RulesState'];
const MOVE = ['m', 'MoveInstance'];
const PLAYER = ['player', 'number'];

const ROLES = {
	apply: { params: [STATE, MOVE], returns: 'void | string | null' },
	legal: { params: [STATE, MOVE, ['proposer', 'number'], ['base', '() => string | null']], returns: 'string | null' },
	legalCustom: { params: [STATE, MOVE, ['proposer', 'number']], returns: 'string | null | LegalVerdict' },
	defaults: { params: [STATE, MOVE], returns: 'void' },
	'computed:global': { params: [STATE], returns: 'unknown' },
	'computed:player': { params: [STATE, PLAYER], returns: 'unknown' },
	setUpBegin: { params: [STATE, ['variant', 'Readonly<Record<string, string>>']], returns: 'void | string | null' },
	distribute: { params: [STATE, ['deck', 'string'], ['index', 'number']], returns: 'Stack | { state: RulesState; at: StackRef }' },
	setUpFinish: { params: [STATE], returns: 'void | string | null' },
	conditionMet: { params: [STATE], returns: 'boolean' },
	playerScore: { params: [STATE, PLAYER], returns: 'number' },
	checkGameFinished: { params: [STATE], returns: '[finished: boolean, winners: number[] | null] | { finished: boolean; winners: number[] }' },
	currentPlayerIndex: { params: [STATE], returns: 'number' },
	turnDone: { params: [STATE, PLAYER], returns: 'string | null' },
	resetForTurnStart: { params: [STATE, PLAYER], returns: 'void | string | null' },
	resetForTurnEnd: { params: [STATE, PLAYER], returns: 'void | string | null' },
	gameScore: { params: [STATE, PLAYER], returns: 'number' },
};

/** Where each type a signature may name comes from, relative to the engine. */
const SIGNATURE_TYPES = {
	RulesState: 'rules/state.ts', MoveInstance: 'rules/types.ts', LegalVerdict: 'legal/types.ts', Stack: 'stack.ts', StackRef: 'state.ts', Comp: 'go/api.ts',
};

/**
 * Go functions with no runtime slot in the hand port (helpers), or whose
 * slot the hand port leaves to the engine. Every helper of the six
 * manifests is listed: a missing entry fails the extraction, so a new Go
 * helper gets a decision, not a silent omission.
 *
 * Fields: `from` (a module-scope declaration of the hand port) or `slot`
 * ('computed:<Name>' | 'trait:gameScore'); `params` (the helper's parameter
 * types, in Go order; names are the hand port's) and `returns`, for a
 * mechanical helper; `signature` and `body` for a wrapper; `kind`; `note`.
 */
const SPECIAL = {
	'pig.(*playerState).ResetForTurn': {
		kind: 'mechanical', from: 'resetForTurn', params: ['RulesState', 'number'], returns: 'void',
		note: 'resetForTurn, which the hand port also installs as resetForTurnStart.',
	},
	'tictactoe.checkRunWon': {
		kind: 'wrapped', from: 'runWinner', signature: '(state: RulesState, runState: string[] | null): string', body: 'return runWinner(runState ?? []);',
		note: 'runWinner(run). Go first returns Empty when len(runState) != dim; the hand port has no such check (every run it builds has three tokens), so the wrapper does not add one.',
	},
	'tictactoe.checkGameFinished': {
		kind: 'wrapped', from: 'boardResult', signature: '(state: RulesState, state_: string[] | null): [boolean, string]',
		body: 'const result = boardResult(state_ ?? []);\n\treturn [result.finished, result.winner];',
		note: 'boardResult(tokens), which returns an object where Go returns (finished, winner). Go also returns (false, Empty) when len(state) != totalDim; the hand port always passes nine tokens and does not check.',
	},
	'tictactoe.rowColToIndex': { kind: 'nopair', note: 'inlined: boardResult\'s local `at(row, col)` closure indexes the token list directly' },
	'tictactoe.(*gameState).tokenValueAtIndex': { kind: 'nopair', note: 'inlined: boardTokens(state) reads every slot in one loop' },
	'memory.(*gameState).CurrentPlayerHasCardsToReveal': {
		kind: 'wrapped', slot: 'computed:CurrentPlayerHasCardsToReveal', params: ['RulesState'], returns: 'boolean',
		note: 'the hand port inlines it into the computed property CurrentPlayerHasCardsToReveal, which is copied here (the same text as ConfigureComputedProperties.func1).',
	},
	'memory.(*gameDelegate).PlayerScore': {
		kind: 'wrapped', slot: 'trait:gameScore',
		note: 'the hand port leaves end.playerScore unset, so the engine scores with the gameScore trait (rules/engine.ts checkGameFinished); Go\'s PlayerScore returns pState.GameScore(). The trait\'s function is copied (the same text as (*playerState).GameScore).',
	},
	'blackjack.(*playerState).HandValue': {
		kind: 'mechanical', from: 'handValue', params: ['RulesState', 'number'], returns: 'number',
		note: 'handValue(state, player), which the hand port also installs as the computed property HandValue.',
	},
	'blackjack.handValue': {
		kind: 'nopair',
		note: 'no counterpart at its signature: Go\'s handValue takes the hand\'s components; the hand port\'s handValue(state, player) is (*playerState).HandValue with this helper inlined',
	},
	'debuganimations.(*gameDelegate).CheckGameFinished': {
		kind: 'none', signature: '(_state: RulesState): [finished: boolean, winners: number[] | null]', body: 'return [false, null];',
		note: 'the hand port has no `end` ("Never ends"), which the engine reads as never finished; Go returns false, nil. The body is that default.',
	},
	'werewolf.finishVoteResolution': {
		kind: 'wrapped', from: 'finishVoteResolution', signature: '(state: RulesState): string | null', body: 'finishVoteResolution(state);\n\treturn null;',
		note: 'finishVoteResolution(state) returns nothing; Go returns an error, which the hand port throws instead (from startVoteTimer).',
	},
	'werewolf.populateFellowWolves': {
		kind: 'wrapped', from: 'populateFellowWolves', signature: '(state: RulesState, _players: number[] | null): void', body: 'populateFellowWolves(state);',
		note: 'populateFellowWolves(state) walks every player; Go walks the slice it is passed, which its one caller builds from every player.',
	},
	'werewolf.resolveVotes': {
		kind: 'wrapped', from: 'resolveVotes', signature: '(state: RulesState): string | null', body: 'resolveVotes(state);\n\treturn null;',
		note: 'resolveVotes(state) returns nothing; Go returns an error (from SetCurrentPhase), which the hand port\'s setPhase throws instead.',
	},
	'werewolf.startVoteTimer': {
		kind: 'wrapped', from: 'startVoteTimer', signature: '(state: RulesState): string | null', body: 'startVoteTimer(state);\n\treturn null;',
		note: 'startVoteTimer(state) throws "timed vote resolver was not configured" where Go returns it as an error.',
	},
	'werewolf.timedVotingEnabled': {
		kind: 'mechanical', from: 'timedVoting', params: ['RulesState'], returns: 'boolean', note: 'timedVoting(state).',
	},
	'werewolf.voteForPhase': {
		kind: 'mechanical', from: 'voteForPhase', params: ['RulesState', 'number', 'number'], returns: 'number', note: 'voteForPhase(state, player, phase).',
	},
};

// ---------------------------------------------------------------------------
// Hand-port slots, as rules/data.test.ts's adapter maps them.

function slotTable(data, hand) {
	const table = new Map();
	const add = (ref, fn, slot, role) => {
		if (ref === undefined || ref === '' || fn === undefined) return;
		const existing = table.get(ref);
		if (existing === undefined) table.set(ref, { fn, slot, role, shared: [] });
		else if (existing.fn !== fn) existing.shared.push(slot);
	};
	const handMoves = new Map(hand.moves.map(move => [move.name, move]));
	for (const move of data.Moves) {
		const hm = handMoves.get(move.Name);
		if (!hm) throw new Error('the hand port has no move ' + move.Name);
		const slot = role => 'moves[' + JSON.stringify(move.Name) + '].' + role;
		add(move.Bodies.Apply, hm.apply, slot('apply'), 'apply');
		add(move.Bodies.Legal, hm.legal, slot('legal'), 'legal');
		add(move.Bodies.LegalCustom, hm.legalCustom, slot('legalCustom'), 'legalCustom');
		add(move.Bodies.Defaults, hm.defaults, slot('defaults'), 'defaults');
	}
	const computed = new Map((hand.computed ?? []).map(item => [item.name, item.compute]));
	for (const item of data.Computed) add(item.Compute, computed.get(item.Name), 'computed[' + JSON.stringify(item.Name) + ']', 'computed:' + item.Scope);
	const traits = hand.player.traits ?? {};
	if (data.Traits.Player.GameScore !== undefined && data.Traits.Player.GameScore !== 'score') add(data.Traits.Player.GameScore, traits.gameScore, 'player.traits.gameScore', 'gameScore');
	const turn = data.Traits.Player.Turn;
	if (turn) {
		add(turn.TurnDone, traits.turn?.turnDone, 'player.traits.turn.turnDone', 'turnDone');
		add(turn.ResetForTurnStart, traits.turn?.resetForTurnStart, 'player.traits.turn.resetForTurnStart', 'resetForTurnStart');
		add(turn.ResetForTurnEnd, traits.turn?.resetForTurnEnd, 'player.traits.turn.resetForTurnEnd', 'resetForTurnEnd');
	}
	add(data.Traits.Game.PlayerOrder, hand.game.traits?.playerOrder, 'game.traits.playerOrder', 'playerOrder');
	add(data.SetUp.Begin, hand.setUp.begin, 'setUp.begin', 'setUpBegin');
	add(data.SetUp.Distribute, hand.setUp.distribute, 'setUp.distribute', 'distribute');
	add(data.SetUp.Finish, hand.setUp.finish, 'setUp.finish', 'setUpFinish');
	add(data.End?.ConditionMet, hand.end?.conditionMet, 'end.conditionMet', 'conditionMet');
	add(data.End?.PlayerScore, hand.end?.playerScore, 'end.playerScore', 'playerScore');
	add(data.End?.Check, hand.end?.check, 'end.check', 'checkGameFinished');
	add(data.CurrentPlayerIndex, hand.currentPlayerIndex, 'currentPlayerIndex', 'currentPlayerIndex');
	return { table, computed, traits };
}

// ---------------------------------------------------------------------------
// Source analysis

const COMPILER_OPTIONS = {
	target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
	allowImportingTsExtensions: true, noEmit: true, strict: true, skipLibCheck: true,
};

function isFunctionLike(node) {
	return ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node);
}

/** The function-like node of `file` whose text `fn.toString()` is. */
function locate(file, stripped, fn, what) {
	const text = fn.toString();
	const at = stripped.indexOf(text);
	if (at < 0) return null;
	if (stripped.indexOf(text, at + 1) >= 0) throw new Error(what + ': the function text occurs twice in ' + file.fileName);
	const end = at + text.length;
	let found = null;
	const visit = node => {
		if (node.pos > end || node.end < at) return;
		if (isFunctionLike(node) && node.getEnd() === end) {
			const start = ts.isFunctionDeclaration(node) ? node.getChildren(file).find(c => c.kind === ts.SyntaxKind.FunctionKeyword).getStart(file) : node.getStart(file);
			if (start === at) found = node;
		}
		ts.forEachChild(node, visit);
	};
	visit(file);
	if (found === null) throw new Error(what + ': found the text but no function node at offset ' + at + ' in ' + file.fileName);
	return found;
}

function topDeclaration(file, name) {
	for (const statement of file.statements) {
		if (ts.isFunctionDeclaration(statement) && statement.name?.text === name) return statement;
		if (ts.isVariableStatement(statement)) {
			for (const decl of statement.declarationList.declarations) {
				if (ts.isIdentifier(decl.name) && decl.name.text === name && decl.initializer && (ts.isArrowFunction(decl.initializer) || ts.isFunctionExpression(decl.initializer))) return decl.initializer;
			}
		}
	}
	return null;
}

/** The module-scope statement holding `node`, or null when a function lies between. */
function moduleStatement(node) {
	let at = node;
	while (at.parent && !ts.isSourceFile(at.parent)) {
		if (isFunctionLike(at.parent) || ts.isBlock(at.parent)) return null;
		at = at.parent;
	}
	return at.parent ? at : null;
}

function within(node, outer) {
	return node.pos >= outer.pos && node.end <= outer.end;
}

/** Whether an identifier is a property name, not a reference. */
function isPropertyName(id) {
	const parent = id.parent;
	if (ts.isPropertyAccessExpression(parent) && parent.name === id) return true;
	if ((ts.isPropertyAssignment(parent) || ts.isPropertySignature(parent) || ts.isMethodDeclaration(parent) || ts.isPropertyDeclaration(parent) || ts.isMethodSignature(parent)) && parent.name === id) return true;
	if (ts.isQualifiedName(parent) && parent.right === id) return true;
	if (ts.isBindingElement(parent) && parent.propertyName === id) return true;
	if ((ts.isImportSpecifier(parent) || ts.isExportSpecifier(parent)) && parent.propertyName === id) return true;
	return false;
}

/**
 * What `roots` reach: module-scope statements of the file (transitively),
 * imports (module specifier, imported name, local name, type-only), and the
 * names captured from an enclosing function (which make a copy impossible).
 */
function dependencies(checker, file, roots) {
	const statements = new Set();
	const imports = [];
	const captured = new Set();
	const seen = new Set();
	const walk = (scope, node) => {
		if (ts.isIdentifier(node) && !isPropertyName(node)) {
			let symbol = ts.isShorthandPropertyAssignment(node.parent) && node.parent.name === node ? checker.getShorthandAssignmentValueSymbol(node.parent) : checker.getSymbolAtLocation(node);
			const decl = symbol?.declarations?.[0];
			if (decl && decl.getSourceFile() === file && !within(decl, scope)) {
				if (ts.isImportSpecifier(decl) || ts.isImportClause(decl) || ts.isNamespaceImport(decl)) {
					const importDecl = ts.isImportSpecifier(decl) ? decl.parent.parent.parent : ts.isImportClause(decl) ? decl.parent : decl.parent.parent;
					const clause = importDecl.importClause;
					const typeOnly = clause.isTypeOnly || (ts.isImportSpecifier(decl) && decl.isTypeOnly);
					if (!ts.isImportSpecifier(decl)) throw new Error('only named imports are supported: ' + decl.getText(file));
					imports.push({ module: importDecl.moduleSpecifier.text, name: (decl.propertyName ?? decl.name).text, local: decl.name.text, typeOnly });
				} else {
					const statement = moduleStatement(decl);
					if (statement === null) captured.add(node.text);
					else if (!seen.has(statement)) {
						seen.add(statement);
						statements.add(statement);
						walk(statement, statement);
					}
				}
			}
		}
		ts.forEachChild(node, child => walk(scope, child));
	};
	for (const root of roots) walk(root, root);
	return { statements: [...statements].sort((a, b) => a.pos - b.pos), imports, captured };
}

// ---------------------------------------------------------------------------
// Rendering

function relativeImport(fromDir, target) {
	let rel = path.relative(fromDir, target).split(path.sep).join('/');
	if (!rel.startsWith('.')) rel = './' + rel;
	return rel;
}

function renderImports(pairDir, handFile, imports, extraTypes) {
	const modules = new Map();
	const add = (absolute, name, local, typeOnly) => {
		const spec = relativeImport(pairDir, absolute);
		let group = modules.get(spec);
		if (!group) modules.set(spec, (group = { values: new Set(), types: new Set() }));
		const entry = name === local ? name : name + ' as ' + local;
		(typeOnly ? group.types : group.values).add(entry);
	};
	for (const imp of imports) add(path.resolve(path.dirname(handFile), imp.module), imp.name, imp.local, imp.typeOnly);
	const have = new Set(imports.map(imp => imp.local));
	for (const name of extraTypes) if (!have.has(name)) add(path.join(engineDir, SIGNATURE_TYPES[name]), name, name, true);
	const lines = [];
	for (const spec of [...modules.keys()].sort()) {
		const group = modules.get(spec);
		const values = [...group.values].sort();
		const types = [...group.types].filter(name => !group.values.has(name)).sort();
		if (values.length) lines.push('import { ' + values.join(', ') + " } from '" + spec + "';");
		if (types.length) lines.push('import type { ' + types.join(', ') + " } from '" + spec + "';");
	}
	return lines.join('\n');
}

function signatureTypes(text) {
	return Object.keys(SIGNATURE_TYPES).filter(name => new RegExp('\\b' + name + '\\b').test(text));
}

/** Text of a node with each line after the first moved left by `indent` tabs. */
function dedent(text, indent) {
	const prefix = '\t'.repeat(indent);
	return text.split('\n').map((line, i) => (i > 0 && line.startsWith(prefix) ? line.slice(indent) : line)).join('\n');
}

function lineIndent(file, pos) {
	const { line } = file.getLineAndCharacterOfPosition(pos);
	const start = file.getPositionOfLineAndCharacter(line, 0);
	const m = /^\t*/.exec(file.text.slice(start));
	return m[0].length;
}

function statementText(file, statement) {
	const text = statement.getText(file);
	return dedent(text.replace(/^export\s+/, ''), lineIndent(file, statement.getStart(file)));
}

/** The body of a function node, as the body of a function declared at column 0. */
function bodyText(file, node) {
	const indent = lineIndent(file, node.getStart(file));
	if (ts.isBlock(node.body)) {
		const inner = node.body.getText(file).slice(1, -1);
		return dedent(inner, indent).replace(/^\n/, '').replace(/\n\t*$/, '');
	}
	return '\treturn ' + dedent(node.body.getText(file), indent) + ';';
}

function header(ref, game, kind, slot, note) {
	const lines = [
		'// Stored pair: ' + ref + ' (producer hand, extraction ' + kind + ').',
		'// From server/static/src/engine/games/' + game + '.ts, ' + slot + '.',
	];
	if (note) {
		let current = '//';
		for (const word of note.split(/\s+/)) {
			if (current.length + word.length + 1 > 78) {
				lines.push(current);
				current = '//';
			}
			current += ' ' + word;
		}
		if (current !== '//') lines.push(current);
	}
	lines.push('// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.');
	return lines.join('\n');
}

function assemble(parts) {
	return parts.filter(part => part !== '').join('\n\n') + '\n';
}

// ---------------------------------------------------------------------------

function findFactoryCall(file, factory, moveName) {
	const found = [];
	const visit = node => {
		if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === factory) {
			const arg = node.arguments[0];
			if (arg && ts.isObjectLiteralExpression(arg)) {
				const nameProp = arg.properties.find(p => ts.isPropertyAssignment(p) && ts.isIdentifier(p.name) && p.name.text === 'name');
				if (nameProp && ts.isStringLiteralLike(nameProp.initializer) && nameProp.initializer.text === moveName) found.push(node);
			}
		}
		ts.forEachChild(node, visit);
	};
	visit(file);
	return found;
}

function enclosingFactory(node) {
	for (let at = node.parent; at; at = at.parent) {
		if (ts.isFunctionDeclaration(at) && ts.isSourceFile(at.parent)) return at;
	}
	return null;
}

/**
 * Extracts every pair of the named games (default: all six).
 * Returns { pairs: [{ game, func, role, name, extraction, note, accepted }], none: [{ game, func, role, reason }] }.
 */
export async function extractPairs(games = GAMES) {
	const handFiles = Object.fromEntries(games.map(game => [game, path.join(engineDir, 'games', game + '.ts')]));
	const program = ts.createProgram(Object.values(handFiles), COMPILER_OPTIONS);
	const checker = program.getTypeChecker();
	const pairs = [];
	const none = [];
	for (const game of games) {
		const data = JSON.parse(readFileSync(path.join(repoRoot, 'conformance', 'definitions', game + '.json'), 'utf8'));
		const module = await import(pathToFileURL(handFiles[game]).href);
		const hand = module[game]();
		const file = program.getSourceFile(handFiles[game]);
		const stripped = stripTypeScriptTypes(file.text);
		const { table, computed, traits } = slotTable(data, hand);
		for (const entry of data.Functions) {
			const ref = entry.Func;
			const name = mangle(ref);
			const pairDir = path.join(STORE_DIR, game, name);
			const special = SPECIAL[ref];
			const slot = table.get(ref);
			if (special?.kind === 'nopair') {
				none.push({ game, func: ref, role: entry.Role, reason: special.note });
				continue;
			}
			if (slot === undefined && special === undefined) throw new Error(ref + ': no hand-port slot and no SPECIAL entry (role ' + entry.Role + ')');

			// The function node and where it came from.
			let node = null;
			let where;
			if (special?.from) {
				node = topDeclaration(file, special.from);
				if (node === null) throw new Error(ref + ': no module-scope function ' + special.from + ' in ' + game + '.ts');
				where = special.from;
			} else if (special?.slot) {
				const [kind, slotName] = special.slot.split(':');
				const fn = kind === 'computed' ? computed.get(slotName) : traits[slotName];
				if (fn === undefined) throw new Error(ref + ': the hand port has no ' + special.slot);
				node = locate(file, stripped, fn, ref);
				where = kind === 'computed' ? 'computed[' + JSON.stringify(slotName) + ']' : 'player.traits.' + slotName;
			} else if (slot) {
				node = locate(file, stripped, slot.fn, ref);
				if (node === null) throw new Error(ref + ': the hand port\'s ' + slot.slot + ' is not written in ' + game + '.ts');
				where = slot.slot;
			}
			const kind = special?.kind ?? 'mechanical';
			const noteParts = [];
			if (special?.note) noteParts.push(special.note);
			if (slot && slot.shared.length > 0) noteParts.push('One Go method serves ' + [slot.slot, ...slot.shared].join(' and ') + '; the hand port builds each from the same function expression, copied once.');

			if (kind === 'none') {
				const text = assemble([
					header(ref, game, kind, 'no slot', special.note),
					renderImports(pairDir, handFiles[game], [], signatureTypes(special.signature)),
					'export function ' + name + special.signature + ' {\n\t' + special.body + '\n}',
				]);
				pairs.push({ game, func: ref, role: entry.Role, name, extraction: kind, note: special.note, accepted: text });
				continue;
			}

			const signatureRole = entry.Role === 'computed' ? slot.role : entry.Role;
			let exported;
			let deps;
			let extraction = kind;
			if (special?.body) {
				// A hand-written wrapper over the copied hand-port function.
				deps = dependencies(checker, file, [node]);
				const own = moduleStatement(node);
				if (own && !deps.statements.includes(own)) deps.statements = [...deps.statements, own].sort((a, b) => a.pos - b.pos);
				exported = 'export function ' + name + special.signature + ' {\n\t' + special.body + '\n}';
				deps.extraTypes = signatureTypes(special.signature);
			} else {
				deps = dependencies(checker, file, [node]);
				let paramsSpec;
				let returns;
				if (entry.Role === 'helper') {
					if (!special?.params) throw new Error(ref + ': a helper needs params in SPECIAL');
					paramsSpec = special.params.map((type, i) => [node.parameters[i]?.name.getText(file) ?? '_p' + i, type]);
					returns = special.returns;
				} else {
					const role = ROLES[signatureRole];
					if (role === undefined) throw new Error(ref + ': no calling convention for role ' + signatureRole);
					if (node.parameters.length > role.params.length) throw new Error(ref + ': the hand port takes more parameters than the role');
					paramsSpec = role.params.map(([defaultName, type], i) => [node.parameters[i] ? node.parameters[i].name.getText(file) : '_' + defaultName, type]);
					returns = role.returns;
				}
				if (node.parameters.some(p => !ts.isIdentifier(p.name) || p.initializer || p.dotDotDotToken)) throw new Error(ref + ': a parameter is not a plain name');
				const signature = '(' + paramsSpec.map(([n, t]) => n + ': ' + t).join(', ') + '): ' + returns;
				if (deps.captured.size === 0) {
					// Mechanical: the body itself. If a copied declaration (or the body)
					// refers back to the function's own declaration, it is copied too.
					exported = 'export function ' + name + signature + ' {\n' + bodyText(file, node) + '\n}';
				} else {
					// The function closes over a factory's options: bind the factory
					// to the move's configuration, as the game installs it.
					const factory = enclosingFactory(node);
					const moveName = slot ? /^moves\[(".*")\]\./.exec(slot.slot)?.[1] : undefined;
					if (factory === null || moveName === undefined) throw new Error(ref + ': captures ' + [...deps.captured].join(', ') + ' and is not a move factory\'s member');
					const calls = findFactoryCall(file, factory.name.text, JSON.parse(moveName));
					if (calls.length !== 1) throw new Error(ref + ': ' + calls.length + ' calls of ' + factory.name.text + ' configure ' + moveName);
					const member = /\.(\w+)$/.exec(slot.slot)[1];
					const capturedNames = [...deps.captured].sort();
					deps = dependencies(checker, file, [factory, calls[0]]);
					deps.statements = [...new Set([...deps.statements, factory])].sort((a, b) => a.pos - b.pos);
					const callText = dedent(calls[0].getText(file), lineIndent(file, calls[0].getStart(file)) - 1);
					const role = ROLES[signatureRole];
					const wrapperParams = role.params.map(([n]) => n);
					exported = 'export function ' + name + '(' + role.params.map(([n, t]) => n + ': ' + t).join(', ') + '): ' + role.returns + ' {\n\treturn ' + callText + '.' + member + '!(' + wrapperParams.join(', ') + ');\n}';
					extraction = 'wrapped';
					noteParts.push('The hand port builds this move with ' + factory.name.text + '(options), and its ' + member + ' reads ' + capturedNames.join(', ') + ' from the factory; the wrapper calls the factory with the configuration the game passes and calls its ' + member + '.');
				}
				deps.extraTypes = signatureTypes(signature + ' ' + exported);
			}
			const note = noteParts.join(' ');
			const text = assemble([
				header(ref, game, extraction, where, note),
				renderImports(pairDir, handFiles[game], deps.imports, deps.extraTypes),
				deps.statements.map(statement => statementText(file, statement)).join('\n\n'),
				exported,
			]);
			pairs.push({ game, func: ref, role: entry.Role, name, extraction, note, accepted: text });
		}
	}
	return { pairs, none };
}

if (process.argv[1] && path.resolve(process.argv[1]) === scriptPath) {
	const games = [];
	for (let i = 2; i < process.argv.length; i++) {
		if (process.argv[i] === '--game') games.push(process.argv[++i]);
		else throw new Error('usage: pairs-extract.mjs [--game G ...]');
	}
	const result = await extractPairs(games.length ? games : GAMES);
	process.stdout.write(JSON.stringify(result, null, '\t') + '\n');
}
