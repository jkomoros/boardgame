# 3D components (Stage 2) — Design

Date: 2026-07-27
Issue: #801. Follows #800 (3D dice), precedes #802.
Status: designed from three investigations and one decisive experiment, all measured.

## The decision, up front

**Convex shapes become real solids. Non-convex shapes keep their authored art and gain a
3D presentation.** The boundary is not taste — it is where CSS stops rendering correctly,
and it was measured.

| shape | 3D as | why |
|---|---|---|
| `cube` | real solid | convex; `dieGeometry(6)` verbatim, zero new geometry |
| `token` | real solid | convex; flat-capped prism, h ≈ 0.55d |
| `chip` | real solid | convex; flat-capped prism, h ≈ 0.13d |
| `disc` | real solid | convex; flat-capped prism, and the only asset already drawn face-on |
| `meeple` | authored art + 3D presentation | **non-convex** |
| `pawn` | authored art + 3D presentation | **non-convex** (the neck is a genuine concavity) |

## Why non-convex shapes are not solids

The renderer's entire hidden-surface removal is `backface-visibility: hidden`. That is
*provably* sufficient for a convex solid — front-facing facets tile the silhouette exactly
and never occlude one another — and provably insufficient otherwise. CSS `preserve-3d` has
no z-buffer; it sorts by painting rules.

A meeple prism was rendered under the die's exact conditions and diffed against a Node
z-buffer rasterization under an identical camera, 16 randomized poses per tilt band,
antialiasing seams excluded two ways. Controls: a **d20 scored 0 wrong pixels across 253
poses** (so the harness is sound), and a deliberately-mispainted positive control flagged
79%.

| tilt | worst wrong % | worst 2px-thick region | self-occluded % |
|---|---|---|---|
| 0–15° | ≤0.057 | **0 px** | 2–3 |
| 20° | 0.170 | 1 px | 4 |
| 35° | 0.226 | 6 px | 11 |
| 50° | 0.594 | 38 px | 16 |
| 75° | **2.844** | 115 px | 28 |

A deliberately hostile "comb" prism is clean to 60° and then **collapses at 75%: 12.1% of
the silhouette wrong**, painting the interiors of its slots straight over the near teeth.
So the failure is **not monotone in tilt** — you cannot interpolate your way to safety.

**Explicit `z-index` sorting does not fix it, and that closes the last escape route.**
Far-to-near ordering measured identical to doing nothing; *near-to-far* — definitionally
the wrong painter's order — measured **better** (1.208% vs 2.844%). Chromium runs its own
sort inside a 3D rendering context and `z-index` only perturbs its tie-breaks. Layer
promotion is irrelevant to static ordering: `will-change` vs none was pixel-identical
across all 253 poses.

**The decisive argument is not that the prism fails — it is where it succeeds.** At 20°
tilt only 4% of the silhouette has two front-facing facets over it. The band where a prism
renders correctly is exactly the band where a prism *looks like a tilted flat card with a
thin rim*. Thirty-four DOM elements buy a picture that one authored SVG plus a tilt, a rim
and a contact shadow reproduces — with zero sorting risk at any angle.

Recorded so this is not re-litigated: there is no cyclic overlap in a prism over a simple
polygon (0 cycles in 120 random orientations), so a correct painter's order always exists.
The obstacle is not the geometry. It is that the browser will not sort to it and gives us
no hook to make it.

## What the geometry module needs

The generator's real acceptance predicate was found, and it is **not** "convex faces" as
previously believed. A prism over a simple CCW polygon is accepted iff *(a)* every side
wall's `dot(normal, centroid − vertexMean) > 0` gets the sign right, and *(b)* each cap's
vertices sorted by angle about its own vertex mean come back in polygon order. Both live in
`orientLoop` (`die-geometry.ts:246-268`) and both are independent convexity assumptions.
Verified against 600 random simple polygons at **100% agreement** with the shipped code,
and 167 accepted cases have non-convex cap faces — so non-convex *faces* were never the
constraint.

Everything downstream is already correct for non-convex input, measured: the
signed-tetrahedron volume/centroid/inertia agrees analytically to **1e-16**, the fan
triangulation's out-of-polygon triangles cancel exactly (3.3e-16 against shoelace), and
`validateClosedSurface` is pure combinatorics.

So: **add an opt-in to `RawSolid` for author-ordered loops** and skip the rederivation. It
is a small seam, it is what the convex prisms need anyway, and it leaves the door open if
the browser ever gains a real sort. It must not change any shipped die by a ULP.

## Where the scene lives, and why the stack is not the risk

`#inner` is the carrier. It is what `motionTrackTarget('visual')` returns, and for a token
it is **completely unwritten** — tokens declare no `animatingProperties` and no
`propertyMotionTracks`, so only host tracks are ever compiled.

