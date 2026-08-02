import { Page, expect } from '@playwright/test';
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  gateSnapshot,
  waitForAnimationCounterStability,
  waitForClientQuiescence,
  GateSnapshot,
} from '../helpers.js';

export interface ParityEvent {
  kind: string;      // play | active | settle | gate-open | gate-close | watchdog | install
  detail: string;    // element identity, e.g. 'boardgame-card#card-3'
  version?: number;  // version slot when the hook recorded one
}

export interface ParityTrace {
  events: ParityEvent[];
  gateDelta: GateSnapshot; // counters attributable to the scenario
}

// import.meta.url, not __dirname: this test suite runs under Playwright's
// ESM TS loader (package.json "type": "module"), where __dirname is
// undefined -- confirmed empirically against this checkout's Playwright
// version before writing this file.
const GOLDEN_DIR = join(dirname(fileURLToPath(import.meta.url)), 'goldens');

// Captures the animHooks event log emitted while `scenario` runs, waits for
// full client quiescence, and returns the normalized trace. Element details
// are kept verbatim (tag + id); game ids never appear in hook details.
//
// `scenario` may itself navigate to a fresh game page (e.g. createOfflineGame)
// where __bgAnimTestHooks does not exist yet at call time -- the log-start
// sample is taken lazily, after hooks exist, by polling for them both before
// and after the scenario runs and treating a hooks-reset (fresh page) as a
// zero-based log start.
export async function captureTrace(
  page: Page,
  scenario: () => Promise<void>,
): Promise<ParityTrace> {
  const hooksExisted = await page.evaluate(() => (window as any).__bgAnimTestHooks !== undefined);
  const before = hooksExisted ? await gateSnapshot(page) : { gateOpens: 0, gateCloses: 0, watchdogFirings: 0, plays: 0, settles: 0 };
  const logStart = hooksExisted
    ? await page.evaluate(() => (window as any).__bgAnimTestHooks.log.length)
    : 0;
  await scenario();
  await waitForClientQuiescence(page);
  // Trailing-edge determinism: quiescence is a point-in-time check, but a
  // late fix-up bundle (or a settle scheduled on the next frame) can land
  // right after it passes, leaving the window with plays > settles purely
  // by sampling race. Close the window only after sustained stability.
  await waitForAnimationCounterStability(page, { balance: 'plays' });
  const after = await gateSnapshot(page);
  // If the scenario navigated to a page that didn't have hooks before it
  // ran, the hooks object is a fresh instance and `logStart` (0) is already
  // correct for it; there is no way (and no need) to distinguish that from
  // "the counters started at zero".
  const events: ParityEvent[] = await page.evaluate((start) => {
    return (window as any).__bgAnimTestHooks.log.slice(start).map((e: any) => {
      const out: any = { kind: e.ev, detail: e.detail ?? '' };
      if (e.version !== undefined) out.version = e.version;
      return out;
    });
  }, logStart);
  return {
    events,
    gateDelta: {
      gateOpens: after.gateOpens - before.gateOpens,
      gateCloses: after.gateCloses - before.gateCloses,
      watchdogFirings: after.watchdogFirings - before.watchdogFirings,
      plays: after.plays - before.plays,
      settles: after.settles - before.settles,
    },
  };
}

