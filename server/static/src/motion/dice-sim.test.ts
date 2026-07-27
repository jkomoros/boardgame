import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { dieGeometry, type DieGeometry, type Quat, type Vec3 } from './die-geometry.ts';
import { presentedFaceIndex, resolveReadingRule, WORLD_UP } from './die-faces.ts';
import {
  simulateRoll,
  simulateRollWithDiagnostics,
  simulationSolid,
  type RollConfig,
} from './dice-sim.ts';

/**
 * Vector helpers written out here rather than imported, so every assertion below
 * is checked against arithmetic this file owns. A bug shared between the module
 * and its test would otherwise be invisible.
 */
const sub = (a: Vec3, b: Vec3): Vec3 => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const addv = (a: Vec3, b: Vec3): Vec3 => [a[0] + b[0], a[1] + b[1], a[2] + b[2]];
const dot = (a: Vec3, b: Vec3): number => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const mag = (a: Vec3): number => Math.sqrt(dot(a, a));
const scaleVec = (a: Vec3, factor: number): Vec3 => [a[0] * factor, a[1] * factor, a[2] * factor];

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

/** The smallest rotation angle, in radians, carrying `a` onto `b`. */
function quatAngle(a: Quat, b: Quat): number {
  const na = Math.sqrt(a[0] * a[0] + a[1] * a[1] + a[2] * a[2] + a[3] * a[3]);
  const nb = Math.sqrt(b[0] * b[0] + b[1] * b[1] + b[2] * b[2] + b[3] * b[3]);
  let d = (a[0] * b[0] + a[1] * b[1] + a[2] * b[2] + a[3] * b[3]) / (na * nb);
  d = Math.min(1, Math.abs(d));
  return 2 * Math.acos(d);
}

/** World-frame vertices of one die at one sample. */
function worldVertices(
  solidVertices: readonly Vec3[],
  position: Vec3,
  orientation: Quat,
): Vec3[] {
  return solidVertices.map((vertex) => addv(position, rotate(orientation, vertex)));
}

/**
 * The container the simulator is documented to build from `bounds`: a box of
 * HALF-extents centred on the origin. Written from the doc comment, not read
 * out of the module, so a change of convention fails here.
 */
function containerPlanes(bounds: { x: number; y: number; z: number }): {
  normal: Vec3;
  offset: number;
}[] {
  return [
    { normal: [1, 0, 0], offset: bounds.x },
    { normal: [-1, 0, 0], offset: bounds.x },
    { normal: [0, 1, 0], offset: bounds.y },
    { normal: [0, -1, 0], offset: bounds.y },
    { normal: [0, 0, 1], offset: bounds.z },
    { normal: [0, 0, -1], offset: bounds.z },
  ];
}

const BOUNDS = { x: 6, y: 6, z: 6 } as const;

function config(faceCount: number, seed: number, overrides: Partial<RollConfig> = {}): RollConfig {
  return {
    seed,
    geometry: dieGeometry(faceCount),
    dieCount: 1,
    bounds: BOUNDS,
    ...overrides,
  };
}

/**
 * Face counts that between them cover every branch a shape can take: the
 * triangular barrel (a generated solid, and read from its DOWN face because an
 * odd barrel has no up face), the tetrahedron (read from the face it RESTS on),
 * the cube (the only shape a box special case would ever be tuned for), the
 * trapezohedral d10, the d12 and the d20 (whose raw unit-mass inertia is 4x a
 * d10's before normalisation).
 *
 * The d7 is here for a specific reason: its cap facets used to be stable rests,
 * so it landed on an unreadable cap in most rolls and had to be held out of
 * this list. `die-geometry.ts` now proportions the caps so no facet of them is
 * a stable rest, and the d7 is held to the same landing contract as everything
 * else.
 */
const SHAPES = [3, 4, 6, 7, 10, 12, 20] as const;
const SEEDS = [1, 2, 7, 12345] as const;

/**
 * How far a vertex may sit outside the container, in units of the die's
 * circumradius (which the simulator normalises to 1).
 *
 * A billionth. Containment is not approximate here: the step ends with a clamp
 * that projects any escaped vertex back onto the plane, so the only slack that
 * should ever appear is floating-point residue. Across 480 three-die rolls the
 * worst excursion measured 1.8e-15, so this leaves six orders of margin and
 * would still catch a clamp that had been weakened to a Baumgarte-style partial
 * correction -- which a percent-scale epsilon would wave straight through.
 */
const CONTAINMENT_EPSILON = 1e-9;

/**
 * What the roll's LAST frame is allowed to be, now that the trajectory is cut
 * at the last frame in which the dice visibly moved rather than run out to a
 * dead stop.
 *
 * A trimmed roll cannot be checked for stillness at its end — its end is by
 * construction a frame that was still moving. What can be checked, and is
 * strictly more to the point, is that nothing worth watching was cut: how far
 * the die still had to go when the animation stopped (`restingDrift`), and that
 * it stopped ON the floor rather than a bounce above it.
 *
 * Both numbers are the trim's price and both are measured. Across the 7 shapes
 * x 4 seeds x 2 dice this file rolls: the worst drift is 2.17 degrees, and the
 * worst final-frame clearance above the floor is 3.1e-3 of a circumradius. The
 * bounds below leave those roughly 40% of headroom. The clearance bound is also
 * still well under `CONTACT_MARGIN` (0.02), which is what it has to be to keep
 * catching the bug it was written for — see the test.
 */
const MAX_RESTING_DRIFT_DEGREES = 3;
const MAX_FLOOR_CLEARANCE = 0.005;

/**
 * A die balanced on an edge or a corner must fail this.
 *
 * Five degrees is a real bound for every shape in `SHAPES`: the smallest angle
 * any of them can be tilted by and still be balanced on something is the d7's
 * side edge, at 180/7 = 25.7 degrees. It is NOT a bound that generalises. A
 * barrel's adjacent side-face normals are 360/N apart, so resting on a side
 * edge puts it 180/N from a face normal, and by N = 36 that is 5 degrees
 * exactly: a d36 balanced on an edge would sail through this. Nothing here
 * rolls a barrel that large — `die-geometry.test.ts` scales the same limit by
 * face count where it does go up to a d24 — but the constant is only as strong
 * as the shapes it is applied to, and that is worth saying next to it.
 */
const MAX_COCK_DEGREES = 5;

