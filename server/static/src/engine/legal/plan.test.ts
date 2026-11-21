// The legal plan against real Go (testdata/plans.json, written by
// testdata/gen from GameManager.LegalEvaluateLedger, LegalEvaluatePlan,
// LegalRenderVerdict and move.Legal), plus hand-written tests for the parts a
// Go vector cannot show: that evaluation really stops, and the host-supplied
// base checks.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import type { LegalHostCheck } from './access.ts';
import { defaultLegalRegistry } from './catalog.ts';
import { stackEmpty, stackNotEmpty } from './catalogStack.ts';
import { propAtLeast } from './catalogCompare.ts';
import { proposerIsCurrentPlayer } from './catalogPlayers.ts';
import { type LegalPreconditionEntry, buildPreconditionLedger } from './evaluable.ts';
import type { LegalVisibility } from './evaluable.ts';
import { legalErrorf, renderLegalVerdict } from './message.ts';
import {
  type LegalCustom,
  type LegalPlan,
  assembleLegalSpecList,
  buildLegalPlan,
  contributedPreconditions,
  evaluateLegalPlan,
  legalPlanError,
} from './plan.ts';
import { type LegalPredicateConstructor, LegalSpecError, newLegalPredicate } from './predicate.ts';
import * as fx from './testing/fixtures.ts';
import type { LegalRead, LegalSpec, LegalVerdict } from './types.ts';

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
}

interface GoEntry {
  name: string;
  args: string[];
  serializable: boolean;
  clientEvaluable: boolean;
  fieldDependent: boolean;
  reads: LegalRead[];
}

interface GoMove {
  game: string;
  name: string;
  contributed: LegalSpec[];
  authored: LegalSpec[];
  suppressions: string[];
  specs: LegalSpec[];
  hasCustom: boolean;
  customMode?: string;
  entries: GoEntry[] | null;
  gameRegistered?: string[];
}

type Scenario = [
  move: number,
  fixture: number,
  moveArgs: number,
  proposer: number,
  overall: number,
  verdicts: number[],
  preconditions: number[],
  rendered: number,
  hotPath: number,
  legalError: number,
];

const exported = JSON.parse(readFileSync(path.join(testdata, 'fixtures.json'), 'utf8')) as {
  games: Record<string, fx.WireGame>;
  fixtures: Record<string, fx.WireFixture>;
};

const plans = JSON.parse(readFileSync(path.join(testdata, 'plans.json'), 'utf8')) as {
  moves: GoMove[];
  views: Record<string, Record<string, LegalVisibility>>;
  assemblyErrors: { name: string; base: string; authored: LegalSpec[]; suppressions: string[]; templates: Record<string, string>; error: string }[];
  fixtureNames: string[];
  moveArgs: (fx.WireFixture['move'] | null)[];
  verdicts: GoVerdict[];
  preconditions: LegalPreconditionEntry[];
  texts: string[];
  scenarios: Scenario[];
};

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

function toVerdict(v: GoVerdict): LegalVerdict {
  return v as LegalVerdict;
}

/** The TypeScript LegalCustom for each mode the generator's synthetic moves use. */
const customByMode: Record<string, LegalCustom> = {
  nil: () => null,
  plain: () => 'custom says no',
  structured: () => legalErrorf('gen.custom', { count: 3, who: 'nobody', flag: true }),
  structuredBare: () => legalErrorf('gen.custom_bare'),
  unregistered: () => legalErrorf('gen.not_in_table', { n: 1 }),
  panic: () => {
    throw new Error('custom blew up');
  },
};

// What Go produced for the pieces this package does not implement and a
// replay therefore has to be told: a game's own registered predicates, the
// inProgression base check, and an example game's Go LegalCustom body. Set
// before each scenario is evaluated.
const recorded: { byName: Map<string, LegalVerdict> } = { byName: new Map() };

function recordedVerdict(name: string): LegalVerdict {
  const v = recorded.byName.get(name);
  assert.ok(v, `no recorded Go verdict for ${name}`);
  return v;
}

function hostCheck(v: LegalVerdict): LegalHostCheck {
  if (v.outcome === 'pass') {
    return null;
  }
  if (v.outcome === 'fail') {
    return (v.message?.bindings?.detail as string) ?? '';
  }
  return { unknown: v.reason ?? '' };
}

