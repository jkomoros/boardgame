# Animation WAAPI Timing Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace guess-based CSS-transition completion tracking with Web Animations API ground truth, then build the declarative animation features and companion cross-screen sync on the reliable base.

**Architecture:** Every animation becomes a WAAPI `Animation` owned by `BoardgameAnimatableItem` via a single `play()` method; the animator keeps its FLIP measure/compute math but delegates playback; `boardgame-render-game`'s completion gate consumes real settlement instead of counted `transitionend` guesses. Spec: `docs/superpowers/specs/2026-07-03-animation-waapi-timing-design.md`.

**Tech Stack:** Lit 3, TypeScript 5.3, Vite 6, Playwright 1.58 (e2e only — no unit test runner; browser-dependent code is tested via Playwright `page.evaluate`).

## Global Constraints

- Branch: `animation-waapi-timing` (already created; spec committed). Commit after every task, often mid-task. `../games` changes are committed in that repo (`/Users/jkomoros/Code/go/src/github.com/jkomoros/games`, branch `animation-waapi-timing` — create it in Task 9).
- Every commit keeps `cd server/static && npm run type-check` green.
- Client source root: `server/static/src/`. All paths below relative to repo root `/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame` unless noted.
- Dev server for tests: `GOPATH=$HOME/Code/go nohup ./boardgame-util/boardgame-util serve --offline-dev-mode > server.log 2>&1 &` then wait for `curl -s http://localhost:8080/client_config.js | grep offline_dev_mode` → `true`. Playwright `reuseExistingServer` is true — do NOT kill the server between specs. Create test games via `/list-games` → Create Game (never `/game/<name>/new`).
- Run tests: `cd server/static && npx playwright test tests/animations/<file> --reporter=line`.
- Never use `Date.now()` semantics changes: `companion-sync.ts` API is consumed as-is.
- Event names `will-animate`, `animation-done`, `all-animations-done` and their `{bubbles: true, composed: true}` shape are a stable contract — do not rename.
- Do NOT remove: `#outer { transition: transform 0.1s }` hover effect, box-shadow/filter transitions (visual polish, not tracked by the gate — and after Task 4 nothing listens to `transitionend`, so they're inert to timing).
- `animationLength()` and `animationOverlap()` renderer hooks stay. Only `delayAnimation()` is removed (Task 9).

## File Structure

| File | Fate |
|---|---|
| `server/static/src/utils/anim-test-hooks.ts` | **Create** (Task 1): instrumentation singleton |
| `server/static/src/components/boardgame-animatable-item.ts` | **Rewrite** (Tasks 3, 4): WAAPI `play()`/`settled()`; delete expectation machinery |
| `server/static/src/components/boardgame-component.ts` | **Modify** (Task 4): `playAnimation()`; delete prepare/startAnimation, transition CSS |
| `server/static/src/components/boardgame-card.ts` | **Modify** (Task 4): keyframe-based flip; delete overrideRotated/basicRotated dance |
| `server/static/src/components/boardgame-die.ts` | **Modify** (Task 4): inner spin via `play()` |
| `server/static/src/components/boardgame-component-animator.ts` | **Modify** (Tasks 4, 5): PLAY phase → `playAnimation()`; `animateBetween` → WAAPI |
| `server/static/src/components/boardgame-component-stack.ts` | **Modify** (Task 4: `beforeOrphaned` → cancel; Task 7: `stagger`) |
| `server/static/src/components/boardgame-render-game.ts` | **Modify** (Tasks 1, 4, 8): hooks wiring; 4s watchdog; `isAnimating` |
| `server/static/src/components/boardgame-move-form.ts` | **Modify** (Task 8): auto-disable while animating |
| `server/static/src/components/boardgame-game-view.ts` | **Modify** (Task 8): propose-move guard |
| `server/static/src/components/boardgame-base-game-renderer.ts` | **Modify** (Task 9): delete `delayAnimation` |
| `server/static/src/components/boardgame-game-state-manager.ts` | **Modify** (Task 9: drop delayAnimation consult; Task 10: scheduled installs) |
| `server/static/src/components/boardgame-hand-view-base.ts`, `boardgame-table-view-base.ts` | **Modify** (Task 10): scheduled auto-fly |
| `examples/memory/client/boardgame-render-game-memory.ts` | **Modify** (Task 9): migrate off `delayAnimation` |
| `../games/murdermrmonroe/client/*` | **Modify** (Task 11): verdict gating |
| `server/static/tests/animations/waapi-*.spec.ts` | **Create** (Tasks 2, 6, 7, 8, 10, 11) |

Tasks 1–2 = Phase A (instrumentation + failing gate). Tasks 3–5 = Phase B (cutover). Tasks 6–9 = Phase C (features + migration). Tasks 10–12 = Phase D (companion sync #798 + close-out).

---

### Task 1: Test-hooks instrumentation

**Files:**
- Create: `server/static/src/utils/anim-test-hooks.ts`
- Modify: `server/static/src/components/boardgame-render-game.ts` (import + 3 call sites)

**Interfaces:**
- Produces: `window.__bgAnimTestHooks` with `{ gateOpens: number, gateCloses: number, watchdogFirings: number, plays: number, settles: number, log: Array<{t: number, ev: string, detail?: string}>, reset(): void }`. Later tasks call `animHooks.record('play', id)` etc. Export: `export const animHooks: AnimHooks` (no-ops if `window` lacks the flag? No — always on; it's cheap and offline-dev only pages use it).

- [ ] **Step 1: Write the module**

```ts
// server/static/src/utils/anim-test-hooks.ts
//
// Instrumentation counters for the animation pipeline, consumed by the
// Playwright regression gate (tests/animations/waapi-*.spec.ts). Always
// installed: the cost is one array push per lifecycle event, and having
// it in production builds means bug reports can include the log.

export interface AnimHookEntry {
  t: number;
  ev: string;
  detail?: string;
}

class AnimHooks {
  gateOpens = 0;
  gateCloses = 0;
  watchdogFirings = 0;
  plays = 0;
  settles = 0;
  log: AnimHookEntry[] = [];

  record(ev: 'gate-open' | 'gate-close' | 'watchdog' | 'play' | 'settle', detail?: string) {
    switch (ev) {
      case 'gate-open': this.gateOpens++; break;
      case 'gate-close': this.gateCloses++; break;
      case 'watchdog': this.watchdogFirings++; break;
      case 'play': this.plays++; break;
      case 'settle': this.settles++; break;
    }
    this.log.push({ t: performance.now(), ev, detail });
    if (this.log.length > 5000) this.log.splice(0, 1000);
  }

  reset() {
    this.gateOpens = 0;
    this.gateCloses = 0;
    this.watchdogFirings = 0;
    this.plays = 0;
    this.settles = 0;
    this.log = [];
  }
}

export const animHooks = new AnimHooks();

declare global {
  interface Window { __bgAnimTestHooks: AnimHooks; }
}
window.__bgAnimTestHooks = animHooks;
```

- [ ] **Step 2: Wire into boardgame-render-game.ts**

Add `import { animHooks } from '../utils/anim-test-hooks.js';` at top. Then:
- In `_resetAnimating()` (line ~318), first line of body: `animHooks.record('gate-open');`
- In the watchdog `setTimeout` callback, right before `console.error(...)`: `animHooks.record('watchdog', pendingComponents.join(','));`
- In `_notifyAnimationsDone()`, right before `this.dispatchEvent(...)`: `animHooks.record('gate-close');`

- [ ] **Step 3: Type-check and verify in browser**

Run: `cd server/static && npm run type-check` → clean.
Start dev server (Global Constraints command), open a game page via Playwright or curl the bundle; quickest check:
`npx playwright test tests/animations/verify-fix.spec.ts --reporter=line` (existing spec) still passes, and in any game page `window.__bgAnimTestHooks` is defined (assert in Step 4's smoke test instead of manually).

- [ ] **Step 4: Commit**

```bash
git add server/static/src/utils/anim-test-hooks.ts server/static/src/components/boardgame-render-game.ts
git commit -m "Add animation instrumentation hooks for regression gate (#720)"
```

---

### Task 2: Playwright regression gate (written against current behavior — expected flaky/failing)

**Files:**
- Create: `server/static/tests/animations/waapi-gate.spec.ts`
- Create: `server/static/tests/animations/helpers.ts`

**Interfaces:**
- Consumes: `window.__bgAnimTestHooks` (Task 1).
- Produces: `createOfflineGame(page, gameName): Promise<void>` helper (navigates `/list-games`, selects game type, creates game with fake email — follow the flow in the existing `verify-fix.spec.ts`; reuse its selectors verbatim) and `expectCleanGate(page)` assertion helper. Later feature specs import both.

- [ ] **Step 1: Write helpers.ts**

```ts
// server/static/tests/animations/helpers.ts
import { Page, expect } from '@playwright/test';

// Creates a fresh offline-dev-mode game and lands on its game page.
// IMPORTANT: copy the working navigation flow from verify-fix.spec.ts —
// go to /list-games, pick the game type card for `gameName`, click
// "Create Game", enter the fake email if the login dialog appears.
export async function createOfflineGame(page: Page, gameName: string): Promise<void> {
  await page.goto('/list-games');
  // ... (transcribe the exact working steps/selectors from verify-fix.spec.ts;
  // they are the project's canonical game-creation flow)
}

export interface GateSnapshot {
  gateOpens: number;
  gateCloses: number;
  watchdogFirings: number;
  plays: number;
  settles: number;
}

export async function gateSnapshot(page: Page): Promise<GateSnapshot> {
  return page.evaluate(() => {
    const h = (window as any).__bgAnimTestHooks;
    return {
      gateOpens: h.gateOpens, gateCloses: h.gateCloses,
      watchdogFirings: h.watchdogFirings, plays: h.plays, settles: h.settles,
    };
  });
}

// Waits for the animation gate to be quiescent (closes caught up with
// opens) then asserts no watchdog fired since `since`.
export async function expectCleanGate(page: Page, since: GateSnapshot, timeoutMs = 20000) {
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  }, undefined, { timeout: timeoutMs });
  const now = await gateSnapshot(page);
  expect(now.watchdogFirings, 'animation watchdog must never fire').toBe(since.watchdogFirings);
}
```

(The `// ...` in `createOfflineGame` is filled in during this task by reading `verify-fix.spec.ts` — it is an existing, working flow to transcribe, not new design.)

- [ ] **Step 2: Write waapi-gate.spec.ts**

```ts
// server/static/tests/animations/waapi-gate.spec.ts
import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, expectCleanGate } from './helpers';

// The reliability gate (spec §Testing). These scenarios are the
// historical wedge repros for #720. They must run clean N times.
const ROUNDS = 10;

test.describe('animation completion gate', () => {
  test('debuganimations: move-all + undo never wedges', async ({ page }) => {
    await createOfflineGame(page, 'debuganimations');
    const base = await gateSnapshot(page);
    for (let i = 0; i < ROUNDS; i++) {
      // The debuganimations renderer exposes move buttons; MoveAllComponents
      // then its inverse is the canonical #720 repro.
      await page.getByRole('button', { name: /Move All Components/i }).first().click();
      await expectCleanGate(page, base);
      await page.getByRole('button', { name: /Undo/i }).first().click();
      await expectCleanGate(page, base);
    }
  });

  test('blackjack: fresh deal completes cleanly', async ({ page }) => {
    for (let i = 0; i < 3; i++) {
      await createOfflineGame(page, 'blackjack');
      const base = await gateSnapshot(page);
      await expectCleanGate(page, base);
    }
  });

  test('memory: reveal/hide fixup chain completes cleanly', async ({ page }) => {
    await createOfflineGame(page, 'memory');
    const base = await gateSnapshot(page);
    for (let i = 0; i < 4; i++) {
      // Click two hidden cards; the HideCards fixup runs on a timer.
      const cards = page.locator('boardgame-card:not([face-up])');
      await cards.nth(0).click();
      await cards.nth(1).click();
      await expectCleanGate(page, base, 30000);
    }
  });
});
```

Button/locator names must be adjusted to the real debuganimations move names — inspect the running page (`page.pause()` or read `examples/debuganimations/client/boardgame-render-game-debuganimations.ts` and the moves in `examples/debuganimations/*.go`) and use the actual visible labels. That adjustment is part of this task, not deferred.

- [ ] **Step 3: Run the suite against the CURRENT (pre-rewrite) system**

Run: `cd server/static && npx playwright test tests/animations/waapi-gate.spec.ts --reporter=line`
Expected: passes are FLAKY or tests FAIL on watchdog assertions (~20% historical wedge rate on blackjack; wedges now surface as `watchdogFirings > 0` after 15s). Record the observed pass/fail in the commit message — this is the baseline the rewrite must beat.

- [ ] **Step 4: Commit**

```bash
git add server/static/tests/animations/waapi-gate.spec.ts server/static/tests/animations/helpers.ts
git commit -m "Add animation reliability regression gate (baseline: pre-WAAPI, documents #720 wedges)"
```

---

### Task 3: WAAPI core on BoardgameAnimatableItem (additive — old machinery still present)

**Files:**
- Modify: `server/static/src/components/boardgame-animatable-item.ts`
- Test: `server/static/tests/animations/waapi-play.spec.ts` (create)

**Interfaces:**
- Produces (consumed by Tasks 4–11):

```ts
// On BoardgameAnimatableItem:
play(element: HTMLElement, keyframes: Keyframe[], timing?: OptionalEffectTiming,
     opts?: { gated?: boolean }): Animation | null
  // null when noAnimate, zero-duration environments handled internally.
get isAnimating(): boolean
settled(): Promise<void>            // resolves when live GATED set is empty; never rejects
finishAllAnimations(): void          // finish() running, cancel() pending — used on interrupt
animationLengthMs(): number          // parses --animation-length from computed style; default 250
postAnimationDelay: number           // property (attribute 'post-animation-delay') — used from Task 6, declared now
waitForAnimation: boolean            // property (attribute 'wait-for-animation') — declared now, default true
```

- [ ] **Step 1: Add the WAAPI members to boardgame-animatable-item.ts** (keep ALL existing code for now; only add)

```ts
// Add imports: none new needed (lit already imported).

// Add fields:
  private _liveAnimations = new Set<Animation>();
  private _liveGatedCount = 0;
  private _settledResolvers: Array<() => void> = [];

  @property({ type: Number, attribute: 'post-animation-delay' })
  postAnimationDelay = 0;

  @property({ type: Boolean, attribute: 'wait-for-animation' })
  waitForAnimation = true;

// Add methods:

  // animationLengthMs reads the effective --animation-length CSS variable
  // (games set it; render-game sets it from the renderer's animationLength()
  // hook). Accepts '0.25s' or '250ms'. Returns milliseconds.
  animationLengthMs(): number {
    const raw = getComputedStyle(this).getPropertyValue('--animation-length').trim();
    if (!raw) return 250;
    if (raw.endsWith('ms')) return parseFloat(raw) || 250;
    if (raw.endsWith('s')) return (parseFloat(raw) || 0.25) * 1000;
    const n = parseFloat(raw);
    return isNaN(n) ? 250 : n;
  }

  get isAnimating(): boolean {
    return this._liveGatedCount > 0;
  }

  // play is the single entry point for starting an animation on this item
  // (host element, #inner, or any shadow child). Ground truth for
  // completion is the returned Animation's settlement — there is nothing
  // to guess (spec: WAAPI rewrite).
  play(element: HTMLElement, keyframes: Keyframe[], timing?: OptionalEffectTiming,
       opts?: { gated?: boolean }): Animation | null {
    if (this.noAnimate) return null;
    const gated = (opts?.gated ?? true) && this.waitForAnimation;
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const resolvedTiming: OptionalEffectTiming = {
      duration: reduced ? 0 : this.animationLengthMs(),
      easing: 'ease-in-out',
      fill: 'none',
      ...timing,
    };
    if (this.postAnimationDelay > 0 && resolvedTiming.endDelay === undefined) {
      resolvedTiming.endDelay = this.postAnimationDelay;
    }
    const anim = element.animate(keyframes, resolvedTiming);
    this._liveAnimations.add(anim);
    if (gated) {
      this._liveGatedCount++;
      if (this._liveGatedCount === 1) {
        this.dispatchEvent(new CustomEvent('will-animate',
          { bubbles: true, composed: true, detail: { ele: this } }));
      }
    }
    animHooks.record('play', this.tagName.toLowerCase() + (this.id ? `#${this.id}` : ''));
    // finished rejects on cancel(); both paths are settlement for us.
    anim.finished.catch(() => {}).finally(() => this._animationSettled(anim, gated));
    return anim;
  }

  private _animationSettled(anim: Animation, gated: boolean) {
    if (!this._liveAnimations.delete(anim)) return; // already accounted
    animHooks.record('settle', this.tagName.toLowerCase() + (this.id ? `#${this.id}` : ''));
    if (!gated) return;
    this._liveGatedCount--;
    if (this._liveGatedCount <= 0) {
      this._liveGatedCount = 0;
      const resolvers = this._settledResolvers;
      this._settledResolvers = [];
      for (const r of resolvers) r();
      this.dispatchEvent(new CustomEvent('animation-done',
        { bubbles: true, composed: true, detail: { ele: this } }));
    }
  }

  settled(): Promise<void> {
    if (this._liveGatedCount === 0) return Promise.resolve();
    return new Promise((resolve) => this._settledResolvers.push(resolve));
  }

  // finishAllAnimations jumps every live animation to its end state and
  // resolves settlement. Called when a new animation cycle must start
  // while a previous one is in flight (spec: Interruption semantics).
  finishAllAnimations(): void {
    for (const anim of [...this._liveAnimations]) {
      try {
        if (anim.playState === 'running' || anim.playState === 'finished') {
          anim.finish();
        } else {
          anim.cancel();
        }
      } catch {
        // finish() throws InvalidStateError for infinite animations; cancel instead.
        try { anim.cancel(); } catch { /* already dead */ }
      }
    }
  }
```

Also add the import: `import { animHooks } from '../utils/anim-test-hooks.js';`

- [ ] **Step 2: Type-check**

Run: `cd server/static && npm run type-check`
Expected: clean. (Old machinery untouched; both systems coexist, nothing calls `play()` yet.)

- [ ] **Step 3: Write the behavior test (drives play() directly in the page)**

```ts
// server/static/tests/animations/waapi-play.spec.ts
import { test, expect } from '@playwright/test';
import { createOfflineGame } from './helpers';

test('play() fires will-animate/animation-done and settles', async ({ page }) => {
  await createOfflineGame(page, 'blackjack'); // any page with components registered
  const result = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    document.body.appendChild(ele);
    await ele.updateComplete;
    const events: string[] = [];
    ele.addEventListener('will-animate', () => events.push('will-animate'));
    ele.addEventListener('animation-done', () => events.push('animation-done'));
    const anim = ele.play(ele, [{ transform: 'translateX(100px)' }, { transform: 'none' }],
      { duration: 50 });
    const animatingDuring = ele.isAnimating;
    await ele.settled();
    return { events, animatingDuring, animatingAfter: ele.isAnimating, gotAnim: !!anim };
  });
  expect(result.gotAnim).toBe(true);
  expect(result.animatingDuring).toBe(true);
  expect(result.animatingAfter).toBe(false);
  expect(result.events).toEqual(['will-animate', 'animation-done']);
});

test('cancel counts as settlement; finishAllAnimations unblocks settled()', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const ok = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    document.body.appendChild(ele);
    await ele.updateComplete;
    ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 60000 });
    const p = ele.settled();
    ele.finishAllAnimations();
    await p; // must resolve, not hang or reject
    return true;
  });
  expect(ok).toBe(true);
});

