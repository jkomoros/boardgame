import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  dieGeometry,
  finishSolid,
  type DieFace,
  type DieGeometry,
  type Quat,
  type RawSolid,
  type Vec3,
} from './die-geometry.ts';
// Test-only dependencies, both one layer above this module. A barrel's caps
// were once proved unrestable by a structural argument about the apex VERTEX,
// which is true and does not imply the thing that matters; only rolling the
// die shows whether it lands readable. See `describe('barrel resting poses')`.
import { presentedFaceIndex, resolveReadingRule, WORLD_UP } from './die-faces.ts';
import { simulateRoll } from './dice-sim.ts';

const sub = (a: Vec3, b: Vec3): Vec3 => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const dot = (a: Vec3, b: Vec3): number => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const mag = (a: Vec3): number => Math.sqrt(dot(a, a));
const cross = (a: Vec3, b: Vec3): Vec3 => [
  a[1] * b[2] - a[2] * b[1],
  a[2] * b[0] - a[0] * b[2],
  a[0] * b[1] - a[1] * b[0],
];

/** Rotate `v` by quaternion `q` through the rotation matrix, not the shorthand. */
function rotate(q: Quat, v: Vec3): Vec3 {
  const length = Math.sqrt(q[0] * q[0] + q[1] * q[1] + q[2] * q[2] + q[3] * q[3]);
  const [x, y, z, w] = [q[0] / length, q[1] / length, q[2] / length, q[3] / length];
  const m = [
    1 - 2 * (y * y + z * z), 2 * (x * y - z * w), 2 * (x * z + y * w),
    2 * (x * y + z * w), 1 - 2 * (x * x + z * z), 2 * (y * z - x * w),
    2 * (x * z - y * w), 2 * (y * z + x * w), 1 - 2 * (x * x + y * y),
  ];
  return [
    m[0] * v[0] + m[1] * v[1] + m[2] * v[2],
    m[3] * v[0] + m[4] * v[1] + m[5] * v[2],
    m[6] * v[0] + m[7] * v[1] + m[8] * v[2],
  ];
}

/** Tolerances named by the brief. */
const PLANARITY_TOLERANCE = 1e-9;
const INERTIA_TOLERANCE = 1e-6;

function close(actual: number, expected: number, tolerance: number, message: string): void {
  assert.ok(
    Math.abs(actual - expected) <= tolerance,
    `${message}: expected ${expected}, got ${actual} (delta ${Math.abs(actual - expected)} > ${tolerance})`,
  );
}

/** Sylvester's criterion: all leading principal minors of a symmetric matrix are positive. */
function isPositiveDefinite(m: readonly number[]): boolean {
  const minor1 = m[0];
  const minor2 = m[0] * m[4] - m[1] * m[3];
  const minor3 =
    m[0] * (m[4] * m[8] - m[5] * m[7]) -
    m[1] * (m[3] * m[8] - m[5] * m[6]) +
    m[2] * (m[3] * m[7] - m[4] * m[6]);
  return minor1 > 0 && minor2 > 0 && minor3 > 0;
}

interface Plane {
  readonly normal: Vec3;
  readonly point: Vec3;
}

/**
 * Every bounding plane of the solid. A barrel's readable faces are its side
 * faces only, so the caps have to come from `capFaces` before the grid oracle
 * can tell inside from outside — which is exactly what that field is for.
 */
function boundingPlanes(geometry: DieGeometry): Plane[] {
  return [...geometry.faces, ...geometry.capFaces].map((face) => ({
    normal: face.normal,
    point: face.centroid,
  }));
}

/** The whole closed surface, fan-triangulated, winding preserved. */
function surfaceTriangles(geometry: DieGeometry): [Vec3, Vec3, Vec3][] {
  const triangles: [Vec3, Vec3, Vec3][] = [];
  for (const face of [...geometry.faces, ...geometry.capFaces]) {
    for (let i = 1; i + 1 < face.polygon.length; i++) {
      triangles.push([face.polygon[0], face.polygon[i], face.polygon[i + 1]]);
    }
  }
  return triangles;
}

/**
 * The volume centroid of the solid as returned, computed from its surface by
 * the divergence theorem. Should be the origin: `finishSolid` centres on it.
 */
function volumeCentroid(geometry: DieGeometry): Vec3 {
  let volume = 0;
  const weighted = [0, 0, 0];
  for (const [a, b, c] of surfaceTriangles(geometry)) {
    const tetVolume = dot(a, cross(b, c)) / 6;
    volume += tetVolume;
    for (let axis = 0; axis < 3; axis++) {
      weighted[axis] += ((a[axis] + b[axis] + c[axis]) * tetVolume) / 4;
    }
  }
  assert.ok(volume > 0, 'surface encloses no volume');
  return [weighted[0] / volume, weighted[1] / volume, weighted[2] / volume];
}

/**
 * Independent inertia oracle: brute-force grid quadrature over the solid's
 * interior, using half-space tests for point-in-solid. Deliberately shares no
 * code with the tetrahedron decomposition under test.
 */
function gridInertia(geometry: DieGeometry, steps: number): number[] {
  let min: Vec3 = [Infinity, Infinity, Infinity];
  let max: Vec3 = [-Infinity, -Infinity, -Infinity];
  for (const vertex of geometry.vertices) {
    min = [Math.min(min[0], vertex[0]), Math.min(min[1], vertex[1]), Math.min(min[2], vertex[2])];
    max = [Math.max(max[0], vertex[0]), Math.max(max[1], vertex[1]), Math.max(max[2], vertex[2])];
  }
  const span: Vec3 = [max[0] - min[0], max[1] - min[1], max[2] - min[2]];
  const planes = boundingPlanes(geometry);
  const covariance = new Array<number>(9).fill(0);
  let count = 0;
  for (let i = 0; i < steps; i++) {
    const x = min[0] + (span[0] * (i + 0.5)) / steps;
    for (let j = 0; j < steps; j++) {
      const y = min[1] + (span[1] * (j + 0.5)) / steps;
      for (let k = 0; k < steps; k++) {
        const z = min[2] + (span[2] * (k + 0.5)) / steps;
        const point: Vec3 = [x, y, z];
        let inside = true;
        for (const plane of planes) {
          if (dot(plane.normal, sub(point, plane.point)) > 0) {
            inside = false;
            break;
          }
        }
        if (!inside) continue;
        count++;
        for (let r = 0; r < 3; r++) {
          for (let c = 0; c < 3; c++) covariance[r * 3 + c] += point[r] * point[c];
        }
      }
    }
  }
  assert.ok(count > 0, 'grid quadrature found no interior samples');
  for (let n = 0; n < 9; n++) covariance[n] /= count;
  const trace = covariance[0] + covariance[4] + covariance[8];
  return covariance.map((value, index) => (index % 4 === 0 ? trace : 0) - value);
}

const STANDARD_FACE_COUNTS = [4, 6, 8, 10, 12, 20];
const BARREL_FACE_COUNTS = [3, 7, 16];

