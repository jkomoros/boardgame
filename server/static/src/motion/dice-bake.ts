/**
 * A simulated die trajectory, baked into a pure function of animation progress.
 *
 * `dice-sim.ts` produces poses on its own 60 Hz grid; the motion-track compiler
 * wants a `(progress: number) => string` it can sample on a uniform grid of its
 * own choosing (64 keyframes by default, 256 at most). This module is the bridge:
 * it interpolates between simulation samples — lerping position, slerping
 * orientation — and formats the result as a CSS transform.
 *
 * ## The Y axis, which is the whole reason this module is not a one-liner
 *
 * The simulation's world is right-handed with +Y UP, matching `die-faces.ts`'s
 * `WORLD_UP`. CSS's transform space has +Y pointing DOWN the screen. They are
 * opposite, and nothing upstream converts between them: `dice-sim.ts` says so in
 * its header and leaves the job here on purpose.
 *
 * The conversion is the reflection `S = diag(1, -1, 1)`, applied in exactly one
 * place — `cssMatrixFromPose` — and applied to the WHOLE pose:
 *
 *   - a point maps to `S p`, so a die falling toward physics -Y translates
 *     toward CSS +Y, which is down the screen;
 *   - an orientation maps to the similarity `S R S`, i.e. `R'_ij = s_i s_j R_ij`.
 *
 * Both halves are needed. Negating only the translation would leave every
 * rotation mirrored: a die whose physics pose points a face upward would point
 * it downward on screen and land showing the wrong number. Conjugating a
 * rotation by a reflection is still a rotation (`det(S R S) = det R = 1`), so no
 * mirror image reaches the screen — but it does reverse the SENSE of every
 * rotation, which is exactly right, because screen-up and world-up are opposite.
 *
 * A tempting shortcut when dice come out upside down is to flip `WORLD_UP` in
 * `die-faces.ts` instead. Do not: that module's tests pin it to +Y, the reading
 * rules are stated in terms of it, and the flip is a property of the RENDERING
 * frame, not of the physics.
 *
 * ## Why literal `matrix3d`, never `var()` or `calc()`
 *
 * Chromium only promotes a transform animation to the compositor when every
 * keyframe value is a literal transform list. A `calc()` or `var()` anywhere in
 * it silently demotes the animation to the main thread, where a multi-second
 * tumble at 60 Hz stutters under any layout work. So every value this module
 * emits is a bare `matrix3d(...)` of sixteen plain decimal numbers — no units,
 * no custom properties, no scientific notation (which some parsers reject).
 *
 * ## Units
 *
 * Simulated dice are unit-circumradius, so trajectory positions are in DIE
 * RADII, not pixels (see `dice-sim.ts`'s normalisation contract). `radiusPx`
 * says what one radius is worth on screen; it defaults to 1, which is only
 * useful for tests.
 */

import type { Quat, Vec3 } from './die-geometry.ts';
import type { DieTrajectory } from './dice-sim.ts';

export interface BakeOptions {
  /**
   * On-screen length of one die circumradius, in CSS pixels. Trajectories are
   * in circumradii, so this is the only scale in the pipeline. Default 1.
   */
  readonly radiusPx?: number;
}

/**
 * Decimal places kept in the emitted numbers.
 *
 * Six is well past what a screen can show — a rotation entry of 1e-6 tilts a
 * 100 px die by two thousandths of a pixel — while keeping every value out of
 * exponential notation (the smallest non-zero magnitude is exactly 1e-6) and
 * keeping 256 keyframes of matrix strings small.
 */
const DECIMALS = 6;

/**
 * The physics-to-CSS axis signs: `S = diag(1, -1, 1)`. See the file docs; this
 * is the single place the Y flip lives.
 */
const CSS_AXIS_SIGN = [1, -1, 1] as const;

/** Two quaternions closer than this in |dot| are treated as the same orientation. */
const RESTING_TOLERANCE = 1e-6;
/** Below this arc, slerp degenerates and a normalised lerp is exact to float noise. */
const SLERP_LINEAR_THRESHOLD = 0.9995;

