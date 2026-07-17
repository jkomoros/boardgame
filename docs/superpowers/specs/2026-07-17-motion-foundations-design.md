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

The first implementation tranche extracts behavior-preserving geometry
primitives already duplicated by FLIP, `animateBetween()`, and effect anchors.
It does not replace the card animator or introduce a universal motion graph.

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

### 1. Execution and scheduling

`BoardgameAnimatableItem.play()` remains the execution kernel. A later tranche
may extract a scheduler service, but only after both current callers can use it
without weakening component-owned settlement or queue participation.

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

## Structural-motion manifest

The next useful bridge is a read-only manifest emitted after the component
animator solves a cycle:

```ts
interface StructuralMotion {
  readonly subjectId: string;
  readonly kind: 'move' | 'appear' | 'depart' | 'morph';
  readonly from: GeometryRect;
  readonly to: GeometryRect;
  readonly delayMs: number;
  readonly durationMs: number;
}
```

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

## Staged implementation

1. Extract and test geometry capture and pure solvers.
2. Migrate structural FLIP, `animateBetween()`, and effect anchors to them.
3. Keep all browser behavior and timing unchanged.
4. Expose a private solved-motion record from the component animator.
5. Add regression tests for movement classification and manifest lifetime.
6. Add effect-only observation/decorating without subject materialization.
7. Design a privacy-safe component subject snapshot protocol.
8. Reassess whether a larger internal motion representation is justified by
   concrete duplication; do not introduce one speculatively.

## Non-goals

- Replacing FLIP with effect descriptors.
- Making effects queue-critical.
- Automatically choosing between decorating and synthesizing travel.
- Cloning arbitrary elements as effect subjects.
- General screen, audio, haptic, Canvas, or WebGL effect backends.
- Removing the current Lit microtask/frame measurement barrier in this tranche.
