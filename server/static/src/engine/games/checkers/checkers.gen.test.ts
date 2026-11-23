// The generated checkers (games/checkers/, `boardgame-util transpile
// ./examples/checkers`): its 19 translated functions are the thing under
// test, held to real Go two ways (the recordings and the replay code are
// testing/checkers.ts, shared with rules/checkers.test.ts):
//
//  1. Three Go fuzz-driver documents (rules/testdata/checkers/seed-000{0,1,6})
//     through testing/fuzz.ts's runFuzzDocument: every version, every viewer,
//     every recorded rejection. Checkers has no checked-in fuzz corpus under
//     conformance/fuzz/ and is not in the conformance corpus; the
//     BOARDGAME_FUZZ_DIR run (generated.testkit.test.ts's convention) replays
//     a transient Go run against this game too.
//  2. The constructed positions of testdata/checkers.vectors.json.gz: Move
//     Token Legal texts and ledgers, Crown Token and Finish Turn legality,
//     every Apply, the end condition and scores.
//
// PENDING: one function is refused by the translator and has no body in
// bodies.gen.ts, so the generated index.ts cannot load checkers on its own:
//
//   checkers.(*gameDelegate).ConfigurePredicateConstructors
//     predicate-constructor checkers.spaceIsBlack
//
// It is bound for the agent loop. Until its body exists, this test supplies
// the TEST-SIDE STAND-IN testing/checkers.ts spaceIsBlackConstructor (the
// stand-in rules/checkers.test.ts has always used), wired in front of the
// generated `bodies`. Everything else under test is generated.
import assert from 'node:assert/strict';
import { existsSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { compileGame, isGameSpecError } from '../../rules/compile.ts';
import { gameFromData } from '../../rules/data.ts';
import type { CompiledGame, GameSpec } from '../../rules/index.ts';
import { CHECKERS_DOCUMENTS, CHECKERS_PREDICATE_REF, CHECKERS_VECTORS, type PositionCounts, type Vectors, checkPositions, emptyCounts, readGzipJSON, spaceIsBlackConstructor } from '../../testing/checkers.ts';
import { type FuzzDocument, type FuzzResult, fuzzPassTable, runFuzzDocument } from '../../testing/fuzz.ts';
import { fuzzProblems, withPendingBodies } from '../generated.testkit.test.ts';
import { bodies } from './bodies.gen.ts';
import { definition } from './definition.gen.ts';
import { checkersGenerated } from './index.ts';

/** The FnRefs the translator refused, which bodies.gen.ts therefore lacks. */
/** None now: the predicate constructor was translated by a coding agent from its brief and accepted (see bodies.gen.ts, `@generated agent:`). */
const PENDING: string[] = [];

const checkersWithStandIn = (): GameSpec => checkersGenerated();

let compiled: CompiledGame | undefined;
const game = (): CompiledGame => {
	compiled ??= compileGame(checkersWithStandIn());
	return compiled;
};

test('checkers: the pending FnRefs are exactly the refused functions, and index.ts names them when it cannot load', () => {
	assert.notEqual(bodies(CHECKERS_PREDICATE_REF), undefined, 'the agent-written predicate constructor has no body');
	assert.equal(checkersGenerated().name, 'checkers');
	// Every other function slot the definition names has a generated body.
	const data = definition();
	const named = new Set<string>();
	for (const move of data.Moves) for (const ref of Object.values(move.Bodies)) if (ref) named.add(ref);
	for (const ref of [data.SetUp.Begin, data.SetUp.Distribute, data.SetUp.Finish, data.End.ConditionMet, data.End.PlayerScore, data.End.Check, data.Traits.Player.GameScore]) if (ref) named.add(ref);
	for (const ref of Object.values(data.Traits.Player.Turn ?? {})) if (ref) named.add(ref);
	for (const ref of Object.values(data.ComponentLegal ?? {})) named.add(ref);
	for (const p of data.Predicates) named.add(p.Construct);
	const missing = [...named].filter(ref => bodies(ref) === undefined).sort();
	assert.deepEqual(missing, PENDING);
});

test('the generated checkers compiles with the stand-in predicate', () => {
	assert.equal(game().moves.length, 7);
});

const fuzzResults: FuzzResult[] = [];
const replay = (label: string, file: string, into: FuzzResult[]): void => {
	test(label + ': ' + path.basename(file, '.json.gz'), () => {
		const doc = readGzipJSON<FuzzDocument>(file);
		const result = runFuzzDocument(game(), doc, { spec: checkersWithStandIn });
		result.name = 'checkers/' + path.basename(file, '.json.gz');
		into.push(result);
		const problems = fuzzProblems(result, doc);
		assert.equal(problems.length, 0, '\n  ' + problems.join('\n  '));
	});
};
for (const file of CHECKERS_DOCUMENTS) replay('generated checkers, Go fuzz document', file, fuzzResults);

test('generated checkers: the three Go documents', t => {
	t.diagnostic('\n' + fuzzPassTable(fuzzResults));
	assert.equal(fuzzResults.length, 3);
	assert.equal(fuzzResults.reduce((n, r) => n + r.rejections.compared, 0), 1, 'seed 6 records one refusal');
});

// The constructed positions, in slices so each test stays far under the
// suite's per-test timeout.
const vectorsPresent = existsSync(CHECKERS_VECTORS);
const SLICES = 4;
const totals: PositionCounts = emptyCounts();
let vectors: Vectors | undefined;
const loadVectors = (): Vectors => {
	vectors ??= readGzipJSON<Vectors>(CHECKERS_VECTORS);
	return vectors;
};
for (let slice = 0; slice < SLICES; slice++) {
	test('generated checkers, constructed positions, slice ' + (slice + 1) + ' of ' + SLICES, { skip: vectorsPresent ? false : CHECKERS_VECTORS + ' is absent' }, () => {
		const v = loadVectors();
		const positions = v.positions.filter((_, i) => i % SLICES === slice);
		const failures = checkPositions(game(), v, positions, totals);
		assert.deepEqual(failures, []);
	});
}
test('generated checkers, constructed positions: totals', { skip: vectorsPresent ? false : CHECKERS_VECTORS + ' is absent' }, t => {
	t.diagnostic('checkers vectors: ' + JSON.stringify(totals));
	assert.equal(totals.positions, loadVectors().positions.length);
	assert.ok(totals.legalMoves > 0, 'no legal Move Token in the vectors');
	assert.ok(totals.appliedOk > 0, 'no Apply succeeded');
	assert.ok(totals.captures > 0, 'no capture');
});

// A transient Go run (`boardgame-util fuzz run --game checkers --seeds A..B
// --out DIR`), replayed against the generated game.
const transient = process.env.BOARDGAME_FUZZ_DIR;
const transientDir = transient !== undefined && transient !== '' ? path.join(transient, 'checkers') : '';
if (transientDir !== '' && existsSync(transientDir)) {
	const files = readdirSync(transientDir).filter(f => /^seed-\d+\.json\.gz$/.test(f)).sort().map(f => path.join(transientDir, f));
	const transientResults: FuzzResult[] = [];
	for (const file of files) replay('generated checkers fuzz (BOARDGAME_FUZZ_DIR)', file, transientResults);
	test('generated checkers BOARDGAME_FUZZ_DIR pass table', t => {
		t.diagnostic('\n' + fuzzPassTable(transientResults));
		const sum = { versions: 0, viewerVersions: 0, rejections: 0 };
		for (const r of transientResults) {
			sum.versions += r.document.versions;
			sum.viewerVersions += r.document.versions * r.document.viewers;
			sum.rejections += r.rejections.compared;
		}
		t.diagnostic(`BOARDGAME_FUZZ_DIR checkers: documents ${transientResults.length}, versions ${sum.versions}, viewer-versions ${sum.viewerVersions}, rejections ${sum.rejections}, passed ${transientResults.filter(r => r.passed).length}`);
		assert.equal(transientResults.length, files.length);
	});
}
