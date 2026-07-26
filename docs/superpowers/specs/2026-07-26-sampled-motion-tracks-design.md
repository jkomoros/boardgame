# Motion Curve Tracks — Design

Date: 2026-07-26
Status: revised after API critique and hazard review
Motivation: `2026-07-26-3d-dice-design.md` needs to play a physically-baked
trajectory, which today's two-endpoint motion track cannot express.

## The gap

A `ComponentMotionTrack` is an endpoint pair (`component-track.ts:11-23`)
compiled into two WAAPI keyframes. Motion that is not "interpolate A→B under
one easing" has no route through the track system: today a baked dice
trajectory, tomorrow any path-following, multi-bounce or simulation-authored
motion.

### Honest motivation (corrected)

The first draft claimed the alternative — calling `play()` with a raw
keyframe list — "forfeits channel ownership". **That claim is wrong.**
`componentMotionTracks` constructs its channel set *inside a single
invocation* (`component-track.ts:66`), and `playMotionTracks` keeps no
per-element registry (`_liveAnimations` is keyed by `Animation`,
`boardgame-animatable-item.ts:65`). Two separate track batches on the same
element already contend freely today, with `composite: 'replace'` letting the
later silently win while the earlier keeps holding the gate.

What the extension genuinely buys, and the only grounds it should be argued
on:

- **Harness visibility** — sampled motion appears in the geometry goldens.
- **Gate participation** — the roll holds the completion gate correctly.
- **Reduced-motion handling** — one policy, not per-producer discipline.
- **Uniform compilation** — one shape flows to the structural plan.

Per-element channel bookkeeping is a real gap, but it is a *separate* fix
(see "Deferred"), not something this extension delivers.

## The reframing

Per-keyframe `easing: 'linear'` is a **no-op** — WAAPI keyframe easing
already defaults to linear. The distortion a sampled trajectory suffers comes
from the *effect-level* `ease-in-out` the kernel injects
(`boardgame-animatable-item.ts:289-293`, surviving `{...defaults,
...requested}` at `timing.ts:77`).

So this is not "a track carries more keyframes" but **"a track can claim its
channel's _timeline_."** That has a concrete implementation consequence:
`playMotionTracks` currently passes **one** `timing` object to every track in
the batch (`boardgame-animatable-item.ts:198-204`), so timing must become
**per-track derived** rather than shared.

Effect level, not per-keyframe, for a second reason:
`boardgame-component-animator.ts:559-568` reads executed easing off
`effect.getTiming()` into `StructuralExecutedTiming`
(`structural-plan.ts:52-60`), published to renderer and effect-layer
consumers. Encoding character in keyframes would make that field report
`linear` for everything and quietly stop meaning anything.

## The API: curve in, samples out

The producer supplies a **function of progress**; the compiler owns the
sampling grid, evaluates the curve, and never retains it.

```ts
export interface ComponentMotionSample {
  readonly offset: number;   // WAAPI's word — emitted verbatim into a Keyframe
  readonly value: string;
}

export interface ComponentMotionTrack {
  readonly target: ComponentMotionTarget;
  readonly property: ComponentMotionProperty;
  readonly samples: readonly ComponentMotionSample[];  // >= 2, uniform, spanning [0,1]
  readonly timeline: 'eased' | 'sampled';
  readonly resting?: string;
}

type MotionEndpointsInput = Readonly<{ from: string; to: string }>;

type MotionCurveInput = Readonly<{
  curve: (progress: number) => string;  // pure, cheap; sampled at compile time
  resolution?: number;                  // clamped to [2, 256], default 64
  resting?: string;                     // default: curve(1)
}>;

// Curve input is permitted ONLY on the visual channel — see "Host sampling is
// forbidden". Endpoint input remains available on both targets.
type ComponentMotionTrackInput =
  | (Readonly<{ target: ComponentMotionTarget; property: ComponentMotionProperty }> & MotionEndpointsInput)
  | (Readonly<{ target: 'visual'; property: ComponentMotionProperty }> & MotionCurveInput);
```

