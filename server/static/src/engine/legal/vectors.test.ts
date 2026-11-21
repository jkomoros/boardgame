// Go-generated predicate vectors (testdata/vectors.json, written by
// testdata/gen from the real Go catalog). Where the conformance corpus pins
// an outcome and a template key, these pin everything Go produced:
//
//   - the full verdict of ~3900 (spec, fixture, proposer) evaluations: the
//     outcome, the template key, every binding and its type, and the exact
//     Unknown reason text;
//   - each resolved predicate's construction metadata: reads and facets,
//     cost, client evaluability, proposer dependence, emitted templates and
//     bindings, read-type contracts;
//   - the error text of every constructor rejection.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { defaultLegalConstructors, defaultLegalRegistry } from './catalog.ts';
import { goQuote } from './goformat.ts';
import { type LegalPredicate, legalPredicateSerializable, resolveLegalSpecs, legalSpecError, isLegalSpecError } from './predicate.ts';
import type { LegalSpecError } from './predicate.ts';
import { type BoundFixture, type WireFixture, type WireGame, bindWireFixture } from './testing/fixtures.ts';
import type { LegalSpec, LegalVerdict } from './types.ts';

function testdataDir(): string {
  let dir = path.dirname(fileURLToPath(import.meta.url));
  for (;;) {
    if (existsSync(path.join(dir, 'legal', 'testdata', 'conformance'))) {
      return path.join(dir, 'server', 'static', 'src', 'engine', 'legal', 'testdata');
    }
    const parent = path.dirname(dir);
    assert.notEqual(parent, dir, 'could not find the repository root above ' + import.meta.url);
    dir = parent;
  }
}

const testdata = testdataDir();

interface GoVerdict {
  outcome: string;
  message?: { template: string; bindings?: Record<string, string | number | boolean> };
  reason?: string;
  panic?: string;
}

interface GoPredicate {
  name: string;
  args: string[];
  reads: { path: string; facet: string }[];
  cost: number;
  clientEvaluable: boolean;
  usesProposer: boolean;
  serializable: boolean;
  emittedTemplates: string[];
  emittedBindings: Record<string, string[]>;
  requiredReadTypes: Record<string, string>;
  allowedReadTypes: Record<string, string[]>;
}

const exported = JSON.parse(readFileSync(path.join(testdata, 'fixtures.json'), 'utf8')) as {
  games: Record<string, WireGame>;
  fixtures: Record<string, WireFixture>;
};

const vectors = JSON.parse(readFileSync(path.join(testdata, 'vectors.json'), 'utf8')) as {
  specs: LegalSpec[];
  predicates: GoPredicate[];
  verdicts: GoVerdict[];
  constructorErrors: { spec: LegalSpec; error: string }[];
  vectors: [source: string, spec: number, fixture: string, proposer: number, verdict: number][];
};

const registry = defaultLegalRegistry();

/**
 * Go: resolveSpecViaRegistry in legal/conformance_test.go, which is how these
 * vectors were resolved - the constructor called directly with a nil chest,
 * with none of resolveLegalSpecs' boot validation (so a vector may carry a
 * deliberately malformed path and observe what Evaluate does with it).
 */
function resolveDirect(spec: LegalSpec): LegalPredicate {
  const ctor = registry.get(spec.name);
  if (ctor === undefined) {
    throw legalSpecError(`legal: unknown predicate name ${goQuote(spec.name)}`);
  }
  return ctor.construct(spec, null, resolveDirect);
}

const predicates: LegalPredicate[] = vectors.specs.map(resolveDirect);

const bound = new Map<string, BoundFixture>();
function fixture(name: string): BoundFixture {
  let b = bound.get(name);
  if (b === undefined) {
    const wire = exported.fixtures[name];
    assert.ok(wire, `fixture ${name} is not in testdata/fixtures.json`);
    b = bindWireFixture(wire, exported.games[wire.game]);
    bound.set(name, b);
  }
  return b;
}

/** A TypeScript verdict in the shape the Go exporter writes one. */
function normalize(v: LegalVerdict): GoVerdict {
  const out: GoVerdict = { outcome: v.outcome };
  if (v.message) {
    out.message = { template: v.message.template };
    if (v.message.bindings && Object.keys(v.message.bindings).length > 0) {
      out.message.bindings = v.message.bindings;
    }
  }
  if (v.reason) {
    out.reason = v.reason;
  }
  return out;
}

