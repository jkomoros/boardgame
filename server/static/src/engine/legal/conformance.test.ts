// The conformance harness: every file under legal/testdata/conformance/ runs
// against the TypeScript evaluator, read in place (never copied).
//
// The corpus names its fixtures; the fixtures themselves are built in Go
// (legal/conformance_test.go's legalFixtureBuilders). testdata/fixtures.json
// is those same fixtures exported by the Go generator in testdata/gen, which
// refuses to write anything unless the real Go catalog agrees with every
// corpus case against them.
//
// A case passes when the outcome matches `expect` and, where the case pins
// one, the failure's template key matches `template` - the same two checks
// Go's TestConformanceCorpus makes.
import assert from 'node:assert/strict';
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { defaultLegalConstructors, defaultLegalRegistry } from './catalog.ts';
import { evalLegalPredicate, resolveLegalSpecs } from './predicate.ts';
import { type WireFixture, type WireGame, bindWireFixture } from './testing/fixtures.ts';
import type { LegalOutcome, LegalSpec } from './types.ts';

/** The repository root: the nearest ancestor holding the conformance corpus. */
function repoRoot(): string {
  let dir = path.dirname(fileURLToPath(import.meta.url));
  for (;;) {
    if (existsSync(path.join(dir, 'legal', 'testdata', 'conformance'))) {
      return dir;
    }
    const parent = path.dirname(dir);
    assert.notEqual(parent, dir, 'could not find legal/testdata/conformance above ' + import.meta.url);
    dir = parent;
  }
}

const root = repoRoot();
const corpusDir = path.join(root, 'legal', 'testdata', 'conformance');
const testdata = path.join(root, 'server', 'static', 'src', 'engine', 'legal', 'testdata');

interface ConformanceCase {
  spec: LegalSpec;
  fixture: string;
  proposer: number;
  expect: LegalOutcome;
  template?: string;
}

interface ConformanceFile {
  predicate: string;
  cases: ConformanceCase[];
}

const exported = JSON.parse(readFileSync(path.join(testdata, 'fixtures.json'), 'utf8')) as {
  games: Record<string, WireGame>;
  fixtures: Record<string, WireFixture>;
};

const corpusFiles = readdirSync(corpusDir)
  .filter(name => name.endsWith('.json'))
  .sort();

const registry = defaultLegalRegistry();

test('the corpus is present', () => {
  assert.ok(corpusFiles.length > 0, 'no conformance corpus files found under ' + corpusDir);
});

for (const fileName of corpusFiles) {
  const file = JSON.parse(readFileSync(path.join(corpusDir, fileName), 'utf8')) as ConformanceFile;

  test(`conformance: ${fileName}`, async t => {
    assert.ok(file.cases.length >= 3, `${fileName} has ${file.cases.length} cases, want >= 3`);

    for (const [i, c] of file.cases.entries()) {
      await t.test(`case ${i}: ${c.fixture} as proposer ${c.proposer} -> ${c.expect}`, () => {
        assert.equal(c.spec.name, file.predicate, 'case spec name does not match the file predicate');

        const fixture = exported.fixtures[c.fixture];
        assert.ok(fixture, `fixture ${c.fixture} is not in testdata/fixtures.json; regenerate it`);
        const bound = bindWireFixture(fixture, exported.games[fixture.game]);

        // Resolved the way the Go harness resolves: default catalog, no chest.
        const [predicate] = resolveLegalSpecs([c.spec], registry);
        const verdict = evalLegalPredicate(predicate, {
          state: bound.state,
          move: bound.move,
          proposer: c.proposer,
          framework: bound.framework,
        });

        assert.equal(verdict.outcome, c.expect, `verdict was ${JSON.stringify(verdict)}`);
        if (c.template) {
          assert.equal(verdict.message?.template, c.template, `verdict was ${JSON.stringify(verdict)}`);
        }
      });
    }
  });
}

test('every catalog predicate has a corpus file, and every corpus file names a catalog predicate', () => {
  // Go: legal/conformance_completeness_test.go. "inProgression" lives in
  // package moves and has no corpus file in Go either.
  const catalog = defaultLegalConstructors()
    .map(ctor => ctor.name)
    .filter(name => name !== 'inProgression')
    .sort();
  const covered = corpusFiles.map(name => name.replace(/\.json$/, '')).sort();
  assert.deepEqual(covered, catalog);
});