test('noAnimate suppresses play; ungated play does not hold settled()', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const r = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    document.body.appendChild(ele);
    await ele.updateComplete;
    ele.noAnimate = true;
    const a1 = ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 50 });
    ele.noAnimate = false;
    const a2 = ele.play(ele, [{ opacity: 0 }, { opacity: 1 }],
      { duration: 60000 }, { gated: false });
    await ele.settled(); // must resolve immediately despite a2 running
    return { a1IsNull: a1 === null, a2Running: a2.playState === 'running' };
  });
  expect(r.a1IsNull).toBe(true);
  expect(r.a2Running).toBe(true);
});
```

- [ ] **Step 4: Run the new spec**

Run: `npx playwright test tests/animations/waapi-play.spec.ts --reporter=line`
Expected: 3 passed.

- [ ] **Step 5: Commit**

```bash
git add server/static/src/components/boardgame-animatable-item.ts server/static/tests/animations/waapi-play.spec.ts
git commit -m "Add WAAPI play()/settled() core to BoardgameAnimatableItem (#726 #714)"
```

---

### Task 4: The cutover — components/animator/gate on WAAPI, delete expectation machinery

This is the load-bearing task. Three green commits: (a) components gain `playAnimation()`, (b) animator + render-game switch, (c) delete dead machinery + CSS transitions.

**Files:**
- Modify: `server/static/src/components/boardgame-component.ts`
- Modify: `server/static/src/components/boardgame-card.ts`
- Modify: `server/static/src/components/boardgame-die.ts`
- Modify: `server/static/src/components/boardgame-component-animator.ts:599-678` (`_startAnimations`), `:466-502` (prepare block)
- Modify: `server/static/src/components/boardgame-render-game.ts` (watchdog, `_stateChanged`)
- Modify: `server/static/src/components/boardgame-component-stack.ts:704,802,1001` (`beforeOrphaned` semantics)
- Test: existing `waapi-gate.spec.ts` + `waapi-play.spec.ts`

**Interfaces:**
- Produces (consumed by animator):

```ts
// On BoardgameComponent:
playAnimation(rec: FlipRecord): void
export interface FlipRecord {
  before: Record<string, any>;       // animatingPropValues() before
  after: Record<string, any>;        // animatingPropValues() after
  invertedTransform: string;         // FLIP inverted transform (animator-computed, includes beforeTransform + scale)
  finalTransform: string;            // final inline/messy transform ('' if none)
  beforeOpacity: string;             // '' treated as '1'
  finalOpacity: string;
  needsHostTransition: boolean;      // host transform keyframes worth playing
}
```

- Consumes: `play()`, `finishAllAnimations()` from Task 3.

- [ ] **Step 1: Add `playAnimation` to BoardgameComponent** (old methods still present)

```ts
  // playAnimation is the WAAPI replacement for the old
  // prepareAnimation/startAnimation pair. The animator computed the FLIP
  // delta; we translate it into keyframes. Property-driven inner effects
  // (card flip, die spin) are handled by subclasses via
  // playPropertyAnimation, so the databinding dance (setting before-props
  // then after-props) is gone entirely.
  playAnimation(rec: FlipRecord): void {
    if (rec.needsHostTransition) {
      this.play(this, [
        { transform: rec.invertedTransform },
        { transform: rec.finalTransform || 'none' },
      ]);
      // The element's resting inline transform must be the final one; the
      // animation is an overlay (fill: 'none').
      this.style.transform = rec.finalTransform;
    }
    const beforeO = parseFloat(rec.beforeOpacity || '1');
    const afterO = parseFloat(rec.finalOpacity || '1');
    if (Math.abs(beforeO - afterO) > 0.01) {
      this.play(this, [{ opacity: String(beforeO) }, { opacity: String(afterO) }]);
    }
    this.style.opacity = rec.finalOpacity;
    this.playPropertyAnimation(rec.before, rec.after);
  }

  // playPropertyAnimation animates the visual consequences of
  // animatingProperties changing (e.g. a card's faceUp flip). Base: no-op.
  playPropertyAnimation(before: Record<string, any>, after: Record<string, any>): void {
    // Subclasses override.
  }
