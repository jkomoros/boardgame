/**
 * THE POSE A SOLID IS PRESENTED IN.
 *
 * A solid drawn facing the viewer square-on projects to a flat outline and
 * reads as the 2D sprite it replaces. This module computes the turns that lean
 * it instead: one nominated facet is aimed near the camera, at least one other
 * facet is brought far enough round to be visible, and — when asked — the
 * nominated facet's own content is rolled upright.
 *
 * ONE routine produces the pose, for a solid that has never moved and for one
 * the physics has just put down; the two callers differ only in the frame they
 * hand it and in the content-down direction they ask to have straightened. See
 * `readingPose`, and `landedContentDown` for the second of those.
 *
 * Nothing here is about dice. `presented` is "the facet the viewer is meant to
 * read", which for a die is the face carrying the value and for a 3D token will
 * be whichever face carries its art. The only inputs are a list of facet
 * directions in CSS space and an index into it.
 *
 * Everything is in the CSS frame; `screen-frame.ts` owns the map into it.
 * Angles are in the degrees CSS `rotate3d` wants, and CSS `rotate3d` is the
 * right-handed Rodrigues rotation in the same coordinate triple, so no sign
 * fixing is needed anywhere below.
 */

import {
  add,
  cross,
  dot,
  magnitude,
  normalize,
  scale as scaleVec,
  vec3,
  type Quat,
} from '../motion/die-geometry.ts';
import {
  CAMERA_AXIS,
  cssNumber,
  facetBasis,
  toScreen,
  type Vec3,
} from './screen-frame.ts';
import type { SolidSurface } from './facet-placement.ts';

/**
 * Where the presented face is pointed, in CSS space, when the solid is at rest
 * and its shape can afford the whole lean.
 *
 * Not straight at the camera (`+Z`): a solid facing the viewer square-on
 * projects to a flat outline and reads as the 2D sprite this replaces. Pointing
 * the presented face slightly down and to the left instead puts the camera
 * above and to the right of it, so a d6 shows its presented face plus the
 * faces above and to its right — a die seen on a table.
 *
 * This lean is enough ONLY while the solid's other faces are within ~90 degrees
 * of the presented one; `companionTilt` covers the rest. And it is a MAXIMUM,
 * not a constant: `presentedTiltLimit` shortens it for a solid whose facets sit
 * close together, so that the face the player is reading is always the most
 * square-on one.
 */
const RESTING_VIEW: Vec3 = normalize(vec3(-0.32, 0.26, 1));

/**
 * `RESTING_VIEW` split into the direction it leans and how far it leans, so
 * that the lean can be SHORTENED without moving to a differently-shaped pose:
 * turning `+Z` about `RESTING_TILT_AXIS` by `RESTING_TILT_DEGREES` reproduces
 * `RESTING_VIEW` exactly, and by less than that reproduces the same view seen
 * from a little nearer the face's own normal.
 */
const RESTING_TILT_AXIS: Vec3 = normalize(cross(CAMERA_AXIS, RESTING_VIEW));
const RESTING_TILT_DEGREES: number =
  (Math.acos(Math.min(1, Math.max(-1, dot(CAMERA_AXIS, RESTING_VIEW)))) * 180) / Math.PI;

/**
 * How much more square-on the presented facet must be than every other facet on
 * the solid, in degrees. See `presentedTiltLimit`, which is what enforces it.
 *
 * Four degrees is not a tolerance, it is a legibility budget: two facets within
 * a couple of degrees of each other in foreshortening are equally big on screen
 * and the player has no reason to read one rather than the other, which on a
 * d20 is the difference between reading 17 and reading 18. Larger would be
 * safer still and buys nothing — the presented facet is also the CENTRED one,
 * so a real margin plus the centre is unambiguous — while costing lean, and
 * lean is what makes the solid read as a solid at all.
 */
const READING_MARGIN = 4;

/** `RESTING_VIEW`'s direction, leaned only `tiltDegrees` off the camera axis. */
function restingView(tiltDegrees: number): Vec3 {
  return applyTurn(CAMERA_AXIS, { axis: RESTING_TILT_AXIS, degrees: tiltDegrees });
}

