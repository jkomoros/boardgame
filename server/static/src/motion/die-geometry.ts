/**
 * Pure geometry for a die of an arbitrary face count.
 *
 * No DOM, no packages, no randomness: this module is the shape oracle that the
 * physics simulation, the keyframe bake, and the renderer all build on, so it
 * has to be exhaustively unit-testable on its own.
 *
 * Face counts 4, 6, 8, 12 and 20 get their face-transitive Platonic solid and
 * 10 gets the pentagonal trapezohedron of a real d10. Every other count N >= 3
 * gets a generated barrel: N side faces around an axis, capped at both ends by
 * a cone of triangular facets meeting at an apex rather than by a flat face, in
 * proportions that make every cap facet an UNSTABLE rest (see `barrelSolid`), so
 * the die always comes to rest on a readable side face. For a barrel the
 * readable faces are the side faces only, which is why `faceCount ===
 * faces.length` holds for every shape; the cap triangles are still part of the
 * surface and are exposed separately as `capFaces`.
 */

/** A point or direction in 3D. Deliberately a plain tuple: no dependencies. */
export type Vec3 = readonly [number, number, number];

/**
 * A rotation as (x, y, z, w). Defined here so the simulator can share it.
 * Consumed by task 5 (`presentedFaceIndex(geometry, orientation: Quat)`) and by
 * task 6's rigid-body integrator; unused inside this module by design.
 */
export type Quat = readonly [number, number, number, number];

/** See `Quat`: the rest orientation tasks 5 and 6 start from. */
export const QUAT_IDENTITY: Quat = Object.freeze([0, 0, 0, 1] as const);

export function vec3(x: number, y: number, z: number): Vec3 {
  return Object.freeze([x, y, z] as const);
}

export function add(a: Vec3, b: Vec3): Vec3 {
  return vec3(a[0] + b[0], a[1] + b[1], a[2] + b[2]);
}

export function subtract(a: Vec3, b: Vec3): Vec3 {
  return vec3(a[0] - b[0], a[1] - b[1], a[2] - b[2]);
}

export function scale(a: Vec3, factor: number): Vec3 {
  return vec3(a[0] * factor, a[1] * factor, a[2] * factor);
}

export function dot(a: Vec3, b: Vec3): number {
  return a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
}

export function cross(a: Vec3, b: Vec3): Vec3 {
  return vec3(
    a[1] * b[2] - a[2] * b[1],
    a[2] * b[0] - a[0] * b[2],
    a[0] * b[1] - a[1] * b[0],
  );
}

export function magnitude(a: Vec3): number {
  return Math.sqrt(dot(a, a));
}

/** Normalize, or throw: a zero-length direction is always a construction bug. */
export function normalize(a: Vec3): Vec3 {
  const length = magnitude(a);
  if (!(length > 0)) throw new Error('cannot normalize a zero-length vector');
  return scale(a, 1 / length);
}

export interface DieFace {
  readonly normal: Vec3;
  readonly centroid: Vec3;
  /** Ordered counter-clockwise seen from outside the solid. */
  readonly polygon: readonly Vec3[];
}

