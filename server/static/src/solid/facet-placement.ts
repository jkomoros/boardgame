/**
 * Drawing a closed surface as a solid: ONE polygon becomes ONE element.
 *
 * A caller hands over a closed surface — `[...faces, ...capFaces]` — where every
 * polygon carries its own outward `normal`, its `centroid` and its vertex loop.
 * Everything below turns one such polygon into one absolutely positioned
 * element, and it is deliberately a single routine: a d6's rectangles, a d20's
 * triangles, a d12's pentagons, a d10's kites, a barrel's non-square side
 * rectangles plus cap triangles, and whatever prism or slab a 3D token turns out
 * to be, all go through it unchanged. NOTHING HERE MAY ASSUME A SQUARE FACET —
 * a d7's side face is 2.37 long by 0.87 wide.
 *
 * Coordinate frames. The surface is in the physics frame (+Y up, right handed);
 * `screen-frame.ts` owns the map into CSS's left-handed frame and is the only
 * place that map exists.
 *
 * Units. Lengths are emitted in `em`, and the caller's stage element is expected
 * to set `font-size` to the solid's overall size, so `1em` is that size and the
 * whole solid scales with a custom property with no JavaScript remeasurement.
 * That is the only reason a caller can set a size to anything (`120px`, `6rem`,
 * `10vmin`) and have the solid follow. Geometry units are scaled by
 * `0.5 / circumradius` so the solid's bounding sphere is exactly `1em` across
 * whatever its face count — `die-geometry.ts` builds each solid at its own
 * natural scale (circumradius 1.000 for a d8, 1.902 for a d20) and documents
 * that consumers must normalize themselves.
 *
 * ## The content square, which is what makes one code path enough
 *
 * Every mark a facet carries is laid out inside its CONTENT SQUARE — the largest
 * axis-aligned square that fits inside that facet's own polygon, shrunk by
 * `CONTENT_MARGIN` — and never inside the facet's bounding box. The distinction
 * is the whole game on a solid whose facets are not squares: a d7 side face is
 * 2.37 by 0.87, so a pip grid sized to the box smears the dots along the barrel
 * and a numeral sized to the box hangs off both ends. Sized to the inscribed
 * square instead, one layout draws a legible face on a cube's square, a d20's
 * triangle, a d12's pentagon, a d10's kite and a barrel's long rectangle.
 *
 * This module derives the square and publishes it as four CSS custom properties
 * plus its side in `em`. WHAT GOES IN IT is not this module's business.
 */

import { dot, normalize, scale as scaleVec, subtract } from '../motion/die-geometry.ts';
import { cssNumber, facetBasis, toScreen, type Vec3 } from './screen-frame.ts';

/**
 * One polygon of a closed surface. Structurally what `die-geometry.ts`'s
 * `DieFace` is, named without the die so a token can produce one.
 */
export interface SolidFace {
  readonly normal: Vec3;
  readonly centroid: Vec3;
  /** Ordered counter-clockwise seen from outside the solid. */
  readonly polygon: readonly Vec3[];
}

/**
 * A whole closed surface. Structurally what `die-geometry.ts`'s `DieGeometry`
 * is, narrowed to the three fields rendering needs: a `DieGeometry` may be
 * passed wherever this is asked for.
 *
 * `faces` are the polygons that carry content and `capFaces` the rest of the
 * closed surface; `[...faces, ...capFaces]` is the complete surface with no
 * polygon appearing twice, so a face index is also an index into the facet list
 * this module returns.
 */
export interface SolidSurface {
  readonly faces: readonly SolidFace[];
  readonly capFaces: readonly SolidFace[];
  /** Distance from the centroid to the farthest vertex, in the surface's own units. */
  readonly circumradius: number;
}

/**
 * The content square is this fraction of the largest square that fits inside
 * the polygon, so the marks have air around them rather than touching the
 * facet's edges. A cube's facet IS its own inscribed square, so this is also
 * how much of a d6's face the pips occupy: about what the flat die they
 * replace used (which was 0.63 of a reel face, the number `#inner.reel .face`
 * still sets so the fallback keeps its original look).
 */
