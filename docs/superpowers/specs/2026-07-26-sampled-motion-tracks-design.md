# Motion Curve Tracks — Design

Date: 2026-07-26
Status: revised after API critique; hazard review pending
Motivation: `2026-07-26-3d-dice-design.md` needs to play a physically-baked
trajectory, which today's two-endpoint motion track cannot express.

## The gap

A `ComponentMotionTrack` is an endpoint pair (`component-track.ts:11-23`)
compiled into two WAAPI keyframes. The channel compiler enforces the system's
central invariant — **one owner per `target:property` channel** — and throws
on contention (`:73`).

Motion that is not "interpolate A→B under one easing" therefore has no route
through the track system. Today that is a baked dice trajectory; tomorrow any
path-following, multi-bounce, or simulation-authored motion. The shortcut —
calling `play()` with a raw keyframe list — **silently forfeits channel
ownership**, because ownership is established during *planning*
(`structural-plan.ts:101-115` publishes the channel set before anything
animates), not at playback.

## The reframing

The first draft of this design was "a track can carry more keyframes." That
is wrong, and it rested on a factual error: per-keyframe `easing: 'linear'`
is a **no-op**, since WAAPI keyframe easing already defaults to linear. The
distortion a sampled trajectory suffers comes from the *effect-level*
`easing: 'ease-in-out'` the kernel injects (`boardgame-animatable-item.ts:290`),
which warps iteration progress before the sample list is ever indexed.

So the real statement is: **a track can claim its channel's _timeline_, not
just its values.** An ordinary track takes its character from the effect's
easing; a curve track already encodes character in its own shape and
therefore owns time on that channel too. Two time warps on one channel is
structurally the same error as two owners on one channel, and earns the same
loud throw.

## The API: curve in, samples out

The producer supplies a **function of progress**; the compiler owns the
sampling grid and evaluates the curve to frozen data, never retaining it.

```ts
export interface ComponentMotionSample {
  readonly offset: number;   // WAAPI's word — emitted verbatim into a Keyframe
  readonly value: string;
}

export interface ComponentMotionTrack {
  readonly target: ComponentMotionTarget;
  readonly property: ComponentMotionProperty;
  readonly samples: readonly ComponentMotionSample[];  // always >= 2, uniform, spanning [0,1]
  readonly timeline: 'eased' | 'sampled';
  readonly resting?: string;   // authored resting style; see reduced motion
}

type MotionEndpointsInput = Readonly<{ from: string; to: string }>;

type MotionCurveInput = Readonly<{
  curve: (progress: number) => string;  // pure, cheap; called `resolution` times at compile time
  resolution?: number;                  // clamped to [2, 256], default 64
  resting?: string;                     // default: curve(1)
}>;
```

`curve` and `progress` are this system's existing vocabulary — the parity
harness's whole ontology is *motion curves* (`geometry-helpers.ts:6,22-45`)
and `progress` already means "fraction of the active interval"
(`release.ts:4-5`). `offset` is WAAPI's, because it is emitted verbatim.

### Why the compiler owns the grid

The earlier draft let producers author offsets, then spent five invariants
policing them (range, monotonicity, endpoint anchoring, minimum count,
maximum count). Every one exists *only* because the producer was handed a
grid it should never have had. A compiler-generated uniform grid deletes all
five. A validator you can delete beats a validator you write well.

This is also the module's established split: inputs are requests, compiled
tracks are normalized facts (`exactTrack`, `component-track.ts:36-60`), the
same shape as `compileMotionPresence` (`presence.ts:8-19`) and
`motionRelease` (`release.ts:39-70`).

A closure is the right *authoring* boundary and the wrong *storage*: it can
capture DOM, is non-deterministic, cannot be meaningfully frozen, and cannot
be diffed by a golden — everything this system refuses to carry across
boundaries (cf. `sanitizeMotionSubjectSnapshot`, `subject.ts:26`). Hence:
function in, frozen samples out.

### Endpoint form rewrites to two samples

One representation everywhere. `{from, to}` compiles to samples at offsets 0
and 1, which is a WAAPI no-op for a two-keyframe list (even distribution
yields the same offsets). Only `component-track.test.ts`'s `deepEqual`
literals churn — a one-time unit-test edit, not a permanent dual code path.