export interface DieGeometry {
  /** Readable faces. For a barrel this counts side faces; the caps are points. */
  readonly faceCount: number;
  readonly vertices: readonly Vec3[];
  readonly faces: readonly DieFace[];
  /**
   * The rest of the closed surface: the faces that exist but carry no value.
   * Empty for every closed-form solid; for a barrel it is the 2N cap triangles.
   * `[...faces, ...capFaces]` is the complete surface, with no polygon appearing
   * twice, which is what a renderer must draw and what a point-in-solid test
   * must clip against. Kept out of `faces` so `faceCount === faces.length`.
   */
  readonly capFaces: readonly DieFace[];
  /**
   * 9 entries, row-major, about the centroid, for unit mass.
   *
   * NOT size-normalized: this is the tensor of the solid at the scale it is
   * built at, and those scales differ between face counts (see `nominalRadius`).
   * Unit-mass inertia goes as R^2, and the d20 is built 1.89x the size of the
   * d10, so a d20's Izz measures 3.73x a d10's (4.11x on the trace). Consumers
   * must divide lengths by their own normalizing radius and the tensor by its
   * SQUARE, or dice of different face counts will tumble at visibly different
   * rates. `dice-sim.ts`, the only consumer of this tensor, normalizes by
   * `boundingRadius` — not by `nominalRadius`; see both fields.
   */
  readonly inertiaTensor: readonly number[];
  /**
   * HALF THE DIE'S NOMINAL BOX: the radius the RENDERER normalizes by, so that
   * a solid of any face count draws at a common `--die-size`. Divide lengths by
   * it; `facet-placement.ts` does exactly that, at `0.5 / nominalRadius` per em.
   *
   * NOT normalized across face counts: it runs from 1.000 for the d8 to 1.902
   * for the d20, because each solid keeps its natural closed-form coordinates.
   *
   * ## Nominal, which is the circumradius only sometimes
   *
   * For every solid with a closed form this IS the circumradius, and
   * `nominalRadius === boundingRadius`. For a BARREL it is deliberately not: a
   * barrel is 2.1 to 2.6 times longer than it is wide, so normalizing it by its
   * circumsphere spends the whole die box on a diagonal nobody reads and leaves
   * the readable side faces — whose content is bounded by the barrel's WIDTH —
   * at a bit over 0.4 of the box. A d7's numeral measured 4.3px on a 50px die
   * and could not be read from a screenshot at all. So a barrel is normalized
   * by its SHORT axis instead (see `barrelSolid`): the die box is its width,
   * its length overflows the box, and every mark on it roughly doubles.
   *
   * `boundingRadius` is the honest circumsphere and is what the PHYSICS
   * normalizes by, so the two are not interchangeable — see `dice-sim.ts`. The
   * field is named for what it means rather than for what it usually equals
   * precisely because a `circumradius` that is not a circumradius is the kind
   * of lie every unit test in this pipeline would pass over in silence.
   */
  readonly nominalRadius: number;
  /**
   * Distance from the centroid to the farthest vertex: the radius of the
   * smallest bounding SPHERE, always, for every shape. This one really is the
   * circumradius, for every solid, including a barrel.
   *
   * Equal to `nominalRadius` except on a barrel, where it is 1.37x (d3) to
   * 2.63x (large N) larger. `dice-sim.ts` normalizes by this rather than by
   * `nominalRadius`, because the simulator's tray is measured in die radii and
   * a die whose bounding sphere overflowed the tray would spend the throw
   * embedded in a wall.
   */
  readonly boundingRadius: number;
  /**
   * Which face of a resting die carries its value.
   *
   * `'up-face'` for anything that presents a single face upward. `'top-vertex'`
   * for the d4: it rests ON a face, three faces tilt equally upward, and a real
   * d4 is read from the apex. `'down-face'` for an ODD-SIDED BARREL: resting on
   * a side face it points an EDGE at the ceiling, so its two best up-face
   * candidates are a floating-point tie (1.1e-16 apart for a d7) while the face
   * it rests on wins by more than 1e-3 — which is also how physical odd barrel
   * dice are read.
   */
  readonly readingRule: ReadingRule;
}

/** See `DieGeometry.readingRule`. */
export type ReadingRule = 'up-face' | 'down-face' | 'top-vertex';

/** Coplanarity/convexity slack. Constructions here land ~1e-15 from exact. */
const EPSILON = 1e-9;

/**
 * A solid before it has been centred and measured: a vertex list, the complete
 * closed surface as index loops, and which of those loops are readable faces.
 *
 * Surface loops may be given in any vertex order and any winding: `finishSolid`
 * owns the winding invariant and orients every loop exactly once.
 */
export interface RawSolid {
  readonly vertices: readonly Vec3[];
  readonly surface: readonly (readonly number[])[];
  readonly readable: readonly number[];
  readonly readingRule: ReadingRule;
  /**
   * What `DieGeometry.nominalRadius` should be, in THESE coordinates, for a
   * solid that does not want to be normalized by its circumsphere. Omitted —
   * the usual case — the circumsphere is used and `nominalRadius` is the
   * circumradius.
   *
   * Only `barrelSolid` supplies it, and it supplies the barrel's short
   * semi-axis. Stated about the raw coordinates rather than the centred ones
   * because a solid that needs this knows its own construction; `finishSolid`
   * centres on the volume centroid, so a raw solid that is not already centred
   * on it must say so with that shift already taken out.
   */
  readonly nominalRadius?: number;
}

/** Newell's method: robust for polygons with more than three vertices. */
function polygonNormal(polygon: readonly Vec3[]): Vec3 {
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
  return normalize(vec3(x, y, z));
}

/**
 * The plane normal of a coplanar point set in any order, from the triple that
 * spans the most area. Newell's method needs the points already sorted around
 * the polygon, which is exactly what we do not have yet.
 */
