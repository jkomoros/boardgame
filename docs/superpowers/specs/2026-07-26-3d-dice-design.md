# 3D Dice — Design

Date: 2026-07-26
Status: awaiting review
Companion: `2026-07-26-dice-roll-treatments.md` (the 16-treatment catalog this design is calibrated against)

## Goal

Replace the die's 2D slot-machine reel with a real 3D cube that rolls with
genuine physics and lands on the server-decided value, without leaving the
animation architecture that the parity harness protects. Design the API so
the common case is one element with no configuration, and the elaborate
treatments in the catalog remain reachable later.

## Non-goals for this slice

- Author-defined containers (invisible bounds, rendered trays) — designed
  for, not implemented.
- Lift-to-foreground staging — explicitly the cliff; its own later slice.
- Rendering polyhedra other than the cube — the architecture is
  shape-parameterized from the start (see "Generalizing to other dice"), but
  only the cube's face geometry ships in this slice.
- Symbol faces — the client can already render arbitrary face content, but
  the server model cannot *say* a face is a symbol (`dice.Value` is
  `Faces []int`, carrying no shape, symbol or colour). That is a Go-side
  change and is out of scope here.
- Multi-die choreography beyond simultaneous rolls.

## Why only three axes are new

The catalog's eight capability axes decompose unevenly. Four of them —
container/collision surface, rest & persistence, post-settle interactivity,
handoff to other components — are already served by the stack/zone system: a
tray that consumes a stack inherits rest positions, persistence, per-component
actions and stack-to-stack migration with no new machinery. The genuinely new
surface is:

- **A — face presentation**: pips/numerals × reel/cube.
- **C — staging**: inline (this slice) vs placed container vs lifted overlay.
- **F — choreography**: how N dice share a time budget.

The API should therefore stay small, and should not reinvent containment.

## Architecture

### Rendering: CSS 3D on the visual channel

A cube of six faces in a `transform-style: preserve-3d` carrier, with
`perspective` on the wrapper — exactly the pattern `boardgame-card` already
uses for its flip (`boardgame-card.ts:65,88`), which is the only 3D in the
client today and is already pinned by a geometry golden.