/**
 * How far off the camera axis the most face-on of the OTHER facets is allowed
 * to sit, in degrees. Past 90 it is a back-face and `backface-visibility:
 * hidden` culls it outright, so a fixed tilt of 23.6 degrees (`RESTING_VIEW`)
 * is not enough for a solid whose faces are far apart in normal angle: a
 * tetrahedron's other three normals are 109.47 degrees from the presented one
 * — 86 to 133 degrees off the camera axis once RESTING_VIEW is applied — so a
 * d4 renders as a single flat triangle, which is exactly the 2D die this
 * replaces. (A d8's are 70.5 apart, a d12's 63.4, a d20's 41.8, a barrel's
 * side faces closer still; none of them need any of this.)
 *
 * 75 rather than a hair under 90 because a facet within a few degrees of
 * edge-on is a sliver, not a visible face.
 */
const COMPANION_VIEW_LIMIT = 75;

/**
 * The most the pose may be tilted to bring that facet into view. The presented
 * face is the value the player has to read, so it stays the dominant one: the
 * tilt moves every direction by at most its own angle, which keeps the
 * presented face within 23.6 + 30 degrees of the camera axis while the facet
 * it reveals sits at 75.
 */
const MAX_COMPANION_TILT = 30;

/** An axis-angle turn, in the degrees CSS wants. */
export interface Turn {
  readonly axis: Vec3;
  readonly degrees: number;
}

/** One turn, as the CSS function that performs it. */
export function rotate3d(turn: Turn): string {
  const { axis, degrees } = turn;
  return `rotate3d(${cssNumber(axis[0])},${cssNumber(axis[1])},${cssNumber(axis[2])},${cssNumber(degrees)}deg)`;
}

/**
 * The minimal rotation carrying direction `from` to direction `to`, or `null`
 * when they already agree: axis `from x to`, angle `atan2(|from x to|, from .
 * to)`. CSS `rotate3d` is the right-handed Rodrigues rotation in the same
 * coordinate triple, so no sign fixing.
 */
export function minimalTurn(from: Vec3, to: Vec3): Turn | null {
  const axis = cross(from, to);
  const sine = magnitude(axis);
  const cosine = dot(from, to);
  if (sine < 1e-9) {
    // Already there, or pointing exactly backwards: a half turn about any
    // perpendicular then does it, and `facetBasis` names one.
    if (cosine > 0) return null;
    return { axis: facetBasis(from).u, degrees: 180 };
  }
  return {
    axis: scaleVec(axis, 1 / sine),
    degrees: (Math.atan2(sine, cosine) * 180) / Math.PI,
  };
}

/** Rodrigues: `v` turned about the UNIT axis of `turn`, right-handed. */
export function applyTurn(v: Vec3, turn: Turn | null): Vec3 {
  if (!turn) return v;
  const radians = (turn.degrees * Math.PI) / 180;
  const cosine = Math.cos(radians);
  return add(
    add(scaleVec(v, cosine), scaleVec(cross(turn.axis, v), Math.sin(radians))),
    scaleVec(turn.axis, dot(turn.axis, v) * (1 - cosine)),
  );
}

/** Rotate `v` by the unit quaternion `q` (`v + 2w(a x v) + 2a x (a x v)`). */
export function rotateByQuat(q: Quat, v: Vec3): Vec3 {
  const axis = vec3(q[0], q[1], q[2]);
  const t = scaleVec(cross(axis, v), 2);
  return add(add(v, scaleVec(t, q[3])), cross(axis, t));
}

/**
 * The extra turn, if any, that brings `companion` — the most face-on of the
 * facets OTHER than the presented one, already in its resting direction — to
 * `COMPANION_VIEW_LIMIT` of the camera axis, so that the solid reads as a solid
 * and not as one flat polygon. `null` when it is visible enough already, which
 * is every solid except the tetrahedron. See `COMPANION_VIEW_LIMIT`.
 *
 * WHAT THIS BUYS IS FRAGILE, and the fragility is why the content roll is a
 * roll of the PICTURE (see `readingPose`'s step 3). The tilt aims at one named
 * direction; any later turn that is not about the camera axis moves that
 * direction off the limit it was just brought to. A landed die used to get
 * exactly such a turn — its content was straightened about the presented
 * facet's own normal, which fixes that normal and so looked safe — and it cost
 * the d4 the whole of this: measured over 12 seeded landings, the best
 * companion left the tilt at exactly 75.0 degrees and arrived on screen
 * anywhere from 71.0 to 88.4, with 3 of the 12 past the point the facet is a
 * hairline and the tetrahedron reads as a flat triangle again.
 *
 * Rotating about `companion x cameraAxis` carries `companion` towards the
 * camera along the shortest path, and moves everything else — the presented
 * face included — by at most the same angle. Which is exactly why `headroom`
 * exists: it is how much further the presented face may be leaned before it
 * stops being the most square-on facet (`presentedTiltLimit`), and "at most the
 * same angle" is what makes capping the tilt at it sufficient.
 */