function planeNormal(points: readonly Vec3[]): Vec3 {
  let best = vec3(0, 0, 0);
  let bestArea = 0;
  for (let i = 1; i < points.length; i++) {
    for (let j = i + 1; j < points.length; j++) {
      const candidate = cross(
        subtract(points[i], points[0]),
        subtract(points[j], points[0]),
      );
      const area = magnitude(candidate);
      if (area > bestArea) {
        bestArea = area;
        best = candidate;
      }
    }
  }
  if (bestArea < EPSILON) throw new Error('degenerate face: collinear vertices');
  return normalize(best);
}

function meanPoint(points: readonly Vec3[]): Vec3 {
  let x = 0;
  let y = 0;
  let z = 0;
  for (const point of points) {
    x += point[0];
    y += point[1];
    z += point[2];
  }
  return vec3(x / points.length, y / points.length, z / points.length);
}

/**
 * Order a coplanar index loop counter-clockwise about `normal`, then flip the
 * winding if it does not face away from `interior`.
 */
function orientLoop(
  loop: readonly number[],
  vertices: readonly Vec3[],
  interior: Vec3,
): number[] {
  const points = loop.map((index) => vertices[index]);
  const centroid = meanPoint(points);
  const normal = planeNormal(points);
  const outward = dot(normal, subtract(centroid, interior)) > 0 ? normal : scale(normal, -1);
  const radial = normalize(subtract(points[0], centroid));
  const binormal = cross(outward, radial);
  const ordered = [...loop].sort((a, b) => {
    const da = subtract(vertices[a], centroid);
    const db = subtract(vertices[b], centroid);
    return (
      Math.atan2(dot(da, binormal), dot(da, radial)) -
      Math.atan2(dot(db, binormal), dot(db, radial))
    );
  });
  const check = polygonNormal(ordered.map((index) => vertices[index]));
  if (dot(check, outward) < 0) ordered.reverse();
  return ordered;
}

/**
 * Every supporting face of the convex hull of `vertices`, as index loops.
 *
 * Brute force over vertex triples: a plane is a face plane when no vertex lies
 * strictly outside it. Only used for the small closed-form solids (at most 20
 * vertices), so the cubic triple enumeration is irrelevant; barrels are built
 * directly instead.
 *
 * Loops come back in vertex-index order, which is not a polygon order. Winding
 * is `finishSolid`'s job, so this deliberately does not orient them.
 */
function convexHullFaces(vertices: readonly Vec3[]): number[][] {
  const found = new Map<string, number[]>();
  for (let i = 0; i < vertices.length; i++) {
    for (let j = i + 1; j < vertices.length; j++) {
      for (let k = j + 1; k < vertices.length; k++) {
        const raw = cross(
          subtract(vertices[j], vertices[i]),
          subtract(vertices[k], vertices[i]),
        );
        if (magnitude(raw) < EPSILON) continue;
        const normal = normalize(raw);
        const offset = dot(normal, vertices[i]);
        let above = 0;
        let below = 0;
        for (const vertex of vertices) {
          const distance = dot(normal, vertex) - offset;
          if (distance > EPSILON) above++;
          if (distance < -EPSILON) below++;
        }
        if (above > 0 && below > 0) continue;
        const loop: number[] = [];
        for (let n = 0; n < vertices.length; n++) {
          if (Math.abs(dot(normal, vertices[n]) - offset) <= EPSILON) loop.push(n);
        }
        const key = [...loop].sort((a, b) => a - b).join(',');
        if (found.has(key)) continue;
        found.set(key, loop);
      }
    }
  }
  return [...found.values()];
}

const PHI = (1 + Math.sqrt(5)) / 2;

/** The five Platonic dice, given by their vertices; faces come from the hull. */
function platonicVertices(faceCount: number): readonly Vec3[] {
  switch (faceCount) {
    case 4:
      return [vec3(1, 1, 1), vec3(1, -1, -1), vec3(-1, 1, -1), vec3(-1, -1, 1)];
    case 6: {
      const vertices: Vec3[] = [];
      for (const x of [-1, 1]) {
        for (const y of [-1, 1]) {
          for (const z of [-1, 1]) vertices.push(vec3(x, y, z));
        }
      }
      return vertices;
    }
    case 8:
      return [
        vec3(1, 0, 0), vec3(-1, 0, 0),
        vec3(0, 1, 0), vec3(0, -1, 0),
        vec3(0, 0, 1), vec3(0, 0, -1),
      ];
    case 12: {
      const vertices: Vec3[] = [];
      for (const x of [-1, 1]) {
        for (const y of [-1, 1]) {
          for (const z of [-1, 1]) vertices.push(vec3(x, y, z));
        }
      }
      for (const a of [-1, 1]) {
        for (const b of [-1, 1]) {
          vertices.push(vec3(0, (a * 1) / PHI, b * PHI));
          vertices.push(vec3((a * 1) / PHI, b * PHI, 0));
          vertices.push(vec3(a * PHI, 0, (b * 1) / PHI));
        }
      }
      return vertices;
    }
    case 20: {
      const vertices: Vec3[] = [];
      for (const a of [-1, 1]) {
        for (const b of [-1, 1]) {
          vertices.push(vec3(0, a, b * PHI));
          vertices.push(vec3(a, b * PHI, 0));
          vertices.push(vec3(a * PHI, 0, b));
        }
      }
      return vertices;
    }
    default:
      throw new Error(`no Platonic solid has ${faceCount} faces`);
  }
}

