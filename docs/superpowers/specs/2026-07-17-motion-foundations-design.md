# Shared motion foundations

**Branch:** `motion-foundations`

## Purpose

Rationalize structural component animation and semantic delight effects by
sharing their lowest-level mechanics without merging their meaning, ownership,
or failure domains.

Components continue to explain visual continuity across authoritative state:
the same card moved, resized, rotated, appeared, disappeared, or changed face.
Effects continue to explain meaning: the transition was rewarding, dangerous,
magical, or worthy of attention.

The implemented tranches extract behavior-preserving geometry and timing
primitives already duplicated by FLIP, `animateBetween()`, and effect anchors,
then publish structural playback as immutable intentions and outcomes. They do
not replace the card animator or introduce a universal motion graph.

## Invariants

- Structural motion remains queue-critical; decorative effects remain
  non-gating and disposable.
- A failure in game-authored effect planning cannot prevent structural motion
  or state settlement.
- Card identity, sanitization, stack provenance, cloned fallback content, and
  faux-component reconstruction remain component-system policy.
- Effects never take ownership of a component's structural transform.
- Missing subjects, anchors, and optional decoration degrade to explicit
  skipped results.
- Companion timing, reduced-motion behavior, interruption, and settlement
  continue to flow through the existing `play()` primitive.
- Geometry solvers are pure, finite, and independently testable.

## Primitive layers

### 1. Timing compilation

A pure compiler turns an animation's requested timing, policy, current clock,
and optional version window into one of two values:

- a finite WAAPI timing plus the active version context and declared settlement
  budget; or
- an explicit timing skip when no visible version-slot budget remains.

It owns version-window validation, late joining, duration/end-delay clipping,
explicit local starts, and backwards fill while waiting. It neither starts an
animation nor decides whether that animation gates state settlement.

`BoardgameAnimatableItem.play()` remains the execution and ownership kernel.
The component animator and effect layer keep their different gating policies,
but consume the same compiled timing rules.

### 2. Geometry

One internal module owns:

- viewport-space and offset-parent-space rectangle capture;
- rectangle centers and center-to-center inversion deltas;
- finite scale-ratio calculation;
- FLIP inversion translation and scale;
- the threshold for deciding whether geometry visibly changed.

Viewport geometry is appropriate for overlays and explicit cross-root travel.
Offset geometry remains appropriate for the existing structural FLIP pipeline.
Sharing the representation and solvers does not pretend those coordinate
spaces are interchangeable.

### 3. Component continuity

The component animator retains:

- the before/install/after measurement transaction;
- stable component ID matching;
- `IDsLastSeen` source inference;
- cloned card content;
- faux component creation;
- component-specific visual-state capture and playback;
- staggering and queue-critical settlement.

### 4. Semantic effects

The effect layer retains descriptor compilation, themes, budgets, named anchor
scope, deterministic particles, cancellation, and reduced-motion substitutes.
It consumes shared viewport geometry but does not control structural motion.

## Structural-motion plans

The component animator emits a private, read-only solved plan. This is the
implemented shape, abbreviated only by omitting field comments:

```ts
interface StructuralMotionPlan {
  readonly source: 'flip' | 'explicit';
  readonly generation: number;
  readonly phase: 'planned' | 'executing' | 'settled';
  readonly segments: readonly StructuralMotionSegment[];
}

interface StructuralMotionSegment {
  readonly subjectId: string;
  readonly presence: 'retained' | 'appearing' | 'departing';
  readonly provenance: StructuralProvenance;
  readonly spatial?: {
    readonly offsetFrom?: OffsetGeometry;
    readonly offsetTo?: OffsetGeometry;
    readonly viewportFrom: ViewportGeometry;
    readonly viewportTo: ViewportGeometry;
    readonly inversion: FlipGeometry;
  };
  readonly transform?: { readonly before: string; readonly after: string };
  readonly properties: readonly StructuralPropertyChange[];
  readonly opacity?: { readonly before: number; readonly after: number };
  readonly timingRequest: StructuralTimingRequest;
  readonly execution: StructuralExecution;
}
```

Spatial, property, transform, and opacity changes are orthogonal rather than
collapsed into one exclusive `kind`. The spatial record carries both the
legacy FLIP offset coordinates and viewport coordinates captured during the
same measurement transaction. Property values that cannot be safely and
immutably represented are marked `opaque`; plans never retain arbitrary game
objects.

### Publication lifecycle