`curve` and `progress` are the system's existing vocabulary (the parity
harness's ontology is *motion curves*, `geometry-helpers.ts:6,22-45`;
`progress` already means fraction-of-active-interval, `release.ts:4-5`).
`offset` is WAAPI's because it is emitted verbatim.

### Why the compiler owns the grid

An earlier draft let producers author offsets, then spent five invariants
policing them (range, monotonicity, endpoint anchoring, min count, max
count). All five exist only because the producer was handed a grid it should
never have had. A compiler-generated uniform grid deletes them, and matches
the module's established split: inputs are requests, compiled tracks are
normalized frozen facts (`exactTrack`, `:36-60`; cf. `presence.ts:8-19`,
`release.ts:39-70`).

A closure is the right *authoring* boundary and the wrong *storage*: it can
capture DOM, is non-deterministic, cannot be meaningfully frozen or diffed by
a golden — everything this system refuses to carry across boundaries (cf.
`sanitizeMotionSubjectSnapshot`, `subject.ts:26`). Function in, frozen
samples out.

### Endpoint form rewrites to two samples

One representation everywhere; `{from, to}` compiles to offsets 0 and 1,
which is a WAAPI no-op for a two-keyframe list. Only unit-test literals
churn.

## Constraints the hazard review forced

### Host sampling is forbidden, in the type

A sampled `host` track would break three contracts: the FLIP resting write
(`boardgame-component.ts:356`), `StructuralMotionPath`'s two-point viewport
contract (`structural-plan.ts:27-31`), and trail-echo synchronization, which
animates echoes along a straight `fromCenter → toCenter` line using the
effect easing (`boardgame-effect-layer.ts:996, 1017-1035`). The dice design
already argues the tumble belongs on `visual`; this encodes it rather than
leaving it to discipline.

### Sampled tracks require `timing: 'immediate'`

Under the `'version'` policy, `timing.ts:130-141` clamps duration — offsets
being fractional, a 1500ms bake in a 600ms slot plays uniformly at **2.5×**:
geometrically faithful, physically wrong, and silent. Worse,
`timing.ts:127-129` can resolve to `{kind:'skip'}` → `play()` returns null →
`playMotionTracks` reports `not-started` → the animator cancels *sibling*
tracks (`boardgame-component-animator.ts:1792`), so a staggered die can
vanish outright. And because `_ambientAnimationContext()` is null in solo
play, all of this happens **only in companion mode**.

### Validation belongs at the producer boundary, not in a throw

The stated "loud throw" strategy does not exist at runtime: `_planMotionTracks`
wraps compilation in try/catch, logs, and returns host-only tracks
(`boardgame-component-animator.ts:592-607`); `playMotionTracks` catches and
returns `playback-error` (`boardgame-animatable-item.ts:219-223`). A throw is
therefore *silence* in the animator path — while in the die's path
(`boardgame-die.ts:248`, called bare inside `updated()`) the same throw
aborts a render pass. Neither is what "loud" meant.

So: validate in the **bake** (pure, unit-tested); **clamp** quality knobs
rather than rejecting, matching `effect-budget.ts`'s clamp-and-degrade
precedent (`reserveEffectBudget` returns null to degrade, never throws); and
surface a real loss into `StructuralExecution` with a `motion-invalid`
reason (`structural-plan.ts:75-78`) so a test can assert it instead of a
`console.error` nobody reads.

### `will-animate` must fire per gated play

It currently dispatches only on the 0→1 gated transition
(`boardgame-animatable-item.ts:311-326`). That is harmless today only because
every track in a batch shares one timing. Once sampled tracks carry their own
longer duration — and `compileComponentMotionTracks` orders `visualTracks`
**last** (`component-track.ts:104-125`) — the watchdog would arm from the
short host FLIP's settle time and force-close mid-tumble. `AnimationGate`'s
`willAnimate` is already idempotent and monotone (`animation-gate.ts:110-118`),
so dispatching per gated play is safe; alternatively `playMotionTracks`
declares the batch max up front.

