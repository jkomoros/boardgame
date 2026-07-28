# 3D components (Stage 2) — evidence pack

Date: 2026-07-28
Branch: `worktree-three-d-dice`
Plan: `docs/superpowers/plans/2026-07-27-3d-components.md`
Spec: `docs/superpowers/specs/2026-07-27-3d-components-design.md`

`boardgame-token`'s four convex shapes (`cube`, `token`, `chip`, `disc`) are now
real 3D solids built on the shared `src/solid/` pipeline. The two non-convex
ones (`meeple`, `pawn`) keep their authored SVG and gain depth without a mesh.

**The one thing to carry away, because three documents used to say otherwise: a
token does NOT use live CSS 3D.** Its pose is a constant, so its projection is
computed once at build time and the renderer emits flat, untransformed
`clip-path` polygons with the back faces culled — no `perspective`, no
`preserve-3d`, no per-facet transform, no layer promotion. A die is different:
it genuinely tumbles, so its pose is not a constant and it still needs a real 3D
context. They are two consumers of one pipeline, using different halves of it.

This pack is in three parts. Part 1 discharges each design decision with the
measurement that settles it. Part 2 is the three findings that **overturned an
earlier belief** — the load-bearing ones, recorded at length because each was
believed, written down, and then proved wrong by a measurement. Part 3 is the
defects found on the way.

---

## Part 1 — the decisions, and what discharges each

### D1 — Convex shapes become solids; non-convex ones do not

**The decision.** `cube`, `token`, `chip` and `disc` are generated solids.
`meeple` and `pawn` keep authored art. The boundary is not taste; it is where
CSS stops rendering correctly.

**Why it needed measuring.** The renderer's entire hidden-surface removal is
"draw only the camera-facing facets". That is *provably* sufficient on a closed
convex surface — the front-facing facets tile the silhouette exactly once and
never occlude one another — and provably insufficient otherwise. CSS
`preserve-3d` has no z-buffer; within a 3D context Chromium sorts by a plane
heuristic, not per pixel.

**Discharged.** A meeple prism was rendered under the die's exact conditions and
diffed against a Node z-buffer rasterization under an identical camera, 16
randomized poses per tilt band, antialiasing excluded two ways:

| tilt | worst wrong % | worst 2px-thick region | self-occluded % |
|---|---|---|---|
| 0–15° | ≤0.057 | **0 px** | 2–3 |
| 20° | 0.170 | 1 px | 4 |
| 35° | 0.226 | 6 px | 11 |
| 50° | 0.594 | 38 px | 16 |
| 75° | **2.844** | 115 px | 28 |

Controls: a **d20 at 0 wrong pixels across 253 poses** (so the harness is
sound), and a deliberately mis-painted positive control flagged 79%.

**And the failure is not monotone in tilt**, so no angle limit is safe. A
deliberately hostile "comb" prism is clean to 60° and then collapses at 75°:
**12.1% of the silhouette wrong**, painting the interiors of its slots straight
over the near teeth.

**The decisive argument is not that the prism fails — it is where it succeeds.**
At 20° tilt only 4% of the silhouette has two front-facing facets over it. The
band where a prism renders correctly is exactly the band where a prism *looks
like a tilted flat card with a thin rim* — which one authored SVG plus a tilt, a
rim and a contact shadow reproduces, for 34 fewer DOM elements and zero sorting
risk at any angle.

Recorded so it is not re-litigated: there is **no cyclic overlap** in a prism
over a simple polygon (0 cycles in 120 random orientations), so a correct
painter's order always exists. The obstacle is not the geometry — it is that the
browser will not sort to it and gives no hook to make it. See **F3** below for
the experiment that closed the last escape route.

**Held by.** `solid-render-truth.spec.ts` (the harness, kept permanently),
`token-3d.spec.ts`, `token-art-depth.spec.ts`.

### D2 — Keep CSS 3D rather than crossing to WebGL

**Discharged with a prototype, not an assumption**, and the three
plausible-sounding *wrong* reasons are recorded so they cannot block a correct
decision later:

- A z-buffer **would** make the sorting problem vanish entirely. That is the
  strongest argument for WebGL and it is real. This design **declines** the
  capability rather than suffering a bug, and that is what settles it.