- `prepare()` begins a generation and immediately invalidates the previous
  solved plan.
- Measurement builds private mutable drafts. Drafts are not observable.
- After the second microtask, layout measurement, component `updateComplete`,
  and a final generation check, the animator creates an immutable `planned`
  intention with each segment's stagger and requested duration.
- The intention is installed immediately before any `playAnimation()` call.
  Playback returns the real WAAPI animations, replacing it with an immutable
  `executing` plan containing compiled timing or an explicit skipped outcome.
- Animation settlement replaces the segment outcome with `finished` or
  `cancelled`; when every segment is terminal the plan becomes `settled`.
- A new `prepare()` invalidates the plan even if the previous generation was
  interrupted. Stale async work may neither publish nor play.
- The plan remains diagnostic/internal after settlement until the next
  generation. It does not imply that an effect may consume it yet.

Spatial records carry both root-relative offsets for legacy FLIP playback and
historical viewport endpoints captured during the same measurement transaction.
No later projection attempts to reconstruct an endpoint after layout changed.

The animator exposes an internal subscription boundary for these immutable
plan revisions. Observation is read-only and non-gating: observers cannot
replace playback, and an observer failure is isolated from structural motion.
An interrupted generation emits a terminal cancelled revision before the next
generation invalidates the animator's current-plan slot.

A pure event compiler compares two revisions and emits only newly observed
segment statuses (`planned`, `started`, `skipped`, `finished`, or `cancelled`).
Event identity includes source, generation, segment index, and status. It never
fabricates an intermediate state if a consumer missed a revision. This is the
smallest durable seam for diagnostics and future decoration adapters; it is not
yet a game-author API.

Provenance is explicit. Retained subjects have exact identity continuity;
appearing and departing endpoints inferred from `IDsLastSeen` carry their stack
history evidence rather than pretending to be exact.

It is observational. Effects may decorate a matching movement with overlay
trails or arrival cues, but cannot cancel, replace, delay, or gate it. A missing
motion is an explicit skip, never an instruction to synthesize a second flight.

The likely public descriptor is deliberately explicit:

```ts
fx.decorateMotion({
  subject: fx.component(cardId),
  arrival: fx.burst({ tone: 'reward', intensity: 'small' }),
})
```

Independent subject travel remains a distinct recipe. Subject materialization
will require an opt-in component protocol so arbitrary DOM cloning cannot leak
hidden card content or depend on uncloneable shadow DOM.

## Physical animation ownership

Avoid a general runtime channel arbiter. Conflicts are prevented by DOM
ownership:

- component host: structural position and size;
- component-owned inner element: face and orientation changes;
- renderer/effect overlay: emphasis, particles, trails, and arrival cues;
- document overlay: screen treatments.

## Implementation ledger

1. **Done:** extract and test geometry capture and pure solvers.
2. **Done:** migrate structural FLIP, `animateBetween()`, and effect anchors.
3. **Done:** extract version/local timing calculation into a pure compiler.
4. **Done:** route component playback, explicit flights, and effects through
   that compiler without merging their ownership or gating policies.
5. **Done:** brand geometry snapshots by coordinate space.
6. **Done:** specify and implement the solved-plan publication barrier.
7. **Done:** publish immutable intentions, actual executions, skips, and terminal
   outcomes for FLIP and explicit flights.
8. **Done:** expose read-only plan observation and pure lifecycle-event
   compilation. **Not done:** connect those events to effect recipes.
9. **Next:** add point-only structural decoration (arrival/departure cues) that
   requires no component materialization and cannot gate playback.
10. **Deferred:** design a privacy-safe subject snapshot protocol before trails
    or travelers may visually reproduce a card/component.
11. Reassess a larger representation only when concrete duplication justifies
    it.

## Supported surface today

Game authors should use `effectsForTransition()`, `this.effects.play()`, named
anchors, and `animateBetween()` as documented in `docs/animation-effects.md` and
`docs/companion-mode-authoring.md`. Structural plans, observers, and lifecycle
events are framework-internal while their contracts settle. Shipping the
internal seam first prevents an attractive experimental API from becoming a
permanent compatibility burden.

## Non-goals

- Replacing FLIP with effect descriptors.
- Making effects queue-critical.
- Automatically choosing between decorating and synthesizing travel.
- Cloning arbitrary elements as effect subjects.
- General screen, audio, haptic, Canvas, or WebGL effect backends.
- Removing the current Lit microtask/frame measurement barrier in this tranche.