/**
 * The pentagonal trapezohedron of a real d10: two apexes and two offset rings
 * of five. The kites are planar only when the apex height satisfies
 * `apex = ring * (5 + 2*sqrt(5))`, which is what fixes the ring height here.
 */
function trapezohedronVertices(): readonly Vec3[] {
  const apex = 1;
  const ringHeight = apex / (5 + 2 * Math.sqrt(5));
  const vertices: Vec3[] = [];
  for (let k = 0; k < 5; k++) {
    const upper = (2 * Math.PI * k) / 5;
    const lower = upper + Math.PI / 5;
    vertices.push(vec3(Math.cos(upper), Math.sin(upper), ringHeight));
    vertices.push(vec3(Math.cos(lower), Math.sin(lower), -ringHeight));
  }
  vertices.push(vec3(0, 0, apex));
  vertices.push(vec3(0, 0, -apex));
  return vertices;
}

/**
 * Barrel proportions: why a pointed cap is not enough, and what fixes it.
 *
 * A barrel is a ring of radius 1, a band of half-height `h`, and two cones of
 * height `c` rising to an apex, so the die is `2(h + c)` long and 2 wide. The
 * intent is that it can only come to rest on a readable SIDE face. A pointed
 * apex does NOT deliver that: each cap is 2N flat triangular facets, and a
 * facet is a stable rest exactly when the centre of mass — the origin, by
 * symmetry — projects inside it.
 *
 * Write a = h + c and alpha = pi/N. One cap facet has its apex at (0, 0, a) and
 * a base edge of the ring at height h; by symmetry the centre of mass projects
 * onto that facet's median, a fraction
 *
 *     t = a * c / (c^2 + cos^2 alpha)
 *
 * of the way from the apex to the base edge, and the facet's support margin —
 * the signed distance from that projection to the base edge, positive inside —
 * works out to
 *
 *     margin = (cos^2 alpha - h * c) / sqrt(c^2 + cos^2 alpha).
 *
 * So the whole question is one product:
 *
 *     h * c > cos^2(pi/N)   <=>   the cap facets are UNSTABLE.        (*)
 *
 * The original barrel used h = sin(pi/N) (which made the side faces square)
 * with a flat c = 0.5, which satisfies (*) only for N = 3: a d7 rested on a cap
 * facet in 64% of unretried rolls, 61 degrees off a readable face. `dice-sim.ts`
 * found that; no structural check about the apex VERTEX can, because the apex
 * is a legitimately unique extreme point the whole time.
 *
 * `BARREL_CAP_SAFETY` is how many times the threshold this module puts h*c at —
 * `h * c = BARREL_CAP_SAFETY * cos^2(pi/N)` for every barrel, so every barrel
 * clears (*) by the same factor. sqrt(3) is not a tuned number: it is
 * exactly the ratio the shipped d3 — the one barrel whose caps were already
 * unstable — already had (0.5 * sin 60 / cos^2 60 = sqrt(3)), so every other
 * face count simply inherits the margin of the case that was right. It leaves
 * every cap facet unstable by 0.168 (large N) to 0.189 (the d3) of a BOUNDING
 * radius — the circumsphere, which is the unit the physics normalizes by and
 * the unit every margin in this doc is quoted in — against the +0.161 by which
 * a d7's cap facets were STABLE before.
 *
 * Splitting that product between band and cap is what decides the die's SHAPE,
 * and it is why the square-side-face rule had to go. h + c is minimised at
 * h = c = sqrt(BARREL_CAP_SAFETY) * cos alpha; the square rule instead drives
 * h to 0 as N grows, so c ~ N/pi and the die becomes a needle (a d24 eleven
 * times longer than wide, a d100 forty-eight times) whose margins collapse to
 * 0.0002 of a bounding radius — unstable on paper, balanceable in practice. So
 * the band keeps the square height only while that is at least the
 * length-minimising height, which is true for N = 3 alone, and is stretched to
 * the minimising height otherwise. Every barrel is then between 1.37 and 2.64
 * times longer than it is wide, the d3 is unchanged TO WITHIN A ULP, and the
 * side faces are rectangles rather than squares for N >= 5.
 *
 * "To within a ulp" and not "bit for bit": `barrelCapHeight(3)` now comes out
 * of `sqrt(3) * cos^2(60) / sin(60)` rather than the literal `0.5`, and that
 * evaluates to 0.5000000000000002 — one ulp high. The bounding radius shifts
 * in its last bit (the nominal radius is the ring radius and is exactly 1
 * whatever the caps do), and because a tumbling die is chaotic that ulp grows:
 * measured against a d3 pinned to exactly 0.5, poses across 20 two-die rolls
 * diverge by up to 2.3e-5 of a bounding radius, while all 20 durations and
 * every presented face stay identical. Still physically nothing, but it is not
 * zero, and the distinction matters because the argument that sqrt(3) is
 * INHERITED from the d3 rather than tuned to it rests on the d3 coming out the
 * same — an argument only as good as its statement of what "the same" means.
 *
 * ## Signed off, not inherited (Task 11)
 *
 * sqrt(3) leaves a visible discontinuity in the family — the d3 is 1.37 long
 * and everything from the d5 up is 2.13 to 2.63 — because the factor is the
 * d3's own and the d3 is the one barrel the length-minimising split does not
 * bind for. A factor of 1.2 removes most of that (d3 1.21, d5 1.77, d100 2.19)
 * and it was measured, rendered and REJECTED, for three reasons:
 *
 *   1. It buys nothing legible. A side face's glyphs are sized by the largest
 *      square inscribed in that face, and for N >= 5 that square is bounded by
 *      the face's WIDTH (the ring chord, 2 sin(pi/N)), which the aspect ratio
 *      does not touch. All 1.2 changes is the barrel's LENGTH, and since a
 *      barrel is normalized by its short axis (`nominalRadius`, the ring
 *      radius, which 1.2 leaves at exactly 1) that length is not the die box:
 *      at a fixed `--die-size` a shorter barrel's marks come out the same size,
 *      not larger. The 17% this doc once claimed was measured back when a
 *      barrel was normalized by its circumsphere, so shortening it shrank the
 *      normalizing radius; that is no longer how a barrel is sized. Rendered
 *      side by side at 160px, a d7's and a d16's numerals are the same size to
 *      the eye in both; the die is simply stubbier.
 *   2. It costs most of the safety margin the factor exists for: cap facets go
 *      from unstable by 0.168 of a bounding radius to unstable by 0.062, which
 *      is below the 0.1 bound `never has a stable cap facet, from the d3 to the
 *      d100` asserts — a bound that is itself backed by measurement (the
 *      pre-fix geometry landed a d16 unreadable 171 times in 240). Adopting
 *      1.2 means loosening a safety test to fit, which is the wrong direction
 *      for a shape that only matters when the physics gets it wrong.
 *   3. It changes the d3 (1.37 -> 1.21), and the d3 being unchanged is the
 *      whole argument that this factor is inherited rather than tuned.
 *
 * The discontinuity is therefore deliberate: the d3 is short because a
 * triangular prism does not need to be long to be unstable on its caps, and
 * every other barrel is as short as the tipping threshold lets it be.
 */
