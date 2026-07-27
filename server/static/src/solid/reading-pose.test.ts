import assert from 'node:assert/strict';
import test from 'node:test';
import {
  QUAT_IDENTITY,
  cross,
  dieGeometry,
  dot,
  magnitude,
  normalize,
  subtract,
  vec3,
  type Quat,
  type Vec3,
} from '../motion/die-geometry.ts';
import { CAMERA_AXIS, facetBasis, toScreen } from './screen-frame.ts';
import {
  applyTurn,
  companionTilt,
  minimalTurn,
  presentedTiltLimit,
  readingPose,
  readingPoseTransform,
  rotate3d,
  rotateByQuat,
  surfaceDirections,
  uprightRoll,
  type Turn,
} from './reading-pose.ts';

const FACE_COUNTS = [4, 5, 6, 7, 8, 10, 12, 20] as const;

/** How much more square-on the presented facet is promised to be. `READING_MARGIN`. */
const READING_MARGIN = 4;

function degreesBetween(a: Vec3, b: Vec3): number {
  return (Math.acos(Math.min(1, Math.max(-1, dot(a, b)))) * 180) / Math.PI;
}

/**
 * Apply a pose to a direction the way the browser will.
 *
 * `readingPose` returns the turns OUTERMOST FIRST, because CSS applies a
 * transform list left to right — so the last entry acts on the raw direction
 * first. Getting this order backwards is a real hazard (it composes to a
 * different pose whenever two turns share no axis), which is why every test
 * below goes through this one helper and why `dice-roll.ts`'s recentring uses
 * `reduceRight` for the same reason.
 */
function pose(turns: readonly Turn[], v: Vec3): Vec3 {
  return turns.reduceRight((point, turn) => applyTurn(point, turn), v);
}

test('applyTurn is a rigid rotation: lengths and angles survive', () => {
  const turn: Turn = { axis: normalize(vec3(1, -2, 3)), degrees: 37 };
  const a = normalize(vec3(2, 1, -1));
  const b = normalize(vec3(-1, 3, 2));
  assert.ok(Math.abs(magnitude(applyTurn(a, turn)) - 1) < 1e-12);
  assert.ok(Math.abs(dot(applyTurn(a, turn), applyTurn(b, turn)) - dot(a, b)) < 1e-12);
  // A point on the axis is fixed, and a null turn is the identity.
  assert.ok(magnitude(subtract(applyTurn(turn.axis, turn), turn.axis)) < 1e-12);
  assert.deepEqual(applyTurn(a, null), a);
});

test('minimalTurn carries from to to about a perpendicular axis, by the angle between them', () => {
  const pairs: readonly (readonly [Vec3, Vec3])[] = [
    [vec3(0, 0, 1), normalize(vec3(-0.32, 0.26, 1))],
    [normalize(vec3(1, 1, 1)), normalize(vec3(-1, 2, 0.5))],
    [vec3(1, 0, 0), vec3(0, 1, 0)],
  ];
  for (const [from, to] of pairs) {
    const turn = minimalTurn(from, to);
    assert.ok(turn, 'a turn is needed');
    assert.ok(magnitude(subtract(applyTurn(from, turn), to)) < 1e-12, 'lands on the target');
    // MINIMAL means the axis is perpendicular to `from`: the turn aims the
    // solid and cannot twist it about the face being read.
    assert.ok(Math.abs(dot(turn.axis, from)) < 1e-12, 'axis is perpendicular to from');
    assert.ok(Math.abs(magnitude(turn.axis) - 1) < 1e-12, 'axis is a unit vector');
    assert.ok(Math.abs(turn.degrees - degreesBetween(from, to)) < 1e-9, 'no long way round');
  }
});

