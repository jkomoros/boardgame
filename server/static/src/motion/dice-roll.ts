/**
 * Planning ONE die's throw: seed, trajectory, trim, and the scene it plays in.
 *
 * The `dice-*` family runs `sim -> roll -> bake`: `dice-sim.ts` is the physics,
 * this module decides WHICH throw a particular die on a particular board is
 * making and frames it, and `dice-bake.ts` turns the resulting poses into CSS.
 * Everything here is a pure function of its arguments — no DOM, no clock, no
 * component — which is what makes a die's tumble reproducible bit for bit.
 *
 * ## Why a die never re-rolls until the physics agrees
 *
 * The outcome is the SERVER's, and it is known before any pixel moves. So the
 * simulation is never asked to produce it: the die is thrown ONCE from a seed,
 * `die-faces.ts` reads which face that throw landed, and `assignFaceValues`
 * paints the server's value onto that face while permuting the rest into a
 * still-legitimate die. Re-simulating until the physics agrees would cost
 * `sides^dice` throws in expectation and is unbounded; this costs one, always.
 * The visible consequence is that the die's OTHER faces carry different numbers
 * after every roll, which is the price of a real tumble and is invisible in
 * practice.
 *
 * The seed comes from `(component ID, roll identity)` and nothing else, so a
 * remount mid-roll — a re-render, a tab returning to the foreground, a replay —
 * rebuilds the same throw bit for bit instead of re-throwing the die.
 *
 * ## The scene: why the baked trajectory is not enough on its own
 *
 * `dice-bake.ts` emits the die's pose in the SIMULATION's world: a tray centred
 * on the origin, with the die's landing spot wherever it happens to be on the
 * floor, and world-up mapped to screen-up. Rendered as-is, that world is
 * useless: the face the die landed on points straight UP THE SCREEN, which is
 * edge-on to a viewer looking down the +Z axis, so the player reads a side face
 * — and the die comes to rest on the tray's floor, about three quarters of a box
 * below the middle of the box it is laid out in, and wherever it drifted to.
 *
 * Two constant turns/translations fix that (see `rollScene`):
 *
 *   1. WHERE THE CAMERA STANDS, which is `readingPose` — the same routine, and
 *      therefore the same framing, the die is shown in before it has ever
 *      rolled. It is handed the facet normals AS THE THROW LEFT THEM, so it
 *      aims at the face this particular throw landed rather than at a fixed
 *      elevation above the tray, and it subsumes the landing square-up that
 *      used to be a step of its own: the read face is put exactly on the
 *      reading direction whether the throw settled flat (nearly all of them)
 *      or COCKED (`RollTrajectory.cocked`, a die the simulator could not settle
 *      in eight throws). There is no floor drawn, so rotating the whole world
 *      by a couple of degrees to straighten a cocked die costs nothing, and the
 *      value is never displayed on a face that is not really up.
 *   2. a RECENTRING translation, so the die comes to rest at the centre of its
 *      own box whatever spot on the floor it landed on.
 *
 * ## Why the camera travels with the die
 *
 * A solid is drawn as one box per face with `backface-visibility: hidden`, and
 * that property does not ask where the camera is: it hides a facet whose
 * accumulated matrix has `m33 <= 0`, i.e. whose outward normal leans away from
 * `+Z`. That test is only the same question as "can the camera see it?" when the
 * facet sits on the camera's axis.
 *
 * A tumble carries the solid across its own box — a median of 56px on a 100px
 * die, up to 91px — so with a camera pinned to the middle of the box the two
 * tests disagree for a facet up to ~17 degrees off axis. The disagreement is
 * visible both ways: a facet the camera CAN see gets culled (a see-through hole
 * in the solid, 10-25% of the silhouette, lasting several frames), and a facet
 * it cannot see gets drawn over the front of the die. Measured over 20 seeded
 * rolls per shape, a d20 spent 22% of its frames with a hole in it.
 *
 * So the die's travel is applied OUTSIDE the projection and its pose inside:
 * every keyframe is
 *
 *     translate3d(travel) perspective(D) translate3d(0,0,depth) <turns> matrix3d(pose)
 *
 * CSS applies a `perspective()` function to everything to its right in the list,
 * so the solid is projected about ITS OWN centre and only then moved into place.
 * The camera therefore rides with the die, every facet stays within a
 * circumradius of the axis (~6 degrees on a 100px die), and the orthographic
 * test `backface-visibility` performs is the one the camera would give. The
 * price is that lateral travel is no longer foreshortened, which is a few pixels
 * on a throw and nothing a player can see.
 *
 * The depth term stays INSIDE the projection: it is on the axis, so it cannot
 * reopen the disagreement, and it is what makes the die grow as it comes up off
 * the tray toward the viewer.
 */