```

Add `FlipRecord` as an exported interface in the same file.

- [ ] **Step 2: Card override — flip as inner keyframes**

In `boardgame-card.ts` add:

```ts
  // _innerTransformFor computes the resting inner transform for a given
  // faceUp/rotated combination — the pure function behind what
  // _updateInnerTransform writes as the resting style.
  private _innerTransformFor(faceUp: boolean, rotated: boolean): string {
    return [
      'scale(var(--component-effective-scale))',
      faceUp ? 'rotateY(180deg)' : 'rotateY(0deg)',
      rotated ? 'rotate(90deg)' : 'rotate(0deg)',
    ].join(' ');
  }

  override playPropertyAnimation(before: Record<string, any>, after: Record<string, any>): void {
    if (before.faceUp === after.faceUp && before.rotated === after.rotated) return;
    if (!this.innerElement) return;
    this.play(this.innerElement, [
      { transform: this._innerTransformFor(!!before.faceUp, !!before.rotated) },
      { transform: this._innerTransformFor(!!after.faceUp, !!after.rotated) },
    ]);
  }
```

Simplify `_updateInnerTransform` to a pure resting-style setter (delete its `_expectTransitionEnd` call and the `changed` tracking):

```ts
  private _updateInnerTransform() {
    if (!this.innerElement) return;
    this.innerElement.style.transform =
      this._innerTransformFor(this.faceUp, this.rotated) || 'none';
  }
