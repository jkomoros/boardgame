# Animatable Item Unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every game-semantic animated element derives from `BoardgameAnimatableItem`; the legacy stack CSS-transition path is retired; #714's discovery/gating gaps are closed — all with mechanically verified zero regression of stack animations.

**Architecture:** A Playwright parity harness (trace + paused-geometry goldens recorded from the merge base) gates every phase. Classing migrations replace CSS keyframes with the gated WAAPI `play()` kernel. The completion gate is extracted to a pure kernel so player-info subtrees can participate. The stack's ambient CSS transition is replaced by an explicit `layoutTransform` setter that self-animates via `play()` — inheriting `noAnimate` suppression for free.

**Tech Stack:** Lit 3, TypeScript, WAAPI, Playwright (`npm run test:e2e`), node:test (`npm run test:unit`), boardgame-util offline-dev-mode server.

**Spec:** `docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md` — read it first; its Verification protocol and declared-changes list govern every task.

## Global Constraints

- Working dir for all client commands: `server/static/` inside the worktree `.claude/worktrees/animatable-item-unification`.
- Go commands in the worktree need `GOWORK=off`.
- E2E needs the server: `cd <worktree root> && nohup ./boardgame-util/boardgame-util serve --offline-dev-mode > server.log 2>&1 &` then verify `curl -s http://localhost:8080/client_config.js | grep offline_dev_mode` shows `true`. Build `boardgame-util` first if missing: `GOWORK=off go build -o boardgame-util/boardgame-util ./boardgame-util`.
- **Zero-regression rule:** only behavior changes explicitly declared in the spec may land. Any other observed deviation: stop, preserve current behavior, record an evidence pack under `docs/superpowers/specs/evidence/`.
- **Verify before write:** line numbers below were read at commit `4f8253f6`. Re-read the target region before each edit; adapt mechanically if drifted.
- Every phase boundary: `npm run type-check && npm run type-check:strict && npm run test:unit` clean, parity suite green, then the sub-agent critic/verifier gate (spec §Verification protocol) before the next phase starts.
- Commit format: imperative subject, `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>` trailer.

---

## Phase 0 — Parity harness

### Task 1: Trace recorder + trace parity scenarios