/** A trajectory validated and flattened into the arrays the curve reads. */
interface Baked {
  readonly times: Float64Array;
  /** 3 per sample. */
  readonly positions: Float64Array;
  /**
   * 4 per sample, unit length, and sign-aligned so consecutive entries have a
   * non-negative dot product. See `bake`: that alignment is what makes every
   * slerp below take the short arc.
   */
  readonly orientations: Float64Array;
  readonly count: number;
  readonly radiusPx: number;
}

function radiusOf(options?: BakeOptions): number {
  const radiusPx = options?.radiusPx ?? 1;
  if (!Number.isFinite(radiusPx) || radiusPx <= 0) {
    throw new Error(`radiusPx must be a positive finite number, got ${radiusPx}`);
  }
  return radiusPx;
}

function finite(value: number, what: string): number {
  if (!Number.isFinite(value)) throw new Error(`${what} must be finite, got ${value}`);
  return value;
}

/**
 * Validate a trajectory and flatten it, normalising every quaternion and
 * choosing each one's SIGN.
 *
 * `q` and `-q` are the same orientation but opposite slerp paths, and an
 * integrator is free to hand back either — a single sign change between two
 * samples turns a 2-degree step into a 358-degree one, which reads as the die
 * snapping backwards for one frame. Rather than testing the sign inside the
 * inner loop, the whole track is put into one hemisphere here, once: every
 * sample is flipped, if needed, to have a non-negative dot with its predecessor.
 * Interpolation downstream is then unconditionally the short arc.
 */
function bake(die: DieTrajectory, options: BakeOptions | undefined): Baked {
  const samples = die.samples;
  if (!Array.isArray(samples) && !(samples && typeof samples.length === 'number')) {
    throw new Error('a die trajectory must carry at least one sample');
  }
  const count = samples.length;
  if (count < 1) throw new Error('a die trajectory must carry at least one sample');

  const times = new Float64Array(count);
  const positions = new Float64Array(count * 3);
  const orientations = new Float64Array(count * 4);

  for (let i = 0; i < count; i++) {
    const sample = samples[i];
    times[i] = finite(sample.t, `sample ${i} time`);
    if (i > 0 && !(times[i] > times[i - 1])) {
      throw new Error(
        `sample times must be strictly increasing: sample ${i} is at ${times[i]}, after ${times[i - 1]}`,
      );
    }
    for (let axis = 0; axis < 3; axis++) {
      positions[i * 3 + axis] = finite(sample.position[axis], `sample ${i} position`);
    }
    writeUnitQuat(sample.orientation, orientations, i * 4, `sample ${i} orientation`);
    if (i > 0) {
      let dot = 0;
      for (let n = 0; n < 4; n++) dot += orientations[(i - 1) * 4 + n] * orientations[i * 4 + n];
      if (dot < 0) for (let n = 0; n < 4; n++) orientations[i * 4 + n] = -orientations[i * 4 + n];
    }
  }

  // `DieTrajectory` documents `restingOrientation` as the final sample's
  // orientation, and `restingTransform` has to agree with `curve(1)` down to the
  // string or a die visibly jumps when its animation is removed. Rather than
  // pick one of two nearly-equal values, insist they are the same orientation
  // and then use the sample everywhere.
  const resting = new Float64Array(4);
  writeUnitQuat(die.restingOrientation, resting, 0, 'resting orientation');
  let restingDot = 0;
  for (let n = 0; n < 4; n++) restingDot += resting[n] * orientations[(count - 1) * 4 + n];
  if (Math.abs(restingDot) < 1 - RESTING_TOLERANCE) {
    throw new Error(
      'resting orientation disagrees with the final sample; the two would render as different poses',
    );
  }

  return { times, positions, orientations, count, radiusPx: radiusOf(options) };
}

function writeUnitQuat(q: Quat, out: Float64Array, offset: number, what: string): void {
  const x = finite(q[0], what);
  const y = finite(q[1], what);
  const z = finite(q[2], what);
  const w = finite(q[3], what);
  const length = Math.sqrt(x * x + y * y + z * z + w * w);
  if (!(length > 0)) throw new Error(`${what} is a zero-length quaternion`);
  out[offset] = x / length;
  out[offset + 1] = y / length;
  out[offset + 2] = z / length;
  out[offset + 3] = w / length;
}