- **"Canvas cannot keep the compositing premise" is FALSE and must not be
  written down.** A main-thread rAF loop rendered **0 of 46** frames under an
  800ms block — but an **OffscreenCanvas in a Worker rendered 47 of 47**, and
  the full architecture at 68 components rendered 46 of 46.
- **"The parity harness forbids it" overstates the harness.** The die's tumble
  already has no geometry golden, and `die-shape.spec.ts` is renderer-agnostic.
  The harness constrains *how* to test, not *what* to build.
- **The real cost is text**: the legibility system (inscribed content squares,
  corner-mark insets, the 6px floor) is DOM typography, and reproducing it in GL
  means an SDF glyph atlas and re-deriving every tuned constant — plus losing
  the element inspector on a pipeline whose bugs were repeatedly found by
  inspecting elements.

Banked so a future crossing is cheap: the WebGL context cap is exactly **16**
(one per component is dead on arrival at 24 checkers pieces); a per-component
canvas inside `#inner` rides a real 485px cross-stack FLIP with **0.00px**
tracking error over 166 frames; a hand-rolled renderer is **2.3KB gzipped**;
accessibility and hit-testing are unchanged. One architecture is ruled out
permanently: a **shared board-wide overlay canvas** that tracks components by
reading their rects stranded a piece **255px** from its own focus ring under an
800ms block and could not self-correct on the first frame back.

### D3 — The geometry seam: let a solid declare its own winding

**The decision.** `RawSolid` gains `oriented?: boolean`, which skips
`orientLoop`'s rederivation.

**Why.** `orientLoop` carried two *independent* convexity assumptions — it
derived a face's outward direction from `dot(normal, centroid − vertexMean) > 0`
and re-sorted each loop's vertices by angle about their own mean, discarding the
author's order. The real acceptance predicate was found rather than assumed:
verified against **600 random simple polygons at 100% agreement** with the
shipped code, and **167 accepted cases have non-convex cap faces** — so
non-convex *faces* were never the constraint.

Everything downstream was already correct for non-convex input, measured:
signed-tetrahedron volume/centroid/inertia agrees analytically to **1e-16**, the
fan triangulation's out-of-polygon triangles cancel to **3.3e-16** against
shoelace, and `validateClosedSurface` is pure combinatorics.

**Discharged.** Commit `5c82f22d`. The default branch is character-for-character
the old expression, and the regression bar was **no shipped die may change by a
single ULP** — every solid's vertices, normals, centroids, `nominalRadius`,
`boundingRadius` and inertia tensor captured before and asserted byte-identical
after. It skips *derivation*, not *validation*: `validateIndices` and
`validateClosedSurface` still run.

### D4 — A prism builder that is not a die

**The decision.** `src/solid/prism.ts`: N side walls about an axis, two flat
caps, parameterized by side count and height-over-diameter. `token` 0.55, `chip`
0.13, `disc` 0.10.

**Discharged.** Commit `a234d584`. Tests: closed 2-manifold (every directed edge
once, reverse present), outward normals, planar caps, analytic volume, and facet
count exactly `sides + 2`. It imports **two types and nothing else** from
`src/solid/`, both erased at runtime, so nothing named for a die is reachable
from it — which is what makes `src/solid/` a shared pipeline rather than the
dice pipeline with a second caller.

`disc` is 0.10 rather than as flat as possible because **at 0.05 its rim is
under a pixel at checkers' size** and the piece becomes indistinguishable from a
flat polygon.

The prism is built about **+Z** (at the camera) where a die is built about +Y
(at the ceiling), so an unposed chip shows its cap face-on and its silhouette
*is* the cross-section — which is the sizing contract in D6. `SHAPES[…].align`
is the one rotation that reconciles the two families, and a sign error there
renders a cube seen from below next to a disc seen from above.

### D5 — The resting pose is a pure function of state, never remembered

**The decision, and the design called it the biggest risk.** A stack pools and
reparents component hosts, and nothing re-derives a pose on reuse — so a node
carries whatever the previous occupant left. That is the failure `_clearRoll`
exists to prevent for the die, multiplied by pooling, faux components, and
motion carriers that are `noAnimate` and can never self-correct by playing
anything.

