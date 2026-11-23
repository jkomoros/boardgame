#!/usr/bin/env node
/**
 * The mutation harness for the fuzz corpus (fabric/TRANSPILE-PROPOSAL.md,
 * "Mutations"). Two modes:
 *
 *   node scripts/fuzz-mutations.mjs historical [--only id,id] [--docs PATH ...] [--jobs N] [--json FILE]
 *
 *     Runs each of the 20 historical mutations of fabric/status/engine-rules.md
 *     (src/engine/testing/mutate.ts HISTORICAL_MUTATIONS) against the fuzz
 *     corpus and reports, per mutation, how many documents fail and on which
 *     field. A mutation the corpus does not catch is a SURVIVOR.
 *
 *   node scripts/fuzz-mutations.mjs body --file games/pig/bodies.gen.ts --function pig_moveRollDice_Apply [--docs PATH ...] [--max N] [--jobs N] [--json FILE]
 *
 *     The routine generator: every comparison of the function flipped, every
 *     integer literal shifted by one, every statement deleted, one at a time
 *     (the first --max of them, in source order), each run against the named
 *     documents until one fails. Survivors are listed.
 *
 * --docs takes directories (every seed-*.json.gz below them) or files;
 * default: the checked-in conformance/fuzz. Paths are relative to the
 * repository root. --jobs bounds how many child processes run at once.
 *
 * No mutated file is written. Each mutant runs in its own Node process whose
 * module loader substitutes the mutated text for the one file while it
 * loads (node:module registerHooks). Every file under src/engine/ and
 * conformance/ is hashed before and after the run, and the run fails if any
 * changed. Nothing is bounded by time: runs are bounded by the mutant count,
 * and a mutant that makes a loop of the target function run away throws
 * after a fixed number of iterations.
 */