/** Index of the last sample at or before `t`, by binary search. */
function segmentAt(times: Float64Array, count: number, t: number): number {
  let low = 0;
  let high = count - 1;
  while (low < high) {
    const middle = (low + high + 1) >> 1;
    if (times[middle] <= t) low = middle;
    else high = middle - 1;
  }
  return low;
}

const POSE_POSITION = new Float64Array(3);
const POSE_ORIENTATION = new Float64Array(4);

/** Interpolate the pose at trajectory time `t` into the module scratch. */
function poseAt(baked: Baked, t: number): void {
  const { times, positions, orientations, count } = baked;
  const first = times[0];
  const last = times[count - 1];
  const clamped = t <= first ? first : t >= last ? last : t;
  const index = segmentAt(times, count, clamped);
  if (index >= count - 1) {
    POSE_POSITION[0] = positions[index * 3];
    POSE_POSITION[1] = positions[index * 3 + 1];
    POSE_POSITION[2] = positions[index * 3 + 2];
    POSE_ORIENTATION[0] = orientations[index * 4];
    POSE_ORIENTATION[1] = orientations[index * 4 + 1];
    POSE_ORIENTATION[2] = orientations[index * 4 + 2];
    POSE_ORIENTATION[3] = orientations[index * 4 + 3];
    return;
  }

  const span = times[index + 1] - times[index];
  const u = span > 0 ? (clamped - times[index]) / span : 0;

  for (let axis = 0; axis < 3; axis++) {
    const a = positions[index * 3 + axis];
    const b = positions[(index + 1) * 3 + axis];
    POSE_POSITION[axis] = a + (b - a) * u;
  }
  slerpInto(orientations, index * 4, (index + 1) * 4, u, POSE_ORIENTATION);
}

/**
 * Spherical linear interpolation between two already sign-aligned quaternions
 * (see `bake`), so the arc taken here is always the short one.
 *
 * Nearly-parallel pairs fall back to a normalised lerp: the `sin` denominators
 * below go to zero there, and at that separation the two curves differ by far
 * less than the rounding applied to the emitted string anyway.
 */
function slerpInto(
  source: Float64Array,
  a: number,
  b: number,
  u: number,
  out: Float64Array,
): void {
  let cosine = 0;
  for (let n = 0; n < 4; n++) cosine += source[a + n] * source[b + n];
  cosine = Math.min(1, Math.max(-1, cosine));

  let weightA: number;
  let weightB: number;
  if (cosine > SLERP_LINEAR_THRESHOLD) {
    weightA = 1 - u;
    weightB = u;
  } else {
    const theta = Math.acos(cosine);
    const sine = Math.sin(theta);
    weightA = Math.sin((1 - u) * theta) / sine;
    weightB = Math.sin(u * theta) / sine;
  }

  let x = source[a] * weightA + source[b] * weightB;
  let y = source[a + 1] * weightA + source[b + 1] * weightB;
  let z = source[a + 2] * weightA + source[b + 2] * weightB;
  let w = source[a + 3] * weightA + source[b + 3] * weightB;
  const length = Math.sqrt(x * x + y * y + z * z + w * w);
  if (!(length > 0)) throw new Error('interpolated orientation collapsed to a zero quaternion');
  x /= length;
  y /= length;
  z /= length;
  w /= length;
  out[0] = x;
  out[1] = y;
  out[2] = z;
  out[3] = w;
}

/**
 * A number as CSS wants it: plain decimal, no exponent, no `-0`.
 *
 * `toFixed` is used rather than `String` precisely because `String(1e-7)` is
 * `"1e-7"`, which is not a value every transform parser accepts.
 */
function formatNumber(value: number): string {
  if (!Number.isFinite(value)) {
    throw new Error(`cannot format a non-finite transform component: ${value}`);
  }
  let text = value.toFixed(DECIMALS);
  if (text.includes('.')) {
    text = text.replace(/0+$/, '');
    if (text.endsWith('.')) text = text.slice(0, -1);
  }
  return text === '-0' ? '0' : text;
}

