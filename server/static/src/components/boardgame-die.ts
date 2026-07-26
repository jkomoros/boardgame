import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';
import { componentMotionTracks } from '../motion/component-track.js';
import type { ComponentMotionTarget } from '../motion/component-track.js';
import {
  add,
  cross,
  dieGeometry,
  dot,
  magnitude,
  normalize,
  scale as scaleVec,
  subtract,
  vec3,
  type DieFace,
  type DieGeometry,
  type Vec3,
} from '../motion/die-geometry.js';
import { resolveReadingRule } from '../motion/die-faces.js';

/**
 * Drawing a die as a solid.
 *
 * `die-geometry.ts` hands over a closed surface — `[...faces, ...capFaces]` —
 * where every polygon carries its own outward `normal`, its `centroid` and its
 * vertex loop. Everything below turns ONE such polygon into ONE absolutely
 * positioned element, and it is deliberately a single routine: a d6's
 * rectangles, a d20's triangles, a d12's pentagons, a d10's kites and a
 * barrel's non-square side rectangles plus cap triangles all go through it
 * unchanged. Nothing here may assume a square facet — a d7's side face is
 * 2.37 long by 0.87 wide.
 *
 * Coordinate frames. The geometry is in the physics frame (+Y up, right
 * handed). CSS screen space is x right, y DOWN, z toward the viewer. The two
 * are related by a 180-degree rotation about X, `(x, y, z) -> (x, -y, -z)`,
 * which is a PROPER rotation: simply flipping Y would mirror the solid, and a
 * mirrored facet renders its glyphs backwards (task 9 paints those).
 *
 * Units. Lengths are emitted in `em`, and `#stage` sets `font-size:
 * var(--effective-die-size)`, so `1em` is the die's size and the whole solid
 * scales with the custom property with no JavaScript remeasurement. That is
 * the only reason a caller can set `--die-size` to anything (`120px`, `6rem`,
 * `10vmin`) and have the solid follow. Geometry units are scaled by
 * `0.5 / circumradius` so the die's bounding sphere is exactly `1em` across
 * whatever its face count — `die-geometry.ts` builds each solid at its own
 * natural scale (circumradius 1.000 for a d8, 1.902 for a d20) and documents
 * that consumers must normalize themselves.
 */

/** Screen up in CSS space: CSS y points down. */
const SCREEN_UP: Vec3 = vec3(0, -1, 0);

/** Straight at the viewer, in CSS space: the direction a facet must face. */
const CAMERA_AXIS: Vec3 = vec3(0, 0, 1);

/**
 * Where the presented face is pointed, in CSS space, when the die is at rest.
 *
 * Not straight at the camera (`+Z`): a solid facing the viewer square-on
 * projects to a flat outline and reads as the 2D die this replaces. Pointing
 * the presented face slightly down and to the left instead puts the camera
 * above and to the right of it, so a d6 shows its presented face plus the
 * faces above and to its right — a die seen on a table. The physics-driven
 * resting pose replaces this when the roll is wired up.
 *
 * This fixed tilt is enough ONLY while the solid's other faces are within
 * ~90 degrees of the presented one; `companionTilt` covers the rest.
 */
const RESTING_VIEW: Vec3 = normalize(vec3(-0.32, 0.26, 1));

/**
 * How far off the camera axis the most face-on of the OTHER facets is allowed
 * to sit, in degrees. Past 90 it is a back-face and `backface-visibility:
 * hidden` culls it outright, so a fixed tilt of 23.6 degrees (`RESTING_VIEW`)
 * is not enough for a solid whose faces are far apart in normal angle: a
 * tetrahedron's other three normals are 109.47 degrees from the presented one
 * — 86 to 133 degrees off the camera axis once RESTING_VIEW is applied — so a
 * d4 renders as a single flat triangle, which is exactly the 2D die this
 * replaces. (A d8's are 70.5 apart, a d12's 63.4, a d20's 41.8, a barrel's
 * side faces closer still; none of them need any of this.)
 *
 * 75 rather than a hair under 90 because a facet within a few degrees of
 * edge-on is a sliver, not a visible face.
 */
const COMPANION_VIEW_LIMIT = 75;

/**
 * The most the pose may be tilted to bring that facet into view. The presented
 * face is the value the player has to read, so it stays the dominant one: the
 * tilt moves every direction by at most its own angle, which keeps the
 * presented face within 23.6 + 30 degrees of the camera axis while the facet
 * it reveals sits at 75.
 */
const MAX_COMPANION_TILT = 30;

/**
 * Painting a face.
 *
 * Content resolves in one order, per face: an author-supplied SYMBOL SET
 * first, then generated PIPS, then a NUMERAL. Nothing here enumerates a
 * layout: the die used to stop at six because its pip patterns were six
 * hard-coded CSS classes, and the replacement computes the pattern from the
 * value on a 3x3 lattice.
 *
 * Everything is laid out inside a facet's CONTENT SQUARE -- the largest
 * axis-aligned square that fits inside that facet's own polygon, shrunk by
 * `CONTENT_MARGIN` -- and never inside the facet's bounding box. The
 * distinction is the whole game on a solid whose facets are not squares: a d7
 * side face is 2.37 by 0.87, so a pip grid sized to the box smears the dots
 * along the barrel, and a numeral sized to the box hangs off both ends. Sized
 * to the inscribed square instead, the same code draws a legible face on a
 * cube's square, a d20's triangle, a d12's pentagon, a d10's kite and a
 * barrel's long rectangle.
 */