interface Replay {
  plan: LegalPlan;
  templates: Record<string, string>;
}

const replays = new Map<number, Replay>();

function replayFor(moveIndex: number, scenario: Scenario): Replay {
  const existing = replays.get(moveIndex);
  if (existing) {
    return existing;
  }
  const goMove = plans.moves[moveIndex];
  const entries = goMove.entries ?? [];
  const fixture = exported.fixtures[plans.fixtureNames[scenario[1]]];
  // Any non-null variant of the move gives the reader boot validation needs.
  const moveArgs = plans.scenarios
    .filter(s => s[0] === moveIndex)
    .map(s => plans.moveArgs[s[2]])
    .find(m => m !== null);
  const bound = fx.bindWireFixture({ ...fixture, move: moveArgs ?? null }, exported.games[fixture.game]);

  const stubs: LegalPredicateConstructor[] = (goMove.gameRegistered ?? []).map(name => {
    const entry = entries.find(e => e.name === name) as GoEntry;
    return {
      name,
      construct: spec =>
        newLegalPredicate({
          name,
          args: [...(spec.args ?? [])],
          reads: entry.reads,
          cost: 0,
          clientEvaluable: entry.clientEvaluable,
          evaluate: () => recordedVerdict(name),
        }),
    };
  });

  let custom: LegalCustom | null = null;
  if (goMove.hasCustom) {
    custom = goMove.customMode ? customByMode[goMove.customMode] : () => recordedVerdict('custom');
  }

  const plan = buildLegalPlan({
    moveName: goMove.name,
    contributed: goMove.contributed,
    authored: goMove.authored,
    suppressions: goMove.suppressions,
    custom,
    registry: defaultLegalRegistry(...stubs),
    chest: bound.chest,
    exampleState: bound.state,
    moveReader: bound.move ? bound.move.reader() : null,
    templates: bound.templates,
  });
  const replay = { plan, templates: bound.templates };
  replays.set(moveIndex, replay);
  return replay;
}

test('the plan vectors are present', () => {
  assert.ok(plans.scenarios.length > 1500, `only ${plans.scenarios.length} scenarios`);
  assert.ok(plans.moves.length > 40);
});

test('plan assembly matches Go for every recorded move', () => {
  for (const goMove of plans.moves) {
    assert.deepEqual(
      assembleLegalSpecList(goMove.contributed, goMove.authored, goMove.suppressions),
      goMove.specs,
      `${goMove.game}/${goMove.name}`,
    );
  }
});

const byGame = new Map<string, Scenario[]>();
for (const scenario of plans.scenarios) {
  const game = plans.moves[scenario[0]].game;
  byGame.set(game, [...(byGame.get(game) ?? []), scenario]);
}

// Go's own two answers for one scenario - the full ledger's overall verdict
// and the hot path's - are compared too. They differ in exactly four vectors,
// all an admin proposer on a hand-mutated state: the hot path serves a
// verdict memoized per (move, state version, proposer, predicate) during game
// setup, and the generator rearranges a state without advancing its version.
// The TypeScript has no memo, so it always gives the ledger's answer.
let goHotPathDisagreements = 0;