test('the vector file is the expected size', () => {
  assert.ok(vectors.vectors.length > 3000, `only ${vectors.vectors.length} vectors`);
  assert.equal(vectors.predicates.length, vectors.specs.length);
});

// One subtest per group keeps a failure readable without 3900 test lines.
const groups = new Map<string, typeof vectors.vectors>();
for (const vector of vectors.vectors) {
  const group = vector[0].startsWith('corpus:') ? 'corpus' : vector[0];
  groups.set(group, [...(groups.get(group) ?? []), vector]);
}

for (const [group, members] of groups) {
  test(`Go verdicts: ${group} (${members.length} vectors)`, () => {
    for (const [source, specIndex, fixtureName, proposer, verdictIndex] of members) {
      const b = fixture(fixtureName);
      const want = vectors.verdicts[verdictIndex];
      const where = `${source}: ${JSON.stringify(vectors.specs[specIndex])} on ${fixtureName} as ${proposer}`;
      let got: GoVerdict;
      try {
        got = normalize(predicates[specIndex].evaluate({ state: b.state, move: b.move, proposer, framework: b.framework }));
      } catch (err) {
        got = { outcome: '', panic: err instanceof Error ? err.message : String(err) };
      }
      assert.deepEqual(got, want, where);
    }
  });
}

test('Go verdicts cover every outcome of every catalog predicate except the unreachable ones', () => {
  const seen = new Map<string, Set<string>>();
  for (const [, specIndex, , , verdictIndex] of vectors.vectors) {
    const name = vectors.specs[specIndex].name;
    seen.set(name, (seen.get(name) ?? new Set()).add(vectors.verdicts[verdictIndex].outcome));
  }
  // stackConstraints treats every unreadable input as "nothing to check", and
  // inPhase has no failing read: neither can be unknown with a state present.
  const noUnknown = new Set(['stackConstraints', 'inPhase']);
  for (const ctor of defaultLegalConstructors()) {
    if (ctor.name === 'inProgression') {
      continue; // package moves; covered by plan.test.ts
    }
    const outcomes = seen.get(ctor.name) ?? new Set();
    const want = noUnknown.has(ctor.name) ? ['pass', 'fail'] : ['pass', 'fail', 'unknown'];
    for (const outcome of want) {
      assert.ok(outcomes.has(outcome), `no Go vector where ${ctor.name} is ${outcome}`);
    }
  }
});

test('predicate metadata matches Go for every vector spec', () => {
  for (const [i, want] of vectors.predicates.entries()) {
    const p = predicates[i];
    const got: GoPredicate = {
      name: p.name,
      args: [...p.args],
      reads: p.reads.map(read => ({ path: read.path, facet: read.facet })),
      cost: p.cost,
      clientEvaluable: p.clientEvaluable,
      usesProposer: p.usesProposer,
      serializable: legalPredicateSerializable(p),
      emittedTemplates: [...p.emittedTemplates],
      emittedBindings: Object.fromEntries(Object.entries(p.emittedBindings ?? {}).map(([k, v]) => [k, [...v]])),
      requiredReadTypes: { ...(p.requiredReadTypes ?? {}) },
      allowedReadTypes: Object.fromEntries(Object.entries(p.allowedReadTypes ?? {}).map(([k, v]) => [k, [...v]])),
    };
    assert.deepEqual(got, want, JSON.stringify(vectors.specs[i]));
  }
});

test('constructor errors match Go for every rejected spec', () => {
  assert.ok(vectors.constructorErrors.length > 50);
  for (const { spec, error } of vectors.constructorErrors) {
    assert.throws(
      () => resolveDirect(spec),
      (err: unknown) => {
        assert.ok(isLegalSpecError(err), String(err));
        assert.equal(err.message, error, JSON.stringify(spec));
        return true;
      },
    );
    // resolveLegalSpecs wraps the same text the way Go's resolveLegalSpecs does.
    const wrapped = error.startsWith('legal: unknown predicate name')
      ? `boardgame: legal spec ${goQuote(spec.name)}: unknown predicate name ${goQuote(spec.name)}`
      : `boardgame: legal spec ${goQuote(spec.name)}: constructor failed: ${error}`;
    assert.throws(() => resolveLegalSpecs([spec], registry), { name: 'LegalSpecError', message: wrapped });
  }
});
