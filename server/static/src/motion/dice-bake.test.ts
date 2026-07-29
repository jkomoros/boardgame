import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { dieGeometry, QUAT_IDENTITY, type Quat, type Vec3 } from './die-geometry.ts';
import { simulateRoll, type DieSample, type DieTrajectory } from './dice-sim.ts';
import { restingTransform, trajectoryCurve } from './dice-bake.ts';

// ---------------------------------------------------------------------------
// Helpers. Written out here rather than imported, so every expectation below is
// checked against arithmetic this file owns: a convention shared between the
// module and its test would otherwise be unfalsifiable.
// ---------------------------------------------------------------------------

const ORIGIN: Vec3 = [0, 0, 0];

function sample(t: number, position: Vec3, orientation: Quat): DieSample {
  return { t, position, orientation };
}

function trajectory(samples: readonly DieSample[]): DieTrajectory {
  return { samples, restingOrientation: samples[samples.length - 1].orientation };
}

function axisAngle(axis: Vec3, radians: number): Quat {
  const length = Math.sqrt(axis[0] * axis[0] + axis[1] * axis[1] + axis[2] * axis[2]);
  const s = Math.sin(radians / 2);
  return [(axis[0] / length) * s, (axis[1] / length) * s, (axis[2] / length) * s, Math.cos(radians / 2)];
}

/** Body-to-world rotation of a unit quaternion, row-major. */
function rotationMatrix(q: Quat): number[] {
  const length = Math.sqrt(q[0] * q[0] + q[1] * q[1] + q[2] * q[2] + q[3] * q[3]);
  const [x, y, z, w] = [q[0] / length, q[1] / length, q[2] / length, q[3] / length];
  return [
    1 - 2 * (y * y + z * z), 2 * (x * y - z * w), 2 * (x * z + y * w),
    2 * (x * y + z * w), 1 - 2 * (x * x + z * z), 2 * (y * z - x * w),
    2 * (x * z - y * w), 2 * (y * z + x * w), 1 - 2 * (x * x + y * y),
  ];
}

/** The smallest rotation angle, in radians, carrying `a` onto `b`. */
function quatAngle(a: Quat, b: Quat): number {
  const na = Math.sqrt(a[0] * a[0] + a[1] * a[1] + a[2] * a[2] + a[3] * a[3]);
  const nb = Math.sqrt(b[0] * b[0] + b[1] * b[1] + b[2] * b[2] + b[3] * b[3]);
  const cosine = Math.abs(
    (a[0] * b[0] + a[1] * b[1] + a[2] * b[2] + a[3] * b[3]) / (na * nb),
  );
  return 2 * Math.acos(Math.min(1, cosine));
}

/**
 * Parse an emitted transform, asserting on the way through that it is a literal
 * `matrix3d` of sixteen plain decimal numbers.
 *
 * The number pattern is deliberately strict: it rejects `var(...)`, `calc(...)`,
 * units, and exponential notation, all of which are things a naive formatter
 * emits and a compositor either refuses or refuses to accelerate.
 */
function parseMatrix3d(value: string): number[] {
  const match = /^matrix3d\(([^()]*)\)$/.exec(value);
  assert.ok(match, `expected a literal matrix3d value, got ${value}`);
  const parts = match[1].split(',').map(part => part.trim());
  assert.equal(parts.length, 16, `expected 16 components in ${value}`);
  return parts.map(part => {
    assert.match(part, /^-?\d+(\.\d+)?$/, `component ${part} is not a plain decimal number`);
    const parsed = Number(part);
    assert.ok(Number.isFinite(parsed), `component ${part} is not finite`);
    return parsed;
  });
}

/**
 * The transform the module is expected to emit, built independently.
 *
 * The physics frame is right-handed with +Y up; CSS screen-Y points DOWN. The
 * change of basis is the reflection S = diag(1, -1, 1): a point maps to `S p`
 * and an orientation to the similarity `S R S`, which is `s_i s_j R_ij`.
 */
