/**
 * A flat-capped right prism: N side walls around an axis, closed at both ends
 * by a flat polygon.
 *
 * This is the shape a 3D `boardgame-token` is. A `token` is a chunky prism
 * (height about 0.55 of its diameter), a `chip` a thin one (0.13), and a `disc`
 * thinner still — one builder, three ratios. It is deliberately NOT a die: it
 * has no face values, no reading rule, no physics and no inertia tensor, and
 * its only imports are two TYPES from this directory, both erased at runtime,
 * so nothing named for a die is reachable from here at all. `src/solid/` exists
 * so a non-die component can render a solid, and this module is the first thing
 * that actually proves it.
 *
 * ## The axis points at the camera, not at the ceiling
 *
 * The prism is built around +Z, which in the body frame is straight at the
 * viewer (`screen-frame.ts`'s `CAMERA_AXIS` after the map into CSS). So a chip
 * with no pose applied shows its flat cap face-on and its silhouette IS the
 * cross-section polygon — which is the whole sizing contract for a token: the
 * solid's drawn extent fills the box the flat SVG used to fill, rather than the
 * die's contract where the box is a bounding sphere with a reserved footprint.
 * A caller that wants the prism lying on a board tilts it; the tilt is a pose,
 * and a pose belongs to the component, not to the geometry.
 *
 * ## Units
 *
 * The cross-section is built on a circumcircle of radius 1, so the prism is 2
 * across and `heightRatio` is height over that diameter. `nominalRadius` is 1,
 * i.e. `1em` buys the cross-section's WIDTH — see `SolidSurface.nominalRadius`
 * and `boundingRadius` below for the part that overflows.
 */

import type { SolidFace, SolidSurface } from './facet-placement.ts';
import type { Vec3 } from './screen-frame.ts';

/** The cross-section's circumradius. Everything else is stated against it. */
const CIRCUMRADIUS = 1;

/**
 * A prism, as the closed surface `facet-placement.ts` draws, plus the two facts
 * a component laying it out needs that a bare `SolidSurface` does not carry.
 */
export interface PrismSurface extends SolidSurface {
  /**
   * Every corner of the solid, bottom ring then top ring, each in the same
   * order the cross-section walks. Face polygons reference these exact frozen
   * tuples, so two polygons meeting along an edge share vertex IDENTITY and not
   * merely equal numbers — which is what makes the closed-manifold check on
   * this surface a check and not an approximation.
   */
  readonly vertices: readonly Vec3[];
  /**
   * Distance from the centre to the farthest corner. Always LARGER than
   * `nominalRadius`, because a prism's diagonal is longer than its width: a
   * caller that spins one has to reserve `boundingRadius / nominalRadius` times
   * the box, exactly as `boardgame-die.ts` does for a barrel. A caller that
   * only tilts it slightly does not.
   */
  readonly boundingRadius: number;
}

/** Newell's method: the plane normal of an ordered loop, CCW seen from outside. */
function loopNormal(polygon: readonly Vec3[]): Vec3 {
  let x = 0;
  let y = 0;
  let z = 0;
  for (let i = 0; i < polygon.length; i++) {
    const current = polygon[i];
    const next = polygon[(i + 1) % polygon.length];
    x += (current[1] - next[1]) * (current[2] + next[2]);
    y += (current[2] - next[2]) * (current[0] + next[0]);
    z += (current[0] - next[0]) * (current[1] + next[1]);
  }
  const length = Math.sqrt(x * x + y * y + z * z);
  if (!(length > 0)) throw new Error('degenerate prism facet: zero-area loop');
  return Object.freeze([x / length, y / length, z / length] as const);
}

function loopCentroid(polygon: readonly Vec3[]): Vec3 {
  let x = 0;
  let y = 0;
  let z = 0;
  for (const point of polygon) {
    x += point[0];
    y += point[1];
    z += point[2];
  }
  const n = polygon.length;
  return Object.freeze([x / n, y / n, z / n] as const);
}

/**
 * The normal and centroid are DERIVED from the loop rather than stated
 * alongside it, so a polygon and the normal a renderer rotates it by cannot
 * drift apart. `prism.test.ts` cross-checks both against the closed form.
 */
function faceOf(polygon: readonly Vec3[]): SolidFace {
  return Object.freeze({
    normal: loopNormal(polygon),
    centroid: loopCentroid(polygon),
    polygon: Object.freeze(polygon),
  });
}

/**
 * Build a right prism with `sides` side walls and a height of `heightRatio`
 * times its cross-section diameter.
 *
 * `sides` is the whole facet budget for the shape: the surface is `sides + 2`
 * polygons and therefore `sides + 2` DOM elements, one per facet. Measured, 55
 * tokens on screen at 24 sides is 1,430 facet elements and 42.8fps at rest with
 * no layer promotion; at 12 sides it is 770 elements and a steady 60fps. The
 * budget is roughly 800 elements free and 1,400 a cliff, so 12 is the ceiling a
 * stack-hosted token gets, and raising it means re-measuring rather than
 * re-reasoning.
 *
 * The caps are flat polygons, not the cones `motion/die-geometry.ts` builds for
 * a barrel: a barrel's caps are shaped so the die cannot come to rest on one,
 * which is a physics requirement, and a token has no physics.
 */
export function prismSurface(sides: number, heightRatio: number): PrismSurface {
  if (!Number.isInteger(sides) || sides < 3) {
    throw new Error(`a prism needs an integer side count >= 3, got ${sides}`);
  }
  if (!Number.isFinite(heightRatio) || heightRatio <= 0) {
    throw new Error(`a prism needs a positive height ratio, got ${heightRatio}`);
  }
  // heightRatio is height over DIAMETER, so the half-height is the ratio times
  // the circumradius. A "flat" disc is a small ratio, never zero: a zero-height
  // prism has no side walls to draw and encloses nothing.
  const halfHeight = heightRatio * CIRCUMRADIUS;

  const bottom: Vec3[] = [];
  const top: Vec3[] = [];
  for (let i = 0; i < sides; i++) {
    const angle = (2 * Math.PI * i) / sides;
    const x = CIRCUMRADIUS * Math.cos(angle);
    const y = CIRCUMRADIUS * Math.sin(angle);
    bottom.push(Object.freeze([x, y, -halfHeight] as const));
    top.push(Object.freeze([x, y, halfHeight] as const));
  }

  // Winding, all of it: the cross-section is walked counter-clockwise seen from
  // +Z, so the +Z cap keeps that order and the -Z cap reverses it, and a wall
  // taken bottom-edge-then-up is counter-clockwise seen from outside. Nothing
  // here is rederived from a centroid, which is why a non-convex cross-section
  // would work the same way (see `RawSolid.oriented` in the die geometry, the
  // seam that had to learn the same lesson).
  const faces: SolidFace[] = [
    faceOf(top),
    faceOf([...bottom].reverse()),
  ];
  for (let i = 0; i < sides; i++) {
    const j = (i + 1) % sides;
    faces.push(faceOf([bottom[i], bottom[j], top[j], top[i]]));
  }

  return Object.freeze({
    faces: Object.freeze(faces),
    // Empty, as it is for every closed-form solid. `capFaces` means "part of
    // the surface that carries no value", and a prism's flat caps are the
    // polygons a token is MOST likely to print something on, so calling them
    // caps here would be the wrong word for the right shape.
    capFaces: Object.freeze([]),
    vertices: Object.freeze([...bottom, ...top]),
    nominalRadius: CIRCUMRADIUS,
    boundingRadius: Math.hypot(CIRCUMRADIUS, halfHeight),
  });
}