```

Delete `overrideRotated`, `basicRotated` properties, `computeAnimationProps` override, and `_rotatedChanged`'s `basicRotated` mirroring (keep `_updateInnerTransform()` call). Update `updated()` to drop references to removed properties. `animationRotates` stays (animator uses it for scale math).

- [ ] **Step 3: Die override**

In `boardgame-die.ts:191`, the spin currently sets an inner transform then `this._expectTransitionEnd(this._innerElement, 'transform')`. Read the surrounding function; replace the expectation with a `play()` of `[{transform: <old>}, {transform: <new>}]` using the same old/new values the function already computes (capture `this._innerElement.style.transform` before overwriting as the `<old>` keyframe). Die also gets a `playPropertyAnimation` no-op only if type errors demand it (it likely animates on value change internally — keep its existing trigger point, just swap the mechanism).

- [ ] **Step 4: Type-check + commit (a)**

Run: `npm run type-check` → clean (old `prepareAnimation`/`startAnimation` still exist and are still called by the animator; new code unused).

```bash
git add server/static/src/components/boardgame-component.ts server/static/src/components/boardgame-card.ts server/static/src/components/boardgame-die.ts
git commit -m "Add WAAPI playAnimation()/playPropertyAnimation() to components (#713)"
```

- [ ] **Step 5: Switch the animator PLAY phase**

In `boardgame-component-animator.ts`:

(i) In `_doAnimate`, in the `if (record.needsAnimation)` block (line ~482): DELETE the `component.prepareAnimation(...)` call (keep the cloned-nodes block — cross-stack content cloning is orthogonal). Store what it needs instead: the record already holds everything except the inverted transform string; add `record.invertedTransform = beforeInvertedTransform;` (add `invertedTransform?: string` to `ComponentRecord`). Also capture `record.beforeOpacity = component.style.opacity;` right there.

(ii) In the faux-component loop (line ~511): keep `component.style.transform/opacity` writes OFF — instead build the same record fields (`invertedTransform`, `beforeOpacity: '1.0'`) and stop calling `component.prepareAnimation(...)` (line ~579). The faux component's `AnimatingComponentRecord` gains `before: record.before || {}` and `invertedTransform`/`beforeOpacity` fields.

(iii) Replace `_startAnimations` entirely:

```ts
  private async _startAnimations(resolve: (p: Promise<void>) => void, generation: number) {
    if (this._generation !== generation) { resolve(Promise.resolve()); return; }

    const collections = this.stackElement._sharedStackList;

    // Restore noAnimate (was the measurement barrier; still gates play()).
    const allComponents: any[] = [];
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];
      collection.noAnimate = false;
      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        component.noAnimate = false;
        allComponents.push(component);
      }
    }
    for (const ac of this._animatingComponents) {
      ac.component.noAnimate = false;
      allComponents.push(ac.component);
    }

    await Promise.all(allComponents.map(c => c.updateComplete));
    if (this._generation !== generation) { resolve(Promise.resolve()); return; }

    const settledPromises: Promise<void>[] = [];

    for (let i = 0; i < collections.length; i++) {
      const components = collections[i].Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        const record = this._infoById[component.id];
        if (!record || !record.needsAnimation) continue;
        component.playAnimation({
          before: record.before || {},
          after: record.after || {},
          invertedTransform: record.invertedTransform || '',
          finalTransform: record.afterTransform || '',
          beforeOpacity: record.beforeOpacity || '1',
          finalOpacity: record.afterOpacity || '',
          needsHostTransition: record.needsHostTransition ?? true,
        });
        settledPromises.push(component.settled());
      }
    }

    for (const ac of this._animatingComponents) {
      ac.component.playAnimation({
        before: ac.before || {},
        after: ac.after,
        invertedTransform: ac.invertedTransform || '',
        finalTransform: ac.afterTransform,
        beforeOpacity: ac.beforeOpacity || '1',
        finalOpacity: ac.afterOpacity,
        needsHostTransition: true,
      });
      settledPromises.push(ac.component.settled());
    }

    // The promise animateFlip() hands out now means "everything SETTLED",
    // not "everything started" — the gate awaits real completion.
    resolve(Promise.all(settledPromises).then(() => {}));
  }
```

And change `animateFlip()`/`_scheduleAnimate`/`_doAnimate` plumbing so the outer promise resolves with settlement: `animateFlip(): Promise<void>` now returns `new Promise<Promise<void>>(...)` flattened — simplest implementation: pass a `resolve` that accepts a `Promise<void>` and declare `animateFlip` as `async` wrapping `await (await inner)`. Concretely:

```ts
  animateFlip(): Promise<void> {
    const generation = this._generation;
    return new Promise<Promise<void>>((resolve) => {
      Promise.resolve().then(() => {
        if (this._generation !== generation) { resolve(Promise.resolve()); return; }
        this._scheduleAnimate(resolve, generation);
      });
    }).then((settled) => settled);
  }