export function companionTilt(companion: Vec3, headroom: number): Turn | null {
  const cosine = Math.min(1, Math.max(-1, dot(companion, CAMERA_AXIS)));
  const offAxis = (Math.acos(cosine) * 180) / Math.PI;
  const degrees = Math.min(offAxis - COMPANION_VIEW_LIMIT, MAX_COMPANION_TILT, headroom);
  if (!(degrees > 0)) return null;
  const axis = cross(companion, CAMERA_AXIS);
  const sine = magnitude(axis);
  // Dead ahead or dead behind: no shortest path to pick, and dead ahead is
  // not a case that needs one anyway.
  if (sine < 1e-9) return null;
  return { axis: scaleVec(axis, 1 / sine), degrees };
}

/**
 * The roll, about the camera axis, that turns `posed` -- a facet's local +y
 * after the pose has been applied -- back to screen-down, or `null` when it is
 * there already or has no screen direction to speak of.
 *
 * About the CAMERA axis rather than about the facet's own normal because a
 * roll about `+Z` leaves every direction's z component alone: it cannot move a
 * facet towards or away from the viewer, so it cannot undo `companionTilt`'s
 * work or change which facet is nearest the camera. It is a rotation of the
 * PICTURE, and `atan2` in the screen plane is therefore exact rather than an
 * approximation that a tilted facet would spoil.
 */
export function uprightRoll(posed: Vec3): Turn | null {
  // The facet's local +y projected on screen. Vanishes only if the facet's
  // plane contains the view direction, i.e. the facet is edge-on -- which the
  // PRESENTED facet, pointed within ~54 degrees of the camera, never is.
  if (Math.hypot(posed[0], posed[1]) < 1e-9) return null;
  const degrees = (Math.atan2(posed[0], posed[1]) * 180) / Math.PI;
  if (Math.abs(degrees) < 1e-9) return null;
  return { axis: CAMERA_AXIS, degrees };
}

/**
 * Every facet's outward normal in CSS space, in surface order
 * (`[...faces, ...capFaces]`, so a face index is also an index into this) —
 * optionally as a throw left them, by applying a landing orientation first.
 *
 * The one place the body frame becomes the screen frame for a NORMAL, which is
 * what lets the pose below be written once and used for a solid that has never
 * moved and for one the physics has just put down.
 */
export function surfaceDirections(surface: SolidSurface, landed?: Quat): readonly Vec3[] {
  return [...surface.faces, ...surface.capFaces].map((face) => normalize(toScreen(
    landed ? rotateByQuat(landed, face.normal) : face.normal)));
}

/**
 * How far off the camera axis facet `presented` may be leaned and still be the
 * most square-on facet on the solid, in degrees.
 *
 * A facet whose normal is `d` degrees from the presented one cannot get closer
 * to the camera axis than `d - a` when the presented facet sits `a` degrees off
 * it — that is the triangle inequality on the sphere, and it holds whichever
 * way the pose leans and wherever a throw happened to leave the solid. So
 * `a <= (d - READING_MARGIN) / 2` makes the presented facet the most square-on
 * by at least `READING_MARGIN`, and taking `d` as the SMALLEST angle to any
 * other facet makes it so against all of them at once.
 *
 * `d` is a property of the solid and nothing else: 109.5 degrees for a d4, 90
 * for a d6, 70.5 for a d8, 51.7 for a d10, 63.4 for a d12, 41.8 for a d20 and
 * 37.2 for a d7 (a side face to the nearest cap triangle). Only the last two
 * come out under `RESTING_TILT_DEGREES` and are actually shortened by this; for
 * every other shape the limit is slack and the pose is exactly what it was.
 *
 * Rotation-invariant, so it may be handed either frame's directions.
 */
