/**
 * Deterministic rigid-body simulation of dice tumbling inside a box.
 *
 * Pure arithmetic over `die-geometry.ts`: no DOM, no packages, no wall clock and
 * — critically — no `Math.random`. The renderer re-derives a roll from
 * `(component id, RollCount)` every time it mounts, so the same seed has to
 * reproduce the same trajectory bit for bit, forever. Everything stochastic here
 * comes out of a mulberry32 stream seeded from `RollConfig.seed`, and the only
 * transcendental in the whole module is `Math.sqrt`, which IEEE-754 requires to
 * be correctly rounded. Nothing depends on the platform's `sin`/`cos`/`exp`.
 *
 * ## Units, and the size-normalisation trap
 *
 * `die-geometry.ts` does NOT normalise its solids: the bounding radius runs from
 * 1.000 (d8) to 1.902 (d20), and unit-mass inertia goes as R^2, so a d20's
 * inertia tensor is about 4x a d10's at the scale each is built at. Fed straight
 * into the angular dynamics that makes different face counts tumble at visibly
 * different rates for no reason a test would explain.
 *
 * This module normalises ON ENTRY, in `simulationSolid`: vertices and face
 * planes are divided by `boundingRadius` and the inertia tensor by
 * `boundingRadius^2`, so EVERY simulated die is a unit-radius, unit-mass solid.
 * One consequence for every consumer: all lengths in a `RollTrajectory` —
 * positions, and the `bounds` you hand in — are in units of the die's BOUNDING
 * radius, not pixels and not the geometry's native coordinates. A renderer
 * scales the whole trajectory by whatever it wants a die radius to be on screen.
 *
 * `boundingRadius` is deliberately not `circumradius`: the latter is the
 * RENDERER's normalisation and is a barrel's short axis. The two agree for every
 * closed-form solid, and a barrel is drawn 2.1-2.6x larger than the sphere it is
 * simulated in — see `die-geometry.ts` and `dice-roll.ts`'s `TRAY_BOUNDS`.
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
 * is why nearly all depenetration is done by impulse.
 *
 * NEARLY all. The linearised constraint still leaves the arc a vertex on a
 * spinning die travels within one step, so the step ends with a positional
 * clamp that TELEPORTS the body back inside the container (`maxWallCorrection`
 * reports how far). A teleport upward is free potential energy: measured across
 * 84 three-die rolls the worst single step gained 6.8e-5 of the launch energy
 * this way, which is above the suite's `RELATIVE_STEP_TOLERANCE` of 1e-5 on its
 * own — the energy assertion passes because the same step's contact and drag
 * dissipation is larger still, not because the clamp is free. Shrink the
 * dissipation or grow the clamp and that assertion is what fails first.
 *
 * ## Restitution
 *
 * Bounce is applied by Poisson's hypothesis, not by giving each contact its own
 * target separating velocity; see `applyRestitution` for why the obvious
 * formulation pumps energy into a die that lands flat.
 *
 * ## Why the inner loop looks the way it does
 *
 * `die-geometry.ts`'s vector helpers return FROZEN tuples, which is right for
 * geometry built once and wrong for a solver that runs ten impulse iterations
 * over a dozen contacts at 720 Hz. Written with them, a two-die roll cost 100 ms
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
  /**
   * Any finite number, and every bit of it counts: `2**32 + 1`, `1.5` and `1`
   * are three different rolls. See `createRandom` for the exact contract — the
   * caller that derives a seed from `(component id, RollCount)` needs it.
   */
  readonly seed: number;
  /**
   * The shape of every die, or one shape PER DIE.
   *
   * A single geometry is the common case and means exactly what it always did:
   * `dieCount` dice of that shape. A list is a throw of mixed shapes — 2d6 and
   * a d20 together, or dice alongside some other component this simulator is
   * reused for — and must have exactly `dieCount` entries, in the order the
   * dice appear in `RollTrajectory.dice`.
   *
   * Nothing below this line assumes the dice agree: contacts, inertia, the
   * broad phase and the settle test are all read off the individual body's own
   * solid. The one thing that IS shared is scale, because `simulationSolid`
   * normalises every shape to circumradius 1 — a mixed throw is a throw of
   * dice that are all the same SIZE and different shapes. A caller that wants
   * a d20 physically larger than its d6 has to say so somewhere this module
   * does not yet have a word for.
   */
  readonly geometry: DieGeometry | readonly DieGeometry[];
  readonly dieCount: number;
  /**
   * HALF-extents of the container box, in die circumradii. See the file docs.
   *
   * The floor of 1.5 is only what the SPAWN needs. A die also has to be able to
   * fall over once it lands, and a tray near that floor cannot let it: with 3
   * d20 over 25 seeds, `{x: 2, y: 5, z: 2}` came out cocked 17 times and
   * `{3, 3, 3}` once, against never at `{4, 4, 4}` and above. Four circumradii
   * of half-extent is the size to reach for; below it, read
   * `RollTrajectory.cocked`.
   */
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
  /**
   * Worst `cos(tilt)` over the dice at rest, where tilt is the angle between
   * straight down and the nearest READABLE face normal. 1 is dead flat.
   *
   * Public because `cocked` needs a magnitude behind it: a caller that wants to
   * treat 1 degree differently from 15 can, and a test can assert on it.
   */
  readonly restAlignment: number;
  /**
   * True when at least one die is not lying flat on a readable face, i.e. every
   * retry landed cocked and the least-cocked attempt was returned anyway.
   *
   * This is NOT hypothetical in a cramped container. `bounds` is only required
   * to be 1.5 circumradii of half-extent, and the retry cannot rescue a die
   * that has nowhere to fall over: measured with 3 d20 over 25 seeds, half-
   * extents of `{x: 2, y: 5, z: 2}` came out cocked 17 times, by up to 15.8
   * degrees, with all 8 attempts exhausted on 18 of them; `{3, 3, 3}` cocked
   * once at 3.3 degrees; `{4, 4, 4}` and up, never. The 6-cubed tray the suite
   * uses is comfortably in the "never" regime, which is why this went unseen.
   *
   * A cocked die matters because the renderer reads a value off the face it
   * decides is up, and on a cocked die that face is not actually up — so the
   * animation ends showing a number the die is not really displaying. React to
   * this by widening the tray, hiding the roll, or annotating the result;
   * ignoring it is a choice, but it should be a choice.
   */
  readonly cocked: boolean;
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
  /**
   * The farthest any vertex sits from the centre, MEASURED rather than assumed.
   *
   * `simulationSolid` divides by `geometry.circumradius`, so this is 1 to
   * within rounding for anything it returns. It is carried anyway because the
   * broad phase and `closingSpeedBound` need a bound on how far a surface point
   * can be from its own centre, and writing that as the literal `1` bakes this
   * function's normalisation into a routine two hundred lines away — where it
   * would have to be found and changed by whoever first hands the solver a
   * solid that is not a die.
   */
  readonly radius: number;
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
  /**
   * How far, in degrees, the pose a die is left on differs from the pose the
   * physics eventually reached — the price of cutting the dead tail. The worst
   * over the dice. See `liveFrameCount`.
   */
  readonly restingDrift: number;
  readonly stepCount: number;
  /**
   * Every physics step the call executed, across every attempt.
   *
   * `stepCount` is the length of the trajectory that came back; this is the
   * work that produced it, and the two differ by exactly what the retries
   * cost. A test that wants to assert what a roll COSTS has to assert on this
   * one — the other says nothing about the throws that were thrown away.
   */
  readonly simulatedSteps: number;
  /**
   * How many throws it took to land every die flat on a readable face. See
   * `MAX_ATTEMPTS`; 1 for the overwhelming majority of rolls.
   */
  readonly attempts: number;
  /**
   * Where each attempt's trace begins in `energyPerStep`, which spans EVERY
   * attempt and not just the one that was kept.
   *
   * The energy invariant holds within an attempt and says nothing across the
   * boundary, because the next attempt starts with a fresh launch. Reported so
   * the invariant can be checked on the throws that were discarded too: a
   * genuine energy pump was hiding in one of them (see `PENETRATION_SLOP`),
   * invisible for as long as a rejected throw took its history with it.
   */
  readonly attemptStarts: readonly number[];
  /**
   * `restAlignment` for each die on its own, in the order of `trajectory.dice`.
   *
   * This is what makes the retry per die: the roll re-throws the dice whose
   * entry is below `SETTLE_ALIGNMENT` and leaves every other die's launch
   * alone. Public because `restAlignment` is the `min` of these and a caller
   * that wants to know WHICH die is cocked cannot recover it from the minimum.
   */
  readonly dieAlignment: readonly number[];
  /**
   * The same number as `RollTrajectory.restAlignment`, kept here so a test that
   * already holds the diagnostics does not have to reach through `trajectory`.
   */
  readonly restAlignment: number;
}

