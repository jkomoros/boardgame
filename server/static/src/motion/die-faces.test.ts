import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { dieGeometry, type DieGeometry, type Quat, type Vec3 } from './die-geometry.ts';
import {
  antipodalFacePairs,
  assignFaceValues,
  presentedFaceIndex,
  resolveReadingRule,
  WORLD_UP,
} from './die-faces.ts';

// ---------------------------------------------------------------------------
// Test-local vector/quaternion algebra.
//
// Deliberately written from scratch rather than imported from die-geometry.ts
// or die-faces.ts. `presentedFaceIndex` rotates the up vector BACKWARDS into
// the body frame with an optimised quaternion-vector product; the oracle below
// rotates each face normal FORWARDS with a rotation matrix. Two different
// expressions, so agreement is evidence rather than tautology.
// ---------------------------------------------------------------------------

const sub = (a: Vec3, b: Vec3): Vec3 => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const dot = (a: Vec3, b: Vec3): number => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const mag = (a: Vec3): number => Math.sqrt(dot(a, a));
const cross = (a: Vec3, b: Vec3): Vec3 => [
  a[1] * b[2] - a[2] * b[1],
  a[2] * b[0] - a[0] * b[2],
  a[0] * b[1] - a[1] * b[0],
];

/** A rotation matrix, 9 entries, row-major: `m[row * 3 + col]`. */
type Mat3 = readonly number[];

const matVec = (m: Mat3, v: Vec3): Vec3 => [
  m[0] * v[0] + m[1] * v[1] + m[2] * v[2],
  m[3] * v[0] + m[4] * v[1] + m[5] * v[2],
  m[6] * v[0] + m[7] * v[1] + m[8] * v[2],
];

/** Full quaternion product; the sandwich rotation below is built on it. */
function quatMul(a: Quat, b: Quat): Quat {
  return [
    a[3] * b[0] + a[0] * b[3] + a[1] * b[2] - a[2] * b[1],
    a[3] * b[1] - a[0] * b[2] + a[1] * b[3] + a[2] * b[0],
    a[3] * b[2] + a[0] * b[1] - a[1] * b[0] + a[2] * b[3],
    a[3] * b[3] - a[0] * b[0] - a[1] * b[1] - a[2] * b[2],
  ];
}

/** q * (v, 0) * q^-1 — the textbook definition, not an optimised form. */
function rotateSandwich(q: Quat, v: Vec3): Vec3 {
  const inverse: Quat = [-q[0], -q[1], -q[2], q[3]];
  const out = quatMul(quatMul(q, [v[0], v[1], v[2], 0]), inverse);
  return [out[0], out[1], out[2]];
}

/** Shepperd's method. Validated below against `rotateSandwich` vs `matVec`. */
function quatFromMatrix(m: Mat3): Quat {
  const trace = m[0] + m[4] + m[8];
  if (trace > 0) {
    const s = Math.sqrt(trace + 1) * 2;
    return [(m[7] - m[5]) / s, (m[2] - m[6]) / s, (m[3] - m[1]) / s, 0.25 * s];
  }
  if (m[0] > m[4] && m[0] > m[8]) {
    const s = Math.sqrt(1 + m[0] - m[4] - m[8]) * 2;
    return [0.25 * s, (m[1] + m[3]) / s, (m[2] + m[6]) / s, (m[7] - m[5]) / s];
  }
  if (m[4] > m[8]) {
    const s = Math.sqrt(1 + m[4] - m[0] - m[8]) * 2;
    return [(m[1] + m[3]) / s, 0.25 * s, (m[5] + m[7]) / s, (m[2] - m[6]) / s];
  }
  const s = Math.sqrt(1 + m[8] - m[0] - m[4]) * 2;
  return [(m[2] + m[6]) / s, (m[5] + m[7]) / s, 0.25 * s, (m[3] - m[1]) / s];
}

/**
 * The 24 rotations that map the coordinate axes onto signed coordinate axes:
 * every permutation of the axes crossed with every sign triple, keeping only
 * the 24 with determinant +1 (the other 24 are reflections, not rotations).
 * These are exactly the rotation group of the cube.
 */