const BARREL_CAP_SAFETY = Math.sqrt(3);

/** Half the axial length of a barrel's side band. See `BARREL_CAP_SAFETY`. */
function barrelHalfHeight(sideCount: number): number {
  const alpha = Math.PI / sideCount;
  return Math.max(Math.sin(alpha), Math.sqrt(BARREL_CAP_SAFETY) * Math.cos(alpha));
}

/**
 * How far each pointed cap rises above the side band: enough that
 * `halfHeight * capHeight` clears the tipping threshold `cos^2(pi/N)` by
 * `BARREL_CAP_SAFETY`, which is what keeps the die off its caps.
 */
function barrelCapHeight(sideCount: number): number {
  const alpha = Math.PI / sideCount;
  return (BARREL_CAP_SAFETY * Math.cos(alpha) ** 2) / barrelHalfHeight(sideCount);
}

/**
 * A barrel with `sideCount` readable side faces and two pointed caps. The caps
 * are fans of triangles meeting at an apex vertex, never a flat face, and are
 * steep enough (see `BARREL_CAP_SAFETY`) that no facet of them is a stable
 * rest, so every resting pose presents a readable side face.
 *
 * An odd-sided barrel resting on a side face points an EDGE at the ceiling, so
 * it has no up face to read and is read from the face it rests on instead.
 *
 * ## Why a barrel is normalized by its SHORT axis
 *
 * `nominalRadius: 1` — the ring radius — rather than the circumsphere, and it
 * is the single most legible thing about these dice. The die box is `2 *
 * nominalRadius` across, so normalizing by the circumsphere spends the whole
 * box on the barrel's LENGTH, which nothing is printed along, and leaves its
 * width at `1 / aspect` of the box: 0.42 for a d7, 0.40 for a d9, 0.39 for a d16.
 * Every mark shrinks with it, because a side face's content is the largest
 * square inscribed in it and that square is bounded by the face's width (the
 * ring chord, `2 sin(pi/N)`) for every N >= 5. Measured on a 50px die before
 * this change: a d7's numeral 4.3px and its corner marks 3.2px, a d9's 3.3 and
 * 2.4, a d16's 1.8 and 1.3 — against 5.1px for a d20 and 8.4 for a d12. A d7
 * could not be read from a screenshot at all across eight rolls.
 *
 * Normalized by the short axis instead, the barrel's WIDTH is the die box, its
 * length overflows it by the aspect ratio (1.37x for the d3, 2.13x to 2.63x
 * from the d5 up, and never more than 2.632x however many sides — see
 * `BARREL_CAP_SAFETY`), and every mark grows by that same ratio. The same d7
 * numeral is 10.3px on a 50px die and its corner marks 7.7px, i.e. ahead of the
 * d20 rather than half of it.
 *
 * What the overflow costs, and who pays it: the solid is drawn outside its `1em`
 * box, so the RENDERER has to ask a layout for more room than `--die-size` —
 * `boardgame-die.ts`'s `solidExtent` turns this ratio into the box `#scaler`
 * reserves, and `die-shape.spec.ts` pins that the drawn solid stays inside it
 * for every shape. It is not paid by the layout: a d7 that overlapped its
 * neighbour by 78px, which is what happened before that box existed, is not a
 * trade anybody agreed to. What IS the trade is that a barrel asks for a box
 * 1.37x to 2.63x wider than a cube at the same `--die-size`; a die that takes
 * more room and can be read is worth more than one that fits and cannot.
 * Nothing downstream clips it (`boardgame-die.ts` keeps `overflow: visible` all
 * the way down, because any clip would collapse the 3D context anyway).
 *
 * What it does NOT change is the physics: `dice-sim.ts` normalizes by
 * `boundingRadius`, so the simulated barrel is the same size in the same tray
 * it always was and every throw is unchanged bit for bit.
 */