describe('simulateRoll determinism', () => {
  it('produces a bitwise-identical trajectory for the same seed', () => {
    for (const faceCount of SHAPES) {
      const first = simulateRoll(config(faceCount, 4242, { dieCount: 2 }));
      const second = simulateRoll(config(faceCount, 4242, { dieCount: 2 }));
      assert.deepStrictEqual(second, first, `d${faceCount} is not deterministic`);
    }
  });

  it('produces different trajectories for different seeds', () => {
    const a = simulateRoll(config(6, 1));
    const b = simulateRoll(config(6, 2));
    assert.notDeepStrictEqual(b, a, 'seeds 1 and 2 produced the same roll');
    assert.notDeepStrictEqual(
      b.dice[0].restingOrientation,
      a.dice[0].restingOrientation,
      'seeds 1 and 2 came to rest identically',
    );
  });

  /**
   * Seeds that used to collide, and the shape of the bug they came from.
   *
   * `createRandom` truncated its seed to an int32, so everything in each of
   * these pairs hashed to the same uint32 and produced the SAME ROLL: a
   * fractional part was thrown away, anything at or above 2^32 wrapped, and
   * negatives wrapped onto the top of the range. The renderer derives its seed
   * from `(component id, RollCount)`, so a hash wider than 32 bits or a ratio
   * would have replayed one roll's animation for a different roll with nothing
   * anywhere saying so. (Today's renderer hashes with FNV-1a and hands over a
   * uint32, so it would trip none of these; the contract is on the parameter,
   * which is a `number`, not on the one caller that exists.)
   *
   * Bitwise-identical trajectories are the assertion, not merely "different
   * resting faces": two rolls can land the same way by chance, but they cannot
   * agree on every sample of a 40-sample tumble unless the streams are equal.
   */
  const ALIASING_PAIRS: readonly (readonly [number, number])[] = [
    [3, 3.7],
    [3, 3.2],
    [1, 2 ** 32 + 1],
    [7, 2 ** 32 + 7],
    [-1, 2 ** 32 - 1],
    [0, 1e-9],
    // Two more the same reasoning implies: the low half of a double is not the
    // only half that matters, and the sign bit is not noise.
    [1, -1],
    [2 ** 53, 2 ** 53 + 2],
  ];

  it('gives every distinct finite seed its own roll', () => {
    for (const [a, b] of ALIASING_PAIRS) {
      assert.notEqual(a, b, `${a} and ${b} are the same number; the pair proves nothing`);
      assert.notDeepStrictEqual(
        simulateRoll(config(6, b)),
        simulateRoll(config(6, a)),
        `seeds ${a} and ${b} produced the same roll`,
      );
    }
  });

  it('treats -0 and 0 as the same seed, because === does', () => {
    // The one collision that is deliberate. A caller computing a seed has no
    // way to know which zero it produced, so they had better roll the same.
    assert.deepStrictEqual(simulateRoll(config(6, -0)), simulateRoll(config(6, 0)));
  });

  it('is unaffected by an intervening simulation', () => {
    const first = simulateRoll(config(20, 99));
    simulateRoll(config(6, 5));
    simulateRoll(config(12, 5));
    const again = simulateRoll(config(20, 99));
    assert.deepStrictEqual(again, first, 'simulateRoll carries state between calls');
  });

  it('never reads Math.random', () => {
    const real = Math.random;
    Math.random = () => {
      throw new Error('simulateRoll must not use Math.random');
    };
    try {
      simulateRoll(config(20, 3, { dieCount: 3 }));
    } finally {
      Math.random = real;
    }
  });
});

describe('simulateRoll shape', () => {
  it('samples a monotonically increasing timeline that ends at durationMs', () => {
    const roll = simulateRoll(config(6, 11, { dieCount: 2 }));
    assert.equal(roll.dice.length, 2);
    for (const die of roll.dice) {
      assert.ok(die.samples.length > 30, `only ${die.samples.length} samples`);
      assert.equal(die.samples[0].t, 0);
      for (let i = 1; i < die.samples.length; i++) {
        assert.ok(die.samples[i].t > die.samples[i - 1].t, `sample ${i} did not advance`);
      }
      assert.equal(die.samples[die.samples.length - 1].t, roll.durationMs);
      assert.deepStrictEqual(
        die.restingOrientation,
        die.samples[die.samples.length - 1].orientation,
        'restingOrientation is not the final sample',
      );
    }
    assert.equal(
      roll.dice[0].samples.length,
      roll.dice[1].samples.length,
      'dice have different sample counts',
    );
  });

  it('emits only finite numbers', () => {
    for (const faceCount of SHAPES) {
      const roll = simulateRoll(config(faceCount, 808, { dieCount: 2 }));
      for (const die of roll.dice) {
        for (const sample of die.samples) {
          for (const value of [...sample.position, ...sample.orientation, sample.t]) {
            assert.ok(Number.isFinite(value), `d${faceCount} produced ${value}`);
          }
        }
      }
    }
  });

  it('rejects a nonsensical configuration', () => {
    assert.throws(() => simulateRoll(config(6, 1, { dieCount: 0 })));
    assert.throws(() => simulateRoll(config(6, 1, { bounds: { x: 6, y: 0, z: 6 } })));
    assert.throws(() => simulateRoll(config(6, Number.NaN)));
    assert.throws(() => simulateRoll(config(6, 1, { restitution: 1.5 })));
    assert.throws(() => simulateRoll(config(6, 1, { friction: -1 })));
  });
});

