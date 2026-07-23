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
  readonly visualSubject?: MotionSubjectSnapshot;
  readonly presence: 'retained' | 'appearing' | 'departing';
  readonly provenance: StructuralProvenance;
  readonly path?: {
    readonly kind: 'stationary' | 'travel';
    readonly from: ViewportGeometry;
    readonly to: ViewportGeometry;
  };
  readonly channels: readonly ComponentMotionChannel[];
  readonly timingRequest: StructuralTimingRequest;
  readonly execution: StructuralExecution;
}
```

The published plan is a capability, not a mirror of animator internals. Its
viewport path says where a safe subject was and is, and distinguishes stationary
morphs from travel. Its channel names say what the executor owns. Raw FLIP
offsets, inversion transforms, opacity endpoints, animating-property names, and
before/after property values remain private to planning and playback. This both
shrinks the consumer contract and makes it impossible for a decoration observer
to learn primitive hidden-state values merely because a component animates them.

### Component motion tracks

The smallest executable unit below a card flip is an immutable single-channel
track:

```ts
interface ComponentMotionTrack {
  readonly target: 'host' | 'visual';
  readonly property: 'transform' | 'opacity';
  readonly from: string;
  readonly to: string;
}
```

`host` belongs to structural position and visibility. `visual` belongs to the
component's inner presentation surface. A compiler combines framework-owned
host tracks with component-produced visual tracks, removes no-ops, copies and
freezes every endpoint, and rejects two writers for the same target/property
channel. It deliberately does not model arbitrary CSS or DOM selectors.

The animator now asks a component to plan this list once. The list determines
whether the component participates, is recorded as channel intent in the
structural plan, and is consumed unchanged by the shared WAAPI execution and
gating kernel. This replaces the former split where the animator guessed that
`faceUp` needed work and `boardgame-card` later imperatively started a separate
animation. Card face/rotation changes are now pure `visual:transform` track
production; FLIP remains `host:transform`, and fades remain `host:opacity`.

Track endpoints are execution descriptions, not game-author effect APIs.
Semantic effects still observe lifecycle/geometry and never acquire a write
channel on either component surface.

Reduced motion is resolved by the shared timing compiler as a complete visual
policy: requested delays, explicit durations, and synchronized version waits
collapse to zero, while semantic post-animation holds remain as `endDelay` so
state such as a matched pair is still readable before capture. An explicit
`animateBetween()` duration cannot bypass that policy. Decorative recipes that remain useful
without travel opt into their own short stationary opacity substitute before
calling the motion kernel.

The track-to-keyframe compiler and executor live one layer lower, on the common
animatable-item kernel. Both stack-managed cards and standalone dice therefore
use the same frozen endpoint representation, finite target resolution, timing
compiler, reduced-motion policy, completion gate, and settlement bookkeeping.
The Pig die's selected-face spin is a `visual:transform` track rather than a
direct `play()` call.

Execution resolves the finite target surfaces during planning and again at the
playback barrier. A missing optional visual surface drops its channel without
cancelling valid host travel. If a later WAAPI start fails, the executor cancels
the channels already started in that plan. Component planning is similarly
isolated: a throwing component visual hook is reported, then the animator
falls back to compiler-owned host transform/opacity tracks so its measurement
barrier and completion promise cannot be stranded. The established
`playAnimation()` array return remains compatible; its ordering is exactly the
immutable track ordering, which lets the animator attach execution timing to
the correct named channel rather than relying on anonymous array position.

Lifecycle scope remains honest. Card tracks are planned inside the animator's
measurement transaction and appear in structural plans. A die is a standalone
state-driven animatable item, so its local track is gated by the ambient version
context but is not fabricated into a FLIP plan and cannot satisfy `fx.motion()`
or `fx.trail()`. Unifying execution does not pretend the two have identical
provenance.

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

Private playback records retain the offset-space data required by FLIP. The
published capability retains only historical viewport endpoints captured in
that same measurement transaction; no consumer later reconstructs an endpoint
after layout changed.

The animator exposes one ordered, replayable lifecycle stream. A pure compiler
emits newly observed segment statuses (`planned`, `armed`, `active-observed`,
`skipped`, `finished`, or `cancelled`) followed by exactly one `generation-settled` marker
after every segment is terminal. Event identity includes source, generation,
segment index, and status. A late observer replays the current generation in
the same order. It never fabricates an unobserved intermediate state, and an
observer failure is isolated from structural motion. This lets decorators close
unmatched anchors directly on the settled marker; there is no second plan
observer or microtask ordering convention. The stream is read-only and
non-gating and is not yet a game-author command API.

### Point-only decoration adapter

The render host registers authoritative effect descriptors before starting the
new generation's structural playback. It first lets `prepare()` publish any
old-generation cancellation, then opens a fresh effect epoch; this prevents a
cancelled old animation of the same component from satisfying the new
transition's anchor.

`fx.motion(subjectId, 'departure')` resolves from the segment's captured
viewport `from` center on `active-observed`. The animator samples the actual
primary Animation timeline with one shared frame monitor; successful completion
synthesizes the observation immediately before `finished` if sampling missed a
short or backgrounded interval, while cancellation never fabricates it. The default
`fx.motion(subjectId, 'arrival')` resolves from its `to` center only on
`finished`. Skipped/cancelled motion and an absent subject settle the decorative
handle explicitly instead of leaving a pending promise. A settled plan closes
all unmatched anchors. The adapter consumes only automatic FLIP events;
explicit `animateBetween()` flights do not silently become author effects.

Motion anchors are accepted by point recipes (`pulse` and `burst`) but not by
`travel`. This is an intentional capability boundary: point decoration needs
only coordinates, while a trail requires the separate sanitized silhouette
capability described below.

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

Independent subject travel remains a distinct recipe. Arbitrary materialization
is still excluded so DOM cloning cannot leak hidden card content or depend on
uncloneable shadow DOM.

### Privacy-safe subject trails

`motionSubjectSnapshot()` is the narrow capability boundary for visual
subject-following decoration. Its first representation is intentionally only a
silhouette shape: rectangle, rounded rectangle, or circle. The animator copies
that value through an exact sanitizer before publishing it on a segment. Any
extra property rejects the whole snapshot, so content, artwork, DOM, styles,
and hidden game state cannot accidentally become observable.

`fx.decorateMotion({ subject, trail, departure, arrival })` is recursively
prepared before structural playback, even inside ordinary composition. Its
trail waits for the matching FLIP `armed` event. The overlay
then derives its endpoints from the segment's viewport geometry and its timing
envelope from actual compiled execution records. It owns only disposable echo
elements, consumes the shared visual-node budget, and is cancelled by the
matching source/generation/subject cancellation. It never owns the component
transform or gates structural settlement. Reduced motion substitutes a
stationary arrival pulse.

The trail deliberately has no independent timing policy: a follower that can
drift into a different version slot is not a follower. Artwork-bearing subjects
remain a future, separately reviewed protocol rather than a compatible field
added to the safe silhouette value.

When a sequence reaches a prepared motion decoration, it joins that existing
handle. It never replays a cached structural start after the subject has already
arrived. This is the semantic distinction between scheduled effects and
lifecycle subscriptions.

Reusable visual profiles remain typed game-local descriptor factories. The
existing semantic layers—recipe, tone, intensity, timing, theme, and
composition—already contain the durable policy. A framework-wide named preset
registry is intentionally deferred until multiple games prove an identical
semantic contract; sharing a keyframe shape alone is insufficient.

Structural cohorts begin one layer lower than group effects: a pure cadence
compiler overlays explicit ordered subject IDs onto compatibility stack delays.
The declaration owns no geometry and effects cannot influence it. Explicit
timing replaces legacy stagger for members; malformed or overlapping
declarations reject the whole explicit set and preserve legacy timing. This is
the narrow primitive shared by current stack cascades and future deal/gather
vocabularies.

The lifecycle now binds observations to source/generation/segment identity and
distinguishes armed playback from sampled activation. That supports one narrow
success barrier, `fx.afterMotion({ subjects, effect })`, over exact automatic
FLIP segments. True cohort identity, progress, and cohort-wide cancellation
remain deferred; the barrier is composition and never a structural owner.

Structural continuity is resolved before geometry or carrier creation:

`DOM sightings -> continuity intent -> geometry -> tracks -> cadence -> timing/execution -> lifecycle`

The continuity solver treats exact before/after identity as authoritative and
uses collection `IDsLastSeen` only as fallible endpoint evidence. Inferred
appearance or departure requires one uniquely newest collection other than the
known exact endpoint. Ties, duplicate exact sightings, malformed versions, and
same-collection-only history are unresolved and deliberately produce no
fabricated travel. Full history versions and candidate sets remain private;
public structural provenance contains only the selected collection and coarse
evidence.

Historical presentation and temporary embodiment are separate private
capabilities. A component may opt into capturing element clones of its already
rendered, visible default-slot light DOM; bare text, named slots, and
`dom-bind` are excluded. The runtime-branded token contains no public state,
and fresh clones are used for every installation. Identity-preserving legacy
cards retain their established `fallback` slot and document identity behavior;
new safe-mode components use the reserved `motion-history` slot, strip IDs and
focus hooks, and require the same host tag. A viewer-local cache scoped to the
animator/game-surface lifetime, refreshed only from exact visible subjects,
preserves last-visible artwork through hidden generations without publishing
it in structural plans. It
supports a retained host whose current rendering has become opaque, a
departing subject with no live host, and a later reappearance.

Only the latter requires a motion carrier. The destination collection creates
a fresh host from its ordinary component recipe, prepares it from closed
default property data, and places it in the generation-owned animation
container. That carrier has no subject DOM ID and is inert, pointer-free, and
`aria-hidden`; ordinary retained hosts are never made inert. Presentation stays
inside the animator/component boundary and is never added to structural plans,
events, silhouettes, or effects. Only an exact visible subject refreshes the
surface-scoped cache; hidden state and stack history can reuse that viewer-known
presentation but can never create or update artwork.

An inferred collection endpoint has a finite, collection-owned presence
policy: `scale-fade` or `travel-only`. The pure compiler turns that semantic
choice into frozen numeric scale/opacity facts. CSS serialization happens only
at the existing host-track composition seam. Neither an appearing live host nor
a departing carrier is mutated during measurement; the facts become ordinary
track endpoints and playback remains the first writer. This replaces the old
imperative `setUnknownAnimationState()` convention without inventing a general
CSS transform language.

Endpoint orientation is likewise a finite component fact: `natural` or
`quarter-turned`. The component type interprets its logical state, while the
result belongs to the captured before/after endpoint. Geometry exchanges the
source width/height axis only when those facts differ. This replaces the
card-shaped `animationRotates()` question with axis parity; it does not encode
rotation angle or parse authored transforms. Host FLIP correction and the
card's inner face/orientation track remain separate owners. The current scalar
solver deliberately assumes uniform scaling and stable intrinsic aspect ratio;
arbitrary aspect morphing would require an explicit future `scaleX/scaleY`
design.

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
8. **Done:** expose one ordered, replayable lifecycle stream with an explicit
   generation-settled marker.
9. **Done:** add `fx.motion()` point-only structural decoration for pulse/burst
   departure and arrival cues. It requires no component materialization, starts
   only from actual execution events, and cannot gate playback.
10. **Done:** add an exact, privacy-safe silhouette snapshot protocol and
    `fx.decorateMotion()`/`fx.trail()` follower that derives geometry and timing
    from real FLIP execution, prepares before playback, degrades under
    budgets/reduced motion, and cancels by generation.
11. **Done:** extract immutable component motion tracks, make channel ownership
    explicit, plan once for participation and playback, and migrate card
    face/rotation animation off its imperative WAAPI hook.
12. **Done:** move track-to-keyframe execution into the common animatable-item
    kernel and migrate the standalone Pig die spin, eliminating the last direct
    state-driven component `play()` bypass while preserving lifecycle scope.
13. **Done:** make target startup atomic, isolate planner/playback failures,
    redact structural plans to capabilities, label execution channels, enforce
    reduced motion for explicit timing, and keep trails on compositor channels.
14. **Done:** extract immutable cohort cadence scheduling, route compatibility
    stack stagger through it, add a generation-bound ordered-ID author hook,
    and prove ordering across independently rendered stacks.
15. **Done:** replace subject-keyed outcomes with stable segment refs and
    monotonic index-keyed updates; distinguish armed playback from sampled
    `active-observed`, bind decorators to exact refs, and add the success-only
    FLIP `fx.afterMotion()` barrier.
16. **Done:** extract a pure structural continuity solver, make inferred
    endpoints permutation-invariant, and fail closed instead of animating an
    arbitrary collection when history is ambiguous.
17. **Done:** replace animator-owned card cloning and faux-card preparation
    conventions with private historical-presentation capture plus fresh inert
    motion carriers; preserve retained and departing artwork without widening
    public motion snapshots.
18. **Done:** replace hard-coded unknown-host mutation with a semantic
    collection presence policy compiled to numeric endpoint facts and ordinary
    host tracks.
19. **Done:** replace the card-only rotation boolean with explicit endpoint
    axis-orientation facts consumed by the pure geometry solver.
20. **Done:** extract a pure retained-carrier viewport-flight compiler; replace
    argument-order direction with `fly({ subjectId, source, carrier })`, wait on
    registered render ownership rather than a fixed frame count, preserve the
    carrier's computed resting transform, and make explicit-flight interruption
    and concurrent event histories terminal and generation-specific.
21. **Done:** add atomically validated, generation-bound, transition-local
    transfer declarations for retained carriers; scope endpoint
    lookup to one renderer, settle the batch with FLIP, reject channel-owner
    conflicts, and correct transition hooks to the real viewer-sanitized
    `AnimationKey` wire contract.
22. **Done:** partition transfer ownership from exact identity facts; consume
    after-only stack declarations into automatic FLIP, separate declared
    spatial origin from safe stack-default semantic pose, flow declared timing
    through every host/visual track, and migrate Hand arrivals off imperative
    lifecycle/frame inference. Viewport deltas fail closed under transformed
    ancestors pending a reviewed affine projection primitive.
23. **Done:** add a generation-bound structural progress barrier for new
    renderers. A pure declaration selects exact subjects or all armed
    FLIP/transfer primaries; one WAAPI sampler observes executed active
    intervals; opaque install-cycle IDs make progress release and settlement
    idempotent and stale-safe. The contract is named as destructive
    buffered-queue cutover, not concurrent generations. The successor-aware
    `animationOverlap()` state-clock contract remains as a compatibility lane
    and takes precedence when overridden; it cannot be represented by a
    current-transition-only structural declaration.
24. **Compatibility correction:** authored Table/Hand choreography uses pure
    transition-local transfer declarations, but the zero-author defaults retain
    their master behavior through `animateBetween()`: local baselines, concurrent
    final-pose Hand arrivals, and decorative ungated Table stubs. Moving those
    defaults onto the richer lifecycle is a visual change and requires explicit
    product approval.
25. **Deferred:** replace local, lossy Table/Hand inference with a
    privacy-reviewed server transfer envelope when games need zero-author
    defaults or true cross-surface correlation. A source-carried departure also
    requires a detached carrier; `fly()` deliberately does not claim it.
26. **Deferred:** design any artwork-bearing subject representation as a new,
    explicitly reviewed capability; do not widen the silhouette snapshot.
27. Reassess a larger representation only when concrete duplication justifies
    it.

## Supported surface today

Game authors should use `effectsForTransition()`,
`motionCohortsForTransition()`, `motion.stagger()`,
`motionTransfersForTransition()`, `motion.transfer()`,
`motionReleaseForTransition()`, `motion.release()`, `this.effects.play()`,
named anchors, `fx.motion()`, `fx.decorateMotion()`, `fx.trail()`, and
`fx.afterMotion()`, plus retained-carrier `fly()` for local imperative
presentation as documented in
`docs/animation-effects.md` and
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