**Measured precedent.** A stale write to a card's `#inner` transform mid-flight
was invisible for the whole flight and became **permanent** the instant the
animation ended, leaving the card face-up and stuck at 45°.

**Discharged structurally rather than by discipline.** Everything about a
token's presentation is a pure function of `(type, color)` in a DOM-free module
(`src/components/token-solid.ts`), and the depth treatment for `meeple`/`pawn`
is CSS keyed off the type's own class, which `_computeClasses` recomputes from
`this.type` on every render — no imperative write, nothing to forget to clear.
The art selectors are **derived** (`LEGAL_TYPES.filter(t => !isTokenSolidShape(t))`),
so promoting a shape to a real solid stops it being art-treated automatically.

**Held by.** `token-3d.spec.ts` — *shows the new component's pose on a recycled
host* drives a real stack through a membership change that recycles a host, and
*clears the visual carrier before the host is orphaned*; `token-art-depth.spec.ts`
— *follow a recycled host in both directions*.

### D6 — Size by drawn extent, never by circumsphere

**The decision.** The solid is scaled so its posed, projected silhouette fills
the existing box (`fitScale`), rather than sizing the nominal sphere to the box
as `boardgame-die.ts` does.

**Why.** A cube's face spans `1/√3` = **57.7%** of its circumsphere, so a 3D
cube dropped in at the die's scale renders **~40% smaller** than the SVG it
replaces, in checkers, tictactoe and `pass`. And a stack-hosted component
**cannot reserve extra space** the way a die's `#scaler` does: stack margins,
the board layout's `aspect-ratio: 1` clamp, the hardcoded 100px spread/fan
margins and the FLIP scale ratio all key off that box.

**Held by.** `token-3d.spec.ts` — *draws a silhouette that fills the token's own
box*; `token-box.spec.ts` — *is square, and says so*, plus *lets the authored
art keep its own proportions inside that square*.

The camera depth is expressed in **token widths** (`CAMERA_DEPTH_WIDTHS = 6`)
rather than the card's flat `perspective: 1000px`, because `pass` puts 55 tokens
on screen at two different `--component-scale` values: a fixed pixel depth would
foreshorten the two sets differently and the same piece would be a different
shape depending on where it sat.

### D7 — One camera for every shape

**The decision.** `CAMERA_LEAN_DEGREES = 50`, shared by every shape; only the
shape's own alignment and spin differ.

**Chosen by looking, at the sizes these actually render at, which is the only
way it could have been chosen.** The boards are drawn top-down, which argues for
a high camera — but a `disc` is a tenth as thick as it is wide, and a high
camera draws its rim at zero pixels and the whole solid as a flat polygon.
Rendered side by side at 30, 60 and 120px and on a real checkers board: **65°
leaves a disc and a chip indistinguishable from the flat art**, 58 is marginal
at 30px, and **50 puts a legible rim under every prism** while still reading as
a piece standing on a board rather than lying on one. A cube at 50° shows its
top and two sides, which is also what the isometric `token_cube.svg` draws.

### D8 — Lighting the two families from the same place

**The decision.** `LIGHT` is upper-left-and-slightly-front, and Lambert shading
is `AMBIENT + DIFFUSE·(n·l)` clamped to `[0.5, 1.08]`.

**The constants are not free.** `token_disc.svg`'s bevel gradient runs #C80000 →
#700000 against a #C20000 face, i.e. **1.03 down to 0.58** of the face's own
brightness, with its cap at 1.00 by definition. The shipped constants reproduce
that range on the shape the SVG draws: at the resting pose a disc's cap comes
out at **1.02** and the visible rim sweeps **0.83 → 0.59** — 1.02..0.59 against
the art's 1.03..0.58. So the solid is shaded like the art it replaces rather
than like a renderer. The floor exists for the shapes that are not discs: a cube
presents faces pointing much further from the light than a rim ever does.

Upper-left is the house convention the art states outright — `token_disc.svg`
says "Light source: upper-left". (`token_cube.svg`, much older Illustrator
output, disagrees and lights from the right; the authored assets win.) See **F2**
for what measuring this actually found.

