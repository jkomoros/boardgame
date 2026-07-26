/**
 * Deterministic rigid-body simulation of dice tumbling inside a box.
 *
 * Pure arithmetic over `die-geometry.ts`: no DOM, no packages, no wall clock and
 * — critically — no `Math.random`. The renderer re-derives a roll from
 * `(component id, state version)` every time it mounts, so the same seed has to
 * reproduce the same trajectory bit for bit, forever. Everything stochastic here
 * comes out of a mulberry32 stream seeded from `RollConfig.seed`, and the only
 * transcendental in the whole module is `Math.sqrt`, which IEEE-754 requires to
 * be correctly rounded. Nothing depends on the platform's `sin`/`cos`/`exp`.
 *
 * ## Units, and the size-normalisation trap
 *
 * `die-geometry.ts` does NOT normalise its solids: circumradius runs from 1.000
 * (d8) to 1.902 (d20), and unit-mass inertia goes as R^2, so a d20's inertia
 * tensor is about 4x a d10's at the scale each is built at. Fed straight into
 * the angular dynamics that makes different face counts tumble at visibly
 * different rates for no reason a test would explain.
 *
 * This module normalises ON ENTRY, in `simulationSolid`: vertices and face
 * planes are divided by `circumradius` and the inertia tensor by
 * `circumradius^2`, so EVERY simulated die is a unit-circumradius, unit-mass
 * solid. One consequence for every consumer: all lengths in a `RollTrajectory`
 * — positions, and the `bounds` you hand in — are in units of the die's
 * circumradius, not pixels and not the geometry's native coordinates. A renderer
 * scales the whole trajectory by whatever it wants a die radius to be on screen.
 *
 * ## Frame
 *
 * Right-handed, +Y up, matching `die-faces.ts`'s `WORLD_UP`. CSS screen-Y points
 * DOWN, so a renderer composing a transform from these poses has to flip Y
 * itself; nothing here does it for it. The container is a box of HALF-extents
 * `bounds` centred on the origin, with a floor, a ceiling and four walls.
 *
 * ## Contact model
 *
 * Contacts are found by iterating the solid's ACTUAL VERTICES against the
 * container planes — never by special-casing a box, because the same code has to
 * settle a d20 and a 7-sided barrel. Die-against-die uses the same vertices
 * against the other die's face planes (a point-in-convex-polyhedron test), which
 * is why `capFaces` matters: a barrel's caps are part of its surface even though
 * they carry no value.
 *
 * Contacts are speculative: a vertex becomes a contact while it still has a gap,
 * and the solver's target separating velocity is `-gap/dt`, so the vertex is
 * allowed to close the gap this step and no further. That is what keeps a die
 * spinning at 30 rad/s from burying a corner in the floor between steps, and it
 * means depenetration is done by impulse rather than by teleporting the body
 * (which would silently add potential energy and break the energy invariant).
 *
 * ## Why the inner loop looks the way it does
 *
 * `die-geometry.ts`'s vector helpers return FROZEN tuples, which is right for
 * geometry built once and wrong for a solver that runs ten impulse iterations
 * over a dozen contacts at 480 Hz. Written with them, a two-die roll cost 100 ms
 * — a visible hitch on every mount. The integrator and the contact solver
 * therefore work on preallocated `Float64Array`s with the arithmetic spelled out
 * component by component, and allocate nothing per step; the helpers are still
 * used at the edges, where clarity is worth more than the nanoseconds.
 */

import {
  dot,
  scale,
  vec3,
  type DieGeometry,
  type Quat,
  type Vec3,
} from './die-geometry.ts';

// ---------------------------------------------------------------------------
// Public interface
// ---------------------------------------------------------------------------

export interface RollConfig {
  /** Any finite number. Equal seeds produce bitwise-equal trajectories. */
  readonly seed: number;
  readonly geometry: DieGeometry;
  readonly dieCount: number;
  /** HALF-extents of the container box, in die circumradii. See the file docs. */
  readonly bounds: { readonly x: number; readonly y: number; readonly z: number };
  /** Vigour of the throw; the launch kinetic energy is linear in it. Default 1. */
  readonly energy?: number;
  /** Downward acceleration, circumradii per second squared. Default 42. */
  readonly gravity?: number;
  /** Bounciness in [0, 1]. Default 0.32. */
  readonly restitution?: number;
  /** Coulomb friction coefficient, >= 0. Default 0.45. */
  readonly friction?: number;
}

export interface DieSample {
  /** Milliseconds from the start of the roll. */
  readonly t: number;
  readonly position: Vec3;
  readonly orientation: Quat;
}

export interface DieTrajectory {
  readonly samples: readonly DieSample[];
  /** Always the final sample's orientation; read it with `presentedFaceIndex`. */
  readonly restingOrientation: Quat;
}

export interface RollTrajectory {
  readonly durationMs: number;
  readonly dice: readonly DieTrajectory[];
}

/** A half-space of the container or of a die: inside means `n . p <= offset`. */
export interface SolidPlane {
  readonly normal: Vec3;
  readonly offset: number;
}

/**
 * A die's geometry rescaled to the simulator's units: circumradius exactly 1.
 *
 * Exported because it IS the normalisation contract — a test or a renderer that
 * wants to reconstruct world-space vertices from a `DieSample` must use these
 * vertices, not `geometry.vertices`, or it will be off by `circumradius`.
 */
export interface SimulationSolid {
  readonly vertices: readonly Vec3[];
  /** Every plane of the closed surface, readable faces and caps alike. */
  readonly planes: readonly SolidPlane[];
  /** Body-frame unit-mass inertia at unit circumradius, row-major 3x3. */
  readonly inertia: readonly number[];
  readonly inverseInertia: readonly number[];
}

/**
 * A roll plus the internals a test needs to prove the solver is behaving.
 * `simulateRoll` is this with the diagnostics dropped.
 */