### Easing lives at the effect level and is contested loudly

A `timeline: 'sampled'` track pins effect easing to `linear`, and
`playMotionTracks` **throws** if a caller also supplies `timing.easing` for
that channel. A producer wanting eased physics bakes it into the curve, in
one inspectable place.

Effect level, not per-keyframe — and not only because per-keyframe linear is
a no-op. `boardgame-component-animator.ts:553-566` reads observability
straight off `effect.getTiming().easing` into `StructuralExecutedTiming`
(`structural-plan.ts:52-60`). Pushing easing into keyframes would make that
field report `'linear'` for every animation in the system and stop meaning
anything — the same class of mistake as a rAF loop that emits no WAAPI
animation.

### `resting` makes reduced-motion correctness structural

Under reduced motion the kernel returns `duration: 0` with `fill: 'none'`
(`timing.ts:102-115`), so the effect is out of effect immediately and
computed style reverts to authored resting style: **the last sample is never
applied.** The compiler cannot assert its way out of this — it has no idea
what resting style says.

`boardgame-component.ts:356-357` already writes final resting style after
playback, but *for the host channel only*, which is why `boardgame-card.ts`
maintains `_updateInnerTransform` by hand. Promote it to the track: a curve
track carries `resting` (default `curve(1)`), and `playMotionTracks` writes
it to the resolved target after starting the animation. Reduced-motion
correctness becomes structural rather than per-producer discipline, and a
`resting` that intentionally differs from the last sample becomes explicit
rather than a silent divergence. Endpoint tracks default to `undefined` (no
write), leaving card and host behavior untouched.

### A constant curve throws

The `from === to` elision (`component-track.ts:70`) exists because the FLIP
compiler routinely emits zero-delta endpoints
(`compileComponentMotionTracks`). Nothing routinely emits a constant
trajectory except a broken producer — and note the elision runs *before* the
channel is claimed, so an elided track silently vacates its channel. A
constant curve is exactly the failure that should be loud.

### `resolution` is clamped, not capped

A quality knob, clamped to `[2, 256]` with default 64 — the precedent being
opacity clamping (`component-track.ts:52`), with throws reserved for
structural and ownership errors. Producers derive it from the duration they
are already passing to `play()`, which resolves the track compiler being
duration-blind.

On cost: ~90 `matrix3d` keyframes is a setup and memory cost (~1.4 KB per
die), not a per-frame one — transform animations run on the compositor. The
scaling wall for dice is the one the dice design already names: twenty
composited faces per d20 inside `preserve-3d`, times N.

## Required harness change (in scope for this spec)

`geometry-helpers.ts:179-198` normalizes the transform channel as
`matrixDist(p, first) / matrixDist(last, first)` — it **assumes monotone
travel**. A tumble's intermediate distance vastly exceeds its net distance,
so recorded values run ≫ 1 and the `0.08` tolerance (`:237-240`) is applied
on a scale it was never calibrated for: brittle goldens rather than wrong
ones. Worse, a near-whole-turn net rotation drives `totalMatrix ≈ 0`, and the
curve is dropped entirely as all-null noise.

The dice design half-noticed this (it warns about 360° multiples) but not the
normalization blowup. Since it is the sampled track that breaks the
assumption, fixing it is scoped here — normalize by **path length** (sum of
successive sample distances) rather than net displacement, or record an
explicit per-curve scale.

## Rejected alternatives

- **Discriminated union at the compiled track level** — makes `from`/`to`
  conditionally present on a *published* artifact (it rides in
  `MotionTrackPlayback.track`, `boardgame-animatable-item.ts:31-35`, and its
  channel identity flows into `structural-plan.ts:43`), forcing every future
  consumer to branch. One compiled shape; discriminate at the input.
- **Owned-channel registration on `play()`** — ownership is a compile-time
  property of a plan, thrown during planning and published into the
  structural draft. A playback-time registry is a second ownership system
  whose failures surface mid-flight.
- **Storing the closure** — see above.

## Deferred

Adaptive sampling (dense at impacts, sparse in free flight) could encode a
bounce in ~25 samples instead of ~90. Not worth reintroducing producer-owned
offsets for now; the compiled shape already tolerates non-uniform offsets, so
it stays a pure input-side addition later.