describe('simulationSolid', () => {
  it('normalises every shape to a circumradius of 1', () => {
    for (const faceCount of SHAPES) {
      const solid = simulationSolid(dieGeometry(faceCount));
      const radius = Math.max(...solid.vertices.map(mag));
      assert.ok(
        Math.abs(radius - 1) < 1e-12,
        `d${faceCount} circumradius is ${radius}, not 1`,
      );
    }
  });

  it('has a real size spread to cancel in the first place', () => {
    // Unit-mass inertia goes as R^2, and the geometry module builds each solid
    // at its own natural scale. Measured: 3.14 (d20) against 0.76 (d10). That
    // 4x is exactly what a die tumbling at the wrong rate for its face count
    // would come from, so this pins the fact the normalisation exists to
    // cancel — if `die-geometry.ts` ever starts normalising itself, everything
    // downstream becomes a silent no-op and this is what says so.
    const trace = (t: readonly number[]): number => t[0] + t[4] + t[8];
    const raw = SHAPES.map((faceCount) => trace(dieGeometry(faceCount).inertiaTensor));
    const spread = Math.max(...raw) / Math.min(...raw);
    assert.ok(
      spread > 3.5,
      `raw inertia traces span only ${spread}x; normalisation may be a no-op now`,
    );

    // There used to be a second assertion here: that the spread of the
    // NORMALISED inverse traces was at least 1.5x narrower than the raw one.
    // It passed with 9% to spare and it bounded nothing. What survives
    // normalisation is genuine SHAPE variation — a tetrahedron really does
    // resist rotation differently from a d20 — so there is no principled
    // number to hold it against, and mutating the normalisation to `/ R`
    // instead of `/ R^2` still satisfied it. The test below, and the closed
    // forms after that, are what actually pin the transform.
  });

  it('is exactly invariant to the scale the geometry happens to be built at', () => {
    // The real bound the "spread narrows" heuristic was reaching for. A die
    // uniformly scaled by any factor is the SAME die to the physics, so
    // `simulationSolid` must return bit-comparable output for both. This is the
    // property that makes `circumradius` cancel; it fails outright for the
    // classic mistake of dividing the tensor by `circumradius` instead of its
    // square, which no comparison of one shape against another can see.
    const stretched = (geometry: DieGeometry, factor: number): DieGeometry => ({
      ...geometry,
      vertices: geometry.vertices.map((v) => scaleVec(v, factor)),
      faces: geometry.faces.map((face) => ({
        normal: face.normal,
        centroid: scaleVec(face.centroid, factor),
        polygon: face.polygon.map((v) => scaleVec(v, factor)),
      })),
      capFaces: geometry.capFaces.map((face) => ({
        normal: face.normal,
        centroid: scaleVec(face.centroid, factor),
        polygon: face.polygon.map((v) => scaleVec(v, factor)),
      })),
      // Unit-mass second moments are lengths squared; the circumradius is a
      // length. Both scalings are the geometry module's own contract.
      inertiaTensor: geometry.inertiaTensor.map((entry) => entry * factor * factor),
      circumradius: geometry.circumradius * factor,
    });

    for (const faceCount of SHAPES) {
      const geometry = dieGeometry(faceCount);
      const reference = simulationSolid(geometry);
      for (const factor of [0.125, 3, 1000]) {
        const scaled = simulationSolid(stretched(geometry, factor));
        for (let i = 0; i < reference.vertices.length; i++) {
          for (let axis = 0; axis < 3; axis++) {
            assert.ok(
              Math.abs(scaled.vertices[i][axis] - reference.vertices[i][axis]) < 1e-12,
              `d${faceCount} at ${factor}x: vertex ${i} moved`,
            );
          }
        }
        for (let i = 0; i < reference.planes.length; i++) {
          assert.ok(
            Math.abs(scaled.planes[i].offset - reference.planes[i].offset) < 1e-12,
            `d${faceCount} at ${factor}x: plane ${i} offset moved`,
          );
        }
        for (let i = 0; i < 9; i++) {
          assert.ok(
            Math.abs(scaled.inertia[i] - reference.inertia[i]) < 1e-12,
            `d${faceCount} at ${factor}x: inertia entry ${i} is ${scaled.inertia[i]}, not ${reference.inertia[i]}`,
          );
        }
      }
    }
  });

  it('matches the published inertia of a unit-circumradius Platonic solid', () => {
    // The independent oracle for the normalisation. Each of these is a textbook
    // closed form for a uniform solid of unit mass whose circumradius is 1, so
    // it is derived from nothing this codebase computes -- which is what makes
    // it able to tell `/ circumradius^2` from `/ circumradius`, a distinction
    // invisible to any test that only compares shapes against each other.
    const expected: readonly (readonly [number, number])[] = [
      // Tetrahedron: m*a^2/20 with edge a = R*sqrt(8/3).
      [4, 8 / 3 / 20],
      // Cube: m*s^2/6 with side s = 2R/sqrt(3).
      [6, 4 / 3 / 6],
      // Octahedron: m*R^2/5.
      [8, 1 / 5],
    ];
    for (const [faceCount, moment] of expected) {
      const inertia = simulationSolid(dieGeometry(faceCount)).inertia;
      for (const index of [0, 4, 8]) {
        assert.ok(
          Math.abs(inertia[index] - moment) < 1e-12,
          `d${faceCount} principal moment ${index}: expected ${moment}, got ${inertia[index]}`,
        );
      }
    }
  });

  it('bounds every vertex by every plane of the closed surface', () => {
    for (const faceCount of SHAPES) {
      const solid = simulationSolid(dieGeometry(faceCount));
      for (const plane of solid.planes) {
        for (const vertex of solid.vertices) {
          assert.ok(
            dot(plane.normal, vertex) <= plane.offset + 1e-9,
            `d${faceCount} vertex lies outside a surface plane`,
          );
        }
      }
    }
  });
});

describe('simulateRoll containment', () => {
  for (const faceCount of SHAPES) {
    it(`keeps every vertex of a d${faceCount} inside the container`, () => {
      const solid = simulationSolid(dieGeometry(faceCount));
      const planes = containerPlanes(BOUNDS);
      for (const seed of SEEDS) {
        const roll = simulateRoll(config(faceCount, seed, { dieCount: 2 }));
        for (const die of roll.dice) {
          for (const sample of die.samples) {
            for (const vertex of worldVertices(solid.vertices, sample.position, sample.orientation)) {
              for (const plane of planes) {
                const outside = dot(plane.normal, vertex) - plane.offset;
                assert.ok(
                  outside <= CONTAINMENT_EPSILON,
                  `d${faceCount} seed ${seed} at t=${sample.t}: vertex ${outside} outside plane ${plane.normal}`,
                );
              }
            }
          }
        }
      }
    });
  }
});

