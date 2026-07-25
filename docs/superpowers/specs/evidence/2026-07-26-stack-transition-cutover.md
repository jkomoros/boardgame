# Stack transition cutover — evidence pack (Task 12)

**Change:** `boardgame-component-stack` stops relying on the ambient CSS
transition and hands each component its layout transform through the Task 11
`BoardgameComponent.layoutTransform` setter (self-animating via the gated
WAAPI `play()` kernel). The legacy transition CSS and the `noAnimate`
container-class (`.no-animate`) suppression path are deleted. Realizes #713.

Commits: see `Retire legacy stack CSS transitions`.

---

## 1. What changed in code

| File | Change |
| --- | --- |
| `boardgame-component-stack.ts` | `_updateComponentClasses` no longer writes `component.style.transform`; it assembles all pieces (messy rotate + pile translate/rotate + fan rotate/translateY) and does a **single** `component.layoutTransform = pieces.join(' ')` write per component. `_fanComponents()` → `_fanComponentPieces(): string[][]` returns per-index pieces instead of appending to `style.transform`. Deleted the `#container ::slotted([boardgame-component]) { transition: transform … , opacity … }` rule and the `#container.no-animate … { transition: unset }` rule. Deleted the stack's `_noAnimate` field / `noAnimate` accessor, the `shouldUpdate` `noAnimate` branch, and the `noAnimate` argument to `_classes()` / the render template. |
| `boardgame-component-animator.ts` | Removed the three **collection-level** `collection.noAnimate = true/false` assignments (measurement barrier set, restore, and abort restore). **Component-level** `component.noAnimate` toggles are unchanged — they remain load-bearing (they snap `play()` calls issued while the barrier is up). |
| `boardgame-component.ts` | Comment-only: documents the resolved design invariant at the setter's `noAnimate` gate. |
| `tests/animations/parity/geometry.spec.ts` (+golden) | New scenario `debuganimations: fan draw relayout curves` closing the fan/pile-relayout-during-cycle coverage gap. |

### Sole-purpose verification before each deletion

- **`.no-animate` container class / stack `noAnimate` accessor.** Grep across
  `src` + `tests` confirmed the stack's `noAnimate` is read only by (a) the
  setter's own `container.classList.toggle('no-animate', …)`, (b) `_classes()`
  at render, and (c) the `shouldUpdate` early-out. `noAnimate` was never a Lit
  `@property` on the stack, so its `shouldUpdate` branch was already
  unreachable dead code (a non-reactive setter can never appear in
  `changedProperties`). No external consumer. The only writer was the
  animator's three collection-level assignments. → all removable together.
- **Collection-level `collection.noAnimate` (animator).** Sole purpose was to
  toggle the `.no-animate` container class. With the class gone, the accessor
  is gone, so the assignments are removed. Component-level `component.noAnimate`
  is a **separate** barrier (set in the same loops) and stays.
- **Opacity transition.** The deleted rule also transitioned `opacity`. The
  live `#container` slotted components never have their `opacity` animated as a
  *layout* effect (opacity writes at `stack:setUnknownAnimationState` / the
  motion-carrier factories target detached carriers in `#animating-components`,
  not `#container` children, and those carriers are `noAnimate` FLIP subjects).
  The parity geometry harness samples the opacity channel of every live
  animation and the full suite passes unregenerated, so no live opacity
  animation regressed. Documented here as the one deletion whose sole-purpose
  claim rests on "no live consumer" rather than "single re-homed consumer".

---

## 2. THE OPEN DESIGN QUESTION — resolved

**Question (Task 11 review carry-forward):** can a setter-driven layout
animation co-exist with a same-cycle animator FLIP on the same host (two
composited transform animations = double animation)?

**Brief's hypothesis:** during animator cycles the stack's layout write lands
inside the animator's measurement window where `component.noAnimate === true`,
so the setter snaps without playing → FLIP owns all motion → no double
animation.