function expectedMatrix(position: Vec3, orientation: Quat, radiusPx = 1): number[] {
  const r = rotationMatrix(orientation);
  const sign = [1, -1, 1];
  const m = r.map((value, index) => sign[Math.floor(index / 3)] * sign[index % 3] * value);
  return [
    m[0], m[3], m[6], 0,
    m[1], m[4], m[7], 0,
    m[2], m[5], m[8], 0,
    position[0] * radiusPx, -position[1] * radiusPx, position[2] * radiusPx, 1,
  ];
}

function assertMatrixClose(actual: number[], expected: number[], tolerance: number, label: string): void {
  for (let i = 0; i < 16; i++) {
    assert.ok(
      Math.abs(actual[i] - expected[i]) <= tolerance,
      `${label}: component ${i} was ${actual[i]}, expected ${expected[i]}`,
    );
  }
}

const FALLING = trajectory([
  sample(0, [0, 2, 0], QUAT_IDENTITY),
  sample(50, [1, 0, -1], axisAngle([0, 0, 1], 0.6)),
  sample(100, [2, -2, -2], axisAngle([0, 0, 1], 1.2)),
]);

// ---------------------------------------------------------------------------

describe('trajectoryCurve endpoints', () => {
  it('reproduces the first and last samples at progress 0 and 1', () => {
    const curve = trajectoryCurve(FALLING, 100);
    const first = FALLING.samples[0];
    const last = FALLING.samples[FALLING.samples.length - 1];
    assertMatrixClose(parseMatrix3d(curve(0)), expectedMatrix(first.position, first.orientation), 1e-6, 'progress 0');
    assertMatrixClose(parseMatrix3d(curve(1)), expectedMatrix(last.position, last.orientation), 1e-6, 'progress 1');
  });

  it('holds the resting pose when the animation outlasts the trajectory', () => {
    const curve = trajectoryCurve(FALLING, 200);
    assert.equal(curve(1), restingTransform(FALLING));
    assert.equal(curve(0.75), restingTransform(FALLING));
  });

  it('clamps progress outside [0, 1] rather than extrapolating', () => {
    const curve = trajectoryCurve(FALLING, 100);
    assert.equal(curve(-0.25), curve(0));
    assert.equal(curve(4), curve(1));
  });
});

describe('trajectoryCurve output form', () => {
  it('emits only literal matrix3d values, never var() or calc()', () => {
    const roll = simulateRoll({ seed: 7, geometry: dieGeometry(6), dieCount: 1, bounds: { x: 4, y: 4, z: 4 } });
    const curve = trajectoryCurve(roll.dice[0], roll.durationMs, { radiusPx: 24 });
    for (let i = 0; i < 256; i++) {
      const value = curve(i / 255);
      assert.ok(!value.includes('var('), `sample ${i} used a custom property: ${value}`);
      assert.ok(!value.includes('calc('), `sample ${i} used calc(): ${value}`);
      const components = parseMatrix3d(value);
      assert.equal(components[15], 1);
      assert.equal(components[3], 0);
      assert.equal(components[7], 0);
      assert.equal(components[11], 0);
    }
  });

  it('emits a rotation block that stays orthonormal along a real roll', () => {
    const roll = simulateRoll({ seed: 11, geometry: dieGeometry(20), dieCount: 1, bounds: { x: 4, y: 4, z: 4 } });
    const curve = trajectoryCurve(roll.dice[0], roll.durationMs);
    for (let i = 0; i < 64; i++) {
      const m = parseMatrix3d(curve(i / 63));
      for (let column = 0; column < 3; column++) {
        const c = [m[column * 4], m[column * 4 + 1], m[column * 4 + 2]];
        const length = Math.sqrt(c[0] * c[0] + c[1] * c[1] + c[2] * c[2]);
        assert.ok(Math.abs(length - 1) < 1e-5, `column ${column} of sample ${i} had length ${length}`);
      }
    }
  });
});