### The resting-pose contract is the component's

With `fill: 'none'`, both `finish()` and natural completion render the
**resting style**, not the last sample (`boardgame-component.ts:354-357`
states this explicitly). `playAnimation` writes resting style only for
`host`; the `visual` channel has no framework write, which is why the card
maintains `_updateInnerTransform` by hand.

A curve track therefore carries `resting` (default `curve(1)`), and
`playMotionTracks` writes it to the resolved target after starting the
animation. The compiler cannot verify it against computed style, so the
component asserts at runtime (compare computed transform to the last sample,
`console.error` on mismatch) and a unit test pins that the bake's final
sample equals the resting pose for the landed value.

### Bake literal matrices

Chromium composites transform keyframes only when values are compositable.
Literal `matrix3d(...)` is; `var()`/`calc()` — as today's die uses
(`boardgame-die.ts:237`) — forces main-thread animation. Bake to literal
matrices.

### `resolution` clamps and decimates

Clamped to `[2, 256]`, default 64. Over-budget bakes **decimate** (a
90-sample trajectory resampled to N is the same trajectory) rather than
failing, per the budget precedent above. The treatments catalog already
anticipates ~12 individually-simulated dice with shared trajectories beyond.

## Required harness change (in scope)

`geometry-helpers.ts:179-193` fingerprints transforms as
`matrixDist(p, first) / matrixDist(last, first)` — normalizing by **net
displacement**, which assumes monotone travel. Against a tumble this fails
four ways:

1. **Values ≫ 1, non-monotone.** Path length far exceeds net displacement, so
   the absolute `0.08` tolerance (`:240, 255-261`) is applied on an
   uncalibrated scale — brittle goldens.
2. **Near-zero denominator.** A trajectory landing near its start drives
   `totalMatrix → 0`: just above the `0.01` threshold ratios explode, just
   below the channel becomes null — and an all-null curve is **dropped
   entirely** (`:214-215`). The most complex animation in the app would
   contribute *nothing*, silently.
3. **Rotation drowned by translation.** `matrixDist` Frobenius-sums unitless
   rotation entries (≤2.83) with translations in **raw pixels**; a 60px
   travel makes rotation ~5% of the norm, degenerating the channel into a
   duplicate of `progress`.
4. **Run-to-run instability.** The landing orientation depends on the
   server-rolled value, so `last.matrix` — the denominator — changes per run.

Fixes, scoped here because the sampled track is what breaks the assumption:
normalize by **path length** (`Σ‖mᵢ − mᵢ₋₁‖`), **separate rotation from
translation** before normalizing, record the die golden from a **fixed-value
fixture** rather than a live roll, and assert the die's curve is **never
all-null** so failure mode 2 cannot fail open.

Also note the harness samples `frac = 1`, which is exactly the after-phase
boundary — it records *resting style*, not the last sample (H7 again). If
those differ, the golden normalizes against a matrix the animation never
reached.

## Rejected alternatives

- **Discriminated union on the compiled track** — makes `from`/`to`
  conditionally present on a published artifact (it rides in
  `MotionTrackPlayback.track`, `boardgame-animatable-item.ts:31-35`, and its
  channel identity flows into `structural-plan.ts:43`). One compiled shape;
  discriminate at the input.
- **Owned-channel registration on `play()`** — ownership is a compile-time
  property published into the structural draft; a playback-time registry is a
  second system failing mid-flight.
- **Storing the closure** — opaque, unfreezable, undiffable.

## Deferred

- **Per-element live-channel bookkeeping** in `playMotionTracks` (cancel or
  throw on a contested live channel). A real gap today, unrelated to this
  extension — see "Honest motivation".
- **Adaptive sampling** (dense at impacts, sparse in free flight). The
  compiled shape already tolerates non-uniform offsets, so it stays a
  pure input-side addition.