// ---------------------------------------------------------------------------
// Tuning
// ---------------------------------------------------------------------------

/**
 * 1080 Hz. The step has to be short compared with `1 / angularSpeed`: the
 * contact constraint is linear in velocity while a vertex on a spinning die
 * travels an ARC, so a corner dips below the floor between solves by about
 * `(w dt)^2 r / 2`, and the positional clamp that cleans that up is the one
 * part of the step that is not physics.
 *
 * It was 720, and the roll's tuning is what moved it. Landing speeds go as
 * `sqrt(gravity)` and gravity is now 65 rather than 42, so the post-impact spin
 * peak went up with it; measured across 175 three-die rolls the worst clamp
 * correction was 1.02e-2 of a circumradius at 720 Hz, against a suite bound of
 * 1e-2, and 7.7e-3 at 1080. The same change carried the worst single-step
 * energy gain from 1.5e-5 (over the suite's 1e-5 budget) to 4.1e-9. The cost is
 * 1.5x the steps per second of simulated time, most of which the shorter rolls
 * hand straight back.
 */
const PHYSICS_HZ = 1080;
/**
 * 180 Hz: three samples per 60 Hz display frame.
 *
 * This is not a rendering rate, it is the RESOLUTION at which a trajectory is
 * handed over, and it has to resolve the motion the simulation produces. A
 * rotation between two samples is interpolated the short way round, which is
 * the right answer only while a sample-to-sample turn is well under half a
 * revolution — and `dice-bake.test.ts` requires eightfold margin on exactly
 * that, `angle < 2*pi/9`, i.e. 40 degrees per sample.
 *
 * At 60 Hz that ceiling was what capped the throw: the dice could not be spun
 * fast enough to make two full turns without putting 96 degrees into one
 * sample, so they turned about once and read as a topple rather than a tumble.
 * Sampling three times as often divides the per-sample angle by three — the
 * same motion now peaks at 32.3 degrees — and the launch spin was raised into
 * the room that opened up. See `LAUNCH_SPIN_MIN`.
 *
 * What it costs is the size of a `RollTrajectory`: a median roll went from 35
 * samples to 105. Measured over 210 rolls, `trajectoryCurve` takes 0.018 ms to
 * build against 0.007 ms before, and 61 evaluations of the returned curve —
 * one per frame of a second-long animation — cost 0.172 ms against 0.163 ms,
 * because evaluating is a binary search and not a scan. Both are noise beside
 * the ~8 ms the simulation itself takes. The bake compiler clamps to 256
 * keyframes, which a 180 Hz roll reaches at 1.4 seconds rather than 4.3; the
 * measured 99th percentile roll is under 2 seconds, so the longest rolls now
 * compile slightly coarser than their samples, which is a resolution the
 * compiler has always been free to choose.
 */
const SAMPLE_HZ = 180;
const STEPS_PER_SAMPLE = PHYSICS_HZ / SAMPLE_HZ;
const STEP_SECONDS = 1 / PHYSICS_HZ;

const MAX_SECONDS = 5;
/** Gauss-Seidel passes over the contact set. A resting cube needs about 4. */
const SOLVER_ITERATIONS = 10;

const DEFAULT_ENERGY = 1;
/**
 * Gravity, friction and the contact drag below are the roll's PACE, and they
 * were retuned together against measurement rather than one at a time.
 *
 * The complaint they answer: a roll blocks the game's animation cycle for its
 * whole length, and at the renderer's 1.6 x 2.0 x 1.6 tray the median d12 took
 * 1.48 seconds and the median d20 1.32 to say a number the player could read
 * after half of it. Harder gravity shortens the fall, more friction stops the
 * slide, and more rolling resistance stops the long slow topple at the end.
 * Median duration over 24 seeds per shape, before -> after:
 *
 *     d3   883 -> 461     d7  1417 ->  539     d12  1483 -> 606
 *     d4   800 -> 350     d10 1133 ->  622     d20  1317 -> 600
 *     d6   983 -> 428
 *
 * and the share of frames turning under a degree — dead frames, the last third
 * of a filmstrip in which nothing happens — went from a 33% worst median to
 * 0-19%. The tumble moved with them, but only once the sample grid could carry
 * it: see `SAMPLE_HZ` and `LAUNCH_SPIN_MIN`.
 */