function axisAlignedRotations(): { matrix: Mat3; quat: Quat }[] {
  const permutations = [
    [0, 1, 2], [0, 2, 1], [1, 0, 2], [1, 2, 0], [2, 0, 1], [2, 1, 0],
  ];
  const out: { matrix: Mat3; quat: Quat }[] = [];
  for (const p of permutations) {
    for (const sx of [1, -1]) {
      for (const sy of [1, -1]) {
        for (const sz of [1, -1]) {
          const signs = [sx, sy, sz];
          const m = new Array<number>(9).fill(0);
          // Column j of the matrix is signs[j] * e_{p[j]}.
          for (let j = 0; j < 3; j++) m[p[j] * 3 + j] = signs[j];
          const det =
            m[0] * (m[4] * m[8] - m[5] * m[7]) -
            m[1] * (m[3] * m[8] - m[5] * m[6]) +
            m[2] * (m[3] * m[7] - m[4] * m[6]);
          if (det < 0) continue;
          out.push({ matrix: m, quat: quatFromMatrix(m) });
        }
      }
    }
  }
  return out;
}

/** Shortest-arc rotation carrying `from` onto `to`; both must be unit length. */
function shortestArc(from: Vec3, to: Vec3): Quat {
  const d = dot(from, to);
  if (d > 1 - 1e-12) return [0, 0, 0, 1];
  if (d < -1 + 1e-12) {
    // Antiparallel: any perpendicular axis, half turn.
    const seed: Vec3 = Math.abs(from[0]) < 0.9 ? [1, 0, 0] : [0, 1, 0];
    const axis = cross(from, seed);
    const length = mag(axis);
    return [axis[0] / length, axis[1] / length, axis[2] / length, 0];
  }
  const axis = cross(from, to);
  const q: Quat = [axis[0], axis[1], axis[2], 1 + d];
  const length = Math.sqrt(q[0] ** 2 + q[1] ** 2 + q[2] ** 2 + q[3] ** 2);
  return [q[0] / length, q[1] / length, q[2] / length, q[3] / length];
}

/** An orientation in which the die sits flat on face `index`. */
function restingOn(geometry: DieGeometry, index: number): Quat {
  return shortestArc(geometry.faces[index].normal, [0, -1, 0]);
}

/** The oracle: face whose FORWARD-rotated normal is nearest +Y. */
function nearestUpFace(geometry: DieGeometry, matrix: Mat3): number {
  let best = 0;
  let bestScore = -Infinity;
  geometry.faces.forEach((face, index) => {
    const score = dot(matVec(matrix, face.normal), [0, 1, 0]);
    if (score > bestScore) {
      bestScore = score;
      best = index;
    }
  });
  return best;
}

/** Test-local antipodal pairing, computed straight off the normals. */
function oppositeFace(geometry: DieGeometry, index: number): number {
  const target = geometry.faces[index].normal;
  for (let j = 0; j < geometry.faces.length; j++) {
    if (j === index) continue;
    const other = geometry.faces[j].normal;
    if (mag([target[0] + other[0], target[1] + other[1], target[2] + other[2]]) < 1e-9) return j;
  }
  return -1;
}

const sorted = (values: readonly number[]): number[] => [...values].sort((a, b) => a - b);

/**
 * The triple product of the normals carrying the three lowest values. Positive
 * is the Western handedness: a standard d6's 1, 2 and 3 read counter-clockwise
 * about the corner they share.
 */
function handedness(geometry: DieGeometry, values: readonly number[]): number {
  const lowest = sorted(values).slice(0, 3);
  const taken = new Set<number>();
  const normals = lowest.map((value) => {
    const index = values.findIndex((v, i) => v === value && !taken.has(i));
    taken.add(index);
    return geometry.faces[index].normal;
  });
  return dot(normals[0], cross(normals[1], normals[2]));
}

// ---------------------------------------------------------------------------

describe('test harness self-check', () => {
  it('derives quaternions that agree with their rotation matrices', () => {
    // Without this the oracle in the presented-face tests is worthless: it
    // compares a matrix-rotated normal against a quaternion the implementation
    // consumes, so the two must describe the same rotation.
    const probes: Vec3[] = [[1, 0, 0], [0, 1, 0], [0, 0, 1], [0.3, -0.7, 0.2]];
    for (const { matrix, quat } of axisAlignedRotations()) {
      for (const probe of probes) {
        const viaMatrix = matVec(matrix, probe);
        const viaQuat = rotateSandwich(quat, probe);
        assert.ok(
          mag(sub(viaMatrix, viaQuat)) < 1e-12,
          `matrix ${JSON.stringify(matrix)} and quat ${JSON.stringify(quat)} disagree`,
        );
      }
    }
  });

  it('enumerates exactly the 24 rotations of the cube', () => {
    const rotations = axisAlignedRotations();
    assert.equal(rotations.length, 24);
    const keys = new Set(rotations.map((r) => r.matrix.join(',')));
    assert.equal(keys.size, 24);
  });

  it('agrees with die-geometry that every d6/d20 face has an antipode', () => {
    for (const faceCount of [6, 8, 10, 12, 20, 16]) {
      const geometry = dieGeometry(faceCount);
      for (let i = 0; i < faceCount; i++) {
        assert.notEqual(oppositeFace(geometry, i), -1, `d${faceCount} face ${i} has no antipode`);
      }
    }
  });
});