describe('die geometry', () => {
  for (const faceCount of [...STANDARD_FACE_COUNTS, ...BARREL_FACE_COUNTS]) {
    describe(`d${faceCount}`, () => {
      const geometry = dieGeometry(faceCount);

      it('reports the requested face count and exposes exactly that many readable faces', () => {
        assert.equal(geometry.faceCount, faceCount);
        assert.equal(geometry.faces.length, faceCount);
      });

      it('has unit-length face normals', () => {
        for (const [index, face] of geometry.faces.entries()) {
          close(mag(face.normal), 1, 1e-12, `face ${index} normal length`);
        }
      });

      it('points every face normal away from the centroid', () => {
        // The solid is centred on its centroid, so the origin is the centroid.
        for (const [index, face] of geometry.faces.entries()) {
          assert.ok(
            dot(face.normal, face.centroid) > 0,
            `face ${index} normal points toward the centroid`,
          );
        }
      });

      it('has planar face polygons', () => {
        // Where this earns its keep: the barrel path, whose loops are asserted
        // to be planar rather than derived from a plane. On the closed-form
        // path it is near-tautological — `convexHullFaces` only admits a vertex
        // into a loop when it is already within 1e-9 of that loop's plane, so a
        // genuinely non-planar solid (a mis-scaled d10 apex, say) still passes
        // here and is caught downstream by the polygon-arity assertions.
        for (const [index, face] of geometry.faces.entries()) {
          assert.ok(face.polygon.length >= 3, `face ${index} polygon has too few vertices`);
          for (const vertex of face.polygon) {
            close(
              dot(face.normal, sub(vertex, face.centroid)),
              0,
              PLANARITY_TOLERANCE,
              `face ${index} vertex off its plane`,
            );
          }
        }
      });

      it('reports a face centroid that is the mean of its polygon', () => {
        for (const [index, face] of geometry.faces.entries()) {
          for (let axis = 0; axis < 3; axis++) {
            const mean =
              face.polygon.reduce((sum, vertex) => sum + vertex[axis], 0) / face.polygon.length;
            close(face.centroid[axis], mean, 1e-12, `face ${index} centroid axis ${axis}`);
          }
        }
      });

      it('is convex: every vertex lies on the inner side of every face plane', () => {
        for (const [index, face] of geometry.faces.entries()) {
          for (const vertex of geometry.vertices) {
            assert.ok(
              dot(face.normal, sub(vertex, face.centroid)) <= PLANARITY_TOLERANCE,
              `vertex ${vertex.join(',')} is outside face ${index}`,
            );
          }
        }
      });

      it('uses only face polygon vertices drawn from the vertex list', () => {
        for (const face of geometry.faces) {
          for (const vertex of face.polygon) {
            assert.ok(
              geometry.vertices.some(
                (candidate) =>
                  Math.abs(candidate[0] - vertex[0]) < 1e-12 &&
                  Math.abs(candidate[1] - vertex[1]) < 1e-12 &&
                  Math.abs(candidate[2] - vertex[2]) < 1e-12,
              ),
              `face polygon vertex ${vertex.join(',')} is not in the vertex list`,
            );
          }
        }
      });

      it('is centred on its volume centroid', () => {
        // Recomputed from the returned surface, not from the vertex mean: for
        // every face count here the two coincide, so this assertion only bites
        // on an asymmetric solid (see 'volume centroid of an asymmetric solid').
        const centroid = volumeCentroid(geometry);
        for (let axis = 0; axis < 3; axis++) {
          close(centroid[axis], 0, 1e-12, `volume centroid axis ${axis} is not at the origin`);
        }
      });

      it('exposes a closed surface: readable faces plus caps', () => {
        const closedFormCaps = STANDARD_FACE_COUNTS.includes(faceCount);
        assert.equal(
          geometry.capFaces.length,
          closedFormCaps ? 0 : 2 * faceCount,
          `d${faceCount} has the wrong number of cap faces`,
        );
        for (const face of geometry.capFaces) {
          close(mag(face.normal), 1, 1e-12, 'cap face normal length');
          assert.ok(dot(face.normal, face.centroid) > 0, 'cap face normal points inward');
        }
      });

      it('reports the bounding radius as the farthest vertex distance', () => {
        const farthest = Math.max(...geometry.vertices.map((vertex) => mag(vertex)));
        close(geometry.boundingRadius, farthest, 1e-12, 'boundingRadius');
      });

      /**
       * `nominalRadius` is the radius the RENDERER normalizes by — half the
       * die's nominal box — and for everything except a barrel that is the
       * circumsphere. A barrel is normalized by its SHORT axis instead, so its
       * length overflows the box and its faces are drawn at the size the box
       * deserves. See `barrelSolid`.
       */
      it('normalizes by the circumsphere, except a barrel, by its short axis', () => {
        if (STANDARD_FACE_COUNTS.includes(faceCount)) {
          close(geometry.nominalRadius, geometry.boundingRadius, 1e-12, 'nominal radius');
          return;
        }
        // A barrel's long axis is z, so its short semi-axis is the radius of
        // the smallest cylinder about that axis.
        const shortAxis = Math.max(
          ...geometry.vertices.map((vertex) => Math.hypot(vertex[0], vertex[1])),
        );
        close(geometry.nominalRadius, shortAxis, 1e-12, 'barrel nominal radius');
        // Strictly smaller than the circumsphere, which is the whole point: at
        // a fixed `--die-size` every length on the solid grows by this ratio.
        assert.ok(
          geometry.boundingRadius / geometry.nominalRadius >= (faceCount === 3 ? 1.35 : 2.1),
          `d${faceCount} only gains ${geometry.boundingRadius / geometry.nominalRadius}x from short-axis normalization`,
        );
      });

      it('has a symmetric, positive-definite inertia tensor', () => {
        assert.equal(geometry.inertiaTensor.length, 9);
        for (const value of geometry.inertiaTensor) assert.ok(Number.isFinite(value));
        close(geometry.inertiaTensor[1], geometry.inertiaTensor[3], 1e-15, 'tensor xy/yx');
        close(geometry.inertiaTensor[2], geometry.inertiaTensor[6], 1e-15, 'tensor xz/zx');
        close(geometry.inertiaTensor[5], geometry.inertiaTensor[7], 1e-15, 'tensor yz/zy');
        assert.ok(
          isPositiveDefinite(geometry.inertiaTensor),
          `inertia tensor is not positive definite: ${geometry.inertiaTensor.join(',')}`,
        );
      });

      it('reports the reading rule the solid actually admits', () => {
        // 'up-face' unless the solid presents no single face upward: the d4
        // (three faces tilt equally up, so it is read from the apex) and every
        // ODD-sided barrel (an edge points up, and the two best up-face
        // candidates are a floating-point tie).
        const oddBarrel = !STANDARD_FACE_COUNTS.includes(faceCount) && faceCount % 2 === 1;
        const expected =
          faceCount === 4 ? 'top-vertex' : oddBarrel ? 'down-face' : 'up-face';
        assert.equal(geometry.readingRule, expected);
      });

      it('reports the same rule die-faces.ts derives from the shape', () => {
        // `resolveReadingRule` works the rule out at runtime by asking whether
        // every face has an antipode. Asserting the two agree for every shape is
        // what lets that derivation be dropped in favour of the stated rule.
        assert.equal(resolveReadingRule(geometry), geometry.readingRule);
      });
    });
  }

  describe('closed-form solids', () => {
    it('builds a cube whose inertia matches m*s^2/6 on the diagonal', () => {
      const cube = dieGeometry(6);
      // Vertices at (+/-1, +/-1, +/-1): side length 2.
      const side = 2;
      const expected = (side * side) / 6;
      for (const index of [0, 4, 8]) {
        close(cube.inertiaTensor[index], expected, INERTIA_TOLERANCE, `cube diagonal ${index}`);
      }
      for (const index of [1, 2, 3, 5, 6, 7]) {
        close(cube.inertiaTensor[index], 0, INERTIA_TOLERANCE, `cube off-diagonal ${index}`);
      }
      close(cube.boundingRadius, Math.sqrt(3), 1e-12, 'cube circumradius');
      close(cube.nominalRadius, Math.sqrt(3), 1e-12, 'cube nominal radius');
      for (const face of cube.faces) assert.equal(face.polygon.length, 4);
    });

    it('builds a regular tetrahedron whose inertia matches m*a^2/20', () => {
      const tetrahedron = dieGeometry(4);
      const edge = mag(sub(tetrahedron.vertices[0], tetrahedron.vertices[1]));
      const expected = (edge * edge) / 20;
      for (const index of [0, 4, 8]) {
        close(
          tetrahedron.inertiaTensor[index],
          expected,
          INERTIA_TOLERANCE,
          `tetrahedron diagonal ${index}`,
        );
      }
      for (const face of tetrahedron.faces) assert.equal(face.polygon.length, 3);
    });

    it('builds a regular octahedron whose inertia matches m*R^2/5', () => {
      const octahedron = dieGeometry(8);
      // m*R^2/5 is stated about the CIRCUMRADIUS, so this reads the honest one.
      const expected = (octahedron.boundingRadius * octahedron.boundingRadius) / 5;
      for (const index of [0, 4, 8]) {
        close(
          octahedron.inertiaTensor[index],
          expected,
          INERTIA_TOLERANCE,
          `octahedron diagonal ${index}`,
        );
      }
      for (const face of octahedron.faces) assert.equal(face.polygon.length, 3);
    });

    it('gives every face-transitive solid an isotropic inertia tensor', () => {
      for (const faceCount of [4, 6, 8, 12, 20]) {
        const geometry = dieGeometry(faceCount);
        const trace =
          geometry.inertiaTensor[0] + geometry.inertiaTensor[4] + geometry.inertiaTensor[8];
        for (const index of [0, 4, 8]) {
          close(
            geometry.inertiaTensor[index],
            trace / 3,
            INERTIA_TOLERANCE,
            `d${faceCount} diagonal ${index} is not isotropic`,
          );
        }
        for (const index of [1, 2, 3, 5, 6, 7]) {
          close(
            geometry.inertiaTensor[index],
            0,
            INERTIA_TOLERANCE,
            `d${faceCount} off-diagonal ${index} is not zero`,
          );
        }
      }
    });

    it('builds a dodecahedron with twelve pentagons and an icosahedron with twenty triangles', () => {
      const dodecahedron = dieGeometry(12);
      for (const face of dodecahedron.faces) assert.equal(face.polygon.length, 5);
      assert.equal(dodecahedron.vertices.length, 20);

      const icosahedron = dieGeometry(20);
      for (const face of icosahedron.faces) assert.equal(face.polygon.length, 3);
      assert.equal(icosahedron.vertices.length, 12);
    });

    it('builds a pentagonal trapezohedron for the d10', () => {
      const d10 = dieGeometry(10);
      assert.equal(d10.vertices.length, 12);
      for (const face of d10.faces) assert.equal(face.polygon.length, 4);
      // Two apexes on the axis of symmetry, ten zig-zag equatorial vertices.
      const onAxis = d10.vertices.filter(
        (vertex) => Math.abs(vertex[0]) < 1e-12 && Math.abs(vertex[1]) < 1e-12,
      );
      assert.equal(onAxis.length, 2);
      // Not isotropic: a trapezohedron is only transversely isotropic.
      close(d10.inertiaTensor[0], d10.inertiaTensor[4], INERTIA_TOLERANCE, 'd10 Ixx vs Iyy');
      assert.ok(
        Math.abs(d10.inertiaTensor[8] - d10.inertiaTensor[0]) > 1e-3,
        'd10 should not be isotropic',
      );
    });
  });

  describe('inertia cross-checked against grid quadrature', () => {
    // What this gate actually bounds. The observed disagreement at 90 steps is
    // at most 0.068% of the trace (the d20; every other shape is under 0.06%),
    // and it shrinks as the grid is refined. But the tolerance is scaled by the
    // *trace*, which is about three times a single diagonal entry, so
    // `0.005 * trace` is not a 0.5% bound on an entry: measured directly, the
    // smallest uniform relative error in the tensor that it rejects is 1.2%
    // (d7) to 1.7% (d20). Call it ~1.5%. That is still a real constraint -- it
    // catches dropped or mis-signed terms by an order of magnitude -- but it is
    // three times looser than "0.5%" reads, so do not lean on it for anything
    // subtle. Note also that the oracle takes its clipping planes from the
    // geometry under test, so it constrains the tensor computation only; it is
    // blind to the solid's proportions (see the square-side-face assertion).
    const GRID_STEPS = 90;
    const GRID_TOLERANCE_FRACTION = 0.005;

    for (const faceCount of [6, 10, 12, 20, ...BARREL_FACE_COUNTS]) {
      it(`matches an independent numeric integration for the d${faceCount}`, () => {
        const geometry = dieGeometry(faceCount);
        const numeric = gridInertia(geometry, GRID_STEPS);
        const scale =
          geometry.inertiaTensor[0] + geometry.inertiaTensor[4] + geometry.inertiaTensor[8];
        for (let index = 0; index < 9; index++) {
          close(
            numeric[index],
            geometry.inertiaTensor[index],
            GRID_TOLERANCE_FRACTION * scale,
            `d${faceCount} tensor entry ${index} disagrees with grid quadrature`,
          );
        }
      });
    }
  });

  describe('barrels', () => {
    for (const faceCount of BARREL_FACE_COUNTS) {
      it(`builds a d${faceCount} barrel with pointed caps that are not readable faces`, () => {
        const geometry = dieGeometry(faceCount);
        // faceCount counts SIDE faces only; the caps exist in the vertex list
        // but are not readable, so faces.length must equal faceCount.
        assert.equal(geometry.faceCount, faceCount);
        assert.equal(geometry.faces.length, faceCount);
        // Two rings of faceCount vertices plus two apex vertices.
        assert.equal(geometry.vertices.length, 2 * faceCount + 2);
        for (const face of geometry.faces) assert.equal(face.polygon.length, 4);

        const sorted = [...geometry.vertices].sort((a, b) => a[2] - b[2]);
        const bottom = sorted[0];
        const top = sorted[sorted.length - 1];
        for (const apex of [bottom, top]) {
          close(apex[0], 0, 1e-12, 'apex is off the axis in x');
          close(apex[1], 0, 1e-12, 'apex is off the axis in y');
          // A pointed cap: the apex is the unique extreme vertex along the axis.
          const ties = geometry.vertices.filter((vertex) => Math.abs(vertex[2] - apex[2]) < 1e-9);
          assert.equal(ties.length, 1, 'the cap is flat, so the die could rest on an end');
          // Strictly inside every side-face plane.
          for (const face of geometry.faces) {
            assert.ok(
              dot(face.normal, sub(apex, face.centroid)) < -1e-6,
              'apex is not strictly inside the side faces',
            );
          }
        }
      });

      it(`gives the d${faceCount} barrel rectangular side faces on a unit ring`, () => {
        // The side faces are rectangles: the ring chord wide (which fixes the
        // ring radius at 1) and the band tall. They are only SQUARE for the d3;
        // for every other N the band is stretched, which is what keeps the caps
        // unstable without a needle (see 'barrel resting poses').
        const geometry = dieGeometry(faceCount);
        const chord = 2 * Math.sin(Math.PI / faceCount);
        for (const [index, face] of geometry.faces.entries()) {
          assert.equal(face.polygon.length, 4);
          const edges = face.polygon.map((vertex, i) =>
            sub(face.polygon[(i + 1) % face.polygon.length], vertex),
          );
          // Opposite edges equal, adjacent edges perpendicular: a rectangle.
          for (let i = 0; i < 4; i++) {
            close(
              mag(edges[i]),
              mag(edges[(i + 2) % 4]),
              1e-12,
              `d${faceCount} side face ${index} is not a parallelogram`,
            );
            close(
              dot(edges[i], edges[(i + 1) % 4]) / (mag(edges[i]) * mag(edges[(i + 1) % 4])),
              0,
              1e-12,
              `d${faceCount} side face ${index} corner ${i} is not square`,
            );
          }
          // Two horizontal edges of exactly the ring chord (which is what pins
          // the ring radius at 1) and two edges parallel to the axis.
          const horizontal = edges.filter((edge) => Math.abs(edge[2]) < 1e-12);
          const vertical = edges.filter(
            (edge) => Math.abs(edge[0]) < 1e-12 && Math.abs(edge[1]) < 1e-12,
          );
          assert.equal(horizontal.length, 2, `d${faceCount} side face ${index} is not upright`);
          assert.equal(vertical.length, 2, `d${faceCount} side face ${index} is not upright`);
          for (const edge of horizontal) {
            close(mag(edge), chord, 1e-12, `d${faceCount} side face ${index} is not a chord wide`);
          }
          // Centred on the band, one apothem out from the axis.
          close(
            mag([face.centroid[0], face.centroid[1], 0]),
            Math.cos(Math.PI / faceCount),
            1e-12,
            'side face is off the unit ring',
          );
          close(face.centroid[2], 0, 1e-12, 'side face is not centred on the band');
        }
      });

      it(`gives the d${faceCount} barrel a transversely isotropic inertia tensor`, () => {
        const geometry = dieGeometry(faceCount);
        close(
          geometry.inertiaTensor[0],
          geometry.inertiaTensor[4],
          INERTIA_TOLERANCE,
          `d${faceCount} Ixx vs Iyy`,
        );
        for (const index of [1, 2, 3, 5, 6, 7]) {
          close(
            geometry.inertiaTensor[index],
            0,
            INERTIA_TOLERANCE,
            `d${faceCount} off-diagonal ${index}`,
          );
        }
      });
    }

    it('does not use a barrel for the face counts that have closed forms', () => {
      for (const faceCount of STANDARD_FACE_COUNTS) {
        const geometry = dieGeometry(faceCount);
        assert.notEqual(
          geometry.vertices.length,
          2 * faceCount + 2,
          `d${faceCount} looks like a barrel`,
        );
      }
    });
  });

  describe('barrel resting poses', () => {
    /**
     * The property the whole barrel construction exists for: a barrel comes to
     * rest on a readable SIDE face and never on a cap facet.
     *
     * Both halves below are needed and neither substitutes for the other. The
     * previous version of this suite proved "the cap is pointed" structurally —
     * the apex is the unique extreme vertex along the axis — which is true, is
     * still asserted above, and says nothing about whether the die can rest on
     * one of the 2N flat facets that make the cap up. It could: `dice-sim.ts`
     * landed a d7 on a cap facet in 64% of unretried rolls. So the analytic
     * half measures facet stability directly, and the empirical half rolls the
     * die.
     */

    /**
     * Signed distance from the centre of mass's projection onto this face's
     * plane to the nearest polygon edge. Positive means the projection falls
     * INSIDE the polygon, i.e. the solid can balance on this face; negative
     * means it topples off it.
     *
     * The solid is centred on its centroid, so the centre of mass is the origin
     * and `dot(normal, centroid)` is the plane's distance from it.
     */
    function supportMargin(face: DieFace): number {
      const depth = dot(face.normal, face.centroid);
      const projected: Vec3 = [
        face.normal[0] * depth,
        face.normal[1] * depth,
        face.normal[2] * depth,
      ];
      let margin = Infinity;
      for (let i = 0; i < face.polygon.length; i++) {
        const a = face.polygon[i];
        const b = face.polygon[(i + 1) % face.polygon.length];
        const edge = sub(b, a);
        // Left of every edge, walking the polygon counter-clockwise about its
        // outward normal, is inside it.
        margin = Math.min(margin, dot(cross(edge, sub(projected, a)), face.normal) / mag(edge));
      }
      return margin;
    }

    /**
     * Face counts spanning the whole barrel range: the smallest barrel, the
     * odd ones the reading rule cares about, an even one, and two so many-sided
     * that the solid is nearly a cylinder — which is the regime where keeping
     * the caps unstable is hardest and where a needle would show up.
     */
    const STABILITY_FACE_COUNTS = [3, 5, 7, 9, 11, 16, 24, 100] as const;

    /**
     * How far outside a cap facet the centre of mass must project, as a
     * fraction of the BOUNDING radius. Scale-free on purpose: `dice-sim.ts`
     * normalises every die to a bounding radius of 1, so this is the margin in
     * the units the physics actually works in — and it is deliberately not
     * `nominalRadius`, which is the RENDERER's normalisation and is a barrel's
     * short axis rather than its circumsphere. The measured values run -0.168
     * (large N) to -0.189 (d3); before the fix a d7 scored +0.161, i.e. stable
     * by about as much as it is now unstable.
     */
    const CAP_INSTABILITY_MARGIN = 0.1;

    it('never has a stable cap facet, from the d3 to the d100', () => {
      for (const faceCount of STABILITY_FACE_COUNTS) {
        const geometry = dieGeometry(faceCount);
        assert.equal(geometry.capFaces.length, 2 * faceCount);
        for (const [index, face] of geometry.capFaces.entries()) {
          const margin = supportMargin(face) / geometry.boundingRadius;
          assert.ok(
            margin < -CAP_INSTABILITY_MARGIN,
            `d${faceCount} cap facet ${index} is a stable rest: margin ${margin} of a bounding radius`,
          );
        }
      }
    });

    /**
     * How far INSIDE a readable side face the centre of mass must project, as
     * a fraction of the BOUNDING radius (see `CAP_INSTABILITY_MARGIN`), and the
     * honest edge of what this module supports.
     *
     * `margin > 0` alone is not a bound: the side margin is `sin(pi/N)` in raw
     * units and the bounding radius tends to 2.632, so it falls off as `1/N` —
     * 0.634 at the d3, 0.050 at the d24, 0.012 at the d100, and 0.0012 at a
     * hypothetical d1000. That last one is "stable" in exactly the way the
     * needle the `BARREL_CAP_SAFETY` doc rejects was stable: true on paper,
     * balanceable by a breath in practice. `dieGeometry` puts no upper limit on
     * `faceCount`, so this number is the limit, not the type signature: the
     * shapes this module claims to produce real dice for are those whose side
     * faces are at least a hundredth of a bounding radius wide, i.e. N <= 119.
     * Beyond that the geometry is still well-formed and the physics still runs,
     * but nothing here asserts a die that many sides comes to rest readably.
     */
    const SIDE_STABILITY_MARGIN = 0.01;

    it('keeps every readable side face a stable rest', () => {
      // The other direction of the same measurement. Steepening the caps until
      // the die cannot rest on them would be no use if it also stopped resting
      // on its sides: the margin here is exactly half the side face's width,
      // because the centre of mass projects onto the face's centre.
      for (const faceCount of STABILITY_FACE_COUNTS) {
        const geometry = dieGeometry(faceCount);
        for (const [index, face] of geometry.faces.entries()) {
          const margin = supportMargin(face);
          const normalised = margin / geometry.boundingRadius;
          assert.ok(
            normalised > SIDE_STABILITY_MARGIN,
            `d${faceCount} side face ${index} rests on only ${normalised} of a bounding radius`,
          );
          close(
            margin,
            Math.sin(Math.PI / faceCount),
            1e-12,
            `d${faceCount} side face ${index} rests on less than its half-width`,
          );
        }
      }
    });

    it('runs out of side-face stability where the doc says it does', () => {
      // The claim above, as an assertion rather than a comment: N = 119 is the
      // last face count whose side faces clear the margin, and N = 120 is the
      // first that does not. If the barrel proportions change, this is what
      // says the supported range moved.
      const normalisedMargin = (faceCount: number): number => {
        const geometry = dieGeometry(faceCount);
        return supportMargin(geometry.faces[0]) / geometry.boundingRadius;
      };
      assert.ok(
        normalisedMargin(119) > SIDE_STABILITY_MARGIN,
        `d119 margin is ${normalisedMargin(119)}, so the supported range is narrower than documented`,
      );
      assert.ok(
        normalisedMargin(120) <= SIDE_STABILITY_MARGIN,
        `d120 margin is ${normalisedMargin(120)}, so the supported range is wider than documented`,
      );
    });

    it('stays a die rather than a needle', () => {
      // A cap steep enough to be unstable is a cap that makes the die longer.
      // Left unchecked that is a race to a knitting needle -- with square side
      // faces the band half-height goes as sin(pi/N), so the cap height needed
      // would go as N/pi and a d100 would be 48 times longer than it is wide.
      for (let faceCount = 3; faceCount <= 200; faceCount++) {
        if (STANDARD_FACE_COUNTS.includes(faceCount)) continue;
        const geometry = dieGeometry(faceCount);
        const halfLength = Math.max(...geometry.vertices.map((vertex) => Math.abs(vertex[2])));
        // Ring radius is 1, so the die is 2 wide and 2 * halfLength long.
        assert.ok(
          halfLength <= 2.7,
          `d${faceCount} is ${halfLength.toFixed(2)} times longer than it is wide`,
        );
        assert.ok(halfLength >= 1, `d${faceCount} is flatter than it is wide`);
      }
    });

    it('lands on a readable side face in the physics, for every seed', () => {
      // The check that would have caught the original defect. `simulateRoll`
      // re-throws a cocked die, so a shape that lands unreadable only sometimes
      // could still slip through a small sample -- but the pre-fix barrels
      // failed this badly even with the retries (12% of settled d5s through 83%
      // of d24s, tilted about 61 degrees), so the retry loop is not what is
      // being measured here.
      /**
       * The tolerance has to shrink with the face count, because a fixed one
       * stops meaning anything.
       *
       * Adjacent side-face normals of an N-barrel are `2 * pi / N` apart, so a
       * barrel balanced on a side EDGE — the failure this is looking for — sits
       * only `pi / N` from a face normal: 60 degrees at the d3, 25.7 at the d7,
       * 11.25 at the d16, and 7.5 at the d24. A flat 5 degrees still separates
       * a good landing from an edge at every count tested here, but only just
       * at the top, and by N = 36 it would not separate them at all. Half the
       * edge angle keeps the same PROPORTIONAL margin at every count, so the
       * d24 is held to 3.75 degrees rather than 5 and the assertion stays a
       * statement about the shape instead of about the number 5.
       */
      const tiltLimit = (faceCount: number): number => Math.min(5, 180 / (2 * faceCount));
      for (const faceCount of [3, 5, 7, 9, 16, 24]) {
        const MAX_TILT_DEGREES = tiltLimit(faceCount);
        const geometry = dieGeometry(faceCount);
        // An odd barrel is read from the face it RESTS on, an even one from the
        // face opposite; either way the presented face must be square to the
        // floor, and the die must be lying on a readable face at all.
        const readDirection: Vec3 =
          resolveReadingRule(geometry) === 'down-face' ? [0, -1, 0] : WORLD_UP;
        for (const seed of [1, 2, 3, 5, 8, 13, 21, 34]) {
          const roll = simulateRoll({
            seed,
            geometry,
            dieCount: 2,
            bounds: { x: 6, y: 6, z: 6 },
          });
          for (const [die, trajectory] of roll.dice.entries()) {
            const orientation = trajectory.restingOrientation;
            const label = `d${faceCount} seed ${seed} die ${die}`;

            const resting = Math.max(
              ...geometry.faces.map((face) => dot(rotate(orientation, face.normal), [0, -1, 0])),
            );
            const restTilt = (Math.acos(Math.min(1, resting)) * 180) / Math.PI;
            assert.ok(
              restTilt <= MAX_TILT_DEGREES,
              `${label} came to rest ${restTilt.toFixed(1)} degrees off any readable face -- it is on a cap facet`,
            );

            const presented = presentedFaceIndex(geometry, orientation);
            const aligned = dot(
              rotate(orientation, geometry.faces[presented].normal),
              readDirection,
            );
            const readTilt = (Math.acos(Math.min(1, aligned)) * 180) / Math.PI;
            assert.ok(
              readTilt <= MAX_TILT_DEGREES,
              `${label} presents face ${presented} at ${readTilt.toFixed(1)} degrees off the reading direction`,
            );
          }
        }
      }
    });
  });

  describe('volume centroid of an asymmetric solid', () => {
    /**
     * A square pyramid: base side 2 in the plane z = 0, apex at z = 2.
     *
     * Every face count `dieGeometry` supports is centrally symmetric, so its
     * vertex mean and its volume centroid coincide and neither the centroid
     * computation nor the centring in `finishSolid` does any work. A pyramid
     * separates them by construction: the vertex mean sits at z = 2/5 (five
     * vertices, four of them on the base) while the volume centroid sits at
     * z = h/4 = 1/2. Anything that centres on the vertex mean, or fails to
     * centre at all, lands the base somewhere other than z = -1/2.
     */
    function squarePyramid(): RawSolid {
      const vertices: Vec3[] = [
        [1, 1, 0],
        [1, -1, 0],
        [-1, -1, 0],
        [-1, 1, 0],
        [0, 0, 2],
      ];
      const surface = [
        [0, 1, 2, 3], // base: a cap, not a readable face
        [4, 0, 1],
        [4, 1, 2],
        [4, 2, 3],
        [4, 3, 0],
      ];
      return { vertices, surface, readable: [1, 2, 3, 4], readingRule: 'up-face' };
    }

    const pyramid = finishSolid(squarePyramid());

    it('centres on the volume centroid, not on the vertex mean', () => {
      const heights = pyramid.vertices.map((vertex) => vertex[2]);
      const apex = Math.max(...heights);
      const base = Math.min(...heights);
      // Volume centroid at h/4 above the base, so base -> -0.5, apex -> +1.5.
      // The vertex mean would put them at -0.4 and +1.6; no centring at all
      // would leave them at 0 and +2.
      close(base, -0.5, 1e-12, 'pyramid base height after centring');
      close(apex, 1.5, 1e-12, 'pyramid apex height after centring');
      const centroid = volumeCentroid(pyramid);
      for (let axis = 0; axis < 3; axis++) {
        close(centroid[axis], 0, 1e-12, `pyramid volume centroid axis ${axis}`);
      }
    });

    it('reports an inertia tensor about the volume centroid', () => {
      // Analytic, for a right square pyramid of base side a = 2, height h = 2,
      // unit mass, about its centroid:
      //   Izz = (a^2 + b^2) / 20            = 0.4
      //   Ixx = Iyy = b^2 / 20 + 3 h^2 / 80 = 0.35
      // Both are centroidal: about the vertex mean Ixx would be 0.44, and about
      // the base-plane origin 0.6.
      close(pyramid.inertiaTensor[0], 0.35, INERTIA_TOLERANCE, 'pyramid Ixx');
      close(pyramid.inertiaTensor[4], 0.35, INERTIA_TOLERANCE, 'pyramid Iyy');
      close(pyramid.inertiaTensor[8], 0.4, INERTIA_TOLERANCE, 'pyramid Izz');
      for (const index of [1, 2, 3, 5, 6, 7]) {
        close(pyramid.inertiaTensor[index], 0, INERTIA_TOLERANCE, `pyramid off-diagonal ${index}`);
      }
    });

    it('keeps the readable/cap split and orients every face outward', () => {
      assert.equal(pyramid.faceCount, 4);
      assert.equal(pyramid.capFaces.length, 1);
      assert.equal(pyramid.capFaces[0].polygon.length, 4);
      for (const face of [...pyramid.faces, ...pyramid.capFaces]) {
        assert.ok(
          dot(face.normal, face.centroid) > 0,
          'a face of the pyramid points toward the centroid',
        );
      }
    });
  });

  describe('author-ordered loops', () => {
    /**
     * `RawSolid.oriented` is the opt-in that lets a NON-CONVEX solid through.
     *
     * The three shapes below are prisms over simple, non-convex polygons, and
     * every one of them is rejected without the flag — not because anything
     * downstream cannot handle them, but because `orientLoop` rederives a
     * winding it can only derive for a convex-ish solid, and the half-edge
     * check then correctly rejects the permuted loops it produced. The tests
     * pin both halves: the rejection without the flag (so the flag is provably
     * load-bearing) and the analytically correct solid with it.
     */
    type Point2 = readonly [number, number];

    /** An L: a 2x2 square missing its top-right quadrant. Area 3, one reflex corner. */
    const L_SHAPE: Point2[] = [[0, 0], [2, 0], [2, 1], [1, 1], [1, 2], [0, 2]];
    /** A U: a 3x3 square with a 1x2 notch cut down from the top. Area 7. */
    const U_SHAPE: Point2[] = [[0, 0], [3, 0], [3, 3], [2, 3], [2, 1], [1, 1], [1, 3], [0, 3]];
    /**
     * A meeple silhouette: 27 vertices, 10 of them reflex — splayed legs with a
     * notch between them, arms out with an underarm indent, and a head on a
     * neck. This is the shape the whole opt-in is for, and nothing about it is
     * star-shaped: the notch between the legs faces straight back at the
     * vertex mean.
     */
    const MEEPLE: Point2[] = [
      [-4.0, 0.0], [-1.4, 0.0], [-0.9, 2.0], [-0.35, 2.9], [0.35, 2.9], [0.9, 2.0],
      [1.4, 0.0], [4.0, 0.0], [4.4, 1.6], [3.4, 3.6], [4.8, 4.0], [5.0, 5.0],
      [4.4, 5.8], [2.6, 5.4], [2.0, 6.6], [1.5, 7.2], [0.9, 9.0], [0.0, 9.6],
      [-0.9, 9.0], [-1.5, 7.2], [-2.0, 6.6], [-2.6, 5.4], [-4.4, 5.8], [-5.0, 5.0],
      [-4.8, 4.0], [-3.4, 3.6], [-4.4, 1.6],
    ];

    /** How many corners turn the wrong way: 0 for a convex polygon. */
    function reflexCorners(polygon: Point2[]): number {
      const n = polygon.length;
      let count = 0;
      for (let i = 0; i < n; i++) {
        const a = polygon[(i + n - 1) % n];
        const b = polygon[i];
        const c = polygon[(i + 1) % n];
        if ((b[0] - a[0]) * (c[1] - b[1]) - (b[1] - a[1]) * (c[0] - b[0]) < 0) count++;
      }
      return count;
    }

    /**
     * Area, centroid and second moments of a simple polygon, straight from the
     * shoelace formulas. Deliberately shares no code with the module's
     * tetrahedron decomposition: this is the independent oracle.
     */
    function polygonMoments(polygon: Point2[]) {
      let area = 0;
      let cx = 0;
      let cy = 0;
      let momentXX = 0;
      let momentYY = 0;
      let momentXY = 0;
      for (let i = 0; i < polygon.length; i++) {
        const [x0, y0] = polygon[i];
        const [x1, y1] = polygon[(i + 1) % polygon.length];
        const shoelace = x0 * y1 - x1 * y0;
        area += shoelace / 2;
        cx += (x0 + x1) * shoelace;
        cy += (y0 + y1) * shoelace;
        momentXX += (y0 * y0 + y0 * y1 + y1 * y1) * shoelace;
        momentYY += (x0 * x0 + x0 * x1 + x1 * x1) * shoelace;
        momentXY += (x0 * y1 + 2 * x0 * y0 + 2 * x1 * y1 + x1 * y0) * shoelace;
      }
      cx /= 6 * area;
      cy /= 6 * area;
      return {
        area,
        cx,
        cy,
        // Second moments about the polygon's own centroid, by parallel axis.
        xx: momentXX / 12 - area * cy * cy,
        yy: momentYY / 12 - area * cx * cx,
        xy: momentXY / 24 - area * cx * cy,
      };
    }

    /**
     * Extrude a CCW polygon along z into a closed prism, wound outward by hand:
     * the +z cap in polygon order, the -z cap reversed, and each side wall as
     * `bottom_i -> bottom_j -> top_j -> top_i`, whose winding is outward
     * exactly because the polygon is CCW.
     */
    function extrude(polygon: Point2[], height: number, oriented: boolean): RawSolid {
      const n = polygon.length;
      const half = height / 2;
      const vertices: Vec3[] = [
        ...polygon.map(([x, y]): Vec3 => [x, y, -half]),
        ...polygon.map(([x, y]): Vec3 => [x, y, half]),
      ];
      const surface: number[][] = [
        polygon.map((_, i) => n + i),
        polygon.map((_, i) => n - 1 - i),
      ];
      for (let i = 0; i < n; i++) {
        const j = (i + 1) % n;
        surface.push([i, j, n + j, n + i]);
      }
      return {
        vertices,
        surface,
        readable: surface.map((_, i) => i),
        readingRule: 'up-face',
        ...(oriented ? { oriented: true } : {}),
      };
    }

    /** The volume the returned surface actually encloses, by the divergence theorem. */
    function enclosedVolume(geometry: DieGeometry): number {
      let volume = 0;
      for (const [a, b, c] of surfaceTriangles(geometry)) volume += dot(a, cross(b, c)) / 6;
      return volume;
    }

    const SHAPES = [
      { name: 'L-shape', polygon: L_SHAPE, height: 1, corners: 6, reflex: 1, area: 3 },
      { name: 'U-shape', polygon: U_SHAPE, height: 2, corners: 8, reflex: 2, area: 7 },
      { name: 'meeple silhouette', polygon: MEEPLE, height: 3, corners: 27, reflex: 10, area: 53.595 },
    ] as const;

    for (const { name, polygon, height, corners, reflex, area } of SHAPES) {
      describe(`a prism over the ${name}`, () => {
        it('is genuinely non-convex, which is the whole point of the case', () => {
          assert.equal(polygon.length, corners);
          assert.equal(reflexCorners([...polygon]), reflex);
          const measured = polygonMoments([...polygon]).area;
          assert.ok(measured > 0, 'polygon is wound counter-clockwise');
          close(measured, area, 1e-12, `${name} area`);
        });

        it('is REJECTED without the opt-in, because the winding gets rederived', () => {
          // Not a nice-to-have: this is the error a caller sees today. The
          // rederived loops are permuted, so two of them claim the same
          // directed edge and the half-edge check refuses the surface.
          assert.throws(
            () => finishSolid(extrude([...polygon], height, false)),
            /surface is not a manifold: directed edge \d+->\d+ is used twice/,
          );
        });

        // Built per test rather than once for the suite: if the opt-in ever
        // stops working the throw belongs to the test that needed it, not to
        // the whole `describe` body.
        const build = () => finishSolid(extrude([...polygon], height, true));
        const moments = polygonMoments([...polygon]);

        it('encloses the analytic prism volume', () => {
          const geometry = build();
          const expected = moments.area * height;
          const actual = enclosedVolume(geometry);
          close(actual, expected, expected * 1e-15, `${name} prism volume`);
        });

        it('reports the analytic inertia tensor of an extruded polygon', () => {
          const geometry = build();
          // Unit mass, about the centroid, for a polygon extruded along z:
          // Izz is the polygon's polar moment per unit area, and the two
          // in-plane axes pick up the extrusion's own h^2/12.
          const expected = [
            moments.xx / moments.area + (height * height) / 12,
            -moments.xy / moments.area,
            0,
            -moments.xy / moments.area,
            moments.yy / moments.area + (height * height) / 12,
            0,
            0,
            0,
            (moments.xx + moments.yy) / moments.area,
          ];
          for (let i = 0; i < 9; i++) {
            // Relative where the term is real, absolute where it is zero.
            const tolerance = Math.max(Math.abs(expected[i]) * 1e-14, 1e-12);
            close(geometry.inertiaTensor[i], expected[i], tolerance, `${name} inertia[${i}]`);
          }
        });

        it('points every face outward and keeps every cap planar', () => {
          const geometry = build();
          const n = polygon.length;
          const expectedNormals: Vec3[] = [[0, 0, 1], [0, 0, -1]];
          for (let i = 0; i < n; i++) {
            const [x0, y0] = polygon[i];
            const [x1, y1] = polygon[(i + 1) % n];
            const length = Math.hypot(x1 - x0, y1 - y0);
            // Outward from a CCW polygon is the edge direction turned right.
            expectedNormals.push([(y1 - y0) / length, -(x1 - x0) / length, 0]);
          }
          assert.equal(geometry.faces.length, n + 2);
          geometry.faces.forEach((face, index) => {
            const expected = expectedNormals[index];
            for (let axis = 0; axis < 3; axis++) {
              close(face.normal[axis], expected[axis], 1e-12, `${name} face ${index} normal ${axis}`);
            }
            for (const vertex of face.polygon) {
              close(
                dot(face.normal, sub(vertex, face.centroid)),
                0,
                PLANARITY_TOLERANCE,
                `${name} face ${index} is not planar`,
              );
            }
          });
        });

        it('keeps the author\'s vertex order rather than re-sorting it', () => {
          const geometry = build();
          // The +z cap was handed over as the polygon itself, in order. An
          // angular re-sort about its own vertex mean does not give an L, a U
          // or a meeple back — which is exactly why the unflagged build above
          // throws — so seeing the loop come back verbatim is the assertion
          // that step is skipped and not merely harmless.
          const n = polygon.length;
          const cap = geometry.faces[0].polygon;
          assert.equal(cap.length, n);
          for (let i = 0; i < n; i++) {
            assert.deepStrictEqual(cap[i], geometry.vertices[n + i], `cap vertex ${i}`);
          }
        });
      });
    }

    it('still rejects a claimed winding that is not true', () => {
      // The opt-in skips DERIVATION, not VALIDATION. An author who reverses one
      // wall and says the surface is oriented gets the manifold error, not a
      // silently inside-out solid.
      const raw = extrude([...L_SHAPE], 1, true);
      const broken = raw.surface.map((loop, index) => (index === 3 ? [...loop].reverse() : loop));
      assert.throws(
        () => finishSolid({ ...raw, surface: broken }),
        /not a manifold|not closed/,
      );
    });

    it('measures a convex solid identically whether the winding is declared or derived', () => {
      // The regression guarantee in miniature. Hand the same already-correct
      // surface in twice, once flagged and once not: every measured number must
      // come back identical BIT FOR BIT, because on a convex solid the flag has
      // nothing to change. `deepStrictEqual` compares numbers with `Object.is`,
      // so a -0 or a one-ULP drift fails here.
      //
      // The one thing that legitimately differs is where each loop STARTS:
      // `orientLoop` re-sorts by angle measured from `points[0]`, which rotates
      // the loop to begin at its own smallest angle rather than at the vertex
      // the author wrote first. A cyclic rotation is the same polygon, and the
      // normals and centroids below are computed identically either way.
      const square: Point2[] = [[-1, -1], [1, -1], [1, 1], [-1, 1]];
      const derived = finishSolid(extrude(square, 3, false));
      const declared = finishSolid(extrude(square, 3, true));

      assert.deepStrictEqual(declared.inertiaTensor, derived.inertiaTensor);
      assert.deepStrictEqual(declared.vertices, derived.vertices);
      assert.deepStrictEqual(declared.nominalRadius, derived.nominalRadius);
      assert.deepStrictEqual(declared.boundingRadius, derived.boundingRadius);
      assert.equal(declared.faceCount, derived.faceCount);
      declared.faces.forEach((face, index) => {
        assert.deepStrictEqual(face.normal, derived.faces[index].normal, `face ${index} normal`);
        assert.deepStrictEqual(face.centroid, derived.faces[index].centroid, `face ${index} centroid`);
        // Same cycle, possibly started elsewhere.
        const other = derived.faces[index].polygon;
        assert.equal(face.polygon.length, other.length);
        const first = face.polygon[0];
        const offset = other.findIndex((point) => point.every((v, k) => Object.is(v, first[k])));
        assert.ok(offset >= 0, `face ${index} does not share a vertex with the derived loop`);
        face.polygon.forEach((point, i) => {
          assert.deepStrictEqual(point, other[(offset + i) % other.length], `face ${index} vertex ${i}`);
        });
      });
    });
  });

  describe('structurally invalid raw solids', () => {
    /**
     * `finishSolid` is an exported seam whose input is hand-built by whoever
     * calls it, so it cannot assume a well-formed surface. Every case here used
     * to be accepted silently, and the first one is the dangerous one: an open
     * surface still integrates to a POSITIVE volume, because the divergence
     * theorem closes the hole through the interior point, so the `volume > 0`
     * guard waves it through and the caller gets an inertia tensor that is
     * wrong by a few percent with no signal at all — and that tensor is what
     * `dice-sim.ts` integrates the die's tumble with.
     */
    function squarePyramid(): {
      vertices: Vec3[];
      surface: number[][];
      readable: number[];
      readingRule: RawSolid['readingRule'];
    } {
      return {
        vertices: [
          [1, 1, 0],
          [1, -1, 0],
          [-1, -1, 0],
          [-1, 1, 0],
          [0, 0, 2],
        ],
        surface: [
          [0, 1, 2, 3], // base
          [4, 0, 1],
          [4, 1, 2],
          [4, 2, 3],
          [4, 3, 0],
        ],
        readable: [1, 2, 3, 4],
        readingRule: 'up-face',
      };
    }

    it('accepts the well-formed pyramid every case here is a mutation of', () => {
      assert.equal(finishSolid(squarePyramid()).faceCount, 4);
    });

    it('rejects an open surface, which the volume guard cannot see', () => {
      const raw = squarePyramid();
      const open: RawSolid = {
        ...raw,
        surface: raw.surface.slice(1),
        readable: [0, 1, 2, 3],
      };
      // Before the half-edge check this returned faceCount 4, capFaces 0 and a
      // tensor of 0.336/0.336/0.400 where the closed pyramid gives
      // 0.35/0.35/0.4: a 4% error, silently.
      assert.throws(() => finishSolid(open), /not closed/);
    });

    it('rejects a face handed in twice', () => {
      const raw = squarePyramid();
      const doubled: RawSolid = { ...raw, surface: [...raw.surface, [0, 1, 2, 3]] };
      assert.throws(() => finishSolid(doubled), /not a manifold/);
    });

    it('rejects a surface too small to enclose anything', () => {
      const raw = squarePyramid();
      assert.throws(
        () => finishSolid({ ...raw, surface: raw.surface.slice(0, 3), readable: [0] }),
        /cannot enclose a volume/,
      );
    });

    it('rejects a loop with an out-of-range vertex index', () => {
      const raw = squarePyramid();
      raw.surface[2] = [4, 1, 9];
      // Used to be an opaque TypeError from inside a vector helper.
      assert.throws(() => finishSolid(raw), /references vertex 9/);
    });

    it('rejects a loop that uses one vertex twice', () => {
      const raw = squarePyramid();
      raw.surface[2] = [4, 1, 1];
      assert.throws(() => finishSolid(raw), /uses vertex 1 twice/);
    });

    it('rejects a loop with fewer than three vertices', () => {
      const raw = squarePyramid();
      raw.surface[2] = [4, 1];
      assert.throws(() => finishSolid(raw), /at least 3/);
    });

    it('rejects a duplicated readable index, which would double-count a face', () => {
      const raw = squarePyramid();
      // Used to return faceCount 5 for a four-faced pyramid, breaking
      // `faceCount === faces.length` from outside the module.
      assert.throws(() => finishSolid({ ...raw, readable: [1, 1, 2, 3] }), /appears twice/);
    });

    it('rejects an out-of-range readable index', () => {
      const raw = squarePyramid();
      assert.throws(
        () => finishSolid({ ...raw, readable: [1, 2, 3, 9] }),
        /readable face 9 is not one of the 5 surface loops/,
      );
      assert.throws(() => finishSolid({ ...raw, readable: [1, 2, 3, -1] }), /readable face -1/);
    });
  });

  describe('invalid face counts', () => {
    for (const faceCount of [2, 1, 0, -6, 3.5, NaN, Infinity]) {
      it(`rejects ${faceCount}`, () => {
        assert.throws(() => dieGeometry(faceCount), /face count/i);
      });
    }
  });
});