import {
  magnitude,
  scale as scaleVec,
  subtract,
  vec3,
  type DieGeometry,
  type Quat,
  type Vec3,
} from './die-geometry.ts';
import { simulateRoll, type DieSample, type DieTrajectory, type RollTrajectory } from './dice-sim.ts';
import { trajectoryCurve } from './dice-bake.ts';
import { cssNumber, toScreen } from '../solid/screen-frame.ts';
import {
  applyTurn,
  readingPose,
  rotate3d,
  surfaceDirections,
  type Turn,
} from '../solid/reading-pose.ts';

/**
 * HALF-extents of the tray a die is thrown in, in die BOUNDING radii — the unit
 * `dice-sim.ts` normalises to, which is `DieGeometry.boundingRadius` and not
 * `circumradius`.
 *
 * The tray is invisible, so its size is a purely visual budget: it is how far
 * the die may travel from the centre of its own box. At 1.6 a d6 stays within
 * about one box width of centre and drops a little over one box height, which
 * reads as a throw without the die leaving the region a game laid out for it. A
 * roomier tray (`dice-sim.ts` recommends 4 for three dice) settles more
 * reliably, but that recommendation is about dice knocking EACH OTHER cocked; a
 * lone die measured over 300 seeds per solid landed cocked once, for a d7, and
 * the pose's aim at the landed face covers even that.
 *
 * `dice-sim.ts` rejects anything under 1.5 (its spawn needs the clearance).
 */
export const TRAY_BOUNDS = Object.freeze({ x: 1.6, y: 2.0, z: 1.6 });

/**
 * One 60Hz frame, in ms: the grid the baked curve is sampled on.
 *
 * The compiler clamps the resolution to [2, 256], so a roll past ~4.3 seconds
 * samples coarser than a frame. `dice-sim.ts` caps a throw at 5 seconds and the
 * measured 99th percentile is under 3, so that ceiling is not normally reached.
 */
export const FRAME_MS = 16.7;

/**
 * The seed for one roll, from the identity of the roll and nothing else.
 *
 * FNV-1a over `id#identity`: the point is only that distinct rolls land on
 * distinct uint32s without structure, which `dice-sim.ts`'s own splitmix
 * avalanche then spreads across its stream. It IS the roll's identity — a test
 * that wants to know which throw a die must be playing has to derive it the same
 * way, and so would a replay or a dice tray built later.
 *
 * `rollIdentity` is the die's own `DynamicValues.RollCount` wherever the server
 * reports one, and the game's state version only as a fallback — see
 * `boardgame-die.ts`'s `_rollIdentity`, which is where the choice is made and
 * argued.
 */