describe('the physics-to-CSS Y flip', () => {
  /**
   * The anchor test for the whole conversion, stated in the CSS frame so that no
   * round trip through the module's own convention can satisfy it. The simulator
   * puts +Y up; CSS puts +Y down. A die the physics says is FALLING must
   * therefore translate DOWN the screen, i.e. its matrix3d translation-Y — the
   * fourteenth component — must increase with progress.
   */
  it('turns a physically downward fall into an increasing screen-Y translation', () => {
    const falling = trajectory([
      sample(0, [0, 3, 0], QUAT_IDENTITY),
      sample(100, [0, -3, 0], QUAT_IDENTITY),
    ]);
    const curve = trajectoryCurve(falling, 100, { radiusPx: 10 });
    const start = parseMatrix3d(curve(0))[13];
    const middle = parseMatrix3d(curve(0.5))[13];
    const end = parseMatrix3d(curve(1))[13];
    assert.ok(start < middle && middle < end, `screen Y went ${start} -> ${middle} -> ${end}`);
    assert.equal(start, -30);
    assert.equal(end, 30);
  });

  /**
   * The same flip for orientation, and the reason it cannot be a bare sign flip
   * on the translation. A physics rotation of +90 degrees about +Z carries the
   * body's +X axis onto world +Y, which is UP. Up on screen is -Y, so the image
   * of the body's X axis — the matrix's first column — must be (0, -1, 0).
   * Without the flip it comes out (0, +1, 0) and the die tumbles the wrong way.
   */
  it('turns a rotation that points a body axis upward into a rotation that points it up the screen', () => {
    const upright = trajectory([sample(0, ORIGIN, axisAngle([0, 0, 1], Math.PI / 2))]);
    const m = parseMatrix3d(trajectoryCurve(upright, 100)(0));
    assert.ok(Math.abs(m[0] - 0) < 1e-6, `x column x was ${m[0]}`);
    assert.ok(Math.abs(m[1] + 1) < 1e-6, `x column y was ${m[1]}, expected -1 (up the screen)`);
    assert.ok(Math.abs(m[2] - 0) < 1e-6, `x column z was ${m[2]}`);
  });

  it('leaves X and Z alone', () => {
    const drift = trajectory([sample(0, [1.5, 0, -2.5], QUAT_IDENTITY)]);
    const m = parseMatrix3d(trajectoryCurve(drift, 100, { radiusPx: 4 })(0));
    assert.equal(m[12], 6);
    assert.equal(m[13], 0);
    assert.equal(m[14], -10);
  });

  it('emits a proper rotation, not a mirrored one', () => {
    // A reflection applied to only half of the pose would flip handedness, which
    // shows up as a determinant of -1 and renders a mirror-image die.
    const tilted = trajectory([sample(0, ORIGIN, axisAngle([1, 2, -3], 1.1))]);
    const m = parseMatrix3d(trajectoryCurve(tilted, 100)(0));
    const r = [m[0], m[4], m[8], m[1], m[5], m[9], m[2], m[6], m[10]];
    const determinant =
      r[0] * (r[4] * r[8] - r[5] * r[7]) -
      r[1] * (r[3] * r[8] - r[5] * r[6]) +
      r[2] * (r[3] * r[7] - r[4] * r[6]);
    assert.ok(Math.abs(determinant - 1) < 1e-6, `determinant was ${determinant}`);
  });
});

