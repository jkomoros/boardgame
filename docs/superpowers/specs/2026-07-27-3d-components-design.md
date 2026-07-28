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
  3D token promotes no layers at all. The measured ~330-layer cliff is a dice constraint
  tokens simply do not inherit. This matters: `pass` shows **55 tokens at once**.
- **No face content**, no reading pose, no `readingRule`, no `capFaces` semantics, no
  legibility floor. Tokens carry no marks.

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