export function presentedTiltLimit(directions: readonly Vec3[], presented: number): number {
  let nearest = 180;
  for (let index = 0; index < directions.length; index++) {
    if (index === presented) continue;
    const cosine = Math.min(1, Math.max(-1, dot(directions[presented], directions[index])));
    nearest = Math.min(nearest, (Math.acos(cosine) * 180) / Math.PI);
  }
  return Math.max(0, (nearest - READING_MARGIN) / 2);
}

/**
 * THE POSE A SOLID IS READ IN, as a list of turns, from the facet directions it
 * is being applied to.
 *
 * There is one of these and there must be exactly one. A die shows this pose
 * twice: before it has ever rolled, and after a throw has put it down. Those
 * used to be two independent producers, and they disagreed by 51.7 degrees —
 * the pre-roll pose left the presented face 22.4 degrees off the camera axis
 * and leaning down-left, the post-roll scene left it at 35 and leaning the
 * other way. On a d20 that is enough for a NEIGHBOURING triangle to be more
 * square-on than the face carrying the value, so the die announced 17 and drew
 * a big central 18; measured over 25 seeded rolls it happened 24 times. Two
 * producers of one pose is the bug, so there is now one, and the two callers
 * differ only in the frame they hand it and in one flag.
 *
 * In order of application (CSS applies a transform list left to right, so the
 * returned list reads outermost first):
 *
 *   1. `base`, the minimal turn pointing the presented facet's normal at
 *      `restingView(...)` — the lean this shape can afford. Minimal means the
 *      axis is perpendicular to that normal, so this AIMS the solid without
 *      twisting it about the face being read: it is where the camera stands,
 *      not how the solid came to rest.
 *   2. `lean`, whatever extra tilt it takes for at least one other facet to be
 *      visible (`companionTilt`), so the solid reads as a solid whatever its
 *      face count. Null for everything but a d4.
 *   3. `roll`, the turn that leaves the presented face's CONTENT the right way
 *      up — and this one only when a `contentDown` is supplied to straighten.
 *
 * The roll exists because nothing about steps 1 and 2 is a statement about
 * which way the face's own printing points: `minimalTurn` swings the presented
 * facet round to face the viewer about an axis perpendicular to its normal and
 * carries the facet's local +y wherever the shortest path leaves it. Measured
 * without the roll, a d4 presenting face 1 was 122 degrees out and a d10
 * presenting face 2 was 116 — an upside-down number.
 *
 * IT IS A ROLL OF THE PICTURE, ABOUT THE CAMERA AXIS, and that is the whole
 * reason it is safe to add to a pose that has just been carefully aimed. A
 * rotation about `+Z` leaves every direction's z component alone, so it cannot
 * move a facet towards or away from the viewer: whatever `presentedTiltLimit`
 * and `companionTilt` established about how square-on each facet is, it
 * establishes still. See `uprightRoll`.
 *
 * WHICH VECTOR IS STRAIGHTENED IS THE CALLER'S, and that is the only difference
 * between the two poses this module serves. A solid that has never moved hands
 * over `facetBasis(directions[presented]).v` — the same routine
 * `facet-placement.ts` lays the facet's box out along, so the two agree by
 * construction. A solid the PHYSICS has put down cannot: `facetBasis` is
 * defined against a FIXED direction (`SCREEN_UP`), so it is not equivariant —
 * `facetBasis(R·n).v` is not `R·facetBasis(n).v` — while the facet element's
 * actual local +y is the BODY one carried through the landing rotation. It
 * hands over `landedContentDown` instead, and gets the identical treatment.
 *
 * (The landed correction used to be a separate turn about the presented
 * facet's OWN normal, emitted on `#orient` inside the tumble's pose. It put the
 * numeral upright just as exactly, and it provably could not cost the presented
 * facet its dominance — a rotation about that normal fixes it and preserves
 * every pairwise angle. What it could and did cost was step 2: it is not a
 * rotation about the camera axis, so it moves every OTHER facet's depth, and on
 * a d4 that took the companion `companionTilt` had just placed at 75 degrees to
 * as far as 88.4 and the solid back to a flat triangle. See `companionTilt`.)
 *
 * Only the presented facet's content can be corrected, and only it should be:
 * one roll is one degree of freedom, and the other facets keep the orientation
 * their own geometry gives them, which is right — they are seen at an angle
 * anyway.
 */