for (const [game, scenarios] of byGame) {
  test(`Go plans: ${game} (${scenarios.length} scenarios)`, () => {
    for (const scenario of scenarios) {
      const [moveIndex, fixtureIndex, moveArgsIndex, proposer, overallIndex, verdictIndexes, preconditionIndexes, renderedIndex, hotPathIndex, legalErrorIndex] = scenario;
      const goMove = plans.moves[moveIndex];
      const entries = goMove.entries ?? [];
      const fixtureName = plans.fixtureNames[fixtureIndex];
      const where = `${game}/${goMove.name} on ${fixtureName} with ${JSON.stringify(plans.moveArgs[moveArgsIndex]?.props ?? null)} as ${proposer}`;

      const replay = replayFor(moveIndex, scenario);
      const fixture = exported.fixtures[fixtureName];
      const bound = fx.bindWireFixture({ ...fixture, move: plans.moveArgs[moveArgsIndex] }, exported.games[fixture.game]);

      recorded.byName = new Map();
      entries.forEach((entry, i) => {
        if (entry.name === 'custom' || entry.name === 'inProgression' || (goMove.gameRegistered ?? []).includes(entry.name)) {
          recorded.byName.set(entry.name, toVerdict(plans.verdicts[verdictIndexes[i]]));
        }
      });

      const ctx = {
        state: bound.state,
        move: bound.move,
        proposer,
        framework: {
          currentPhase: bound.framework.currentPhase,
          moveInProgression: () => hostCheck(recordedVerdict('inProgression')),
        },
      };

      // The full ledger: one entry per predicate, in plan order.
      const ledger = evaluateLegalPlan(replay.plan, ctx, true);
      assert.ok(ledger.entries);
      assert.deepEqual(
        ledger.entries.map(e => ({
          name: e.name,
          args: [...e.args],
          serializable: e.serializable,
          clientEvaluable: e.clientEvaluable,
          fieldDependent: e.fieldDependent,
          reads: e.reads.map(r => ({ path: r.path, facet: r.facet })),
        })),
        entries,
        where,
      );
      assert.deepEqual(
        ledger.entries.map(e => normalize(e.verdict)),
        verdictIndexes.map(i => plans.verdicts[i]),
        where,
      );
      assert.deepEqual(normalize(ledger.verdict), plans.verdicts[overallIndex], where);

      // The hot path stops at the first non-pass and reports the same verdict.
      const hot = evaluateLegalPlan(replay.plan, ctx);
      assert.equal(hot.entries, null);
      assert.deepEqual(normalize(hot.verdict), plans.verdicts[overallIndex], where);

      // Rendered against the game's template table: the text move.Legal() returns.
      const text = renderLegalVerdict(ledger.verdict, replay.templates);
      assert.equal(text, plans.texts[renderedIndex], where);
      assert.equal(legalPlanError(replay.plan, ctx, replay.templates), text, where);
      const goAgrees = hotPathIndex === renderedIndex && (legalErrorIndex === -1 || legalErrorIndex === renderedIndex);
      if (!goAgrees) {
        goHotPathDisagreements++;
        assert.equal(proposer, -2, `Go's hot path and ledger disagree outside the known memo case: ${where}`);
      }

      // The per-viewer wire ledger, for the viewer the server builds it for.
      const view = {
        viewer: proposer,
        visibility: plans.views[fixtureName][String(proposer)] ?? null,
        currentPlayerIndex: fixture.currentPlayerIndex,
      };
      assert.deepEqual(
        buildPreconditionLedger(view, ledger.entries),
        preconditionIndexes.map(i => plans.preconditions[i]),
        where,
      );
    }
  });
}

test("Go's hot path and ledger disagree only in the four memoized-setup vectors", () => {
  assert.equal(goHotPathDisagreements, 4);
});

test('plan vectors exercise the things they exist for', () => {
  const overall = (s: Scenario) => plans.verdicts[s[4]];
  const names = (s: Scenario) => (plans.moves[s[0]].entries ?? []).map(e => e.name);
  const outcomes = (s: Scenario) => s[5].map(i => plans.verdicts[i].outcome);
  // Several predicates fail at once and the first in plan order wins.
  assert.ok(plans.scenarios.some(s => outcomes(s).filter(o => o !== 'pass').length >= 3));
  // A later failure exists but an earlier unknown is the plan's verdict.
  assert.ok(plans.scenarios.some(s => overall(s).outcome === 'unknown' && outcomes(s).includes('fail')));
  // Each base check this package implements is, somewhere, the first failure.
  // inProgression is in real plans too (checkers' Place Token) but only ever
  // passes there; its fail and unknown paths are covered by the hand-written
  // base-check test below.
  assert.ok(plans.scenarios.some(s => names(s).includes('inProgression')));
  for (const base of ['inPhase', 'stackConstraints', 'proposerIsCurrentPlayer']) {
    assert.ok(
      plans.scenarios.some(s => {
        const firstBad = outcomes(s).findIndex(o => o !== 'pass');
        return firstBad >= 0 && names(s)[firstBad] === base;
      }),
      `no scenario where ${base} is the first failure`,
    );
  }
  // The any compositor passes, fails and is unknown at the top of a plan.
  for (const want of ['pass', 'fail', 'unknown']) {
    assert.ok(plans.scenarios.some(s => names(s).some((n, i) => n === 'any' && outcomes(s)[i] === want)), `any is never ${want}`);
  }
  // LegalCustom is reached and rejects; and is never reached when an earlier check fails.
  assert.ok(plans.scenarios.some(s => names(s).at(-1) === 'custom' && overall(s).message?.template === 'custom says no'));
  // An admin proposer bypasses a predicate that fails for everyone else.
  assert.ok(plans.scenarios.some(s => s[3] === -2 && plans.moves[s[0]].name === 'Gen Admin' && outcomes(s)[1] === 'pass'));
  // Bindings are stripped from a ledger line the viewer could not evaluate.
  assert.ok(plans.preconditions.some(p => p.verdict === 'fail' && !p.evaluable && p.message && !p.message.bindings));
  assert.ok(plans.preconditions.some(p => p.verdict === 'fail' && p.evaluable && p.message?.bindings));
});