// Compares against (or in PARITY_RECORD=1 mode, rewrites) the golden.
// Event ORDER is compared per-element (each element's own event sequence
// must match exactly); global interleaving across distinct elements is
// allowed to vary -- WAAPI settlement order between unrelated elements is
// not deterministic. Counters must match exactly; watchdog must be 0.
//
// `structural` relaxes the comparison for scenarios whose animation COUNT is
// inherently randomized by game logic the test cannot control (blackjack's
// deal length depends on the shuffled deck; pig's post-roll cycles depend on
// the rolled value).
//
// STRUCTURAL MODE USED TO READ NOTHING OUT OF THE GOLDEN. Without
// `exactCycles` -- which is exactly blackjack and pig -- `golden` was parsed
// and never dereferenced; `existsSync` was the only use the file's contents
// got. `blackjack-deal.json`'s 3,214 recorded events and `pig-roll.json`'s 37
// pinned four scalars between them, and a run that opened the gate once and
// animated a single card satisfied every blackjack assertion: 1,057 of 1,058
// plays could vanish silently.
//
// What the golden now supplies, chosen by measuring what survives a reshuffle
// (four fresh randomized games per scenario, plus the goldens themselves):
//
//   * ITS OWN ELEMENT KINDS. Every tag the golden recorded must animate --
//     `requiredKinds` generalized, and read off the file instead of retyped.
//     Extra kinds are still allowed (pig's fade only happens on a scoring
//     roll), so this is one-directional on purpose.
//   * PER-KIND DISTINCT-ELEMENT FLOOR. At least as many DISTINCT elements of
//     each recorded kind must animate as the golden recorded. This is the
//     invariant a shuffled deal cannot move: blackjack relayouts the whole
//     deck, so exactly 52 distinct `boardgame-card`s animate in the golden
//     AND in all four fresh deals measured, while their play counts ranged
//     1,058-1,523. A floor rather than equality because the recorded goldens
//     are behind current behavior in the OTHER direction (debuganimations
//     records 54 distinct cards where four consecutive runs now produce 55),
//     and the hole being closed is elements DISAPPEARING.
//   * A VOLUME FLOOR where the scenario has one (`minPlayFraction`), as a
//     fraction of the golden's own play count.
//
// Not asserted, with reasons: exact kind-set equality (branch-dependent
// extras would flake it); exact play counts (that is what structural mode
// exists for); a volume floor for pig (its post-roll cycle count genuinely
// branches on the rolled value -- measured 3 plays on three runs and 11 on a
// fourth, so no fraction of the golden's 11 is both safe and meaningful).
//
// `assertPerElementCycleGrammar` runs for EVERY mode; see its own note.
export function expectTraceMatchesGolden(
  trace: ParityTrace,
  name: string,
  opts: {
    structural?: {
      requiredKinds: string[];
      exactCycles?: boolean;
      /**
       * Floor on `plays`, as a fraction of the golden's. Omit for scenarios
       * whose volume genuinely branches. Set it from a MEASURED spread, and
       * say what was measured at the call site.
       */
      minPlayFraction?: number;
    };
  } = {},
): void {
  const goldenPath = join(GOLDEN_DIR, `${name}.json`);
  expect(trace.gateDelta.watchdogFirings, 'watchdog must never fire').toBe(0);
  expect(trace.gateDelta.settles, 'every play must settle inside the capture window')
    .toBe(trace.gateDelta.plays);
  assertPerElementCycleGrammar(trace.events);
  if (!opts.structural) {
    // Not asserted in structural mode: those scenarios' windows contain
    // whole-game setup whose cycle traffic is game-randomness-dependent;
    // their open/close balance is covered instead by the waapi-gate suite's
    // creation and reinstall regression tests.
    expect(trace.gateDelta.gateCloses, 'every gate open must close inside the capture window')
      .toBe(trace.gateDelta.gateOpens);
  }
  if (process.env.PARITY_RECORD === '1') {
    mkdirSync(dirname(goldenPath), { recursive: true });
    writeFileSync(goldenPath, JSON.stringify(trace, null, 2) + '\n');
    return;
  }
  if (!existsSync(goldenPath)) {
    throw new Error(`missing golden ${name}; record with PARITY_RECORD=1`);
  }
  const golden: ParityTrace = JSON.parse(readFileSync(goldenPath, 'utf-8'));
  if (opts.structural) {
    expect(trace.gateDelta.gateOpens, 'scenario must drive at least one gated cycle')
      .toBeGreaterThan(0);
    if (opts.structural.exactCycles) {
      // Cycle STRUCTURE is deterministic even when per-component play
      // counts are not (FLIP skips no-op transforms, and messy-stack
      // rotations are hashed from per-game random component ids, so how
      // many components actually play varies per game).
      expect(trace.gateDelta.gateOpens, 'gated cycle count must match the golden')
        .toBe(golden.gateDelta.gateOpens);
      expect(trace.gateDelta.gateCloses, 'gated cycle closes must match the golden')
        .toBe(golden.gateDelta.gateCloses);
    }
    const kinds = elementKinds(trace.events);
    for (const required of opts.structural.requiredKinds) {
      expect(kinds, `required element kind "${required}" must animate`).toContain(required);
    }
    // Everything below reads the golden. See the note above this function for
    // why each of these, and only these, survives a reshuffle.
    const observed = distinctElementsPerKind(trace.events);
    const recorded = distinctElementsPerKind(golden.events);
    for (const [kind, goldenCount] of [...recorded].sort((a, b) => a[0].localeCompare(b[0]))) {
      expect(observed.get(kind) ?? 0,
        `the golden recorded ${goldenCount} distinct "${kind}" element(s) animating; `
        + 'at least that many must still animate')
        .toBeGreaterThanOrEqual(goldenCount);
    }
    if (opts.structural.minPlayFraction !== undefined) {
      const floor = Math.ceil(golden.gateDelta.plays * opts.structural.minPlayFraction);
      expect(trace.gateDelta.plays,
        `scenario must play at least ${opts.structural.minPlayFraction} of the golden's `
        + `${golden.gateDelta.plays} animations`)
        .toBeGreaterThanOrEqual(floor);
    }
    return;
  }
  expect(trace.gateDelta).toEqual(golden.gateDelta);
  expect(perElement(canonicalize(trace.events))).toEqual(perElement(canonicalize(golden.events)));
}

