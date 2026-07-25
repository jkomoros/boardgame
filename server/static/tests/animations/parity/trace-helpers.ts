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
// the rolled value). Structural mode still enforces the invariants that
// catch real regressions -- every open matched by a close, zero watchdogs,
// every play settled (asserted for every mode above), at least one gated
// cycle -- plus that every kind named in `requiredKinds` actually animated.
// Exact kind-set equality is deliberately NOT asserted: branch-dependent
// extras (a score fade that only happens on a scoring roll) would make it
// flaky for exactly the scenarios structural mode exists for. The golden is
// still recorded for human inspection/debugging.
export function expectTraceMatchesGolden(
  trace: ParityTrace,
  name: string,
  opts: { structural?: { requiredKinds: string[]; exactCycles?: boolean } } = {},
): void {
  const goldenPath = join(GOLDEN_DIR, `${name}.json`);
  expect(trace.gateDelta.watchdogFirings, 'watchdog must never fire').toBe(0);
  expect(trace.gateDelta.settles, 'every play must settle inside the capture window')
    .toBe(trace.gateDelta.plays);
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
    return;
  }
  expect(trace.gateDelta).toEqual(golden.gateDelta);
  expect(perElement(canonicalize(trace.events))).toEqual(perElement(canonicalize(golden.events)));
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