describe('trajectoryCurve interpolation', () => {
  it('lerps position and slerps orientation between samples', () => {
    const die = trajectory([
      sample(0, ORIGIN, QUAT_IDENTITY),
      sample(100, [2, -4, 6], axisAngle([1, 0, 0], Math.PI / 2)),
    ]);
    const m = parseMatrix3d(trajectoryCurve(die, 100, { radiusPx: 3 })(0.5));
    assertMatrixClose(m, expectedMatrix([1, -2, 3], axisAngle([1, 0, 0], Math.PI / 4), 3), 1e-6, 'midpoint');
  });

  it('interpolates at a resolution the samples do not line up with', () => {
    // 64 uniform samples over a 100 ms trajectory whose own steps are 30 ms:
    // every interior grid point falls strictly inside a segment.
    const die = trajectory([
      sample(0, [0, 0, 0], QUAT_IDENTITY),
      sample(30, [3, 0, 0], QUAT_IDENTITY),
      sample(60, [6, 0, 0], QUAT_IDENTITY),
      sample(90, [9, 0, 0], QUAT_IDENTITY),
    ]);
    const curve = trajectoryCurve(die, 90);
    for (let i = 0; i < 64; i++) {
      const progress = i / 63;
      const x = parseMatrix3d(curve(progress))[12];
      assert.ok(Math.abs(x - progress * 9) < 1e-5, `x at ${progress} was ${x}`);
    }
  });

  /**
   * `q` and `-q` are the same orientation but opposite slerp paths. A bake that
   * ignores the sign takes the 5.88 rad way round here instead of the 0.4 rad
   * way, which reads as a die snapping backwards between two adjacent frames.
   */
  it('takes the short arc when adjacent samples have opposite quaternion signs', () => {
    const turn = axisAngle([0, 0, 1], 0.4);
    const flipped: Quat = [-turn[0], -turn[1], -turn[2], -turn[3]];
    const die = trajectory([sample(0, ORIGIN, QUAT_IDENTITY), sample(100, ORIGIN, flipped)]);
    const curve = trajectoryCurve(die, 100);
    assertMatrixClose(
      parseMatrix3d(curve(0.5)),
      expectedMatrix(ORIGIN, axisAngle([0, 0, 1], 0.2)),
      1e-6,
      'short-arc midpoint',
    );
    // And the endpoint is still the sample it was handed, sign notwithstanding.
    assertMatrixClose(parseMatrix3d(curve(1)), expectedMatrix(ORIGIN, turn), 1e-6, 'flipped endpoint');
  });

  it('keeps sign handling working across a chain of alternating samples', () => {
    const samples: DieSample[] = [];
    for (let i = 0; i <= 8; i++) {
      const q = axisAngle([0, 1, 0], i * 0.25);
      const flip = i % 2 === 0 ? 1 : -1;
      samples.push(sample(i * 10, ORIGIN, [q[0] * flip, q[1] * flip, q[2] * flip, q[3] * flip]));
    }
    const curve = trajectoryCurve(trajectory(samples), 80);
    let previous = parseMatrix3d(curve(0));
    for (let i = 1; i <= 256; i++) {
      const current = parseMatrix3d(curve(i / 256));
      const angle = rotationDelta(previous, current);
      assert.ok(angle < 0.05, `step ${i} rotated ${angle} rad, far past the 0.25 rad per 10 ms track`);
      previous = current;
    }
  });
});

/** The angle between the rotation blocks of two emitted matrices, in radians. */
// The angle between two rotations, measured so that SMALL angles survive the
// 6-decimal rounding in the emitted matrix string.
//
// The obvious estimator, `acos((trace(A^T B) - 1) / 2)`, is unusable here:
// d(theta)/d(trace) is -1/(2 sin theta), so as the step shrinks it amplifies the
// rounding without bound. Measured against exact rotations rounded to 6 decimals,
// worst-case error by step size:
//
//     theta     acos       atan2      amplification
//     0.0010    1.00e-3    8.6e-7     1166x
//     0.0065    1.46e-4    9.4e-7      156x
//     0.0500    2.50e-5    8.8e-7       28x
//     1.0000    1.15e-6    1.0e-6      1.2x
//
// This test samples 4096 steps, so theta is ~0.0065 and `acos` contributed
// ~1.15e-4 of pure measurement noise -- larger than the tolerance, and about
// 2% of the bound being checked. Reading the skew part directly is O(theta) and
// is not amplified, so its error stays flat at ~1e-6 for every step size.
function rotationDelta(a: number[], b: number[]): number {
  // (A^T B)[row][column]; the inputs are column-major, so element (r, c) of a
  // matrix M lives at M[c * 4 + r] and A^T B at (r, c) is sum_k A(k,r) B(k,c).
  const product = (row: number, column: number): number => {
    let sum = 0;
    for (let k = 0; k < 3; k++) sum += a[row * 4 + k] * b[column * 4 + k];
    return sum;
  };
  const skewX = product(2, 1) - product(1, 2);
  const skewY = product(0, 2) - product(2, 0);
  const skewZ = product(1, 0) - product(0, 1);
  const sine = Math.hypot(skewX, skewY, skewZ) / 2;
  const cosine = (product(0, 0) + product(1, 1) + product(2, 2) - 1) / 2;
  return Math.atan2(sine, cosine);
}