// How many DISTINCT elements of each kind appear in a trace. Not how many
// times they played -- that is the number game randomness moves. WHICH
// components take part in a relayout is a property of the board, not of the
// shuffle: blackjack's deal relayouts the whole 52-card deck whatever order it
// comes out in.
function distinctElementsPerKind(events: ParityEvent[]): Map<string, number> {
  const perKind = new Map<string, Set<string>>();
  for (const event of events) {
    if (event.kind === 'gate-open' || event.kind === 'gate-close') continue;
    const hashIdx = event.detail.indexOf('#');
    const kind = hashIdx < 0 ? event.detail : event.detail.slice(0, hashIdx);
    if (kind === '') continue;
    let seen = perKind.get(kind);
    if (seen === undefined) perKind.set(kind, (seen = new Set()));
    seen.add(event.detail);
  }
  return new Map([...perKind].map(([kind, seen]) => [kind, seen.size]));
}

// Every element's own play/active/settle traffic must balance, and its
// sequence must open with a play and close with a settle.
//
// This is an invariant of the OBSERVED stream rather than a golden, and it is
// asserted for every mode deliberately -- including the exact ones, where the
// per-element sequence comparison already implies it. The point is that
// `active` was previously compared NOWHERE, in any mode: the gate counters
// only track plays and settles, so an element that announced itself and then
// never became active (or became active twice) was invisible to the whole
// harness. Aggregate `settles === plays` cannot see it either, because it
// sums across elements -- one element's missing settle hides behind another's
// extra.
//
// Measured across the four goldens and twelve fresh scenario runs: zero
// violations, so this is a floor rather than a guess about how much slack the
// capture window needs. If a window ever does slice an element mid-flight,
// the aggregate `settles === plays` above fails first.
function assertPerElementCycleGrammar(events: ParityEvent[]): void {
  const perEl = perElement(events.filter((e) => e.detail !== ''));
  for (const [detail, sequence] of Object.entries(perEl).sort(
    (a, b) => a[0].localeCompare(b[0]))) {
    const kinds = sequence.map((e) => e.kind);
    const count = (kind: string): number => kinds.filter((k) => k === kind).length;
    expect(
      { plays: count('play'), actives: count('active'), settles: count('settle') },
      `${detail}: every play must have exactly one active and one settle`,
    ).toEqual({ plays: count('play'), actives: count('play'), settles: count('play') });
    expect(kinds[0], `${detail}: an element's first event must be its play`).toBe('play');
    expect(kinds[kinds.length - 1], `${detail}: an element's last event must be its settle`)
      .toBe('settle');
  }
}

// The sorted set of tag names (id suffixes stripped) that appear in a
// trace's events -- the structural-mode identity of "what kinds of things
// animated".
function elementKinds(events: ParityEvent[]): string[] {
  return [...new Set(events.map((e) => {
    const hashIdx = e.detail.indexOf('#');
    return hashIdx < 0 ? e.detail : e.detail.slice(0, hashIdx);
  }))].sort();
}

function perElement(events: ParityEvent[]): Record<string, ParityEvent[]> {
  const out: Record<string, ParityEvent[]> = {};
  for (const e of events) (out[e.detail] ??= []).push(e);
  return out;
}

// Rewrites each event's `detail` id suffix (the part after '#') to a
// canonical index assigned in order of first appearance within this trace,
// independently on each side of the comparison. Some scenarios (e.g. memory,
// which shuffles its deck at game creation) make *which* physical component
// id ends up as "the first clickable card" nondeterministic across separate
// game creations, even though the sequence of actions taken (and thus the
// sequence of elements touched) is identical every time. Golden files still
// store the real recorded ids (useful for debugging); only the comparison
// path canonicalizes, so this never masks a real per-element event-order or
// count regression -- it only makes the comparison blind to *which*
// arbitrary id a given position's component happened to be assigned.
function canonicalize(events: ParityEvent[]): ParityEvent[] {
  const idMap = new Map<string, string>();
  let counter = 0;
  return events.map((e) => {
    const hashIdx = e.detail.indexOf('#');
    if (hashIdx < 0) return e; // no id suffix to normalize (e.g. '', or a bare tag)
    let canon = idMap.get(e.detail);
    if (canon === undefined) {
      canon = `${e.detail.slice(0, hashIdx)}#${counter++}`;
      idMap.set(e.detail, canon);
    }
    return { ...e, detail: canon };
  });
}