test('assembly errors match Go', () => {
  const fixture = exported.fixtures['plan/synthMemory/fresh'];
  const game = exported.games[fixture.game];
  assert.ok(plans.assemblyErrors.length >= 30);
  for (const c of plans.assemblyErrors) {
    const currentPlayer = c.base === 'currentPlayer';
    const bound = fx.bindWireFixture(fixture, game);
    const moveReader = fx.move(currentPlayer ? { TargetPlayerIndex: fx.playerIndex(0) } : {}).reader();
    assert.throws(
      () =>
        buildLegalPlan({
          moveName: 'Gen Bad',
          contributed: contributedPreconditions({ moveName: 'Gen Bad', currentPlayer }),
          authored: c.authored,
          suppressions: c.suppressions,
          registry: defaultLegalRegistry(),
          chest: bound.chest,
          exampleState: bound.state,
          moveReader,
          // Go's table: the defaults, the game's own, and this case's.
          templates: { ...bound.templates, ...c.templates },
        }),
      (err: unknown) => {
        assert.ok(err instanceof LegalSpecError, `${c.name}: ${String(err)}`);
        assert.equal(err.message, c.error, c.name);
        return true;
      },
      c.name,
    );
  }
});

// ---------------------------------------------------------------------------
// Hand-written: what a recorded verdict cannot show.

function countingConstructor(name: string, verdict: LegalVerdict, calls: string[]): LegalPredicateConstructor {
  return {
    name,
    construct: () =>
      newLegalPredicate({
        name,
        reads: [],
        cost: 0,
        evaluate() {
          calls.push(name);
          return verdict;
        },
      }),
  };
}

test('the hot path does not evaluate anything after the first non-pass', () => {
  const calls: string[] = [];
  const registry = defaultLegalRegistry(
    countingConstructor('first', { outcome: 'pass' }, calls),
    countingConstructor('second', { outcome: 'fail', message: { template: 'second.failed' } }, calls),
    countingConstructor('third', { outcome: 'fail', message: { template: 'third.failed' } }, calls),
  );
  const plan = buildLegalPlan({
    moveName: 'M',
    authored: [{ name: 'first' }, { name: 'second' }, { name: 'third' }],
    custom: () => {
      calls.push('custom');
      return 'custom.failed';
    },
    registry,
  });
  const ctx = { state: fx.state(), move: fx.move({}), proposer: 0 };

  assert.deepEqual(evaluateLegalPlan(plan, ctx).verdict, { outcome: 'fail', message: { template: 'second.failed' } });
  assert.deepEqual(calls, ['first', 'second']);

  calls.length = 0;
  const ledger = evaluateLegalPlan(plan, ctx, true);
  assert.deepEqual(calls, ['first', 'second', 'third', 'custom']);
  assert.deepEqual(ledger.verdict, { outcome: 'fail', message: { template: 'second.failed' } });
  assert.deepEqual(ledger.entries?.map(e => e.verdict.outcome), ['pass', 'fail', 'fail', 'fail']);
});

