// Go<->TS conformance test: drives the shared conformance corpus
// (legal/testdata/conformance/*.json) through the TS evaluator and asserts it
// returns the same outcome (+ Fail template) as the Go catalog. The corpus is
// the normative cross-language contract (legal/testdata/conformance/README.md);
// any divergence here is a bug on whichever side disagrees.
//
// Fixtures are the states the Go conformance suite builds, serialized by
// `EXPORT_CONFORMANCE_FIXTURES=1 go test ./legal/ -run TestExportConformanceFixtures`
// into legal/testdata/conformance/fixtures/<name>.json. Run with
// `npm run test:unit`.
//
// SCOPE (vertical slice): only the predicates the TS evaluator has implemented
// so far are asserted; corpus files for not-yet-implemented predicates are
// skipped with a visible message (so coverage growth is tracked, not silently
// dropped).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { evaluateSpec, IMPLEMENTED_PREDICATES } from './evaluator.ts';
import type { EvalContext, LegalSpec } from './evaluator.ts';

const CORPUS_DIR = join(import.meta.dirname, '../../../../legal/testdata/conformance');
const FIXTURES_DIR = join(CORPUS_DIR, 'fixtures');

interface CorpusCase {
  spec: LegalSpec;
  fixture: string;
  proposer: number;
  expect: 'pass' | 'fail' | 'unknown';
  template?: string;
}
interface CorpusFile {
  predicate: string;
  cases: CorpusCase[];
}
interface ExportedFixture {
  fixture: string;
  currentPlayerIndex: number;
  state: EvalContext['state'];
  move: Record<string, unknown> | null;
}

const loadJSON = <T,>(p: string): T => JSON.parse(readFileSync(p, 'utf8')) as T;
const fixtureCache = new Map<string, ExportedFixture>();
function loadFixture(name: string): ExportedFixture {
  let fx = fixtureCache.get(name);
  if (!fx) {
    fx = loadJSON<ExportedFixture>(join(FIXTURES_DIR, `${name}.json`));
    fixtureCache.set(name, fx);
  }
  return fx;
}

// The predicate corpus files we currently assert (the implemented subset).
const IMPLEMENTED = new Set(IMPLEMENTED_PREDICATES);

for (const predicate of IMPLEMENTED) {
  const corpusPath = join(CORPUS_DIR, `${predicate}.json`);
  if (!existsSync(corpusPath)) continue;
  const corpus = loadJSON<CorpusFile>(corpusPath);
  test(`conformance: ${predicate} (${corpus.cases.length} cases)`, () => {
    assert.ok(existsSync(FIXTURES_DIR), `missing ${FIXTURES_DIR} — run EXPORT_CONFORMANCE_FIXTURES=1 go test ./legal/ -run TestExportConformanceFixtures`);
    for (const c of corpus.cases) {
      const fx = loadFixture(c.fixture);
      const ctx: EvalContext = {
        state: fx.state,
        move: fx.move,
        currentPlayerIndex: fx.currentPlayerIndex,
        proposer: c.proposer,
      };
      const verdict = evaluateSpec(c.spec, ctx);
      assert.equal(
        verdict.outcome,
        c.expect,
        `${predicate} case ${JSON.stringify(c.spec.args)} on ${c.fixture}: expected ${c.expect}, got ${verdict.outcome}`,
      );
      if (c.expect === 'fail' && c.template) {
        assert.equal(
          verdict.template,
          c.template,
          `${predicate} case ${JSON.stringify(c.spec.args)} on ${c.fixture}: expected template ${c.template}, got ${verdict.template}`,
        );
      }
    }
  });
}
