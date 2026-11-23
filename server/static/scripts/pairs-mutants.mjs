#!/usr/bin/env node
/**
 * Body mutants of the stored pairs' accepted.ts (fabric/TRANSPILE-PROPOSAL.md,
 * "Mutations", applied to the pair store): for each pair, every comparison
 * of its exported function flipped, every integer literal shifted by one,
 * every statement deleted (src/engine/testing/mutate.ts bodyMutants), one at
 * a time. Each mutant is the game built from its definition and the stored
 * pairs (games/pairs.testkit.test.ts) with that one function replaced by the
 * mutant, replayed against the game's conformance and checked-in fuzz
 * documents until the first that fails. A mutant no document fails is a
 * survivor.
 *
 *   node scripts/pairs-mutants.mjs [--game G ...] [--func FnRef ...] [--jobs N] --json FILE
 *
 * Only the exported function's body is mutated, not the declarations
 * copied beside it (a wrapped pair's factory, a helper it calls). A helper
 * pair is skipped: the loader never resolves a helper, so its accepted.ts
 * cannot change the game (its callers carry their own copy of the hand
 * port's function).
 *
 * No file is written except --json. A mutant is loaded under a query
 * (accepted.ts?pairs-mutant=N) whose source a module hook supplies, so the
 * store on disk is never touched; the run checks that the store's files
 * hash the same before and after. Bounded by mutant and document counts; a
 * mutant that makes a loop run away throws after mutate.ts's fixed
 * iteration count.
 */