const DEFAULT_GRAVITY = 85;
const DEFAULT_RESTITUTION = 0.32;
const DEFAULT_FRICTION = 0.6;

/** Air drag, per second. Small: friction does the real work. */
const LINEAR_DRAG = 0.3;
const ANGULAR_DRAG = 0.3;
/**
 * Rolling resistance, applied only while a die is touching something.
 *
 * This is the constant that ends the roll. Coulomb friction alone does settle a
 * die, but a rounder solid finishes with a slow topple from one face to the
 * next that costs a third of a second and shows the player nothing, and at 7 —
 * where this used to sit — a d12's 95th percentile was still 2.85 seconds.
 * Raising it trades tumble for brevity, so it is bounded on both sides: at 26
 * the die stops turning noticeably before it has finished settling, which shows
 * up as `restingDrift` climbing, and too low and the median d12 goes back over
 * a second. 34 is where the two meet at the current launch spin — it went up
 * with the spin, because a die thrown twice as hard has twice as much to shed
 * before it can settle.
 */
const CONTACT_ANGULAR_DRAG = 34;

/** A vertex this close to a surface is a contact candidate even at rest. */
const CONTACT_MARGIN = 0.02;
/**
 * How far ahead a speculative contact reaches: a vertex closing at `v` is made
 * a contact once its gap drops below `CONTACT_MARGIN + v * dt * LOOKAHEAD`.
 * Named because the broad phase has to reach at least as far, or it would skip
 * a pair the narrow phase would have found.
 */
const CONTACT_LOOKAHEAD = 1.5;
/**
 * Below this approach speed a contact is treated as resting and gets no
 * restitution. Bouncing a die that is settling is both unphysical (a real die
 * has no coefficient of restitution at 0.1 mm/s) and the classic way a
 * sequential-impulse solver jitters forever instead of coming to rest.
 */
const RESTITUTION_THRESHOLD = 1.2;
/**
 * Cap on how much penetration one step's impulse is allowed to undo.
 *
 * This is a SPEED limit in disguise — the bias target is `depth / dt`, so 0.02
 * at 720 Hz let a contact demand 14.4 circumradii per second of separation, and
 * whatever separating speed the bias buys is left in the bodies as kinetic
 * energy once the overlap is gone. That is energy from nowhere, and at
 * `restitution: 1`, where nothing else in the step dissipates, one d4 contact
 * spent it all at once: a single step gained 4.4e-4 of the launch energy
 * against the suite's 1e-5 budget.
 *
 * 0.005 is still more than twice the deepest overlap ever measured (2.2e-3
 * across the whole restitution sweep), so the cap only binds on a spike that
 * should not happen, which is exactly what a safety valve is for. Measured
 * worst single-step gain across that sweep: 4.4e-4 at 0.02, and 1.7e-8 at
 * anything from 0.005 down to zero, with the deepest overlap unchanged.
 */
const MAX_DEPENETRATION_SPEED = 3.6;
/**
 * Overlap a contact is allowed to keep: the bias only pushes on what is deeper
 * than this.
 *
 * The depenetration bias is the one part of the normal solve that is not a
 * collision response — it is a position correction driven through the velocity
 * channel, so whatever separating speed it buys is left in the bodies as
 * kinetic energy afterwards, and that is energy from nowhere. Contacts here are
 * speculative, so genuine overlap is only ever the residue of the arc a vertex
 * travels within one step, and correcting the last two thousandths of it buys
 * nothing visible while costing exactly that. Measured across the suite's
 * restitution sweep (two dice, d4/d6/d20, 7 restitutions x 4 seeds) the worst
 * single-step energy gain as a fraction of the launch energy:
 *
 *     no slop -> 2.9e-5      slop 0.002 -> 1.7e-7
 *
 * against a suite budget of 1e-5, and the deepest overlap anywhere in that
 * sweep went 2.49e-2 -> 2.33e-2, i.e. slightly BETTER rather than worse. The
 * 2.9e-5 was not new: the retry used to run whole separate throws and hand back
 * only the chosen one's energy trace, so a throw that was rejected for landing
 * cocked took its energy history with it.
 */
const PENETRATION_SLOP = 0.002;

/**
 * What one sample has to do to count as motion rather than as tail: degrees of
 * rotation and circumradii of travel, per `SAMPLE_HZ` interval. See
 * `liveFrameCount`.
 *
 * These are RATES wearing per-sample clothing — 120 degrees a second and 0.72
 * circumradii a second — so they moved with the sample grid when it went from
 * 60 Hz to 180 and mean exactly what they meant before: at a 100px die, 1.7 px
 * of surface swing and 0.2 px of drift per 60 Hz display frame. Anything
 * slower than that is a die the player has already stopped watching.
 */
const TAIL_ROTATION = 0.667;
const TAIL_TRAVEL = 0.004;

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
 * for a cube) or leaning on a wall is nowhere near either number.
 *
 * The limit of a fixed angle: a barrel resting on a SIDE EDGE is only pi/N from
 * a side-face normal, which is 25.7 degrees for a d7 but 5 degrees at N = 36 and
 * 1.8 at N = 100. This constant therefore stops distinguishing a flat rest from
 * an edge rest somewhere around a d36. Nothing ships anywhere near that — the
 * suite tops out at a d24, where pi/N is 7.5 degrees — and a barrel that many
 * sides cannot be read at a glance anyway; if one is ever wanted, this wants to
 * become `min(3 degrees, pi / (2N))` and the suite's limits with it.
 * (Cap facets are a different question and no longer a live one: they used to
 * be stable rests 61 degrees off readable, and `die-geometry.ts` now
 * proportions every barrel so no cap facet is a stable rest at all.)
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
 * How many times the throw may be repeated before the roll gives up and hands
 * back the best it saw.
 *
 * A cocked die is re-thrown, exactly as at a real table — and only the cocked
 * die is: each die's launch comes from its own seed, and an attempt resamples
 * the seeds of the dice that landed badly and no others (see
 * `simulateRollWithDiagnostics`). What that removes is the compounding. The
 * retry used to redraw the whole tray, so settling was a coin that had to come
 * up heads for every die at once and the success rate fell as `p^dieCount`;
 * it never failed at one die and failed every single time at twenty-five,
 * having burned all eight throws to get there.
 *
 * Deterministic: every seed is drawn from one stream rooted at `config.seed`,
 * so the whole retried roll still reproduces bit for bit. If every attempt
 * lands cocked the least-cocked one is returned rather than a throw — a
 * slightly ugly die beats no animation — and `RollTrajectory.cocked` says so
 * out loud, because a caller that is told nothing will render a face that is
 * not really up.
 *
 * The retry is invisible on purpose: a rejected throw is not part of the
 * trajectory. Splicing it in instead was measured, and it is much worse where
 * it matters — in the renderer's own 1.6 x 2.0 x 1.6 tray a d3 cocks half the
 * time, so half of all d3 rolls would visibly re-throw and the 90th-percentile
 * roll went from 1.2 to 4.0 seconds.
 */