function translationDelta(a: number[], b: number[]): number {
  const dx = a[12] - b[12];
  const dy = a[13] - b[13];
  const dz = a[14] - b[14];
  return Math.sqrt(dx * dx + dy * dy + dz * dz);
}

/**
 * The fastest a trajectory's own samples ever move, per millisecond.
 *
 * This is what makes the continuity bound derived rather than picked. Inside a
 * segment the baked pose moves at that segment's rate, and a grid step that
 * straddles a seam moves at a weighted average of two segments' rates, so no
 * step of `dt` can move further than `maxRate * dt` — with no slack invented
 * anywhere. `fastest*` describes the segment that sets the angular rate, which
 * is what the teeth checks below reason about.
 */
function trajectoryRates(die: DieTrajectory): {
  angular: number;
  linear: number;
  maxSegmentAngle: number;
  fastestAngle: number;
  fastestDt: number;
} {
  let angular = 0;
  let linear = 0;
  let maxSegmentAngle = 0;
  let fastestAngle = 0;
  let fastestDt = 1;
  for (let i = 1; i < die.samples.length; i++) {
    const previous = die.samples[i - 1];
    const current = die.samples[i];
    const dt = current.t - previous.t;
    const angle = quatAngle(previous.orientation, current.orientation);
    const distance = Math.hypot(
      current.position[0] - previous.position[0],
      current.position[1] - previous.position[1],
      current.position[2] - previous.position[2],
    );
    maxSegmentAngle = Math.max(maxSegmentAngle, angle);
    if (angle / dt > angular) {
      angular = angle / dt;
      fastestAngle = angle;
      fastestDt = dt;
    }
    linear = Math.max(linear, distance / dt);
  }
  return { angular, linear, maxSegmentAngle, fastestAngle, fastestDt };
}

function assertContinuous(
  die: DieTrajectory,
  durationMs: number,
  steps: number,
  radiusPx: number,
  teeth: boolean,
): void {
  const rates = trajectoryRates(die);
  const stepMs = durationMs / steps;
  const angleBound = rates.angular * stepMs;
  const translationBound = rates.linear * stepMs * radiusPx;

  if (teeth) {
    // The bound has to be small enough to catch the two things it exists for.
    // A slerp that goes the long way round the fastest segment covers
    // `2*pi - angle` in the same `dt`; a seam at a sample boundary dumps a whole
    // segment's rotation into one step. Both are checked against the real
    // numbers rather than assumed.
    const longWayStep = Math.min(
      Math.PI,
      ((2 * Math.PI - rates.fastestAngle) / rates.fastestDt) * stepMs,
    );
    assert.ok(
      longWayStep > angleBound * 8,
      `bound ${angleBound} has no teeth against a long-way slerp step of ${longWayStep}`,
    );
    assert.ok(
      rates.maxSegmentAngle > angleBound * 8,
      `bound ${angleBound} has no teeth against a seam of ${rates.maxSegmentAngle}`,
    );
  }

  const curve = trajectoryCurve(die, durationMs, { radiusPx });
  let previous = parseMatrix3d(curve(0));
  for (let i = 1; i <= steps; i++) {
    const current = parseMatrix3d(curve(i / steps));
    const angle = rotationDelta(previous, current);
    const distance = translationDelta(previous, current);
    // Slack absorbs the 6-decimal rounding in the emitted string, and nothing
    // else. `rotationDelta` reads the skew part directly, so its residual error
    // is ~1e-6 rad at every step size (see its comment) -- 1e-5 is ten times
    // that, and ~0.15% of the bound rather than the ~2% the old estimator
    // needed. Translation is a plain component difference, so its rounding is
    // ~1e-6 px and unamplified.
    assert.ok(angle <= angleBound + 1e-5, `step ${i} of ${steps} rotated ${angle} rad, bound ${angleBound}`);
    assert.ok(
      distance <= translationBound + 1e-5,
      `step ${i} of ${steps} moved ${distance} px, bound ${translationBound}`,
    );
    previous = current;
  }
}

