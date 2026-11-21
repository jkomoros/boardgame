/**
 * Drawing a solid whose POSE IS A CONSTANT: one already-projected polygon
 * becomes one ordinary, untransformed element.
 *
 * `facet-placement.ts` is the other half of this module's job and the one a die
 * needs: it hands each polygon a `matrix3d` that stands it up in a live
 * `preserve-3d` scene, so the browser can re-project it every frame of a tumble.
 * A `boardgame-token` never tumbles. Its pose is a constant, so the projection
 * of every facet is a constant too, and it can be done ONCE, here, in
 * JavaScript — leaving the DOM with nothing but flat, coloured polygons.
 *
 * ## Why that is not merely a simplification
 *
 * It is the difference between 60fps and 30fps on a real board, and the reason
 * is layers, not arithmetic. Chromium hands every element inside a live
 * `preserve-3d` context its own composited layer the moment an ANCESTOR
 * transform starts animating — which is exactly what a stack's FLIP does to a
 * component host on every move. Measured in `pass` (55 tokens, 14 facets each):
 * during a move the token subtrees went from 57 composited layers to **1,047**,
 * and from 1.6 to **88.6 megapixels** of layer area, because a clip-path'ed
 * facet under a `perspective` gets conservative layer bounds two thousand pixels
 * across. Nothing authored on the token asked for that: the promotion is a
 * consequence of the 3D context existing at all while something above it moves.
 *
 * All three of `perspective`, `transform-style: preserve-3d` and the facets'
 * own 3D transforms have to go for it to stop — measured one at a time, and any
 * two of the three still leaves ~1,000 layers. What is left is what this module
 * emits, and it composites exactly like the flat SVG art it replaced.
 *
 * ## Why the picture is the same picture
 *
 * A convex solid's front-facing facets tile its silhouette exactly once, with no
 * overlap — which is the same fact that made `backface-visibility: hidden` a
 * sufficient hidden-surface removal in the 3D version. So the caller can decide
 * facing itself, drop the back faces, and paint the rest in any order at all:
 * there is nothing left to sort. And each facet is filled with ONE flat colour,
 * so its projection is a flat-filled convex polygon and `clip-path: polygon()`
 * draws it exactly. The two renderings are pixel-identical off the seams;
 * `token-flat-truth.spec.ts` measures the difference and requires it to be a
 * one-pixel anti-aliasing band and nothing else.
 *
 * ## Units
 *
 * Points arrive in `em`, measured from the solid's centre, `+x` right and `+y`
 * DOWN — CSS's own screen frame, which is what `screen-frame.ts` maps into. The
 * caller's stage sets `font-size` to the solid's size, exactly as for
 * `facet-placement.ts`, so the whole thing follows a custom property with no
 * JavaScript remeasurement.
 */

import { cssNumber } from './screen-frame.ts';

/** A projected vertex, in `em` from the solid's centre; `+y` is down. */
export interface FlatPoint {
  readonly x: number;
  readonly y: number;
}

/**
 * The complete inline style of the one element that draws this projected
 * polygon, minus its fill.
 *
 * The element is its own bounding box rather than the whole stage, for the
 * reason `facet-placement.ts` sizes a facet to its polygon: a `clip-path` is
 * clipped by the element's own border box, so a box that merely happens to be
 * big enough today is a silhouette that gets shaved when the fit changes by a
 * rounding step. Sized to the polygon, the outermost vertices land ON the box's
 * edge by construction.
 *
 * Positioned from the CENTRE (`left: 50%` plus a negative margin), because the
 * solid is centred in the token's box and the projection is measured from that
 * centre. The polygon itself is then expressed in percentages of the box, which
 * is what makes it survive the box being rounded to device pixels.
 */
export function flatFacetStyle(points: readonly FlatPoint[]): string {
  if (points.length < 3) {
    throw new Error(`flat facet: a polygon needs 3 points, got ${points.length}`);
  }
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;
  for (const point of points) {
    if (!Number.isFinite(point.x) || !Number.isFinite(point.y)) {
      throw new Error('flat facet: a projected vertex was not finite');
    }
    minX = Math.min(minX, point.x);
    minY = Math.min(minY, point.y);
    maxX = Math.max(maxX, point.x);
    maxY = Math.max(maxY, point.y);
  }
  const width = maxX - minX;
  const height = maxY - minY;
  if (!(width > 0) || !(height > 0)) {
    throw new Error('flat facet: a projected polygon collapsed to a line');
  }
  // Percent within the box. An edge-on facet is the caller's to cull -- by the
  // time it gets here the box has area, so neither divisor can be zero.
  const polygon = points
    .map((point) => `${cssNumber(((point.x - minX) / width) * 100)}% `
      + `${cssNumber(((point.y - minY) / height) * 100)}%`)
    .join(',');
  return [
    'position:absolute',
    'left:50%',
    'top:50%',
    `width:${cssNumber(width)}em`,
    `height:${cssNumber(height)}em`,
    `margin-left:${cssNumber(minX)}em`,
    `margin-top:${cssNumber(minY)}em`,
    `clip-path:polygon(${polygon})`,
  ].join(';');
}