**Files:**
- Create: `server/static/tests/animations/parity/trace-helpers.ts`
- Create: `server/static/tests/animations/parity/trace.spec.ts`
- Create: `server/static/tests/animations/parity/goldens/` (recorded JSON)
- Reference (read, don't modify): `tests/animations/helpers.ts`, `tests/animations/waapi-play.spec.ts`, `src/utils/anim-test-hooks.ts`

**Interfaces:**
- Produces: `captureTrace(page, scenario: () => Promise<void>): Promise<ParityTrace>` and `expectTraceMatchesGolden(trace, goldenName: string)` where `ParityTrace = { events: {kind: string, detail: string, version?: number}[], gate: GateSnapshot }`.
- Golden regeneration: `PARITY_RECORD=1 npx playwright test tests/animations/parity/` rewrites goldens; default mode compares.

- [ ] **Step 1: Read `src/utils/anim-test-hooks.ts`** to confirm the shape of `window.__bgAnimTestHooks.log` entries (kind + detail + optional context recorded by `record()`), and read `tests/animations/waapi-play.spec.ts` to copy the established scenario idioms (createOfflineGame, gateSnapshot, expectCleanGate).

- [ ] **Step 2: Write `trace-helpers.ts`**

```typescript
import { Page, expect } from '@playwright/test';
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { gateSnapshot, waitForClientQuiescence, GateSnapshot } from '../helpers.js';

export interface ParityEvent {
  kind: string;      // play | active | settle | gate-open | gate-close | watchdog
  detail: string;    // element identity, e.g. 'boardgame-card#card-3'
  version?: number;  // version slot when the hook recorded one
}

export interface ParityTrace {
  events: ParityEvent[];
  gateDelta: GateSnapshot; // counters attributable to the scenario
}

const GOLDEN_DIR = join(__dirname, 'goldens');

// Captures the animHooks event log emitted while `scenario` runs, waits for
// full client quiescence, and returns the normalized trace. Element details
// are kept verbatim (tag + id); game ids never appear in hook details.
export async function captureTrace(
  page: Page,
  scenario: () => Promise<void>,
): Promise<ParityTrace> {
  const before = await gateSnapshot(page);
  const logStart = await page.evaluate(
    () => (window as any).__bgAnimTestHooks.log.length);
  await scenario();
  await waitForClientQuiescence(page);
  const after = await gateSnapshot(page);
  const events: ParityEvent[] = await page.evaluate((start) => {
    return (window as any).__bgAnimTestHooks.log.slice(start).map((e: any) => {
      const out: any = { kind: e.kind, detail: e.detail ?? '' };
      if (e.context?.version !== undefined) out.version = e.context.version;
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
// allowed to vary — WAAPI settlement order between unrelated elements is
// not deterministic. Counters must match exactly; watchdog must be 0.
export function expectTraceMatchesGolden(trace: ParityTrace, name: string): void {
  const goldenPath = join(GOLDEN_DIR, `${name}.json`);
  expect(trace.gateDelta.watchdogFirings, 'watchdog must never fire').toBe(0);
  if (process.env.PARITY_RECORD === '1') {
    mkdirSync(dirname(goldenPath), { recursive: true });
    writeFileSync(goldenPath, JSON.stringify(trace, null, 2) + '\n');
    return;
  }
  if (!existsSync(goldenPath)) {
    throw new Error(`missing golden ${name}; record with PARITY_RECORD=1`);
  }
  const golden: ParityTrace = JSON.parse(readFileSync(goldenPath, 'utf-8'));
  expect(trace.gateDelta).toEqual(golden.gateDelta);
  expect(perElement(trace.events)).toEqual(perElement(golden.events));
}

function perElement(events: ParityEvent[]): Record<string, ParityEvent[]> {
  const out: Record<string, ParityEvent[]> = {};
  for (const e of events) (out[e.detail] ??= []).push(e);
  return out;
}
```

- [ ] **Step 3: Write `trace.spec.ts`** with four scenarios. Model each scenario's driving actions on the existing specs for that game (read them first):

```typescript
import { test } from '@playwright/test';
import { createOfflineGame } from '../helpers.js';
import { captureTrace, expectTraceMatchesGolden } from './trace-helpers.js';

test.describe('animation parity traces', () => {
  test('debuganimations: card move cycle', async ({ page }) => {
    await createOfflineGame(page, 'debuganimations');
    const trace = await captureTrace(page, async () => {
      // Use the same move-driving action waapi-play.spec.ts uses (a move
      // button that flies a card between stacks). One deterministic move.
      await page.getByRole('button', { name: /Move Card/i }).first().click();
    });
    expectTraceMatchesGolden(trace, 'debuganimations-card-move');
  });

  test('memory: reveal two cards', async ({ page }) => {
    await createOfflineGame(page, 'memory');
    const trace = await captureTrace(page, async () => {
      const cards = page.locator('boardgame-card');
      await cards.nth(0).click();
      await cards.nth(1).click();
    });
    expectTraceMatchesGolden(trace, 'memory-reveal-two');
  });

  test('blackjack: initial deal settles cleanly', async ({ page }) => {
    // Deal auto-runs on creation; capture from game load (allowAlreadySettled
    // pattern from helpers.ts applies — see expectCleanGate docs).
    const trace = await captureTrace(page, async () => {
      await createOfflineGame(page, 'blackjack');
    });
    expectTraceMatchesGolden(trace, 'blackjack-deal');
  });

  test('pig: die roll', async ({ page }) => {
    await createOfflineGame(page, 'pig');
    const trace = await captureTrace(page, async () => {
      await page.getByRole('button', { name: /Roll/i }).click();
    });
    expectTraceMatchesGolden(trace, 'pig-roll');
  });
});
```

Note: `pig` is not in `GAME_TYPE_LABELS` in `tests/animations/helpers.ts` — add `pig: 'Pig'` (verify the display name in the game's client config) to that map as part of this task. `captureTrace` around `createOfflineGame` must tolerate hooks not existing pre-navigation: move the `gateSnapshot`/log-start sampling to after the page has hooks (adapt: for the blackjack scenario take `logStart = 0` after load and full counters as the delta).

- [ ] **Step 4: Record goldens** — server running, then from `server/static`: `PARITY_RECORD=1 npx playwright test tests/animations/parity/trace.spec.ts`. Expected: 4 passed, 4 goldens written.
- [ ] **Step 5: Verify comparison mode passes** — `npx playwright test tests/animations/parity/trace.spec.ts`. Expected: 4 passed.
- [ ] **Step 6: Verify the harness FAILS when behavior changes** — temporarily edit `src/components/boardgame-animatable-item.ts` `play()` to skip recording one hook (or set `noAnimate` short-circuit), rerun, expect failures; revert the edit. This proves the harness has teeth.
- [ ] **Step 7: Commit** — `git add server/static/tests/animations/parity && git commit -m "Add animation trace parity harness"`.

### Task 2: Paused-geometry parity scenarios

**Files:**
- Create: `server/static/tests/animations/parity/geometry-helpers.ts`
- Create: `server/static/tests/animations/parity/geometry.spec.ts`
- Goldens: `server/static/tests/animations/parity/goldens/geometry-*.json`

**Interfaces:**
- Produces: `sampleGeometry(page, scenario, opts): Promise<GeometrySamples>` — pauses every `document.getAnimations()` entry (covers both WAAPI animations **and** CSS transitions, which appear as `CSSTransition`) at fixed fractions of the effective duration and records each moving element's `getBoundingClientRect` + computed transform. `expectGeometryMatchesGolden(samples, name, tolerancePx)`.

- [ ] **Step 1: Write `geometry-helpers.ts`**

```typescript
import { Page, expect } from '@playwright/test';
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';

export interface GeometrySample {
  fraction: number;
  elements: Record<string, { x: number; y: number; w: number; h: number; transform: string }>;
}
export interface GeometrySamples { samples: GeometrySample[] }

const GOLDEN_DIR = join(__dirname, 'goldens');
const FRACTIONS = [0, 0.25, 0.5, 0.75, 1];

// Runs `trigger`, then as soon as animations exist, pauses ALL of them
// (WAAPI + CSSTransition both appear in document.getAnimations()), steps
// currentTime through fixed fractions of each animation's own duration,
// sampling identified elements' geometry at each step, then finishes all
// animations so the gate settles normally.
export async function sampleGeometry(
  page: Page,
  trigger: () => Promise<void>,
  opts: { selector: string },
): Promise<GeometrySamples> {
  await trigger();
  await page.waitForFunction(() => document.getAnimations().length > 0,
    undefined, { timeout: 10000 });
  const samples: GeometrySample[] = [];
  await page.evaluate(() => { for (const a of document.getAnimations()) a.pause(); });
  for (const fraction of FRACTIONS) {
    const sample = await page.evaluate(([frac, selector]) => {
      for (const a of document.getAnimations()) {
        const t = a.effect?.getComputedTiming();
        const total = (Number(t?.delay ?? 0)) + Number(t?.activeDuration ?? 0);
        a.currentTime = (frac as number) * total;
      }
      const out: Record<string, any> = {};
      const roots: (Document | ShadowRoot)[] = [document];
      const walk = (root: Document | ShadowRoot | Element) => {
        for (const el of Array.from((root as Element).querySelectorAll?.(selector as string)
            ?? (root as Document).querySelectorAll(selector as string))) {
          const key = el.tagName.toLowerCase() + (el.id ? `#${el.id}` : '');
          const r = (el as HTMLElement).getBoundingClientRect();
          out[key] = { x: r.x, y: r.y, w: r.width, h: r.height,
            transform: getComputedStyle(el as HTMLElement).transform };
        }
        for (const el of Array.from((root as any).querySelectorAll('*'))) {
          const sr = (el as Element & { shadowRoot: ShadowRoot | null }).shadowRoot;
          if (sr) walk(sr);
        }
      };
      for (const r of roots) walk(r as Document);
      return out;
    }, [fraction, opts.selector] as [number, string]);
    samples.push({ fraction, elements: sample });
  }
  await page.evaluate(() => {
    for (const a of document.getAnimations()) { try { a.finish(); } catch { a.cancel(); } }
  });
  return { samples };
}

