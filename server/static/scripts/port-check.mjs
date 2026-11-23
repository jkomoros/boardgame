#!/usr/bin/env node
/**
 * The TypeScript half of `boardgame-util port` (fabric/TRANSPILE-PROPOSAL.md,
 * "The agent loop"). boardgame-util writes one JSON job to stdin, waits for
 * this process to exit, and reads one JSON result from stdout. Nothing here
 * writes a file: the generated game under test (its bodies.gen.ts,
 * definition.gen.ts and index.ts, assembled by boardgame-util with the
 * candidate function in place) is served to Node's module loader from
 * memory at its real path, games/<game>/, through node:module registerHooks,
 * and to the TypeScript compiler through a compiler host.
 *
 *   node scripts/port-check.mjs replay   < job.json
 *
 *     Replays documents on a generated game: { game, virtual, docs }.
 *     virtual maps paths relative to src/engine/ to their text (empty: the
 *     checked-in files are used). Result: { loaded, error?, documents:
 *     [{ name, passed, report? }] }.
 *
 *   node scripts/port-check.mjs check    < job.json
 *
 *     Validates one function: { game, tsName, expectedSignature,
 *     outputPath, candidate, importable, callees, virtual, docs, mutation }.
 *     1. static: the candidate file has exactly one exported function, the
 *        required name and signature, imports only what the brief allows,
 *        and passes the sandbox and purity rules (testing/sourceRules.ts);
 *     2. types: the candidate (placed at outputPath) and the assembled game
 *        type-check under src/engine/tsconfig.sandbox.json;
 *     3. differential: every document replays with the function in place
 *        (testing/fuzz.ts runFuzzDocument); the first difference carries the
 *        arguments the function was last called with;
 *     4. mutation: testing/mutate.ts bodyMutants of the function, each run
 *        against the documents until one fails; survivors listed.
 *     Each step runs only if the one before it passed.
 *
 * Bounded by counts, never by time: no timeout, sleep or retry. A mutant
 * that makes a loop of the target function run away is stopped by the loop
 * guard's count (testing/mutate.ts LOOP_GUARD_LIMIT).
 */
import { readFileSync } from 'node:fs';
import { registerHooks } from 'node:module';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { gunzipSync } from 'node:zlib';

const scriptPath = fileURLToPath(import.meta.url);
const staticDir = path.resolve(path.dirname(scriptPath), '..');
const engineDir = path.join(staticDir, 'src', 'engine');

const mode = process.argv[2];
const job = JSON.parse(readFileSync(0, 'utf8'));

// ---------------------------------------------------------------------------
// Virtual modules

/** URL (without query) to text, for the files served from memory. */
const virtual = new Map();
/** URL (without query) of the bodies file to the text of each variant, by query. */
const variants = new Map();

function urlOf(rel) {
	return pathToFileURL(path.join(engineDir, rel)).href;
}

function setVirtual(files) {
	for (const [rel, text] of Object.entries(files ?? {})) virtual.set(urlOf(rel), text);
}

registerHooks({
	resolve(specifier, context, nextResolve) {
		const parent = context.parentURL ?? import.meta.url;
		if (specifier.startsWith('.') || specifier.startsWith('file:')) {
			const target = new URL(specifier, parent);
			const bare = target.href.split('?')[0];
			if (virtual.has(bare)) {
				// A variant (a mutant, the recording copy) carries its query to
				// the virtual files it imports, so each variant is its own module graph
				// over the one shared engine.
				const own = target.href.includes('?') ? '?' + target.href.split('?')[1] : '';
				const query = own !== '' ? own : parent.includes('?') ? '?' + parent.split('?')[1] : '';
				return { url: bare + query, format: 'module-typescript', shortCircuit: true };
			}
		}
		return nextResolve(specifier, context);
	},
	load(url, context, nextLoad) {
		const [bare, query] = url.split('?');
		if (virtual.has(bare)) {
			const key = query ?? '';
			const source = variants.get(bare)?.get(key) ?? virtual.get(bare);
			return { format: 'module-typescript', source, shortCircuit: true };
		}
		return nextLoad(url, context);
	},
});

function setVariant(rel, query, text) {
	const url = urlOf(rel);
	if (!variants.has(url)) variants.set(url, new Map());
	variants.get(url).set(query, text);
}

// ---------------------------------------------------------------------------
// Documents and the game

function readDocument(file) {
	return JSON.parse(gunzipSync(readFileSync(file)).toString('utf8'));
}

const engine = {};
async function loadEngine() {
	engine.rules = await import(urlOf('rules/index.ts'));
	engine.fuzz = await import(urlOf('testing/fuzz.ts'));
	engine.conformance = await import(urlOf('testing/conformance.ts'));
}