/** A pip's cell on the 3x3 lattice: [col, row], each 0..2, +row downward. */
type PipCell = readonly [number, number];

/**
 * The lattice cells a pip layout is built from, IN THE ORDER THEY ARE ADDED,
 * as opposite pairs. A layout for `n` is the centre cell when `n` is odd
 * followed by the first `floor(n / 2)` pairs -- which reproduces every
 * familiar die and domino face from 0 to 9 without naming any of them:
 *
 *   0 blank; 1 centre; 2 a diagonal; 3 diagonal + centre; 4 the corners;
 *   5 corners + centre; 6 corners + the side midpoints; 7 six + centre;
 *   8 six + top and bottom midpoints; 9 the full lattice.
 */
const PIP_PAIRS: readonly (readonly [PipCell, PipCell])[] = [
  [[0, 0], [2, 2]],
  [[2, 0], [0, 2]],
  [[0, 1], [2, 1]],
  [[1, 0], [1, 2]],
];

const PIP_CENTRE: PipCell = [1, 1];

/**
 * The largest value still drawn as pips.
 *
 * NINE: the 3x3 lattice that physical dice and dominoes use holds exactly
 * nine, and every count up to it has a canonical symmetric pattern on it. A
 * tenth pip needs a fourth row, which both breaks those familiar patterns and
 * shrinks each dot below what reads at the size a facet actually gets (a d10's
 * kite gives its content square about a third of the die's width). Past nine
 * a numeral is smaller to draw AND faster to read — nobody counts ten dots at
 * a glance — so the die switches over.
 */
const MAX_PIP_VALUE = 9;

/** Pip diameter as a fraction of the content square's side (one lattice cell is a third). */
const PIP_DIAMETER = 0.2;

/** Numeral/glyph height as a fraction of the content square's side. */
const GLYPH_HEIGHT = 0.66;

/**
 * How wide the text may run, as a multiple of the content square, divided by
 * its character count: a two-digit numeral is drawn smaller than a one-digit
 * one so that "20" fits the same square "5" does.
 */
const GLYPH_WIDTH_BUDGET = 1.6;

/** Corner marks are drawn a little taller in their (smaller) square. */
const CORNER_GLYPH_HEIGHT = 0.78;

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
const CORNER_INSET_MAX = 0.45;
const CORNER_INSET_STEPS = 12;

/** A corner mark never grows past this fraction of the centre content square. */
const CORNER_MAX_SIZE = 0.4;

/** The lattice cells for a pip count, computed rather than enumerated. */
function pipCells(count: number): readonly PipCell[] {
  const cells: PipCell[] = [];
  if (count % 2 === 1) cells.push(PIP_CENTRE);
  for (let index = 0; index < Math.floor(count / 2); index++) {
    cells.push(...PIP_PAIRS[index]);
  }
  return cells;
}

/** True for the values `pipCells` has a canonical lattice pattern for. */
function isPipValue(value: number): boolean {
  return Number.isInteger(value) && value >= 0 && value <= MAX_PIP_VALUE;
}

/**
 * Font size for a mark, as a fraction of the square it is drawn in: capped by
 * the square's height, and by a width budget that shrinks with the text's
 * length so a three-character label still fits.
 */
function glyphScale(text: string, heightFraction: number): number {
  return Math.min(heightFraction, GLYPH_WIDTH_BUDGET / Math.max(1, text.length));
}

