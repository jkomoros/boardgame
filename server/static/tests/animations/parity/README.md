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
Scenarios: swap flight, fan-draw relayout, reveal flip, interrupted-swap
retarget, plus component fixtures for `fading-text` and `game-outcome` (the
Phase 1 before/after anchors — full-game flows can't drive them
deterministically). The die roll is deliberately NOT one of them; see the
ledger entry below.

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
inside the tolerant set comparison. Any scenario added for a specific
element — a die roll above all, since a tumble can land near its start
pose — must assert its own curve survived.

## Teeth (verified failure detection)

- Suppressing card `play` hooks → trace suite fails debuganimations, memory,
  blackjack (pig legitimately unaffected).
- `ease-in-out` → `linear` sabotage → geometry fails the swap scenario.
  (Preserved across the path-length change by construction: pure-translation
  curves record byte-identical values under the new normalizer.)
- Flip-MAGNITUDE sabotage (`rotateY(180)` → `rotateY(90)` in
  `boardgame-card.ts`'s `_innerTransformFor`) → geometry fails the memory
  reveal. **RESTORED** (re-verified by hand 2026-07-26): the curve SET no
  longer catches this — path-length normalization is magnitude-invariant, so
  the sabotaged `rotation` channel is byte-identical — a separate scalar
  assertion does. `sweptRotationDegrees()` decomposes each sampled matrix to
  its pure rotation (Gram-Schmidt, so scale/skew divide out) and accumulates
  the angle between successive orientations; `geometry: memory reveal flip
  curves` asserts the largest swept angle in the cycle is 180° ± 3°. Measured
  clean: `[180, 0, 0, …]` — only the flip rotates at all. Under the sabotage:
  `Error: memory's reveal must sweep a half turn; swept angles (deg, desc)
  were [90,0,0,…]`. Scope: this pins memory's half turn ONLY. The assertion
  lives in the scenario, not in the fingerprint, deliberately — see the
  rotation-MAGNITUDE entry below. The helper is scenario-agnostic and is
  meant to be pointed at any scenario whose rotation magnitude is a real
  invariant (next: a fixed-seed die roll); step-wise accumulation is what
  makes a multi-turn tumble measurable, since a 360° roll's start-to-end
  angle is 0.

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
- **Rotation MAGNITUDE — narrowed to rotations nothing pins** — the
  path-length normalizer is magnitude-invariant by construction: a rotation
  through 180° and one through 90° under the same easing produce the SAME
  normalized `rotation` channel. (The previous chord-over-net-displacement
  lens was magnitude-SENSITIVE for rotations by accident, not by design —
  normalized chord is `sin(θ/2)/sin(Θ/2)`, which depends on the total angle
  `Θ`. That entanglement of shape with magnitude is exactly what made the
  lens unusable for a tumbling die, so it had to go, and the flip-shape
  tooth went with it.) **No fingerprint CHANNEL will ever cover this**, and
  the task-3 report's proposed `directness` scalar (`2·sin(Θ/2)/Θ`) is
  unusable for the same reason any magnitude channel is: curves compare as a
  SET across every animating element, and debuganimations' messy-stack tilts
  are per-game RANDOM in magnitude, so such a channel would flake every run
  (and would be per-ROLL random for a tumbling die besides). What covers it
  instead is a per-scenario scalar assertion — `sweptRotationDegrees()` —
  used where the magnitude is a genuine product invariant. Currently that is
  memory's reveal flip and nothing else (see Teeth above). **Still
  uncovered**: every rotation whose magnitude no invariant pins — the
  per-game-random messy-stack tilts above all, and any future rotation added
  to a scenario without its own swept-angle assertion; a change to how far
  those turn is invisible here. Owner: whoever adds a rotating animation
  adds the assertion, or accepts the gap explicitly.
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
- ~~**Roster / player-info animations un-gated**~~ — CLOSED (the #714 gap).
  Roster animatables now hold the gate through the game-view event pipe, with
  `player-info-gate.spec.ts` as the witness in both directions: a roster
  animation forwarded during a real board cycle holds the close, and one with
  no cycle open leaves the gate untouched. The residual topology asymmetry —
  gated but not registry-swept — is described at the end of this file.
- ~~**`expectedSettleMs` watchdog extension end-to-end**~~ — NOW COVERED (was
  "no scenario is long enough to need it; owned by the gate-kernel unit tests").
  `die-roll.spec.ts`'s *a roll past the watchdog floor extends the deadline
  instead of being cut off* mounts a die inside the LIVE renderer — real ambient
  registry, real `will-animate` listener, real cycle opened by a real move —
  declares a 4500ms `postAnimationDelay` on it, and throws it inside pig's own
  cycle. Measured: with the declaration reaching `AnimationGate.willAnimate` the
  die declares 4822ms (322ms of tumble + the hold), the gate stays open 4870ms,
  the watchdog does not fire and the die reports itself settled at 4774ms,
  `finished`; with `expectedSettleMs` dropped on the way into the gate
  (`boardgame-render-game.ts:695`) the gate closes at 4000.8ms — the floor, to
  under a millisecond — the watchdog fires once, and the die never reports
  settling at all, still `running` when the cycle is torn down.

  **The length used to come from the physics and no longer can.** The original
  witness was a d48 seed running to `dice-sim.ts`'s own 5000ms cap. The physics
  retune that followed cut every throw so far that the longest roll over 4,400
  seeded throws (11 shapes × 400 seeds) is **2761ms** — a d30 — so no seed can
  reach the 4000ms floor and the test was deterministically dead. It did not go
  red when that landed; it just stopped exercising anything, and only its own
  premise guard would have said so. The declared hold is both a real product
  path (`postAnimationDelay` is a property of every animatable item and
  `boardgame-component-stack.ts` parses it off an attribute) and a strictly
  better fixture, because the occupancy is CHOSEN rather than sampled from a
  distribution any physics change can move.

  Note that the OTHER app-level roll test (*a multi-second roll never trips the
  gate watchdog*) still passes under that same sabotage — a sub-second pig roll
  never reaches the floor, which is exactly why this blind spot survived until a
  scenario was built that does.
- **The die's tumble has NO geometry golden** — the sampled-motion design
  planned one, recorded from a fixed-value fixture. It was not built, and the
  decision is recorded here rather than left as a silent omission. Why:
  a golden fingerprint is *shape under normalization*, and the die's shape is
  a seeded physics trajectory, so the golden would restate the simulator's
  output rather than any product invariant, and any change to the sim,
  the tray, the camera or the trim would rewrite it wholesale. What pins the
  die instead is `die-roll.spec.ts`, which asserts the RENDERED KEYFRAMES
  against a trajectory recomputed in-page from the component's own exported
  seed derivation — a strictly tighter comparison than a 0.08-tolerance
  normalized curve set, and one that catches a wrong seed, a wrong pixel
  radius or a mirrored basis. What is genuinely lost: the die does not
  participate in the cross-cutting curve SET, so a harness-wide regression in
  how sampled motion is fingerprinted would not show up on the die. That is
  covered instead by `fingerprint-normalization.spec.ts`, which drives the
  normalizer with synthetic non-monotone and return-to-start tumbles — the
  two failure modes the die was the reason to fix — and re-checks the
  invariants over the whole recorded golden corpus. Owner: whoever adds a
  second sampled-motion producer should reconsider, because at that point the
  set comparison starts being worth its cost.
- **The CARD's 3D-context probe is weaker than the token's, by construction** —
  `component-3d-context.spec.ts` proves a `filter` on `#inner` costs the
  component its 3D context by mounting a `translateZ`'d face and measuring its
  perspective magnification. On a token the probe mounts ON `#inner`, so
  flattening it takes the magnification away entirely: 20px → 10px. On a card
  it cannot — `#inner` already carries the flip transform — so the probe hangs
  a scene of its own off a CHILD of `#inner`, keeps its own `perspective(200px)`
  either way, and what is at stake is only the further 1.25× that `#outer`'s
  `perspective: 1000px` contributes through `#inner`: 25px → 20px. Both
  numbers were measured, and the smaller signal is still deterministic and
  still fails if the rotated alt-shadow goes back on `#inner`. What is lost is
  headroom: a future change that cost the card its outer perspective for some
  OTHER reason would land on the same 20 and read as the filter regression.
  Owner: whoever gives `boardgame-card` a real 3D mode should move the probe
  onto `#inner` and recover the 2× tooth.
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