export function readingPose(
  directions: readonly Vec3[],
  presented: number,
  options: {
    /**
     * The direction the presented facet's content reads DOWNWARDS in, in the
     * same frame as `directions`, or `null` to leave the content wherever the
     * aim happens to put it.
     */
    readonly contentDown: Vec3 | null;
  },
): readonly Turn[] {
  const limit = presentedTiltLimit(directions, presented);
  const tilt = Math.min(RESTING_TILT_DEGREES, limit);
  const base = minimalTurn(directions[presented], restingView(tilt));
  let companion: Vec3 | null = null;
  for (let index = 0; index < directions.length; index++) {
    if (index === presented) continue;
    const direction = applyTurn(directions[index], base);
    if (companion === null || direction[2] > companion[2]) companion = direction;
  }
  const lean = companion === null ? null : companionTilt(companion, limit - tilt);
  const roll = options.contentDown
    ? uprightRoll(applyTurn(applyTurn(options.contentDown, base), lean))
    : null;
  return [roll, lean, base].filter((turn): turn is Turn => turn !== null);
}

/**
 * The pose of a solid that has never moved, as a CSS transform list, or `'none'`
 * when it needs no turning at all.
 *
 * Content is rolled upright, because a solid nothing has happened to has to read
 * like the flat sprite it replaces. A solid the physics has moved is posed by
 * `dice-roll.ts`'s `rollScene`, from the same `readingPose` with
 * `landedContentDown` in place of the vector below.
 */
export function readingPoseTransform(surface: SolidSurface, presented: number): string {
  const directions = surfaceDirections(surface);
  // Read the presented facet's local +y from the SAME routine that lays its
  // content box out, so the correction stays tied to what is actually drawn
  // rather than to a second copy of the rule.
  const turns = readingPose(directions, presented, {
    contentDown: facetBasis(directions[presented]).v,
  });
  return turns.length ? turns.map(rotate3d).join(' ') : 'none';
}

/**
 * THE PRESENTED FACET'S CONTENT-DOWN DIRECTION, AS A LANDING LEFT IT — the
 * vector `readingPose` straightens for a solid the physics has put down.
 *
 * `readingPose`'s own default reads the presented facet's local +y back out of
 * `facetBasis(normal)`, which is legitimate for a solid that has never moved:
 * `facet-placement.ts` derives every facet's box from `facetBasis` of that same
 * normal, so the two agree by construction. They do NOT agree once the physics
 * has turned the solid. `facetBasis` is defined against a FIXED direction
 * (`SCREEN_UP`), so it is not equivariant — `facetBasis(R·n).v` is not
 * `R·facetBasis(n).v` — while the facet element's actual local +y is the body
 * one carried through the landing rotation. Uprighting the first vector leaves
 * the second at whatever angle it likes, which is the "ει" a landed d20 used to
 * read as.
 *
 * So this is the body one, carried through: `facetBasis` of the facet's BODY
 * normal, then the landing. The landing rotation acts on a CSS-frame vector as
 * the similarity `S R S` (see `screen-frame.ts`), and `toScreen` is its own
 * inverse, so that is what is written out below. The result is in the same
 * frame `surfaceDirections(surface, landed)` reports, which is the frame
 * `readingPose` wants it in.
 *
 * `null` for a face index the surface does not have, so a caller mid-render
 * gets an unstraightened die rather than a thrown exception.
 */
export function landedContentDown(
  surface: SolidSurface,
  landed: Quat,
  presented: number,
): Vec3 | null {
  const face = surface.faces[presented] ?? surface.capFaces[presented - surface.faces.length];
  if (!face) return null;
  const contentDown = facetBasis(normalize(toScreen(face.normal))).v;
  return toScreen(rotateByQuat(landed, toScreen(contentDown)));
}

/**
 * The pose a solid the physics has put down is read in, as a list of turns:
 * `readingPose` handed the facet normals AS THE THROW LEFT THEM and the content
 * direction the throw left the presented facet's printing pointing.
 *
 * The parallel of `readingPoseTransform` for the other pose, one level down —
 * it returns turns rather than CSS because `dice-roll.ts` needs the list itself
 * to pose the trajectory's positions with, not only to emit.
 */
export function landedReadingPose(
  surface: SolidSurface,
  landed: Quat,
  presented: number,
): readonly Turn[] {
  return readingPose(surfaceDirections(surface, landed), presented, {
    contentDown: landedContentDown(surface, landed, presented),
  });
}