/** The game's spec factory from games/<game>/index.ts, for one variant ('' for the plain one). */
async function gameFactory(game, query) {
	const module = await import(urlOf('games/' + game + '/index.ts') + (query ? '?' + query : ''));
	const make = module[game + 'Generated'];
	if (typeof make !== 'function') throw new Error('games/' + game + '/index.ts exports no ' + game + 'Generated');
	return make;
}

function errorText(err) {
	return String(err?.message ?? err);
}

/** Replays every document; stops at the first failure when `stopAtFirst`. */
function replayAll(compiled, make, docs, stopAtFirst, withReport) {
	const out = [];
	for (const file of docs) {
		let doc;
		let entry;
		try {
			doc = readDocument(file);
			const result = engine.fuzz.runFuzzDocument(compiled, doc, withReport ? { spec: make } : {});
			entry = { name: doc.name, file, passed: result.passed };
			if (!result.passed && withReport) entry.report = result.report;
		} catch (err) {
			entry = { name: doc?.name ?? file, file, passed: false, exception: errorText(err).slice(0, 2000) };
		}
		out.push(entry);
		if (stopAtFirst && !entry.passed) break;
	}
	return out;
}

async function compileVariant(game, query) {
	const make = await gameFactory(game, query);
	return { make, compiled: engine.rules.compileGame(make()) };
}

// ---------------------------------------------------------------------------
// replay

async function replay() {
	setVirtual(job.virtual);
	await loadEngine();
	let game;
	try {
		game = await compileVariant(job.game, '');
	} catch (err) {
		return { loaded: false, error: errorText(err), documents: [] };
	}
	return { loaded: true, documents: replayAll(game.compiled, game.make, job.docs, false, true) };
}

// ---------------------------------------------------------------------------
// check: static

const ALLOWED_SPECIFIER = /^\.\.\/\.\.\/(go|lib|rules|legal)\/[A-Za-z0-9-]+\.ts$/;

function paramsAndResult(ts, decl, text) {
	const params = decl.parameters.map(p => (p.type ? p.type.getText(text) : '<untyped>').replace(/\s+/g, ' '));
	const result = decl.type ? decl.type.getText(text).replace(/\s+/g, ' ') : '<untyped>';
	return { params, result };
}

function staticCheck(ts, sourceRules) {
	const problems = [];
	const name = job.outputPath;
	const text = job.candidate;
	const source = ts.createSourceFile(name, text, ts.ScriptTarget.ES2022, true, ts.ScriptKind.TS);
	const exported = [];
	let fn = null;
	for (const statement of source.statements) {
		const line = source.getLineAndCharacterOfPosition(statement.getStart(source)).line + 1;
		if (ts.isImportDeclaration(statement)) {
			const specifier = statement.moduleSpecifier.text;
			const clause = statement.importClause;
			const names = [];
			if (clause?.name) problems.push(`${name}:${line} has a default import; import names only`);
			if (clause?.namedBindings && ts.isNamespaceImport(clause.namedBindings)) problems.push(`${name}:${line} has a namespace import; import names only`);
			if (clause?.namedBindings && ts.isNamedImports(clause.namedBindings)) {
				for (const element of clause.namedBindings.elements) {
					if (element.propertyName) problems.push(`${name}:${line} renames ${element.propertyName.text} on import`);
					names.push(element.name.text);
				}
			}
			if (specifier === './bodies.gen.ts') {
				for (const n of names) if (!job.callees.includes(n)) problems.push(`${name}:${line} imports ${n} from ./bodies.gen.ts, which is not one of the functions it calls (${job.callees.join(', ') || 'none'})`);
				continue;
			}
			if (!ALLOWED_SPECIFIER.test(specifier)) {
				problems.push(`${name}:${line} imports ${specifier}; allowed are ../../go/*.ts and the names the brief lists`);
				continue;
			}
			const module = specifier.slice('../../'.length);
			for (const n of names) {
				const allowed = job.importable[n];
				if (allowed === undefined || allowed.replace(/^type:/, '') !== module) {
					problems.push(`${name}:${line} imports ${n} from ${specifier}; ${allowed === undefined ? 'not a name the brief allows' : 'it is exported by ../../' + allowed.replace(/^type:/, '')}`);
				}
			}
			continue;
		}
		if (ts.isFunctionDeclaration(statement)) {
			const isExported = statement.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword) ?? false;
			if (!isExported) problems.push(`${name}:${line} declares a function that is not exported; the file holds exactly one function`);
			else exported.push(statement);
			continue;
		}
		problems.push(`${name}:${line} holds a statement that is neither an import nor the function`);
	}
	if (exported.length !== 1) problems.push(`${name} has ${exported.length} exported functions; it must have exactly one, ${job.tsName}`);
	else {
		fn = exported[0];
		if (fn.name?.text !== job.tsName) problems.push(`${name}: the function is ${fn.name?.text}; it must be ${job.tsName}`);
		if (fn.asteriskToken) problems.push(`${name}: the function is a generator`);
		if (fn.modifiers?.some(m => m.kind === ts.SyntaxKind.AsyncKeyword)) problems.push(`${name}: the function is async`);
		const expectedFile = ts.createSourceFile('expected.ts', job.expectedSignature + ' }', ts.ScriptTarget.ES2022, true, ts.ScriptKind.TS);
		const expectedDecl = expectedFile.statements.find(s => ts.isFunctionDeclaration(s));
		const want = paramsAndResult(ts, expectedDecl, expectedFile);
		const got = paramsAndResult(ts, fn, source);
		if (JSON.stringify(want) !== JSON.stringify(got)) {
			problems.push(`${name}: the signature is (${got.params.join(', ')}): ${got.result}; it must be (${want.params.join(', ')}): ${want.result}, as in: ${job.expectedSignature}`);
		}
	}
	const outputDir = path.dirname(path.join(engineDir, job.outputPath));
	const escapes = specifier => path.relative(engineDir, path.resolve(outputDir, specifier)).startsWith('..');
	problems.push(...sourceRules.sandboxProblems(ts, name, text, escapes));
	for (const [what, found] of sourceRules.platformUses(text)) problems.push(`${name} uses ${what}: ${found}`);
	if (sourceRules.loadsDynamically(text)) problems.push(`${name} loads a module dynamically`);
	let functionText = null;
	if (fn !== null) functionText = text.slice(fn.getStart(source), fn.getEnd()) + '\n';
	return { ok: problems.length === 0, problems, functionText };
}

