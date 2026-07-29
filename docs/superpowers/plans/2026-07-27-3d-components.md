# 3D Components (Stage 2) Implementation Plan

**Goal:** Render `boardgame-token`'s convex shapes (`cube`, `token`, `chip`, `disc`) as real
3D solids reusing the dice pipeline, give the non-convex shapes (`meeple`, `pawn`) a 3D
presentation of their existing art, and fix three live bugs found on the way.

**Spec:** `docs/superpowers/specs/2026-07-27-3d-components-design.md` — read it first. Its
decisions were measured, not chosen, and the measurements are in it.

> **SUPERSEDED IN PART — this plan is the plan of record and is kept as written.**
> Task 4 below plans a live CSS 3D scene on a token's `#inner`, and the facet budget in
> Global Constraints is derived from it. That shipped and was then measured out: **a token
> uses no live 3D at all.** Its pose is a constant, so its projection is precomputed at
> build time and the renderer emits flat, untransformed `clip-path` polygons with the back
> faces culled — no `perspective`, no `preserve-3d`, no per-facet transform, no layer
> promotion. That took `pass` at 55 tokens from 30fps to 60.5 during a move. A DIE is
> unchanged: it genuinely tumbles, so its pose is not constant and it still needs a real 3D
> context. There is no ~1400-facet wall; it was the layer cliff, measured at rest where it
> does not bite. For what exists, read
> `docs/superpowers/specs/evidence/2026-07-27-3d-components.md` (finding F1),
> `server/static/src/ARCHITECTURE.md`'s "The solid pipeline", and the design's own
> "SUPERSEDED BY MEASUREMENT" section.

## Global Constraints

- Work in `/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/.claude/worktrees/three-d-dice`
  on `worktree-three-d-dice`. npm/npx from `server/static/`.
- **Facet budget: ~800 elements free, ~1400 is a cliff.** With `pass` at 55 tokens on
  screen, prisms are capped at **12 sides**. Do not spend facets without re-measuring.
- **No layer promotion for tokens.** `will-change: transform` is a dice-specific need.
- **No physics for tokens.** They do not tumble.
- Tokens size by **drawn extent** — the solid's silhouette fills the existing box. No
  `#scaler`, no reserved overflow. A stack-hosted component cannot reserve space.
- Run Playwright in the FOREGROUND. Parity goldens must never be regenerated to make a test
  pass; `git status --short -- server/static/tests/animations/parity/goldens/` must be EMPTY.
- Known load flakes, acceptable as a SOLE failure after an isolated re-run: `waapi-gate:120`,
  `waapi-gate:147`, `waapi-buttons:30`, `waapi-companion:81`, `trace: memory reveal one card`,
  `waapi-attrs`, `fading-text`, `die-shape: d6 pips stay legible`.
- Commits: imperative subject, trailer `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.
  Never stage `.database` or `boardgame-util/boardgame-util`.
- `*.test.ts` is excluded from `tsconfig.json`, so the type checker will NOT catch a
  signature change in a test file. This has already hidden two. Hand-check test files.

---

## Task 1: Author-ordered loops in the geometry seam

**Files:** `src/motion/die-geometry.ts`, `die-geometry.test.ts`

`orientLoop` (`:246-268`) carries two independent convexity assumptions: it derives a face's
outward direction from `dot(normal, centroid − vertexMean) > 0`, and it re-sorts each loop's
vertices by angle about their own mean, discarding the author's order. The real acceptance
predicate was verified against 600 random simple polygons at 100% agreement.

Add an opt-in to `RawSolid` (e.g. `oriented: true`) meaning "my loops are already correctly
wound and ordered — do not rederive." Skip both steps when set.

- Failing tests first: a prism over an L-shape, a U-shape and a 27-vertex meeple silhouette
  are all currently REJECTED (`surface is not a manifold: directed edge N->M used twice`);
  with the opt-in they must be accepted with volumes matching the analytic prism volume to
  1e-15, and the inertia tensor must match an analytic cross-check.
- **Regression bar: no shipped die may change by a single ULP.** Capture every solid's
  vertices, face normals, centroids, `nominalRadius`, `boundingRadius` and inertia tensor
  before the change; assert byte-identical after.
- Commit: `"Let a solid declare its own winding"`.

## Task 2: A prism builder, dice-free

**Files:** create `src/solid/prism.ts` + `.test.ts`

A flat-capped right prism: N side walls around an axis, two flat cap polygons. Parameterized
by side count and height-to-diameter ratio. `token` ≈ 0.55, `chip` ≈ 0.13, `disc` flat.

This is the generalization `src/solid/` exists for, so it must not import anything named for
dice. It produces the same `{normal, centroid, polygon}` shape `facetPlacement` consumes.

- Tests: closed 2-manifold (every directed edge once, reverse present); outward normals;
  cap polygons planar; volume matches `π`-free analytic prism volume; **facet count is
  `sides + 2` and a 12-side prism is 14 facets** (the budget depends on this).
- Commit: `"Build a flat-capped prism from a side count"`.

## Task 3: Keep the z-buffer comparison harness

**Files:** create `tests/animations/parity/solid-render-truth.spec.ts`

The harness built to decide the WebGL question is the only thing that can prove a solid
renders correctly. It validated itself against a d20 at **0 wrong pixels across 253 poses**.
Rebuild it as a permanent spec: render a solid under the real pipeline, rasterize the same
geometry in-page with a real z-buffer, diff, excluding antialiasing seams two ways
(subsample disagreement plus a 2px dilation), and report the largest region surviving one
erosion.

- Assert: a d20 and a d6 have **zero** thick wrong regions; a 12-side prism likewise.
- This is the regression test for any future change to facet placement or culling.
- Commit: `"Prove a solid renders what a z-buffer would"`.

## Task 4: 3D convex tokens

**Files:** `src/components/boardgame-token.ts`, `tests/animations/parity/token-3d.spec.ts`

Render `cube`, `token`, `chip`, `disc` as solids. `cube` is `dieGeometry(6)` verbatim; the
other three are prisms at their aspect ratios, **12 sides**.

The scene goes on `#inner` — verified free (tokens compile no visual tracks) and verified
FLIP-safe (projection ratio held at exactly 2.000 across 143 frames of real grid and pile
swaps including reparenting, zero flattened frames). Copy the card's camera placement
(`perspective` on `#outer`, which is verified to reach through the elevation filter).