const CONTENT_MARGIN = 0.72;

/**
 * A corner mark's square sits this far along the line from its vertex to the
 * facet's centroid. Scanned rather than fixed: the mark wants to be as close
 * to the corner as it can while still being big enough to read, and how far
 * in that is depends on how sharp the corner is (a d4's 60-degree triangle
 * corner needs more inset than a barrel face's right angle).
 */
const CORNER_INSET_MIN = 0.18;
const CORNER_INSET_MAX = 0.65;
const CORNER_INSET_STEPS = 12;

/**
 * A corner mark never grows past this fraction of the centre content square.
 *
 * The cap and the inset range above are a LEGIBILITY budget, not a taste
 * knob: a corner mark is the only thing a corner-printed die (a d4, an
 * odd-sided barrel) shows the player, so it has to survive at the size a real
 * board draws a die at. Measured on a 100px die -- pig's -- a d7's side facet
 * is an 18-by-50px strip carrying four corner marks and a centre numeral, and
 * a barrel's short dimension is what bounds every square inscribed in it: at
 * the previous 0.4/0.45/0.78 the mark came out a 4.1px font, which is dust.
 * These values put it at 6.5px and a d4's at 12.2px with no pair of marks
 * overlapping; the next step up (a 0.65 cap) runs a d7's marks into each other
 * and a d4's into its centre numeral, measured. `die-shape.spec.ts` pins both
 * ends of that: nothing smaller than 6px, and nothing outside its own facet.
 */
const CORNER_MAX_SIZE = 0.6;

/** A point in a facet's own plane; `a` is along the box's local x, `b` its y. */
export interface PlanePoint {
  readonly a: number;
  readonly b: number;
}

/**
 * Half the side of the largest axis-aligned square centred at `(cx, cy)` that
 * fits inside the convex polygon `points`, or 0 when that point is outside it.
 *
 * For each edge with inward unit normal `n` at signed distance `d` from the
 * centre, a square of half-side `s` clears the edge exactly while
 * `s * (|n.a| + |n.b|) <= d` -- its farthest corner in the direction of `-n`.
 * Every facet `die-geometry.ts` produces is convex, so the minimum over the
 * edges is the answer.
 */
export function inscribedSquareHalfSide(
  points: readonly PlanePoint[],
  cx: number,
  cy: number,
): number {
  let twiceArea = 0;
  for (let i = 0; i < points.length; i++) {
    const p = points[i];
    const q = points[(i + 1) % points.length];
    twiceArea += p.a * q.b - q.a * p.b;
  }
  // Winding is whatever the projection made of it; take it from the polygon
  // rather than assuming, so the normals below point inward either way.
  const sign = twiceArea >= 0 ? 1 : -1;
  let best = Infinity;
  for (let i = 0; i < points.length; i++) {
    const p = points[i];
    const q = points[(i + 1) % points.length];
    const ea = q.a - p.a;
    const eb = q.b - p.b;
    const length = Math.hypot(ea, eb);
    if (!(length > 0)) continue;
    // Rotating the edge direction a quarter turn counter-clockwise points to
    // its left, which is the interior for a counter-clockwise winding.
    const na = (-eb / length) * sign;
    const nb = (ea / length) * sign;
    const distance = na * (cx - p.a) + nb * (cy - p.b);
    if (!(distance > 0)) return 0;
    best = Math.min(best, distance / (Math.abs(na) + Math.abs(nb)));
  }
  return Number.isFinite(best) ? best : 0;
}

/**
 * One value printed at one corner of one facet, as the box it goes in.
 *
 * Everything is a PERCENTAGE of the facet's own box on purpose: the mark sets
 * a `font-size` on the text inside it, and an `em` length on the same element
 * would then resolve against that new font-size instead of the solid's size.
 */
export interface SolidCornerMark {
  /** Which face's content this corner carries. Index into `faces`. */
  readonly faceIndex: number;
  readonly left: number;
  readonly top: number;
  readonly width: number;
  readonly height: number;
  /** The square's side in `em`, for sizing the text inside it. */
  readonly size: number;
}