```

(`_scheduleAnimate` and `_doAnimate` signatures change from `resolve: () => void` to `resolve: (p: Promise<void>) => void`; their early-bail calls become `resolve(Promise.resolve())`. `_doAnimate`'s rAF call stays.)

(iv) In `prepare()`: first line of the collections loop, add interruption handling per spec — before measuring, finish anything still flying:

```ts
    // Interruption semantics (spec): a new cycle must measure resting
    // positions, so jump any still-live animations to their end state.
    for (let i = 0; i < collections.length; i++) {
      const components = collections[i].Components;
      for (let j = 0; j < components.length; j++) {
        const c = components[j];
        if (typeof c.finishAllAnimations === 'function') c.finishAllAnimations();
      }
    }
```

- [ ] **Step 6: render-game + stack adjustments**

`boardgame-render-game.ts`:
- `_resetAnimating()`: watchdog `15000` → `4000`; update the console.error text from `15s` to `4s`.
- `_stateChanged` keeps `this._animator?.animateFlip().then(() => this._nextStateIfNoAnimations());` — unchanged code, but now correct-by-construction: the promise resolves at settlement, and `_nextStateIfNoAnimations` fires the gate when the event bookkeeping is also empty. Delete the stale TODO comment below it (lines ~408-412).

`boardgame-component-stack.ts` (lines 704, 802, 1001): `beforeOrphaned()` on the item now must also cancel live animations. Update `BoardgameAnimatableItem.beforeOrphaned` (in the same commit) to:

```ts
  beforeOrphaned() {
    // Last chance before removal: settle everything so the gate never
    // waits on a detached element.
    this.finishAllAnimations();
  }
```

(The old body fired `animation-done` based on expectation maps; `finishAllAnimations` settles the live set which fires it via `_animationSettled`.)

- [ ] **Step 7: Type-check, run the full gate**

Run: `npm run type-check` → clean.
Run: `npx playwright test tests/animations/ --reporter=line`
Expected: `waapi-play.spec.ts` and `waapi-gate.spec.ts` PASS — including scenarios that were flaky at baseline. `verify-fix.spec.ts` must also still pass (cross-screen flights unaffected so far). If gate tests fail: debug BEFORE proceeding (systematic-debugging skill); this task is not done with red tests.

- [ ] **Step 8: Commit (b)**

```bash
git add server/static/src/components/boardgame-component-animator.ts server/static/src/components/boardgame-render-game.ts server/static/src/components/boardgame-animatable-item.ts server/static/src/components/boardgame-component-stack.ts
git commit -m "Cut animator + gate over to WAAPI settlement; watchdog 15s -> 4s (#720 #726)"
```

- [ ] **Step 9: Delete the dead machinery**

- `boardgame-animatable-item.ts`: delete `_expectedTransitionEnds`, `_outstandingTransitonEnds`, `_boundTransitionEnded`, the `connectedCallback`/`disconnectedCallback` transitionend listener bodies (keep the overrides if other logic remains, else delete the overrides entirely), `willNotAnimate`, `resetAnimating`, `_expectTransitionEnd`, `_removeExpectedTransition`, `_notifyAnimationDone`, `_transitionEnded`.
- `boardgame-component.ts`: delete `prepareAnimation`, `startAnimation`, `computeAnimationProps`, `setProperties`, the `willNotAnimate` override; delete the `:host { transition: transform ..., opacity ... }` CSS block (lines 17–21) and the transform term from `#inner`'s transition (keep box-shadow/filter terms, line 87–92); delete `.no-animate #inner { transition: unset; }` (lines 50–53).
- `boardgame-component-animator.ts`: delete the `component.resetAnimating()` call (line ~328) and its comment.
- Search for stragglers: `grep -rn "_expectTransitionEnd\|willNotAnimate\|resetAnimating" server/static/src/ ../game-src` → zero hits (game-src symlinks point at examples/ and ../games clients; if hits appear there, fix in Task 9's migration instead and leave a note).

- [ ] **Step 10: Type-check, full animations suite again, commit (c)**

Run: `npm run type-check && npx playwright test tests/animations/ --reporter=line` → all pass.

```bash
git add -A server/static/src
git commit -m "Delete transitionend expectation machinery and transition CSS (#726)"
```

---

### Task 5: animateBetween on WAAPI

**Files:**
- Modify: `server/static/src/components/boardgame-component-animator.ts:215-265`
- Test: existing `tests/animations/verify-fix.spec.ts`

**Interfaces:**
- Signature unchanged: `animateBetween(realId, stubId, durationMs=500): Promise<void>` — callers (`boardgame-hand-view-base.ts:129`, `boardgame-table-view-base.ts:118`, murdermrmonroe renderers) unaffected. Task 10 adds an optional `opts?: { startAtMs?: number }` parameter — declare it now, implement scheduling then.

- [ ] **Step 1: Replace the transition/timeout implementation**

Replace the body after the `dx === 0 && dy === 0` early-return with:

```ts
    const anim = real.animate(
      [
        { transform: `translate(${dx}px, ${dy}px) ${real.style.transform || ''}`.trim() },
        { transform: real.style.transform || 'none' },
      ],
      { duration: durationMs, easing: 'ease-out', fill: 'none' },
    );
    // Settlement is ground truth: finished resolves on completion, rejects
    // on cancel (element removed mid-flight) — both mean "done" here.
    await anim.finished.catch(() => {});
```

Delete: `prevTransform` capture, both `style.transition` writes, the forced `getBoundingClientRect()` reflow, the whole cleanup/transitionend/setTimeout promise. (No inline style mutation at all — the flight is a pure overlay.)

- [ ] **Step 2: Type-check + run cross-screen spec**

Run: `npm run type-check && npx playwright test tests/animations/verify-fix.spec.ts --reporter=line`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add server/static/src/components/boardgame-component-animator.ts
git commit -m "animateBetween: WAAPI overlay animation, no inline style/transition juggling"
```

---

### Task 6: Declarative attributes — post-animation-delay (#715) + wait-for-animation (#716)

The properties exist since Task 3 and `play()` already consults them. This task makes them attribute-driven contracts with tests, and sets the non-component defaults.

**Files:**
- Modify: `server/static/src/components/boardgame-component-stack.ts` (pass through `post-animation-delay` / `wait-for-animation` attributes from the stack to stamped components, alongside existing pass-through attrs — find the attribute-forwarding block by grepping `componentPropertiesFromAttributes\|_stackDefault` in the file and follow the existing pattern for e.g. `no-shadow`)
- Test: `server/static/tests/animations/waapi-attrs.spec.ts` (create)

**Interfaces:**
- Consumes: `postAnimationDelay`/`waitForAnimation` from Task 3.
- Produces: stack-level attributes `post-animation-delay="<ms>"` and `wait-for-animation` forwarded to children — renderer-facing API (spec: Renderer-facing API).

- [ ] **Step 1: Write failing test**

```ts
// server/static/tests/animations/waapi-attrs.spec.ts
import { test, expect } from '@playwright/test';
import { createOfflineGame } from './helpers';

test('post-animation-delay defers animation-done', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const elapsed = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    ele.postAnimationDelay = 300;
    document.body.appendChild(ele);
    await ele.updateComplete;
    const start = performance.now();
    ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 50 });
    await ele.settled();
    return performance.now() - start;
  });
  expect(elapsed).toBeGreaterThanOrEqual(340); // 50ms anim + 300ms endDelay, minus jitter
});