export function dieRollSeed(componentId: string, rollIdentity: number): number {
  const text = `${componentId}#${rollIdentity}`;
  let hash = 0x811c9dc5;
  for (let index = 0; index < text.length; index++) {
    hash ^= text.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return hash >>> 0;
}

/**
 * The throw a die of this geometry makes for this identity. Pure, deterministic
 * and independent of the DOM: same arguments, same trajectory, forever.
 */
export function dieRollTrajectory(
  geometry: DieGeometry,
  componentId: string,
  rollIdentity: number,
): RollTrajectory {
  return simulateRoll({
    seed: dieRollSeed(componentId, rollIdentity),
    geometry,
    dieCount: 1,
    bounds: TRAY_BOUNDS,
  });
}

/**
 * How far the die may still be from its final pose for a frame to count as
 * part of the trailing HOLD rather than as motion: half a thousandth of a
 * circumradius, and a fiftieth of a degree.
 *
 * `dice-sim.ts` only declares a die at rest after `REST_HOLD_SECONDS` (0.3s)
 * of continuous stillness, and it emits that hold as samples, so the last
 * ~300ms of every trajectory is the die frozen on its final pose. Animating it
 * is not neutral: the roll is GATED, so those 300ms are 30% of a median roll
 * during which the whole game's animation cycle waits on a die that has already
 * stopped. `settledTrajectory` cuts them off.
 *
 * The tolerances are deliberately far below anything a screen can show — at
 * pig's 100px die they are 0.025px of travel and 0.035px of surface swing — so
 * this cannot cut a frame that a player could tell apart from the last one.
 * Measured over 1800 throws (9 face counts x 200 seeds) it removes a median of
 * 300ms and never once changes which face `presentedFaceIndex` reads.
 */
const SETTLED_POSITION_TOLERANCE = 5e-4;
const SETTLED_ANGLE_TOLERANCE = 0.02;

/** The angle between two unit quaternions' orientations, in degrees. */
function quatAngleDegrees(a: Quat, b: Quat): number {
  const cosine = Math.min(1, Math.abs(a[0] * b[0] + a[1] * b[1] + a[2] * b[2] + a[3] * b[3]));
  return (Math.acos(cosine) * 360) / Math.PI;
}

/**
 * The same throw with its trailing dead hold removed: the samples up to and
 * including the FIRST one from which the die never again moves visibly (see
 * `SETTLED_POSITION_TOLERANCE`).
 *
 * Part of the roll's identity — the animation's duration is the trimmed
 * trajectory's last sample time, so anything recomputing what a die must be
 * playing (a test, a replay) has to trim the same way.
 *
 * The returned trajectory's `restingOrientation` is its own final sample's, so
 * `restingTransform` and `trajectoryCurve(...)(1)` stay byte-identical and the
 * scene's reading pose is aimed at the pose actually rendered last.
 */
export function settledTrajectory(die: DieTrajectory): DieTrajectory {
  const samples = die.samples;
  const last = samples[samples.length - 1];
  let first = samples.length - 1;
  while (first > 1) {
    const candidate = samples[first - 1];
    if (magnitude(subtract(candidate.position, last.position)) > SETTLED_POSITION_TOLERANCE) break;
    if (quatAngleDegrees(candidate.orientation, last.orientation) > SETTLED_ANGLE_TOLERANCE) break;
    first--;
  }
  // Nothing to cut, or nothing left to play if it were cut: hand back the
  // throw untouched rather than a trajectory the bake would refuse.
  if (first >= samples.length - 1 || !(samples[first].t > 0)) return die;
  const trimmed = samples.slice(0, first + 1);
  return { samples: trimmed, restingOrientation: trimmed[trimmed.length - 1].orientation };
}

/**
 * How far in front of the solid the camera stands, in die sizes.
 *
 * The single definition: `boardgame-die.ts` interpolates it into the CSS that
 * frames a die which has never rolled, and every keyframe of a roll writes it as
 * a literal `perspective()` (see `rollScene`). A roll framed at a different
 * depth from the resting die would change the solid's size the moment its
 * animation was removed.
 */
export const PERSPECTIVE_DEPTH_DIE_SIZES = 6;

/** The origin of the die's box, in the space the emitted transform poses. */
const ORIGIN: Vec3 = vec3(0, 0, 0);

/**
 * Where the die's centre sits on screen at trajectory time `t`, in px, after
 * the reading pose has been applied.
 *
 * The pose turns are applied with `reduceRight` because the emitted list is
 * outermost first, so it is applied last first.
 *
 * `radiusPx` is HALF THE DIE'S BOX on screen, which is what the caller can
 * measure (`#stage`'s font-size, halved) and what the solid is drawn at. The
 * trajectory is in `boundingRadius` units, and for every closed-form solid
 * those are the same unit. For a BARREL they are not: the die is drawn 2.1-2.6x
 * larger than the sphere it was simulated in, so one trajectory unit is still
 * one die half-box here rather than one drawn bounding radius. The consequence
 * is that a barrel roams the same number of PIXELS a d20 does while being
 * larger, i.e. its tray is tighter relative to itself — which reads as a
 * heavier die and is the conservative direction. Scaling the travel up to match
 * the drawn size instead would have a d7 flying 190px on a 100px die, well
 * outside anything a layout budgeted for it.
 */
function posedPosition(
  samples: readonly DieSample[],
  turns: readonly Turn[],
  radiusPx: number,
  t: number,
): Vec3 {
  return turns.reduceRight(
    (point, turn) => applyTurn(point, turn),
    scaleVec(toScreen(positionAt(samples, t)), radiusPx));
}

/**
 * The simulated position at time `t`, linearly interpolated between samples and
 * clamped to the trajectory's own ends.
 *
 * Deliberately the same shape of interpolation `dice-bake.ts` performs on the
 * pose it bakes, evaluated at the same time for the same progress: the two halves
 * of one keyframe would otherwise describe the die at two different instants.
 */
function positionAt(samples: readonly DieSample[], t: number): Vec3 {
  const last = samples.length - 1;
  if (t <= samples[0].t) return samples[0].position;
  if (t >= samples[last].t) return samples[last].position;
  let low = 0;
  let high = last;
  while (low < high) {
    const middle = (low + high + 1) >> 1;
    if (samples[middle].t <= t) low = middle;
    else high = middle - 1;
  }
  const span = samples[low + 1].t - samples[low].t;
  const u = span > 0 ? (t - samples[low].t) / span : 0;
  const a = samples[low].position;
  const b = samples[low + 1].position;
  return vec3(a[0] + (b[0] - a[0]) * u, a[1] + (b[1] - a[1]) * u, a[2] + (b[2] - a[2]) * u);
}

/** One roll, as the transform `#inner` carries for the whole of it. */
export interface RollScene {
  /** The transform at animation progress `p`, clamped to [0, 1]. */
  readonly transform: (progress: number) => string;
  /**
   * What the element must hold once the animation is gone. Byte-identical to
   * `transform(1)` because it IS `transform(1)`: animations run with
   * `fill: 'none'`, so a single rounding digit of disagreement shows up as the
   * die twitching as it settles.
   */
  readonly resting: string;
}

/**
 * One simulated throw, as the CSS `#inner` carries for every frame of it.
 *
 * The pose is `readingPose` handed the facet normals AS THE THROW LEFT THEM, so
 * a rolled die is framed exactly the way a die that has never rolled is framed —
 * see `readingPose` for why that has to be one routine and not two. Note what it
 * is NOT allowed to do: `minimalTurn` carries the landed face round to the
 * reading direction about an axis perpendicular to that face's own normal, so
 * the pose aims the camera and cannot twist the die about the face being read.
 * Whatever roll the simulation stopped at survives, which is what a real die
 * does.
 *
 * The emitted list is, outermost first:
 *
 *     translate3d(travel) perspective(D) translate3d(0,0,depth) <turns> matrix3d(pose)
 *
 * `perspective()` projects everything to its RIGHT, so the solid is projected
 * about its own centre and only then moved into place — the camera rides with
 * the die, which is what keeps `backface-visibility`'s orthographic culling
 * honest. See the file docs for the measurement behind that.
 *
 * The travel is minus the POSED resting position, so it is zero at progress 1
 * exactly and the die always comes to rest dead centre in its own box, whatever
 * spot on the tray's floor it landed on.
 *
 * `matrix3d` comes from `dice-bake.ts` with the trajectory's positions removed:
 * the bake owns the physics-to-CSS reflection and the slerp, and this module
 * owns where the resulting solid is put. Every number is a literal — a `var()`
 * or `calc()` anywhere in a transform keyframe forfeits compositing and drops a
 * multi-second tumble onto the main thread, which is exactly the behaviour this
 * feature replaces.
 */
export function rollScene(
  geometry: DieGeometry,
  die: DieTrajectory,
  presented: number,
  radiusPx: number,
  durationMs: number,
): RollScene {
  // A turn `minimalTurn` reports as a fraction of a millidegree rather than as
  // `null` would put a dead `rotate3d(..., 0deg)` in every one of up to 256
  // keyframes. Dropped BEFORE the travel below is computed, so the offset is
  // minus the position the emitted list actually poses the die at.
  const turns = readingPose(
    surfaceDirections(geometry, die.restingOrientation), presented,
    { uprightContent: false },
  ).filter((turn) => Math.abs(turn.degrees) > 1e-4);
  const samples = die.samples;
  const first = samples[0].t;
  const last = samples[samples.length - 1].t;
  const rest = posedPosition(samples, turns, radiusPx, last);
  // The same throw with its travel taken out: what is left is the die's
  // ORIENTATION, which is the only part of the pose that belongs inside the
  // projection. The bake's own resting/curve agreement carries over unchanged,
  // because both are computed from this same stripped trajectory.
  const spinning: DieTrajectory = {
    samples: samples.map((sample) => ({
      t: sample.t,
      position: ORIGIN,
      orientation: sample.orientation,
    })),
    restingOrientation: die.restingOrientation,
  };
  const spin = trajectoryCurve(spinning, durationMs, { radiusPx });
  const depthPx = PERSPECTIVE_DEPTH_DIE_SIZES * 2 * radiusPx;
  const pose = turns.length ? ` ${turns.map(rotate3d).join(' ')}` : '';
  const transform = (progress: number): string => {
    const clamped = progress <= 0 ? 0 : progress >= 1 ? 1 : progress;
    const t = Math.min(last, Math.max(first, first + clamped * durationMs));
    const travel = subtract(posedPosition(samples, turns, radiusPx, t), rest);
    return `translate3d(${cssNumber(travel[0])}px,${cssNumber(travel[1])}px,0px)`
      + ` perspective(${cssNumber(depthPx)}px)`
      + ` translate3d(0px,0px,${cssNumber(travel[2])}px)`
      + `${pose} ${spin(clamped)}`;
  };
  // The resting value is `transform(1)` itself, not the bake's `restingTransform`
  // composed by hand with a prefix: the WHOLE list has to agree, not just its
  // last function, and calling the emitter is the only way the two can be
  // byte-identical by construction rather than by inspection.
  return { transform, resting: transform(1) };
}