describe('simulateRoll settling', () => {
  /**
   * The direction the PRESENTED face's normal must point when the die is
   * legitimately at rest.
   *
   * Only an `up-face` die shows the face at the ceiling. A d4 is read from the
   * apex, which names the face it is RESTING ON, and an odd-sided barrel is
   * read from below for the same reason; for both, the presented normal points
   * DOWN. Demanding "up" everywhere would fail those shapes for a reason that
   * has nothing to do with the physics.
   */
  function expectedPresentedDirection(geometry: DieGeometry): Vec3 {
    const sign = resolveReadingRule(geometry) === 'up-face' ? 1 : -1;
    return [WORLD_UP[0] * sign, WORLD_UP[1] * sign, WORLD_UP[2] * sign];
  }

  for (const faceCount of SHAPES) {
    it(`brings a d${faceCount} to rest on a face, not on an edge`, () => {
      const geometry = dieGeometry(faceCount);
      const wanted = expectedPresentedDirection(geometry);
      const cosLimit = Math.cos((MAX_COCK_DEGREES * Math.PI) / 180);
      for (const seed of SEEDS) {
        const roll = simulateRoll(config(faceCount, seed, { dieCount: 2 }));
        for (let d = 0; d < roll.dice.length; d++) {
          const die = roll.dice[d];
          const label = `d${faceCount} seed ${seed} die ${d}`;
          // The cocked-die check. presentedFaceIndex always returns SOME face,
          // even for a cube balanced at 45 degrees where the winner leads by
          // 1e-16, so this tolerance is the only thing standing between a
          // legitimate result and a die wedged on an edge.
          const resting = die.restingOrientation;
          const face = geometry.faces[presentedFaceIndex(geometry, resting)];
          const worldNormal = rotate(resting, face.normal);
          const alignment = dot(worldNormal, wanted);
          const degrees = (Math.acos(Math.min(1, Math.max(-1, alignment))) * 180) / Math.PI;
          assert.ok(
            alignment >= cosLimit,
            `${label}: landed cocked - presented face is ${degrees.toFixed(2)} degrees off`,
          );
        }
      }
    });
  }

  it('ends the animation the moment the dice stop being worth watching', () => {
    // The trim's contract, from both sides.
    //
    // Below: nothing worth watching was cut. `restingDrift` is exactly how far
    // the die still had to turn when the trajectory ended, so a trim that
    // guessed too early shows up here as degrees and not as an opinion.
    //
    // Above: the trim actually happened. The last frame of a trimmed roll is by
    // construction one in which a die still MOVED — everything after it was
    // below the threshold, which is why it was cut — so a module that had
    // quietly stopped trimming would end on a frame indistinguishable from its
    // predecessor and fail this. Without it, `restingDrift < 3 degrees` is
    // satisfied perfectly by not trimming at all.
    for (const faceCount of SHAPES) {
      for (const seed of SEEDS) {
        const diagnostics = simulateRollWithDiagnostics(config(faceCount, seed, { dieCount: 2 }));
        const label = `d${faceCount} seed ${seed}`;
        assert.ok(
          diagnostics.restingDrift < MAX_RESTING_DRIFT_DEGREES,
          `${label}: the roll stopped ${diagnostics.restingDrift.toFixed(2)} degrees short of where the die settled`,
        );
        let turned = 0;
        let travelled = 0;
        for (const die of diagnostics.trajectory.dice) {
          const samples = die.samples;
          const n = samples.length;
          turned = Math.max(
            turned,
            (quatAngle(samples[n - 2].orientation, samples[n - 1].orientation) * 180) / Math.PI,
          );
          travelled = Math.max(travelled, mag(sub(samples[n - 1].position, samples[n - 2].position)));
        }
        // Either kind of motion counts: the trim asks the same of rotation and
        // of travel, and a die that slides to a halt without turning is still
        // a die the player is watching.
        assert.ok(
          turned > 1 || travelled > 0.006,
          `${label}: the roll ends on a dead frame - ${turned.toFixed(2)} degrees and ${travelled.toExponential(2)} circumradii`,
        );
      }
    }
  });

  it('comes to rest ON the floor, not hovering above it', () => {
    // The contact solver creates a contact while the vertex still has a gap and
    // constrains how fast that gap may close. Get the no-impact target wrong --
    // a floor of zero rather than "no opinion" -- and it forbids closing the gap
    // at all: every die lands one contact margin off the floor, tiltable to any
    // angle because it is held by a rigid constraint rather than resting on
    // anything. Nothing else in this file notices, because a hovering die is
    // both perfectly contained and perfectly still.
    for (const faceCount of SHAPES) {
      const solid = simulationSolid(dieGeometry(faceCount));
      for (const seed of SEEDS) {
        const roll = simulateRoll(config(faceCount, seed, { dieCount: 2 }));
        for (let d = 0; d < roll.dice.length; d++) {
          const last = roll.dice[d].samples[roll.dice[d].samples.length - 1];
          const lowest = Math.min(
            ...worldVertices(solid.vertices, last.position, last.orientation).map((v) => v[1]),
          );
          const clearance = lowest + BOUNDS.y;
          assert.ok(
            clearance < MAX_FLOOR_CLEARANCE,
            `d${faceCount} seed ${seed} die ${d} came to rest ${clearance} above the floor`,
          );
        }
      }
    }
  });

  it('says so when it hands back a cocked die', () => {
    /**
     * The retry loop cannot save a die that has nowhere to fall over, and
     * `resolved()` accepts half-extents down to 1.5. Three d20 in a
     * `{x: 2, y: 5, z: 2}` shaft come out cocked in most rolls, by as much as
     * 15.8 degrees, with all eight attempts exhausted. That used to be
     * reported nowhere a caller could see it, so the renderer would read a
     * value off a face that is not actually up and show it as the result.
     *
     * The contract asserted here is not "cramped containers work" — they do
     * not — it is that the caller is TOLD. Every roll must either be flat or
     * be flagged, and the flag must agree with the alignment it reports.
     */
    const cramped = { x: 2, y: 5, z: 2 } as const;
    const cosLimit = Math.cos((MAX_COCK_DEGREES * Math.PI) / 180);
    let flagged = 0;
    const trials = 12;
    for (let seed = 1; seed <= trials; seed++) {
      const roll = simulateRoll(config(20, seed, { dieCount: 3, bounds: cramped }));
      const geometry = dieGeometry(20);
      // Recomputed here from the resting orientations rather than trusted, so
      // `restAlignment` is checked against this file's own arithmetic.
      let worst = Infinity;
      for (const die of roll.dice) {
        let best = -Infinity;
        for (const face of geometry.faces) {
          best = Math.max(best, dot(rotate(die.restingOrientation, face.normal), [0, -1, 0]));
        }
        worst = Math.min(worst, best);
      }
      assert.ok(
        Math.abs(worst - roll.restAlignment) < 1e-12,
        `seed ${seed}: reported restAlignment ${roll.restAlignment}, measured ${worst}`,
      );
      if (roll.cocked) flagged++;
      else {
        const degrees = (Math.acos(Math.min(1, worst)) * 180) / Math.PI;
        assert.ok(
          worst >= cosLimit,
          `seed ${seed}: not flagged as cocked but sits ${degrees.toFixed(2)} degrees off a readable face`,
        );
      }
    }
    // And the shaft really is the hard case, or the assertion above would be
    // vacuous: it has to actually produce cocked dice.
    assert.ok(flagged > 0, `${cramped.x}-wide shaft never cocked a die in ${trials} rolls`);
  });

  it('lands flat, and says so, in a container big enough to fall over in', () => {
    // The other side of the measurement that justifies the tray sizes above.
    // Three d20 over these seeds: 17/25 cocked at {2, 5, 2}, 1/25 at {3, 3, 3},
    // 0/25 at {4, 4, 4} and at {6, 6, 6}. Four circumradii of half-extent is
    // where the retry loop starts winning every time, so it is what a caller
    // should reach for; the suite's own 6-cubed tray is well past it.
    for (let seed = 1; seed <= 12; seed++) {
      const roll = simulateRoll(config(20, seed, { dieCount: 3, bounds: { x: 4, y: 4, z: 4 } }));
      assert.equal(roll.cocked, false, `seed ${seed} cocked a die in a 4-cubed tray`);
    }
  });

  it('settles well inside the duration cap', () => {
    for (const faceCount of SHAPES) {
      for (const seed of SEEDS) {
        const roll = simulateRoll(config(faceCount, seed, { dieCount: 2 }));
        assert.ok(
          roll.durationMs < 5000,
          `d${faceCount} seed ${seed} ran the clock out at ${roll.durationMs}ms`,
        );
        assert.ok(
          roll.durationMs > 400,
          `d${faceCount} seed ${seed} finished in ${roll.durationMs}ms, which is not a roll`,
        );
      }
    }
  });
});