/** A point in a facet's own plane; `a` is along the box's local x, `b` its y. */
interface PlanePoint {
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
function inscribedSquareHalfSide(points: readonly PlanePoint[], cx: number, cy: number): number {
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

/** Body frame (+Y up) to CSS frame (+Y down): a 180-degree turn about X. */
function toScreen(v: Vec3): Vec3 {
  return vec3(v[0], -v[1], -v[2]);
}

/** Short, stable decimal text: keeps generated style strings readable. */
function num(value: number): string {
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
 * reflection, whichever branch is taken.
 */
function facetBasis(w: Vec3): { u: Vec3; v: Vec3 } {
  let u = cross(w, SCREEN_UP);
  // Degenerate exactly for the facets that point straight up or straight
  // down the screen (a d6's top and bottom): any perpendicular will do.
  if (magnitude(u) < 1e-6) u = cross(w, vec3(0, 0, 1));
  const unitU = normalize(u);
  return { u: unitU, v: cross(w, unitU) };
}

/**
 * One value printed at one corner of one facet, as the box it goes in.
 *
 * Everything is a PERCENTAGE of the facet's own box on purpose: the mark sets
 * a `font-size` on the text inside it, and an `em` length on the same element
 * would then resolve against that new font-size instead of the die's size.
 */
interface DieCornerMark {
  /** Which face's value this corner carries. Index into `faces`. */
  readonly faceIndex: number;
  readonly left: number;
  readonly top: number;
  readonly width: number;
  readonly height: number;
  /** The square's side in `em`, for sizing the text inside it. */
  readonly size: number;
}

/** One facet of the solid, as the CSS needed to draw it. */
interface DieFacet {
  /** Stable key: index into `[...faces, ...capFaces]`. */
  readonly key: number;
  /** Index into `faces` (and so into the `faces` property), or -1 for a cap. */
  readonly faceIndex: number;
  readonly style: string;
  /** Empty unless the die is read from a face nobody can see. */
  readonly corners: readonly DieCornerMark[];
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
 * friends), and one corner mark per vertex when `cornerFaces` says this die is
 * printed at its corners. Nothing downstream of here knows what shape it is
 * drawing on.
 */
function facetStyle(
  face: DieFace,
  unitsToEm: number,
  cornerFaces: readonly number[] | null,
): { style: string; corners: readonly DieCornerMark[] } {
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
    .map((p) => `${num(((p.a - minA) / width) * 100)}% ${num(((p.b - minB) / height) * 100)}%`)
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

  const corners: DieCornerMark[] = cornerFaces === null ? [] : points.map((point, index) => {
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
    `width:${num(width)}em`,
    `height:${num(height)}em`,
    `margin-left:${num(-width / 2)}em`,
    `margin-top:${num(-height / 2)}em`,
    `transform:translate3d(${num(t[0])}em,${num(t[1])}em,${num(t[2])}em) `
      + `matrix3d(${[u, v, w].map((axis) => `${num(axis[0])},${num(axis[1])},${num(axis[2])},0`).join(',')},0,0,0,1)`,
    `clip-path:polygon(${clip})`,
    `--content-left:${num(contentBox.left)}%`,
    `--content-top:${num(contentBox.top)}%`,
    `--content-width:${num(contentBox.width)}%`,
    `--content-height:${num(contentBox.height)}%`,
    // The one length here that is an `em`, because the marks inside the
    // content square size themselves against it. It resolves against the
    // facet's INHERITED font-size (the die's size, set once on `#stage`).
    // Nothing may set `font-size` on the facet itself: its own width/height
    // are `em` too and would resize themselves, which is why the text sizes a
    // child span instead.
    `--content-size:${num(contentSize)}em`,
  ].join(';');
  return { style, corners };
}

/** An axis-angle turn, in the degrees CSS wants. */
interface Turn {
  readonly axis: Vec3;
  readonly degrees: number;
}

function rotate3d(turn: Turn): string {
  const { axis, degrees } = turn;
  return `rotate3d(${num(axis[0])},${num(axis[1])},${num(axis[2])},${num(degrees)}deg)`;
}

/**
 * The minimal rotation carrying direction `from` to direction `to`, or `null`
 * when they already agree: axis `from x to`, angle `atan2(|from x to|, from .
 * to)`. CSS `rotate3d` is the right-handed Rodrigues rotation in the same
 * coordinate triple, so no sign fixing.
 */
function minimalTurn(from: Vec3, to: Vec3): Turn | null {
  const axis = cross(from, to);
  const sine = magnitude(axis);
  const cosine = dot(from, to);
  if (sine < 1e-9) {
    // Already there, or pointing exactly backwards: a half turn about any
    // perpendicular then does it, and `facetBasis` names one.
    if (cosine > 0) return null;
    return { axis: facetBasis(from).u, degrees: 180 };
  }
  return {
    axis: scaleVec(axis, 1 / sine),
    degrees: (Math.atan2(sine, cosine) * 180) / Math.PI,
  };
}

/** Rodrigues: `v` turned about the UNIT axis of `turn`, right-handed. */
function applyTurn(v: Vec3, turn: Turn | null): Vec3 {
  if (!turn) return v;
  const radians = (turn.degrees * Math.PI) / 180;
  const cosine = Math.cos(radians);
  return add(
    add(scaleVec(v, cosine), scaleVec(cross(turn.axis, v), Math.sin(radians))),
    scaleVec(turn.axis, dot(turn.axis, v) * (1 - cosine)),
  );
}

/**
 * The extra turn, if any, that brings `companion` — the most face-on of the
 * facets OTHER than the presented one, already in its resting direction — to
 * `COMPANION_VIEW_LIMIT` of the camera axis, so that the die reads as a solid
 * and not as one flat polygon. `null` when it is visible enough already, which
 * is every solid except the tetrahedron. See `COMPANION_VIEW_LIMIT`.
 *
 * Rotating about `companion x cameraAxis` carries `companion` towards the
 * camera along the shortest path, and moves everything else — the presented
 * face included — by at most the same angle.
 */
function companionTilt(companion: Vec3): Turn | null {
  const cosine = Math.min(1, Math.max(-1, dot(companion, CAMERA_AXIS)));
  const offAxis = (Math.acos(cosine) * 180) / Math.PI;
  const degrees = Math.min(offAxis - COMPANION_VIEW_LIMIT, MAX_COMPANION_TILT);
  if (!(degrees > 0)) return null;
  const axis = cross(companion, CAMERA_AXIS);
  const sine = magnitude(axis);
  // Dead ahead or dead behind: no shortest path to pick, and dead ahead is
  // not a case that needs one anyway.
  if (sine < 1e-9) return null;
  return { axis: scaleVec(axis, 1 / sine), degrees };
}

/**
 * The resting pose: the rotation that points the presented face's normal at
 * `RESTING_VIEW`, then whatever extra tilt it takes for at least one other
 * facet to be visible (`companionTilt`), so the die reads as a solid whatever
 * its face count. CSS applies a transform list left to right, so the extra
 * tilt is written FIRST to be applied after the base pose.
 */
function presentationTransform(geometry: DieGeometry, presented: number): string {
  const base = minimalTurn(normalize(toScreen(geometry.faces[presented].normal)), RESTING_VIEW);
  const surface = [...geometry.faces, ...geometry.capFaces];
  let companion: Vec3 | null = null;
  for (let index = 0; index < surface.length; index++) {
    if (index === presented) continue;
    const direction = applyTurn(normalize(toScreen(surface[index].normal)), base);
    if (companion === null || direction[2] > companion[2]) companion = direction;
  }
  const tilt = companion === null ? null : companionTilt(companion);
  const turns = [tilt, base].filter((turn): turn is Turn => turn !== null);
  return turns.length ? turns.map(rotate3d).join(' ') : 'none';
}

/**
 * Which face a die is READ from when `up` is its topmost direction, for a
 * solid that is not read from an up face at all — the one it is resting on,
 * i.e. the one whose normal is most opposed to `up`. `die-faces.ts` uses the
 * same rule for both of its non-`'up-face'` conventions, so this covers a d4
 * (read from the apex vertex) and an odd-sided barrel (read from the edge
 * pointing at the ceiling) with no case analysis.
 */
function faceReadFrom(geometry: DieGeometry, up: Vec3): number {
  let best = 0;
  let bestScore = Infinity;
  for (let index = 0; index < geometry.faces.length; index++) {
    const score = dot(geometry.faces[index].normal, up);
    if (score < bestScore) {
      bestScore = score;
      best = index;
    }
  }
  return best;
}

/** The full facet list for a face count, plus the geometry it came from. */
interface DieSolid {
  readonly geometry: DieGeometry;
  readonly facets: readonly DieFacet[];
  /**
   * True when this solid presents the face it RESTS ON rather than one a
   * player can see — a d4 (`'top-vertex'`) or any odd-sided barrel
   * (`'down-face'`). Such a die prints its values at its corners.
   */
  readonly cornerPrinted: boolean;
}

// Building a solid runs a convex hull for the closed-form shapes, so it is
// cached per face count. `null` records a face count that has no solid, so a
// malformed die does not retry the failure on every render pass.
//
// Deliberately unbounded: the key is a die's face count, so the cache is
// bounded by the number of DISTINCT dice the loaded games define (a handful),
// not by the number of dice on the board or by anything a player can drive.
const SOLID_CACHE = new Map<number, DieSolid | null>();

function dieSolid(faceCount: number): DieSolid | null {
  const cached = SOLID_CACHE.get(faceCount);
  if (cached !== undefined) return cached;
  let solid: DieSolid | null = null;
  try {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    const surface = [...geometry.faces, ...geometry.capFaces];
    // A d4 and every odd-sided barrel are read from the face they REST ON, so
    // painting the value only at each face's centre lands the result face-down
    // against the table. `die-faces.ts` owns which solids those are; a real d4
    // answers it by printing the value at the CORNERS of the faces that stay
    // visible, and each of those corners carries the value that is read when
    // that corner is the top of the die.
    const cornerPrinted = resolveReadingRule(geometry) !== 'up-face';
    solid = {
      geometry,
      cornerPrinted,
      facets: surface.map((face, key) => {
        const readable = key < geometry.faces.length;
        const cornerFaces = cornerPrinted && readable
          ? face.polygon.map((vertex) => faceReadFrom(geometry, vertex))
          : null;
        const { style, corners } = facetStyle(face, unitsToEm, cornerFaces);
        return { key, faceIndex: readable ? key : -1, style, corners };
      }),
    };
  } catch (error) {
    // A face count with no solid (fewer than 3 faces, or a shape the geometry
    // module rejects) falls back to the reel rather than throwing mid-render.
    // Silently is not good enough: a bug in `facetStyle` would land here too
    // and degrade a d20 into a 20-tall reel with nothing in the console to say
    // why, so say it — once per face count, since the result is then cached.
    console.warn(`boardgame-die: no solid for ${faceCount} faces; falling back to the reel`, error);
    solid = null;
  }
  SOLID_CACHE.set(faceCount, solid);
  return solid;
}

class BoardgameDie extends BoardgameAnimatableItem {
  static override styles = [
    ...(BoardgameAnimatableItem.styles ? [BoardgameAnimatableItem.styles] : []),
    css`
      :host {
        --effective-die-scale: var(--die-scale, 1.0);
        /*
         * --die-size is the die's overall size, and is the property a caller
         * sets: any CSS length ('120px', '6rem', '10vmin'). It is the side of
         * the square box the die is laid out in AND the diameter of the
         * sphere the solid is inscribed in, so a die of any face count fits
         * its box in every orientation -- which is what lets a later task
         * tumble it without it escaping the layout.
         *
         * --effective-die-size is the resolved value everything in here
         * measures against; it is not part of the component's API.
         */
        --effective-die-size: var(--die-size, 50px);
        /*
         * How far #inner scrolls per face of the REEL. One die-size, which is
         * a reel face's height -- except on a solid, which has no reel to
         * scroll and sets it to zero (see #inner.solid). It is a variable of
         * its own rather than a re-definition of --effective-die-size so that
         * zeroing it cannot silently zero anything else below #inner that
         * measures against the die's size.
         */
        --reel-step: var(--effective-die-size);
      }

      #scaler {
        height: calc(var(--effective-die-size) * var(--effective-die-scale));
        width: calc(var(--effective-die-size) * var(--effective-die-scale));
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      #main.disabled {
        cursor: default;
      }

      #main {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        border-radius: 6px;
        background: linear-gradient(135deg, #F5F0E8 0%, #E0D9CE 100%);
        overflow: hidden;
        cursor: pointer;
        box-shadow: 0 2px 4px 0 rgba(60, 40, 20, 0.18),
                    0 1px 5px 0 rgba(60, 40, 20, 0.12),
                    0 3px 1px -2px rgba(60, 40, 20, 0.2),
                    inset 0 1px 0 rgba(255, 255, 255, 0.4);
        transform: scale(var(--effective-die-scale));
        transition: box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1);
        border: 0;
        padding: 0;
      }

      #main.interactive:hover {
        box-shadow: 0 8px 10px 1px rgba(60, 40, 20, 0.14),
                    0 3px 14px 2px rgba(60, 40, 20, 0.12),
                    0 5px 5px -3px rgba(60, 40, 20, 0.4);
      }

      /*
       * In solid mode the die's body is the facets themselves, so #main is
       * only the hit target and the 3D scene's positioning context. It must
       * give up overflow:hidden (which would slice the solid, and which
       * flattens any 3D context put on it) along with the flat card look.
       */
      #main.solid,
      #main.solid.interactive:hover {
        position: relative;
        overflow: visible;
        background: none;
        border-radius: 0;
        box-shadow: none;
      }

      #action-status {
        position: absolute;
        top: 100%;
        width: max-content;
        max-width: 16rem;
        margin-top: 0.25rem;
        color: var(--md-sys-color-error, #ba1a1a);
        font-size: 0.75rem;
      }

      /*
       * The solid: a perspective wrapper (#stage), the preserve-3d carrier
       * (#inner -- the element motionTrackTarget('visual') returns, so a later
       * task animates the tumble on it), a resting-pose carrier (#orient) and
       * one element per surface polygon.
       *
       * #orient exists because #inner's transform belongs to the animation
       * kernel: a resting pose written onto #inner would be replaced outright
       * the moment a spin plays (play() pins composite:'replace'), snapping
       * the solid flat mid-roll.
       *
       * Nothing from #stage down may carry a grouping property -- overflow
       * other than visible, a filter, opacity < 1, clip-path, mask -- because
       * each of those forces transform-style back to flat and collapses the
       * solid into a pile of overlapping outlines. That is why #main.solid
       * gives up the overflow:hidden the reel needs. The facets themselves
       * DO carry clip-path, which is fine: they are leaves of the 3D context.
       */
      #stage {
        position: absolute;
        inset: 0;
        /*
         * The one place the die's size becomes a font-size, so that every
         * generated length below can be an em unit and the whole solid follows
         * --die-size with no JavaScript remeasurement.
         */
        font-size: var(--effective-die-size);
        perspective: 6em;
        perspective-origin: 50% 50%;
      }

      #inner {
        position: relative;
        transform: translateY(calc(-1 * var(--reel-step) * var(--selected-face)));
        /* The spin is WAAPI-driven now; no CSS transform transition. */
      }

      #inner.solid {
        position: absolute;
        inset: 0;
        transform-style: preserve-3d;
        /*
         * There is no reel to scroll, so the reel step is zero. The #inner
         * rule above and the WAAPI spin keyframes (_innerTransformForFace)
         * both read --reel-step, so zeroing it here makes the face-change spin
         * a no-op on the solid without touching either -- the roll is a real
         * tumble in a later task, and until then the die must not slide.
         * Scoped to --reel-step and NOT to --effective-die-size, which
         * everything under here still needs at its true value.
         */
        --reel-step: 0px;
      }

      #orient {
        position: absolute;
        inset: 0;
        transform-style: preserve-3d;
      }

      .facet {
        position: absolute;
        left: 50%;
        top: 50%;
        /* width/height/margin/transform/clip-path are generated per facet. */
        backface-visibility: hidden;
        background:
          linear-gradient(135deg, #F5F0E8 0%, #E0D9CE 100%);
        box-shadow: inset 0 0 0 1px rgba(60, 40, 20, 0.14);
        transition: background 0.2s ease-out;
      }

      #main.solid.interactive:hover .facet {
        background:
          linear-gradient(135deg, #FFFDF8 0%, #EFE9DE 100%);
      }

      .facet,
      .face {
        font-family: var(--md-sys-typescale-body-large-font, 'Source Sans 3', sans-serif);
        font-weight: 500;
        color: var(--md-sys-color-on-surface, #1C1810);
      }

      /*
       * A reel face is one die-size square, so its content square is just a
       * centred fraction of it. 63% reproduces the flat die's original pip
       * geometry: on a 50px face a dot lands 10.5px off centre, exactly where
       * it used to, and measures 6.3px across where it used to be 7.
       */
      #inner.reel .face {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        position: relative;
        --content-left: 18.5%;
        --content-top: 18.5%;
        --content-width: 63%;
        --content-height: 63%;
        --content-size: calc(var(--effective-die-size) * 0.63);
      }

      /*
       * Every mark a face carries lives inside its CONTENT SQUARE, whose box
       * the facet supplies as four percentages of its own box plus the square's
       * side in em (--content-size) for the marks to size themselves against.
       * The square is the largest that fits inside the facet's polygon, so a
       * mark that fits the square cannot leave the facet -- which is what makes
       * one layout work on a cube's square, a d20's triangle, a d10's kite and
       * a d7's 2.7:1 barrel face alike.
       */
      .content {
        position: absolute;
        left: var(--content-left);
        top: var(--content-top);
        width: var(--content-width);
        height: var(--content-height);
        --pip-size: calc(var(--content-size) * ${PIP_DIAMETER});
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
      }

      /*
       * Text sizes itself in em against the facet's font-size (the die's size)
       * and is set per mark, because it depends on both the square it is in
       * and how many characters it has to fit.
       */
      .content > span,
      .corner > span {
        line-height: 1;
        white-space: nowrap;
      }

      /*
       * The corner marks of a die that is read from a face nobody can see.
       * Position and size are percentages of the FACET's box, never em: the
       * span inside sets a font-size, and an em on the same element would then
       * resolve against that instead of the die's size.
       */
      .corner {
        position: absolute;
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
        opacity: 0.85;
      }

      /*
       * Pips are placed on the 3x3 lattice of the content square: cell centres
       * at a sixth, a half and five sixths of it. The cells a value occupies
       * are computed (see pipCells), not enumerated in CSS -- which is what
       * used to cap the die at six faces.
       */
      .pip {
        background-color: currentColor;
        height: var(--pip-size);
        width: var(--pip-size);
        border-radius: 50%;
        position: absolute;
        box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.4),
                    0 1px 0 rgba(255, 255, 255, 0.2);
      }
    `
  ];

  @property({ type: Object })
  item: any = null;

  @property({ type: Number })
  value = 0;

  @property({ type: Array })
  faces: number[] = [];

  @property({ type: Number })
  selectedFace = 0;

  /**
   * Face VALUE to the name that face carries — `{ 3: 'Star' }`.
   *
   * The seam an enum plugs into. The framework's `enum` package already sends
   * its values to the client, so a later change supplies this from the enum a
   * die's faces are typed with and nothing else here moves: a name selects the
   * glyph out of `symbols`, and it is what the die announces. With no names
   * attached, a face's name is its own value written out, so a symbol set can
   * be keyed by plain integers and still work.
   */
  @property({ type: Object })
  faceNames: Record<string, string> | null = null;

  /**
   * Face NAME to the glyph drawn for it — `{ Star: '★' }`.
   *
   * The author-supplied symbol set, and the first thing face content resolves
   * to: a face with a glyph draws the glyph whatever its value would otherwise
   * have generated.
   */
  @property({ type: Object })
  symbols: Record<string, string> | null = null;

  @property({ type: Boolean })
  disabled = false;

  @property({ attribute: false })
  action: BoundMoveAction<string, object> | null = null;

  @query('#inner')
  private _innerElement?: HTMLElement;

  private _boundHandleClick?: (e: Event) => void;
  private _unsubscribeAction: (() => void) | null = null;

  override connectedCallback() {
    super.connectedCallback();
    this._boundHandleClick ??= (e: Event) => this._handleClick(e);
    this.renderRoot.addEventListener('click', this._boundHandleClick);
    this._subscribeAction();
  }

  override disconnectedCallback() {
    if (this._boundHandleClick) {
      this.renderRoot.removeEventListener('click', this._boundHandleClick);
    }
    this._unsubscribeAction?.();
    this._unsubscribeAction = null;
    super.disconnectedCallback();
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('selectedFace')) {
      this._selectedFaceChanged(
        this.selectedFace,
        changedProperties.get('selectedFace') as number | undefined
      );
    }

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    }

    if (changedProperties.has('action')) {
      this._subscribeAction();
    }
  }

  private _handleClick(e: Event) {
    if (!isBoundMoveAction(this.action)) {
      if (this.action !== null) e.stopPropagation();
      return;
    }
    if (this.disabled || !this.action.canActivate) {
      e.stopPropagation();
      return;
    }
    e.stopPropagation();
    void this.action.activate();
  }

  private _subscribeAction(): void {
    this._unsubscribeAction?.();
    this._unsubscribeAction = isBoundMoveAction(this.action)
      ? this.action.subscribe(() => this.requestUpdate())
      : null;
  }

  // _innerTransformForFace mirrors the CSS resting transform on #inner for
  // a given selectedFace: translateY(-1 * reel-step * face). The #main element
  // carries the --selected-face var that drives the CSS resting position; here
  // we build an explicit transform so WAAPI can interpolate the spin instead
  // of relying on a CSS transition.
  //
  // IN SOLID MODE THIS ANIMATION IS A DELIBERATE VISUAL NO-OP. #inner.solid
  // sets --reel-step to 0px, so both keyframes resolve to translateY(0) and
  // the solid does not move: a solid has no reel to scroll, and letting the
  // reel's translateY through would slide the die by a multiple of its own
  // size and back on every roll. The track is still scheduled because the
  // gate/play/active events it produces are pinned by the pig-roll parity
  // golden -- the die's animation contract is unchanged by drawing it as a
  // solid. The next task replaces this track with the real tumble.
  private _innerTransformForFace(face: number): string {
    return `translateY(calc(-1 * var(--reel-step) * ${face}))`;
  }

  protected override motionTrackTarget(target: ComponentMotionTarget): HTMLElement | null {
    return target === 'host' ? this : this._innerElement ?? null;
  }

  // Schedules the face-change spin. On a solid the spin it schedules moves
  // nothing on screen by design -- see _innerTransformForFace -- but it is
  // scheduled all the same, because the motion-track events are the die's
  // observable animation contract and a parity golden pins them.
  private _selectedFaceChanged(newValue: number, oldValue: number | undefined) {
    if (!this._innerElement) return;
    // On first render there's no meaningful spin to animate from.
    if (oldValue === undefined || oldValue === newValue) return;
    this.playMotionTracks(componentMotionTracks([{
      target: 'visual',
      property: 'transform',
      from: this._innerTransformForFace(oldValue),
      to: this._innerTransformForFace(newValue),
    }]));
  }

  private _itemChanged(newValue: any) {
    if (!newValue) {
      this.faces = [];
      this.selectedFace = 0;
      this.value = 0;
      return;
    }
    this.faces = newValue.Values.Faces;
    this.selectedFace = newValue.DynamicValues.SelectedFace;
    this.value = newValue.DynamicValues.Value;
  }

  private _classes(disabled: boolean, solid: boolean): string {
    const pieces: string[] = [];
    pieces.push(disabled ? 'disabled' : 'interactive');
    pieces.push(solid ? 'solid' : 'reel');
    return pieces.join(' ');
  }

  /**
   * The solid this die should be drawn as, or `null` when it has none: fewer
   * than three faces, or a face list the geometry module refuses. Those fall
   * back to the reel rather than throwing during a render pass.
   */
  private _solid(): DieSolid | null {
    const faces = this.faces;
    if (!Array.isArray(faces) || faces.length < 3) return null;
    if (!faces.every((face) => Number.isFinite(face))) return null;
    return dieSolid(faces.length);
  }

  /**
   * Which FACE the die presents, as an index into `faces`.
   *
   * `selectedFace` is an index (the server sends `DynamicValues.SelectedFace`
   * alongside a separate `Values.Faces` list of face VALUES); reading it as a
   * value is the silent bug this component invites. Out-of-range values fall
   * back to the first face rather than rendering nothing.
   */
  private _presentedFaceIndex(faceCount: number): number {
    const index = Math.trunc(this.selectedFace);
    return Number.isFinite(index) && index >= 0 && index < faceCount ? index : 0;
  }

  /**
   * The name a face carries: the enum's string name once one is attached, and
   * otherwise the value written out. Also the die's accessible label for that
   * face, which is why it is one function and not two.
   */
  private _nameForValue(value: number): string {
    const names = this.faceNames;
    if (names && typeof names === 'object') {
      const name = names[String(value)];
      if (typeof name === 'string' && name.length > 0) return name;
    }
    return String(value);
  }

  /** The author-supplied glyph for a face name, or null when there is none. */
  private _glyphForName(name: string): string | null {
    const symbols = this.symbols;
    if (!symbols || typeof symbols !== 'object') return null;
    const glyph = symbols[name];
    return typeof glyph === 'string' && glyph.length > 0 ? glyph : null;
  }

  /**
   * Whether THIS DIE draws its unlettered faces as pips.
   *
   * A property of the whole die, not of each face: a die that mixed dots on
   * one face with a numeral on the next would read as two different dice, so
   * one value past the lattice's capacity (see `MAX_PIP_VALUE`) moves all of
   * them to numerals. That is what makes a d6 pipped and a d20 numbered
   * without either being named anywhere.
   *
   * Corner-printed dice (a d4, an odd barrel) are always numbered: their value
   * has to fit in a small square at a corner, where dots do not read, and a
   * face carrying pips in the middle and numerals at its corners reads as a
   * mistake.
   */
  private _usesPips(solid: DieSolid | null): boolean {
    if (solid?.cornerPrinted) return false;
    const faces = Array.isArray(this.faces) ? this.faces : [];
    return faces.every((value) =>
      isPipValue(value) || this._glyphForName(this._nameForValue(value)) !== null);
  }

  /**
   * A face's content, resolved in the one order this component documents:
   * author symbol set, then generated pips, then a numeral. `label` is what
   * the die ANNOUNCES for that face, and it always describes what is actually
   * drawn — which is the assertion that catches drawing one face's value while
   * announcing another's.
   */
  private _resolveFace(value: number, usePips: boolean): {
    kind: 'symbol' | 'pips' | 'numeral';
    text: string;
    cells: readonly PipCell[];
    label: string;
  } {
    const name = this._nameForValue(value);
    const glyph = this._glyphForName(name);
    if (glyph !== null) {
      return { kind: 'symbol', text: glyph, cells: [], label: name };
    }
    // A named face with no glyph draws its number and announces both: the
    // number is what is on the facet, the name is what it means.
    const label = name === String(value) ? String(value) : `${name} (${value})`;
    if (usePips && isPipValue(value)) {
      return { kind: 'pips', text: '', cells: pipCells(value), label };
    }
    return { kind: 'numeral', text: String(value), cells: [], label };
  }

  /** The content square of one face: its dots, or its glyph/numeral. */
  private _faceContent(value: number, usePips: boolean) {
    const content = this._resolveFace(value, usePips);
    if (content.kind === 'pips') {
      return html`<div class="content">
        ${content.cells.map(([col, row]) => html`<div
          class="pip"
          style="left:calc(${num(((col + 0.5) / 3) * 100)}% - var(--pip-size) / 2);top:calc(${num(((row + 0.5) / 3) * 100)}% - var(--pip-size) / 2)"
        ></div>`)}
      </div>`;
    }
    return html`<div class="content"><span
      style="font-size:calc(var(--content-size) * ${num(glyphScale(content.text, GLYPH_HEIGHT))})"
    >${content.text}</span></div>`;
  }

  /**
   * The values printed at a facet's corners, for a die read from a face nobody
   * can see. Each mark carries the value that would be READ if that corner
   * were the top of the die — so a d4 resting on face 2 shows a 2 at the apex
   * of the three faces still facing the player. Always a single glyph or
   * numeral, never dots: the square at a corner is a third of the face's.
   */
  private _cornerContent(facet: DieFacet) {
    return facet.corners.map((corner) => {
      const value = this.faces[corner.faceIndex];
      const content = this._resolveFace(value, false);
      return html`<div
        class="corner"
        data-corner-face-index="${corner.faceIndex}"
        style="left:${num(corner.left)}%;top:${num(corner.top)}%;width:${num(corner.width)}%;height:${num(corner.height)}%"
      ><span
        style="font-size:${num(corner.size * glyphScale(content.text, CORNER_GLYPH_HEIGHT))}em"
      >${content.text}</span></div>`;
    });
  }

  // The degenerate fallback: the pre-3D vertical reel of flat faces, scrolled
  // to the selected one by #inner's translateY. It has no geometry, so no
  // corner marks -- and nothing with fewer than three faces is read from a
  // face it rests on anyway.
  private _renderReel() {
    const usePips = this._usesPips(null);
    return html`
      <div id="inner" class="reel">
        ${repeat(this.faces, (face) => face, (face) => html`
          <div class="face" data-face-value="${face}" data-face-label="${this._resolveFace(face, usePips).label}"
            >${this._faceContent(face, usePips)}</div>
        `)}
      </div>
    `;
  }

  private _renderSolid(solid: DieSolid) {
    const presented = this._presentedFaceIndex(solid.geometry.faceCount);
    const orient = presentationTransform(solid.geometry, presented);
    const usePips = this._usesPips(solid);
    return html`
      <div id="stage">
        <div id="inner" class="solid">
          <div id="orient" style="transform:${orient}">
            ${repeat(solid.facets, (facet) => facet.key, (facet) => facet.faceIndex < 0
              ? html`<div class="facet cap" style="${facet.style}"></div>`
              : html`<div
                    class="facet"
                    style="${facet.style}"
                    data-face-index="${facet.faceIndex}"
                    data-face-value="${this.faces[facet.faceIndex]}"
                    data-face-label="${this._resolveFace(this.faces[facet.faceIndex], usePips).label}"
                  >${this._faceContent(this.faces[facet.faceIndex], usePips)}${this._cornerContent(facet)}</div>`)}
          </div>
        </div>
      </div>
    `;
  }

  /**
   * What the die announces: the label of the face it is presenting, so a
   * screen reader is told the same thing the facet draws. Absent entirely for
   * a die with no faces, which is what an unconfigured `<boardgame-die>` is.
   */
  private _ariaLabel(interactive: boolean, solid: DieSolid | null): string {
    const base = interactive ? 'Roll die' : 'Die';
    const faces = Array.isArray(this.faces) ? this.faces : [];
    if (faces.length === 0) return base;
    const presented = this._presentedFaceIndex(faces.length);
    return `${base} showing ${this._resolveFace(faces[presented], this._usesPips(solid)).label}`;
  }

  override render() {
    const action = this.action;
    const bound = isBoundMoveAction(action);
    const interactive = bound;
    const effectiveDisabled = this.disabled || !interactive || (bound && !action.canActivate);
    const baseReason = bound
      ? action.reason?.message
      : action ? 'Bind required move input with .with(...)' : null;
    const reason = bound && action.preview.kind === 'failed' && action.preview.retryable
      ? `${baseReason ?? 'Move legality check failed'}. Activate to retry.`
      : baseReason;
    const solid = this._solid();
    return html`
      <div id="scaler">
        <button
          id="main"
          type="button"
          aria-label=${this._ariaLabel(interactive, solid)}
          aria-describedby=${reason ? 'action-status' : ''}
          aria-busy=${String(bound && action.submission.kind === 'pending')}
          ?disabled=${effectiveDisabled}
          style="--selected-face:${this.selectedFace}"
          class="${this._classes(effectiveDisabled, solid !== null)}">
          ${solid ? this._renderSolid(solid) : this._renderReel()}
        </button>
        ${reason ? html`<span id="action-status" role="status">${reason}</span>` : ''}
      </div>
    `;
  }
}

customElements.define('boardgame-die', BoardgameDie);