/** One facet of the solid, as the CSS needed to draw it. */
export interface SolidFacet {
  /** Stable key: index into `[...faces, ...capFaces]`. */
  readonly key: number;
  /** Index into `faces` (and so into any per-face content list), or -1 for a cap. */
  readonly faceIndex: number;
  readonly style: string;
  /**
   * The content square's side in `em` — the same number `style` publishes as
   * `--content-size`, in a form a caller can do arithmetic on.
   *
   * Published because how big a mark comes out is a fact about the GEOMETRY,
   * and the only place it can be checked before the browser has drawn anything.
   * `boardgame-die.ts` multiplies it by the die's pixel size to decide whether
   * the marks on this shape are big enough to read (see its legibility floor).
   * Deriving that from here rather than from a constant is what keeps the check
   * honest when a solid's proportions change.
   */
  readonly contentSize: number;
  /** Empty unless the caller asked for corner marks on this facet. */
  readonly corners: readonly SolidCornerMark[];
}

/**
 * Place one surface polygon.
 *
 * The element is a plain box whose centre sits at the solid's centre (`left`/
 * `top` at 50% less half its size), so `transform-origin` — its own centre —
 * is the solid's origin. `translate3d(...) matrix3d(...)` then rotates the box
 * into the facet's plane about that origin and moves it out to the facet, the
 * translation being read in the PARENT's frame because CSS applies a transform
 * list left to right.
 *
 * The box is the facet's own bounding rectangle in its own plane — NOT a
 * square, and not centred on the polygon's centroid either (a triangle's
 * vertex mean is not the centre of its bounding box), so the translation
 * carries the bounding-box offset as well. `clip-path` then cuts the box down
 * to the actual polygon, which is what makes triangles, kites, pentagons and
 * rectangles one code path.
 *
 * It also derives where CONTENT goes on that facet, in the same one routine
 * and from the same polygon: the content square (`--content-left` and
 * friends), and one corner mark per vertex when `cornerFaces` is supplied —
 * `cornerFaces[i]` being which face's content the mark at vertex `i` carries.
 * Nothing downstream of here knows what shape it is drawing on, or what the
 * marks say.
 */