**FLIP is not a 3D hazard.** Measured in real games, not fixtures: a `preserve-3d` scene on
a token's `#inner` held its projection ratio at exactly 2.000 across a 74-frame grid swap
(68 tokens/frame) and a 69-frame pile swap including reparenting — **zero flattened frames
in either**. Ancestor `scale`, non-uniform scale, ancestor `filter`, and a full host
`translate/rotate/scale` flight all measured 2.000. The reason is structural: the scene is
projected inside `#inner`'s own space and flattened at its boundary; everything the stack
and animator do happens above that boundary and is plain 2D composition.

Two honest consequences: a scaled flight renders a scaled *raster* of the solid rather than
re-rendering it, and motion carriers are permanently `noAnimate`, so a departing token can
display a solid but never animate one.

## The actual biggest risk: the resting pose on a pooled host

A stack **pools and reparents component hosts** across membership changes. Nothing
re-derives a pose on reuse. So the same DOM node carries whatever pose the *previous
occupant's* last write left on it — the exact failure the die's `_clearRoll` exists to
prevent, multiplied by pooling, faux components, and motion carriers that can never
self-correct by playing anything.

Measured precedent: a stale write to a card's `#inner` transform mid-flight was invisible
for the whole flight and became **permanent** the instant the animation ended, leaving the
card face-up and stuck at 45°.

**Therefore the pose must be a pure function of current state, re-derived on every render,
and explicitly cleared on orphan/recycle.** Not written once and remembered.

## Sizing

The two models collide and the token's must win. A token's `#inner` box **is** its drawn
extent (30px default, verified across every shape and layout); the die's `--die-size` is a
**bounding-sphere diameter** with a separately reserved footprint. A 3D cube dropped in at
the same scale would render **~40% smaller** than the SVG it replaces (a cube's face spans
1/√3 = 57.7% of its circumsphere) in checkers, tictactoe and pass.

A stack-hosted component also **cannot reserve extra space**: stack margins, the board
layout's `aspect-ratio: 1` clamp, the hardcoded `100px` spread/fan margins and the FLIP
scale ratio all key off that box.

So 3D tokens size by **drawn extent**, not by circumsphere — the solid is scaled so its
silhouette fills the existing box. No `#scaler`, no reserved overflow, no new layout
contract.

## What tokens do NOT inherit from dice

- **No physics.** Tokens do not tumble. The simulator's narrow phase is a convex
  half-space test that, on a meeple, intersects to the **empty set** — two meeples would
  pass through each other. Not a small extension; explicitly out of scope.
- **No layer promotion.** `will-change: transform` on facets exists so promotion is in
  place *before a roll starts*; a static token has no such need, and declining it means a
  3D token promotes no layers at all. **(Half wrong — corrected below. Declining
  `will-change` does not decline promotion: a live `preserve-3d` context is promoted
  wholesale the moment an ancestor animates, and a stack's FLIP animates the host on
  every move. A token has no live 3D context now.)**
- **No face content**, no reading pose, no `readingRule`, no `capFaces` semantics, no
  legibility floor. Tokens carry no marks.

### The facet budget — a second wall, under the layer cliff

> **SUPERSEDED 2026-07-28 — see "The wall was the layer cliff after all" below.** The
> table here is real but was measured only AT REST, and its conclusion (an independent
> element-count wall) is wrong. The cost was promotion all along; it just does not happen
> until something animates.

Declining promotion does **not** buy unlimited facets. The ~330-layer cliff is a
*promotion* cost; there is an independent wall in raw `clip-path`-ed element count.
Measured at rest, no animation, in a real game:

| tokens | sides/prism | facet elements | promoted | fps |
|---|---|---|---|---|
| 55 | 6 | 440 | no | **60.0** |
| 55 | 12 | 770 | no | **60.0** |
| 55 | 24 | 1430 | no | **42.8** |
| 55 | 24 | 1430 | yes | 42.4 |
| 68 | 24 | 1768 | no | **32.4** |

Promotion changes nothing (42.4 vs 42.8), which both confirms the layer reasoning and
proves this is a different wall. **Budget: ~800 facet elements is free, ~1400 is not.**
With `pass` at 55 tokens, that caps a prism at **12 sides**. Whether 12 sides reads as
round for a `disc` or `chip` is an open question for implementation — if it does not, that
is a trip-wire (below), not a licence to spend facets.

### The wall was the layer cliff after all

Implementation shipped 12-sided prisms on this budget, and `pass` then ran at **30fps
during a Pass move** against the flat SVG art's 59.6 — linear in facet count, and
unmoved by promoting the container, exactly as the table above predicts an
element-count wall would behave. It is not one.

`LayerTree` says what is actually happening. 55 tokens, 14 facets each, measured
through CDP:

| | composited layers | painted | layer area |
|---|---|---|---|
| flat SVG art, hosts animating | 59 | 57 | 1.6 Mpx |
| 3D solids, at rest | — | — | (nothing promoted) |
| 3D solids, hosts animating | **1104** | **1047** | **88.6 Mpx** |

**Chromium promotes every element inside a live `preserve-3d` context to its own
composited layer the moment an ANCESTOR transform animates** — which is what a stack's
FLIP does to a component host on every single move. Nothing the token authored asked
for it; declining `will-change` does not decline this. And the layers are not
token-sized: a `clip-path`-ed facet seen through a `perspective` gets conservative
bounds ~2000px on a side, which is where 88.6 megapixels of raster comes from.

