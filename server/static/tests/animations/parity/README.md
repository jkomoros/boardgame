# Animation parity harness

Regression net for the animatable-item unification project
(`docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md`).
Two complementary suites pin the CURRENT animation behavior so migrations can
prove zero regression mechanically.

## Running

Needs the offline-dev server serving this checkout (`boardgame-util serve
--offline-dev-mode`; in a worktree, launch with a GOPATH shim so the import
path resolves to the worktree). Then from `server/static/`:

```bash
npx playwright test tests/animations/parity/            # compare vs goldens
PARITY_RECORD=1 npx playwright test tests/animations/parity/  # re-record
```

Re-record ONLY when a behavior change is intended and declared; every golden
diff must be explainable hunk by hunk.

## What each suite pins

**Trace suite** (`trace.spec.ts`): animation-hook event streams. Per-scenario
contracts, matched to what each game can promise deterministically:

| Golden | Contract | Why not stricter |
|---|---|---|
| `memory-reveal-one` | exact counters + per-element event sequences (ids canonicalized by first appearance) | deck shuffle renames ids; everything else is deterministic |
| `debuganimations-card-move` | exact gated-cycle counts + required kinds | FLIP skips no-op transforms and messy rotations hash per-game random ids, so play counts vary (observed 92–138 for the same action) |
| `blackjack-deal` | structural invariants + required kinds | deal length depends on the shuffled deck |
| `pig-roll` | structural invariants + required kinds | post-roll cycles depend on the rolled value |

All modes assert: watchdog 0, every play settles in-window; exact modes also
assert open/close balance.