describe('trajectoryCurve continuity', () => {
  for (const seed of [3, 17, 42]) {
    it(`never jumps more than the trajectory's own rate allows (seed ${seed})`, () => {
      const roll = simulateRoll({ seed, geometry: dieGeometry(6), dieCount: 2, bounds: { x: 4, y: 4, z: 4 } });
      for (const die of roll.dice) {
        // The compiler's maximum resolution: 256 keyframes, 255 steps apart.
        assertContinuous(die, roll.durationMs, 255, 20, false);
        // And a grid fine enough that the same bound provably has teeth: a
        // seam or a long-way slerp would exceed it by more than eightfold.
        assertContinuous(die, roll.durationMs, 4096, 20, true);
      }
    });
  }
});

describe('restingTransform', () => {
  it('equals curve(1) exactly, so nothing jumps when the animation is removed', () => {
    const roll = simulateRoll({ seed: 5, geometry: dieGeometry(12), dieCount: 3, bounds: { x: 5, y: 5, z: 5 } });
    for (const die of roll.dice) {
      const curve = trajectoryCurve(die, roll.durationMs, { radiusPx: 18 });
      assert.equal(restingTransform(die, { radiusPx: 18 }), curve(1));
    }
  });

  it('equals curve(1) for a hand-built trajectory too', () => {
    assert.equal(restingTransform(FALLING), trajectoryCurve(FALLING, 100)(1));
  });

  it('scales with radiusPx', () => {
    const die = trajectory([sample(0, [0, -1.5, 0], QUAT_IDENTITY)]);
    assert.equal(parseMatrix3d(restingTransform(die, { radiusPx: 40 }))[13], 60);
    assert.equal(parseMatrix3d(restingTransform(die))[13], 1.5);
  });
});

describe('trajectoryCurve validation', () => {
  it('rejects a trajectory with no samples', () => {
    assert.throws(() => trajectoryCurve({ samples: [], restingOrientation: QUAT_IDENTITY }, 100), /at least one sample/);
  });

  it('rejects non-monotonic sample times', () => {
    const die = trajectory([
      sample(0, ORIGIN, QUAT_IDENTITY),
      sample(50, ORIGIN, QUAT_IDENTITY),
      sample(50, ORIGIN, QUAT_IDENTITY),
    ]);
    assert.throws(() => trajectoryCurve(die, 100), /strictly increasing/);
  });

  it('rejects a duration shorter than the trajectory, which would truncate the roll', () => {
    assert.throws(() => trajectoryCurve(FALLING, 60), /shorter than/);
    for (const bad of [0, -1, Number.NaN, Number.POSITIVE_INFINITY]) {
      assert.throws(() => trajectoryCurve(FALLING, bad), /duration/);
    }
  });

  it('rejects non-finite geometry', () => {
    const badPosition = trajectory([sample(0, [0, Number.NaN, 0], QUAT_IDENTITY)]);
    assert.throws(() => trajectoryCurve(badPosition, 100), /finite/);
    const badQuat = trajectory([sample(0, ORIGIN, [0, 0, 0, 0])]);
    assert.throws(() => trajectoryCurve(badQuat, 100), /zero-length|finite/);
  });

  it('rejects a resting orientation that disagrees with the final sample', () => {
    const die: DieTrajectory = {
      samples: [sample(0, ORIGIN, QUAT_IDENTITY)],
      restingOrientation: axisAngle([0, 1, 0], 0.5),
    };
    assert.throws(() => trajectoryCurve(die, 100), /resting orientation/);
    assert.throws(() => restingTransform(die), /resting orientation/);
  });

  it('rejects a non-finite radius', () => {
    assert.throws(() => trajectoryCurve(FALLING, 100, { radiusPx: 0 }), /radiusPx/);
    assert.throws(() => trajectoryCurve(FALLING, 100, { radiusPx: Number.NaN }), /radiusPx/);
  });

  it('rejects a progress that is not a number', () => {
    const curve = trajectoryCurve(FALLING, 100);
    assert.throws(() => curve(Number.NaN), /progress/);
  });
});