const MAX_ATTEMPTS = 8;
/**
 * How many attempts in a row may fail to reduce the number of cocked dice
 * before the roll accepts that they are not going to.
 *
 * Some trays cannot be settled at all: twenty-five d20 in a 6-cubed tray land
 * in a heap, and a die resting on another die is cocked no matter what launch
 * it was given. The retry cannot see that in advance, but it can see that it is
 * not making progress, and stopping then is the difference between three
 * throws and eight.
 */
const RETRY_PATIENCE = 3;

/** Spawn grid pitch, in circumradii: two unit dice need more than 2 apart. */
const SPAWN_PITCH = 2.4;
/** Clearance kept between a spawned die and the container walls. */
const SPAWN_CLEARANCE = 1.2;
/**
 * Launch speeds at `energy: 1`. Scaled by `sqrt(energy)`, so KE is linear.
 *
 * The spin range is 60 to 102 rad/s, i.e. 9.5 to 16 turns per second, which
 * over the quarter-second a die spends in the air is the two to three visible
 * tumbles a thrown die makes. It was 20 to 44 and the dice turned about once in
 * total — 387 to 431 degrees of median rotation across the shapes, a slow
 * topple rather than a throw. They now turn 806 to 953 degrees, 2.2 to 2.6 full
 * revolutions.
 *
 * ## The ceiling this sits under, which is not physical
 *
 * A roll is handed over as samples, and a rotation between two of them is
 * interpolated the short way round; `dice-bake.test.ts` requires eightfold
 * margin on that, `angle < 2*pi/9`, i.e. 40 degrees per SAMPLE. An impact can
 * spin a die up well past its launch, so this range peaks at 32.3 degrees a
 * sample and the headroom is real but not large.
 *
 * That ceiling is per sample and not per second, which is the whole reason this
 * number could be raised: at 60 Hz the same 60-102 rad/s put 96 degrees into
 * one sample and the bake refused it, correctly. `SAMPLE_HZ` went to 180 first;
 * this followed. Anyone reaching for more spin has to move them together, and
 * `SAMPLE_HZ` is the one with the consequences — see its own note.
 */
const LAUNCH_HORIZONTAL = 5.5;
const LAUNCH_SPIN_MIN = 60;
const LAUNCH_SPIN_RANGE = 42;


// ---------------------------------------------------------------------------
// Seeded randomness
// ---------------------------------------------------------------------------

/** splitmix32's finaliser: the avalanche step, on its own. */
function mix32(value: number): number {
  let x = value >>> 0;
  x = Math.imul(x ^ (x >>> 16), 0x21f0aaad) >>> 0;
  x = Math.imul(x ^ (x >>> 15), 0x735a2d97) >>> 0;
  return (x ^ (x >>> 15)) >>> 0;
}

/**
 * Scratch for reading a double's bits. A `DataView` rather than a
 * `Uint32Array` over a `Float64Array` because the latter is host-endian and
 * this module's whole contract is that a seed reproduces on any engine;
 * `getUint32(_, true)` pins little-endian regardless of the machine.
 */
const SEED_BITS = new DataView(new ArrayBuffer(8));

/**
 * mulberry32. Chosen because its whole state is one uint32 and every operation
 * is integer, so it reproduces exactly on any conforming engine — which is the
 * entire point of having a PRNG here instead of `Math.random`.
 *
 * ## The seed contract
 *
 * ALL 64 bits of the seed are hashed, not just an int32 of it. This used to be
 * `Math.trunc(seed) ^ ...`, an implicit ToInt32, and the aliasing that produced
 * was structural rather than rare: `3` and `3.7` rolled identically, so did `1`
 * and `2**32 + 1`, and so did `-1` and `2**32 - 1`. The renderer derives its
 * seed from `(component id, RollCount)`, and a hash that overflows 32 bits or
 * lands on a fraction would have replayed one roll's animation for a
 * different roll with no signal anywhere. That renderer's own hash is FNV-1a
 * and so hands over a uint32 today, which none of those families would trip;
 * the guarantee below is about the parameter's type rather than about one
 * caller's arithmetic, because the hash on the far side of a `number` is the
 * caller's business and is free to change. Distinct finite doubles now give
 * distinct hash inputs; `-0` is folded onto `0` (`seed + 0`) because the two
 * are `===` and callers have no way to tell which one they produced.
 *
 * What is NOT promised: mulberry32's state is 32 bits, so there are only 2^32
 * possible streams and distinct seeds must sometimes collide. The guarantee is
 * that collisions are unstructured — a pseudorandom ~2^-32 per pair — rather
 * than the systematic families ToInt32 created.
 *
 * The hash is a splitmix32 avalanche of each half in turn. Without avalanching,
 * consecutive seeds differ by one bit of state and mulberry32's FIRST few
 * outputs stay correlated — which here means seed 1 and seed 2 throw their dice
 * in nearly the same direction. Every seed-dependent quantity in this module is
 * drawn in the first dozen outputs, so that is exactly the regime that matters.
 */
function createRandom(seed: number): () => number {
  SEED_BITS.setFloat64(0, seed + 0, true);
  let state = mix32(SEED_BITS.getUint32(0, true) ^ 0x9e3779b9);
  state = mix32(state ^ SEED_BITS.getUint32(4, true));
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
 * `geometry` rescaled so its bounding sphere has radius exactly 1.
 *
 * See the file docs: this is where the size trap is defused. Lengths divide by
 * `boundingRadius`; the unit-mass inertia tensor, whose entries are second
 * moments of length, divides by `boundingRadius^2`. Face NORMALS are invariant
 * under a uniform scale, so they carry through untouched — and so, usefully,
 * does `presentedFaceIndex`, which reads only normals and argmax/argmin over
 * uniformly scaled distances. Reading a resting orientation therefore gives the
 * same answer against the original geometry as against this one.
 *
 * `boundingRadius` and NOT `circumradius`, which the renderer normalizes by and
 * which is a barrel's short axis rather than its circumsphere. The two are the
 * same number for every closed-form solid and differ by up to 2.63x on a
 * barrel, and it is the bounding sphere that this module needs: `bounds` is a
 * tray measured in die radii, and a die whose sphere overflowed the tray would
 * spend the whole throw wedged in a wall. See `die-geometry.ts`.
 */
export function simulationSolid(geometry: DieGeometry): SimulationSolid {
  const radius = geometry.boundingRadius;
  if (!(radius > 0)) throw new Error('die geometry has a non-positive bounding radius');
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
    radius: Math.max(...vertices.map((vertex) => Math.sqrt(dot(vertex, vertex)))),
  });
}

