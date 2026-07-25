# Triage: 4 failing waapi-attrs / waapi-buttons tests (2026-07-26)

## Question

Phase 3 gate: are the 4 failures pre-existing relative to the stack
CSS-transition cutover (57f85dc8, "Retire legacy stack CSS transitions"), or
caused by it? A prior "identical on baseline" claim had been asserted without
captured evidence; this document captures a real baseline experiment.

## The 4 failures (as reported by the fresh verifier, reproduced verbatim)

1. `waapi-attrs.spec.ts:37` "stack forwards post-animation-delay to stamped
   components" — `found=false` (no stack with stamped components).
2. `waapi-attrs.spec.ts:82` "stagger produces strictly increasing per-index
   animation delays" — TimeoutError clicking "To Hidden" (disabled: "not
   possible right now").
3. `waapi-buttons.spec.ts:30` "move buttons disable during animation and
   re-enable after" — `is-animating` attribute true BEFORE any click.
4. `waapi-buttons.spec.ts:62` "a move proposed while isAnimating is true is
   swallowed" — precondition `isAnimating` false when it should be true.

## Baseline experiment (no stash; copy-aside + `git show`)

Method: copied HEAD's `boardgame-component-stack.ts` and
`boardgame-component.ts` aside, overwrote both with their `f475ab4d`
(pre-cutover, post-layoutTransform-setter) contents, let the worktree vite
dev server HMR them in, ran both spec files, then restored the copies
(verified `git diff --stat` clean vs HEAD apart from the standing
`.database`/`boardgame-util` binary noise).

### HEAD run (7 tests, 1 worker)

```
4 failed
  waapi-attrs.spec.ts:37  stack forwards post-animation-delay ... expect(v.found).toBe(true) — Received: false
  waapi-attrs.spec.ts:82  stagger ... TimeoutError: locator.click: Timeout 10000ms exceeded ("To Hidden")
  waapi-buttons.spec.ts:30 move buttons disable ... line 41 expect(isAnimatingAttr).toBe(false) — Received: true
  waapi-buttons.spec.ts:62 swallowed ... line 120 precondition isAnimating .toBe(true) — Received: false
3 passed (32.2s)
```

### f475ab4d-baseline run (same 7 tests)

```
4 failed
  waapi-attrs.spec.ts:37  IDENTICAL failure (found=false)
  waapi-attrs.spec.ts:82  fails EARLIER: line 127 expect(setup.count).toBeGreaterThan(0) — Received: 0
                          (no stack had >1 stamped components at query time)
  waapi-buttons.spec.ts:30 IDENTICAL failure (is-animating true before click)
  waapi-buttons.spec.ts:62 IDENTICAL failure (precondition isAnimating false)
3 passed (21.7s)
```

3 of 4 failures byte-identical on the baseline; the 4th (stagger) fails on
both, at a slightly earlier line on baseline — same upstream symptom family
(stacks not yet stamped / initial cascade in flight). **Verdict: all four
are pre-existing relative to the 57f85dc8 cutover.**

## Root cause (live-probe evidence)

A temporary probe spec dumped page state immediately after
`createOfflineGame` and again 4s later:

```
blackjack immediate : stackCount=3, stampedStacks=0, isAnimating=false,
                      hooks {gateOpens:0, gateCloses:0, plays:0, settles:0}, pendingBundles=0
blackjack +4s       : stackCount=7, deck stamped 52, isAnimating=TRUE,
                      hooks {gateOpens:9, gateCloses:8, plays:635, settles:583}, pendingBundles=11
debuganimations imm.: stampedStacks=14, isAnimating=TRUE,
                      hooks {gateOpens:2, gateCloses:1, plays:198, settles:151}, pendingBundles=1,
                      "To Hidden" disabled ("...is not possible right now")
debuganimations +4s : isAnimating=false, hooks {3,3,277,277}, pendingBundles=0,
                      "To Hidden" enabled
```

Single first domino: `createOfflineGame` returns as soon as the renderer
mounts and the test hooks install, but the initial state bundle(s) —
blackjack's auto-deal cascade, debuganimations' gated initial animations
(player-info/roster/registry work from this branch: 49f398b4, 8bf1a757,
7a876a65) — arrive and animate for several seconds afterwards. All four
tests act during that window:

1. attrs:37 queries stacks before the deal bundle stamps any components →
   `found=false`.
2. attrs:82 applies `--animation-length: 3s` while the initial cascade is
   still running, stretching the cascade itself; buttons stay disabled
   (is-animating, #721) past the 10s click actionability timeout. On the
   baseline run the same race surfaced one step earlier (stacks not yet
   stamped → `setup.count=0`).
3. buttons:30 asserts a closed-gate baseline while the initial cascade is
   legitimately in flight → is-animating true.
4. buttons:62's `gateOpens > before.gateOpens` wait is satisfied by an
   initial-cascade cycle rather than the clicked move's own cycle; the
   "mid-animation" dispatch then lands after that spurious cycle closed →
   precondition isAnimating=false.

This is the DECLARED behavior side of the ledger: gating more animatables on
initial load (this branch's purpose) lengthened the initial cascade; the
tests' implicit "page is quiescent right after creation" assumption was the
defect. Per the triage rule, tests adapt with comments.

## Fix

Test-harness only, composed from existing primitives:

- `helpers.ts`: new `settleInitialLoad(page)` — waits for the first gate
  open (so all-zero counters can't trivially satisfy stability before the
  initial bundle arrives), then `waitForAnimationCounterStability(balance:
  'all')`, then `waitForClientQuiescence`.
- Each of the four failing tests calls it right after `createOfflineGame`,
  with a comment explaining the specific race it prevents.

No product code changed. No golden files touched.

## Verification

- `npx playwright test tests/animations/waapi-attrs.spec.ts
  tests/animations/waapi-buttons.spec.ts --reporter=line` → 7 passed (two
  consecutive runs: 1.0m, 59.2s).
- `npx playwright test tests/animations/parity/ --reporter=line` → 31
  passed (2.5m); `git status` shows no parity/golden changes.
- `npx tsc --noEmit` → exit 0.