That is why the at-rest table looks like an element wall. At rest nothing is promoted,
so what it measured was the ordinary paint cost of 1,430 clipped elements; under
animation the same scene crosses the layer cliff and the two effects are
indistinguishable from the frame rate alone.

**Removing any two of {`perspective`, `preserve-3d`, the facets' own 3D transforms}
still leaves ~1,000 layers.** All three have to go — which for a token they can,
because *a token's pose is a constant*. It is projected once in JavaScript and drawn as
flat `clip-path` polygons, with the back faces culled rather than hidden
(`src/solid/flat-facets.ts`). `pass` then holds **60.5fps through a Pass move** and
composites exactly like the flat art it replaced: 57 painted layers, 1.6 Mpx.

This does **not** transfer to the die, which tumbles and therefore needs a live 3D
context and the promotion that goes with it. It is the reason the two are separate
code paths.

## Why not WebGL

Asked and answered properly, with a prototype, rather than assumed. **Keep CSS 3D** — but
for the right reasons, because three plausible-sounding wrong reasons would block the
correct decision later.

- **A z-buffer would make the sorting problem vanish entirely** — no tilt band, no comb
  pathology, no non-monotone failure. That is the strongest argument for WebGL and it is
  real. It is also a capability this design **declines** rather than a bug it suffers, and
  that is what settles it.
- **"Canvas cannot keep the compositing premise" is FALSE, and must not be written down.**
  A main-thread rAF loop rendered **0 of 46** frames under an 800ms block — but an
  **OffscreenCanvas in a Worker rendered 47 of 47**, and the full architecture at 68
  components rendered 46 of 46. The premise is recoverable by the same trick the baking
  already uses: precompute, hand it to something that is not the main thread.
- **"The parity harness forbids it" overstates the harness.** The die's tumble already has
  no geometry golden — it is covered by comparing rendered keyframes against a recomputed
  trajectory — and `die-shape.spec.ts`'s screencast + hull-deficit harness is entirely
  renderer-agnostic. The harness constrains *how* to test, not *what* to build.
- **The real cost is text.** The legibility system — inscribed content squares, corner-mark
  insets, the 6px floor — is DOM typography, and reproducing it in GL means an SDF glyph
  atlas and re-deriving every tuned constant. That, plus losing the element inspector on a
  pipeline whose bugs were repeatedly found by inspecting elements, is the actual price.

Measured and worth banking, because they make a future crossing cheap: the WebGL context
cap is exactly **16** (one context per component is dead on arrival at 24 checkers pieces);
a per-component canvas inside `#inner` rides a real 485px cross-stack FLIP with **0.00px**
tracking error over 166 frames; a hand-rolled renderer is **2.3KB gzipped** with no new
dependency and no asset pipeline (meshes extrude from the SVGs already shipped);
accessibility and hit-testing are **unchanged**, because they live on the host and the
canvas is `pointer-events: none`.

One architecture is ruled out permanently: a **shared board-wide overlay canvas** that
tracks components by reading their rects. Under an 800ms block it stranded a piece **255px**
from its own focus ring and hit target, and could not self-correct on the first frame back.

**Trip-wires that flip this decision.** Any one firing means revisit:
1. A shape in scope must be a real solid at any angle (i.e. a game wants a meeple that
   genuinely *is* a meeple).
2. ~~A `disc`/`chip` needs more than ~16 sides to read as round, or any game ships more
   than ~70 3D pieces — either pushes past the 1400-facet wall.~~ **Retired 2026-07-28:
   there is no 1400-facet wall (see above). 55 chips draw 330 elements flattened and
   hold 60fps through a move; doubling the sides would draw ~660. Side count is a
   legibility question now, not a budget one.**
3. Real contact shadows or inter-piece occlusion become design goals.

## Bugs found in passing (fix on-branch, with evidence)

1. **`--component-aspect-ratio` overrides are dead code.** `#outer.pawn { …: 2.0 }` and
   `#outer.meeple { …: 1.25 }` never apply: custom-property substitution happens where
   `--component-effective-height` is *declared* (`:host`), where the ratio is always 1.0.
   Every shape renders in a 30×30 box; a pawn letterboxes to ~13×30.
2. **The debuganimations demo token is invisible.** It binds `color`/`type` but no `.item`,
   so `spacer = true` and `visibility: hidden`. `token-throb.spec.ts` cannot catch it
   because it asserts on `#outer`'s computed filter, which is unaffected by visibility.
3. **`faux-components` is dead on a raw stack.** The property declares no `attribute:`, so
   the observed name is `fauxcomponents`; only `boardgame-component-zone` maps the dashed
   form. Both debuganimations uses are no-ops, so that faux path has never run there.

## Testing

The parity harness is the net. Additionally, the **z-buffer comparison harness built for
the decision is worth keeping** — it is the only thing that can prove a solid renders
correctly, it validated itself against a d20 at 0 wrong pixels, and it is the regression
test for any future change to facet placement or culling.