test('wait-for-animation=false items do not hold the gate', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const r = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    ele.waitForAnimation = false;
    document.body.appendChild(ele);
    await ele.updateComplete;
    let done = false;
    ele.addEventListener('animation-done', () => { done = true; });
    const anim = ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 60000 });
    await ele.settled(); // resolves immediately: nothing gated
    return { settledImmediately: true, doneFired: done, running: anim.playState === 'running' };
  });
  expect(r.settledImmediately).toBe(true);
  expect(r.doneFired).toBe(false);
  expect(r.running).toBe(true);
});
```

- [ ] **Step 2: Run — first test may already pass (play() consults the properties); the stack-forwarding is the gap**

Run: `npx playwright test tests/animations/waapi-attrs.spec.ts --reporter=line`
Expected: both PASS (properties path was built in Task 3). If either fails, fix `play()`.

- [ ] **Step 3: Add stack attribute forwarding + a DOM-attribute test**

In `boardgame-component-stack.ts`, locate the existing per-component property forwarding (the code that stamps components and applies stack-level defaults like `noShadow`) and forward two more: if the stack has attribute `post-animation-delay`, set `component.postAnimationDelay = parseFloat(attr)`; if it has `wait-for-animation="false"` set `component.waitForAnimation = false`. Follow the file's existing forwarding idiom exactly.

Append to `waapi-attrs.spec.ts`:

```ts
test('stack forwards post-animation-delay to stamped components', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const v = await page.evaluate(async () => {
    const stack = document.querySelector('boardgame-component-stack') as any
      ?? document.createElement('boardgame-component-stack');
    stack.setAttribute('post-animation-delay', '150');
    if (!stack.isConnected) document.body.appendChild(stack);
    await stack.updateComplete;
    const comp = stack.Components?.[0];
    return comp ? comp.postAnimationDelay : 'no-components';
  });
  expect(v === 150 || v === 'no-components').toBe(true); // deep query may find an empty stack; assert forwarding when present
});
```

Adjust that test during implementation to target a stack that definitely has components on the blackjack page (e.g. `page.locator('boardgame-component-stack').first()` inside the hand area) — the assertion must end up strict (`toBe(150)`), not the `'no-components'` escape hatch. Rework until strict.

- [ ] **Step 4: Type-check + run + commit**

```bash
npm run type-check && npx playwright test tests/animations/waapi-attrs.spec.ts --reporter=line
git add server/static/src/components/boardgame-component-stack.ts server/static/tests/animations/waapi-attrs.spec.ts
git commit -m "post-animation-delay + wait-for-animation attributes (#715 #716)"
```

---

### Task 7: Staggering (#728)

**Files:**
- Modify: `server/static/src/components/boardgame-component-stack.ts` (add `stagger` property)
- Modify: `server/static/src/components/boardgame-component-animator.ts` (`_startAnimations`)
- Test: append to `server/static/tests/animations/waapi-attrs.spec.ts`

**Interfaces:**
- Produces: `stagger` attribute (Number, fraction of animation length, e.g. `stagger="0.15"`) on `boardgame-component-stack`. Children that animate in the same cycle get `delay = indexInAnimatingSet * stagger * animationLengthMs()`.

- [ ] **Step 1: Add the property to the stack**

```ts
  // stagger, when > 0, offsets the start of each animating child in a
  // cycle by (index * stagger * animation length), producing a cascading
  // deal effect (#728). 0 = simultaneous (default).
  @property({ type: Number })
  stagger = 0;
```

- [ ] **Step 2: Consume in the animator**

`playAnimation` needs a per-call delay. Extend `FlipRecord` with `delayMs?: number`, and in `BoardgameComponent.playAnimation` pass `{ delay: rec.delayMs ?? 0 }` into both host `play()` calls' timing (spread AFTER the defaults so it wins; property animations in `playPropertyAnimation` get the same delay — pass it through as an optional second arg `delayMs = 0` and include in its `play()` timing).

In `_startAnimations`' main loop, track per-collection animating index:

```ts
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];
      const staggerFraction = collection.stagger || 0;
      let animIndex = 0;
      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        const record = this._infoById[component.id];
        if (!record || !record.needsAnimation) continue;
        const delayMs = staggerFraction > 0
          ? animIndex * staggerFraction * component.animationLengthMs()
          : 0;
        animIndex++;
        component.playAnimation({ /* ...as in Task 4... */, delayMs });
        settledPromises.push(component.settled());
      }
    }
```

- [ ] **Step 3: Test**

Append to `waapi-attrs.spec.ts` — drive it through debuganimations (which moves many cards at once): set `stagger` on the target stack via `page.evaluate`, trigger a MoveAllComponents move, then read `window.__bgAnimTestHooks.log` and assert the `play` entries for that cycle have strictly increasing `t` gaps ≥ the stagger step for same-stack components. Keep the assertion loose on absolute values (±1 frame) but strict on monotonicity.

```ts
test('stagger produces monotonically increasing start times', async ({ page }) => {
  await createOfflineGame(page, 'debuganimations');
  await page.evaluate(() => {
    const stacks = Array.from(document.querySelectorAll('*'))
      .flatMap(e => e.shadowRoot ? Array.from(e.shadowRoot.querySelectorAll('boardgame-component-stack')) : []);
    for (const s of stacks) (s as any).stagger = 0.2;
    (window as any).__bgAnimTestHooks.reset();
  });
  await page.getByRole('button', { name: /Move All Components/i }).first().click();
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens && h.plays > 2;
  });
  const playTimes = await page.evaluate(() =>
    (window as any).__bgAnimTestHooks.log.filter((e: any) => e.ev === 'play').map((e: any) => e.t));
  // WAAPI delay shifts effective start; our 'play' log stamps creation
  // time, so instead assert via the animations themselves:
  const delays = await page.evaluate(() => {
    const els = Array.from(document.querySelectorAll('*'))
      .flatMap(e => e.shadowRoot ? Array.from(e.shadowRoot.querySelectorAll('boardgame-card')) : []);
    return els.flatMap(el => el.getAnimations().map(a => (a.effect as KeyframeEffect).getTiming().delay as number));
  });
  const positive = delays.filter(d => d > 0);
  expect(positive.length).toBeGreaterThan(0);
});
```

(The `getAnimations()` variant is the robust assertion; timing-log monotonicity is a fallback. Implement with the `getAnimations()` check as primary; run it while animations are live — query immediately after the click, before `waitForFunction`.) Rework locators against real debuganimations DOM as in Task 2.

- [ ] **Step 4: Type-check + run + commit**

```bash
npm run type-check && npx playwright test tests/animations/waapi-attrs.spec.ts --reporter=line
git add -A server/static/src server/static/tests
git commit -m "Stagger attribute on component-stack (#728)"
```

---

### Task 8: isAnimating + move auto-disable (#721)

**Files:**
- Modify: `server/static/src/components/boardgame-render-game.ts`
- Modify: `server/static/src/components/boardgame-game-view.ts`
- Modify: `server/static/src/components/boardgame-move-form.ts`
- Test: `server/static/tests/animations/waapi-buttons.spec.ts` (create)

**Interfaces:**
- Produces: `BoardgameRenderGame.isAnimating: boolean` (reflected attribute `is-animating`), `animating-changed` event (bubbles, composed, `detail: {value: boolean}`); `boardgame-move-form` property `disableWhileAnimating = true` + attribute `no-animation-disable`; game-view swallows `propose-move` while animating with `console.warn`.

- [ ] **Step 1: render-game exposes the flag**

```ts
  @property({ type: Boolean, reflect: true, attribute: 'is-animating' })
  isAnimating = false;