import { spawn } from 'node:child_process';
import { createHash } from 'node:crypto';
import { existsSync, readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs';
import { availableParallelism } from 'node:os';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { gunzipSync } from 'node:zlib';

const scriptPath = fileURLToPath(import.meta.url);
const staticDir = path.resolve(path.dirname(scriptPath), '..');
const engineDir = path.join(staticDir, 'src', 'engine');
const repoRoot = path.resolve(staticDir, '..', '..');

if (process.argv[2] === '--child') {
	await child();
} else {
	await main(process.argv.slice(2));
}

// ---------------------------------------------------------------------------
// The child: one mutant, a set of documents.

async function child() {
	const job = JSON.parse(readFileSync(0, 'utf8'));
	const { registerHooks } = await import('node:module');
	if (job.file !== null) {
		const target = pathToFileURL(path.join(engineDir, job.file)).href;
		registerHooks({
			load(url, context, nextLoad) {
				const result = nextLoad(url, context);
				if (url.split('?')[0] !== target) return result;
				return { ...result, source: job.source, shortCircuit: true };
			},
		});
	}
	globalThis.__fuzzMutantLoopSteps = 0;
	const out = { documents: [], error: null };
	try {
		const { compileGame } = await import(pathToFileURL(path.join(engineDir, 'rules', 'index.ts')).href);
		const { runFuzzDocument } = await import(pathToFileURL(path.join(engineDir, 'testing', 'fuzz.ts')).href);
		const compiled = new Map();
		for (const file of job.docs) {
			let doc;
			try {
				doc = JSON.parse(gunzipSync(readFileSync(file)).toString('utf8'));
				let game = compiled.get(doc.game.name);
				if (game === undefined) {
					const spec = await gameSpec(doc.game.name);
					game = compileGame(spec());
					compiled.set(doc.game.name, game);
				}
				const result = runFuzzDocument(game, doc);
				const report = result.report;
				out.documents.push({
					name: doc.name, passed: result.passed,
					field: report?.field, version: report?.version, viewer: report?.viewer, path: report?.path,
					fatal: result.document.fatal,
				});
			} catch (err) {
				out.documents.push({ name: doc?.name ?? file, passed: false, field: 'exception', exception: String(err?.message ?? err).slice(0, 300) });
			}
			if (job.stopAtFirstFailure && !out.documents[out.documents.length - 1].passed) break;
		}
	} catch (err) {
		out.error = String(err?.stack ?? err).slice(0, 1000);
	}
	process.stdout.write(JSON.stringify(out));
}

async function gameSpec(name) {
	// The seven generated games (games/<game>/index.ts exports <game>()), which
	// replaced the six hand ports; checkers never had one.
	const games = ['memory', 'blackjack', 'tictactoe', 'pig', 'debuganimations', 'werewolf', 'checkers'];
	if (!games.includes(name)) throw new Error('no game registered for ' + name);
	const module = await import(pathToFileURL(path.join(engineDir, 'games', name, 'index.ts')).href);
	if (typeof module[name] !== 'function') throw new Error('games/' + name + '/index.ts exports no ' + name);
	return module[name];
}

// ---------------------------------------------------------------------------
// The parent

function parseArgs(argv) {
	const args = { mode: argv[0], docs: [], jobs: Math.max(1, Math.floor(availableParallelism() / 2)), only: null, json: null, file: null, function: null, max: Infinity };
	for (let i = 1; i < argv.length; i++) {
		const flag = argv[i];
		const value = () => {
			if (i + 1 >= argv.length) throw new Error(flag + ' needs a value');
			return argv[++i];
		};
		if (flag === '--docs') {
			while (i + 1 < argv.length && !argv[i + 1].startsWith('--')) args.docs.push(argv[++i]);
		} else if (flag === '--jobs') args.jobs = Number(value());
		else if (flag === '--only') args.only = value().split(',');
		else if (flag === '--json') args.json = value();
		else if (flag === '--file') args.file = value();
		else if (flag === '--function') args.function = value();
		else if (flag === '--max') args.max = Number(value());
		else throw new Error('unknown flag ' + flag);
	}
	if (args.mode !== 'historical' && args.mode !== 'body') throw new Error('usage: fuzz-mutations.mjs historical|body [flags]; see the comment at the top of the script');
	if (!(args.jobs >= 1)) throw new Error('--jobs must be at least 1');
	if (args.mode === 'body' && (args.file === null || args.function === null)) throw new Error('body needs --file and --function');
	if (args.docs.length === 0) args.docs.push('conformance/fuzz');
	return args;
}

function documentsUnder(entries) {
	const out = [];
	const walk = full => {
		if (statSync(full).isDirectory()) {
			for (const name of readdirSync(full).sort()) walk(path.join(full, name));
		} else if (/seed-\d+\.json\.gz$/.test(full)) out.push(full);
	};
	for (const entry of entries) {
		const full = path.resolve(repoRoot, entry);
		if (!existsSync(full)) throw new Error('no such path: ' + entry);
		walk(full);
	}
	if (out.length === 0) throw new Error('no documents under ' + entries.join(', '));
	return out;
}

function treeHash() {
	const hash = createHash('sha256');
	const walk = dir => {
		for (const entry of readdirSync(dir, { withFileTypes: true }).sort((a, b) => (a.name < b.name ? -1 : 1))) {
			const full = path.join(dir, entry.name);
			if (entry.isDirectory()) walk(full);
			else if (entry.isFile()) {
				hash.update(path.relative(repoRoot, full) + '\0');
				hash.update(readFileSync(full));
			}
		}
	};
	walk(engineDir);
	walk(path.join(repoRoot, 'conformance'));
	return hash.digest('hex');
}

function runChild(job) {
	return new Promise((resolve, reject) => {
		const proc = spawn(process.execPath, [scriptPath, '--child'], { cwd: staticDir, stdio: ['pipe', 'pipe', 'pipe'] });
		let stdout = '';
		let stderr = '';
		proc.stdout.on('data', chunk => (stdout += chunk));
		proc.stderr.on('data', chunk => (stderr += chunk));
		proc.on('error', reject);
		proc.on('close', code => {
			try {
				resolve({ ...JSON.parse(stdout), code });
			} catch {
				resolve({ documents: [], error: 'child exited ' + code + ' without a result: ' + stderr.slice(-1000), code });
			}
		});
		proc.stdin.end(JSON.stringify(job));
	});
}

async function pool(items, jobs, work) {
	const results = new Array(items.length);
	let next = 0;
	const worker = async () => {
		while (next < items.length) {
			const i = next++;
			results[i] = await work(items[i], i);
		}
	};
	await Promise.all(Array.from({ length: Math.min(jobs, items.length) }, worker));
	return results;
}

function summarize(result) {
	const failed = result.documents.filter(d => !d.passed);
	const fields = {};
	for (const d of failed) fields[d.field] = (fields[d.field] ?? 0) + 1;
	return { failed: failed.length, compared: result.documents.length, fields, first: failed[0] ?? null };
}

function fieldText(fields) {
	return Object.entries(fields).sort((a, b) => b[1] - a[1]).map(([field, n]) => field + ' ' + n).join(', ');
}

async function main(argv) {
	const args = parseArgs(argv);
	const docs = documentsUnder(args.docs);
	const { HISTORICAL_MUTATIONS, applyMutation, bodyMutants } = await import(pathToFileURL(path.join(engineDir, 'testing', 'mutate.ts')).href);
	const before = treeHash();

	// The unmutated baseline must pass, or nothing below means anything.
	const baseline = summarize(await runChild({ file: null, source: null, docs, stopAtFirstFailure: false }));
	console.log('baseline: ' + (baseline.compared - baseline.failed) + ' of ' + baseline.compared + ' documents pass');
	if (baseline.failed > 0) throw new Error('the unmutated engine fails ' + baseline.failed + ' documents: ' + JSON.stringify(baseline.first));

	let output;
	if (args.mode === 'historical') {
		const selected = HISTORICAL_MUTATIONS.filter(m => args.only === null || args.only.includes(m.id));
		const results = await pool(selected, args.jobs, async mutation => {
			const original = readFileSync(path.join(engineDir, mutation.file), 'utf8');
			let source;
			try {
				source = applyMutation(original, mutation);
			} catch (err) {
				return { id: mutation.id, applied: false, error: String(err.message) };
			}
			const run = await runChild({ file: mutation.file, source, docs, stopAtFirstFailure: false });
			return { id: mutation.id, applied: true, error: run.error, ...summarize(run) };
		});
		const rows = selected.map((mutation, i) => ({ ...mutation, ...results[i] }));
		console.log('\n| Mutation | File | Failed on (conformance corpus) | Fuzz documents failing | First field per failing document | Result |');
		console.log('|---|---|---|---|---|---|');
		for (const row of rows) {
			const killed = row.applied && (row.failed > 0 || row.error);
			const result = !row.applied ? 'NOT APPLIED: ' + row.error : row.error ? 'killed (load error)' : killed ? 'killed' : 'SURVIVED';
			console.log('| ' + row.description + ' | ' + row.file + ' | ' + row.failedOn + ' | ' + (row.applied ? row.failed + ' of ' + row.compared : '-') + ' | ' + (row.applied ? fieldText(row.fields ?? {}) || (row.error ? 'module failed to load' : '') : '') + ' | ' + result + ' |');
		}
		const survivors = rows.filter(row => row.applied && !row.error && row.failed === 0);
		const unapplied = rows.filter(row => !row.applied);
		console.log('\n' + (rows.length - survivors.length - unapplied.length) + ' of ' + rows.length + ' killed; survivors: ' + (survivors.map(r => r.id).join(', ') || 'none') + (unapplied.length ? '; not applied: ' + unapplied.map(r => r.id).join(', ') : ''));
		output = { mode: 'historical', documents: docs.length, rows: rows.map(({ find, replace, ...rest }) => rest) };
	} else {
		const relative = args.file;
		const original = readFileSync(path.join(engineDir, relative), 'utf8');
		const ts = (await import('typescript')).default;
		const generated = bodyMutants(ts, relative, original, args.function);
		const mutants = generated.mutants.slice(0, args.max);
		// The guarded file alone must still pass: the guards change nothing.
		const guarded = summarize(await runChild({ file: relative, source: generated.guarded, docs, stopAtFirstFailure: false }));
		if (guarded.failed > 0) throw new Error('the function with only its loop guards fails: ' + JSON.stringify(guarded.first));
		console.log(relative + ' ' + args.function + ' (lines ' + generated.startLine + '-' + generated.endLine + '): ' + generated.total + ' mutants, running ' + mutants.length + '; ' + generated.skippedDeclarations + ' declarations not deleted');
		const results = await pool(mutants, args.jobs, mutant => runChild({ file: relative, source: mutant.source, docs, stopAtFirstFailure: true }));
		const rows = mutants.map((mutant, i) => {
			const run = results[i];
			const failed = run.documents.find(d => !d.passed);
			const killed = run.error !== null || failed !== undefined;
			return { id: mutant.id, kind: mutant.kind, line: mutant.line, column: mutant.column, before: mutant.before, after: mutant.after, killed, by: run.error !== null ? 'load error' : failed ? failed.name + ' ' + failed.field : null };
		});
		const survivors = rows.filter(row => !row.killed);
		const byKind = {};
		for (const row of rows) {
			byKind[row.kind] ??= { total: 0, survived: 0 };
			byKind[row.kind].total++;
			if (!row.killed) byKind[row.kind].survived++;
		}
		console.log('\n| Kind | Mutants | Survived |\n|---|---|---|');
		for (const [kind, counts] of Object.entries(byKind)) console.log('| ' + kind + ' | ' + counts.total + ' | ' + counts.survived + ' |');
		console.log('\n' + (rows.length - survivors.length) + ' of ' + rows.length + ' killed against ' + docs.length + ' documents.');
		if (survivors.length > 0) {
			console.log('\nSurvivors:\n\n| # | Kind | At | Before | After |\n|---|---|---|---|---|');
			for (const row of survivors) console.log('| ' + row.id + ' | ' + row.kind + ' | ' + relative + ':' + row.line + ':' + row.column + ' | `' + row.before.replace(/\|/g, '\\|') + '` | `' + row.after + '` |');
		}
		output = { mode: 'body', file: relative, function: args.function, documents: docs.length, total: generated.total, ran: rows.length, skippedDeclarations: generated.skippedDeclarations, rows };
	}

	const after = treeHash();
	if (after !== before) throw new Error('a file under src/engine/ or conformance/ changed during the run');
	console.log('\nNo file under src/engine/ or conformance/ changed (tree hash ' + after.slice(0, 12) + ').');
	if (args.json !== null) writeFileSync(path.resolve(args.json), JSON.stringify(output, null, 2) + '\n');
}