export interface RollDiagnostics {
  readonly trajectory: RollTrajectory;
  /** Total mechanical energy of every die, after each physics step. */
  readonly energyPerStep: readonly number[];
  /**
   * The largest single positional nudge the containment clamp ever applied.
   * The clamp is a backstop for the residue the linearised contact constraint
   * leaves behind; if this is large, the impulse solver is not doing its job.
   */
  readonly maxWallCorrection: number;
  /** The deepest any vertex of one die ever reached inside another. */
  readonly maxDieOverlap: number;
  readonly stepCount: number;
  /**
   * How many throws it took to land every die flat on a readable face. See
   * `SETTLE_ALIGNMENT`; 1 for the overwhelming majority of rolls.
   */
  readonly attempts: number;
  /**
   * Worst `cos(tilt)` over the dice at rest, where tilt is the angle between
   * straight down and the nearest READABLE face normal. 1 is dead flat. Below
   * `SETTLE_ALIGNMENT` means the roll ran out of attempts and the returned
   * trajectory is the least-cocked one found.
   */
  readonly restAlignment: number;
}

// ---------------------------------------------------------------------------
// Tuning
// ---------------------------------------------------------------------------

/**
 * 720 Hz. The step has to be short compared with `1 / angularSpeed`: the
 * contact constraint is linear in velocity while a vertex on a spinning die
 * travels an ARC, so a corner dips below the floor between solves by about
 * `(w dt)^2 r / 2`. Spin peaks near 60 rad/s just after a corner impact, which
 * is 3.5e-3 of a circumradius here and 8e-3 at 480 Hz — measurably worse, and
 * the positional clamp that cleans it up is the one part of the step that is
 * not physics.
 */
const PHYSICS_HZ = 720;
/** Samples land on 60 Hz frame boundaries, which is what the renderer wants. */
const SAMPLE_HZ = 60;
const STEPS_PER_SAMPLE = PHYSICS_HZ / SAMPLE_HZ;
const STEP_SECONDS = 1 / PHYSICS_HZ;

const MAX_SECONDS = 5;
/** Gauss-Seidel passes over the contact set. A resting cube needs about 4. */
const SOLVER_ITERATIONS = 10;

const DEFAULT_ENERGY = 1;
const DEFAULT_GRAVITY = 42;
const DEFAULT_RESTITUTION = 0.32;
const DEFAULT_FRICTION = 0.45;

/** Air drag, per second. Small: friction does the real work. */
const LINEAR_DRAG = 0.3;
const ANGULAR_DRAG = 0.9;
/**
 * Rolling resistance, applied only while a die is touching something. Modest,
 * and honestly so: measured over 60 seeds it moves the MEDIAN roll by under
 * 100 ms and trims about 15% off the 95th-percentile tail of the rounder solids
 * (a d12's p95 goes 3333 ms -> 2850 ms). Coulomb friction alone does settle
 * them; this stops the longest rolls from dragging.
 */
const CONTACT_ANGULAR_DRAG = 7;

/** A vertex this close to a surface is a contact candidate even at rest. */
const CONTACT_MARGIN = 0.02;
/**
 * Below this approach speed a contact is treated as resting and gets no
 * restitution. Bouncing a die that is settling is both unphysical (a real die
 * has no coefficient of restitution at 0.1 mm/s) and the classic way a
 * sequential-impulse solver jitters forever instead of coming to rest.
 */
const RESTITUTION_THRESHOLD = 1.2;
/** Cap on how much penetration one step's impulse is allowed to undo. */
const MAX_DEPENETRATION = 0.02;

/** Sustained speeds below these, in contact, count as at rest. */
const REST_LINEAR = 0.05;
const REST_ANGULAR = 0.09;
/** How long that has to hold before the roll is over. */
const REST_HOLD_SECONDS = 0.3;
const REST_HOLD_STEPS = Math.round(REST_HOLD_SECONDS * PHYSICS_HZ);

/**
 * A die counts as landed when some READABLE face normal is within this of
 * straight down, i.e. it is lying flat on a face it can be read from.
 *
 * Three degrees. Deliberately tighter than the five the suite asserts, so the
 * assertion keeps margin to bite with; a die balanced on an edge (45 degrees
 * for a cube), leaning on a wall, or — for a barrel — sitting on one of its
 * unreadable cap facets (61 degrees for a d7) is nowhere near either number.
 *
 * Resting-on-a-readable-face is used rather than presenting-one because it is
 * the same criterion for all three reading rules: an `up-face` die presents the
 * antipode of the face it rests on, and a d4 and an odd barrel present the
 * resting face itself. It also needs no reading rule, so this module does not
 * have to import `die-faces.ts` and the suite's check through
 * `presentedFaceIndex` stays an independent expression of the same fact.
 */
const SETTLE_ALIGNMENT = Math.cos((3 * Math.PI) / 180);
/**
 * A cocked die is re-thrown, exactly as at a real table. Deterministic: each
 * attempt's seed is drawn from a stream rooted at `config.seed`, so the whole
 * retried roll still reproduces bit for bit. If every attempt lands cocked the
 * least-cocked one is returned rather than a throw — a slightly ugly die beats
 * no animation — and `RollDiagnostics.restAlignment` says so.
 */
const MAX_ATTEMPTS = 8;

/** Spawn grid pitch, in circumradii: two unit dice need more than 2 apart. */
const SPAWN_PITCH = 2.4;
/** Clearance kept between a spawned die and the container walls. */
const SPAWN_CLEARANCE = 1.2;
/**
 * Launch speeds at `energy: 1`. Scaled by `sqrt(energy)`, so KE is linear.
 *
 * The spin range is 20 to 44 rad/s, i.e. 3 to 7 turns per second, which over a
 * typical half-second of flight is the two to four visible tumbles a thrown die
 * makes. It was originally a third lower and the dice read as dropped rather
 * than thrown — no assertion here can see that, so the number is set from what
 * a die does and not from what makes the suite pass.
 */
const LAUNCH_HORIZONTAL = 5.5;
const LAUNCH_SPIN_MIN = 20;
const LAUNCH_SPIN_RANGE = 24;

// ---------------------------------------------------------------------------
// Seeded randomness
// ---------------------------------------------------------------------------

/**
 * mulberry32. Chosen because its whole state is one uint32 and every operation
 * is integer, so it reproduces exactly on any conforming engine — which is the
 * entire point of having a PRNG here instead of `Math.random`.
 *
 * The seed is avalanched through a splitmix32 finaliser first. Without it,
 * consecutive seeds differ by one bit of state and mulberry32's FIRST few
 * outputs stay correlated — which here means seed 1 and seed 2 throw their dice
 * in nearly the same direction. Every seed-dependent quantity in this module is
 * drawn in the first dozen outputs, so that is exactly the regime that matters.
 */