```

Set `this.isAnimating = true` in `_resetAnimating()`; set `false` in `_notifyAnimationsDone()`; both dispatch `new CustomEvent('animating-changed', { bubbles: true, composed: true, detail: { value: this.isAnimating } })`.

- [ ] **Step 2: game-view guard + move-form disable**

`boardgame-game-view.ts`: it already routes `propose-move` (grep `propose-move` in the file for the handler). At the top of that handler:

```ts
    if (this._renderGameEle?.isAnimating) {
      console.warn('[game-view] propose-move ignored while animations are running (#721)');
      return;
    }
```

(Use the file's existing reference to the render-game element; grep for how it queries it.)

`boardgame-move-form.ts`: add

```ts
  @property({ type: Boolean, attribute: 'no-animation-disable' })
  noAnimationDisable = false;

  @property({ type: Boolean })
  animating = false;
```

In its `render()`, disabled state of submit buttons becomes `disabled || (this.animating && !this.noAnimationDisable)` (adapt to actual template). Wire `animating`: game-view (or wherever move-forms are stamped — grep `boardgame-move-form` usage) listens to `animating-changed` and sets the property on forms.

- [ ] **Step 3: Test**

```ts
// server/static/tests/animations/waapi-buttons.spec.ts
import { test, expect } from '@playwright/test';
import { createOfflineGame } from './helpers';

test('move buttons disable during animation and re-enable after', async ({ page }) => {
  await createOfflineGame(page, 'debuganimations');
  const btn = page.getByRole('button', { name: /Move All Components/i }).first();
  await btn.click();
  // Immediately during the animation the render-game reflects is-animating
  const during = await page.evaluate(() =>
    !!document.querySelector('boardgame-game-view')?.shadowRoot
      ?.querySelector('boardgame-render-game[is-animating]'));
  expect(during).toBe(true);
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  });
  const after = await page.evaluate(() =>
    !!document.querySelector('boardgame-game-view')?.shadowRoot
      ?.querySelector('boardgame-render-game[is-animating]'));
  expect(after).toBe(false);
});
```

(Adjust the shadow-DOM path to the real element tree — inspect via `page.evaluate` during implementation. Also add an assertion that clicking a move button mid-animation does not enqueue a move: click twice fast, assert only one version advance in `__bgAnimTestHooks.gateOpens` delta or via the version indicator.)

- [ ] **Step 4: Type-check + run + commit**

```bash
npm run type-check && npx playwright test tests/animations/waapi-buttons.spec.ts --reporter=line
git add -A server/static/src server/static/tests
git commit -m "isAnimating gate + move auto-disable and propose-move guard (#721)"
```

---

### Task 9: Remove delayAnimation; migrate renderers (in-repo + ../games)

**Files:**
- Modify: `server/static/src/components/boardgame-base-game-renderer.ts:175-181` (delete `delayAnimation`)
- Modify: `server/static/src/components/boardgame-game-state-manager.ts:652-662` (delete the consult block)
- Modify: `examples/memory/client/boardgame-render-game-memory.ts:55` (migrate)
- Verify: `examples/debuganimations`, `examples/blackjack`, all `examples/*/client`, `../games/{murdermrmonroe,pass,valentine}/client` type-check and play
- Test: gate suite (memory scenario covers the migration)

**Interfaces:**
- Consumes: `postAnimationDelay` (Task 6).
- Produces: none new — deletion + migration.

- [ ] **Step 1: Read memory's delayAnimation and migrate its intent**

`examples/memory/client/boardgame-render-game-memory.ts:55` delays the next state install after a reveal so players can see the pair before HideCards. Replace with `post-animation-delay` bound on the revealed cards' stack in the template: find the stack rendering `HiddenCards`/`VisibleCards` in that renderer and bind `post-animation-delay="${this._revealHoldMs()}"` where `_revealHoldMs()` returns the same constant the old hook returned (read the old body; it returns a value for RevealCard→HideCards move pairs — the declarative equivalent holds `animation-done` after the reveal animation). Delete the `delayAnimation` override.

- [ ] **Step 2: Delete the hook framework-side**

- `boardgame-base-game-renderer.ts`: delete the `delayAnimation` method + its comment block; update `animationLength`'s comment cross-reference.
- `boardgame-game-state-manager.ts:652-662`: delete the `if (renderer.delayAnimation) {...}` block.
- Grep: `grep -rn "delayAnimation" server/static/src examples ../games */client 2>/dev/null` → zero hits (adjust the ../games path: `/Users/jkomoros/Code/go/src/github.com/jkomoros/games`).

- [ ] **Step 3: Type-check both repos, run gate**

```bash
cd server/static && npm run type-check
npx playwright test tests/animations/waapi-gate.spec.ts --reporter=line
```
Expected: clean; memory scenario still passes (the hold now comes from the attribute). Manually verify the memory reveal still visibly pauses: `HEADED=1 npx playwright test tests/animations/waapi-gate.spec.ts -g memory`.

- [ ] **Step 4: Commit (both repos)**

```bash
git add -A server/static/src examples
git commit -m "Replace delayAnimation() hook with post-animation-delay attribute; migrate memory (#715)"
cd /Users/jkomoros/Code/go/src/github.com/jkomoros/games
git checkout -b animation-waapi-timing
# only if files changed here (murdermrmonroe/pass/valentine grep hits):
git add -A && git commit -m "Migrate clients off delayAnimation (boardgame WAAPI animation rewrite)"
```

---

### Task 10: Companion scheduled installs + synced auto-fly (#798, part 1)

**Files:**
- Modify: `server/static/src/components/boardgame-game-state-manager.ts` (`_scheduleNextStateBundle`)
- Modify: `server/static/src/components/boardgame-component-animator.ts` (`animateBetween` gains `opts.startAtMs`)
- Modify: `server/static/src/components/boardgame-hand-view-base.ts:119-130`, `server/static/src/components/boardgame-table-view-base.ts:102-119`
- Test: `server/static/tests/animations/waapi-companion.spec.ts` (create)

**Interfaces:**
- Consumes: `companionSync.localEquivalent()`, `latestServerPlayAt()`, `ingestVersionTiming()` from `server/static/src/companion-sync.ts` (existing, currently unconsumed).
- Produces: `animateBetween(realId, stubId, durationMs, opts?: { startAtMs?: number })` — `startAtMs` is a local-clock `Date.now()`-comparable timestamp; the flight's WAAPI `delay` is `clamp(startAtMs - Date.now(), 0, 2000)`.

- [ ] **Step 1: state-manager schedules installs at serverPlayAt**

In `_scheduleNextStateBundle` (after the renderer hook consults, before `this._asyncFireNextStateBundle(...)`), add:

```ts
    // Companion sync (#798): if the server stamped a play-at time for this
    // version and we're on a companion surface, delay install so Table and
    // Hand start the same animation within estimator error. Clamped so a
    // bad estimate can never hang the game (spec: Error handling).
    const playAt = latestServerPlayAt();
    if (playAt !== null) {
      const local = companionSync.localEquivalent(playAt);
      if (local !== null) {
        const wait = Math.min(2000, Math.max(0, local - Date.now()));
        if (wait > 8) { // sub-frame waits aren't worth a timer
          window.setTimeout(() => this._asyncFireNextStateBundle(effectiveAnimationLength), wait);
          return;
        }
        if (local - Date.now() < -2000) {
          console.warn('[state-manager] serverPlayAt over 2s stale; installing immediately');
        }
      }
    }