function barrelSolid(sideCount: number): RawSolid {
  const halfHeight = barrelHalfHeight(sideCount);
  const apexHeight = halfHeight + barrelCapHeight(sideCount);
  const vertices: Vec3[] = [];
  for (let k = 0; k < sideCount; k++) {
    const angle = (2 * Math.PI * k) / sideCount;
    vertices.push(vec3(Math.cos(angle), Math.sin(angle), halfHeight));
  }
  for (let k = 0; k < sideCount; k++) {
    const angle = (2 * Math.PI * k) / sideCount;
    vertices.push(vec3(Math.cos(angle), Math.sin(angle), -halfHeight));
  }
  const topApex = vertices.push(vec3(0, 0, apexHeight)) - 1;
  const bottomApex = vertices.push(vec3(0, 0, -apexHeight)) - 1;

  const surface: number[][] = [];
  const readable: number[] = [];
  for (let k = 0; k < sideCount; k++) {
    const next = (k + 1) % sideCount;
    readable.push(surface.length);
    surface.push([k, next, sideCount + next, sideCount + k]);
  }
  for (let k = 0; k < sideCount; k++) {
    const next = (k + 1) % sideCount;
    surface.push([topApex, k, next]);
    surface.push([bottomApex, sideCount + k, sideCount + next]);
  }
  return {
    vertices,
    surface,
    readable,
    readingRule: sideCount % 2 === 0 ? 'up-face' : 'down-face',
    // THE SHORT AXIS, which is the ring radius: see `DieGeometry.nominalRadius`
    // and the note above on what it buys.
    nominalRadius: 1,
  };
}

/**
 * Inertia tensor about the origin at unit total mass, by decomposing the closed
 * surface into tetrahedra with their fourth vertex at the origin and summing
 * their second moments at uniform density. One routine for every shape: no
 * per-solid closed forms, because barrels have none.
 */