function createRandom(seed: number): () => number {
  let state = (Math.trunc(seed) ^ 0x9e3779b9) >>> 0;
  state = Math.imul(state ^ (state >>> 16), 0x21f0aaad) >>> 0;
  state = Math.imul(state ^ (state >>> 15), 0x735a2d97) >>> 0;
  state = (state ^ (state >>> 15)) >>> 0;
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let t = state;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** Uniform in [-1, 1). */
function signed(random: () => number): number {
  return random() * 2 - 1;
}

/**
 * A uniformly distributed unit vector written into `out`, by rejection sampling
 * the unit ball and normalising. Deliberately not `(cos, sin)` of a random
 * angle: rejection needs only `sqrt`, and `sqrt` is the one transcendental
 * IEEE-754 pins exactly, so the stream survives a change of JavaScript engine.
 */
function randomUnitVector(random: () => number, out: Float64Array, offset: number): void {
  for (;;) {
    const x = signed(random);
    const y = signed(random);
    const z = signed(random);
    const squared = x * x + y * y + z * z;
    if (squared > 1e-6 && squared <= 1) {
      const inverse = 1 / Math.sqrt(squared);
      out[offset] = x * inverse;
      out[offset + 1] = y * inverse;
      out[offset + 2] = z * inverse;
      return;
    }
  }
}

/** A uniformly distributed orientation: the 4-D version of the same trick. */
function randomUnitQuat(random: () => number, out: Float64Array): void {
  for (;;) {
    const x = signed(random);
    const y = signed(random);
    const z = signed(random);
    const w = signed(random);
    const squared = x * x + y * y + z * z + w * w;
    if (squared > 1e-6 && squared <= 1) {
      const inverse = 1 / Math.sqrt(squared);
      out[0] = x * inverse;
      out[1] = y * inverse;
      out[2] = z * inverse;
      out[3] = w * inverse;
      return;
    }
  }
}

// ---------------------------------------------------------------------------
// Matrices (row-major 3x3) and quaternions (x, y, z, w)
// ---------------------------------------------------------------------------

function invertMatrix(m: readonly number[]): number[] {
  const c00 = m[4] * m[8] - m[5] * m[7];
  const c01 = m[5] * m[6] - m[3] * m[8];
  const c02 = m[3] * m[7] - m[4] * m[6];
  const determinant = m[0] * c00 + m[1] * c01 + m[2] * c02;
  if (!(Math.abs(determinant) > 0)) {
    throw new Error('cannot invert a singular inertia tensor');
  }
  const inverse = 1 / determinant;
  return [
    c00 * inverse,
    (m[2] * m[7] - m[1] * m[8]) * inverse,
    (m[1] * m[5] - m[2] * m[4]) * inverse,
    c01 * inverse,
    (m[0] * m[8] - m[2] * m[6]) * inverse,
    (m[2] * m[3] - m[0] * m[5]) * inverse,
    c02 * inverse,
    (m[1] * m[6] - m[0] * m[7]) * inverse,
    (m[0] * m[4] - m[1] * m[3]) * inverse,
  ];
}

/** `out = R(q)`, the body-to-world rotation of a unit quaternion. */
function quatMatrixInto(q: ArrayLike<number>, out: Float64Array): void {
  const x = q[0];
  const y = q[1];
  const z = q[2];
  const w = q[3];
  out[0] = 1 - 2 * (y * y + z * z);
  out[1] = 2 * (x * y - z * w);
  out[2] = 2 * (x * z + y * w);
  out[3] = 2 * (x * y + z * w);
  out[4] = 1 - 2 * (x * x + z * z);
  out[5] = 2 * (y * z - x * w);
  out[6] = 2 * (x * z - y * w);
  out[7] = 2 * (y * z + x * w);
  out[8] = 1 - 2 * (x * x + y * y);
}

/** `out = a * b * a^T`, the world-frame form of a body-frame tensor. */
function conjugateInto(a: Float64Array, b: ArrayLike<number>, out: Float64Array): void {
  // t = a * b
  const t = CONJUGATE_SCRATCH;
  for (let row = 0; row < 3; row++) {
    for (let column = 0; column < 3; column++) {
      t[row * 3 + column] =
        a[row * 3] * b[column] + a[row * 3 + 1] * b[3 + column] + a[row * 3 + 2] * b[6 + column];
    }
  }
  // out = t * a^T
  for (let row = 0; row < 3; row++) {
    for (let column = 0; column < 3; column++) {
      out[row * 3 + column] =
        t[row * 3] * a[column * 3] +
        t[row * 3 + 1] * a[column * 3 + 1] +
        t[row * 3 + 2] * a[column * 3 + 2];
    }
  }
}
const CONJUGATE_SCRATCH = new Float64Array(9);

// ---------------------------------------------------------------------------
// Normalised geometry
// ---------------------------------------------------------------------------

/**
 * `geometry` rescaled so its circumradius is exactly 1.
 *
 * See the file docs: this is where the size trap is defused. Lengths divide by
 * `circumradius`; the unit-mass inertia tensor, whose entries are second
 * moments of length, divides by `circumradius^2`. Face NORMALS are invariant
 * under a uniform scale, so they carry through untouched — and so, usefully,
 * does `presentedFaceIndex`, which reads only normals and argmax/argmin over
 * uniformly scaled distances. Reading a resting orientation therefore gives the
 * same answer against the original geometry as against this one.
 */
export function simulationSolid(geometry: DieGeometry): SimulationSolid {
  const radius = geometry.circumradius;
  if (!(radius > 0)) throw new Error('die geometry has a non-positive circumradius');
  const inverse = 1 / radius;
  const vertices = geometry.vertices.map((vertex) => scale(vertex, inverse));
  const planes = [...geometry.faces, ...geometry.capFaces].map((face) =>
    Object.freeze({
      normal: face.normal,
      offset: dot(face.normal, face.centroid) * inverse,
    }),
  );
  const inertia = geometry.inertiaTensor.map((entry) => entry * inverse * inverse);
  return Object.freeze({
    vertices: Object.freeze(vertices),
    planes: Object.freeze(planes),
    inertia: Object.freeze(inertia),
    inverseInertia: Object.freeze(invertMatrix(inertia)),
  });
}

/** The same solid flattened into typed arrays, for the allocation-free loop. */
interface Kernel {
  readonly vertexCount: number;
  /** 3 per vertex. */
  readonly vertices: Float64Array;
  readonly planeCount: number;
  /** 3 per plane. */
  readonly planeNormals: Float64Array;
  readonly planeOffsets: Float64Array;
  readonly inertia: Float64Array;
  readonly inverseInertia: Float64Array;
}

function kernelOf(solid: SimulationSolid): Kernel {
  const vertices = new Float64Array(solid.vertices.length * 3);
  solid.vertices.forEach((vertex, index) => vertices.set(vertex, index * 3));
  const planeNormals = new Float64Array(solid.planes.length * 3);
  const planeOffsets = new Float64Array(solid.planes.length);
  solid.planes.forEach((plane, index) => {
    planeNormals.set(plane.normal, index * 3);
    planeOffsets[index] = plane.offset;
  });
  return {
    vertexCount: solid.vertices.length,
    vertices,
    planeCount: solid.planes.length,
    planeNormals,
    planeOffsets,
    inertia: Float64Array.from(solid.inertia),
    inverseInertia: Float64Array.from(solid.inverseInertia),
  };
}

// ---------------------------------------------------------------------------
// Bodies
// ---------------------------------------------------------------------------

interface Body {
  readonly position: Float64Array;
  readonly velocity: Float64Array;
  readonly angular: Float64Array;
  readonly orientation: Float64Array;
  readonly rotation: Float64Array;
  readonly inverseInertiaWorld: Float64Array;
  /** 3 per vertex, world frame. */
  readonly worldVertices: Float64Array;
  touching: boolean;
}

function createBody(kernel: Kernel): Body {
  return {
    position: new Float64Array(3),
    velocity: new Float64Array(3),
    angular: new Float64Array(3),
    orientation: Float64Array.from([0, 0, 0, 1]),
    rotation: new Float64Array(9),
    inverseInertiaWorld: new Float64Array(9),
    worldVertices: new Float64Array(kernel.vertexCount * 3),
    touching: false,
  };
}

/** Recompute everything derived from `position` and `orientation`. */
function refresh(body: Body, kernel: Kernel): void {
  quatMatrixInto(body.orientation, body.rotation);
  conjugateInto(body.rotation, kernel.inverseInertia, body.inverseInertiaWorld);
  const r = body.rotation;
  const v = kernel.vertices;
  const w = body.worldVertices;
  for (let i = 0; i < kernel.vertexCount; i++) {
    const x = v[i * 3];
    const y = v[i * 3 + 1];
    const z = v[i * 3 + 2];
    w[i * 3] = body.position[0] + r[0] * x + r[1] * y + r[2] * z;
    w[i * 3 + 1] = body.position[1] + r[3] * x + r[4] * y + r[5] * z;
    w[i * 3 + 2] = body.position[2] + r[6] * x + r[7] * y + r[8] * z;
  }
}

/** `q += dt/2 * (omega (x) q)`, renormalised: the standard spin integrator. */
function integrateOrientation(body: Body, dt: number): void {
  const q = body.orientation;
  const wx = body.angular[0];
  const wy = body.angular[1];
  const wz = body.angular[2];
  const half = dt * 0.5;
  const x = q[0] + half * (wx * q[3] + wy * q[2] - wz * q[1]);
  const y = q[1] + half * (-wx * q[2] + wy * q[3] + wz * q[0]);
  const z = q[2] + half * (wx * q[1] - wy * q[0] + wz * q[3]);
  const w = q[3] + half * (-wx * q[0] - wy * q[1] - wz * q[2]);
  const length = Math.sqrt(x * x + y * y + z * z + w * w);
  if (!(length > 0)) throw new Error('orientation collapsed to a zero quaternion');
  q[0] = x / length;
  q[1] = y / length;
  q[2] = z / length;
  q[3] = w / length;
}

/** `n . ((I^-1 (r x n)) x r)`: the angular half of a contact's effective mass. */
function angularMass(
  inverse: Float64Array,
  rx: number,
  ry: number,
  rz: number,
  nx: number,
  ny: number,
  nz: number,
): number {
  const cx = ry * nz - rz * ny;
  const cy = rz * nx - rx * nz;
  const cz = rx * ny - ry * nx;
  const ix = inverse[0] * cx + inverse[1] * cy + inverse[2] * cz;
  const iy = inverse[3] * cx + inverse[4] * cy + inverse[5] * cz;
  const iz = inverse[6] * cx + inverse[7] * cy + inverse[8] * cz;
  return nx * (iy * rz - iz * ry) + ny * (iz * rx - ix * rz) + nz * (ix * ry - iy * rx);
}

function applyImpulse(
  body: Body,
  rx: number,
  ry: number,
  rz: number,
  px: number,
  py: number,
  pz: number,
): void {
  body.velocity[0] += px;
  body.velocity[1] += py;
  body.velocity[2] += pz;
  const cx = ry * pz - rz * py;
  const cy = rz * px - rx * pz;
  const cz = rx * py - ry * px;
  const inverse = body.inverseInertiaWorld;
  body.angular[0] += inverse[0] * cx + inverse[1] * cy + inverse[2] * cz;
  body.angular[1] += inverse[3] * cx + inverse[4] * cy + inverse[5] * cz;
  body.angular[2] += inverse[6] * cx + inverse[7] * cy + inverse[8] * cz;
}

// ---------------------------------------------------------------------------
// Contacts
// ---------------------------------------------------------------------------

/**
 * One vertex against one plane. Mutable and pooled: a busy step generates
 * dozens of these and they are rebuilt every step, so allocating them would
 * dominate the cost of the whole simulation.
 */
class Contact {
  a: Body = null as unknown as Body;
  /** `null` for a container wall, which is immovable. */
  b: Body | null = null;
  /** Contact point relative to each body's centre. */
  rax = 0;
  ray = 0;
  raz = 0;
  rbx = 0;
  rby = 0;
  rbz = 0;
  /** Unit, pointing out of the wall (or out of `b`) and into `a`. */
  nx = 0;
  ny = 0;
  nz = 0;
  /** Positive is a gap, negative is penetration. */
  separation = 0;
  /**
   * Separating speed the bounce wants, computed before any impulse — and
   * `-Infinity` when this contact is not an impact.
   *
   * NOT zero. A speculative contact is created while the vertex still has a
   * gap, and its non-penetration target is the NEGATIVE speed `-gap/dt`; a
   * floor of zero would forbid closing the gap at all and every die would come
   * to rest hovering exactly one `CONTACT_MARGIN` above the floor, wedgeable
   * into a corner at any angle. That bug is invisible in a containment test —
   * a hovering die is very definitely contained.
   */
  restitutionTarget = 0;
  effectiveMass = 0;
  normalImpulse = 0;
}

/** Relative velocity of the contact point, `a` with respect to `b`. */
const RELATIVE = new Float64Array(3);
function relativeVelocity(contact: Contact): void {
  const a = contact.a;
  const wa = a.angular;
  RELATIVE[0] = a.velocity[0] + wa[1] * contact.raz - wa[2] * contact.ray;
  RELATIVE[1] = a.velocity[1] + wa[2] * contact.rax - wa[0] * contact.raz;
  RELATIVE[2] = a.velocity[2] + wa[0] * contact.ray - wa[1] * contact.rax;
  const b = contact.b;
  if (b === null) return;
  const wb = b.angular;
  RELATIVE[0] -= b.velocity[0] + wb[1] * contact.rbz - wb[2] * contact.rby;
  RELATIVE[1] -= b.velocity[1] + wb[2] * contact.rbx - wb[0] * contact.rbz;
  RELATIVE[2] -= b.velocity[2] + wb[0] * contact.rby - wb[1] * contact.rbx;
}

function fillContact(
  contact: Contact,
  a: Body,
  b: Body | null,
  px: number,
  py: number,
  pz: number,
  nx: number,
  ny: number,
  nz: number,
  separation: number,
  restitution: number,
): void {
  contact.a = a;
  contact.b = b;
  contact.rax = px - a.position[0];
  contact.ray = py - a.position[1];
  contact.raz = pz - a.position[2];
  contact.rbx = b ? px - b.position[0] : 0;
  contact.rby = b ? py - b.position[1] : 0;
  contact.rbz = b ? pz - b.position[2] : 0;
  contact.nx = nx;
  contact.ny = ny;
  contact.nz = nz;
  contact.separation = separation;
  contact.normalImpulse = 0;
  relativeVelocity(contact);
  const approach = RELATIVE[0] * nx + RELATIVE[1] * ny + RELATIVE[2] * nz;
  contact.restitutionTarget =
    separation < CONTACT_MARGIN && approach < -RESTITUTION_THRESHOLD
      ? -approach * restitution
      : -Infinity;
  contact.effectiveMass =
    (b ? 2 : 1) +
    angularMass(a.inverseInertiaWorld, contact.rax, contact.ray, contact.raz, nx, ny, nz) +
    (b
      ? angularMass(b.inverseInertiaWorld, contact.rbx, contact.rby, contact.rbz, nx, ny, nz)
      : 0);
}

function solveContact(contact: Contact, friction: number, dt: number): void {
  const { a, b, nx, ny, nz } = contact;
  relativeVelocity(contact);
  const separating = RELATIVE[0] * nx + RELATIVE[1] * ny + RELATIVE[2] * nz;

  // Non-penetration: after `dt` the gap must not have gone negative. When it
  // already has, undo at most `MAX_DEPENETRATION` this step so a deep overlap
  // resolves smoothly instead of firing the body out.
  const allowed =
    contact.separation >= 0
      ? -contact.separation / dt
      : Math.min(-contact.separation, MAX_DEPENETRATION) / dt;
  const target = Math.max(allowed, contact.restitutionTarget);

  const change = (target - separating) / contact.effectiveMass;
  const total = Math.max(0, contact.normalImpulse + change);
  const applied = total - contact.normalImpulse;
  contact.normalImpulse = total;
  if (applied !== 0) {
    applyImpulse(a, contact.rax, contact.ray, contact.raz, nx * applied, ny * applied, nz * applied);
    if (b) {
      applyImpulse(
        b,
        contact.rbx,
        contact.rby,
        contact.rbz,
        -nx * applied,
        -ny * applied,
        -nz * applied,
      );
    }
  }

  if (contact.normalImpulse <= 0) return;

  // Coulomb friction, opposing whatever sliding is left after the normal solve.
  // The magnitude is capped both by what would zero the sliding and by
  // `friction * normalImpulse`, so it can never reverse the slide and can never
  // add energy.
  relativeVelocity(contact);
  const along = RELATIVE[0] * nx + RELATIVE[1] * ny + RELATIVE[2] * nz;
  const tx = RELATIVE[0] - along * nx;
  const ty = RELATIVE[1] - along * ny;
  const tz = RELATIVE[2] - along * nz;
  const sliding = Math.sqrt(tx * tx + ty * ty + tz * tz);
  if (sliding < 1e-9) return;
  const ux = tx / sliding;
  const uy = ty / sliding;
  const uz = tz / sliding;
  const tangentMass =
    (b ? 2 : 1) +
    angularMass(a.inverseInertiaWorld, contact.rax, contact.ray, contact.raz, ux, uy, uz) +
    (b
      ? angularMass(b.inverseInertiaWorld, contact.rbx, contact.rby, contact.rbz, ux, uy, uz)
      : 0);
  const magnitude = Math.min(sliding / tangentMass, friction * contact.normalImpulse);
  applyImpulse(
    a,
    contact.rax,
    contact.ray,
    contact.raz,
    -ux * magnitude,
    -uy * magnitude,
    -uz * magnitude,
  );
  if (b) {
    applyImpulse(
      b,
      contact.rbx,
      contact.rby,
      contact.rbz,
      ux * magnitude,
      uy * magnitude,
      uz * magnitude,
    );
  }
}

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------

interface Options {
  readonly energy: number;
  readonly gravity: number;
  readonly restitution: number;
  readonly friction: number;
}

function resolved(config: RollConfig): Options {
  const energy = config.energy ?? DEFAULT_ENERGY;
  const gravity = config.gravity ?? DEFAULT_GRAVITY;
  const restitution = config.restitution ?? DEFAULT_RESTITUTION;
  const friction = config.friction ?? DEFAULT_FRICTION;
  if (!Number.isFinite(config.seed)) throw new Error(`roll seed must be finite, got ${config.seed}`);
  if (!Number.isInteger(config.dieCount) || config.dieCount < 1) {
    throw new Error(`dieCount must be a positive integer, got ${config.dieCount}`);
  }
  for (const axis of ['x', 'y', 'z'] as const) {
    const half = config.bounds[axis];
    if (!Number.isFinite(half) || half < SPAWN_CLEARANCE + 0.3) {
      throw new Error(
        `bounds.${axis} must be a half-extent of at least ${SPAWN_CLEARANCE + 0.3} circumradii, got ${half}`,
      );
    }
  }
  if (!Number.isFinite(energy) || energy <= 0) {
    throw new Error(`energy must be positive, got ${energy}`);
  }
  if (!Number.isFinite(gravity) || gravity <= 0) {
    throw new Error(`gravity must be positive, got ${gravity}`);
  }
  if (!Number.isFinite(restitution) || restitution < 0 || restitution > 1) {
    throw new Error(`restitution must be in [0, 1], got ${restitution}`);
  }
  if (!Number.isFinite(friction) || friction < 0) {
    throw new Error(`friction must be non-negative, got ${friction}`);
  }
  return { energy, gravity, restitution, friction };
}

/**
 * Dice start on a grid near the ceiling, one pitch apart so they never begin
 * interpenetrating, with a randomised orientation, spin and throw direction.
 * The grid wraps into a second row when the container is too narrow.
 */
function spawnBodies(
  config: RollConfig,
  kernel: Kernel,
  energy: number,
  random: () => number,
): Body[] {
  const { bounds, dieCount } = config;
  const usable = 2 * bounds.x - 2 * SPAWN_CLEARANCE;
  const columns = Math.max(1, Math.floor(usable / SPAWN_PITCH) + 1);
  const rows = Math.ceil(dieCount / columns);
  const lowest = bounds.y - SPAWN_CLEARANCE - (rows - 1) * SPAWN_PITCH;
  if (lowest < -bounds.y + SPAWN_CLEARANCE) {
    throw new Error(`container is too small to spawn ${dieCount} dice without overlap`);
  }
  // Speeds go as sqrt(energy) so that the LAUNCH KINETIC ENERGY is linear in
  // `energy`, which is what the name promises and what the suite pins.
  const vigour = Math.sqrt(energy);
  const bodies: Body[] = [];
  for (let i = 0; i < dieCount; i++) {
    const body = createBody(kernel);
    const column = i % columns;
    const row = Math.floor(i / columns);
    body.position[0] = (column - (columns - 1) / 2) * SPAWN_PITCH + signed(random) * 0.12;
    body.position[1] = bounds.y - SPAWN_CLEARANCE - row * SPAWN_PITCH;
    body.position[2] = signed(random) * Math.min(0.5, bounds.z - SPAWN_CLEARANCE);
    body.velocity[0] = signed(random) * LAUNCH_HORIZONTAL * vigour;
    body.velocity[1] = -(1 + 3 * random()) * vigour;
    body.velocity[2] = signed(random) * LAUNCH_HORIZONTAL * vigour;
    randomUnitQuat(random, body.orientation);
    randomUnitVector(random, body.angular, 0);
    const spin = (LAUNCH_SPIN_MIN + LAUNCH_SPIN_RANGE * random()) * vigour;
    body.angular[0] *= spin;
    body.angular[1] *= spin;
    body.angular[2] *= spin;
    refresh(body, kernel);
    bodies.push(body);
  }
  return bodies;
}

// ---------------------------------------------------------------------------
// The loop
// ---------------------------------------------------------------------------

/**
 * `cos` of the angle between straight down and the nearest READABLE face normal
 * of a die in this orientation. 1 means lying dead flat on a face it can be
 * read from; a cube on an edge scores `cos 45deg`, and a barrel on a cap facet
 * — a face that carries no value — scores about `cos 61deg`.
 */
function readableRestAlignment(geometry: DieGeometry, orientation: Quat): number {
  const rotation = new Float64Array(9);
  quatMatrixInto(orientation, rotation);
  let best = -Infinity;
  for (const face of geometry.faces) {
    // The Y row of the rotation, negated: the downward component of R * normal.
    const downward = -(
      rotation[3] * face.normal[0] +
      rotation[4] * face.normal[1] +
      rotation[5] * face.normal[2]
    );
    if (downward > best) best = downward;
  }
  return best;
}

/**
 * Total mechanical energy, unit mass per die, with the container floor as the
 * potential-energy datum. Never used by the simulation itself — it exists so the
 * suite can assert that no step ever gains energy, which is how an unstable
 * impulse solver announces itself long before anything visibly explodes.
 */
const INERTIA_WORLD = new Float64Array(9);
function totalEnergy(
  bodies: readonly Body[],
  kernel: Kernel,
  gravity: number,
  floor: number,
): number {
  let total = 0;
  for (const body of bodies) {
    const v = body.velocity;
    const w = body.angular;
    conjugateInto(body.rotation, kernel.inertia, INERTIA_WORLD);
    const lx = INERTIA_WORLD[0] * w[0] + INERTIA_WORLD[1] * w[1] + INERTIA_WORLD[2] * w[2];
    const ly = INERTIA_WORLD[3] * w[0] + INERTIA_WORLD[4] * w[1] + INERTIA_WORLD[5] * w[2];
    const lz = INERTIA_WORLD[6] * w[0] + INERTIA_WORLD[7] * w[1] + INERTIA_WORLD[8] * w[2];
    total +=
      0.5 * (v[0] * v[0] + v[1] * v[1] + v[2] * v[2]) +
      0.5 * (w[0] * lx + w[1] * ly + w[2] * lz) +
      gravity * (body.position[1] - floor);
  }
  return total;
}

/** One throw. `simulateRollWithDiagnostics` may run several and keep the best. */
function throwOnce(
  config: RollConfig,
  options: Options,
  kernel: Kernel,
  random: () => number,
): RollDiagnostics {
  const { energy, gravity, restitution, friction } = options;
  const bodies = spawnBodies(config, kernel, energy, random);
  const bounds = config.bounds;
  const floor = -bounds.y;
  const dt = STEP_SECONDS;

  // The six container half-spaces, `n . p <= offset`, so `normal` points OUT of
  // the box and a contact impulse is applied along its negation.
  const wallNormals = new Float64Array([1, 0, 0, -1, 0, 0, 0, 1, 0, 0, -1, 0, 0, 0, 1, 0, 0, -1]);
  const wallOffsets = new Float64Array([
    bounds.x, bounds.x, bounds.y, bounds.y, bounds.z, bounds.z,
  ]);

  const linearDrag = Math.max(0, 1 - LINEAR_DRAG * dt);
  const angularDrag = Math.max(0, 1 - ANGULAR_DRAG * dt);
  const contactDrag = Math.max(0, 1 - CONTACT_ANGULAR_DRAG * dt);

  const pool: Contact[] = [];
  let used = 0;
  const take = (): Contact => {
    if (used === pool.length) pool.push(new Contact());
    return pool[used++];
  };

  const tracks: DieSample[][] = bodies.map(() => []);
  const energyPerStep: number[] = [];
  let maxWallCorrection = 0;
  let maxDieOverlap = 0;

  const record = (index: number): void => {
    const t = (index * 1000) / SAMPLE_HZ;
    for (let i = 0; i < bodies.length; i++) {
      const body = bodies[i];
      tracks[i].push(
        Object.freeze({
          t,
          position: vec3(body.position[0], body.position[1], body.position[2]),
          orientation: Object.freeze([
            body.orientation[0],
            body.orientation[1],
            body.orientation[2],
            body.orientation[3],
          ] as const) as Quat,
        }),
      );
    }
  };

  record(0);
  energyPerStep.push(totalEnergy(bodies, kernel, gravity, floor));

  const maxSteps = Math.round(MAX_SECONDS * PHYSICS_HZ);
  let restSteps = 0;
  let step = 0;
  while (step < maxSteps) {
    for (const body of bodies) {
      body.velocity[1] -= gravity * dt;
      body.velocity[0] *= linearDrag;
      body.velocity[1] *= linearDrag;
      body.velocity[2] *= linearDrag;
      body.angular[0] *= angularDrag;
      body.angular[1] *= angularDrag;
      body.angular[2] *= angularDrag;
      body.touching = false;
    }

    used = 0;
    for (const body of bodies) {
      const w = body.worldVertices;
      for (let v = 0; v < kernel.vertexCount; v++) {
        const px = w[v * 3];
        const py = w[v * 3 + 1];
        const pz = w[v * 3 + 2];
        const rx = px - body.position[0];
        const ry = py - body.position[1];
        const rz = pz - body.position[2];
        const a = body.angular;
        const vx = body.velocity[0] + a[1] * rz - a[2] * ry;
        const vy = body.velocity[1] + a[2] * rx - a[0] * rz;
        const vz = body.velocity[2] + a[0] * ry - a[1] * rx;
        for (let p = 0; p < 6; p++) {
          const nx = wallNormals[p * 3];
          const ny = wallNormals[p * 3 + 1];
          const nz = wallNormals[p * 3 + 2];
          const gap = wallOffsets[p] - (nx * px + ny * py + nz * pz);
          // Speculative: a vertex is constrained before it arrives, so the
          // solver can stop it exactly at the surface instead of the clamp
          // having to dig it back out afterwards.
          const closing = vx * nx + vy * ny + vz * nz;
          if (gap >= CONTACT_MARGIN + Math.max(0, closing) * dt * 1.5) continue;
          body.touching = true;
          fillContact(take(), body, null, px, py, pz, -nx, -ny, -nz, gap, restitution);
        }
      }
    }

    for (let i = 0; i < bodies.length; i++) {
      for (let j = 0; j < bodies.length; j++) {
        if (i === j) continue;
        const a = bodies[i];
        const b = bodies[j];
        const dx = a.position[0] - b.position[0];
        const dy = a.position[1] - b.position[1];
        const dz = a.position[2] - b.position[2];
        // Broad phase: unit circumradii, so hulls cannot touch beyond 2 apart.
        if (dx * dx + dy * dy + dz * dz > 2.4 * 2.4) continue;
        maxDieOverlap = Math.max(
          maxDieOverlap,
          collectDieContacts(a, b, kernel, restitution, dt, take),
        );
      }
    }

    for (let iteration = 0; iteration < SOLVER_ITERATIONS; iteration++) {
      for (let c = 0; c < used; c++) solveContact(pool[c], friction, dt);
    }

    for (const body of bodies) {
      if (body.touching) {
        body.angular[0] *= contactDrag;
        body.angular[1] *= contactDrag;
        body.angular[2] *= contactDrag;
      }
      body.position[0] += body.velocity[0] * dt;
      body.position[1] += body.velocity[1] * dt;
      body.position[2] += body.velocity[2] * dt;
      integrateOrientation(body, dt);
      refresh(body, kernel);

      // Backstop. The contact constraint is linear in velocity while a vertex
      // on a spinning die travels an arc, so a corner can dip below a wall by
      // roughly `(w dt)^2 r / 2`. Push it back out; the walls are mutually
      // orthogonal, so one pass per plane is exact.
      let corrected = false;
      for (let p = 0; p < 6; p++) {
        const nx = wallNormals[p * 3];
        const ny = wallNormals[p * 3 + 1];
        const nz = wallNormals[p * 3 + 2];
        let deepest = 0;
        const w = body.worldVertices;
        for (let v = 0; v < kernel.vertexCount; v++) {
          const gap = wallOffsets[p] - (nx * w[v * 3] + ny * w[v * 3 + 1] + nz * w[v * 3 + 2]);
          if (gap < deepest) deepest = gap;
        }
        if (deepest < 0) {
          // `deepest` is negative and the normal faces outward, so this moves
          // the body back INTO the container by exactly the overshoot.
          body.position[0] += nx * deepest;
          body.position[1] += ny * deepest;
          body.position[2] += nz * deepest;
          maxWallCorrection = Math.max(maxWallCorrection, -deepest);
          corrected = true;
        }
      }
      if (corrected) refresh(body, kernel);
    }

    step++;
    energyPerStep.push(totalEnergy(bodies, kernel, gravity, floor));

    let resting = true;
    for (const body of bodies) {
      const v = body.velocity;
      const a = body.angular;
      if (
        !body.touching ||
        v[0] * v[0] + v[1] * v[1] + v[2] * v[2] > REST_LINEAR * REST_LINEAR ||
        a[0] * a[0] + a[1] * a[1] + a[2] * a[2] > REST_ANGULAR * REST_ANGULAR
      ) {
        resting = false;
        break;
      }
    }
    restSteps = resting ? restSteps + 1 : 0;

    if (step % STEPS_PER_SAMPLE === 0) {
      record(step / STEPS_PER_SAMPLE);
      if (restSteps >= REST_HOLD_STEPS) break;
    }
  }

  const durationMs = tracks[0][tracks[0].length - 1].t;
  const dice = tracks.map((samples) =>
    Object.freeze({
      samples: Object.freeze(samples),
      restingOrientation: samples[samples.length - 1].orientation,
    }),
  );

  let restAlignment = Infinity;
  for (const die of dice) {
    restAlignment = Math.min(
      restAlignment,
      readableRestAlignment(config.geometry, die.restingOrientation),
    );
  }
  return Object.freeze({
    trajectory: Object.freeze({ durationMs, dice: Object.freeze(dice) }),
    energyPerStep: Object.freeze(energyPerStep),
    maxWallCorrection,
    maxDieOverlap,
    stepCount: step,
    attempts: 1,
    restAlignment,
  });
}

/**
 * Vertices of `a` that are inside — or about to be inside — the convex hull of
 * `b`, tested against `b`'s own face planes and appended through `take`.
 *
 * Same routine for every shape, and the reason `capFaces` matters: `planes`
 * spans the whole closed surface, so a barrel is a closed convex solid here and
 * not an open tube. Returns the deepest overlap seen, for diagnostics.
 */
function collectDieContacts(
  a: Body,
  b: Body,
  kernel: Kernel,
  restitution: number,
  dt: number,
  take: () => Contact,
): number {
  const r = b.rotation;
  const w = a.worldVertices;
  let deepest = 0;
  for (let v = 0; v < kernel.vertexCount; v++) {
    const px = w[v * 3];
    const py = w[v * 3 + 1];
    const pz = w[v * 3 + 2];
    const dx = px - b.position[0];
    const dy = py - b.position[1];
    const dz = pz - b.position[2];
    // Into b's body frame: R^T d, i.e. dot with the COLUMNS of R.
    const lx = r[0] * dx + r[3] * dy + r[6] * dz;
    const ly = r[1] * dx + r[4] * dy + r[7] * dz;
    const lz = r[2] * dx + r[5] * dy + r[8] * dz;
    // The tightest supporting plane; negative means the vertex is inside.
    let best = -Infinity;
    let bestPlane = 0;
    for (let p = 0; p < kernel.planeCount; p++) {
      const distance =
        kernel.planeNormals[p * 3] * lx +
        kernel.planeNormals[p * 3 + 1] * ly +
        kernel.planeNormals[p * 3 + 2] * lz -
        kernel.planeOffsets[p];
      if (distance > best) {
        best = distance;
        bestPlane = p;
      }
    }
    if (best < 0 && -best > deepest) deepest = -best;
    // Back to the world frame: R n, pointing out of b and so pushing a away.
    const bx = kernel.planeNormals[bestPlane * 3];
    const by = kernel.planeNormals[bestPlane * 3 + 1];
    const bz = kernel.planeNormals[bestPlane * 3 + 2];
    const nx = r[0] * bx + r[1] * by + r[2] * bz;
    const ny = r[3] * bx + r[4] * by + r[5] * bz;
    const nz = r[6] * bx + r[7] * by + r[8] * bz;
    const wa = a.angular;
    const wb = b.angular;
    const rax = px - a.position[0];
    const ray = py - a.position[1];
    const raz = pz - a.position[2];
    const vx =
      a.velocity[0] + wa[1] * raz - wa[2] * ray - (b.velocity[0] + wb[1] * dz - wb[2] * dy);
    const vy =
      a.velocity[1] + wa[2] * rax - wa[0] * raz - (b.velocity[1] + wb[2] * dx - wb[0] * dz);
    const vz =
      a.velocity[2] + wa[0] * ray - wa[1] * rax - (b.velocity[2] + wb[0] * dy - wb[1] * dx);
    const closing = -(vx * nx + vy * ny + vz * nz);
    if (best >= CONTACT_MARGIN + Math.max(0, closing) * dt * 1.5) continue;
    a.touching = true;
    b.touching = true;
    fillContact(take(), a, b, px, py, pz, nx, ny, nz, best, restitution);
  }
  return deepest;
}

/** `simulateRoll`, plus the internals the suite asserts against. */
export function simulateRollWithDiagnostics(config: RollConfig): RollDiagnostics {
  const options = resolved(config);
  const solid = simulationSolid(config.geometry);
  const kernel = kernelOf(solid);
  // One stream seeds the throws, so the retry chain is itself deterministic.
  const seeds = createRandom(config.seed);
  let best: RollDiagnostics | null = null;
  let attempts = 0;
  while (attempts < MAX_ATTEMPTS) {
    attempts++;
    const result = throwOnce(config, options, kernel, createRandom(seeds() * 0x100000000));
    if (best === null || result.restAlignment > best.restAlignment) best = result;
    if (result.restAlignment >= SETTLE_ALIGNMENT) break;
  }
  const chosen = best as RollDiagnostics;
  return Object.freeze({ ...chosen, attempts });
}

/**
 * Tumble `config.dieCount` dice inside the container until they settle.
 *
 * Deterministic in `config`: the same configuration always yields the same
 * trajectory, which is what lets the renderer rebuild a roll from scratch on
 * every mount without the dice jumping.
 */
export function simulateRoll(config: RollConfig): RollTrajectory {
  return simulateRollWithDiagnostics(config).trajectory;
}