test('an unknown stops the hot path just as a fail does', () => {
  const calls: string[] = [];
  const plan = buildLegalPlan({
    moveName: 'M',
    authored: [{ name: 'murky' }, { name: 'clear' }],
    registry: defaultLegalRegistry(
      countingConstructor('murky', { outcome: 'unknown', reason: 'cannot tell' }, calls),
      countingConstructor('clear', { outcome: 'fail', message: { template: 'clear.failed' } }, calls),
    ),
  });
  const ctx = { state: fx.state(), move: fx.move({}), proposer: 0 };
  assert.deepEqual(evaluateLegalPlan(plan, ctx).verdict, { outcome: 'unknown', reason: 'cannot tell' });
  assert.deepEqual(calls, ['murky']);
  assert.equal(legalPlanError(plan, ctx, {}), 'cannot tell');
});

test('LegalCustom runs last, and only when everything before it passed', () => {
  const calls: string[] = [];
  const build = (verdict: LegalVerdict) =>
    buildLegalPlan({
      moveName: 'M',
      authored: [{ name: 'gate' }],
      custom: () => {
        calls.push('custom');
        return legalErrorf('custom.no', { n: 2 });
      },
      registry: defaultLegalRegistry(countingConstructor('gate', verdict, calls)),
    });
  const ctx = { state: fx.state(), move: fx.move({}), proposer: 0 };

  assert.deepEqual(evaluateLegalPlan(build({ outcome: 'pass' }), ctx).verdict, {
    outcome: 'fail',
    message: { template: 'custom.no', bindings: { n: 2 } },
  });
  assert.deepEqual(calls, ['gate', 'custom']);

  calls.length = 0;
  evaluateLegalPlan(build({ outcome: 'fail', message: { template: 'gate.no' } }), ctx);
  assert.deepEqual(calls, ['gate']);
});

test('the custom entry is opaque: not serializable, not client evaluable, field dependent, no reads', () => {
  const plan = buildLegalPlan({ moveName: 'M', custom: () => null, registry: defaultLegalRegistry() });
  const { entries } = evaluateLegalPlan(plan, { state: fx.state(), move: fx.move({}), proposer: 0 }, true);
  assert.deepEqual(entries, [
    { name: 'custom', args: [], verdict: { outcome: 'pass' }, serializable: false, clientEvaluable: false, fieldDependent: true, reads: [] },
  ]);
});

test('a missing plan fails closed', () => {
  assert.deepEqual(evaluateLegalPlan(null, { state: fx.state(), move: null, proposer: 0 }), {
    verdict: { outcome: 'unknown', reason: 'nil legal plan' },
    entries: null,
  });
});

test('contributedPreconditions orders the base checks phase, progression, stack constraints, proposer', () => {
  assert.deepEqual(
    contributedPreconditions({
      moveName: 'Draw Card',
      currentPlayer: true,
      stackConstraints: { source: 'Deck', destination: 'Discard' },
      hasMoveProgression: true,
      legalPhases: [2, 3],
    }),
    [
      { name: 'inPhase', args: ['2', '3'] },
      { name: 'inProgression', args: ['Draw Card'] },
      { name: 'stackConstraints', args: ['Deck', 'Discard'] },
      { name: 'proposerIsCurrentPlayer' },
    ],
  );
  // Only what is configured is contributed.
  assert.deepEqual(contributedPreconditions({ moveName: 'M' }), []);
  assert.deepEqual(contributedPreconditions({ moveName: 'M', legalPhases: [], currentPlayer: true }), [proposerIsCurrentPlayer()]);
});