// ---------------------------------------------------------------------------
// check: types

function typeCheck(ts) {
	const configPath = path.join(engineDir, 'tsconfig.sandbox.json');
	const parsed = ts.getParsedCommandLineOfConfigFile(configPath, {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: d => { throw new Error(ts.flattenDiagnosticMessageText(d.messageText, '\n')); } });
	const files = new Map();
	for (const [rel, text] of Object.entries(job.virtual)) files.set(path.join(engineDir, rel), text);
	files.set(path.join(engineDir, job.outputPath), job.candidate);
	const host = ts.createCompilerHost(parsed.options, true);
	const fileExists = host.fileExists.bind(host);
	const readFile = host.readFile.bind(host);
	const getSourceFile = host.getSourceFile.bind(host);
	host.fileExists = f => files.has(path.resolve(f)) || fileExists(f);
	// The generated game's directory may exist only in memory.
	const dirs = new Set([...files.keys()].map(f => path.dirname(f)));
	const directoryExists = host.directoryExists?.bind(host);
	host.directoryExists = d => dirs.has(path.resolve(d)) || (directoryExists ? directoryExists(d) : ts.sys.directoryExists(d));
	host.readFile = f => files.get(path.resolve(f)) ?? readFile(f);
	host.getSourceFile = (f, language, onError, shouldCreate) => {
		const text = files.get(path.resolve(f));
		if (text !== undefined) return ts.createSourceFile(f, text, language, true);
		return getSourceFile(f, language, onError, shouldCreate);
	};
	// Only what the candidate and the game under test reach: the engine
	// modules they import, not every engine source (those the sandbox type
	// check covers on its own).
	const roots = [...files.keys()];
	const program = ts.createProgram({ rootNames: roots, options: parsed.options, host });
	const diagnostics = ts.getPreEmitDiagnostics(program).filter(d => d.file === undefined || files.has(path.resolve(d.file.fileName)));
	const text = diagnostics.map(d => {
		const message = ts.flattenDiagnosticMessageText(d.messageText, '\n');
		if (d.file === undefined) return message;
		const { line, character } = d.file.getLineAndCharacterOfPosition(d.start ?? 0);
		return path.relative(engineDir, d.file.fileName) + ':' + (line + 1) + ':' + (character + 1) + ' ' + message;
	});
	return { ok: diagnostics.length === 0, diagnostics: text };
}

// ---------------------------------------------------------------------------
// check: the arguments the function was last called with

function summarize(value, depth = 0) {
	if (value === null || typeof value !== 'object') {
		if (typeof value === 'function') return '<function>';
		if (value === undefined) return '<undefined>';
		return value;
	}
	if (typeof value.numPlayers === 'function' && typeof value.timer === 'function') return '<RulesState>';
	if (depth > 4) return '<…>';
	if (Array.isArray(value)) return value.slice(0, 50).map(v => summarize(v, depth + 1));
	const out = {};
	for (const key of Object.keys(value).slice(0, 50)) out[key] = summarize(value[key], depth + 1);
	return out;
}