describe('simulateRoll energy', () => {
  /**
   * Total mechanical energy must never rise. A sequential-impulse solver that
   * has gone unstable announces itself by gaining energy long before anything
   * visibly explodes, so this is the check that catches a bad restitution or a
   * mis-signed inertia term while the trajectory still looks plausible.
   */
  // Ten parts per million of the launch energy per step. The integrator is not
  // symplectic, so a bound of exactly zero would fail on rounding; the worst
  // single-step gain measured across 480 three-die rolls was 4.5e-7, so this
  // sits about 20x above the noise and orders of magnitude below anything a
  // real instability produces.
  //
  // It is not ALL noise, and the margin here is thinner than it looks. The
  // step ends with a positional clamp that teleports an escaped body back
  // inside the container, and a teleport upward is free potential energy: the
  // worst single step measured injects 6.8e-5 of the launch energy that way,
  // seven times this budget. What keeps the assertion true is that the same
  // step's contact and drag dissipation is larger still. See `dice-sim.ts`'s
  // header — if this ever starts failing by a factor of a few, the clamp is
  // where to look, not the contact solver.
  const RELATIVE_STEP_TOLERANCE = 1e-5;

  /**
   * Every step of a roll's energy trace must fall, except the ones a settle
   * retry declared it was putting energy into.
   *
   * The exemption is a list of exact indices the module reports, not a widened
   * budget: a re-thrown die is handed a launch's worth of kinetic energy and no
   * tolerance that admitted that would still catch a solver going unstable. So
   * the assertion skips exactly the declared steps and holds everywhere else,
   * which is strictly stronger than what it could check before — the retry used
   * to run whole separate throws and only the chosen one's trace came back, so
   * the steps of up to seven other throws were never looked at at all.
   */
  function assertDissipates(
    diagnostics: ReturnType<typeof simulateRollWithDiagnostics>,
    label: string,
  ): void {
    const energy = diagnostics.energyPerStep;
    assert.ok(energy.length > 100, `${label}: only ${energy.length} steps`);
    const injected = new Set(diagnostics.attemptStarts);
    const budget = energy[0] * RELATIVE_STEP_TOLERANCE;
    for (let i = 1; i < energy.length; i++) {
      if (injected.has(i)) continue;
      assert.ok(
        energy[i] <= energy[i - 1] + budget,
        `${label} gained energy at step ${i}: ${energy[i - 1]} -> ${energy[i]} (budget ${budget})`,
      );
    }
  }

  for (const faceCount of SHAPES) {
    it(`never gains energy while rolling a d${faceCount}`, () => {
      for (const seed of SEEDS) {
        const diagnostics = simulateRollWithDiagnostics(config(faceCount, seed, { dieCount: 2 }));
        const energy = diagnostics.energyPerStep;
        assertDissipates(diagnostics, `d${faceCount} seed ${seed}`);
        assert.ok(
          energy[energy.length - 1] < energy[0] * 0.25,
          `d${faceCount} seed ${seed} only dissipated to ${energy[energy.length - 1]} of ${energy[0]}`,
        );
      }
    });
  }

  /**
   * Every restitution `resolved()` will accept, not just the default.
   *
   * The old solver gave each contact its own restitution target inside the
   * iteration loop, and a die landing FLAT — several vertices of one face
   * touching at once — had each of them demand a bounce computed from an
   * approach speed the others had already cancelled. The gain grew with
   * restitution and was completely invisible to the tests above, which only
   * ever ran the default 0.32. Measured on this exact configuration against
   * the pre-fix module:
   *
   *     r<=0.7 -> 1.6e-9 (fine)   r=0.85 -> 1.4e-2   r=0.95 -> 8.2e-2
   *     r=1.0  -> 1.2e-1
   *
   * against 4.2e-9 or better at every one of them now, so the budget below
   * clears the fixed solver by six orders and catches the old one by three.
   * `restitution: 1` is in the sweep deliberately: it is legal, it is the
   * worst case, and it is the value at which nothing else in the step is
   * dissipating enough to hide a pump.
   */
  const RESTITUTIONS = [0, 0.32, 0.5, 0.7, 0.85, 0.95, 1] as const;

  it('never gains energy at any restitution the API accepts', () => {
    // A subset of shapes and seeds, because this is 84 rolls: the cube for the
    // flat multi-vertex manifold that is the whole point, the tetrahedron
    // because it was the worst offender before the fix, and the d20 as the
    // roundest thing here.
    for (const restitution of RESTITUTIONS) {
      for (const faceCount of [4, 6, 20] as const) {
        for (const seed of SEEDS) {
          assertDissipates(
            simulateRollWithDiagnostics(config(faceCount, seed, { dieCount: 2, restitution })),
            `d${faceCount} seed ${seed} restitution ${restitution}`,
          );
        }
      }
    }
  });

  it('still bounces: a higher restitution keeps more energy alive', () => {
    // The other half of the sweep above, and the reason it is not vacuous. A
    // solver that satisfied "never gains energy" by quietly dropping the
    // bounce altogether would pass every other assertion in this file, so pin
    // that the parameter does the thing it names. Energy a second into the
    // roll, as a fraction of the launch energy, averaged over the seeds:
    // 0.072 at restitution 0 against 0.231 at 0.95.
    const survived = (restitution: number): number => {
      let total = 0;
      for (const seed of SEEDS) {
        const energy = simulateRollWithDiagnostics(
          config(6, seed, { dieCount: 2, restitution }),
        ).energyPerStep;
        total += energy[Math.min(700, energy.length - 1)] / energy[0];
      }
      return total / SEEDS.length;
    };
    const dead = survived(0);
    const lively = survived(0.95);
    assert.ok(
      lively > dead * 2,
      `restitution barely changed how much energy survives: ${dead} at 0 against ${lively} at 0.95`,
    );
  });

  it('applies only negligible positional correction', () => {
    // Contacts are resolved by impulse; the positional clamp is a backstop for
    // what the LINEARISED constraint drops, namely the arc a vertex on a
    // spinning die travels within one step. That residue is bounded by
    // `(w dt)^2 r / 2`, which at the ~60 rad/s post-impact spin peak and a
    // 720 Hz step is 3.5e-3 of a circumradius; the worst measured was 3.8e-3.
    // Anything approaching a full percent means the impulse solver, not the
    // clamp, has stopped doing the work.
    const bound = 0.01;
    for (const faceCount of SHAPES) {
      for (const seed of SEEDS) {
        const diagnostics = simulateRollWithDiagnostics(config(faceCount, seed, { dieCount: 3 }));
        assert.ok(
          diagnostics.maxWallCorrection < bound,
          `d${faceCount} seed ${seed} needed a ${diagnostics.maxWallCorrection} positional correction`,
        );
      }
    }
  });
});