export function expectGeometryMatchesGolden(
  actual: GeometrySamples, name: string, tolerancePx = 2,
): void {
  const goldenPath = join(GOLDEN_DIR, `${name}.json`);
  if (process.env.PARITY_RECORD === '1') {
    mkdirSync(dirname(goldenPath), { recursive: true });
    writeFileSync(goldenPath, JSON.stringify(actual, null, 2) + '\n');
    return;
  }
  if (!existsSync(goldenPath)) throw new Error(`missing golden ${name}`);
  const golden: GeometrySamples = JSON.parse(readFileSync(goldenPath, 'utf-8'));
  expect(actual.samples.length).toBe(golden.samples.length);
  for (let i = 0; i < golden.samples.length; i++) {
    const g = golden.samples[i]!, a = actual.samples[i]!;
    expect(Object.keys(a.elements).sort()).toEqual(Object.keys(g.elements).sort());
    for (const key of Object.keys(g.elements)) {
      const ge = g.elements[key]!, ae = a.elements[key]!;
      for (const dim of ['x', 'y', 'w', 'h'] as const) {
        expect(Math.abs(ae[dim] - ge[dim]),
          `${name}[${i}] ${key}.${dim}: ${ae[dim]} vs golden ${ge[dim]}`)
          .toBeLessThanOrEqual(tolerancePx);
      }
    }
  }
}
```

- [ ] **Step 2: Write `geometry.spec.ts`** — two scenarios: `debuganimations` card flight (selector `boardgame-card`) and `memory` reveal (selector `boardgame-card`), each `createOfflineGame` → `sampleGeometry(page, <same trigger as Task 1>, {selector: 'boardgame-card'})` → `expectGeometryMatchesGolden`. Fix the viewport in the spec (`test.use({ viewport: { width: 1280, height: 900 } })`) so rects are stable.
- [ ] **Step 3: Record goldens, verify comparison passes, verify teeth** (same procedure as Task 1 Steps 4–6; for teeth, temporarily perturb a keyframe in `boardgame-component.ts` FLIP path or a stack margin, expect failure, revert).
- [ ] **Step 4: Commit** — `"Add paused-geometry parity harness"`.

### Task 3: Harness critic gate + README

- [ ] **Step 1: Dispatch a critic sub-agent** with the harness diff and the question: "Name concrete stack-animation regressions this harness would NOT catch (timing/easing changes, interruption behavior, z-order, silhouette/trail behavior, companion flights, reduced-motion)." 
- [ ] **Step 2: Close actionable holes** — expected known gaps to address: (a) easing changes invisible at 5 fractions → geometry FRACTIONS already samples mid-flight where easing shows (0.25/0.75 differ under ease vs linear by >2px for typical flights; verify by teeth-test with easing swapped); (b) blackjack companion flights not geometry-sampled → add a `blackjack` geometry scenario if the critic confirms feasibility; (c) interruption (retrigger mid-flight) → add a debuganimations trace scenario clicking two moves back-to-back without waiting.
- [ ] **Step 3: Write `tests/animations/parity/README.md`** documenting: how to record, what's compared, tolerances, and the accepted residual blind spots (verbatim from critic findings that were consciously not closed, with reasons).
- [ ] **Step 4: Full-suite baseline** — run `npm run test:e2e` and `npm run test:unit`; both green. Commit `"Close parity harness blind spots"`. **Phase 0 gate complete.**

---

## Phase 1 — Classing migrations

### Task 4: Migrate `boardgame-fading-text` to `BoardgameAnimatableItem`

**Files:**
- Modify: `server/static/src/components/boardgame-fading-text.ts`
- Test: `server/static/tests/animations/parity/fading-text.spec.ts` (new)

**Interfaces:**
- Consumes: `BoardgameAnimatableItem.play(element, keyframes, timing?, opts?)`, `animationLengthMs()`, `finishAllAnimations()`.
- Produces: unchanged public API (`message`, `trigger`, `suppress`, `autoMessage`, `announce`, `animateFade()`); now fires `will-animate`/`animation-done` (declared change).

- [ ] **Step 1: Write the failing test** — `fading-text.spec.ts`: create a memory game, capture a trace around a scoring move (memory match increments score, which drives status-text → fading-text). Assert (a) a `play` event with detail `boardgame-fading-text` appears in the trace, (b) gate delta shows the fade was gated (gateCloses only after settle), (c) watchdog 0. Run: `npx playwright test tests/animations/parity/fading-text.spec.ts` — expected FAIL (no `boardgame-fading-text` play events exist; it's CSS-driven).
- [ ] **Step 2: Implement.** Replace the class body machinery:
  - `export class BoardgameFadingText extends BoardgameAnimatableItem` (import from `./boardgame-animatable-item.js`).
  - Delete: the `.animating #message` animation CSS, `@keyframes fadetext`, the reduced-motion 1ms block, `_animating` property, `_animationEnded`, `@animationend/@animationcancel` bindings, `_classes()`.
  - Keep `#container` visibility CSS but key it off a private `_visible` reactive property.
  - New `animateFade()`:

```typescript
animateFade(): void {
  this.finishAllAnimations();          // retrigger = finish prior fade (parity
  this._visible = true;                // with the old generation-counter reset)
  void this.updateComplete.then(() => {
    const message = this.renderRoot.querySelector('#message') as HTMLElement | null;
    if (!message || !this.isConnected) { this._visible = false; return; }
    const anim = this.play(message, [
      { opacity: 1, transform: 'scale(1.0)' },
      { opacity: 0, transform: 'scale(6.0)' },
    ], { easing: 'ease-out' });        // duration defaults to animationLengthMs()
    if (!anim) { this._visible = false; return; }
    anim.finished.catch(() => {}).finally(() => { this._visible = false; });
  });
}
```

  - Container class becomes `class="${this._visible ? 'animating' : ''}"` (keep the class name so external CSS hooks, if any, still match).
- [ ] **Step 3: Run the new spec** — expected PASS. Run `npx playwright test tests/animations/parity/` — trace goldens for memory WILL differ (new gated plays from fading-text). This is the declared change: regenerate ONLY the affected goldens with `PARITY_RECORD=1 npx playwright test tests/animations/parity/trace.spec.ts`, and verify the geometry goldens still pass UNREGENERATED. Diff the regenerated trace goldens in the commit and confirm the only delta is added `boardgame-fading-text` play/settle events and matching play/settle counters — any other delta is a stop-line.
- [ ] **Step 4: Evidence pack** — write `docs/superpowers/specs/evidence/2026-07-24-fading-text-gating.md`: before/after golden diff, the #714 checklist quote, reduced-motion note (kernel skip vs 1ms sprint).
- [ ] **Step 5: `npm run type-check && npm run type-check:strict && npm run test:e2e`** all green. Commit `"Route fading-text through the gated animation kernel"`.

### Task 5: `boardgame-status-text` extends `BoardgameAnimatableItem`