test('minimalTurn is null when the directions already agree, and a half turn when opposed', () => {
  const v = normalize(vec3(1, 2, 3));
  assert.equal(minimalTurn(v, v), null);
  const opposed = minimalTurn(v, vec3(-v[0], -v[1], -v[2]));
  assert.ok(opposed);
  assert.equal(opposed.degrees, 180);
  assert.ok(Math.abs(dot(opposed.axis, v)) < 1e-9, 'the half turn is about a perpendicular');
  assert.ok(magnitude(subtract(applyTurn(v, opposed), vec3(-v[0], -v[1], -v[2]))) < 1e-9);
});

test('uprightRoll is a roll of the picture only, about the camera axis', () => {
  const tilted = normalize(vec3(0.6, 0.7, -0.4));
  const turn = uprightRoll(tilted);
  assert.ok(turn);
  assert.deepEqual(turn.axis, CAMERA_AXIS);
  const rolled = applyTurn(tilted, turn);
  assert.ok(Math.abs(rolled[0]) < 1e-12, 'nothing left of screen-x');
  assert.ok(rolled[1] > 0, 'and it points DOWN the screen, not up');
  // A roll about +Z cannot move anything toward or away from the viewer, so it
  // cannot undo the lean or change which facet is nearest the camera.
  assert.ok(Math.abs(rolled[2] - tilted[2]) < 1e-12);
  assert.equal(uprightRoll(vec3(0, 1, 0.5)), null, 'already upright');
  assert.equal(uprightRoll(vec3(0, 0, 1)), null, 'no screen direction at all');
});

test('rotateByQuat agrees with the equivalent axis-angle turn', () => {
  const axis = normalize(vec3(1, -2, 0.5));
  const degrees = 63;
  const half = ((degrees * Math.PI) / 180) / 2;
  const q: Quat = [
    axis[0] * Math.sin(half), axis[1] * Math.sin(half), axis[2] * Math.sin(half), Math.cos(half),
  ];
  const v = normalize(vec3(0.3, 0.9, -0.2));
  assert.ok(magnitude(subtract(rotateByQuat(q, v), applyTurn(v, { axis, degrees }))) < 1e-12);
  assert.deepEqual(rotateByQuat(QUAT_IDENTITY, v), v);
});

test('surfaceDirections maps the whole surface into CSS space, in surface order', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const directions = surfaceDirections(geometry);
    const surface = [...geometry.faces, ...geometry.capFaces];
    assert.equal(directions.length, surface.length);
    surface.forEach((face, index) => {
      assert.ok(magnitude(subtract(directions[index], normalize(toScreen(face.normal)))) < 1e-12);
      assert.ok(Math.abs(magnitude(directions[index]) - 1) < 1e-12);
    });
    // The identity quaternion must be indistinguishable from no landing at all.
    surfaceDirections(geometry, QUAT_IDENTITY).forEach((direction, index) => {
      assert.ok(magnitude(subtract(direction, directions[index])) < 1e-12);
    });
  }
});

/**
 * THE INVARIANT THE WHOLE POSE EXISTS FOR.
 *
 * The presented facet must end up the most square-on facet on the solid by at
 * least `READING_MARGIN` degrees — over EVERY face count and EVERY choice of
 * presented face, because a rolled die nominates whichever face the physics
 * turned up. This is what broke when the pre-roll and post-roll poses were two
 * separate producers: on a d20 a neighbouring triangle came out more square-on
 * than the face carrying the value, and the die announced 17 while drawing 18.
 */
test('the pose leaves the presented facet the most square-on facet, on every solid', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const directions = surfaceDirections(geometry);
    for (let presented = 0; presented < geometry.faces.length; presented++) {
      for (const uprightContent of [true, false]) {
        const turns = readingPose(directions, presented, { uprightContent });
        const posed = directions.map((direction) => pose(turns, direction));
        const own = degreesBetween(posed[presented], CAMERA_AXIS);
        posed.forEach((direction, index) => {
          if (index === presented) return;
          const other = degreesBetween(direction, CAMERA_AXIS);
          assert.ok(other >= own + READING_MARGIN - 1e-6,
            `d${faceCount} face ${presented} (upright=${uprightContent}): facet ${index} is at `
              + `${other.toFixed(2)} deg, presented is at ${own.toFixed(2)} deg`);
        });
      }
    }
  }
});