function inertiaAboutOrigin(triangles: readonly (readonly Vec3[])[]): number[] {
  const covariance = new Array<number>(9).fill(0);
  let volume = 0;
  for (const [a, b, c] of triangles) {
    // Signed volume of the tetrahedron (origin, a, b, c); positive when the
    // triangle winds counter-clockwise seen from outside.
    const tetVolume = dot(a, cross(b, c)) / 6;
    volume += tetVolume;
    const sum = add(add(a, b), c);
    for (let i = 0; i < 3; i++) {
      for (let j = 0; j < 3; j++) {
        // E[x_i x_j] over a tetrahedron with one vertex at the origin.
        const moment = sum[i] * sum[j] + a[i] * a[j] + b[i] * b[j] + c[i] * c[j];
        covariance[i * 3 + j] += (tetVolume * moment) / 20;
      }
    }
  }
  if (!(volume > 0)) throw new Error('degenerate solid: non-positive volume');
  for (let n = 0; n < 9; n++) covariance[n] /= volume;
  const trace = covariance[0] + covariance[4] + covariance[8];
  return covariance.map((value, index) => (index % 4 === 0 ? trace : 0) - value);
}

/** Fan-triangulate a convex loop, preserving its winding. */
function triangulate(polygon: readonly Vec3[]): Vec3[][] {
  const triangles: Vec3[][] = [];
  for (let i = 1; i + 1 < polygon.length; i++) {
    triangles.push([polygon[0], polygon[i], polygon[i + 1]]);
  }
  return triangles;
}

/**
 * Reject a `RawSolid` whose index lists are not usable, before any arithmetic
 * touches them.
 *
 * Out-of-range indices would otherwise surface as a `TypeError` from inside a
 * vector helper, and a duplicated `readable` index would quietly break the
 * module's headline invariant `faceCount === faces.length` from outside.
 */
function validateIndices(raw: RawSolid): void {
  if (raw.surface.length < 4) {
    throw new Error(`degenerate solid: ${raw.surface.length} surface loops cannot enclose a volume`);
  }
  for (const [face, loop] of raw.surface.entries()) {
    if (loop.length < 3) {
      throw new Error(`surface loop ${face} has ${loop.length} vertices: a face needs at least 3`);
    }
    const seen = new Set<number>();
    for (const index of loop) {
      if (!Number.isInteger(index) || index < 0 || index >= raw.vertices.length) {
        throw new Error(
          `surface loop ${face} references vertex ${index}, which is not in the 0..${raw.vertices.length - 1} vertex list`,
        );
      }
      if (seen.has(index)) {
        throw new Error(`surface loop ${face} uses vertex ${index} twice`);
      }
      seen.add(index);
    }
  }
  const seenReadable = new Set<number>();
  for (const index of raw.readable) {
    if (!Number.isInteger(index) || index < 0 || index >= raw.surface.length) {
      throw new Error(
        `readable face ${index} is not one of the ${raw.surface.length} surface loops`,
      );
    }
    if (seenReadable.has(index)) {
      throw new Error(`readable face ${index} appears twice: faces would be counted twice`);
    }
    seenReadable.add(index);
  }
}

/**
 * Reject a surface that is not a closed, consistently oriented manifold.
 *
 * The half-edge check: walking every loop in its own winding must use each
 * directed edge exactly once and must supply the reverse of every one of them.
 * A hole (say a pyramid handed over without its base loop) leaves the boundary
 * edges without reverses; a duplicated or inconsistently wound face uses one
 * directed edge twice.
 *
 * This is what the `volume > 0` guard below cannot do. An open pyramid still
 * integrates to a positive volume — the divergence theorem happily closes the
 * hole through the interior point — and returns an inertia tensor 4% wrong with
 * no signal at all, which then feeds the physics.
 */
function validateClosedSurface(loops: readonly (readonly number[])[]): void {
  const directed = new Set<string>();
  for (const loop of loops) {
    for (let i = 0; i < loop.length; i++) {
      const key = `${loop[i]}->${loop[(i + 1) % loop.length]}`;
      if (directed.has(key)) {
        throw new Error(`surface is not a manifold: directed edge ${key} is used twice`);
      }
      directed.add(key);
    }
  }
  for (const key of directed) {
    const [from, to] = key.split('->');
    if (!directed.has(`${to}->${from}`)) {
      throw new Error(`surface is not closed: edge ${key} has no face on its other side`);
    }
  }
}

