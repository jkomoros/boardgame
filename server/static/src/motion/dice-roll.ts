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
 * Two constant turns/translations fix that, composed in front of every keyframe
 * as ONE literal prefix (see `sceneTransform`):
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
 * Both are the same for every keyframe of one roll, so composing them here
 * rather than in `dice-bake.ts` costs one string concatenation per keyframe and
 * keeps the bake a pure function of the physics.
 */

import {
  magnitude,
  scale as scaleVec,
  subtract,
  type DieGeometry,
  type Quat,
} from './die-geometry.ts';
import { simulateRoll, type DieTrajectory, type RollTrajectory } from './dice-sim.ts';
import { cssNumber, toScreen } from '../solid/screen-frame.ts';
import { applyTurn, readingPose, rotate3d, surfaceDirections } from '../solid/reading-pose.ts';

/**
 * HALF-extents of the tray a die is thrown in, in die circumradii.
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
 * The constant prefix that turns one simulated world into one rendered scene:
 * `translate3d(recentre) <the reading pose>`, applied in front of every baked
 * `matrix3d`.
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
 * CSS applies a transform list left to right, so the list reads outermost
 * first: a point is posed, then moved so that the RESTING pose lands on the
 * origin. The recentring offset is therefore minus the POSED resting position
 * and has to be computed after the turns, not before.
 *
 * Every number here is a literal: a `var()` or `calc()` anywhere in a transform
 * keyframe forfeits compositing and drops a multi-second tumble onto the main
 * thread, which is exactly the behaviour this feature replaces.
 */
export function sceneTransform(
  geometry: DieGeometry,
  die: DieTrajectory,
  presented: number,
  radiusPx: number,
): string {
  // A turn `minimalTurn` reports as a fraction of a millidegree rather than as
  // `null` would put a dead `rotate3d(..., 0deg)` in every one of up to 256
  // keyframes. Dropped BEFORE the recentring below is computed, so the offset
  // is minus the position the emitted list actually poses the die at.
  const turns = readingPose(
    surfaceDirections(geometry, die.restingOrientation), presented,
    { uprightContent: false },
  ).filter((turn) => Math.abs(turn.degrees) > 1e-4);
  const rest = die.samples[die.samples.length - 1].position;
  // reduceRight: the list is outermost first, so it is applied last first.
  const posed = turns.reduceRight(
    (point, turn) => applyTurn(point, turn), scaleVec(toScreen(rest), radiusPx));
  const recentre = `translate3d(${cssNumber(-posed[0])}px,${cssNumber(-posed[1])}px,${cssNumber(-posed[2])}px)`;
  return turns.length ? `${recentre} ${turns.map(rotate3d).join(' ')}` : recentre;
}