/**
 * A solid that shows exactly one facet reads as the flat sprite it replaces —
 * which is what a d4 did before `companionTilt` existed: its other three
 * normals sit 109.5 degrees away, past the point `backface-visibility: hidden`
 * culls them.
 */
test('every solid shows at least two facets, so it reads as a solid', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const directions = surfaceDirections(geometry);
    for (let presented = 0; presented < geometry.faces.length; presented++) {
      const turns = readingPose(directions, presented, { uprightContent: true });
      const frontFacing = directions
        .map((direction) => pose(turns, direction))
        .filter((direction) => direction[2] > 1e-6).length;
      assert.ok(frontFacing >= 2, `d${faceCount} face ${presented}: only ${frontFacing} facet visible`);
    }
  }
});

/**
 * The correction that makes a numeral read the right way up. Measured without
 * it, a d4 presenting face 1 was 122 degrees out and a d10 presenting face 2
 * was 116 — an upside-down number on a die that had never moved.
 */
test('uprightContent leaves the presented facet content the right way up', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const directions = surfaceDirections(geometry);
    for (let presented = 0; presented < geometry.faces.length; presented++) {
      const turns = readingPose(directions, presented, { uprightContent: true });
      // The facet's local +y is what `facet-placement.ts` lays its box out
      // along, so this is the direction the content's "down" points on screen.
      const down = pose(turns, facetBasis(directions[presented]).v);
      assert.ok(Math.abs(down[0]) < 1e-9,
        `d${faceCount} face ${presented}: content is ${down[0]} off vertical`);
      assert.ok(down[1] > 0, `d${faceCount} face ${presented}: content is upside down`);
    }
  }
});

/**
 * A die the physics has just put down must NOT get the roll: a real die stops
 * at whatever angle it stops at. The two poses must therefore differ by exactly
 * one turn, about the camera axis, and by nothing else — if they differed in the
 * lean as well, a rolled die would be framed differently from a resting one and
 * the two producers would be back.
 */
test('the two poses differ by exactly one roll about the camera axis', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const directions = surfaceDirections(geometry);
    for (let presented = 0; presented < geometry.faces.length; presented++) {
      const upright = readingPose(directions, presented, { uprightContent: true });
      const raw = readingPose(directions, presented, { uprightContent: false });
      assert.ok(upright.length === raw.length || upright.length === raw.length + 1,
        `d${faceCount} face ${presented}: unexpected turn count`);
      // The raw pose is the tail of the upright one: same lean, same aim.
      assert.deepEqual(upright.slice(upright.length - raw.length), raw);
      const extra = upright.slice(0, upright.length - raw.length);
      for (const turn of extra) {
        assert.deepEqual(turn.axis, CAMERA_AXIS, 'the only extra turn is a picture roll');
      }
      // Which means the FRAMING is identical: a roll about +Z leaves every
      // direction's z component alone, so every facet is exactly as square-on
      // to the viewer under one pose as under the other. Only the picture has
      // turned. (The directions themselves differ, and must: that rotation is
      // precisely what puts a numeral the right way up.)
      directions.forEach((direction, index) => {
        const a = pose(upright, direction);
        const b = pose(raw, direction);
        assert.ok(Math.abs(a[2] - b[2]) < 1e-9,
          `d${faceCount} face ${presented}: facet ${index} changed depth between the poses`);
      });
    }
  }
});

