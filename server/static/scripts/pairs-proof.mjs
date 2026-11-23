#!/usr/bin/env node
/**
 * The TypeScript half of a rule proof (fabric/TRANSPILE-PROPOSAL.md,
 * "Turning a refusal reason into a rule"), driven by
 * boardgame-util/lib/transpile/pairs (TestRuleAgainstPairs and the general
 * substitution proof). Reads a job on stdin:
 *
 *   { "baseline": ["pig", ...],
 *     "substitutions": [{ "game": "pig", "func": "pig.(*moveRollDice).Apply",
 *                         "file": "/abs/translated.ts", "name": "pig_moveRollDice_Apply" }, ...] }
 *
 * For each baseline game, the game built from its definition and the stored
 * pairs (games/pairs.testkit.test.ts) is replayed against its whole corpus:
 * conformance documents and checked-in fuzz documents, rejections included.
 * For each substitution, the same game with that one FnRef served by the
 * function `name` of `file` instead of the pair's accepted.ts, replayed until
 * the first document that fails (bounded by the document count). The first
 * failing document is replayed again with instrumented bodies, for the
 * divergence report's `bodiesExecuted`.
 *
 * Writes { baseline: [...], substitutions: [...] } as JSON on stdout. Writes
 * no file.
 */
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const scriptPath = fileURLToPath(import.meta.url);
const engineDir = path.resolve(path.dirname(scriptPath), '..', 'src', 'engine');
const kit = await import(pathToFileURL(path.join(engineDir, 'games', 'pairs.testkit.test.ts')).href);

async function replay(game, substitutions, stopAtFirstFailure) {
	const out = { documents: 0, versions: 0, rejections: 0, passed: true, failure: null, served: [], substituted: [], fallbacks: [], stockStandIns: [] };
	let built;
	try {
		built = await kit.storeGame(game, substitutions);
	} catch (err) {
		out.passed = false;
		out.failure = { file: null, problems: ['the game did not build: ' + String(err?.message ?? err)] };
		return out;
	}
	for (const entry of kit.corpusDocuments(game)) {
		const doc = kit.readDocument(entry.file);
		let check;
		try {
			check = kit.checkDocument(built.game, entry.file, doc);
		} catch (err) {
			check = { passed: false, problems: ['threw: ' + String(err?.message ?? err)], versions: doc.versions.length, rejections: 0 };
		}
		out.documents++;
		out.versions += check.versions;
		out.rejections += check.rejections;
		if (!check.passed) {
			out.passed = false;
			let problems = check.problems;
			try {
				problems = kit.checkDocument(built.game, entry.file, doc, built.spec).problems;
			} catch {
				// keep the first problems
			}
			if (out.failure === null) out.failure = { file: entry.file, problems: problems.slice(0, 6).map(p => p.slice(0, 2000)) };
			if (stopAtFirstFailure) break;
		}
	}
	out.served = built.table.served;
	out.substituted = built.table.substituted;
	out.fallbacks = built.table.fallbacks;
	out.stockStandIns = built.table.stockStandIns;
	return out;
}

const job = JSON.parse(readFileSync(0, 'utf8'));
const result = { baseline: [], substitutions: [] };
for (const game of job.baseline ?? []) result.baseline.push({ game, ...(await replay(game, new Map(), false)) });
for (const sub of job.substitutions ?? []) {
	let fn;
	try {
		fn = (await import(pathToFileURL(sub.file).href))[sub.name];
	} catch (err) {
		result.substitutions.push({ ...sub, passed: false, failure: { file: null, problems: ['the translation did not load: ' + String(err?.message ?? err)] } });
		continue;
	}
	if (typeof fn !== 'function') {
		result.substitutions.push({ ...sub, passed: false, failure: { file: null, problems: [sub.file + ' exports no function ' + sub.name] } });
		continue;
	}
	const run = await replay(sub.game, new Map([[sub.func, fn]]), true);
	if (!run.substituted.includes(sub.func)) {
		run.passed = false;
		run.failure = run.failure ?? { file: null, problems: ['the loader never asked for ' + sub.func + ', so the substitution was not exercised'] };
	}
	result.substitutions.push({ ...sub, ...run });
}
process.stdout.write(JSON.stringify(result));
