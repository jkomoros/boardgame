// Client evaluability against real Go (testdata/evaluable.json, written by
// testdata/gen from boardgame.LegalReadEvaluable and the Visibility block of
// the state JSON the server sends each viewer), plus the per-viewer ledger's
// binding guard.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
  type LegalPolicy,
  type LegalVisibility,
  buildPreconditionEntry,
  facetSurvives,
  legalEntryEvaluable,
  legalReadEvaluable,
  legalReadsEvaluable,
  legalVisibilityFromPolicies,
  visibleFacets,
} from './evaluable.ts';
import type { LegalVerdictEntry } from './plan.ts';
import { LEGAL_FACETS, type LegalFacet } from './types.ts';

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

const vectors = JSON.parse(readFileSync(path.join(testdataDir(), 'evaluable.json'), 'utf8')) as {
  facetTable: { policy: LegalPolicy; survives: Record<LegalFacet, boolean>; visibility: LegalFacet[] }[];
  readCases: {
    config: string;
    state: string;
    viewer: number;
    currentPlayerIndex: number;
    visibility: LegalVisibility;
    reads: [path: string, facet: LegalFacet, evaluable: boolean][];
  }[];
};

test('facetSurvives matches Go for every policy and facet', () => {
  assert.deepEqual(vectors.facetTable.map(row => row.policy), ['visible', 'order', 'len', 'nonempty', 'hidden']);
  for (const row of vectors.facetTable) {
    for (const facet of LEGAL_FACETS) {
      assert.equal(facetSurvives(row.policy, facet), row.survives[facet], `${row.policy} / ${facet}`);
    }
    // And the facet list the server ships for a property under that policy.
    assert.deepEqual(visibleFacets(row.policy), row.visibility, row.policy);
  }
});

test('legalReadEvaluable matches Go for every path kind, viewer and policy mix', () => {
  let checked = 0;
  const seen = { true: 0, false: 0 };
  for (const c of vectors.readCases) {
    const view = { viewer: c.viewer, visibility: c.visibility, currentPlayerIndex: c.currentPlayerIndex };
    for (const [readPath, facet, want] of c.reads) {
      assert.equal(
        legalReadEvaluable(view, { path: readPath, facet }),
        want,
        `${c.config}/${c.state} viewer ${c.viewer}: ${readPath} (${facet})`,
      );
      seen[String(want) as 'true' | 'false']++;
      checked++;
    }
  }
  assert.ok(checked > 2000, `only ${checked} reads checked`);
  assert.ok(seen.true > 300 && seen.false > 300, JSON.stringify(seen));
});

test('legalVisibilityFromPolicies builds the block the server ships', () => {
  assert.deepEqual(
    legalVisibilityFromPolicies(['Deck', 'Round'], { Deck: 'len' }, [{ props: ['Hand', 'Score'], policies: { Hand: 'hidden', Score: 'visible' } }]),
    {
      Game: { Deck: ['count', 'nonempty'], Round: ['values', 'count', 'occupancy', 'order', 'nonempty'] },
      Players: [{ Hand: [], Score: ['values', 'count', 'occupancy', 'order', 'nonempty'] }],
    },
  );
});

const visibility: LegalVisibility = {
  Game: { Deck: ['count', 'nonempty'], Discard: ['values', 'count', 'occupancy', 'order', 'nonempty'] },
  Players: [
    { Hand: ['values', 'count', 'occupancy', 'order', 'nonempty'], Score: ['values', 'count', 'occupancy', 'order', 'nonempty'] },
    { Hand: ['count', 'nonempty'], Score: ['values', 'count', 'occupancy', 'order', 'nonempty'] },
  ],
};

test('a read is judged by the facet it needs, not by the property', () => {
  const view = { viewer: 0, visibility, currentPlayerIndex: 0 };
  assert.equal(legalReadEvaluable(view, { path: 'game.Deck', facet: 'count' }), true);
  assert.equal(legalReadEvaluable(view, { path: 'game.Deck', facet: 'nonempty' }), true);
  assert.equal(legalReadEvaluable(view, { path: 'game.Deck', facet: 'occupancy' }), false);
  assert.equal(legalReadEvaluable(view, { path: 'game.Deck', facet: 'values' }), false);
});