`SHADOW_DIRECTION` is **derived** from `LIGHT` rather than typed, so moving the
light moves every `meeple`/`pawn` shadow with it and the board cannot end up lit
from two places.

**Held by.** `token-art-depth.spec.ts` — *are lit from the same side as the
solids*, *offset every shadow along the light's own direction*.

### D9 — Colour shared arithmetically, not by two lists agreeing

**The decision.** `TOKEN_COLOR_FILTERS` is one table. `boardgame-token.ts`
generates its `#outer.<color> img` rules from it for the flat art, and
`tokenBaseColor` evaluates the same filter strings **as colour matrices** over
one representative red (`#C50000`) to get a solid's base colour.

So a 3D chip and a 2D meeple beside it are the same hue *by construction*.
**Held by** `token-3d.spec.ts` — *is coloured by the same filters the flat art
is* — which paints a swatch per colour with the very filter string the token's
own rule uses, screenshots it, and requires the arithmetic to match the
browser's own implementation at **distance zero on all nine colours**, plus a
non-vacuity check that the nine are distinct. The arithmetic is not an
approximation of the filter; it *is* the filter.

### D10 — The z-buffer harness is kept, permanently

**The decision.** The harness built to decide the WebGL question became
`tests/animations/parity/solid-render-truth.spec.ts` — the only test in the repo
that can prove a solid renders *correctly*, as opposed to renders *opaquely*
(`die-shape.spec.ts`) or renders *as recorded* (the golden corpus).