/** The bodies text with the target wrapped so each call's arguments are recorded. */
function recordingBodies(bodies) {
	const declaration = 'export function ' + job.tsName + '(';
	if (!bodies.includes(declaration)) return null;
	return bodies.replace(declaration, 'function __portTarget(') +
		'\nexport function ' + job.tsName + '(...args: any[]): any {\n\t(globalThis as any).__portCalls(args);\n\treturn (__portTarget as any)(...args);\n}\n';
}

async function lastArguments(report, file) {
	const rel = 'games/' + job.game + '/bodies.gen.ts';
	const text = recordingBodies(job.virtual[rel]);
	if (text === null) return { calls: 0 };
	setVariant(rel, 'record', text);
	let calls = 0;
	let last = null;
	globalThis.__portCalls = args => {
		calls++;
		last = args;
	};
	try {
		const { compiled } = await compileVariant(job.game, 'record');
		const doc = readDocument(file);
		const version = report.version;
		doc.versions = doc.versions.slice(0, version + 1);
		if (report.field === 'rejection') {
			// The refused proposal at that version is the last input.
			doc.rejections = (doc.rejections ?? []).filter(r => r.atVersion <= version);
			engine.fuzz.runFuzzDocument(compiled, doc);
		} else {
			// One replay of the outside inputs (runFuzzDocument feeds them twice).
			engine.conformance.runDocument(compiled, doc);
		}
	} catch (err) {
		return { calls, error: errorText(err) };
	} finally {
		delete globalThis.__portCalls;
	}
	return { calls, arguments: last === null ? null : summarize(last) };
}

// ---------------------------------------------------------------------------
// check: mutation

async function mutation(ts, mutate) {
	const rel = 'games/' + job.game + '/bodies.gen.ts';
	const bodies = job.virtual[rel];
	const generated = mutate.bodyMutants(ts, rel, bodies, job.tsName);
	const run = async (query, text) => {
		setVariant(rel, query, text);
		globalThis[mutate.LOOP_GUARD_COUNTER] = 0;
		try {
			const { compiled } = await compileVariant(job.game, query);
			const results = replayAll(compiled, null, job.docs, true, false);
			const failed = results.find(r => !r.passed);
			return failed === undefined ? null : failed.name + (failed.exception ? ' (threw)' : '');
		} catch (err) {
			return 'load error: ' + errorText(err).slice(0, 200);
		}
	};
	const guarded = await run('guarded', generated.guarded);
	if (guarded !== null) return { ok: false, error: 'the function with only its loop guards fails: ' + guarded };
	const rows = [];
	for (const mutant of generated.mutants) {
		const by = await run('m' + mutant.id, mutant.source);
		rows.push({ id: mutant.id, kind: mutant.kind, line: mutant.line - generated.startLine + 1, before: mutant.before, after: mutant.after, killed: by !== null, by });
	}
	return {
		ok: true, total: generated.total, killed: rows.filter(r => r.killed).length,
		survivors: rows.filter(r => !r.killed), skippedDeclarations: generated.skippedDeclarations,
	};
}

// ---------------------------------------------------------------------------
// check

async function check() {
	const ts = (await import('typescript')).default;
	const sourceRules = await import(urlOf('testing/sourceRules.ts'));
	const out = {};
	out.static = staticCheck(ts, sourceRules);
	if (!out.static.ok) return out;
	out.types = typeCheck(ts);
	if (!out.types.ok) return out;

	setVirtual(job.virtual);
	await loadEngine();
	let game;
	try {
		game = await compileVariant(job.game, '');
	} catch (err) {
		out.load = { ok: false, error: errorText(err) };
		return out;
	}
	out.load = { ok: true };
	const documents = replayAll(game.compiled, game.make, job.docs, true, true);
	const failed = documents.find(d => !d.passed);
	out.differential = { passed: failed === undefined, documents: documents.length, of: job.docs.length };
	if (failed !== undefined) {
		out.differential.first = failed;
		if (failed.report) out.differential.lastCall = await lastArguments(failed.report, failed.file);
		return out;
	}
	if (job.mutation) out.mutation = await mutation(ts, await import(urlOf('testing/mutate.ts')));
	return out;
}

// ---------------------------------------------------------------------------

let result;
if (mode === 'replay') result = await replay();
else if (mode === 'check') result = await check();
else throw new Error('usage: port-check.mjs replay|check < job.json');
process.stdout.write(JSON.stringify(result));
