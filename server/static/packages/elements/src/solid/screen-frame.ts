/**
 * The CSS screen frame, and the ONE map into it.
 *
 * `src/solid/` is the pipeline that turns a closed convex surface into DOM: a
 * perspective wrapper, a `preserve-3d` carrier, and one absolutely positioned,
 * `clip-path`ed element per polygon. Nothing in this directory knows what the
 * solid IS — a die, a token, a chit stood on its edge — and nothing in it may
 * learn: `boardgame-die.ts` supplies pips and face values on top, and a 3D
 * `boardgame-token` will supply something else on top of the same three modules.
 *
 * ## The two frames, and the trap between them
 *
 * Body/physics space is +X right, +Y UP, +Z toward the viewer: a RIGHT handed
 * triple, and the frame `die-geometry.ts` builds solids in and `dice-sim.ts`
 * simulates in.
 *
 * CSS transform space is +X right, +Y DOWN, +Z toward the viewer: the same three
 * directions with one axis reversed, which makes it a LEFT handed triple.
 *
 * The map between them is therefore a REFLECTION and cannot be anything else.
 * `CSS_AXIS_SIGN` is that reflection, `S = diag(1, -1, 1)`, and `toScreen` is it
 * applied to a direction or a point. It is exported from here because it is
 * consumed in two places that must not be allowed to drift apart:
 *
 *   - `facet-placement.ts`, which places each polygon's box, and
 *   - `motion/dice-bake.ts`, which conjugates a simulated pose by the same `S`.
 *
 * A facet placed by one convention and rotated by a matrix built in the other
 * renders as a MIRRORED solid, with the pair of faces along the axis they
 * disagree on swapped — so a die lands showing the face OPPOSITE the one the
 * physics turned up. That is not hypothetical: this map originally shipped as
 * `(x, -y, -z)`, which is a proper rotation (determinant +1) rather than a
 * reflection, and it did exactly that. Every unit test passed while it was live.
 * `screen-frame.test.ts` now pins `det(S) = -1` directly, which is the single
 * assertion that catches it, and `die-roll.spec.ts` measures the consequence
 * end to end (the winding of a right-handed triple of face normals on screen).
 *
 * The tempting "fix" when solids come out inside-out is to flip `WORLD_UP` in
 * `die-faces.ts`, or to negate a second axis here. Do neither: the flip is a
 * property of the RENDERING frame, and it has to be an odd number of axes.
 */

// The vector vocabulary happens to live in `die-geometry.ts`; nothing about it
// is die-specific and nothing die-specific is imported here. See this
// directory's README-in-comments above: a later change should lift `Vec3` and
// its helpers into a module of their own, at which point only this import line
// moves.
import { cross, magnitude, normalize, vec3, type Vec3 } from '../motion/die-geometry.ts';

export type { Vec3 };

/**
 * The physics-to-CSS axis signs: `S = diag(1, -1, 1)`.
 *
 * THE SINGLE PLACE THE Y FLIP LIVES. `toScreen` applies it to a vector;
 * `dice-bake.ts` applies it to a whole pose as the similarity `S R S` (i.e.
 * `R'_ij = s_i s_j R_ij`) plus `p -> S p`. Adding a consumer means importing
 * this, never restating it.
 */
export const CSS_AXIS_SIGN = [1, -1, 1] as const;

/**
 * Body frame (+Y up, right handed) to CSS frame (+Y down): flip Y, and Y only.
 *
 * A point at body `(x, y, z)`, seen by a camera on the body's +Z side, appears
 * `x` to the right, `y` UP and `z` toward the viewer. CSS places `(cx, cy, cz)`
 * at `cx` right, `cy` DOWN and `cz` toward the viewer, so `cy = -y` and nothing
 * else moves. See the file docs for why this must have determinant -1.
 */
export function toScreen(v: Vec3): Vec3 {
  return vec3(v[0] * CSS_AXIS_SIGN[0], v[1] * CSS_AXIS_SIGN[1], v[2] * CSS_AXIS_SIGN[2]);
}

/** Screen up in CSS space: CSS y points down. */
export const SCREEN_UP: Vec3 = vec3(0, -1, 0);

/** Straight at the viewer, in CSS space: the direction a facet must face. */
export const CAMERA_AXIS: Vec3 = vec3(0, 0, 1);

/** Short, stable decimal text: keeps generated style strings readable. */
export function cssNumber(value: number): string {
  const rounded = Number(value.toFixed(5));
  return Object.is(rounded, -0) ? '0' : String(rounded);
}

/**
 * An orthonormal, right-handed basis `(u, v, w)` for a facet's own plane, with
 * `w` the outward normal. `u` is the facet's local +x (screen right when the
 * facet faces the camera) and `v` its local +y (screen down), so the CSS box
 * lands the same way up as the screen wherever the facet can be read.
 *
 * `det([u v w]) = u . ((w x u) x w) = 1`, so this is a rotation and never a
 * reflection, whichever branch is taken. That matters: this basis becomes the
 * upper 3x3 of the facet's `matrix3d`, and a reflection there would mirror the
 * facet's own content while leaving the solid's silhouette unchanged — a defect
 * a screenshot barely shows and `screen-frame.test.ts` catches outright.
 *
 * `w` is expected in CSS space and expected to be a unit vector; callers
 * normalize `toScreen(normal)` before handing it over.
 */
export function facetBasis(w: Vec3): { u: Vec3; v: Vec3 } {
  let u = cross(w, SCREEN_UP);
  // Degenerate exactly for the facets that point straight up or straight
  // down the screen (a d6's top and bottom): any perpendicular will do.
  if (magnitude(u) < 1e-6) u = cross(w, vec3(0, 0, 1));
  const unitU = normalize(u);
  return { u: unitU, v: cross(w, unitU) };
}