export function facetPlacement(
  face: SolidFace,
  unitsToEm: number,
  cornerFaces: readonly number[] | null,
): { style: string; contentSize: number; corners: readonly SolidCornerMark[] } {
  const w = normalize(toScreen(face.normal));
  const { u, v } = facetBasis(w);
  const centre = scaleVec(toScreen(face.centroid), unitsToEm);
  const points = face.polygon.map((point) => {
    const offset = subtract(scaleVec(toScreen(point), unitsToEm), centre);
    return { a: dot(offset, u), b: dot(offset, v) };
  });
  const minA = Math.min(...points.map((p) => p.a));
  const maxA = Math.max(...points.map((p) => p.a));
  const minB = Math.min(...points.map((p) => p.b));
  const maxB = Math.max(...points.map((p) => p.b));
  const width = maxA - minA;
  const height = maxB - minB;
  if (!(width > 0) || !(height > 0)) {
    throw new Error(`degenerate facet: ${width} x ${height} bounding box`);
  }
  // The box centre in the facet's plane, then out into the parent's frame.
  const boxA = (minA + maxA) / 2;
  const boxB = (minB + maxB) / 2;
  const t = [0, 1, 2].map((i) => centre[i] + u[i] * boxA + v[i] * boxB);
  const clip = points
    .map((p) => `${cssNumber(((p.a - minA) / width) * 100)}% ${cssNumber(((p.b - minB) / height) * 100)}%`)
    .join(', ');

  // The content square: the biggest square that fits inside THIS polygon,
  // centred on its centroid (which is the origin of the (a, b) frame), with a
  // margin. Not the bounding box: on a d7's 2.37-by-0.87 side face the
  // bounding box is nearly three times as long as it is wide, and a pip grid
  // or a numeral sized to it smears along the barrel.
  const contentSize = 2 * inscribedSquareHalfSide(points, 0, 0) * CONTENT_MARGIN;
  // Percentages of the facet's box, so nothing downstream has to know the
  // facet's size and no `em` is at the mercy of a font-size set on the mark.
  const asBox = (a: number, b: number, size: number) => ({
    left: ((a - boxA - size / 2) / width + 0.5) * 100,
    top: ((b - boxB - size / 2) / height + 0.5) * 100,
    width: (size / width) * 100,
    height: (size / height) * 100,
  });
  const contentBox = asBox(0, 0, contentSize);

  const corners: SolidCornerMark[] = cornerFaces === null ? [] : points.map((point, index) => {
    // Walk in from the vertex towards the centroid until the square that fits
    // there is big enough, and stop there: as near the corner as it can be.
    const cap = contentSize * CORNER_MAX_SIZE;
    let best = { size: 0, a: point.a, b: point.b };
    for (let step = 0; step <= CORNER_INSET_STEPS; step++) {
      const t = CORNER_INSET_MIN
        + ((CORNER_INSET_MAX - CORNER_INSET_MIN) * step) / CORNER_INSET_STEPS;
      const a = point.a * (1 - t);
      const b = point.b * (1 - t);
      const size = 2 * inscribedSquareHalfSide(points, a, b) * CONTENT_MARGIN;
      if (size > best.size) best = { size, a, b };
      if (best.size >= cap) break;
    }
    const size = Math.min(best.size, cap);
    return { faceIndex: cornerFaces[index], size, ...asBox(best.a, best.b, size) };
  });

  const style = [
    `width:${cssNumber(width)}em`,
    `height:${cssNumber(height)}em`,
    `margin-left:${cssNumber(-width / 2)}em`,
    `margin-top:${cssNumber(-height / 2)}em`,
    `transform:translate3d(${cssNumber(t[0])}em,${cssNumber(t[1])}em,${cssNumber(t[2])}em) `
      + `matrix3d(${[u, v, w].map((axis) => `${cssNumber(axis[0])},${cssNumber(axis[1])},${cssNumber(axis[2])},0`).join(',')},0,0,0,1)`,
    `clip-path:polygon(${clip})`,
    `--content-left:${cssNumber(contentBox.left)}%`,
    `--content-top:${cssNumber(contentBox.top)}%`,
    `--content-width:${cssNumber(contentBox.width)}%`,
    `--content-height:${cssNumber(contentBox.height)}%`,
    // The one length here that is an `em`, because the marks inside the
    // content square size themselves against it. It resolves against the
    // facet's INHERITED font-size (the solid's size, set once on the stage).
    // Nothing may set `font-size` on the facet itself: its own width/height
    // are `em` too and would resize themselves, which is why the text sizes a
    // child span instead.
    `--content-size:${cssNumber(contentSize)}em`,
  ].join(';');
  return { style, contentSize, corners };
}

export interface SolidFacetOptions {
  /**
   * Which face's content the mark at a given VERTEX of a readable facet
   * carries, for a solid that is read from a face nobody can see.
   *
   * Supplying it is what turns corner printing on. Omitted, no facet gets
   * corner marks, which is what an ordinary solid wants. Called only for the
   * readable faces — caps carry nothing.
   */
  readonly cornerOwner?: (vertex: Vec3) => number;
}

/**
 * The whole surface, as one facet per polygon in surface order
 * (`[...faces, ...capFaces]`, so a face index is also an index into the result).
 *
 * Normalizes the surface to a bounding sphere exactly `1em` across, which is
 * what lets a caller size the solid with one custom property and lets a solid
 * of any face count fit its box in every orientation.
 */
export function solidFacets(
  surface: SolidSurface,
  options: SolidFacetOptions = {},
): readonly SolidFacet[] {
  const unitsToEm = 0.5 / surface.circumradius;
  const cornerOwner = options.cornerOwner;
  return [...surface.faces, ...surface.capFaces].map((face, key) => {
    const readable = key < surface.faces.length;
    const cornerFaces = cornerOwner && readable
      ? face.polygon.map((vertex) => cornerOwner(vertex))
      : null;
    const { style, contentSize, corners } = facetPlacement(face, unitsToEm, cornerFaces);
    return { key, faceIndex: readable ? key : -1, style, contentSize, corners };
  });
}