test('the base checks run before authored specs and take the framework as an input', () => {
  const cards = fx.deck('cards', [null, null]);
  const draw = new fx.FixtureStack('cards', 'growable', [cards[0], cards[1]]);
  const discard = new fx.FixtureStack('cards', 'growable', []);
  const state = fx.state({
    game: { Phase: 0, Draw: draw, Discard: discard, CurrentPlayer: fx.playerIndex(0) },
    players: [{ Score: 0 }, { Score: 0 }],
  });
  const plan = buildLegalPlan({
    moveName: 'Draw Card',
    contributed: contributedPreconditions({
      moveName: 'Draw Card',
      legalPhases: [1],
      hasMoveProgression: true,
      stackConstraints: { source: 'Draw', destination: 'Discard' },
      currentPlayer: true,
    }),
    authored: [propAtLeast('player.Score', 5), stackNotEmpty('game.Draw')],
    registry: defaultLegalRegistry(),
  });
  assert.deepEqual(plan.steps.map(s => s.predicate.name), ['inPhase', 'inProgression', 'stackConstraints', 'proposerIsCurrentPlayer', 'propAtLeast', 'stackNotEmpty']);
  assert.deepEqual(plan.steps.map(s => s.fieldDependent), [false, false, false, true, false, false]);

  const move = fx.move({ TargetPlayerIndex: fx.playerIndex(0) });
  const asked: string[] = [];
  const run = (phaseKey: number, progression: LegalHostCheck, proposer = 0) =>
    evaluateLegalPlan(plan, {
      state,
      move,
      proposer,
      framework: fx.framework({ key: phaseKey, name: phaseKey === 1 ? 'Playing' : 'Setup', ancestors: null }, (name, _state, who) => {
        asked.push(`${name}:${who}`);
        return progression;
      }),
    }).verdict;

  // Every base check fails here; the phase check is first.
  assert.deepEqual(run(0, 'not the next move', 1), {
    outcome: 'fail',
    message: { template: 'legal.in_phase', bindings: { detail: 'Move is not legal in phase Setup' } },
  });
  assert.deepEqual(asked, [], 'progression is not consulted once the phase check has failed');
  // Phase passes: progression is next, and is asked about this move and proposer.
  assert.deepEqual(run(1, 'not the next move', 1), {
    outcome: 'fail',
    message: { template: 'legal.in_progression', bindings: { detail: 'not the next move' } },
  });
  assert.deepEqual(asked, ['Draw Card:1']);
  assert.deepEqual(run(1, { unknown: 'no history' }), { outcome: 'unknown', reason: 'no history' });
  // Then the proposer check, then the author's first spec.
  assert.equal(run(1, null, 1).message?.template, 'legal.proposer_not_your_turn');
  assert.deepEqual(run(1, null), {
    outcome: 'fail',
    message: { template: 'legal.prop_at_least', bindings: { value: 0, min: 5 } },
  });
});

test('inPhase and inProgression are unknown, never pass, when no framework is supplied', () => {
  const plan = (specs: LegalSpec[]) => buildLegalPlan({ moveName: 'M', contributed: specs, registry: defaultLegalRegistry() });
  const ctx = { state: fx.state({ game: { Phase: 0 } }), move: null, proposer: 0 };
  assert.equal(evaluateLegalPlan(plan([{ name: 'inPhase', args: ['1'] }]), ctx).verdict.outcome, 'unknown');
  assert.equal(evaluateLegalPlan(plan([{ name: 'inProgression', args: ['M'] }]), ctx).verdict.outcome, 'unknown');
  // A phase check with no phases is legal everywhere and needs no framework.
  assert.equal(evaluateLegalPlan(plan([{ name: 'inPhase', args: [] }]), ctx).verdict.outcome, 'pass');
});

test('inPhase accepts a descendant of a legal phase through the phase ancestors', () => {
  const plan = buildLegalPlan({ moveName: 'M', contributed: [{ name: 'inPhase', args: ['1'] }], registry: defaultLegalRegistry() });
  const at = (ancestors: number[] | null) =>
    evaluateLegalPlan(plan, {
      state: fx.state({ game: { Phase: 0 } }),
      move: null,
      proposer: 0,
      framework: fx.framework({ key: 5, name: 'Playing > Bidding', ancestors }),
    }).verdict;
  assert.equal(at([5, 1, 0]).outcome, 'pass');
  assert.deepEqual(at(null), {
    outcome: 'fail',
    message: { template: 'legal.in_phase', bindings: { detail: 'Move is not legal in phase Playing > Bidding' } },
  });
});

test('suppression removes a contributed spec and never an authored one', () => {
  const contributed = [{ name: 'inPhase', args: ['1'] }, proposerIsCurrentPlayer()];
  const authored = [stackEmpty('game.Discard'), { name: 'inPhase', args: ['2'] }];
  assert.deepEqual(assembleLegalSpecList(contributed, authored, ['inPhase']), [proposerIsCurrentPlayer(), ...authored]);
  assert.throws(
    () => buildLegalPlan({ moveName: 'M', contributed, authored, suppressions: ['stackConstraints'], registry: defaultLegalRegistry() }),
    /move "M" suppresses "stackConstraints".*contributed spec names: "inPhase", "proposerIsCurrentPlayer"\)$/,
  );
});