describe('WORLD_UP', () => {
  it('is +Y', () => {
    assert.deepEqual([...WORLD_UP], [0, 1, 0]);
  });
});

describe('resolveReadingRule', () => {
  it('reads the up face on every centrally symmetric solid', () => {
    for (const faceCount of [6, 8, 10, 12, 20, 14, 16]) {
      assert.equal(resolveReadingRule(dieGeometry(faceCount)), 'up-face', `d${faceCount}`);
    }
  });

  it('reads the top vertex on a d4', () => {
    assert.equal(resolveReadingRule(dieGeometry(4)), 'top-vertex');
  });

  it('reads the down face on an odd-sided barrel, which has no up face', () => {
    for (const faceCount of [3, 5, 7, 15]) {
      assert.equal(resolveReadingRule(dieGeometry(faceCount)), 'down-face', `d${faceCount}`);
    }
  });
});

describe('antipodalFacePairs', () => {
  it('pairs every face of a centrally symmetric solid with its opposite', () => {
    for (const faceCount of [6, 8, 10, 12, 20, 16]) {
      const geometry = dieGeometry(faceCount);
      const pairs = antipodalFacePairs(geometry);
      assert.ok(pairs, `d${faceCount} should pair`);
      for (let i = 0; i < faceCount; i++) {
        assert.equal(pairs[i], oppositeFace(geometry, i), `d${faceCount} face ${i}`);
        assert.equal(pairs[pairs[i]], i, `d${faceCount} pairing is not an involution at ${i}`);
      }
    }
  });

  it('returns null when some face has no opposite', () => {
    for (const faceCount of [4, 3, 7, 15]) {
      assert.equal(antipodalFacePairs(dieGeometry(faceCount)), null, `d${faceCount}`);
    }
  });
});

