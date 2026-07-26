import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { dieGeometry, type DieGeometry, type Vec3 } from './die-geometry.ts';

const sub = (a: Vec3, b: Vec3): Vec3 => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const dot = (a: Vec3, b: Vec3): number => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const mag = (a: Vec3): number => Math.sqrt(dot(a, a));

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
 * Rebuild a barrel's cap planes from its vertices. A barrel's readable faces
 * are its side faces only, so the caps have to be reconstructed here before the
 * grid oracle can tell inside from outside.
 */
function barrelCapPlanes(geometry: DieGeometry): Plane[] {
  const byHeight = [...geometry.vertices].sort((a, b) => a[2] - b[2]);
  const apexes = [byHeight[0], byHeight[byHeight.length - 1]];
  const planes: Plane[] = [];
  for (const apex of apexes) {
    const ring = geometry.vertices
      .filter(
        (vertex) =>
          vertex !== apex &&
          Math.sign(vertex[2]) === Math.sign(apex[2]) &&
          Math.abs(vertex[2]) < Math.abs(apex[2]),
      )
      .sort((a, b) => Math.atan2(a[1], a[0]) - Math.atan2(b[1], b[0]));
    assert.equal(ring.length, geometry.faceCount, 'barrel cap ring is the wrong size');
    for (let i = 0; i < ring.length; i++) {
      const a = ring[i];
      const b = ring[(i + 1) % ring.length];
      const edge1 = sub(a, apex);
      const edge2 = sub(b, apex);
      const raw: Vec3 = [
        edge1[1] * edge2[2] - edge1[2] * edge2[1],
        edge1[2] * edge2[0] - edge1[0] * edge2[2],
        edge1[0] * edge2[1] - edge1[1] * edge2[0],
      ];
      const length = mag(raw);
      const unit: Vec3 = [raw[0] / length, raw[1] / length, raw[2] / length];
      const point: Vec3 = [
        (apex[0] + a[0] + b[0]) / 3,
        (apex[1] + a[1] + b[1]) / 3,
        (apex[2] + a[2] + b[2]) / 3,
      ];
      const outward: Vec3 =
        dot(unit, point) > 0 ? unit : [-unit[0], -unit[1], -unit[2]];
      planes.push({ normal: outward, point });
    }
  }
  return planes;
}

/**
 * Independent inertia oracle: brute-force grid quadrature over the solid's
 * interior, using half-space tests for point-in-solid. Deliberately shares no
 * code with the tetrahedron decomposition under test.
 */
function gridInertia(geometry: DieGeometry, steps: number, extraPlanes: Plane[] = []): number[] {
  let min: Vec3 = [Infinity, Infinity, Infinity];
  let max: Vec3 = [-Infinity, -Infinity, -Infinity];
  for (const vertex of geometry.vertices) {
    min = [Math.min(min[0], vertex[0]), Math.min(min[1], vertex[1]), Math.min(min[2], vertex[2])];
    max = [Math.max(max[0], vertex[0]), Math.max(max[1], vertex[1]), Math.max(max[2], vertex[2])];
  }
  const span: Vec3 = [max[0] - min[0], max[1] - min[1], max[2] - min[2]];
  const planes: Plane[] = [
    ...geometry.faces.map((face) => ({ normal: face.normal, point: face.centroid })),
    ...extraPlanes,
  ];
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

      it('is centred on its centroid', () => {
        // Vertex mean is not the volume centroid in general, but every solid we
        // generate is symmetric enough that the origin must be inside it.
        for (const face of geometry.faces) {
          assert.ok(
            dot(face.normal, face.centroid) > 0,
            'the origin is not strictly inside the solid',
          );
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
    // The observed disagreement at 90 steps is under 0.07% of the trace for
    // every shape here, and it shrinks as the grid is refined, so 0.5% is a
    // real constraint rather than a rubber stamp.
    const GRID_STEPS = 90;
    const GRID_TOLERANCE_FRACTION = 0.005;

    for (const faceCount of [6, 10, 12, 20, ...BARREL_FACE_COUNTS]) {
      it(`matches an independent numeric integration for the d${faceCount}`, () => {
        const geometry = dieGeometry(faceCount);
        const extraPlanes = BARREL_FACE_COUNTS.includes(faceCount)
          ? barrelCapPlanes(geometry)
          : [];
        const numeric = gridInertia(geometry, GRID_STEPS, extraPlanes);
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

  describe('invalid face counts', () => {
    for (const faceCount of [2, 1, 0, -6, 3.5, NaN, Infinity]) {
      it(`rejects ${faceCount}`, () => {
        assert.throws(() => dieGeometry(faceCount), /face count/i);
      });
    }
  });
});