**And it validates itself**, which is what makes any of its other numbers
believable: a **d20 at zero wrong pixels** (not merely zero thick regions —
a d20 is necessarily rendered correctly, so any error it reports is the
harness's own), and a **deliberately mis-painted d20 at >90% of the silhouette
wrong** with a thick region over 1000px on every pose. Without the second, the
first is consistent with a harness that reports zero for everything.

**No golden, ever.** The reference is recomputed from the geometry every run;
the reference maths is written out in the spec rather than imported from `src/`,
deliberately, because a shared sign error would cancel and the test would agree
with the bug.

Antialiasing is excluded **two ways** so that raising a tolerance is never the
answer: the reference is supersampled 3×3 and a pixel whose nine subsamples
disagree is not judged at all; those seams are dilated by 2px; what survives is
differenced exactly and **eroded once**. A 1px hairline does not survive an
erosion; a facet painted in front of the facet that should have hidden it does.
Commit `b65cab2c`.

---

## Part 2 — the three findings that overturned an earlier belief

These are the load-bearing ones. Each was believed, written down as a design
constraint, and then contradicted by a measurement.

### F1 — There is no facet-count wall. It was the layer cliff all along.

**What was believed, and written into the plan as a hard constraint:** "**Facet
budget: ~800 elements free, ~1400 is a cliff.** With `pass` at 55 tokens, prisms
are capped at **12 sides**. Do not spend facets without re-measuring." It came
from a real table, measured at rest in a real game:

| tokens | sides/prism | facet elements | promoted | fps |
|---|---|---|---|---|
| 55 | 6 | 440 | no | 60.0 |
| 55 | 12 | 770 | no | 60.0 |
| 55 | 24 | 1430 | no | **42.8** |
| 55 | 24 | 1430 | yes | 42.4 |
| 68 | 24 | 1768 | no | **32.4** |

Promotion changing nothing (42.4 vs 42.8) was read as *confirming* that this was
a second, independent wall under the known ~330-layer promotion cliff.

**What actually happened.** Implementation shipped 12-sided prisms on that
budget — 770 elements, comfortably inside it — and `pass` ran at **30fps during
a Pass move** against the flat SVG art's 59.6. Linear in facet count, and
unmoved by promoting the container, exactly as an element-count wall would
behave.

**The measurement that overturned it.** CDP's `LayerTree`, 55 tokens at 14
facets each:

| | composited | painted | layer area |
|---|---|---|---|
| flat SVG art, hosts animating | 59 | 57 | 1.6 Mpx |
| 3D solids, at rest | — | — | nothing promoted |
| 3D solids, hosts animating | **1104** | **1047** | **88.6 Mpx** |

**Chromium promotes every element inside a live `preserve-3d` context to its own
composited layer the moment an ANCESTOR transform animates**, so it can sort
them on the compositor. A stack's FLIP animates the component host on every
single move. Nothing the token authored asked for this, and *declining
`will-change` does not decline it* — which was the other thing the design
believed. The layers are not token-sized either: a `clip-path`ed facet seen
through a `perspective` gets conservative bounds ~2000px on a side, which is
where 88.6 megapixels of raster comes from.

It explains every symptom the facet theory could not: linear in facets (one
layer each), zero for flat art (no 3D context), fine at rest (nothing
animating), and unmoved by promoting the container (`will-change` on `#solid`
added 55 layers on top of the 770 it did not remove — the promotion was never
the container's to give).

**And it explains why the at-rest table looked like an element wall.** At rest
nothing is promoted, so what that table measured was the ordinary paint cost of
1,430 clipped elements. Under animation the same scene crosses the layer cliff
and the two effects are indistinguishable from the frame rate alone.

**The fix, and why it is not a compromise.** A token's pose is a **constant**.
So its projection is a constant, and the perspective divide can be done **once
at build time**, in JavaScript, emitting flat untransformed `clip-path`
polygons. Removing the three offenders one at a time, measured: `perspective`
alone → still 1047 layers; `preserve-3d` alone → 992; both plus
`backface-visibility` → 992. **All three had to go**, including the facets' own
3D transforms — and baking the perspective into a per-facet `matrix3d` under a
flat parent does not help either, because a 3D-transformed element is promoted
on its own.

Back faces are then **culled rather than hidden**, which the same convexity
theorem makes exactly as sufficient as `backface-visibility` was, and which
means nothing needs sorting at all. Element counts fall out of the cull, not out
of fewer sides: a 12-side prism draws 6 or 7 of its 14 polygons, a cube 3 of 6.
**12 sides is unchanged.**

| `pass`, 55 tokens | before | after | flat-art control |
|---|---|---|---|
| at rest | 31.0 | **59.8** | — |
| during a real move | 30.0–31.3 | **60.5** (median) | 59.6 |

| game | facets | fps | painted layers |
|---|---|---|---|
| `pass` (55 tokens) | 770 → 330 | 31.0 → 59.8 | 1189 → **192** |
| `checkers` (24 discs) | 336 → 144 | 59.9 → 59.7 | 461 → **29** |
| `debuganimations` (30 tokens) | 180 → 90 | 59.9 → 59.8 | 416 → **176** |

**Consequences recorded.** The facet budget and the 12-side cap were derived
from the cliff; with no 3D context a token's facets are ordinary clipped divs
and the budget no longer binds the same way. **The 12-side cap stays on
*appearance* grounds** (a dodecagon reads as a dodecagon above ~100px), not
performance — which retired trip-wire #2 in the design outright. Nothing about
the die changed: it tumbles, so it still needs a live 3D context and the
promotion that goes with it. That is the reason the two are separate code paths.

**Held by** `token-3d.spec.ts` — *promotes no layers, even while an ancestor
transform animates* — which asserts **layers, not frames**, deliberately: fps on
a shared machine is a coin flip, the layer count is exact, and **at rest the
broken version passes every other assertion in the file**. Commit `b41ddc9e`.

**Found in passing, and it matters for reading any "at rest" number on this
branch:** `pass` and `checkers` are **never actually at rest**. A relayout loop
(`slotchange` → `_updateComponentClasses` → the `layoutTransform` setter →
`play()`) starts fresh 500ms animations on every component about twice a second,
forever. `debuganimations` does not do it (plays counter flat at 277 over 3 idle
seconds; `pass` climbs by ~240, `checkers` by ~120). It is pre-existing and
identical with flat art — but it is why `pass` measured 29fps "at rest" here
where an earlier report measured 60.3, and it deserves its own investigation.

### F2 — The meeple and pawn are lit from the wrong side, and an eyeball said otherwise

**What was believed.** An eyeball pass over the rendered assets concluded the
pawn's specular highlight was on the **upper left** — i.e. that the art already
agreed with the solids' light and nothing needed doing.

**The measurement that overturned it.** A quadrant sampler: mean luma of the
left half against the right half of the drawn silhouette's top band, at 200px on
white.

| shape | left − right |
|---|---|
| cube | +1.0 |
| token | +0.1 |
| chip | 0.0 |
| disc | −0.02 |
| **meeple** | **−8.2** |
| **pawn** | **−17.1** |

The solids are near-symmetric by construction (a 12-gon rim under a mostly
overhead light). The two art shapes are decisively the *other* way — both lit
from the upper **right**. That was the single biggest reason a board mixing them
read as two art styles, and it is exactly the kind of thing an eyeball gets
wrong and a sampler does not.

**The fix is one line and costs nothing:** `transform: scaleX(-1)`. Both assets
are mirror-symmetric in silhouette, so mirroring moves the light across without
touching the shape. Measured afterwards: meeple **+8.2**, pawn **+17.1**.

The rest of the depth treatment, all keyed to the same `LIGHT`:

- **Edge treatment** — a hard-edged `drop-shadow` hugging the silhouette, offset
  along `SHADOW_DIRECTION`. Deliberately **no matching light rim**: the solids'
  brightest surface is 1.02× base, so a white highlight is not something this
  light produces, and rendered at 110px it read as a cut-out sticker outline.
- **Contact shadow** — a soft ellipse at the foot, in the elevation shadows' own
  `rgba(60,40,20)`, sized off the piece's **drawn** width rather than its box. An
  SVG keeps its own proportions inside a square box, so a pawn spans 0.43 of its
  box and a box-sized shadow read as a puddle. Only the standing shapes get one;
  a disc already meets the board along its dark rim.
- **Tilt** — a 2D `scaleY(0.94)` about the foot. Small on purpose: fully
  reprojecting a standing piece to the 50° camera would foreshorten it by
  `cos(50°) = 0.64` and lay it down, which is the mesh work this design declines.
  It is a 2D scale rather than a `rotateX` **because a 3D transform is a
  composited layer the moment an ancestor animates** — F1, applied prospectively,
  and `token-art-depth.spec.ts` carries the same layer tripwire to enforce it.

`#art` is a wrapper element rather than styling the `img` directly, because an
`#outer.<colour> img` rule outranks anything the img could be given inside this
shadow tree — an edge filter on the img would have survived on red and vanished
on the other nine colours.

**The mixed-board judgement, looked at rather than asserted.** `cube · disc ·
chip · token · meeple · pawn` side by side at 30, 60 and 120px, red on parchment
and blue/yellow on charcoal. **What holds:** same hue across all six (one colour
table drives both paths); same light — at 120px the cube's bright face, every
prism's lit rim, the meeple's bright extrusion and the pawn's specular are all
upper-left; all six grounded, the solids by their dark bottom rim and the art
shapes by their contact ellipse. At 60px — the size checkers and murdermrmonroe
actually draw at — the set reads as one set. **What does not, and cannot without
a mesh:** the art shapes are smooth gradients where the solids are flat-shaded
facets, the meeple's extrusion shows Illustrator banding at 120px, and the
cameras differ irreconcilably (solids from 50° above the board, the meeple art a
standing piece seen nearly front-on, the pawn a side elevation at eye level).
That is the honest limit the design names: presented, not modelled.

Commit `d7357c8e`.

### F3 — Explicit `z-index` sorting does not fix a non-convex solid, and the wrong order measures *better*

**What was believed.** The obvious escape route from D1: if CSS will not sort a
non-convex prism correctly, sort it by hand — compute the painter's order and
put it in `z-index`.

**The measurement that overturned it.** Far-to-near ordering — the correct
painter's order — measured **identical to doing nothing**. And *near-to-far*,
definitionally the **wrong** order, measured **better**: 1.208% wrong against
2.844%.

That is not noise; it is a signature. Chromium runs its own sort inside a 3D
rendering context and `z-index` only perturbs its tie-breaks, so the two
orderings are not being honoured at all — they are nudging a heuristic, and one
of the nudges happens to help.

Two further controls close the route completely. **Layer promotion is irrelevant
to static ordering**: `will-change` vs none was **pixel-identical across all 253
poses** — a separate fact from F1, which is about promotion under *animation*.
And there is **no cyclic overlap to blame**: 0 cycles in 120 random
orientations, so a correct painter's order provably exists and the browser
simply will not be told it.

This is what makes D1 a decision rather than a limitation with an unexplored
workaround, and it is the reason "just sort them" must not be proposed again.

---

## Part 3 — defects found on the way

Each fixed on-branch, red-first.

| # | Defect | Commit |
|---|---|---|
| 1 | `--component-aspect-ratio` overrides were dead code | `aa0497df` |
| 2 | The debuganimations demo token was invisible | `19c67a9c` |
| 3 | `faux-components` was dead on a raw stack | `59ce38de` |
| 4 | A gate that never opened could not settle the initial load | `eb7c3f43` |
| 5 | A token with an unnamed colour froze permanently | `f76e59a3` |
| 6 | `/api/list/game` 500s for a stored game type this server dropped | `cd66e3af` |
| 7 | A test compared two `undefined`s and could not fail | `11bada92` |
| 8 | The `[spacer]` selector matched nothing, so placeholders accumulated | `79f8d78e` |

**1. `--component-aspect-ratio` overrides — deleted, with reasons.** `#outer.pawn
{ …: 2.0 }` and `#outer.meeple { …: 1.25 }` never applied: a custom property
that references another is substituted where it is **declared**, and
`--component-effective-height` is declared at `:host`, above `#outer`. Measured:
a pawn computed `--component-aspect-ratio: 2.0` on `#outer` while
`--component-effective-height` computed to `calc(calc(1.0 * 30px) * 1.0)`, and
every shape laid out 30×30. **Deleted rather than made to work**, for three
measured reasons: every consumer assumes the box is square (D6); the art is
already in true proportion (`token_pawn.svg` is 89.536 × 207.215 and draws at
its own 0.432 inside any box, so the rules would have resized the letterbox, not
removed it); and the numbers were wrong anyway — 2.0 against the pawn asset's
2.31, 1.25 against the meeple's 1.11 — with 2.0 rendered and looked at, drawing
a 240px pawn beside a 120px cube. Pinned by `token-box.spec.ts`: a shape may not
declare a ratio its box does not have.

**2. The debuganimations demo token was invisible.** It bound `color`/`type` but
no `.item`, so `spacer = true` and `visibility: hidden`. `token-throb.spec.ts`
could not catch it because it asserts on `#outer`'s computed filter, which
visibility does not affect. Fixed by binding a **frozen** constant — a fresh
object every render would re-run `_itemChanged` and rewrite `spacer` and the
reflected `id` on every keystroke of the two selects. The new spec asserts what
visibility actually changes: `spacer`, computed `visibility`, and whether red
pixels reached the screen inside the element's own rect.

**3. `faux-components` was dead on a raw stack — and this one is a whole
class.** Lit derives a reactive property's observed attribute by **lowercasing**
the property name, not dash-casing it, so `fauxComponents` observed
`fauxcomponents` and the dashed spelling every author writes was a silent no-op.
`boardgame-component-zone` mapped the dashed form for *itself* and forwarded the
value down as a property, which is why the zone path always worked and the raw
path never did — including both of `debuganimations`' uses, in the one game
written to exercise it. Measured before the fix: `getAttribute` `"5"`, property
`0`, zero faux hosts.

It was **held once and then landed**, which is itself worth recording. Making it
work builds four faux hosts in each of debuganimations' two stacks, and those
hosts changed what debuganimations *animates* — three geometry goldens observed
a 280ms `box-shadow` transition on a card's `#inner` that does not occur at all
without the fix (0 such transitions against 8). Three suppression attempts
failed. So it was reverted (`65c7d3b5`) rather than re-recorded under a "goldens
must end EMPTY" instruction, and then re-landed (`59ce38de`) once the behaviour
change was explicitly declared, with the three geometry goldens re-recorded
**scoped to the affected tests, never the directory** — a whole-directory record
rewrites the trace goldens from a fresh randomized deal — and all four trace
goldens verified byte-identical by sha256.

`noDefaultSpacer`, `autoMessage` and five roster/lobby boolean bindings were
dead for the same reason, some for years. **No runtime test can see a bug whose
symptom is "nothing happened"**, so the class is closed mechanically:
`src/components/property-attribute-names.test.ts` requires every multi-word
reactive property in `src/` to declare its attribute explicitly (dash-case, or
`attribute: false`), and requires a declared name to *be* the dash-case one,
since declaring the wrong one reproduces the bug with extra steps. It carries
its own premise guard so a refactor cannot make it vacuous.

**8. The `[spacer]` selector matched nothing** — the neighbouring trap the lint
above deliberately does not catch, because it is a different one. `spacer` is a
single word, so its observed *name* was right; what was missing was
`reflect: true`, so the attribute was never written to the DOM at all.
`boardgame-component-stack` looks its placeholder up by attribute in three
places — refresh, teardown, and `_slotChanged` (twice: `haveSpacer`, and the
removal loop). All three matched nothing, always. So `haveSpacer` was
permanently false and every `_slotChanged` on an empty stack built another
placeholder, and the removal loop's NodeList was permanently empty so none was
ever removed. Measured, cycling one raw stack empty/full six times: **3, 3, 4,
4, 5, 5, 6, 6, 7, 7, 8, 8, 9** hosts — monotone, +1 per empty phase, never
returning to 0 when full, with the selector matching 0 at all thirteen steps. A
freshly loaded `debuganimations` carried **60 placeholder hosts across 15
stacks** where 15 were wanted, each a real custom element with its own shadow
root. Nothing threw and nothing was visible, because a placeholder is
`visibility: hidden` — which is why it survived. Two changes: `spacer` reflects,
and `_slotChanged` writes the attribute synchronously at the creation site,
because Lit reflects a microtask later and `_slotChanged` runs four times during
a debuganimations mount (reflection alone still left 4 hosts in one stack).
After both: 1 when empty, 0 when full, and debuganimations carries 1. **No
golden moved** — unlike defect 3, this only removes hidden hosts that were never
animating. `stack-spacer-reflect.spec.ts`.

**4, 5, 6, 7 — the ones that blocked verification rather than shipping.** The
gate-accounting fix (`eb7c3f43`) came out of defect 3's first attempt:
`settleInitialLoad` required `gateOpens === gateCloses`, and
`boardgame-render-game._rendererLoaded` closes a gate that by design never
opened, for a renderer that mounts after its state. `gateCloses >= gateOpens` is
what "all" actually means; the fix stayed on-branch even when defect 3 was
reverted, because the invariant was wrong on its own terms. The unnamed-colour
freeze (`f76e59a3`) is the sharpest: `classMap` records an empty class as
applied and then removes it with `classList.remove('')`, which throws a
`SyntaxError` from inside Lit's own update — aborting `performUpdate`, which Lit
never retries, freezing the class list, item, spacer flag and content at the
poisoned pass permanently. It **emptied the entire checkers board, silently**,
because checkers passes `''` for every component whose colour a player may not
see. The list-games 500 (`cd66e3af`) was a nil `*managerInfo` dereference for
any *stored* game whose type is not in this server's `config.json` — 190
recovered panics in one log, 39 of 41 test failures in one run — fixed on-branch
because a green suite was not otherwise reachable, and because "drop a game from
config.json and your entire game list 500s" is production-shaped. Defect 7 is
the one worth generalizing: `token-solid.test.ts` compared `.pose` on two
objects that no longer *have* a `pose` field, so both sides were `undefined` and
the assertion could not fail — and `*.test.ts` is outside `tsconfig.json`'s
`include`, so the type checker never saw it. **Hand-check test files after a
signature change**; this has now hidden three.

---

## What the docs say now

- `TUTORIAL.md` — a `boardgame-token` section: which shapes are solids and which
  keep authored art *and why*, the sizing contract (drawn extent, square box,
  and the contrast with `--die-size`), the ten colours, and the honest limits.
- `server/static/src/ARCHITECTURE.md` — "The solid pipeline (`src/solid/`), and
  its two consumers", stating the constant-pose/precomputed-flat versus
  tumbling/live-3D split explicitly, with the layer table.
- `tests/animations/parity/README.md` — "Solids: the specs with no golden, and
  why they have none": `solid-render-truth.spec.ts` and how it validates itself,
  `token-flat-truth.spec.ts`, the `token-3d.spec.ts` layer tripwire, and the
  `property-attribute-names.test.ts` lint.
