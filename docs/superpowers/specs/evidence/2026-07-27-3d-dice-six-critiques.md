# 3D dice — six-dimension adversarial critique

Date: 2026-07-27
Branch: `worktree-three-d-dice` at `82f5a553` (31 commits off master `b8618541`)
Method: six independent critics, one per dimension, each barred from re-deriving what
`.superpowers/sdd/progress.md` already records. Only the visual critic had browser access,
so its measurements are unpolluted by contention.

The verdicts disagree, and that is the useful part. Five critics rated the branch anywhere
from "overwhelmingly idiomatic" to "premise verified". The sixth — the only one that
**looked at it** — said: *"No, I would not ship this to players as-is."*

That gap is the single most important output of this exercise. Every suite is green, the
compositing premise is verified, robustness held under 4,400 seeded rolls, and the thing
still has a visible rendering tear on every d20 roll. **Correctness testing and looking at
it are not substitutes for each other**, and this branch is the proof.

---

## Cross-critic convergences

Findings that two or more independent critics reached separately. These carry the most
weight, because each critic was given a different lens and no shared conclusion.

| Finding | Reached by | Note |
|---|---|---|
| A sanitized `item` crashes the render pass | creator-API (#4), robustness (D4) | Robustness pinned the mechanism: a sanitized component expands to `{}` (`selectors.ts:424`), and Lit 3.3.3 does not wrap `_$didUpdate`, so it rejects `updateComplete` and produces an unhandled rejection *every update*. Pre-existing on master; the roll logic now sits downstream of it. `boardgame-card.ts:423` guards correctly and the die does not. |
| Barrel dice are unreadable at real sizes | creator-API (#7), visual (#3) | Two independent compounding causes, one per critic. API: `--die-size` still defaults to 50px, inherited from the flat die, while its semantics changed to bounding-SPHERE diameter. Visual: the barrel is normalized by its circumradius and is ~2.5:1 long, so its ink fills **0.45** of the box against 1.13–1.27 for d10/d12/d20. Result: a d7 numeral is **4.3px** at the default. |
| The die has no "landed" signal | creator-API (#6), visual (#11) | API: the component dispatches zero events, so pig celebrates from `effectsForTransition` at *cycle start*. Visual confirmed the consequence on screen: the golden pulse fires at the die's layout anchor while the solid has already jumped ~60px away, and finishes on a 600ms version slot while the roll runs 1–2s. |
| Announced value can diverge from the shown one | robustness (D2), visual (#11) | Robustness found three `playMotionTracks` early returns that fire *before* the resting write, leaving the die in its raw body frame with `aria-label` still correct — the same divergence class the final whole-branch review caught for the d20, through a different door. Visual found the complementary gap: an `aria-label` change on a button is **not announced**, so the result never reaches a screen reader at all. |
| The d4 double-prints its value | creator-API, visual (#11) | Centre numeral plus three corner numerals: four numbers on one face, matching no real d4. |

---

## Blockers (would ship a visible defect)

### B1 — See-through holes through the die during a roll
`boardgame-die.ts` `#stage` (~:1172) sets `perspective: 6em; perspective-origin: 50% 50%`,
but the roll's **translation** lives on `#inner`, *inside* that 3D context. The tumble
carries the solid 10–91px from the fixed perspective origin (median ~56px on a 100px die),
so the view vector to a facet can exceed 20° off `+Z` — while `.facet { backface-visibility:
hidden }` culls against `+Z`, not against the camera.

Measured, 20 seeded rolls per shape sampled at 60fps:

| shape | frames with a visible facet culled (a hole) | frames drawing a facet the camera cannot see |
|---|---|---|
| d20 | 288/1300 (22.2%) | 401/1300 (30.8%) |
| d7 | 181/1276 (14.2%) | 211/1276 (16.5%) |
| d12 | 183/1276 (14.3%) | 104/1276 (8.2%) |
| d6 | 77/925 (8.3%) | 44/925 (4.8%) |

Holes are 10–25% of the silhouette and persist ~6 consecutive frames — a visible tear, not
a one-frame glitch. They cluster in the first ~150ms (die furthest from the origin) and
clear as it returns to centre, which is exactly what the geometry predicts.

**Fix:** move the roll's translation *out* of the perspective context (onto `#main` or a
wrapper above `#stage`) and keep only rotation on `#inner`.

### B2 — Red error text under the die on every pig roll
Polled from the die's own `#action-status` through a real roll: "Roll Dice is not possible
right now" (red, directly under the die) from 78–724ms, then "Wait for the current animation
to finish" 794–1233ms. Screenshots show the red line drawn *through* the tumbling die. It
carries `role="status"`, so a screen-reader user is told the roll is impossible — and is
never told the result. Likely pre-existing gating copy, but this branch **triples the window
it is visible for** (up to ~2s on a d12).

### B3 — `#inner`'s inline transform is never cleared
Nothing in the codebase writes `style.transform = ''` on `#inner`. `#orient` (resting pose)
is a *child* of `#inner` (physics pose), so they compose multiplicatively, and the
"mutually exclusive" invariant is one-way. Two paths return `_roll` to null after a
successful roll without self-healing — a face-count change on the same element
(`:1536`), and `_planRoll()` returning null (`:1638`, on unmeasurable size or a sim/bake
throw). Lit reuses the static `#inner` node with the stale style intact.

Measured over 900 seeded rolls, the stale prefix leaves the die **60–106px outside its own
100px slot**, permanently, until a *successful* roll overwrites it. `_startRoll`'s docblock
claims the opposite of what the code does. No test changes a face count or forces a plan
failure after a roll — this is the path the final whole-branch review hunted and missed,
because it looked inside a single roll's lifecycle where it does not exist.

### B4 — Barrel legibility
See convergence table. A d7 at the 50px default is a smudge that could not be read from a
screenshot in any of 8 rolls. Two fixes, both needed: default `--die-size` to 100px, and
normalize the barrel by its **short axis** (letting the long axis overflow), which roughly
doubles the mark size for free.

---

## Should-fix

- **The landed number is at a random rotation.** Across 8 landed d20s showing 13, **zero**
  were upright; one read as "ει". This was a deliberate decision on the grounds that a real
  die stops at a random angle — the visual critic's counter-argument is decisive: it is a
  rotation *about the presented normal*, so normalizing it **cannot** disturb the
  most-square-on-face guarantee that `7715aeef` just bought. The app already normalizes in
  its fresh-mount pose and that pose is dramatically easier to read. Fixing this also fixes
  the reload discontinuity (#10) and gives reduced-motion users the tidier pose.
- **The 3D die is less prominent than the flat die it replaces.** At the same `--die-size:
  100px`: ink 74.8×69.8 vs 100×100, pips 8.1px vs 12.6px, and `box-shadow: none` vs a full
  elevation shadow with hover lift. A player who just wants to read a number was better
  served before. Needs a contact shadow and a stronger hover affordance.
- **Duration and spin.** d6 is well judged (median 683ms) but d12 medians **1283ms** and
  d20 **1000ms**, and total rotation is ~1 turn, not the 2–3 claimed (d6 369°, d20 478°,
  one d20 at 194°). 16–40% of frames rotate <1°. Recommendation: spin faster (target 2–3
  turns) and trim the tail at <2°/frame rather than only fully-dead frames, targeting
  450–700ms on *every* shape.
- **Entry is a hard cut** of 10–91px against a max in-flight step of 5–11px/frame, and some
  rolls enter from *below* — a die rising off the table. Cap the offset (~0.4 die-widths)
  and bias it downward.
- **No landing beat.** Nothing marks the result arriving; combined with the slow tail you
  cannot tell when it finished. Worth more than adding bounce.
- **`MAX_PIP_VALUE = 9`** draws 7- and 8-pip lattices on d8 *triangles* at 4.2px/pip. No
  real d8 is pipped; lower the cutoff to 6.
- **The presented face barely wins on a d20** (`towardsCamera` 0.946 vs runner-up 0.891;
  the "≥2.8% of projected area" guarantee is below perceptual threshold). A faint tint or
  darker ink on the presented facet would settle it without touching geometry.

---

## Structural — cheap now, expensive later

- **`sceneTransform` bakes the reading pose into the die's own transform**, rotating the
  whole simulated world per roll. This single decision blocks a visible dice tray, multiple
  dice sharing a world, zoom-on-roll, and persistent dice — four of the sixteen catalogued
  treatments, and the four ranked highest after the basic roll. `#orient` already exists as
  the layer to move it to.
- **`RollConfig` carries one geometry and one kernel** (`dice-sim.ts:1440` and ~8 consumers),
  blocking mixed shapes in one throw. The same edit should split the all-or-nothing retry
  (`:1337`, `:1447`) where one awkward die currently rethrows all 25.
- **The `visual` channel's layering contract is undefined.** `#inner` is simultaneously the
  WAAPI target and the `preserve-3d` carrier; the die resolves the double-booking by
  *muting* the resting pose. That works for one producer. `boardgame-card` already owns
  `#inner`'s transform for its flip, and **every `boardgame-token` has a `filter` on
  `#inner`** (`boardgame-component.ts:85-87` via `boardgame-token.ts:140`) — plus an
  animated one for the throb — which forces `transform-style: flat`. Every token in the app
  is currently unable to host a 3D scene on its visual channel. Moving that filter to
  `#outer` is cheap while nothing depends on its stacking.

### Issue #801 is wrong and should be corrected
It names ancestor-flattened `preserve-3d` as the central obstacle to 3D tokens/meeples.
It is not: both the die and the card build perspective → `preserve-3d` entirely **inside
their own shadow root**, an ancestor transform flattens the *result* into itself rather
than the descendant's local context, and `boardgame-component-stack` contains zero
`overflow`/`filter`/`opacity`/`transform` declarations. The die already demonstrates it —
`#main` carries a `transform` and is the direct parent of `#stage`. The real obstacle is
the token's own `filter`, above.

Also worth recording: `boardgame-die` extends `BoardgameAnimatableItem`, not
`BoardgameComponent`, and carries no `[boardgame-component]` attribute — so it cannot be a
stack child at all. The branch did not solve the stack question; it never met it.

---

## Verified premises (negative results worth keeping)

- **Compositing holds, unambiguously.** `ActiveTransformAnimation` on the animating element
  for every shape, and behaviourally: with the main thread hard-blocked by an 800ms spin,
  the tumble rendered 45–48 distinct frames (~57fps) against 1–2 for a
  non-compositable control. Literal `matrix3d` bought exactly what it was meant to.
  *Correction:* the code's claim that `var()`/`calc()` in a transform keyframe forfeits
  compositing is overstated — Chromium resolves `calc()` over a static custom property at
  compose time. Right decision, wrong stated mechanism.
- **Robustness held broadly.** 4,400 seeded rolls across 11 face counts: zero throws, zero
  non-finite tokens in any emitted `matrix3d`, `restingTransform === curve(1)` byte-exact
  in all 4,400. Hostile simulator configs (bounds 1e300, gravity 1e12) never produced a
  non-finite pose. Reduced motion and every interruption path are structurally safe rather
  than accidentally safe. No memory leak: 0.89KB retained per mount-roll-unmount cycle
  across 300 iterations with forced GC.
- **The solids are excellent at rest.** 24-step rotation sweeps of every shape: correct
  silhouettes at every angle, no z-fighting, no edge gaps, no facet popping, correct depth
  sorting. Pips foreshorten to correct ellipses. The d10's kites and equator are right.
- **`dice-bake.ts` is a genuinely general primitive** — it consumes
  `{samples: [{t, position, orientation}]}` and nothing in it is about dice.
- **Sampled curve tracks landed on the right channel and the type enforces it**
  (`component-track.ts:136-151` refuses curves *and* resting on `host`), deciding the
  ownership boundary #801 asks about before a second component exists to get it wrong.
- **Relabeling already covers** Fudge dice, symbol dice, blank faces, repeated values and
  loaded dice with no new code — verified by probe.

---

## Known scaling limit

`preserve-3d` promotes every facet to its own composited layer, and the cliff arrives **at
rest**, not while rolling. An unrelated animation drops to 40fps beside a stationary d64
(329 layers) and 19fps beside a d100 (509 layers). It is layer count, not pixels: a d100 at
50px measured *worse* than at 120px. Nothing shipping today is affected (pig's d6 is 15
layers; ten d20s hold 44fps), but a d100 percentile die is exactly the shape a game would
reach for, and the cost cannot be escaped by not rolling.

Two related numbers: simulation cost tracks **vertex** count, not face count (a d12 is
twice a d20), and the multi-die simulator path is roughly quadratic (10 dice ≈ 400–630ms),
latent only because the component always passes `dieCount: 1`.

---

## Idiomatic-hygiene items

- The physics→CSS reflection exists in **two** places — `dice-bake.ts`'s `CSS_AXIS_SIGN`,
  whose comment claims it is "the single place the Y flip lives", and `boardgame-die.ts`'s
  `toScreen`. The branch's own history is the argument: `toScreen` shipped as `(x,-y,-z)`,
  a proper rotation into a left-handed frame, rendering the mirror solid and presenting the
  wrong face, while every unit test passed.
- **~1,060 lines of DOM-free geometry live inside the Lit component** (54% of a 1,980-line
  file) with zero unit coverage, against the convention `spatial-board-geometry.ts` +
  `.test.ts` establishes in the same directory. Three functions are exported purely as a
  test seam — the seam this repo normally creates by *moving the code*.
- **A third seeded-hash implementation**, byte-identical to `effects/particle-burst.ts`'s —
  and that older copy carries precisely the 32-bit aliasing bug the dice work just fixed.
  A live determinism defect in another system, and the case the "prefer existing
  primitives" guidance exists for.
- `src/motion/` now holds both `geometry.ts` (140 lines of FLIP rect math) and
  `die-geometry.ts` (711 lines of convex hulls and inertia tensors). `src/dice/` is where a
  maintainer would look.
- Stale docs in `dice-sim.ts` (`:6`, `:94`, `:367`) still describe the seed as
  `(component id, state version)`, changed to `RollCount` in `fab98c37`.