**Verdict: the hypothesis is FALSE as stated, but there is NO regression.**
The actual call ordering (traced, not assumed) and the parity goldens both
confirm the cutover is behavior-preserving.

### Actual call ordering (traced)

`render-game._stateChanged` runs, in order:
1. `_resetAnimating()` (gate open) → `_animator.prepare()` (captures "before";
   `prepare` first calls `finishAllAnimations()` on every component, ending any
   prior self-play so the FLIP "First" is a resting sample).
2. `this.renderer.state = newState` — Lit commits the new component set. The
   stack's `updated()` / `slotchange` → `_slotChanged` → `_updateComponentClasses`
   runs here and issues `component.layoutTransform = …`.
3. `presentationPlanning.then(startStructuralMotion)` → `animateFlip()` →
   **double microtask** → `_doAnimate`, which only THEN raises the
   component-level `noAnimate` barrier and reads geometry.

Step 2 (the layout write) is committed **microtasks before** step 3 raises the
barrier — by design: the animator's documented double-microtask delay exists
specifically to let all databinding/slotchange cascades finish before
`_doAnimate`. So at the moment the stack writes `layoutTransform`,
`component.noAnimate` is **still false**, and for index-derived layouts
(`fan`/`pile`) — and for the id-hashed `messy` rotation when the element at a
slot rebinds to a different card id — the setter **self-plays**.

**Empirical confirmation** (instrumented `layoutTransform` setter, native
Playwright clicks driving real moves in `debuganimations`):

| Cycle | changed `layoutTransform` writes | self-plays started | `noAnimate` at write |
| --- | --- | --- | --- |
| `#fan` Public Shuffle | 13 | 13 | `false` (all) |
| `#fan` Draw | 9 | 9 | `false` (all) |
| `#shortstacks` Swap | 0 | 0 | — (messy STACK rotation is id-stable; value unchanged → no-op) |

### Why it is NOT a regression

The retired CSS `transition: transform var(--animation-length) ease-in-out`
fired at the **exact same** step-2 slotchange moment — the container was not
`.no-animate` then either (that class was also only applied in `_doAnimate`,
after the double microtask). Both mechanisms therefore started an ambient
transform animation on the same host, at the same instant, with the same
easing (`ease-in-out`) and the same duration source (`--animation-length`,
500ms in `debuganimations`), co-existing with the same FLIP. **The mechanism
swapped (CSSTransition → WAAPI); the observable motion did not.**

This is pinned mechanically by the new geometry golden
`geometry-debuganimations-fan-draw`, **recorded from the pre-cutover commit
(old CSS transition path)** and matched **unregenerated** by the setter path
(3 consecutive green runs; old-vs-old self-consistency verified first). The
geometry harness samples both `CSSTransition` and WAAPI animations from
`getAnimations()`, so an old-code golden legitimately fingerprints the CSS
transition curve, and a new-code run must reproduce it within tolerance 0.08 —
which it does, on the `progress`, `transform`, `opacity`, `timing`, and
`zIndex` channels.

### The genuine load-bearing invariant (corrected)

> The `layoutTransform` setter reproduces the retired CSS transition's exact
> duration source (`--animation-length`), easing (`ease-in-out`), and
> retarget-from-computed semantics. Because both fire at the same pre-barrier
> slotchange moment, wherever the CSS transition animated, the setter's
> self-play animates identically — including concurrently with a FLIP. The
> component-level `noAnimate` barrier suppresses only writes issued *while the
> barrier is up* (the animator's own measurement-time card-flip / faux-carrier
> style mutations); it is not what governs the stack's layout write.

Documented in code at both sites: the setter's `noAnimate` gate
(`boardgame-component.ts`) and the animator's barrier-set site
(`boardgame-component-animator.ts`).

---

## 3. Golden verdicts

**Regenerated: NONE.** Every pre-existing parity golden passed UNREGENERATED.