**Size by drawn extent.** A cube's face spans 1/√3 of its circumsphere, so sizing by
circumsphere would render every token ~40% smaller than the SVG it replaces. Scale the solid
so its silhouette fills `--component-effective-width`.

**THE CRITICAL REQUIREMENT — the resting pose.** Stacks pool and reparent hosts, and nothing
re-derives a pose on reuse, so a node carries whatever the previous occupant left. This is
the failure `_clearRoll` exists to prevent for the die, multiplied by pooling, faux
components, and motion carriers that are `noAnimate` and can never self-correct.

The pose must be a **pure function of current state, re-derived every render**, and
explicitly cleared when the component is orphaned or recycled. Test it directly: drive a real
stack through a membership change that recycles a host, and assert the reused element shows
the new component's pose and not the old one's.

- Also: keep `spacer` (no `item`) from building a scene at all.
- Verify the six-shape debuganimations selector still switches live, checkers' 24 discs and
  tictactoe's 9 chips still render, and `pass` at 55 tokens holds 60fps.
- Commit: `"Render a token's convex shapes as solids"`.

## Task 5: 3D presentation for meeple and pawn

**Files:** `src/components/boardgame-token.ts`, the same spec

These keep their authored SVG. Give them depth without a mesh: a small tilt, a rim/edge
treatment, and a contact shadow consistent with the solids' lighting — so a board mixing a
meeple with a cube reads as one scene.

The tilt must stay small enough to read as a piece standing on a board rather than a card
lying on one, and must be a pure function of state like Task 4's pose.

- Verify a meeple and a pawn beside a cube and a disc at the same size look like they belong
  together. This is a judgment step: look at it and report what you saw.
- Commit: `"Give the non-convex shapes depth without a mesh"`.

## Task 6: The three bugs

**Files:** `src/components/boardgame-token.ts`, `boardgame-component-stack.ts`,
`examples/debuganimations/client/boardgame-render-game-debuganimations.ts`

Each needs a red-first test proving the bug, then the fix.

1. **`--component-aspect-ratio` overrides are dead.** `#outer.pawn { …: 2.0 }` and
   `#outer.meeple { …: 1.25 }` never apply — substitution happens where
   `--component-effective-height` is declared, at `:host`, where the ratio is always 1.0.
   Every shape renders 30×30; a pawn letterboxes to ~13×30. Fix so a non-square shape can
   actually be non-square, or delete the dead rules and say why.
2. **The debuganimations demo token is invisible.** It binds `color`/`type` but no `.item`,
   so `spacer = true` and `visibility: hidden`. Note `token-throb.spec.ts` cannot catch this
   because it asserts on `#outer`'s computed filter, which visibility does not affect.
3. **`faux-components` is dead on a raw stack.** The property declares no `attribute:`, so
   the observed name is `fauxcomponents`; only `boardgame-component-zone` maps the dashed
   form, so both debuganimations uses are no-ops and that faux path has never run there.
- Commit each separately.

## Task 7: Docs and close-out

- `TUTORIAL.md`: extend the components section — which shapes are 3D, that shape drives the
  solid, sizing, and the honest limit that `meeple`/`pawn` are presented rather than modelled.
- `server/static/src/ARCHITECTURE.md`: the shared `src/solid/` pipeline and its two consumers.
- Parity README: the new render-truth spec and the facet budget.
- Evidence pack at `docs/superpowers/specs/evidence/2026-07-27-3d-components.md`: the design
  decisions and which measurement discharges each.
- Full sweep: `npm run test:unit`, `npm run type-check`, `npm run test:renderer`,
  `npx playwright test tests/animations/`, and from the repo root `GOWORK=off go build ./...`
  and `go test ./...`. Goldens EMPTY.
- Final whole-branch review.