describe('presentedFaceIndex', () => {
  it('returns the face nearest +Y for all 24 axis-aligned cube orientations', () => {
    const cube = dieGeometry(6);
    const rotations = axisAlignedRotations();
    assert.equal(rotations.length, 24);
    for (const { matrix, quat } of rotations) {
      const scores = cube.faces
        .map((face) => dot(matVec(matrix, face.normal), [0, 1, 0]))
        .sort((a, b) => b - a);
      // The answer must be unambiguous or the assertion below means nothing.
      assert.ok(scores[0] - scores[1] > 0.5, `ambiguous up face: ${JSON.stringify(scores)}`);
      assert.equal(
        presentedFaceIndex(cube, quat),
        nearestUpFace(cube, matrix),
        `orientation ${JSON.stringify(quat)}`,
      );
    }
  });

  it('presents the opposite face when a symmetric die rests on a face', () => {
    for (const faceCount of [6, 8, 10, 12, 20, 14, 16]) {
      const geometry = dieGeometry(faceCount);
      for (let down = 0; down < faceCount; down++) {
        assert.equal(
          presentedFaceIndex(geometry, restingOn(geometry, down)),
          oppositeFace(geometry, down),
          `d${faceCount} resting on face ${down}`,
        );
      }
    }
  });

  it('tolerates a die that lands a few degrees off flat', () => {
    const geometry = dieGeometry(20);
    for (let down = 0; down < 20; down++) {
      const flat = restingOn(geometry, down);
      // Tip it 4 degrees about +X, the settling tolerance task 6 works to.
      const half = (4 * Math.PI) / 360;
      const tilt: Quat = [Math.sin(half), 0, 0, Math.cos(half)];
      assert.equal(
        presentedFaceIndex(geometry, quatMul(tilt, flat)),
        oppositeFace(geometry, down),
        `d20 tilted off face ${down}`,
      );
    }
  });

  it('accepts an unnormalised quaternion', () => {
    // A fixed-step integrator's orientation drifts off the unit sphere, so this
    // is the shape of input task 6 will actually hand over. The rotation
    // formula is quadratic in the quaternion's length, so on a die with fine
    // angular resolution an unrenormalised quaternion reads the wrong face.
    for (const faceCount of [12, 20]) {
      const geometry = dieGeometry(faceCount);
      for (let down = 0; down < faceCount; down++) {
        const flat = restingOn(geometry, down);
        for (const degrees of [0, 3, 11]) {
          const half = (degrees * Math.PI) / 360;
          const tilt: Quat = [Math.sin(half), 0, 0, Math.cos(half)];
          const quat = quatMul(tilt, flat);
          const expected = presentedFaceIndex(geometry, quat);
          for (const factor of [0.5, 2.5, 7]) {
            const scaled = [
              quat[0] * factor, quat[1] * factor, quat[2] * factor, quat[3] * factor,
            ] as unknown as Quat;
            assert.equal(
              presentedFaceIndex(geometry, scaled),
              expected,
              `d${faceCount} face ${down} tilted ${degrees}deg, scaled by ${factor}`,
            );
          }
        }
      }
    }
  });

  it('refuses a zero quaternion rather than reading NaN', () => {
    assert.throws(() => presentedFaceIndex(dieGeometry(6), [0, 0, 0, 0]), /zero-length/);
  });

  describe('odd-sided barrels', () => {
    it('has a genuinely ambiguous up face, which is why it is read from below', () => {
      // This is the empirical fact that motivates the down-face rule. If it
      // ever stops holding, the rule should be revisited rather than kept.
      for (const faceCount of [3, 7, 15]) {
        const geometry = dieGeometry(faceCount);
        const orientation = restingOn(geometry, 0);
        const up = geometry.faces
          .map((face) => dot(rotateSandwich(orientation, face.normal), [0, 1, 0]))
          .sort((a, b) => b - a);
        assert.ok(
          up[0] - up[1] < 1e-9,
          `d${faceCount} up face is not a tie after all: ${up[0]} vs ${up[1]}`,
        );
      }
    });

    it('presents the face it rests on', () => {
      for (const faceCount of [3, 5, 7, 15]) {
        const geometry = dieGeometry(faceCount);
        for (let down = 0; down < faceCount; down++) {
          assert.equal(
            presentedFaceIndex(geometry, restingOn(geometry, down)),
            down,
            `d${faceCount} resting on face ${down}`,
          );
        }
      }
    });

    it('picks the down face by a clear margin, not by floating-point noise', () => {
      for (const faceCount of [3, 5, 7, 15]) {
        const geometry = dieGeometry(faceCount);
        const orientation = restingOn(geometry, 0);
        const down = geometry.faces
          .map((face) => dot(rotateSandwich(orientation, face.normal), [0, 1, 0]))
          .sort((a, b) => a - b);
        assert.ok(down[1] - down[0] > 1e-3, `d${faceCount} down face is a tie too: ${down}`);
      }
    });
  });

  describe('d4 top-vertex rule', () => {
    it('has no usable up face: the three visible faces tie', () => {
      const geometry = dieGeometry(4);
      const orientation = restingOn(geometry, 0);
      const up = geometry.faces
        .map((face) => dot(rotateSandwich(orientation, face.normal), [0, 1, 0]))
        .sort((a, b) => b - a);
      assert.ok(
        up[0] - up[2] < 1e-9,
        `d4's three upward faces should tie, got ${JSON.stringify(up)}`,
      );
    });

    it('resolves to the face opposite the single highest vertex', () => {
      const geometry = dieGeometry(4);
      for (let down = 0; down < 4; down++) {
        const orientation = restingOn(geometry, down);
        const heights = geometry.vertices.map((v) => rotateSandwich(orientation, v)[1]);
        const top = heights.indexOf(Math.max(...heights));
        // Exactly one vertex is strictly highest.
        assert.equal(
          heights.filter((h) => h > heights[top] - 1e-9).length,
          1,
          `d4 on face ${down} has no unique top vertex`,
        );
        const apex = geometry.vertices[top];
        // The face opposite that vertex is the one whose polygon does not
        // include it. Deliberately a vertex-membership test rather than the
        // plane-distance test the implementation uses: a sign error in that
        // distance is exactly the bug this suite caught, so the oracle must
        // not be written the same way.
        const missing = geometry.faces
          .map((face, index) => ({ index, face }))
          .filter(({ face }) => !face.polygon.some((p) => mag(sub(p, apex)) < 1e-9));
        assert.equal(missing.length, 1, `d4 on face ${down}: apex should miss exactly one face`);
        assert.equal(
          missing[0].index,
          down,
          `d4 on face ${down}: the face the apex misses should be the one it rests on`,
        );
        assert.equal(presentedFaceIndex(geometry, orientation), down);
      }
    });
  });
});