describe('simulateRoll with several dice', () => {
  for (const faceCount of [6, 20] as const) {
    it(`keeps three d${faceCount} from interpenetrating`, () => {
      const geometry = dieGeometry(faceCount);
      const solid = simulationSolid(geometry);
      for (const seed of SEEDS) {
        const roll = simulateRoll(config(faceCount, seed, { dieCount: 3 }));
        assert.equal(roll.dice.length, 3);
        const sampleCount = roll.dice[0].samples.length;
        for (let s = 0; s < sampleCount; s++) {
          const poses = roll.dice.map((die) => die.samples[s]);
          const vertices = poses.map((pose) =>
            worldVertices(solid.vertices, pose.position, pose.orientation),
          );
          for (let a = 0; a < poses.length; a++) {
            for (let b = 0; b < poses.length; b++) {
              if (a === b) continue;
              // Depth of the deepest vertex of a inside the convex hull of b.
              for (const vertex of vertices[a]) {
                const local = rotate(
                  [
                    -poses[b].orientation[0],
                    -poses[b].orientation[1],
                    -poses[b].orientation[2],
                    poses[b].orientation[3],
                  ],
                  sub(vertex, poses[b].position),
                );
                let depth = Infinity;
                for (const plane of solid.planes) {
                  depth = Math.min(depth, plane.offset - dot(plane.normal, local));
                }
                assert.ok(
                  depth <= 0.01,
                  `d${faceCount} seed ${seed}: die ${a} is ${depth} inside die ${b} at t=${poses[a].t}`,
                );
              }
            }
          }
        }
      }
    });
  }

  it('actually exercises the die-against-die constraint', () => {
    // Without this, "they do not interpenetrate" would be satisfied by three
    // dice that simply never met.
    let closest = Infinity;
    for (const seed of SEEDS) {
      const roll = simulateRoll(config(6, seed, { dieCount: 3 }));
      for (let s = 0; s < roll.dice[0].samples.length; s++) {
        for (let a = 0; a < 3; a++) {
          for (let b = a + 1; b < 3; b++) {
            closest = Math.min(
              closest,
              mag(sub(roll.dice[a].samples[s].position, roll.dice[b].samples[s].position)),
            );
          }
        }
      }
    }
    // Two unit-circumradius dice cannot have their centres this close without
    // their hulls being in contact for at least part of the roll.
    assert.ok(closest < 1.9, `dice never came closer than ${closest} circumradii`);
  });

  it('starts three dice apart and lands them apart', () => {
    const roll = simulateRoll(config(6, 55, { dieCount: 3 }));
    const last = roll.dice.map((die) => die.samples[die.samples.length - 1].position);
    for (let a = 0; a < 3; a++) {
      for (let b = a + 1; b < 3; b++) {
        assert.ok(mag(sub(last[a], last[b])) > 1, `dice ${a} and ${b} finished on top of each other`);
      }
    }
  });
});

describe('simulateRoll retries the die that landed badly, not the throw', () => {
  /**
   * A whole-tray retry is a coin that has to come up heads for every die at
   * once, so its success rate is the single-die rate raised to the DIE COUNT
   * and it falls off a cliff as the tray fills up. Measured against the
   * all-or-nothing loop, d20 in a 6-cubed tray, 12 seeds:
   *
   *     dice   cocked           attempts        ms
   *        1    0/12 -> 0/12    1.08 -> 1.00      10 ->   11
   *        5    0/12 -> 0/12    2.00 -> 2.00      42 ->   50
   *       10    3/12 -> 0/12    4.83 -> 3.83     339 ->  280
   *       25   12/12 -> 12/12   8.00 -> 5.75    3379 -> 2349
   *
   * Ten dice is not an exotic configuration and the throw was already failing
   * at it. Twenty-five is a different problem and no retry policy fixes it:
   * they land in a heap, and a die resting on another die is cocked whatever
   * launch it was given. What the retry can do there is notice and stop, which
   * is where most of that case's saving comes from.
   */
  const TRAY = { x: 6, y: 6, z: 6 } as const;

  it('does not get likelier to hand back a cocked die as the tray fills', () => {
    // Up to five dice, which is as many as this tray can hold WITHOUT them
    // landing on each other; see the note above for why ten and twenty-five are
    // a different problem that no retry policy solves.
    for (const dieCount of [1, 5] as const) {
      for (let seed = 1; seed <= 12; seed++) {
        const roll = simulateRoll(config(20, seed, { dieCount, bounds: TRAY }));
        assert.equal(
          roll.cocked,
          false,
          `${dieCount} d20, seed ${seed}: came back cocked at ${roll.restAlignment}`,
        );
      }
    }
  });

  it('stops throwing a tray that is not getting any better', () => {
    // The cost assertion, in throws and physics steps rather than milliseconds
    // so it says the same thing on every machine. Twenty-five d20 in this tray
    // cannot be settled -- they land in a heap -- and the old loop spent all
    // eight throws finding that out on every single seed, 8.00 of a possible 8.
    // Stopping when the cocked count stops falling is most of that case's
    // saving; the bound below is loose because progress is noisy (a throw can
    // fix two dice and cock a third), and it is worth having anyway because
    // 8.00 out of 8, every time, is what it is measuring against.
    let attempts = 0;
    let steps = 0;
    const seeds = 8;
    for (let seed = 1; seed <= seeds; seed++) {
      const diagnostics = simulateRollWithDiagnostics(
        config(20, seed, { dieCount: 25, bounds: TRAY }),
      );
      assert.ok(diagnostics.trajectory.cocked, `seed ${seed}: 25 d20 settled; this test needs the hopeless case`);
      attempts += diagnostics.attempts;
      steps += diagnostics.simulatedSteps;
    }
    assert.ok(
      attempts / seeds < 7,
      `averaged ${attempts / seeds} throws on a tray that was never going to settle`,
    );
    // In steps as well as throws, against the ceiling the old loop actually
    // reached: eight throws, each free to run the whole 5-second cap.
    assert.ok(
      steps / seeds < 8 * 5 * 1080,
      `averaged ${steps / seeds} physics steps, which is a full eight-throw budget`,
    );
    // And the same accounting on a roll that settles first time, where the two
    // step counts must agree exactly or `simulatedSteps` is not measuring work.
    const one = simulateRollWithDiagnostics(config(20, 3, { dieCount: 1, bounds: TRAY }));
    assert.equal(one.attempts, 1);
    assert.equal(one.simulatedSteps, one.stepCount);
  });

  it('reports every die\'s alignment, which is what it retries on', () => {
    const diagnostics = simulateRollWithDiagnostics(config(20, 4, { dieCount: 3, bounds: TRAY }));
    assert.equal(diagnostics.dieAlignment.length, 3);
    assert.equal(diagnostics.restAlignment, Math.min(...diagnostics.dieAlignment));
    // Recomputed from the resting orientations rather than trusted, so the
    // per-die numbers the retry acts on are checked against this file's own
    // arithmetic and not merely against each other.
    const geometry = dieGeometry(20);
    for (let d = 0; d < 3; d++) {
      let best = -Infinity;
      for (const face of geometry.faces) {
        best = Math.max(
          best,
          dot(rotate(diagnostics.trajectory.dice[d].restingOrientation, face.normal), [0, -1, 0]),
        );
      }
      assert.ok(
        Math.abs(best - diagnostics.dieAlignment[d]) < 1e-12,
        `die ${d}: reported ${diagnostics.dieAlignment[d]}, measured ${best}`,
      );
    }
  });

  it('keeps the whole energy history, including the throws it threw away', () => {
    // A retried roll used to hand back only the chosen throw's energy trace, so
    // a rejected throw took its history with it -- and a real energy pump was
    // hiding in one (see `PENETRATION_SLOP` in the module). `attemptStarts`
    // marks where each throw's trace begins, because the invariant holds WITHIN
    // a throw and says nothing across the boundary, where a fresh launch is.
    const diagnostics = simulateRollWithDiagnostics(
      config(20, 3, { dieCount: 25, bounds: TRAY }),
    );
    assert.ok(diagnostics.attempts > 1, 'this test needs a roll that retried');
    assert.equal(diagnostics.attemptStarts.length, diagnostics.attempts);
    assert.equal(diagnostics.attemptStarts[0], 0);
    for (let i = 1; i < diagnostics.attemptStarts.length; i++) {
      assert.ok(
        diagnostics.attemptStarts[i] > diagnostics.attemptStarts[i - 1],
        'attempt traces are not in order',
      );
    }
    assert.ok(
      diagnostics.energyPerStep.length > diagnostics.stepCount,
      'energyPerStep covers only the throw that was kept',
    );
  });
});