/**
 * The scratch pose, as a CSS `matrix3d`.
 *
 * THE Y FLIP LIVES HERE, and nowhere else in this module. The physics frame is
 * +Y up and CSS's is +Y down, so the whole pose is conjugated by the reflection
 * `S = diag(1, -1, 1)`: `p -> S p` and `R -> S R S`, the latter written out as
 * `s_i s_j R_ij`. See the file docs for why both halves are required and why the
 * result is still a proper rotation.
 *
 * `matrix3d` takes its sixteen components in COLUMN-major order, so the first
 * four are the image of the x axis and the last four are the translation.
 */
function cssMatrixFromPose(radiusPx: number): string {
  const x = POSE_ORIENTATION[0];
  const y = POSE_ORIENTATION[1];
  const z = POSE_ORIENTATION[2];
  const w = POSE_ORIENTATION[3];
  // Body-to-world rotation, row-major, in the physics frame.
  const r = [
    1 - 2 * (y * y + z * z), 2 * (x * y - z * w), 2 * (x * z + y * w),
    2 * (x * y + z * w), 1 - 2 * (x * x + z * z), 2 * (y * z - x * w),
    2 * (x * z - y * w), 2 * (y * z + x * w), 1 - 2 * (x * x + y * y),
  ];
  for (let row = 0; row < 3; row++) {
    for (let column = 0; column < 3; column++) {
      r[row * 3 + column] *= CSS_AXIS_SIGN[row] * CSS_AXIS_SIGN[column];
    }
  }
  const t = [
    POSE_POSITION[0] * CSS_AXIS_SIGN[0] * radiusPx,
    POSE_POSITION[1] * CSS_AXIS_SIGN[1] * radiusPx,
    POSE_POSITION[2] * CSS_AXIS_SIGN[2] * radiusPx,
  ];
  const components = [
    r[0], r[3], r[6], 0,
    r[1], r[4], r[7], 0,
    r[2], r[5], r[8], 0,
    t[0], t[1], t[2], 1,
  ];
  return `matrix3d(${components.map(formatNumber).join(', ')})`;
}

function transformAt(baked: Baked, t: number): string {
  poseAt(baked, t);
  return cssMatrixFromPose(baked.radiusPx);
}

/**
 * Bake `die` into a pure function of animation progress, for
 * `MotionCurveInput.curve`.
 *
 * `durationMs` is the wall-clock span progress covers: progress `p` reads the
 * trajectory at `p * durationMs`. The intended argument is the roll's own
 * `RollTrajectory.durationMs`, which is exactly the final sample's time. A
 * LONGER duration is allowed and simply holds the die at rest for the tail; a
 * shorter one is rejected, because it would silently truncate the roll and leave
 * `curve(1)` disagreeing with `restingTransform` — the die would jump the
 * instant its animation was removed.
 *
 * Progress outside [0, 1] clamps rather than extrapolating: a physics trajectory
 * has nothing to say beyond its own ends.
 */
export function trajectoryCurve(
  die: DieTrajectory,
  durationMs: number,
  options?: BakeOptions,
): (progress: number) => string {
  const baked = bake(die, options);
  if (!Number.isFinite(durationMs) || durationMs <= 0) {
    throw new Error(`curve duration must be a positive finite number of ms, got ${durationMs}`);
  }
  const span = baked.times[baked.count - 1] - baked.times[0];
  if (durationMs < span - 1e-9) {
    throw new Error(
      `curve duration ${durationMs}ms is shorter than the ${span}ms trajectory, which would truncate the roll`,
    );
  }
  return (progress: number): string => {
    if (!Number.isFinite(progress)) {
      throw new Error(`curve progress must be finite, got ${progress}`);
    }
    const clamped = progress <= 0 ? 0 : progress >= 1 ? 1 : progress;
    return transformAt(baked, baked.times[0] + clamped * durationMs);
  };
}

/**
 * The transform a settled die should hold once its animation is gone.
 *
 * Byte-identical to `trajectoryCurve(die, ...)(1)` by construction — both are
 * the final sample through the same formatter. Animations run with
 * `fill: 'none'`, so the element snaps to its resting style the moment the
 * animation finishes; if these two disagreed by so much as a rounding digit,
 * every die would twitch at the end of its roll.
 */
export function restingTransform(die: DieTrajectory, options?: BakeOptions): string {
  const baked = bake(die, options);
  return transformAt(baked, Number.POSITIVE_INFINITY);
}