The 3D transform lives on the **`visual` channel (`#inner`)**, never on the
host. This is not stylistic: the host's `transform` has up to three owners
(the stack's `layoutTransform` setter, the animator's FLIP host track, and
`playAnimation`'s final write), `composite: 'replace'` means the
last-started animation wins outright rather than blending, and the
`layoutTransform` setter's "did the computed transform change?" probe reads
`getComputedStyle`, which a live host animation shadows. A tumble on the host
would make genuine relayouts silently no-op and would be mutually destructive
with FLIP. The channel compiler enforces one owner per `target:property`
(`component-track.ts`), so the split is checkable.

No canvas, no WebGL: a rAF render loop emits no WAAPI animations, so it is
invisible to the geometry sampler, contributes no motion curves, emits no
`play`/`settle` hooks, and never holds the completion gate. That would delete
regression coverage for the most visually complex thing in the app.

### Simulation: pure, hand-rolled, offline of the DOM

A new pure module (`src/motion/dice-sim.ts`) with no DOM dependency:

```
simulateRoll(config: RollConfig): Trajectory
  RollConfig  = { seed, dieCount, dieSize, bounds, energy, gravity, restitution, friction }
  Trajectory  = { durationMs, dice: Array<{ samples: Array<{ t, position: Vec3, orientation: Quat }>, restingUpFace: FaceNormal }> }
```

Boxes against an axis-aligned container, impulse-based collision response
with restitution and friction, a real inertia tensor so tumbling reads
correctly. Dice collide with each other. Being pure and seeded makes it
unit-testable without a browser and deterministic across surfaces.

### Outcomes by relabeling, not by re-rolling

Rejection-sampling seeds until the dice land on the wanted values costs
`sides^dice` simulations — ~8,000 for three d20s, unbounded worst case. It is
the wrong algorithm.

Rigid-body dynamics depend on geometry and mass distribution, not on which
pips are printed where. So: simulate **once**, read each die's resting
orientation, compute which body-space face normal ends up presented, and
assign the predetermined value to that face slot — laying out the remaining
values in a valid standard arrangement (for a d6: opposite faces sum to 7,
fixed chirality; exactly four rotations about that axis qualify, so one always
exists). Zero retries, exact every time, any number of dice with any face
count, full inter-die collisions preserved.

This is undetectable: every roll shows a legitimately-arranged die, and the
only thing that varies between rolls is the die's initial orientation — which
is exactly what varies when you shake dice in your hand.

**Residual**: a die settling on an edge or corner with no face presented.
This is a property of the trajectory alone, *uncorrelated with the desired
value*, so re-simulating is bounded and cannot loop on a particular outcome.
Detect it by testing the resting orientation against a tolerance; a
well-damped sim settles flat.

### Baking: trajectory → WAAPI keyframes

The trajectory is sampled into a `matrix3d` keyframe list and handed to
`play()`. The physics engine authors the motion; it never renders it.
Playback is therefore native WAAPI, which buys — for free — harness
visibility, gate participation, reduced-motion handling, and correct
interruption.

The geometry sampler passes 16-component `matrix3d` through untouched
(`geometry-helpers.ts` `parseMatrix`), so a baked tumble is fully measurable.

### Determinism

Seed derived from `(component ID, state version)`. Consequences: every
surface in companion mode shows the identical roll without needing
synchronized timing; a replayed or re-rendered state reproduces the same
motion; and parity goldens are stable rather than recording one arbitrary
sample of a random process.

### Timing and the gate

- **`timing: 'immediate'`.** The `'version'` policy clamps to the server's
  slot (600ms inside an 800ms slot). Critically, `animationContext` is null in
  solo play, so a 1.5s roll would look correct solo and be silently truncated
  **only in companion mode** — an easy bug to never notice. `'immediate'`
  avoids the clamp; determinism (above) provides the cross-surface agreement
  that `'version'` would otherwise have given.
- **Gated, blocking until settled** (decided): the roll holds the completion
  gate, so move buttons stay disabled and the next state waits.
- **Declare the full duration on the first gated play of the cycle.**
  `will-animate` fires only on the 0→1 gated transition, so a wind-up
  animation preceding the tumble would leave the watchdog armed at its 4s
  floor and force-close mid-air. The roll must be a single play, or the first
  one. A watchdog firing fails the trace suite unconditionally.
- This will be the **first feature to exercise the `expectedSettleMs`
  watchdog extension in a browser** — the parity README lists it as an
  accepted blind spot owned only by unit tests. It ships with an e2e witness.

### Reduced motion

The kernel resolves reduced motion to `duration: 0` with `fill: 'none'`, so
an animation-carried orientation would render nothing — the same trap that
made the token throb vanish. The landed orientation must therefore come from
**resting style**, not from animation fill: the die writes its final
orientation to `#inner`'s style and the tumble animates from a start pose to
that same resting pose. Under reduced motion the die simply appears landed,
and announces its value.

### Degradation ladder

For face counts or shapes the cube renderer cannot express, fall back to
today's reel rather than throwing. Deliberately *not* following the
`unknown stack layout` throw precedent: graceful degradation is what makes
"exotic is possible" affordable, and a d20 rendering as a numeral reel is
strictly better than an exception.

## Generalizing to other dice

The point of the architecture is that a d20 is *data*, not a rewrite. Three
things have to be shape-parameterized from the start, or they calcify around
the cube:

**One geometry table drives everything.** A die shape is described once as
`{ vertices, faces: [{ normal, centroid, polygon }], inertiaTensor }`. Both
the simulator and the renderer consume that same table — the CSS face
transforms are *derived* from each face's normal and centroid rather than
hand-authored per shape. Adding a d8 means adding a geometry entry, not
touching either subsystem.

**Vertex-based contact resolution.** The simulator must not special-case box
faces. Detecting penetration per *vertex* against the container planes and
applying the impulse at that contact point is both simple and shape-agnostic:
the identical code tumbles a cube, an octahedron and an icosahedron once the
vertex list and inertia tensor change. Writing it box-specific first is the
one decision that would make d20 a rewrite, so it is ruled out now even
though this slice only ships a cube.

**Relabeling holds for every standard die.** d4/d6/d8/d12/d20 are Platonic
solids and the d10 is a pentagonal trapezohedron — all face-transitive, so
any face can carry any value and the "simulate once, paint the wanted value
on whichever face lands up" solution is valid for all of them. Numbering
conventions loosen as face count rises (the d6's opposite-faces-sum-to-7 rule
has no universal d20 analogue), which makes the constraint solver *easier*,
not harder. One genuine quirk: a d4 rests on a face and is conventionally
read from the top vertex, so its "presented value" lookup differs from every
other die and needs an explicit per-shape reading rule rather than a shared
"which normal points up" assumption.

**Rendering cost is per-shape but bounded.** A cube's faces are rectangles;
a d12's are pentagons and a d20's are triangles, cut with `clip-path`. That
is real per-shape work, but it is confined to the geometry table and does not
touch the sim, the baking, the gate integration or the harness.

## Face content: pips, numerals, and beyond

Faces are DOM elements, so their content is free-form and this dimension is
largely already solved:

- **Numerals already work today.** The numeral `<span>` is suppressed by CSS
  only for the classes `one`…`six`, so any other face value already renders
  as a number — which is exactly what a d20 wants (nobody prints 20 pips).
- **Pips** stay the default where they read well: a six-or-fewer-sided die
  with consecutive values from 1.
- **Automatic selection, author override.** Pips for the classic d6 case,
  numerals otherwise, with an explicit knob for authors who want numerals on
  a d6 (or pips on something unusual).
- **Arbitrary content — including symbols — is a client non-issue.** Because
  a face is a DOM element, slotting an icon into it needs no new client
  machinery. The blocker for symbol dice (Catan event dice, Warhammer
  hit/wound) is purely that Go's `Faces []int` cannot express "this face is a
  wheat icon", and that `Value` doubles as the accessible human-readable
  result. That is the server-model seam named in the non-goals, and it is
  worth fixing *once* for both symbol faces and shape metadata rather than
  twice.

## API surface (this slice)

```html
<boardgame-die .item=${die}></boardgame-die>
```

Zero configuration renders a cube, tumbles in place for ~800ms, stays inside
its own border box, announces the value to a live region, and snaps under
reduced motion.

Designed-but-deferred, named now so they are not tripped into accidentally:

- `stage="inline | container | lift"` — staging is its own attribute
  precisely because it is a cliff, not a slope. `inline` is the only value
  implemented in this slice.
- `energy` — feel knob (lazy tumble → vigorous throw).
- Roll **budget** at the group level rather than per-die `duration`: three
  Yahtzee rolls a turn and fifteen Warhammer dice at once push opposite
  directions, and per-die timing breaks both. Individually-simulated dice cap
  around a dozen, with shared trajectories beyond.

Also fixed in passing: `--effective-die-size` is hardcoded at 50px, so pig's
`.die { height: 100px }` is dead code. Size becomes a real, documented knob.

## Ownership partition (written down before coding, per the catalog's tension #2)

| Thing | Owner |
|---|---|
| host `transform` (resting layout) | stack's `layoutTransform` setter |
| host `transform` (structural motion) | the animator's FLIP host track |
| `#inner` `transform` (3D orientation + flight) | the die itself, via one `visual:transform` track |
| resting orientation style | the die, written outside the animation |

The sim never borrows the host transform in this slice, because inline
staging never leaves the die's border box. When containers land, the tray
owns resting slots and the sim borrows the flight — that partition gets
written down then, not improvised.

## Testing

- **Unit (no browser)**: sim determinism for a fixed seed; energy
  conservation bounds; settling detection; the relabeling solver (every
  desired value is reachable from any resting orientation, and the produced
  arrangement is always a valid standard die).
- **Geometry golden**: the tumble curve. Two harness hazards to design
  around: a net rotation that is a whole multiple of 360° produces an
  identical start and end matrix, so the curve is dropped as noise — the
  landing orientation must be net-different; and multi-turn spins alias under
  5-point sampling, so the golden must not depend on distinguishing 360° from
  720°.
- **Trace golden**: pig's existing `pig-roll` golden will shift (structural
  mode, so it tolerates the change, but it must be re-recorded deliberately
  with the diff explained).
- **E2E witness**: the watchdog-extension path described above.
- **Reduced motion**: the die shows the landed face with no tumble.
- **Interruption**: force-settling mid-roll (cycle sweep, orphan settle, the
  geometry sampler's own `finish()`) leaves the correct face presented.

## Risks

1. **Hand-rolled angular dynamics are fiddly.** Convincing tumbling needs a
   correct inertia tensor and contact-point impulses. Mitigation: the sim
   lives behind a narrow interface (config in, trajectory out), so it can be
   replaced with a real engine without touching rendering, baking or gate
   integration.
2. **Keyframe volume.** A 1.5s roll at 60fps is ~90 samples per die. Needs
   measuring: sample-rate reduction with interpolation may be required, and
   the geometry sampler seeks within these.
3. **Perspective through ancestors.** `preserve-3d` is flattened by an
   ancestor `transform`, `filter`, `opacity < 1` or `overflow`. Inline
   staging keeps the 3D context inside the die's own subtree, which avoids
   this — but it must be verified against real game pages, since every
   candidate ancestor in this app clips.