describe('simulateRoll with a geometry per die', () => {
  /**
   * The whole point of the per-body thread: a throw may mix shapes.
   *
   * Every assertion here is made against the die's OWN solid, which is what a
   * single shared kernel cannot express. A simulator that quietly used the
   * first geometry for all three dice passes containment (the shapes are all
   * circumradius 1 after normalisation, so the box never notices) and fails
   * this, because a d20's vertices are not a d4's and a face that a d4 can be
   * read from is not a face of a d20.
   */
  const MIXED = [3, 4, 6, 20] as const;

  it('lands each die flat on a face of its own geometry', () => {
    const geometries = MIXED.map(dieGeometry);
    const cosLimit = Math.cos((MAX_COCK_DEGREES * Math.PI) / 180);
    for (const seed of SEEDS) {
      const roll = simulateRoll({
        seed,
        geometry: geometries,
        dieCount: geometries.length,
        bounds: BOUNDS,
      });
      assert.equal(roll.dice.length, geometries.length);
      for (let d = 0; d < geometries.length; d++) {
        const geometry = geometries[d];
        const sign = resolveReadingRule(geometry) === 'up-face' ? 1 : -1;
        const wanted: Vec3 = [WORLD_UP[0] * sign, WORLD_UP[1] * sign, WORLD_UP[2] * sign];
        const resting = roll.dice[d].restingOrientation;
        const face = geometry.faces[presentedFaceIndex(geometry, resting)];
        const alignment = dot(rotate(resting, face.normal), wanted);
        const degrees = (Math.acos(Math.min(1, Math.max(-1, alignment))) * 180) / Math.PI;
        assert.ok(
          alignment >= cosLimit,
          `seed ${seed} d${MIXED[d]}: landed ${degrees.toFixed(2)} degrees off its own face`,
        );
      }
    }
  });

  it('keeps every die inside the container using its own hull', () => {
    const geometries = MIXED.map(dieGeometry);
    const solids = geometries.map(simulationSolid);
    const planes = containerPlanes(BOUNDS);
    for (const seed of SEEDS) {
      const roll = simulateRoll({
        seed,
        geometry: geometries,
        dieCount: geometries.length,
        bounds: BOUNDS,
      });
      for (let d = 0; d < geometries.length; d++) {
        for (const sample of roll.dice[d].samples) {
          for (const vertex of worldVertices(solids[d].vertices, sample.position, sample.orientation)) {
            for (const plane of planes) {
              assert.ok(
                dot(plane.normal, vertex) - plane.offset <= CONTAINMENT_EPSILON,
                `seed ${seed} d${MIXED[d]} at t=${sample.t}: outside plane ${plane.normal}`,
              );
            }
          }
        }
      }
    }
  });

  it('keeps a mixed pair from interpenetrating, each against the other hull', () => {
    // The die-die narrow phase reads one body's VERTICES against the other
    // body's PLANES. With one shared kernel those always come from the same
    // solid and the asymmetry is invisible; with two shapes, getting either
    // side backwards buries a d4's corner in a d20.
    const geometries = [dieGeometry(4), dieGeometry(20)];
    const solids = geometries.map(simulationSolid);
    for (const seed of SEEDS) {
      const roll = simulateRoll({ seed, geometry: geometries, dieCount: 2, bounds: BOUNDS });
      for (let s = 0; s < roll.dice[0].samples.length; s++) {
        const poses = roll.dice.map((die) => die.samples[s]);
        for (const [a, b] of [[0, 1], [1, 0]] as const) {
          for (const vertex of worldVertices(solids[a].vertices, poses[a].position, poses[a].orientation)) {
            const local = rotate(
              [-poses[b].orientation[0], -poses[b].orientation[1], -poses[b].orientation[2], poses[b].orientation[3]],
              sub(vertex, poses[b].position),
            );
            let depth = Infinity;
            for (const plane of solids[b].planes) {
              depth = Math.min(depth, plane.offset - dot(plane.normal, local));
            }
            assert.ok(depth <= 0.01, `seed ${seed}: die ${a} is ${depth} inside die ${b} at t=${poses[a].t}`);
          }
        }
      }
    }
  });

  it('reproduces the single-geometry roll exactly when every die is the same shape', () => {
    // The compatibility contract. A config naming ONE geometry must not merely
    // still work, it must produce the identical trajectory it always did, or
    // every consumer's animation moves for no reason anyone asked for.
    for (const faceCount of SHAPES) {
      const geometry = dieGeometry(faceCount);
      const single = simulateRoll(config(faceCount, 4242, { dieCount: 3 }));
      const listed = simulateRoll({
        seed: 4242,
        geometry: [geometry, geometry, geometry],
        dieCount: 3,
        bounds: BOUNDS,
      });
      assert.deepStrictEqual(listed, single, `d${faceCount} diverged when listed per die`);
    }
  });

  it('refuses a geometry list that does not match dieCount', () => {
    const two = [dieGeometry(6), dieGeometry(20)];
    assert.throws(() => simulateRoll({ seed: 1, geometry: two, dieCount: 3, bounds: BOUNDS }));
    assert.throws(() => simulateRoll({ seed: 1, geometry: two, dieCount: 1, bounds: BOUNDS }));
    assert.throws(() => simulateRoll({ seed: 1, geometry: [], dieCount: 0, bounds: BOUNDS }));
  });
});