/**
 * The same solid flattened into typed arrays, for the allocation-free loop.
 *
 * One per SHAPE, not one per die: a body holds a reference to its own kernel
 * (see `Body.kernel`), and a throw of five d6 and three d20 builds two.
 */
interface Kernel {
  readonly geometry: DieGeometry;
  readonly vertexCount: number;
  /** 3 per vertex. */
  readonly vertices: Float64Array;
  readonly planeCount: number;
  /** 3 per plane. */
  readonly planeNormals: Float64Array;
  readonly planeOffsets: Float64Array;
  readonly inertia: Float64Array;
  readonly inverseInertia: Float64Array;
  /** See `SimulationSolid.radius`: the bound the broad phase reaches for. */
  readonly radius: number;
}

function kernelOf(geometry: DieGeometry, solid: SimulationSolid): Kernel {
  const vertices = new Float64Array(solid.vertices.length * 3);
  solid.vertices.forEach((vertex, index) => vertices.set(vertex, index * 3));
  const planeNormals = new Float64Array(solid.planes.length * 3);
  const planeOffsets = new Float64Array(solid.planes.length);
  solid.planes.forEach((plane, index) => {
    planeNormals.set(plane.normal, index * 3);
    planeOffsets[index] = plane.offset;
  });
  return {
    geometry,
    vertexCount: solid.vertices.length,
    vertices,
    planeCount: solid.planes.length,
    planeNormals,
    planeOffsets,
    inertia: Float64Array.from(solid.inertia),
    inverseInertia: Float64Array.from(solid.inverseInertia),
    radius: solid.radius,
  };
}

/**
 * One kernel per DISTINCT geometry in the throw, in die order.
 *
 * Deduplicated by object identity, which is the cheap half of the job: the
 * usual `[g, g, g]` and the usual single geometry both build exactly one
 * kernel, so a throw of twenty-five identical dice does not flatten the same
 * solid twenty-five times. Two structurally equal but distinct geometry objects
 * build two kernels and simulate identically; that is a wasted flatten and
 * never a wrong answer.
 */
function kernelsFor(geometries: readonly DieGeometry[]): Kernel[] {
  const built = new Map<DieGeometry, Kernel>();
  return geometries.map((geometry) => {
    let kernel = built.get(geometry);
    if (kernel === undefined) {
      kernel = kernelOf(geometry, simulationSolid(geometry));
      built.set(geometry, kernel);
    }
    return kernel;
  });
}

// ---------------------------------------------------------------------------
// Bodies
// ---------------------------------------------------------------------------

