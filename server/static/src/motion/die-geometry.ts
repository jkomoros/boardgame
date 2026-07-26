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
 * a pointed apex vertex rather than a flat face, so the die cannot come to rest
 * on an end. For a barrel the readable faces are the side faces only, which is
 * why `faceCount === faces.length` holds for every shape.
 */

/** A point or direction in 3D. Deliberately a plain tuple: no dependencies. */
export type Vec3 = readonly [number, number, number];

/** A rotation as (x, y, z, w). Defined here so the simulator can share it. */
export type Quat = readonly [number, number, number, number];

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
  /** 9 entries, row-major, about the centroid, for unit mass. */
  readonly inertiaTensor: readonly number[];
  readonly circumradius: number;
  /** A d4 rests on a face and is read from the apex pointing up. */
  readonly readingRule: 'up-face' | 'top-vertex';
}

/** Coplanarity/convexity slack. Constructions here land ~1e-15 from exact. */
const EPSILON = 1e-9;

/**
 * A solid before it has been centred and measured: a vertex list, the complete
 * closed surface as index loops, and which of those loops are readable faces.
 */
interface RawSolid {
  readonly vertices: readonly Vec3[];
  readonly surface: readonly (readonly number[])[];
  readonly readable: readonly number[];
  readonly readingRule: 'up-face' | 'top-vertex';
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
 */
function convexHullFaces(vertices: readonly Vec3[]): number[][] {
  const interior = meanPoint(vertices);
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
        found.set(key, orientLoop(loop, vertices, interior));
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

/** Half the axial length of a barrel's side band; makes side faces square. */
function barrelHalfHeight(sideCount: number): number {
  return Math.sin(Math.PI / sideCount);
}

/** How far each pointed cap rises above the side band. */
const BARREL_CAP_HEIGHT = 0.5;

/**
 * A barrel with `sideCount` readable side faces and two pointed caps. The caps
 * are fans of triangles meeting at an apex vertex, never a flat face, so the
 * die cannot come to rest on an end and every readable face is a side face.
 */
function barrelSolid(sideCount: number): RawSolid {
  const halfHeight = barrelHalfHeight(sideCount);
  const apexHeight = halfHeight + BARREL_CAP_HEIGHT;
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
  return { vertices, surface, readable, readingRule: 'up-face' };
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

function finishSolid(raw: RawSolid): DieGeometry {
  const interior = meanPoint(raw.vertices);
  const oriented = raw.surface.map((loop) => orientLoop(loop, raw.vertices, interior));

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

  const faces = raw.readable.map((index) => {
    const polygon = oriented[index].map((n) => vertices[n]);
    return Object.freeze({
      normal: polygonNormal(polygon),
      centroid: meanPoint(polygon),
      polygon: Object.freeze(polygon),
    });
  });

  return Object.freeze({
    faceCount: faces.length,
    vertices: Object.freeze(vertices),
    faces: Object.freeze(faces),
    inertiaTensor: Object.freeze(inertiaAboutOrigin(triangles)),
    circumradius: Math.max(...vertices.map((vertex) => magnitude(vertex))),
    readingRule: raw.readingRule,
  });
}

function hullSolid(
  vertices: readonly Vec3[],
  readingRule: 'up-face' | 'top-vertex',
): RawSolid {
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
