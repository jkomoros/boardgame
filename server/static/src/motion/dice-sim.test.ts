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

/** Rest thresholds for the settled tail, from sample-to-sample differences. */
const REST_SPEED = 0.1;
const REST_ANGULAR_SPEED = 0.2;
const REST_WINDOW_MS = 200;

/** A die balanced on an edge or a corner must fail this. */
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

  it('scales the inertia tensor with the geometry, so shapes tumble comparably', () => {
    // Unit-mass inertia goes as R^2, and the geometry module builds each solid
    // at its own natural scale. Compare traces of the INVERSE tensor, which is
    // what the integrator actually multiplies an impulse by.
    const trace = (t: readonly number[]): number => t[0] + t[4] + t[8];
    const raw = SHAPES.map((faceCount) => trace(dieGeometry(faceCount).inertiaTensor));
    const normalised = SHAPES.map((faceCount) =>
      trace(simulationSolid(dieGeometry(faceCount)).inverseInertia),
    );
    const spread = (values: number[]): number => Math.max(...values) / Math.min(...values);

    // Measured: 3.14 (d20) against 0.76 (d10). The 4x is what a die tumbling at
    // the wrong rate for its face count would come from, so this is the fact the
    // normalisation exists to cancel — pinned so that if `die-geometry.ts` ever
    // starts normalising itself, this stops being a silent no-op.
    assert.ok(
      spread(raw) > 3.5,
      `raw inertia traces span only ${spread(raw)}x; normalisation may be a no-op now`,
    );
    assert.ok(
      spread(normalised) < spread(raw) / 1.5,
      `normalising barely narrowed the spread: ${spread(raw)}x -> ${spread(normalised)}x`,
    );
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
          const tail = die.samples.filter(
            (sample) => sample.t >= roll.durationMs - REST_WINDOW_MS,
          );
          assert.ok(tail.length >= 8, `${label}: only ${tail.length} samples in the rest window`);
          for (let i = 1; i < tail.length; i++) {
            const dt = (tail[i].t - tail[i - 1].t) / 1000;
            const speed = mag(sub(tail[i].position, tail[i - 1].position)) / dt;
            const spin = quatAngle(tail[i - 1].orientation, tail[i].orientation) / dt;
            assert.ok(speed <= REST_SPEED, `${label}: still moving at ${speed} at t=${tail[i].t}`);
            assert.ok(
              spin <= REST_ANGULAR_SPEED,
              `${label}: still spinning at ${spin} rad/s at t=${tail[i].t}`,
            );
          }

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
            clearance < 1e-3,
            `d${faceCount} seed ${seed} die ${d} came to rest ${clearance} above the floor`,
          );
        }
      }
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
  const RELATIVE_STEP_TOLERANCE = 1e-5;

  for (const faceCount of SHAPES) {
    it(`never gains energy while rolling a d${faceCount}`, () => {
      for (const seed of SEEDS) {
        const diagnostics = simulateRollWithDiagnostics(config(faceCount, seed, { dieCount: 2 }));
        const energy = diagnostics.energyPerStep;
        assert.ok(energy.length > 100, `only ${energy.length} steps`);
        const budget = energy[0] * RELATIVE_STEP_TOLERANCE;
        for (let i = 1; i < energy.length; i++) {
          assert.ok(
            energy[i] <= energy[i - 1] + budget,
            `d${faceCount} seed ${seed} gained energy at step ${i}: ${energy[i - 1]} -> ${energy[i]} (budget ${budget})`,
          );
        }
        assert.ok(
          energy[energy.length - 1] < energy[0] * 0.25,
          `d${faceCount} seed ${seed} only dissipated to ${energy[energy.length - 1]} of ${energy[0]}`,
        );
      }
    });
  }

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