interface Body {
  /**
   * This die's own shape. Every routine below reads geometry through here and
   * never through the config, which is what lets one throw mix shapes.
   */
  readonly kernel: Kernel;
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
    kernel,
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
function refresh(body: Body): void {
  const kernel = body.kernel;
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
   * The restitution coefficient this contact bounces with, or 0 when it is not
   * an impact — a resting contact, or one whose approach speed is below
   * `RESTITUTION_THRESHOLD`. Decided before any impulse is applied, from the
   * pre-solve approach speed. See `applyRestitution`.
   */
  restitution = 0;
  /** Normal-relative speed before any impulse; negative means approaching. */
  approachSpeed = 0;
  effectiveMass = 0;
  normalImpulse = 0;
}

/**
 * An upper bound on how fast any point of `a` can be approaching any point of
 * `b`. A surface point is at most `kernel.radius` from its own centre, so it
 * moves at most `|v| + radius * |omega|`. Used only by the broad phase.
 */
function closingSpeedBound(a: Body, b: Body): number {
  const speed = (body: Body): number => {
    const v = body.velocity;
    const w = body.angular;
    return (
      Math.sqrt(v[0] * v[0] + v[1] * v[1] + v[2] * v[2]) +
      body.kernel.radius * Math.sqrt(w[0] * w[0] + w[1] * w[1] + w[2] * w[2])
    );
  };
  return speed(a) + speed(b);
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
  gravityStep: number,
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
  // The impact speed restitution is entitled to is the one the bodies had when
  // the step began, NOT the one they have after this step's gravity increment
  // has been added to them. Bouncing off the increment as well hands a die
  // `gravity * dt` of free speed at every contact — invisible at the default
  // 0.32, where the same step throws most of the impact away, and a per-step
  // energy gain of 2.4e-3 of the launch energy at restitution 1, where nothing
  // else is throwing anything away. A wall is immovable so the increment shows
  // up in full; two dice both received it, so it cancels out of their relative
  // velocity and only the wall case corrects.
  const approach =
    RELATIVE[0] * nx + RELATIVE[1] * ny + RELATIVE[2] * nz + (b === null ? gravityStep * ny : 0);
  contact.approachSpeed = approach;
  contact.restitution =
    separation < CONTACT_MARGIN && approach < -RESTITUTION_THRESHOLD ? restitution : 0;
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
  // already has, undo whatever is deeper than `PENETRATION_SLOP` and at most
  // `MAX_DEPENETRATION` this step, so a deep overlap resolves smoothly instead
  // of firing the body out and a shallow one is simply left alone.
  //
  // The target is NOT zero. A speculative contact is created while the vertex
  // still has a gap, and its non-penetration target is the NEGATIVE speed
  // `-gap/dt`; a floor of zero would forbid closing the gap at all and every
  // die would come to rest hovering exactly one `CONTACT_MARGIN` above the
  // floor, wedgeable into a corner at any angle. That bug is invisible in a
  // containment test — a hovering die is very definitely contained.
  //
  // Nor does the target carry any bounce: restitution is a separate pass, and
  // this loop solves the perfectly inelastic problem. See `applyRestitution`.
  const target =
    contact.separation >= 0
      ? -contact.separation / dt
      : Math.min(Math.max(0, -contact.separation - PENETRATION_SLOP) / dt, MAX_DEPENETRATION_SPEED);

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

/**
 * Bounce, by Poisson's hypothesis: a second impulse of `restitution` times the
 * impulse the inelastic solve already used, applied once, after that solve has
 * converged.
 *
 * ## Why not the obvious thing
 *
 * The obvious formulation gives every contact its own target separating speed,
 * `restitution * (its own pre-solve approach speed)`, and maxes that into the
 * non-penetration target inside the iteration loop. For ONE contact that is
 * exactly Newton's rule and it is what this module used to do. For a die that
 * lands flat it is a pump: the several vertex contacts of a single face each
 * independently demand a bounce computed from a velocity that the other
 * contacts' impulses have already cancelled, and the solver obliges. Measured
 * worst single-step gain as a fraction of the launch energy, ONE die, 40 seeds
 * x 7 shapes:
 *
 *     r=0.32 -> 1.6e-7    r=0.7 -> 1.2e-2    r=0.85 -> 8.3e-2
 *     r=0.95 -> 2.1e-1    r=1.0 -> 2.8e-1
 *
 * `resolved()` accepts every one of those restitutions. More solver iterations
 * do not help — raising `SOLVER_ITERATIONS` to 60 improved r=0.32 to 5.7e-10
 * and left r=0.95 at 2.3e-1 — because the fixed point the iteration converges
 * to is itself the energy-gaining one. It is the formulation, not the solve.
 *
 * ## Why this one cannot pump
 *
 * Write the inelastic solve's accumulated normal impulses as `L >= 0`, the
 * pre-solve normal velocities as `u0` and the post-solve ones as `u1`. Applying
 * `E L`, with `E` diagonal in [0, 1], leaves the energy change
 *
 *     dKE = L^T (I + E)(I - E) u0 / 2 + L^T (I + E)^2 u1 / 2,
 *
 * and `u0 <= 0` wherever `E` is nonzero (a contact only earns a restitution
 * coefficient while it is approaching), so the first term can only take energy
 * out however many contacts fire at once. That is the sense in which
 * restitution is applied once per MANIFOLD: the bounce a face gets is
 * `restitution` times the impulse that whole face was JOINTLY solved to need,
 * not the sum of what each of its vertices would have demanded alone.
 *
 * ## The second cap, and why it is not optional
 *
 * The `u1` term needs `u1 <= 0` and that is only true where the inelastic
 * target was `-gap/dt`. For a contact that is already PENETRATING, the target
 * is the depenetration bias — up to `MAX_DEPENETRATION / dt`, which is 3.6
 * circumradii per second — so `L` there is not a compression impulse at all,
 * and scaling it by `restitution` is scaling a position correction. Poisson
 * alone measured WORSE than what it replaced (r=0.95 gained 5.6e-1 against
 * 2.1e-1, one die, 40 seeds x 7 shapes) entirely on that term.
 *
 * So the impulse is also capped by Newton's rule at this contact: never push
 * harder than would leave the contact separating at `restitution` times its own
 * approach speed. A bias-driven contact is already separating faster than that,
 * so its cap is zero and it gets no bounce; a clean impact has `u1` near zero,
 * so its cap is `restitution * -u0 / m` and does not bind. Taking the smaller
 * of the two leaves each cap covering the case the other misses.
 *
 * For a single isolated impact `L = -u0 / m` and both caps agree on exactly the
 * Newton result, so the default 0.32 bounce is unchanged in feel.
 */
function applyRestitution(contact: Contact): void {
  if (contact.restitution <= 0 || contact.normalImpulse <= 0) return;
  const { a, b, nx, ny, nz } = contact;
  relativeVelocity(contact);
  const separating = RELATIVE[0] * nx + RELATIVE[1] * ny + RELATIVE[2] * nz;
  // `effectiveMass` is the constraint's INVERSE mass, as in `solveContact`:
  // an impulse of `dv / effectiveMass` changes the separating speed by `dv`.
  const newton =
    (-contact.restitution * contact.approachSpeed - separating) / contact.effectiveMass;
  const magnitude = Math.min(contact.restitution * contact.normalImpulse, newton);
  if (magnitude <= 0) return;
  applyImpulse(
    a,
    contact.rax,
    contact.ray,
    contact.raz,
    nx * magnitude,
    ny * magnitude,
    nz * magnitude,
  );
  if (b) {
    applyImpulse(
      b,
      contact.rbx,
      contact.rby,
      contact.rbz,
      -nx * magnitude,
      -ny * magnitude,
      -nz * magnitude,
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
 * `config.geometry` as one entry per die, whichever way it was written.
 *
 * The single-geometry form repeats the SAME OBJECT, so `kernelsFor` dedupes it
 * back to one kernel and the throw is byte-identical to what it was before this
 * module could mix shapes.
 */
function geometriesOf(config: RollConfig): readonly DieGeometry[] {
  if (!Array.isArray(config.geometry)) {
    return new Array<DieGeometry>(config.dieCount).fill(config.geometry as DieGeometry);
  }
  const listed = config.geometry as readonly DieGeometry[];
  if (listed.length !== config.dieCount) {
    throw new Error(
      `geometry lists ${listed.length} shapes but dieCount is ${config.dieCount}; a per-die list must name one shape per die`,
    );
  }
  return listed;
}

/**
 * Dice start on a grid near the ceiling, one pitch apart so they never begin
 * interpenetrating, with a randomised orientation, spin and throw direction.
 * The grid wraps into a second row when the container is too narrow.
 *
 * Each die's launch is drawn from ITS OWN stream, seeded from
 * `launchSeeds[i]`. That is what lets the settle retry be per die: re-throwing
 * die 3 replaces `launchSeeds[3]` and nothing else, so every other die is
 * thrown exactly as it was thrown before. Sharing one stream would have
 * shifted every later die's draws and made "re-throw one die" a re-throw of
 * the whole tray in disguise. See `simulateRollWithDiagnostics`.
 */
function spawnBodies(
  config: RollConfig,
  kernels: readonly Kernel[],
  energy: number,
  launchSeeds: Float64Array,
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
    const random = createRandom(launchSeeds[i]);
    const body = createBody(kernels[i]);
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
    refresh(body);
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
function readableRestAlignment(geometry: DieGeometry, orientation: ArrayLike<number>): number {
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

/** The angle between two unit quaternions' orientations, in degrees. */
function quatDegrees(a: Quat, b: Quat): number {
  const cosine = Math.min(1, Math.abs(a[0] * b[0] + a[1] * b[1] + a[2] * b[2] + a[3] * b[3]));
  return (Math.acos(cosine) * 360) / Math.PI;
}

/**
 * The index of the last frame in which ANY die still visibly moved, which is
 * where the roll's animation should end.
 *
 * A roll does not stop, it decays: the dice trade their last energy for a
 * micro-creep that goes on for a third of a second and that no player can see.
 * Measured at a 100px die: a d12's median roll spent 37% of its frames turning
 * under a degree, and the worst d20 spent 57% — the last third of a d20's
 * filmstrip was frame after identical frame. That is not free. The roll gates
 * the whole animation cycle, so every dead frame is a frame the game waits on a
 * die that has already told the player its answer. It is now 0% to 19%.
 *
 * So the tail is cut at MOTION rather than at stillness. `TAIL_ROTATION` is a
 * fifth of the ~10 degrees per frame that reads as a tumble, and `TAIL_TRAVEL`
 * at a 100px die is half a pixel of drift per frame; a die still doing either
 * is a die still visibly settling, and a die doing neither has stopped as far
 * as the screen is concerned. What it costs is that the pose the roll ends on
 * is not quite the pose the physics would have reached — `restingDrift`
 * measures exactly that, and the suite holds it to a bound far inside the angle
 * at which the wrong face could be read.
 *
 * The cut is taken over ALL dice at once, because a `RollTrajectory` is one
 * timeline: the roll ends when the last die stops, not when the first does.
 */
function liveFrameCount(tracks: readonly (readonly DieSample[])[]): number {
  const frames = tracks[0].length;
  let live = 1;
  for (let i = 1; i < frames; i++) {
    for (const track of tracks) {
      const before = track[i - 1];
      const after = track[i];
      const dx = after.position[0] - before.position[0];
      const dy = after.position[1] - before.position[1];
      const dz = after.position[2] - before.position[2];
      if (
        dx * dx + dy * dy + dz * dz > TAIL_TRAVEL * TAIL_TRAVEL ||
        quatDegrees(before.orientation, after.orientation) > TAIL_ROTATION
      ) {
        live = i;
        break;
      }
    }
  }
  return Math.min(live, frames - 1);
}

/**
 * Total mechanical energy, unit mass per die, with the container floor as the
 * potential-energy datum. Never used by the simulation itself — it exists so the
 * suite can assert that no step ever gains energy, which is how an unstable
 * impulse solver announces itself long before anything visibly explodes.
 */
const INERTIA_WORLD = new Float64Array(9);
function totalEnergy(bodies: readonly Body[], gravity: number, floor: number): number {
  let total = 0;
  for (const body of bodies) {
    const v = body.velocity;
    const w = body.angular;
    conjugateInto(body.rotation, body.kernel.inertia, INERTIA_WORLD);
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
  kernels: readonly Kernel[],
  launchSeeds: Float64Array,
): RollDiagnostics {
  const { energy, gravity, restitution, friction } = options;
  const bodies = spawnBodies(config, kernels, energy, launchSeeds);
  const bounds = config.bounds;
  const floor = -bounds.y;
  const dt = STEP_SECONDS;

  // The six container half-spaces, `n . p <= offset`, so `normal` points OUT of
  // the box and a contact impulse is applied along its negation.
  const wallNormals = new Float64Array([1, 0, 0, -1, 0, 0, 0, 1, 0, 0, -1, 0, 0, 0, 1, 0, 0, -1]);
  const wallOffsets = new Float64Array([
    bounds.x, bounds.x, bounds.y, bounds.y, bounds.z, bounds.z,
  ]);

  // How much downward speed one step of gravity adds; `fillContact` takes it
  // back out before deciding what a contact is entitled to bounce with.
  const gravityStep = gravity * dt;
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
  energyPerStep.push(totalEnergy(bodies, gravity, floor));

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
      for (let v = 0; v < body.kernel.vertexCount; v++) {
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
          if (gap >= CONTACT_MARGIN + Math.max(0, closing) * dt * CONTACT_LOOKAHEAD) continue;
          body.touching = true;
          fillContact(take(), body, null, px, py, pz, -nx, -ny, -nz, gap, restitution, gravityStep);
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
        // Broad phase, stated as the bound it is rather than as a literal.
        // Each hull is bounded by its OWN `kernel.radius`, so the surfaces
        // cannot be closer than `distance - (ra + rb)`; the narrow phase makes
        // a contact at `CONTACT_MARGIN + closing * dt * CONTACT_LOOKAHEAD`, and
        // `closingSpeedBound` bounds how fast any point of one can approach any
        // point of the other using those same radii. Skipping a pair beyond
        // that provably drops no contact — for two shapes as much as for one,
        // which is why neither radius is written here as the literal 1 that
        // `simulationSolid`'s normalisation currently makes it.
        const reach =
          a.kernel.radius +
          b.kernel.radius +
          CONTACT_MARGIN +
          closingSpeedBound(a, b) * dt * CONTACT_LOOKAHEAD;
        if (dx * dx + dy * dy + dz * dz > reach * reach) continue;
        maxDieOverlap = Math.max(maxDieOverlap, collectDieContacts(a, b, restitution, dt, take));
      }
    }

    for (let iteration = 0; iteration < SOLVER_ITERATIONS; iteration++) {
      for (let c = 0; c < used; c++) solveContact(pool[c], friction, dt);
    }
    // One pass, after convergence, and never inside the loop above: that is the
    // whole of why this cannot pump energy. See `applyRestitution`.
    for (let c = 0; c < used; c++) applyRestitution(pool[c]);

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
      refresh(body);

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
        for (let v = 0; v < body.kernel.vertexCount; v++) {
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
      if (corrected) refresh(body);
    }

    step++;
    energyPerStep.push(totalEnergy(bodies, gravity, floor));

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

  const last = tracks[0].length - 1;
  const keep = liveFrameCount(tracks);
  const durationMs = tracks[0][keep].t;
  const dice = tracks.map((samples) => {
    const trimmed = samples.slice(0, keep + 1);
    return Object.freeze({
      samples: Object.freeze(trimmed),
      restingOrientation: trimmed[keep].orientation,
    });
  });

  // Read off the pose that is actually PRESENTED — the trimmed one — so the
  // retry is deciding about the die the player will see and not about a frame
  // that was cut.
  const dieAlignment = dice.map((die, i) =>
    readableRestAlignment(bodies[i].kernel.geometry, die.restingOrientation),
  );
  const restAlignment = Math.min(...dieAlignment);
  let restingDrift = 0;
  for (let i = 0; i < dice.length; i++) {
    restingDrift = Math.max(
      restingDrift,
      quatDegrees(dice[i].restingOrientation, tracks[i][last].orientation),
    );
  }
  return Object.freeze({
    trajectory: Object.freeze({
      durationMs,
      dice: Object.freeze(dice),
      restAlignment,
      cocked: restAlignment < SETTLE_ALIGNMENT,
    }),
    energyPerStep: Object.freeze(energyPerStep),
    maxWallCorrection,
    maxDieOverlap,
    restingDrift,
    stepCount: step,
    simulatedSteps: step,
    attempts: 1,
    attemptStarts: Object.freeze([0]),
    dieAlignment: Object.freeze(dieAlignment),
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
 *
 * The two sides come from DIFFERENT kernels and the asymmetry is the point:
 * the vertices are `a`'s, the planes are `b`'s. The caller runs the pair both
 * ways round, so a d4 against a d20 is tested as four points against twenty
 * planes and then as twelve points against four.
 */
function collectDieContacts(
  a: Body,
  b: Body,
  restitution: number,
  dt: number,
  take: () => Contact,
): number {
  const r = b.rotation;
  const w = a.worldVertices;
  const hull = b.kernel;
  let deepest = 0;
  for (let v = 0; v < a.kernel.vertexCount; v++) {
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
    for (let p = 0; p < hull.planeCount; p++) {
      const distance =
        hull.planeNormals[p * 3] * lx +
        hull.planeNormals[p * 3 + 1] * ly +
        hull.planeNormals[p * 3 + 2] * lz -
        hull.planeOffsets[p];
      if (distance > best) {
        best = distance;
        bestPlane = p;
      }
    }
    if (best < 0 && -best > deepest) deepest = -best;
    // Back to the world frame: R n, pointing out of b and so pushing a away.
    const bx = hull.planeNormals[bestPlane * 3];
    const by = hull.planeNormals[bestPlane * 3 + 1];
    const bz = hull.planeNormals[bestPlane * 3 + 2];
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
    if (best >= CONTACT_MARGIN + Math.max(0, closing) * dt * CONTACT_LOOKAHEAD) continue;
    a.touching = true;
    b.touching = true;
    // Two dice both got this step's gravity increment, so it cancels out of
    // their relative velocity and the correction is zero for a die-die contact.
    fillContact(take(), a, b, px, py, pz, nx, ny, nz, best, restitution, 0);
  }
  return deepest;
}

/** `simulateRoll`, plus the internals the suite asserts against. */
export function simulateRollWithDiagnostics(config: RollConfig): RollDiagnostics {
  const options = resolved(config);
  const kernels = kernelsFor(geometriesOf(config));
  // One stream hands out LAUNCH SEEDS, one per die, so the whole retried roll
  // still reproduces bit for bit.
  const seeds = createRandom(config.seed);
  const launchSeeds = new Float64Array(config.dieCount);
  for (let i = 0; i < launchSeeds.length; i++) launchSeeds[i] = seeds() * 0x100000000;

  const cockedCount = (result: RollDiagnostics): number =>
    result.dieAlignment.reduce((total, a) => total + (a < SETTLE_ALIGNMENT ? 1 : 0), 0);

  const energyPerStep: number[] = [];
  const attemptStarts: number[] = [];
  let simulatedSteps = 0;
  let best: RollDiagnostics | null = null;
  let bestCocked = Infinity;
  let stale = 0;
  let attempts = 0;
  while (attempts < MAX_ATTEMPTS) {
    attempts++;
    const result = throwOnce(config, options, kernels, launchSeeds);
    attemptStarts.push(energyPerStep.length);
    energyPerStep.push(...result.energyPerStep);
    simulatedSteps += result.simulatedSteps;
    // Keep the throw that left the FEWEST dice cocked, and only then the one
    // that left them least cocked. The old criterion was the min alignment
    // alone, which cannot tell one badly cocked die from twenty.
    const cocked = cockedCount(result);
    if (best === null || cocked < bestCocked || (cocked === bestCocked && result.restAlignment > best.restAlignment)) {
      if (cocked < bestCocked) stale = 0;
      bestCocked = Math.min(bestCocked, cocked);
      best = result;
    } else {
      stale++;
    }
    if (cocked === 0) break;
    // Stop when re-throwing has stopped helping. Twenty-five d20 in a 6-cubed
    // tray simply pile up — measured, every one of 12 seeds ended with a die
    // resting on another die, at every attempt — and there is no launch seed
    // that fixes a die with nowhere flat to land. Burning the remaining throws
    // to prove it again cost 3.4 seconds of synchronous main-thread work for a
    // result that was decided by attempt three.
    //
    // Only a tray that is failing WHOLESALE stops early: with a single cocked
    // die an attempt that does not fix it is ordinary bad luck, and cutting
    // that roll's throws short costs exactly the settling this retry exists
    // for — measured, it put a d3 back to cocking 2 rolls in 60 in the
    // renderer's own tray.
    if (bestCocked >= 2 && stale >= RETRY_PATIENCE) break;
    // The retry, and the whole of what makes it per die: only the dice that
    // landed badly get a new launch seed. Everything else is thrown again
    // EXACTLY as it was, which is why the next attempt is not another roll of
    // the same compound die.
    //
    // "Exactly" is about the launch and not about the outcome: dice collide, so
    // a die whose neighbour was re-thrown can still land somewhere new. What
    // goes away is the compounding. A whole-tray retry had to come up heads for
    // every die at once — `p^dieCount` — so it degraded from never failing at
    // one die to failing every time at twenty-five, and burned all eight throws
    // doing it. Measured on d20 in a 6-cubed tray, 8 seeds, before -> after:
    //
    //     dice    attempts        cocked        ms
    //        1    1.13 -> 1.13    0/8 -> 0/8      4 ->    4
    //        5    1.25 -> 1.25    0/8 -> 0/8     44 ->   38
    //       10    5.25 -> 1.75    2/8 -> 0/8    484 ->  169
    //       25    8.00 -> 2.38    8/8 -> 1/8   3379 -> 1040
    for (let i = 0; i < launchSeeds.length; i++) {
      if (result.dieAlignment[i] < SETTLE_ALIGNMENT) launchSeeds[i] = seeds() * 0x100000000;
    }
  }
  const chosen = best as RollDiagnostics;
  return Object.freeze({
    ...chosen,
    energyPerStep: Object.freeze(energyPerStep),
    attemptStarts: Object.freeze(attemptStarts),
    simulatedSteps,
    attempts,
  });
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