**Files:**
- Modify: `server/static/src/components/boardgame-status-text.ts:1-10`

- [ ] **Step 1:** Change `import { LitElement, html, css, nothing } from 'lit'` → `import { html, css, nothing } from 'lit'`; add `import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';`; `export class BoardgameStatusText extends BoardgameAnimatableItem`. No other change.
- [ ] **Step 2:** `npm run type-check && npm run type-check:strict` clean; full parity suite green with NO golden changes (status-text itself plays nothing yet). Commit `"Make status-text an animatable item"`.

### Task 6: Migrate `boardgame-game-outcome` arrival to gated `play()`

**Files:**
- Modify: `server/static/src/components/boardgame-game-outcome.ts`
- Test: extend `server/static/tests/renderer/` — read `tests/renderer/renderer-fixture.spec.ts` first and add the assertion where outcome display is already exercised; if no existing coverage, add `tests/animations/parity/game-outcome.spec.ts` driving a quick tictactoe/memory game to completion via admin moves.

- [ ] **Step 1: Write the failing test** — assert a `play` event with detail `boardgame-game-outcome` when the verdict appears, gated, watchdog 0.
- [ ] **Step 2: Implement** — `extends BoardgameAnimatableItem`; delete `animation: outcome-arrive 220ms ease-out both;`, `@keyframes outcome-arrive`, and the reduced-motion `animation: none` block. Add:

```typescript
override updated(changed: Map<PropertyKey, unknown>) {
  super.updated(changed);
  const revealed = this.finished && !this.animating;
  if (revealed && !this._arrivalPlayed) {
    this._arrivalPlayed = true;
    const outcome = this.renderRoot.querySelector('#outcome') as HTMLElement | null;
    if (outcome) {
      this.play(outcome, [
        { opacity: 0, transform: 'scale(0.96)' },
        { opacity: 1, transform: 'scale(1)' },
      ], { duration: 220, easing: 'ease-out', fill: 'backwards' });
    }
  }
  if (!revealed) this._arrivalPlayed = false;
}
private _arrivalPlayed = false;
```