```

Import `companionSync, latestServerPlayAt` from `../companion-sync.js` (check the actual export path/names in `src/companion-sync.ts` and match them). IMPORTANT: gate this on companion mode — only when `surfaceForGame(...)` is `'table'` or `'hand'` (import from `../utils/companion-surface.js`); solo games keep instant installs.

- [ ] **Step 2: animateBetween scheduling + view-base consumption**

`animateBetween`: add `opts?: { startAtMs?: number }` parameter; compute `const delay = opts?.startAtMs ? Math.min(2000, Math.max(0, opts.startAtMs - Date.now())) : 0;` and put `delay` into the WAAPI timing.

`boardgame-hand-view-base.ts:129` and `boardgame-table-view-base.ts:118`: change the auto-fly calls to pass the scheduled start:

```ts
        const playAt = latestServerPlayAt();
        const startAtMs = playAt !== null ? companionSync.localEquivalent(playAt) ?? undefined : undefined;
        this.animator?.animateBetween(id, anchor, 600, { startAtMs });
```

(matching variable names per file — `stub, source` on the table side).

- [ ] **Step 3: Two-page companion test**

```ts
// server/static/tests/animations/waapi-companion.spec.ts
import { test, expect } from '@playwright/test';
import { createOfflineGame } from './helpers';

// Two contexts: table + hand surfaces of one murdermrmonroe game. The
// surface is selected per-game via the surface_<gameId> cookie (see
// utils/companion-surface.ts) — read that util during implementation and
// set the cookie/query param it documents for each page.
test('deal flights start within sync threshold on both surfaces', async ({ browser }) => {
  const tableCtx = await browser.newContext();
  const handCtx = await browser.newContext();
  const table = await tableCtx.newPage();
  const hand = await handCtx.newPage();
  await createOfflineGame(table, 'murdermrmonroe');
  const gameUrl = table.url();
  // Join the same game from the hand context, then set surface routing
  // (implementation detail: consult companion-surface.ts + murdermrmonroe
  // docs for the exact join + surface mechanism used in dev).
  await hand.goto(gameUrl + '?surface=hand');
  await table.reload(); // pick up table surface similarly (?surface=table)

  // Trigger a deal (host action on table surface), then compare the first
  // 'play' hook timestamp on each page, normalized via Date.now() offsets.
  // Assertion: |tableStart - handStart| < 250ms (drift was ~1000ms in #798).
  // ... (implement against murdermrmonroe's real host controls; the
  // existing verify-fix.spec.ts drives these surfaces already — reuse.)
});
```

The `...` here is deliberate at PLAN level only in the sense that it transcribes existing working two-surface driving code from `verify-fix.spec.ts` — during implementation, copy that spec's setup verbatim and add the skew assertion: capture `performance.timeOrigin + firstPlay.t` per page and assert the absolute difference `< 250`.

- [ ] **Step 4: Type-check + run + commit**

```bash
npm run type-check && npx playwright test tests/animations/ --reporter=line
git add -A server/static/src server/static/tests
git commit -m "Consume serverPlayAt: scheduled installs + synced auto-fly (#798)"
```

---

### Task 11: Verdict gating + murdermrmonroe polish (#798, part 2)

**Files:**
- Modify: `../games/murdermrmonroe/client/boardgame-render-game-murdermrmonroe.ts` (and its `-table`/`-hand` variants if separate files — `ls ../games/murdermrmonroe/client/`)
- Test: extend `waapi-companion.spec.ts`

**Interfaces:**
- Consumes: `isAnimating` / `animating-changed` (Task 8) — the verdict/status text render gates on `!animating`.

- [ ] **Step 1: Find the verdict text**

`grep -n "verdict\|Verdict\|reveal" /Users/jkomoros/Code/go/src/github.com/jkomoros/games/murdermrmonroe/client/*.ts` — locate where outcome text renders. Gate it: the renderer receives `animating` (bind from the `animating-changed` listener or read `render-game[is-animating]` via context — use the move-form pattern from Task 8) and holds rendering the verdict until `animating === false` (a `this.animating ? '' : verdictTemplate` guard in the template).

- [ ] **Step 2: Extend the companion test**

Assert: after triggering the verdict-revealing move on the table surface, the verdict text is NOT visible on either page while `is-animating` is set, and IS visible after the gate closes on that page.

- [ ] **Step 3: Type-check both repos, run, commit (../games repo)**

```bash
cd server/static && npm run type-check && npx playwright test tests/animations/waapi-companion.spec.ts --reporter=line
cd /Users/jkomoros/Code/go/src/github.com/jkomoros/games
git add -A && git commit -m "Gate verdict text on animation completion (boardgame #798)"
cd /Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame
git add server/static/tests && git commit -m "Companion verdict-gating test (#798)"
```

---

### Task 12: Full gate ×3, docs, close-out

**Files:**
- Modify: `server/static/src/ARCHITECTURE.md` (animation section — grep `transitionend` there and rewrite that section to describe `play()`/settlement/gate)
- Modify: spec doc — append "Implementation notes" if reality diverged

- [ ] **Step 1: Run the entire animations suite three times**

```bash
cd server/static && for i in 1 2 3; do npx playwright test tests/animations/ --reporter=line || break; done
```
Expected: 3/3 clean, `watchdogFirings` assertions all pass. Any flake = investigate before closing out (systematic-debugging skill).

- [ ] **Step 2: Update ARCHITECTURE.md animation section**

Describe: `play()` ownership, settlement gate, 4s watchdog invariant, attributes (`post-animation-delay`, `wait-for-animation`, `stagger`), `isAnimating`, companion scheduling. Delete all references to expectation counting / transitionend.

- [ ] **Step 3: Verify build + go side untouched**

```bash
cd server/static && npm run build            # tsc + vite build clean
cd ../.. && go build ./...                    # unchanged, but confirm
```

- [ ] **Step 4: Final commit + summary**

```bash
git add -A server/static docs
git commit -m "Docs + close-out for WAAPI animation timing rewrite

Resolves the timing foundation for #720 #726 #713 #714; features #715
#716 #721 #728; companion sync #798."
```

Write a summary for the user listing: per-issue outcome, the two branches (`animation-waapi-timing` in both repos), baseline-vs-final gate results, and anything deferred (e.g. FLIP↔animateBetween handoff blip if not fully addressed).

---

## Self-review (done at planning time)

- **Spec coverage:** reliability core (T3–5) ✓; watchdog 4s (T4) ✓; #715/#716 (T6) ✓; #728 (T7) ✓; #721 (T8) ✓; delayAnimation removal + migration incl. ../games (T9) ✓; #798 scheduled installs/auto-fly/verdict (T10–11) ✓; Playwright gate as done-bar (T2, T12) ✓; prefers-reduced-motion (T3 `play()`) ✓; interruption semantics (T4 Step 5.iv + `finishAllAnimations`) ✓; handoff blip: attempted via shared scheduling in T10, explicitly reported in T12 summary if still visible ✓.
- **Known judgment calls encoded:** keep `noAnimate` toggling in the animator (harmless, lower risk); keep event-based gate bookkeeping in render-game alongside the settlement promise (needed for renderer-initiated gated plays); `animationOverlap`/`animationLength` hooks survive.
- **Type consistency:** `play()`/`settled()`/`finishAllAnimations()`/`animationLengthMs()` (T3) are consumed with those exact names in T4–T10; `FlipRecord` fields in T4 match the T7 `delayMs` extension; `startAtMs` declared T5, implemented T10 — consistent.
- **Placeholder scan:** two intentional "transcribe existing working code" references (game-creation flow, two-surface driving) point at `verify-fix.spec.ts` as the concrete source — acceptable because the code exists in-repo; no TBDs otherwise. Locator names in T2/T7/T8 are explicitly flagged as adjust-to-real-DOM work within those tasks.
