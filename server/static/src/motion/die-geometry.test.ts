import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  dieGeometry,
  finishSolid,
  type DieGeometry,
  type RawSolid,
  type Vec3,
} from './die-geometry.ts';

const sub = (a: Vec3, b: Vec3): Vec3 => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const dot = (a: Vec3, b: Vec3): number => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const mag = (a: Vec3): number => Math.sqrt(dot(a, a));
const cross = (a: Vec3, b: Vec3): Vec3 => [
  a[1] * b[2] - a[2] * b[1],
  a[2] * b[0] - a[0] * b[2],
  a[0] * b[1] - a[1] * b[0],
];

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

      it('reports the circumradius as the farthest vertex distance', () => {
        const farthest = Math.max(...geometry.vertices.map((vertex) => mag(vertex)));
        close(geometry.circumradius, farthest, 1e-12, 'circumradius');
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

      it('uses the up-face reading rule unless it is a d4', () => {
        assert.equal(geometry.readingRule, faceCount === 4 ? 'top-vertex' : 'up-face');
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
      close(cube.circumradius, Math.sqrt(3), 1e-12, 'cube circumradius');
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
      const expected = (octahedron.circumradius * octahedron.circumradius) / 5;
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

      it(`gives the d${faceCount} barrel square side faces`, () => {
        // The documented proportion: half-height sin(pi/N) makes the vertical
        // edge match the ring chord, so every side face is a square. Nothing
        // else pins the barrel's aspect ratio, so assert it directly.
        const geometry = dieGeometry(faceCount);
        const chord = 2 * Math.sin(Math.PI / faceCount);
        for (const [index, face] of geometry.faces.entries()) {
          for (let i = 0; i < face.polygon.length; i++) {
            const edge = mag(sub(face.polygon[(i + 1) % face.polygon.length], face.polygon[i]));
            close(edge, chord, 1e-12, `d${faceCount} side face ${index} edge ${i} is not square`);
          }
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

  describe('invalid face counts', () => {
    for (const faceCount of [2, 1, 0, -6, 3.5, NaN, Infinity]) {
      it(`rejects ${faceCount}`, () => {
        assert.throws(() => dieGeometry(faceCount), /face count/i);
      });
    }
  });
});