**Geometry suite** (`geometry.spec.ts`): motion-curve fingerprints. Every
animation in a scenario (deep shadow-root walk — `document.getAnimations()`
returns NOTHING for shadow-tree animations in this Chromium) is paused and
seeked to fractions 0/.25/.5/.75/1 of its own delay+duration. Curves compare
as a SET under 0.08 tolerance (counts are per-game random; count regressions
are the trace suite's job). Wave-union sampling captures chained cohorts.
Scenarios: swap flight, reveal flip, interrupted-swap retarget, plus
component fixtures for `fading-text` and `game-outcome` (the Phase 1
before/after anchors — full-game flows can't drive them deterministically).

Per curve, five channels:

| Channel | What it is | Null when |
|---|---|---|
| `progress` | bounding-rect-center travel | net displacement ≤ 2px (not a travel animation) |
| `rotation` | transform matrix's LINEAR part — rotation/scale/skew/perspective, every entry except the translation column | that part's path ≤ 0.01 |
| `translation` | transform matrix's translation column (tx, ty, tz), in px | its path ≤ 0.01 |
| `opacity` | raw measured opacity | constant within 0.01 |
| `timing` / `zIndex` | declared `[duration, delay]` on a 25ms grid; computed z-index | z-index constant for the whole flight |

The three motion channels are normalized by **cumulative path length**
(`Σ‖pᵢ − pᵢ₋₁‖` accumulated, over the total), NOT by net displacement, so
each is monotone non-decreasing and in `[0,1]` by construction. That matters
for anything that does not travel monotonically from A to B — a tumbling
die, or any out-and-back — where the old chord-over-net-displacement ratio
produced values well past 1 on an absolute 0.08 tolerance, non-monotone
samples, and (for a landing near the start pose) a near-zero denominator
that either exploded the ratios or nulled the channel outright. It was not
hypothetical: `fading-text`'s `scale(1) → scale(6)` fade snaps back to base
on the final sample, so its whole transform channel used to record as null.
`rotation` and `translation` are separate channels because their scales are
incomparable — rotation entries are unitless (2.83 for a whole matrix) while
translation is raw pixels, so one Frobenius sum over both let a 60px travel
drown the rotation to ~5% of the norm and degenerate into a duplicate of
`progress`. `fingerprint-normalization.spec.ts` pins these invariants twice:
on synthetic non-monotone input, and over the recorded golden corpus itself.

**A curve whose `progress`, `rotation`, `translation` AND `opacity` are all
null is DROPPED** (sub-threshold noise — near-no-op FLIPs whose presence is
per-game random — asserts nothing). Recording a scenario therefore fails
loudly only if it produces NO curves at all; a scenario for a specific
element must additionally assert that its own curve survived, or a
regression that stops the element moving would just shrink the curve set
inside the tolerant set comparison. A die scenario in particular must assert
it is never all-null.

## Teeth (verified failure detection)

- Suppressing card `play` hooks → trace suite fails debuganimations, memory,
  blackjack (pig legitimately unaffected).
- `ease-in-out` → `linear` sabotage → geometry fails the swap scenario.
  (Preserved across the path-length change by construction: pure-translation
  curves record byte-identical values under the new normalizer.)
- ~~Flip-shape sabotage (`rotateY(180)` → `rotateY(90)`) → geometry fails the
  memory reveal.~~ **NO LONGER BITES** — see the rotation-MAGNITUDE blind
  spot below. Re-verified by hand on 2026-07-26: with the sabotage applied,
  `geometry: memory reveal flip curves` PASSES.

## Accepted residual blind spots (harness-critic ledger)

Reviewed adversarially at Phase 0 close; these are ACCEPTED, with owners:

- **`will-animate` DOM event VOLUME** — the trace goldens record `play`,
  `active`, `settle`, `gate-open` and `gate-close`; they do *not* record
  `will-animate`. So a change in how often that event fires is structurally
  invisible to this harness. This is not hypothetical: the 3D-dice branch made
  `will-animate` a per-gated-play *declaration* rather than a 0→1 transition,
  which doubles its volume on an ordinary FLIP-plus-fade (two host tracks), and
  every golden stayed byte-identical. What the goldens *do* pin is the part
  that matters — `gateOpens`/`gateCloses` stay 1/1 and `plays`/`settles` are
  unchanged, because `AnimationGate.willAnimate` is idempotent (a keyed write
  plus a strictly monotone deadline). Exposure is therefore limited to a
  NON-idempotent listener being added later. Owner: the three current
  listeners were each traced by hand (`animation-gate.willAnimate`,
  `boardgame-render-game._componentWillAnimate`,
  `boardgame-game-view._rosterWillAnimate`); any new `will-animate` listener
  must be checked for idempotence by review, because no test will catch it.
- **Rotation MAGNITUDE (NEW, and a real regression in coverage)** — the
  path-length normalizer is magnitude-invariant by construction: a rotation
  through 180° and one through 90° under the same easing produce the SAME
  normalized `rotation` channel. The previous chord-over-net-displacement
  lens was magnitude-SENSITIVE for rotations by accident, not by design —
  normalized chord is `sin(θ/2)/sin(Θ/2)`, which depends on the total angle
  `Θ` (a 180° flip recorded `[0, 0.2, 0.71, 0.98, 1]`, a 90° one would have
  recorded `[0, 0.14, 0.54, 0.89, 1]`, and the 0.17 midpoint gap exceeded the
  0.08 tolerance). That entanglement of shape with magnitude is exactly what
  made the lens unusable for a tumbling die, so it had to go — but the flip-
  shape tooth went with it, VERIFIED by re-running the documented sabotage
  above. Nothing else in the harness pins how FAR a rotation turns. Restoring
  it needs a separate scalar, e.g. a per-channel `directness` = net
  displacement over path length (`2·sin(Θ/2)/Θ`: 0.64 for a half turn, 0.90
  for a quarter, ~1.0 for the small per-game-random messy-stack tilts, 0 for
  an out-and-back), which stays deterministic for every current scenario —
  but would be per-ROLL random for a tumbling die, so a die scenario would
  have to opt out of it. Deliberately NOT added here: it is an undeclared
  channel, and planting a per-roll-random value in the die golden is the
  failure mode this whole harness exists to prevent. Owner: unassigned —
  needs an explicit decision before Phase 3 relies on flip shape.
- **Absolute endpoints / raw positions** — per-game layout randomness makes
  raw-rect goldens unreproducible. Wrong-final-position bugs that preserve
  curve shape are not caught here; the existing behavioral suites
  (`waapi-gate`, `waapi-companion`) and geometry's z/timing channels bound
  the exposure.
- **Blackjack companion flight SHAPE** — only cross-surface skew sync is
  pinned (`waapi-companion.spec.ts`); the flight's own curve has no golden
  (deal randomness). Phase 3 must rely on the debuganimations curves plus
  companion suite.
- **Reduced-motion goldens** — Phase 1 *declares* a behavior change here
  (kernel skip vs CSS 1ms sprint), so pinning the current behavior would
  contradict the approved spec. Kernel-level reduced-motion behavior is
  covered by `waapi-play`/`waapi-companion`.
- **Mobile viewport** — geometry runs at 1280×900 only; curves are
  size-normalized in principle. Low value vs cost.
- **Roster / player-info animations** — currently un-gated (the #714 gap);
  the Phase 2 change lands with its own gate-witness test (plan Task 10).
- **`expectedSettleMs` watchdog extension end-to-end** — no scenario is long
  enough to need it; owned by the gate-kernel unit tests (plan Task 8).
- **0.08 tolerance** — validated against large-effect teeth (0.17–0.21
  midpoint deltas); a sub-tolerance easing tweak (<0.08 at every fraction)
  would pass. The midpoint sample carries most discriminative power.
- **Short-lived animations** — a sub-two-frame animation could finish before
  the wave pause catches it; would surface as a loud existential-match
  failure (flake), not a silent pass.
- **Ambient / infinite ungated animations across a cycle** — NOW COVERED (was
  a silent gap). An infinite highlight throb (`{ gated: false }`) is not a
  completion-cycle participant, so no gate/trace/geometry golden observes it,
  yet the cycle-start registry sweep used to cancel it every state change (the
  ambient-animation-sweep regression). `token-throb.spec.ts` now pins throb
  survival through a real render-game cycle AND through a DOM reparent, and
  `finish-gated-animations.spec.ts` pins the `finishGatedAnimations` kernel
  contract directly (gated force-settled, ungated left running; and
  `finishAllAnimations` still kills ambient loops for the tree-departure
  paths). The out-of-tree ungated-ambient contract — any game-authored
  `{ gated: false }` loop must outlive a cycle — is protected by the sweeps'
  `finishGatedAnimations` semantics. See
  `docs/superpowers/specs/evidence/2026-07-26-ambient-animation-sweep.md`.
- **`composite: 'replace'` as the same-host no-double-motion invariant** —
  when a stack's `layoutTransform` self-play and the same cycle's FLIP host
  track both animate one host's transform, safety comes from `play()`'s
  pinned `composite: 'replace'` (the higher animation wins outright each
  frame), not from easing/duration matching. This is enforced in `play()`
  (`src/components/boardgame-animatable-item.ts`) but is NOT
  golden-detectable: curves are displacement-normalized, so a uniform
  doubling under `composite: 'add'` divides back out of the comparison.
  See `docs/superpowers/specs/evidence/2026-07-26-stack-transition-cutover.md`.

### Phase 2 topology asymmetry (accepted, documented)

The registry/context providers live on `boardgame-render-game`; the roster
is its DOM SIBLING. Roster-hosted animatables are therefore gated (via the
game-view event pipe) but NOT registry-swept: on a cycle handoff a
mid-flight board animatable is force-finished (snaps) while a roster one
completes smoothly. Benign — a late roster settle is a kernel no-op. Roster
items also resolve a null ambient animation context (default timing), which
is correct: their animations are local effects, not version-slot
participants.

Orphan-settle (a roster animatable removed from the DOM mid-animation) IS
now covered: `BoardgameAnimatableItem.disconnectedCallback` force-settles
any still-gated item a microtask after disconnect (deferred so a
same-tick reparent never snaps a live animation), and
`boardgame-game-view`'s `_rosterWillAnimate` additionally subscribes to the
item's `settled()` promise as a done channel that does not depend on DOM
presence — the bubbled `animation-done` event this suite otherwise relies
on cannot reach render-game from a detached node. See
`tests/animations/parity/player-info-gate.spec.ts`'s third test and
`docs/superpowers/specs/evidence/2026-07-25-roster-orphan-settle.md`.