| Golden | Kind | Verdict |
| --- | --- | --- |
| `geometry-debuganimations-swap` | geometry | pass, unregenerated |
| `geometry-debuganimations-interrupted-swap` | geometry (crown jewel: mid-flight retarget) | pass, unregenerated |
| `geometry-memory-reveal` | geometry | pass, unregenerated |
| `geometry-fixture-fading-text` | geometry | pass, unregenerated |
| `geometry-fixture-game-outcome` | geometry | pass, unregenerated |
| `debuganimations-card-move` | trace | pass, unregenerated |
| `memory-reveal-one` | trace | pass, unregenerated |
| `blackjack-deal` | trace | pass, unregenerated |
| `pig-roll` | trace | pass, unregenerated |

**Trace-golden note.** The plan anticipated the trace suite MIGHT show new
`play`/`settle` events for layout tweaks that previously rode invisible CSS
transitions. In practice **no trace golden changed**: the traced scenarios
(`debuganimations` Swap, `memory` reveal, `blackjack` deal, `pig` roll) drive
`stack`-layout membership changes whose `layoutTransform` is either constant
(`''`) or id-stable (messy rotation), so the setter is a no-op there — the new
self-play surfaces only in `fan`/`pile` relayouts, which those four scenarios
do not trigger. That path is instead covered by the new geometry scenario.

**New golden added (not a regeneration):** `geometry-debuganimations-fan-draw`,
recorded from the **old CSS** path and matched by the new path — it is the
before/after parity anchor for the fan/pile-relayout-during-cycle case (the
only scenario that exercises setter self-play concurrent with a FLIP).

---

## 4. Before / after behavior table

| Aspect | Before (CSS transition) | After (layoutTransform setter) |
| --- | --- | --- |
| Layout transform animation mechanism | ambient CSS `transition: transform …` on `#container` slotted components | WAAPI `play()` self-animation from the pre-snap computed transform |
| Duration source | `var(--animation-length, 0.25s)` | `animationLengthMs()` (reads `--animation-length`; same value) |
| Easing | `ease-in-out` | `ease-in-out` |
| Fires at | slotchange relayout (pre-barrier) | slotchange relayout (pre-barrier) — same instant |
| Mid-flight retarget origin | current on-screen (computed) value | `getComputedStyle(this).transform` captured fresh per set (same) |
| Suppression during measurement | `.no-animate` container class unset the transition | writes issued while `component.noAnimate` is up snap (no play) |
| Gate participation | invisible to the completion gate (CSS transitions are unhooked) | **now gated** — layout tweaks emit `will-animate`/`animation-done` and hold the gate until settled (declared #713 change; the setter uses `timing:'immediate'`) |
| `opacity` transition on live components | present in the rule, no live layout consumer | removed (no live consumer; parity opacity channel unchanged) |
| Double animation (self-play + FLIP) | N/A (CSS transition + FLIP already co-existed) | equivalent co-existence; geometry golden confirms identical curves |

---

## 5. Verification run summary

- Parity suite (`tests/animations/parity/`): **31 passed** (incl. new fan-draw
  scenario), goldens unregenerated.
- `waapi-gate` + `waapi-play`: **9 passed**.
- `waapi-companion` (Table/Hand cross-surface flights — named riskiest
  consumers): **8 passed**; measured cross-surface skew 0.4ms.
- `npm run test:unit`: **254 passed**.
- `npm run type-check`: clean. `type-check:strict` unchanged pre-existing
  failures only (232 → 230; the cutover removed 2, introduced 0).
- Pre-existing failures NOT caused by this change (fail identically on the
  stashed baseline; out of Task 12 scope): `waapi-attrs.spec.ts` "stack
  forwards post-animation-delay", "stagger produces strictly increasing
  per-index animation delays"; `waapi-buttons.spec.ts` "move buttons disable
  during animation and re-enable after", "a move proposed while isAnimating is
  true is swallowed". All four time out on a disabled "To Hidden" button
  ("…is not possible right now") in this environment on baseline as well.