// ---------------------------------------------------------------------------

/**
 * The canonical Western d6: 1 opposite 6, 2 opposite 5, 3 opposite 4, and
 * 1, 2, 3 counter-clockwise about the corner they share (equivalently, their
 * normals form a right-handed frame).
 */
function canonicalCubeValue(normal: Vec3): number {
  const axes: { direction: Vec3; value: number }[] = [
    { direction: [1, 0, 0], value: 1 },
    { direction: [0, 1, 0], value: 2 },
    { direction: [0, 0, 1], value: 3 },
    { direction: [0, 0, -1], value: 4 },
    { direction: [0, -1, 0], value: 5 },
    { direction: [-1, 0, 0], value: 6 },
  ];
  const unit = normal.map((c) => c / mag(normal)) as unknown as Vec3;
  for (const axis of axes) {
    if (dot(unit, axis.direction) > 0.99) return axis.value;
  }
  throw new Error(`not an axis-aligned cube normal: ${JSON.stringify(normal)}`);
}

/**
 * Independent oracle for "is this a valid standard d6?".
 *
 * Not a restatement of the sum rule or the chirality rule: it searches the
 * cube's 24 rotations for one that carries the canonical Western die exactly
 * onto `values`. An arrangement that is a rotation of the standard die is the
 * standard die, and nothing else passes.
 */
function isStandardCubeLabelling(geometry: DieGeometry, values: readonly number[]): boolean {
  return axisAlignedRotations().some(({ matrix }) =>
    geometry.faces.every(
      (face, index) => canonicalCubeValue(matVec(matrix, face.normal)) === values[index],
    ),
  );
}