- [ ] **Step 3:** Test passes; type-checks clean; parity goldens unaffected (outcome scenarios aren't in the Phase 0 goldens — the new spec owns this coverage). Commit `"Play game-outcome arrival through the gated kernel"`.

### Task 7: Route `boardgame-token` throb through ungated `play()`

**Files:**
- Modify: `server/static/src/components/boardgame-token.ts` (throb CSS at ~`:32-60`)
- Test: `server/static/tests/animations/parity/token-throb.spec.ts` (new) — debuganimations renders tokens (`boardgame-render-game-debuganimations.ts:483`).

- [ ] **Step 1: Failing test** — drive debuganimations to a highlighted-token state (read the debuganimations renderer to find what sets `highlighted`); assert via `page.evaluate` that the token host has a live infinite `Animation` (`document.getAnimations()` filtered to the token's shadow root) AND that the trace shows **no** `play`-hook gating events blocking the gate (gate closes while throb still runs).
- [ ] **Step 2: Implement** — remove the `animation-*: throb` CSS lines and `@keyframes throb` (keep the `--throb-color-*` custom properties and selector structure). In `updated()`, when the throbbing state class/property becomes active, start:

```typescript
this._throb?.cancel();
this._throb = this.play(target, [
  { filter: 'drop-shadow(0 0 0.25em var(--throb-color-to)) drop-shadow(0 0 0.25em var(--throb-color-to))' },
  { filter: 'drop-shadow(0 0 0.25em var(--throb-color-from)) drop-shadow(0 0 0.25em var(--throb-color-from))' },
], { duration: 1000, easing: 'ease-in-out', direction: 'alternate', iterations: Infinity },
   { gated: false, timing: 'immediate' });
```

with `this._throb?.cancel()` when the state clears (and in `disconnectedCallback`). Note: WAAPI keyframes can't use `var()` in all engines — resolve the two colors with `getComputedStyle(this).getPropertyValue(...)` at start time; re-start on state change so theme switches mid-throb are acceptable (same as today: CSS var changes mid-keyframe are also edge-y; verify parity by eye in Step 3's headed run).
- [ ] **Step 3:** Test passes; run `HEADED=1 npx playwright test tests/animations/parity/token-throb.spec.ts` once and visually confirm the throb; type-checks; full parity suite green (goldens untouched — throb is ungated and unhooked). Commit `"Route token throb through ungated play()"`. 
- [ ] **Step 4: Phase 1 gate** — dispatch regression-critic + harness-critic + fresh-verifier sub-agents per spec §Verification protocol. Resolve findings before Phase 2.

---

## Phase 2 — Discovery and gate topology

### Task 8: Extract `motion/animation-gate.ts` kernel (no behavior change)

**Files:**
- Create: `server/static/src/motion/animation-gate.ts`
- Create: `server/static/src/motion/animation-gate.test.ts`
- Modify: `server/static/src/components/boardgame-render-game.ts:316-738` (gate internals)

**Interfaces:**
- Produces:

```typescript
export interface AnimationGateCallbacks {
  onOpen(): void;                       // gate transitions closed -> open
  onAllDone(): void;                    // last participant settled (or watchdog)
  onWatchdog(pending: readonly string[], budgetMs: number): void;
  setTimer(cb: () => void, ms: number): unknown;   // injectable for tests
  clearTimer(handle: unknown): void;
  now(): number;
}
export class AnimationGate {
  constructor(cb: AnimationGateCallbacks,
    opts?: { floorMs?: number; marginMs?: number });
  open(cycleId: number): void;                    // replaces _resetAnimating bookkeeping
  willAnimate(ele: object, label: string, expectedSettleMs?: number): void;
  animationDone(ele: object): void;
  settleIfEmpty(cycleId?: number): void;          // replaces _nextStateIfNoAnimations
  readonly isOpen: boolean;
  readonly pendingCount: number;
}
```

- [ ] **Step 1: Write `animation-gate.test.ts`** (node:test, following an existing `src/motion/*.test.ts` for import style) covering: open→willAnimate→animationDone fires onAllDone once; duplicate done is a no-op; done for unknown ele ignored; second open cancels prior watchdog and resets; watchdog fires onWatchdog+onAllDone after floor when a participant never settles; a willAnimate with `expectedSettleMs` beyond the current deadline re-arms to `expectedSettleMs + marginMs`; a shorter one does not shrink the deadline; `settleIfEmpty` fires onAllDone only when no participants and matching cycleId. Use fake timers via the injected `setTimer/clearTimer/now`. HARNESS-CRITIC REQUIREMENT (gap 9): the `expectedSettleMs` extension tests are the ONLY coverage of the watchdog-deadline extension anywhere (no e2e scenario is long enough to exercise it), so they are mandatory, not optional — include a case proving a long declared cycle (stagger+duration+endDelay > floor) is NOT force-closed at the floor.
- [ ] **Step 2: Run** `npm run test:unit` — expected FAIL (module missing).
- [ ] **Step 3: Implement the kernel** by transplanting the logic verbatim from `boardgame-render-game.ts` `_resetAnimating`/`_armWatchdog`/`_componentWillAnimate`/`_componentAnimationDone`/`_nextStateIfNoAnimations`/`_notifyAnimationsDone` (lines ~629-738), parameterizing timers/now and replacing `animHooks`/event dispatch with the callbacks. Constants: `floorMs = 4000`, `marginMs = 1500` defaults (from `_WATCHDOG_FLOOR_MS`/`_WATCHDOG_MARGIN_MS`).
- [ ] **Step 4:** `npm run test:unit` green.
- [ ] **Step 5: Adopt in render-game** — replace the inlined fields (`_activeAnimations`, `_maxExpectedSettleMs`, `_watchdogDeadlineEpoch`, `_animationWatchdogTimer`, the two constants) with one `AnimationGate` instance whose callbacks preserve EXACTLY today's side effects: `onOpen` → `animHooks.record('gate-open')`, `isAnimating = true`, `_applyAnimatingToRenderer()`, `animating-changed` dispatch; `onAllDone` → the `_notifyAnimationsDone` body's effects (`gate-close` hook, `isAnimating = false`, `animating-changed`, `all-animations-done` with cycleId); `onWatchdog` → the existing `console.error` text and `animHooks.record('watchdog', pending.join(','))`. `willAnimate` label = `tag#id` derived exactly as today (`boardgame-render-game.ts:665-671`). Keep the DOM listeners; they now delegate into the gate. `disconnectedCallback` calls the gate's clear (add `dispose()` if needed — implement as `clearTimer` of any armed watchdog).
- [ ] **Step 6:** Full parity suite + `npm run test:e2e` — ALL goldens must pass UNCHANGED (this is pure refactor). Type-checks clean. Commit `"Extract animation gate kernel from render-game"`.

### Task 9: Ambient animatable-item registry

**Files:**
- Modify: `server/static/src/components/boardgame-animatable-item.ts` (connect/disconnect)
- Modify: `server/static/src/components/boardgame-render-game.ts` (provide registry; use at cycle start)
- Create: `server/static/src/motion/animatable-registry.ts` + `animatable-registry.test.ts`

**Interfaces:**
- Produces: `export class AnimatableRegistry { register(item): void; unregister(item): void; items(): readonly BoardgameAnimatableItem-like[] }` (typed via a minimal structural interface `{ finishAllAnimations(): void; animationContext: unknown }` to avoid an import cycle). Providers expose it as a property `animatableRegistry` discovered by the same parent/slot walk as `_ambientAnimationContext()` (`boardgame-animatable-item.ts:170-186`).

- [ ] **Step 1: Unit test** — registry add/remove/idempotence; walk-up discovery is covered by the e2e step below (jsdom-free node:test cannot exercise shadow DOM).
- [ ] **Step 2: Implement registry module**; in `BoardgameAnimatableItem.connectedCallback`/`disconnectedCallback` (add overrides; call `super`), walk up (factor the walk out of `_ambientAnimationContext` into a shared private helper `_ambientLookup(propName)`) to find `animatableRegistry` and register/unregister. Cache the found registry on the instance for symmetric unregistration even after detach.
- [ ] **Step 3: render-game provides** `animatableRegistry = new AnimatableRegistry()` and, in `_resetAnimating` (cycle start), iterates `items()` calling `finishAllAnimations()` and installing `this.animationContext` — replacing the component-only paths (`_clearAllAnimatingComponents` stays; the animator's own iteration remains authoritative for stack components — registry `finishAllAnimations` is idempotent for already-settled items).
- [ ] **Step 4: E2E test** — `tests/animations/parity/registry.spec.ts`: in pig, assert via `page.evaluate` that the standalone die is registered (reach `boardgame-render-game` element, check its registry contains an element with tag `boardgame-die`), and that after two rapid rolls the first roll's animation was finished (no overlapping die animations: `document.getAnimations()` count on the die ≤ 1 mid-second-roll).
- [ ] **Step 5:** Parity goldens unchanged (finishing already-settled items is a no-op in steady state; the rapid-retrigger path adds no hooks). Full suites green. Commit `"Add ambient animatable registry"`. Evidence pack `evidence/2026-07-24-registry-reset.md` if any golden shifted (stop-line otherwise).

### Task 10: Player-info animations join the gate

**Files:**
- Modify: `server/static/src/components/boardgame-game-view.ts` (~`:610` area) — pipe roster subtree gate events into render-game's gate.
- Modify: `server/static/src/components/boardgame-render-game.ts` — expose `gateWillAnimate(e: CustomEvent)` / `gateAnimationDone(e: CustomEvent)` thin delegates to its `AnimationGate`.
- Test: `server/static/tests/animations/parity/player-info-gate.spec.ts`

- [ ] **Step 1: Evidence pack FIRST** (`evidence/2026-07-24-player-info-ungated.md`): a spec snippet demonstrating the desync — in memory (which shows per-player status-text score), capture a trace on a scoring move and record that fading-text `play` hooks now fire (post-Task 4) from the roster subtree while `gateDelta.gateCloses` arrives without waiting for them (compare settle timestamps vs gate-close index in the log). Quote #714's checklist item. This justifies the declared change.
- [ ] **Step 2: Failing test** — assert that on a scoring move, gate-close appears AFTER the roster fading-text settle in the hook log. HARNESS-CRITIC REQUIREMENT (gap 3): this test is the roster gating deliverable's only parity witness — it must assert both directions: (a) roster animations now hold the gate (close comes after their settle), and (b) a roster animation outside any open cycle does NOT reopen or wedge the gate (drive a roster-only animation with no board cycle and assert gate counters unchanged and no watchdog).
- [ ] **Step 3: Implement** — in `boardgame-game-view.ts` where the roster element is rendered (find `<boardgame-player-roster` in its template), add `@will-animate=${this._rosterWillAnimate} @animation-done=${this._rosterAnimationDone}` handlers that call `this._renderGame?.gateWillAnimate(e)` / `gateAnimationDone(e)` (add a `@query` for the render-game element if absent; guard: only forward while the gate is open — otherwise a roster animation outside any cycle (e.g. hover-triggered) must NOT reopen/queue: `if (!renderGame?.isAnimating) return;` for `will-animate`; always forward `animation-done` so a participant admitted at open can settle).
- [ ] **Step 4:** Test passes. Regenerate ONLY trace goldens whose diff shows added roster participants; geometry goldens untouched. Full suites + type-checks green. Commit `"Gate player-info animations with the board"`. 
- [ ] **Step 5: Phase 2 gate** — critics + fresh verifier per protocol.

---

## Phase 3 — Retire legacy stack transitions

### Task 11: `layoutTransform` setter on `BoardgameComponent`

**Files:**
- Modify: `server/static/src/components/boardgame-component.ts`
- Create: `server/static/src/components/boardgame-component.layout.test.ts` is NOT possible (DOM) — coverage comes from the parity geometry suite; add a focused e2e `tests/animations/parity/layout-transform.spec.ts` only if debuganimations has a messy/pile/fan stack whose membership change re-lays-out survivors (read the renderer; it does — pile/fan stacks).

**Interfaces:**
- Produces: `set layoutTransform(value: string)` / `get layoutTransform(): string` on `BoardgameComponent`. Semantics: setting a *different* value snaps `this.style.transform` to the new value and, when not suppressed (`noAnimate` false, element connected, prior computed transform differs), plays a gated host animation from the previous **computed** transform (mid-flight retarget parity with CSS transitions) to the new value with `duration: animationLengthMs()`, `easing: 'ease-in-out'` — exactly the retired CSS `transition: transform var(--animation-length) ease-in-out`.

- [ ] **Step 1: Implement**

```typescript
private _layoutTransform = '';

get layoutTransform(): string { return this._layoutTransform; }

set layoutTransform(value: string) {
  if (value === this._layoutTransform) return;
  const before = this.isConnected ? getComputedStyle(this).transform : 'none';
  this._layoutTransform = value;
  this.style.transform = value;
  if (this.noAnimate || !this.isConnected) return;
  const after = getComputedStyle(this).transform;
  if (before === after) return;
  this.play(this, [{ transform: before }, { transform: after }],
    { easing: 'ease-in-out' });   // duration defaults to animationLengthMs()
}
```

Also: the animator writes `component.style.transform` directly during FLIP (`boardgame-component-animator.ts:1195` region is the STACK's write; the animator's own writes are at e.g. `:983-984`, `:1318`, `:1356`) — those must keep working: the setter only mediates STACK layout writes; direct `style.transform` writes by the animator bypass it by design. Sync `_layoutTransform` cache in `updated()` is NOT needed — the stack is the only `layoutTransform` caller.
- [ ] **Step 2:** Type-checks clean. No behavior change yet (nothing calls it). Commit `"Add layoutTransform self-animating setter"`.

### Task 12: Stack uses the setter; delete the CSS transition path

**Files:**
- Modify: `server/static/src/components/boardgame-component-stack.ts` — `_updateComponentClasses` (`:1181-1195`), `_fanComponents` (`:1282`), transition CSS (`:70-78`), the `noAnimate`/`.no-animate` container plumbing (find `noAnimate` accessors in this file).
- Modify: `server/static/src/components/boardgame-component-animator.ts` — collection-level `noAnimate` toggles (`:646`, `:1200`) IF they only drive the deleted container class (verify: read the stack's `noAnimate` setter first; component-level `noAnimate` at `:647`, `:649`, `:1205` STAYS — it suppresses the new setter's self-play during measurement).

- [ ] **Step 1: Refactor transform assembly** — in `_updateComponentClasses`, accumulate `transformPieces` but do NOT assign `component.style.transform` at `:1195`; for `fan` layout, fold `_fanComponents`' per-component `rotationTransformation + translateTransformation` into the same pieces array (restructure `_fanComponents` to RETURN pieces per index instead of appending to style). Single final write per component: `component.layoutTransform = pieces.join(' ')`.
- [ ] **Step 2: Delete the transition CSS** (`:70-78` — both the `transition:` rule and the `.no-animate` unset rule) and the stack's container-class `no-animate` toggling; remove the animator's collection-level toggles only if Step 2's read confirmed they exist solely for that class.
- [ ] **Step 3: Parity verdict** — run the FULL parity suite (trace + geometry) with goldens UNREGENERATED. Expected: all green. The hook trace WILL show new `play`/`settle` events for layout tweaks that previously rode CSS transitions (messy/pile/fan re-layouts inside scenarios). If trace goldens fail on exactly and only added `boardgame-card`/`boardgame-token` play+settle pairs with matching counter deltas AND all geometry goldens pass untouched: this is the declared mechanism swap; regenerate trace goldens, document the diff in `evidence/2026-07-24-stack-transition-cutover.md`. ANY geometry deviation > tolerance is a stop-line: fix, don't regenerate.
- [ ] **Step 4:** `npm run test:e2e` full (companion `waapi-companion.spec.ts` especially — Table/Hand flights), `test:unit`, both type-checks. Commit `"Retire legacy stack CSS transitions"`.
- [ ] **Step 5: Phase 3 gate** — critics (regression critic explicitly prompted: "compare interruption/retarget behavior of CSS transitions vs the setter's computed-from capture; name any scenario where they diverge") + fresh verifier.

---

## Task 13: Close-out

- [ ] **Step 1: Docs** — update `server/static/src/ARCHITECTURE.md` (animation section: all game-semantic elements are `BoardgameAnimatableItem`; gate kernel; registry; layoutTransform) and `docs/animation-effects.md` if it references the CSS transition path. Update issue-relevant text in `TUTORIAL.md` only if it mentions status-text/fading-text animation mechanics.
- [ ] **Step 2: Full verification sweep** — from `server/static`: `npm run type-check && npm run type-check:strict && npm run test:unit && npm run test:e2e`; from worktree root: `GOWORK=off go build ./... && GOWORK=off go test ./server/...`. All green, output captured.
- [ ] **Step 3: Final review** — superpowers:requesting-code-review against the whole branch diff; resolve findings.
- [ ] **Step 4:** Update GitHub issues #714/#713 comment drafts (do not post without user) summarizing what's done. Commit docs `"Document unified animatable architecture"`. Report: branch state, evidence packs, golden diffs, residual blind spots.