/**
 * Orient, centre and measure a raw solid.
 *
 * Exported as a testing seam: every face count this module supports produces a
 * centrally symmetric solid, so through `dieGeometry` alone the vertex mean and
 * the volume centroid always coincide and the centring below would be untested.
 * Tests build a deliberately asymmetric solid (where they differ) and hand it
 * here. Not part of the public geometry API.
 *
 * Structurally invalid input is rejected rather than measured: see
 * `validateIndices` and `validateClosedSurface`.
 */
export function finishSolid(raw: RawSolid): DieGeometry {
  validateIndices(raw);
  const interior = meanPoint(raw.vertices);
  const oriented = raw.surface.map((loop) => orientLoop(loop, raw.vertices, interior));
  validateClosedSurface(oriented);

  // Volume-weighted centroid, from tetrahedra hung off the interior point.
  let volume = 0;
  let weighted = vec3(0, 0, 0);
  for (const loop of oriented) {
    for (const triangle of triangulate(loop.map((index) => raw.vertices[index]))) {
      const [a, b, c] = triangle.map((point) => subtract(point, interior));
      const tetVolume = dot(a, cross(b, c)) / 6;
      volume += tetVolume;
      // The tetrahedron's centroid is the mean of its four vertices, one of
      // which is the interior point itself, i.e. the origin of a, b and c.
      weighted = add(weighted, scale(add(add(a, b), c), tetVolume / 4));
    }
  }
  if (!(volume > 0)) throw new Error('degenerate solid: non-positive volume');
  const centroid = add(interior, scale(weighted, 1 / volume));

  const vertices = raw.vertices.map((vertex) => subtract(vertex, centroid));
  const triangles: Vec3[][] = [];
  for (const loop of oriented) {
    triangles.push(...triangulate(loop.map((index) => vertices[index])));
  }

  const faceAt = (index: number): DieFace => {
    const polygon = oriented[index].map((n) => vertices[n]);
    return Object.freeze({
      normal: polygonNormal(polygon),
      centroid: meanPoint(polygon),
      polygon: Object.freeze(polygon),
    });
  };
  const readable = new Set(raw.readable);
  const faces = raw.readable.map(faceAt);
  // Everything else on the surface: a barrel's cap triangles. Same shape as a
  // readable face so a renderer can draw `[...faces, ...capFaces]` uniformly.
  const capFaces = oriented
    .map((_, index) => index)
    .filter((index) => !readable.has(index))
    .map(faceAt);

  const boundingRadius = Math.max(...vertices.map((vertex) => magnitude(vertex)));
  const nominalRadius = raw.nominalRadius ?? boundingRadius;
  if (!(nominalRadius > 0) || nominalRadius > boundingRadius + EPSILON) {
    throw new Error(
      `nominalRadius must be positive and no larger than the bounding radius ${boundingRadius}, got ${nominalRadius}`,
    );
  }

  return Object.freeze({
    faceCount: faces.length,
    vertices: Object.freeze(vertices),
    faces: Object.freeze(faces),
    capFaces: Object.freeze(capFaces),
    inertiaTensor: Object.freeze(inertiaAboutOrigin(triangles)),
    nominalRadius,
    boundingRadius,
    readingRule: raw.readingRule,
  });
}

function hullSolid(vertices: readonly Vec3[], readingRule: ReadingRule): RawSolid {
  const surface = convexHullFaces(vertices);
  return {
    vertices,
    surface,
    readable: surface.map((_, index) => index),
    readingRule,
  };
}

const CLOSED_FORM_FACE_COUNTS = new Set([4, 6, 8, 10, 12, 20]);

/** Whether `faceCount` has a real-world solid rather than a generated barrel. */
export function hasClosedFormSolid(faceCount: number): boolean {
  return CLOSED_FORM_FACE_COUNTS.has(faceCount);
}

/** Build the polyhedron a die with `faceCount` readable faces should be. */
export function dieGeometry(faceCount: number): DieGeometry {
  if (!Number.isInteger(faceCount) || faceCount < 3) {
    throw new Error(`unsupported die face count: ${faceCount} (need an integer face count >= 3)`);
  }
  if (faceCount === 10) return finishSolid(hullSolid(trapezohedronVertices(), 'up-face'));
  if (hasClosedFormSolid(faceCount)) {
    // A d4 rests on a face, so there is no up-face: it is read from the apex.
    const readingRule = faceCount === 4 ? 'top-vertex' : 'up-face';
    return finishSolid(hullSolid(platonicVertices(faceCount), readingRule));
  }
  return finishSolid(barrelSolid(faceCount));
}