import { spawn } from 'node:child_process';
import { createHash } from 'node:crypto';
import { readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { availableParallelism } from 'node:os';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const scriptPath = fileURLToPath(import.meta.url);
const staticDir = path.resolve(path.dirname(scriptPath), '..');
const engineDir = path.join(staticDir, 'src', 'engine');
const repoRoot = path.resolve(staticDir, '..', '..');
const storeDir = path.join(repoRoot, 'boardgame-util', 'lib', 'transpile', 'testdata', 'pairs');

if (process.argv[2] === '--child') await child();
else await main(process.argv.slice(2));

function storePairs() {
	const out = [];
	for (const game of readdirSync(storeDir, { withFileTypes: true })) {
		if (!game.isDirectory()) continue;
		for (const entry of readdirSync(path.join(storeDir, game.name), { withFileTypes: true })) {
			if (!entry.isDirectory()) continue;
			const dir = path.join(storeDir, game.name, entry.name);
			const meta = JSON.parse(readFileSync(path.join(dir, 'meta.json'), 'utf8'));
			out.push({ game: game.name, name: entry.name, func: meta.func, role: meta.role, file: path.join(dir, 'accepted.ts') });
		}
	}
	return out.sort((a, b) => (a.game + a.name < b.game + b.name ? -1 : 1));
}

function sha256(text) {
	return createHash('sha256').update(text).digest('hex');
}

// ---------------------------------------------------------------------------

async function child() {
	const job = JSON.parse(readFileSync(0, 'utf8'));
	const sources = new Map();
	const { registerHooks } = await import('node:module');
	registerHooks({
		load(url, context, nextLoad) {
			const result = nextLoad(url, context);
			const source = sources.get(url);
			if (source === undefined) return result;
			return { ...result, source, shortCircuit: true };
		},
	});
	const ts = (await import('typescript')).default;
	const { bodyMutants } = await import(pathToFileURL(path.join(engineDir, 'testing', 'mutate.ts')).href);
	const kit = await import(pathToFileURL(path.join(engineDir, 'games', 'pairs.testkit.test.ts')).href);
	const docs = new Map();
	const documentsOf = game => {
		if (!docs.has(game)) docs.set(game, kit.corpusDocuments(game).map(entry => ({ file: entry.file, doc: kit.readDocument(entry.file) })));
		return docs.get(game);
	};
	const results = [];
	let counter = 0;
	for (const pair of job.pairs) {
		const source = readFileSync(pair.file, 'utf8');
		const generated = bodyMutants(ts, pair.file, source, pair.name);
		const result = { game: pair.game, func: pair.func, name: pair.name, total: generated.total, killed: 0, survivors: [], acceptedSha256: sha256(source), kills: {} };
		// The baseline: the guarded source, substituted, must pass.
		const run = async text => {
			const url = pathToFileURL(pair.file).href + '?pairs-mutant=' + ++counter;
			sources.set(url, text);
			globalThis.__fuzzMutantLoopSteps = 0;
			try {
				const fn = (await import(url))[pair.name];
				const built = await kit.storeGame(pair.game, new Map([[pair.func, fn]]));
				for (const { file, doc } of documentsOf(pair.game)) {
					const check = kit.checkDocument(built.game, file, doc);
					if (!check.passed) return { passed: false, file, field: check.result.report?.field ?? 'replay' };
				}
				if (!built.table.substituted.includes(pair.func)) return { passed: false, file: null, field: 'not-resolved' };
				return { passed: true };
			} catch (err) {
				return { passed: false, file: null, field: 'exception: ' + String(err?.message ?? err).slice(0, 120) };
			} finally {
				sources.delete(url);
			}
		};
		const baseline = await run(generated.guarded);
		if (!baseline.passed) {
			result.error = 'the unmutated pair fails: ' + JSON.stringify(baseline);
			results.push(result);
			continue;
		}
		for (const mutant of generated.mutants) {
			if (process.env.PAIRS_MUTANTS_PROGRESS) process.stderr.write('  ' + pair.func + ' mutant ' + mutant.id + '/' + generated.total + '\n');
			const outcome = await run(mutant.source);
			if (outcome.passed) {
				result.survivors.push(mutant.id + ' ' + mutant.kind + ' ' + mutant.line + ':' + mutant.column + ' ' + mutant.before + ' -> ' + mutant.after);
			} else {
				result.killed++;
				const key = (outcome.file ?? '(load)').replace(/^conformance\//, '') + ' ' + outcome.field;
				result.kills[key] = (result.kills[key] ?? 0) + 1;
			}
		}
		results.push(result);
		process.stderr.write('progress: ' + pair.game + ' ' + pair.func + ' ' + result.killed + '/' + result.total + '\n');
	}
	process.stdout.write(JSON.stringify(results));
}

// ---------------------------------------------------------------------------

function runChild(job) {
	return new Promise((resolve, reject) => {
		const proc = spawn(process.execPath, ['--no-warnings', scriptPath, '--child'], { cwd: staticDir, stdio: ['pipe', 'pipe', 'pipe'] });
		let stdout = '';
		let stderr = '';
		proc.stdout.on('data', chunk => (stdout += chunk));
		proc.stderr.on('data', chunk => {
			stderr += chunk;
			if (process.env.PAIRS_MUTANTS_PROGRESS) process.stderr.write(chunk);
		});
		proc.on('error', reject);
		proc.on('close', code => {
			try {
				resolve(JSON.parse(stdout));
			} catch {
				reject(new Error('child exited ' + code + ': ' + stderr.slice(-2000)));
			}
		});
		proc.stdin.end(JSON.stringify(job));
	});
}

function storeHash() {
	const hash = createHash('sha256');
	for (const pair of storePairs()) hash.update(pair.file + '\0' + readFileSync(pair.file, 'utf8'));
	return hash.digest('hex');
}

async function main(argv) {
	const args = { games: [], funcs: [], jobs: Math.max(1, Math.floor(availableParallelism() / 2)), json: null };
	for (let i = 0; i < argv.length; i++) {
		const flag = argv[i];
		const value = () => argv[++i];
		if (flag === '--game') args.games.push(value());
		else if (flag === '--func') args.funcs.push(value());
		else if (flag === '--jobs') args.jobs = Number(value());
		else if (flag === '--json') args.json = value();
		else throw new Error('unknown flag ' + flag + '; see the comment at the top of the script');
	}
	if (args.json === null) throw new Error('--json FILE is required');
	const before = storeHash();
	const all = storePairs().filter(pair => (args.games.length === 0 || args.games.includes(pair.game)) && (args.funcs.length === 0 || args.funcs.includes(pair.func)));
	const skipped = all.filter(pair => pair.role === 'helper').map(pair => ({ game: pair.game, func: pair.func, reason: 'helper: the loader never resolves it, so its accepted.ts cannot change the game' }));
	const pairs = all.filter(pair => pair.role !== 'helper');
	const chunks = Array.from({ length: Math.min(args.jobs, pairs.length) }, () => []);
	pairs.forEach((pair, i) => chunks[i % chunks.length].push(pair));
	const results = (await Promise.all(chunks.map(chunk => runChild({ pairs: chunk })))).flat();
	if (storeHash() !== before) throw new Error('the store changed during the run');
	results.sort((a, b) => (a.game + a.name < b.game + b.name ? -1 : 1));
	const command = 'node scripts/pairs-mutants.mjs ' + argv.join(' ');
	writeFileSync(args.json, JSON.stringify({ command, pairs: results, skipped }, null, '\t') + '\n');
	let total = 0;
	let killed = 0;
	for (const r of results) {
		total += r.total;
		killed += r.killed;
		console.log(r.game + ' ' + r.func + ': ' + r.killed + ' of ' + r.total + ' killed' + (r.error ? ' ERROR ' + r.error : '') + (r.survivors.length ? '; survivors: ' + r.survivors.join(' | ') : ''));
	}
	console.log('\n' + results.length + ' pairs, ' + killed + ' of ' + total + ' mutants killed; ' + skipped.length + ' helper pairs skipped');
	if (results.some(r => r.error)) process.exitCode = 1;
}