describe('simulateRoll physical plausibility', () => {
  it('lets a die fall, tumble and travel', () => {
    // Not a correctness bound, a sanity bound: a simulation that dropped the
    // angular integration or the initial throw would still satisfy every
    // assertion above while looking like a brick landing on a table.
    for (const faceCount of SHAPES) {
      const roll = simulateRoll(config(faceCount, 2024, { dieCount: 1 }));
      const samples = roll.dice[0].samples;
      let travelled = 0;
      let turned = 0;
      for (let i = 1; i < samples.length; i++) {
        travelled += mag(sub(samples[i].position, samples[i - 1].position));
        turned += quatAngle(samples[i - 1].orientation, samples[i].orientation);
      }
      assert.ok(travelled > 5, `d${faceCount} only travelled ${travelled}`);
      // At least one and a half tumbles. The launch spin is 3 to 7 turns per
      // second and flight is about half a second, so a healthy roll clears this
      // comfortably; a die that fell without spinning would score near zero.
      assert.ok(
        turned > 3 * Math.PI,
        `d${faceCount} only turned ${(turned / Math.PI).toFixed(2)} pi radians`,
      );
      const lowest = Math.min(...samples.map((sample) => sample.position[1]));
      const highest = Math.max(...samples.map((sample) => sample.position[1]));
      assert.ok(highest - lowest > 2, `d${faceCount} barely fell (${highest - lowest})`);
    }
  });

  it('rolls at a pace a player will sit through, in the tray the renderer uses', () => {
    /**
     * The feel contract, in the container it is actually felt in.
     *
     * Written out here rather than imported from `dice-roll.ts` for the same
     * reason `containerPlanes` is: this file checks the module against numbers
     * it owns. If the renderer's tray moves, this stops describing the shipped
     * roll and should be updated deliberately.
     */
    const TRAY = { x: 1.6, y: 2.0, z: 1.6 } as const;
    // Twenty seeds a shape. A single sample is worthless here — the spread
    // within one shape is wider than the gap between shapes — so every bound
    // below is on the MEDIAN and the seeds are fixed so it is reproducible.
    const seeds = 20;
    const median = (values: number[]): number => {
      const sorted = [...values].sort((a, b) => a - b);
      return sorted[Math.floor(sorted.length / 2)];
    };
    for (const faceCount of SHAPES) {
      const geometry = dieGeometry(faceCount);
      const durations: number[] = [];
      const turns: number[] = [];
      const dead: number[] = [];
      for (let seed = 1; seed <= seeds; seed++) {
        const roll = simulateRoll({ seed: seed * 7919, geometry, dieCount: 1, bounds: TRAY });
        const samples = roll.dice[0].samples;
        durations.push(roll.durationMs);
        let turned = 0;
        let still = 0;
        for (let i = 1; i < samples.length; i++) {
          const degrees = (quatAngle(samples[i - 1].orientation, samples[i].orientation) * 180) / Math.PI;
          turned += degrees;
          if (degrees < 1) still++;
        }
        turns.push(turned);
        dead.push(still / (samples.length - 1));
      }
      // A roll gates the whole animation cycle, so its length is a budget and
      // not a taste. Before the pace tuning the worst median was 1483 ms.
      assert.ok(
        median(durations) <= 800,
        `d${faceCount} takes a median ${median(durations).toFixed(0)}ms to say a number`,
      );
      // And it has to be a throw rather than a topple. One full turn is the
      // floor, not the goal: see `LAUNCH_SPIN_MIN` for the sampling ceiling
      // that stops this being the two-to-three turns a real die makes.
      assert.ok(
        median(turns) >= 360,
        `d${faceCount} turns a median ${median(turns).toFixed(0)} degrees, which is not a tumble`,
      );
      // Frames in which the die turns under a degree are frames the player is
      // waiting through. Worst median before the trim and the pace tuning: 37%.
      assert.ok(
        median(dead) <= 0.25,
        `d${faceCount} spends a median ${(median(dead) * 100).toFixed(0)}% of its frames not moving`,
      );
    }
  });

  it('makes the launch kinetic energy linear in the energy parameter', () => {
    // The spawn positions do not depend on `energy`, so the potential term of
    // the opening total is identical across these three rolls and the whole
    // difference is kinetic. Speeds scale as sqrt(energy), so that difference
    // must be exactly linear -- which pins the parameter's meaning rather than
    // merely observing that it does something.
    const launch = (energy: number): number =>
      simulateRollWithDiagnostics(config(6, 606, { energy })).energyPerStep[0];
    const low = launch(0.4);
    const mid = launch(1);
    const high = launch(2.5);
    const ratio = (high - mid) / (mid - low);
    assert.ok(
      Math.abs(ratio - (2.5 - 1) / (1 - 0.4)) < 1e-9,
      `launch energy is not linear in energy: ratio ${ratio}, expected 2.5`,
    );
    assert.ok(high > mid && mid > low, 'more energy did not mean a faster throw');
  });

  it('tumbles more when thrown harder', () => {
    const tumble = (energy: number): number => {
      let total = 0;
      for (const seed of SEEDS) {
        const samples = simulateRoll(config(6, seed, { energy })).dice[0].samples;
        for (let i = 1; i < samples.length; i++) {
          total += quatAngle(samples[i - 1].orientation, samples[i].orientation);
        }
      }
      return total;
    };
    assert.ok(tumble(3) > tumble(0.3) * 1.5, 'energy does not change how much the dice spin');
  });
});