test('player reads follow the current player; proposer and quantified reads need every player', () => {
  const asZero = { viewer: 0, visibility, currentPlayerIndex: 0 };
  const whenOnesTurn = { viewer: 0, visibility, currentPlayerIndex: 1 };
  assert.equal(legalReadEvaluable(asZero, { path: 'player.Hand', facet: 'values' }), true);
  assert.equal(legalReadEvaluable(whenOnesTurn, { path: 'player.Hand', facet: 'values' }), false);
  for (const p of ['proposer.Hand', 'players[*].Hand', 'players[move.Target].Hand']) {
    assert.equal(legalReadEvaluable(asZero, { path: p, facet: 'values' }), false, p);
    assert.equal(legalReadEvaluable(asZero, { path: p, facet: 'count' }), true, p);
  }
  // No concrete current player: nothing player-scoped is evaluable.
  assert.equal(legalReadEvaluable({ viewer: 0, visibility, currentPlayerIndex: -1 }, { path: 'player.Score', facet: 'values' }), false);
});

test('move reads and an admin viewer are always evaluable; everything else fails closed', () => {
  const blind = { viewer: 1, visibility: null, currentPlayerIndex: 0 };
  assert.equal(legalReadEvaluable(blind, { path: 'move.Anything', facet: 'values' }), true);
  assert.equal(legalReadEvaluable(blind, { path: 'game.Deck', facet: 'nonempty' }), false);
  assert.equal(legalReadEvaluable({ ...blind, viewer: -2 }, { path: 'game.Deck', facet: 'values' }), true);
  // A malformed path is never evaluable, even for admin.
  assert.equal(legalReadEvaluable({ ...blind, viewer: -2 }, { path: 'nonsense', facet: 'values' }), false);
  assert.equal(legalReadsEvaluable(blind, []), true);
});

function entry(overrides: Partial<LegalVerdictEntry>): LegalVerdictEntry {
  return {
    name: 'stackCount',
    args: ['game.Deck', '>=', '5'],
    verdict: { outcome: 'fail', message: { template: 'legal.stack_count', bindings: { count: 3, op: '>=', n: 5 } } },
    serializable: true,
    clientEvaluable: true,
    fieldDependent: false,
    reads: [{ path: 'game.Deck', facet: 'count' }],
    ...overrides,
  };
}

test('a ledger entry ships its bindings only to a viewer who could have computed it', () => {
  const view = { viewer: 1, visibility, currentPlayerIndex: 0 };
  assert.deepEqual(buildPreconditionEntry(view, entry({})), {
    name: 'stackCount',
    args: ['game.Deck', '>=', '5'],
    verdict: 'fail',
    evaluable: true,
    message: { template: 'legal.stack_count', bindings: { count: 3, op: '>=', n: 5 } },
  });
  // Same failure, but it read the deck's values, which this viewer cannot see:
  // the template key still ships, the state-derived bindings do not.
  const hidden = entry({ reads: [{ path: 'game.Deck', facet: 'values' }], fieldDependent: true });
  assert.deepEqual(buildPreconditionEntry(view, hidden), {
    name: 'stackCount',
    args: ['game.Deck', '>=', '5'],
    verdict: 'fail',
    evaluable: false,
    provisional: true,
    message: { template: 'legal.stack_count' },
  });
});

test('evaluable needs all three: a known predicate, a wire form, and visible reads', () => {
  const view = { viewer: 0, visibility, currentPlayerIndex: 0 };
  assert.equal(legalEntryEvaluable(view, entry({})), true);
  assert.equal(legalEntryEvaluable(view, entry({ clientEvaluable: false })), false);
  assert.equal(legalEntryEvaluable(view, entry({ serializable: false })), false);
  assert.equal(legalEntryEvaluable(view, entry({ reads: [{ path: 'game.Deck', facet: 'order' }] })), false);
  // The custom wrapper: no reads at all, but no wire form either.
  const custom = entry({ name: 'custom', args: [], reads: [], serializable: false, clientEvaluable: false, verdict: { outcome: 'pass' } });
  assert.deepEqual(buildPreconditionEntry(view, custom), { name: 'custom', verdict: 'pass', evaluable: false });
});