describe('assignFaceValues', () => {
  describe('d6', () => {
    const cube = dieGeometry(6);
    const pips = [1, 2, 3, 4, 5, 6];

    it('solves all 36 (presented, desired) pairs as a real standard die', () => {
      // Exhaustive on purpose: "always solvable" is the whole reason the
      // outcome is painted onto the die instead of re-rolled for.
      for (let presented = 0; presented < 6; presented++) {
        for (const desired of pips) {
          const values = assignFaceValues(cube, pips, presented, desired);
          const label = `presented ${presented}, desired ${desired}`;

          assert.deepEqual(sorted(values), pips, `${label}: not a permutation`);
          assert.equal(values[presented], desired, `${label}: wrong face presented`);

          for (let i = 0; i < 6; i++) {
            assert.equal(
              values[i] + values[oppositeFace(cube, i)],
              7,
              `${label}: faces ${i} and ${oppositeFace(cube, i)} do not sum to 7`,
            );
          }

          assert.ok(
            isStandardCubeLabelling(cube, values),
            `${label}: ${JSON.stringify(values)} is not a rotation of a standard die`,
          );
        }
      }
    });

    it('rejects a mirrored die, so the standard-die oracle has teeth', () => {
      // Swap 2 and 5: sums still 7, chirality inverted.
      const values = [...assignFaceValues(cube, pips, 0, 1)];
      const two = values.indexOf(2);
      const five = values.indexOf(5);
      values[two] = 5;
      values[five] = 2;
      assert.ok(!isStandardCubeLabelling(cube, values), 'a mirrored die must not pass');
    });

    it('honours the min+max sum rule for a non-1-based value set', () => {
      const faces = [0, 1, 2, 3, 4, 5];
      for (let presented = 0; presented < 6; presented++) {
        for (const desired of faces) {
          const values = assignFaceValues(cube, faces, presented, desired);
          assert.deepEqual(sorted(values), faces);
          assert.equal(values[presented], desired);
          for (let i = 0; i < 6; i++) {
            assert.equal(values[i] + values[oppositeFace(cube, i)], 5);
          }
        }
      }
    });

    it('still produces a bijection when the values cannot be paired', () => {
      // 2+13 = 15 but 3+11 = 14: no pairing sums to min+max, so there is no
      // standard arrangement to build and the plain fallback must be used.
      const faces = [2, 3, 5, 7, 11, 13];
      for (let presented = 0; presented < 6; presented++) {
        for (const desired of faces) {
          const values = assignFaceValues(cube, faces, presented, desired);
          assert.deepEqual(sorted(values), sorted(faces));
          assert.equal(values[presented], desired);
          // The documented fallback: the remaining values keep their given
          // order across the remaining faces. Pinning this is what distinguishes
          // "recognised that the values do not pair" from "paired them anyway
          // and got away with it because any bijection is still a bijection".
          const remaining = faces.filter((_, i) => i !== faces.indexOf(desired));
          const placed = values.filter((_, i) => i !== presented);
          assert.deepEqual(placed, remaining, `presented ${presented}, desired ${desired}`);
        }
      }
    });

    it('is deterministic', () => {
      assert.deepEqual(
        [...assignFaceValues(cube, pips, 3, 5)],
        [...assignFaceValues(cube, pips, 3, 5)],
      );
    });

    it('returns a frozen array', () => {
      assert.ok(Object.isFrozen(assignFaceValues(cube, pips, 0, 1)));
    });
  });

  describe('d20', () => {
    const d20 = dieGeometry(20);
    const pips = Array.from({ length: 20 }, (_, i) => i + 1);

    it('reaches every value from every presented face', () => {
      const reached = new Set<string>();
      for (let presented = 0; presented < 20; presented++) {
        for (const desired of pips) {
          const values = assignFaceValues(d20, pips, presented, desired);
          const label = `presented ${presented}, desired ${desired}`;
          assert.deepEqual(sorted(values), pips, `${label}: not a permutation`);
          assert.equal(values[presented], desired, `${label}: wrong face presented`);
          for (let i = 0; i < 20; i++) {
            assert.equal(
              values[i] + values[oppositeFace(d20, i)],
              21,
              `${label}: faces ${i} and ${oppositeFace(d20, i)} do not sum to 21`,
            );
          }
          reached.add(`${presented}:${desired}`);
        }
      }
      assert.equal(reached.size, 400);
    });
  });

  it('keeps a right-handed low-value frame on every solid that has one', () => {
    // Only the d6 has a convention to violate, but the chirality correction is
    // written for any antipodally-paired solid, so it is asserted on all of
    // them: a fix that happens to work for three value pairs and silently
    // no-ops for ten is the failure mode this catches.
    for (const faceCount of [6, 8, 10, 12, 20]) {
      const geometry = dieGeometry(faceCount);
      const pips = Array.from({ length: faceCount }, (_, i) => i + 1);
      for (let presented = 0; presented < faceCount; presented++) {
        for (const desired of pips) {
          const values = assignFaceValues(geometry, pips, presented, desired);
          assert.ok(
            handedness(geometry, values) > 1e-9,
            `d${faceCount} presented ${presented} desired ${desired}: left-handed`,
          );
        }
      }
    }
  });

  describe('shapes with no antipodal pairing', () => {
    it('assigns a d4 bijectively from every presented face', () => {
      const geometry = dieGeometry(4);
      const pips = [1, 2, 3, 4];
      for (let presented = 0; presented < 4; presented++) {
        for (const desired of pips) {
          const values = assignFaceValues(geometry, pips, presented, desired);
          assert.deepEqual(sorted(values), pips);
          assert.equal(values[presented], desired);
        }
      }
    });

    it('assigns an odd barrel bijectively from every presented face', () => {
      const geometry = dieGeometry(7);
      const pips = [1, 2, 3, 4, 5, 6, 7];
      for (let presented = 0; presented < 7; presented++) {
        for (const desired of pips) {
          const values = assignFaceValues(geometry, pips, presented, desired);
          assert.deepEqual(sorted(values), pips);
          assert.equal(values[presented], desired);
        }
      }
    });
  });

  describe('duplicate face values', () => {
    // Face values are a MULTISET everywhere in this module, deliberately: a
    // game-defined face enum may legally repeat a value (three blanks, two
    // skulls). Nothing above exercises that, so these lock in the behaviour
    // before the enum lands. `sorted` compares multisets, so a "permutation"
    // assertion here is genuinely a multiset assertion.
    const cube = dieGeometry(6);

    /** Every distinct value, so `desired` covers the choices without repeats. */
    const distinct = (values: readonly number[]): number[] => [...new Set(values)];

    const cases: { label: string; faces: number[]; sum: number | null }[] = [
      // Pairs up as (1,3),(1,3),(2,2): the standard arrangement is reachable
      // even though two value pairs are identical.
      { label: 'two of each, complementary', faces: [1, 1, 2, 2, 3, 3], sum: 4 },
      // Three copies of one pair. Every pair is (1,6), so the "which pair is
      // forced" search has three equally good answers.
      { label: 'three lows and three highs', faces: [1, 1, 1, 6, 6, 6], sum: 7 },
      // The degenerate multiset. min+max is 0 and every face satisfies it, so
      // the only real content is that a bijection still comes back.
      { label: 'all identical', faces: [0, 0, 0, 0, 0, 0], sum: 0 },
    ];

    for (const { label, faces, sum } of cases) {
      it(`is a bijection of the multiset and presents the value: ${label}`, () => {
        for (let presented = 0; presented < 6; presented++) {
          for (const desired of distinct(faces)) {
            const values = assignFaceValues(cube, faces, presented, desired);
            const where = `${label}: presented ${presented}, desired ${desired}`;
            assert.deepEqual(
              sorted(values),
              sorted(faces),
              `${where}: ${JSON.stringify(values)} is not a permutation of the multiset`,
            );
            assert.equal(values[presented], desired, `${where}: wrong face presented`);
          }
        }
      });

      if (sum !== null) {
        it(`keeps opposite faces summing to min+max: ${label}`, () => {
          for (let presented = 0; presented < 6; presented++) {
            for (const desired of distinct(faces)) {
              const values = assignFaceValues(cube, faces, presented, desired);
              for (let i = 0; i < 6; i++) {
                assert.equal(
                  values[i] + values[oppositeFace(cube, i)],
                  sum,
                  `${label}: presented ${presented}, desired ${desired}: faces ${i} and ` +
                    `${oppositeFace(cube, i)} do not sum to ${sum}`,
                );
              }
            }
          }
        });
      }
    }

    it('places every copy of a repeated value, not just the first', () => {
      // The multiset guarantee is exactly what a Set-based implementation
      // would lose: [1,1,2,2,3,3] must come back with two 1s, not one.
      const faces = [1, 1, 2, 2, 3, 3];
      const values = assignFaceValues(cube, faces, 0, 2);
      const count = (value: number) => values.filter((v) => v === value).length;
      assert.equal(count(1), 2, 'lost a copy of 1');
      assert.equal(count(2), 2, 'lost a copy of 2');
      assert.equal(count(3), 2, 'lost a copy of 3');
    });
  });

  describe('end to end', () => {
    it('shows the desired value on the face the roll actually presented', () => {
      // The whole point of the module, stated once without any helper: land
      // the die, ask which face shows, paint the value, read it back.
      for (const faceCount of [4, 6, 8, 10, 12, 20, 7, 16]) {
        const geometry = dieGeometry(faceCount);
        const pips = Array.from({ length: faceCount }, (_, i) => i + 1);
        for (let down = 0; down < faceCount; down++) {
          const orientation = restingOn(geometry, down);
          const presented = presentedFaceIndex(geometry, orientation);
          for (const desired of pips) {
            const values = assignFaceValues(geometry, pips, presented, desired);
            assert.equal(
              values[presentedFaceIndex(geometry, orientation)],
              desired,
              `d${faceCount} on face ${down} cannot show ${desired}`,
            );
          }
        }
      }
    });
  });

  describe('rejections', () => {
    const cube = dieGeometry(6);

    it('rejects a value list that does not match the face count', () => {
      assert.throws(() => assignFaceValues(cube, [1, 2, 3], 0, 1), /face count/);
    });

    it('rejects a presented index out of range', () => {
      assert.throws(() => assignFaceValues(cube, [1, 2, 3, 4, 5, 6], 6, 1), /presented face/);
      assert.throws(() => assignFaceValues(cube, [1, 2, 3, 4, 5, 6], -1, 1), /presented face/);
      assert.throws(() => assignFaceValues(cube, [1, 2, 3, 4, 5, 6], 1.5, 1), /presented face/);
    });

    it('rejects a desired value that is not one of the face values', () => {
      assert.throws(() => assignFaceValues(cube, [1, 2, 3, 4, 5, 6], 0, 9), /desired value/);
    });
  });
});