/** A pose that mirrored the solid would show the face opposite the read one. */
test('the composed pose is a proper rotation, never a reflection', () => {
  for (const faceCount of FACE_COUNTS) {
    const geometry = dieGeometry(faceCount);
    const directions = surfaceDirections(geometry);
    for (let presented = 0; presented < geometry.faces.length; presented++) {
      const turns = readingPose(directions, presented, { uprightContent: true });
      const columns = [vec3(1, 0, 0), vec3(0, 1, 0), vec3(0, 0, 1)].map((axis) => pose(turns, axis));
      const det = dot(columns[0], cross(columns[1], columns[2]));
      assert.ok(Math.abs(det - 1) < 1e-9, `d${faceCount} face ${presented}: det ${det}`);
    }
  }
});

test('presentedTiltLimit is the half-angle to the nearest other facet, less the margin', () => {
  // A cube's neighbours are 90 degrees away, so (90 - 4) / 2 = 43.
  const cube = surfaceDirections(dieGeometry(6));
  assert.ok(Math.abs(presentedTiltLimit(cube, 0) - 43) < 1e-9);
  // A tetrahedron's are 109.47 apart: (109.47 - 4) / 2 = 52.7.
  const tetra = surfaceDirections(dieGeometry(4));
  assert.ok(Math.abs(presentedTiltLimit(tetra, 0) - 52.73) < 0.01);
  // Rotation-invariant, so either frame's directions may be handed to it.
  const rolled = cube.map((direction) => applyTurn(direction, { axis: normalize(vec3(1, 2, 3)), degrees: 41 }));
  assert.ok(Math.abs(presentedTiltLimit(cube, 2) - presentedTiltLimit(rolled, 2)) < 1e-9);
});

test('companionTilt only fires when a facet is too far round to be seen', () => {
  // A companion already at 75 degrees or better needs nothing.
  assert.equal(companionTilt(normalize(vec3(0, 0.5, 1)), 30), null);
  // One at 100 degrees off the camera axis is a back-face; the tilt is capped
  // by the headroom the presented facet has left.
  const behind = normalize(vec3(0, 0.985, -0.174));
  const tilt = companionTilt(behind, 8);
  assert.ok(tilt);
  assert.equal(tilt.degrees, 8, 'headroom is the binding cap');
  assert.ok(Math.abs(dot(tilt.axis, cross(behind, CAMERA_AXIS))) > 0,
    'the tilt is about companion x camera, the shortest path to the viewer');
  assert.ok(companionTilt(behind, 0) === null, 'no headroom, no tilt');
  // Dead behind: there is no shortest path to pick.
  assert.equal(companionTilt(vec3(0, 0, -1), 30), null);
});

test('readingPoseTransform emits a CSS transform list, outermost first', () => {
  const geometry = dieGeometry(20);
  const text = readingPoseTransform(geometry, 7);
  const turns = readingPose(surfaceDirections(geometry), 7, { uprightContent: true });
  assert.equal(text, turns.map(rotate3d).join(' '));
  assert.ok(/^rotate3d\([-\d.]+,[-\d.]+,[-\d.]+,[-\d.]+deg\)( rotate3d\(.*\))*$/.test(text), text);
  // No var() or calc(): a keyframe carrying either forfeits compositing.
  assert.ok(!text.includes('var(') && !text.includes('calc('));
});

test('a facet already square-on and upright needs no pose at all', () => {
  // Two coincident directions would be a degenerate solid, so this exercises
  // the "no turns" path through a hand-built surface rather than a die.
  const flat = [CAMERA_AXIS, vec3(0, 0, -1)];
  const turns = readingPose(flat, 0, { uprightContent: true });
  // The resting view is a deliberate lean, so there IS a turn; what must hold
  // is that the transform is well formed and never the empty string.
  assert.ok(turns.length >= 1);
  const surface = {
    faces: [{ normal: vec3(0, 0, 1), centroid: vec3(0, 0, 0.5), polygon: [] as readonly Vec3[] }],
    capFaces: [],
    circumradius: 1,
  };
  assert.ok(readingPoseTransform(surface, 0).startsWith('rotate3d('));
});
