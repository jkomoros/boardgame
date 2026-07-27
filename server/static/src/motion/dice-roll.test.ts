import assert from 'node:assert/strict';
import test from 'node:test';
import { QUAT_IDENTITY, dieGeometry, vec3, type Quat, type Vec3 } from './die-geometry.ts';
import type { DieTrajectory } from './dice-sim.ts';
import { FRAME_MS, TRAY_BOUNDS, dieRollSeed, sceneTransform, settledTrajectory } from './dice-roll.ts';
import { readingPose, surfaceDirections } from '../solid/reading-pose.ts';

/**
 * A hand-built trajectory. Nothing here calls `simulateRoll`: the properties
 * below are about the TRIM and the FRAMING, and a synthetic track states the
 * cases (a long dead tail, no dead tail, an all-still track) that a physics
 * throw only produces by luck.
 */
function trajectory(
  samples: readonly { t: number; position: Vec3; orientation?: Quat }[],
): DieTrajectory {
  const filled = samples.map((sample) => ({
    t: sample.t,
    position: sample.position,
    orientation: sample.orientation ?? QUAT_IDENTITY,
  }));
  return { samples: filled, restingOrientation: filled[filled.length - 1].orientation };
}

test('the roll seed is a pure function of the roll identity', () => {
  assert.equal(dieRollSeed('die-1', 3), dieRollSeed('die-1', 3));
  // Determinism is the whole point: a remount mid-roll must rebuild the same
  // throw rather than re-throwing the die in front of the player.
  assert.notEqual(dieRollSeed('die-1', 3), dieRollSeed('die-1', 4));
  assert.notEqual(dieRollSeed('die-1', 3), dieRollSeed('die-2', 3));
  // The component ID and the identity must not be able to trade places: without
  // the separator, ('die', 12) and ('die1', 2) would be one throw.
  assert.notEqual(dieRollSeed('die', 12), dieRollSeed('die1', 2));
});

test('the roll seed is a uint32, which is what the simulator wants', () => {
  for (const [id, identity] of [['a', 0], ['', 1], ['long-component-id', 999999]] as const) {
    const seed = dieRollSeed(id, identity);
    assert.ok(Number.isInteger(seed) && seed >= 0 && seed <= 0xffffffff, `${id}#${identity} -> ${seed}`);
  }
  // Distinct throws land on distinct seeds across a realistic run of rolls.
  const seeds = new Set<number>();
  for (let identity = 0; identity < 500; identity++) seeds.add(dieRollSeed('component-42', identity));
  assert.equal(seeds.size, 500);
});

test('the tray is inside what the simulator will accept, and roomy enough to read as a throw', () => {
  // `dice-sim.ts` rejects anything under 1.5: its spawn needs the clearance.
  assert.ok(TRAY_BOUNDS.x >= 1.5 && TRAY_BOUNDS.y >= 1.5 && TRAY_BOUNDS.z >= 1.5);
  assert.ok(Object.isFrozen(TRAY_BOUNDS));
  // One 60Hz frame, which is the grid the baked curve is sampled on.
  assert.ok(Math.abs(FRAME_MS - 1000 / 60) < 0.1);
});

/**
 * The last ~300ms of every simulated throw is the rest-detection hold: the die
 * frozen on its final pose. The roll is GATED, so playing it holds the whole
 * game's animation cycle open on a die that has already stopped.
 */
test('settledTrajectory cuts the trailing hold and nothing else', () => {
  const moving = [
    { t: 0, position: vec3(0, 3, 0) },
    { t: 100, position: vec3(0, 2, 0) },
    { t: 200, position: vec3(0, 1, 0) },
  ];
  const held = [300, 400, 500, 600].map((t) => ({ t, position: vec3(0, 1, 0) }));
  const cut = settledTrajectory(trajectory([...moving, ...held]));
  // The die is already at its final pose at t=200, so THAT is the first sample
  // it never moves again from — the trim keeps it and drops the four after it.
  assert.equal(cut.samples.length, 3, 'the first still sample is kept, the rest are dropped');
  assert.equal(cut.samples[cut.samples.length - 1].t, 200);
  // The die is left exactly where the throw left it.
  assert.deepEqual(cut.samples[cut.samples.length - 1].position, vec3(0, 1, 0));
  // And the resting orientation is the trimmed track's own last sample, so
  // `restingTransform` and `trajectoryCurve(...)(1)` stay byte-identical.
  assert.deepEqual(cut.restingOrientation, cut.samples[cut.samples.length - 1].orientation);
});

test('settledTrajectory keeps a throw that has no dead tail', () => {
  const die = trajectory([
    { t: 0, position: vec3(0, 3, 0) },
    { t: 100, position: vec3(0, 2, 0) },
    { t: 200, position: vec3(0, 1, 0) },
  ]);
  assert.equal(settledTrajectory(die), die, 'the same object, untouched');
});

test('settledTrajectory never trims a throw down to something unplayable', () => {
  // `trajectoryCurve` rejects a zero-length span, and a one-sample track has no
  // span at all, so the trim has a floor: two samples and a positive final time,
  // whatever the input looks like.
  for (const times of [[0, 100, 200, 300], [0, 100], [0, 50, 100, 150, 200, 250]]) {
    const still = trajectory(times.map((t) => ({ t, position: vec3(0, 1, 0) })));
    const cut = settledTrajectory(still);
    assert.ok(cut.samples.length >= 2, `${times}: ${cut.samples.length} samples left`);
    assert.ok(cut.samples[cut.samples.length - 1].t > 0, `${times}: zero-length animation`);
    assert.deepEqual(cut.restingOrientation, cut.samples[cut.samples.length - 1].orientation);
  }
  // A two-sample track has nothing to give up and comes back untouched.
  const two = trajectory([
    { t: 0, position: vec3(0, 1, 0) },
    { t: 100, position: vec3(0, 1, 0) },
  ]);
  assert.equal(settledTrajectory(two), two);
});

test('settledTrajectory measures rotation as well as travel', () => {
  // A die spinning in place has stopped travelling but has not stopped moving.
  const spin = (degrees: number): Quat => {
    const half = ((degrees * Math.PI) / 180) / 2;
    return [0, Math.sin(half), 0, Math.cos(half)];
  };
  const die = trajectory([
    { t: 0, position: vec3(0, 1, 0), orientation: spin(0) },
    { t: 100, position: vec3(0, 1, 0), orientation: spin(30) },
    { t: 200, position: vec3(0, 1, 0), orientation: spin(60) },
    { t: 300, position: vec3(0, 1, 0), orientation: spin(60) },
    { t: 400, position: vec3(0, 1, 0), orientation: spin(60) },
  ]);
  const cut = settledTrajectory(die);
  // The spin stops at t=200, so the two held samples after it are the dead tail
  // — travel alone would have declared the die settled at t=0.
  assert.equal(cut.samples.length, 3);
  assert.equal(cut.samples[2].t, 200);
});

/**
 * The scene prefix is what makes a rolled die land in the middle of its own box
 * however far across the invisible tray it drifted. Composing the pose and then
 * subtracting the POSED resting position (not the raw one) is the part that is
 * easy to get backwards, and getting it backwards leaves the die resting
 * somewhere off its own square.
 */
test('sceneTransform recentres the landed die on the origin', () => {
  const geometry = dieGeometry(6);
  const die = trajectory([
    { t: 0, position: vec3(0.4, 2.5, -0.3) },
    { t: 500, position: vec3(1.2, 0.5, -0.9) },
  ]);
  const radiusPx = 40;
  const scene = sceneTransform(geometry, die, 0, radiusPx);
  const translate = /^translate3d\(([-\d.]+)px,([-\d.]+)px,([-\d.]+)px\)/.exec(scene);
  assert.ok(translate, `scene must begin with a recentring translate3d: ${scene}`);
  const offset = vec3(Number(translate[1]), Number(translate[2]), Number(translate[3]));

  // Apply the emitted rotations to the resting position ourselves, the way the
  // browser will: the list reads outermost first, so the last entry acts first.
  const turns = readingPose(
    surfaceDirections(geometry, die.restingOrientation), 0, { uprightContent: false },
  ).filter((turn) => Math.abs(turn.degrees) > 1e-4);
  const posed = turns.reduceRight(
    (point, turn) => rotate(point, turn.axis, turn.degrees),
    vec3(1.2 * radiusPx, -0.5 * radiusPx, -0.9 * radiusPx), // toScreen, then to px
  );
  for (let axis = 0; axis < 3; axis++) {
    assert.ok(Math.abs(offset[axis] + posed[axis]) < 1e-3,
      `axis ${axis}: offset ${offset[axis]} does not cancel posed ${posed[axis]}`);
  }
  // Literals only: a var() or calc() in a keyframe forfeits compositing and
  // drops a multi-second tumble onto the main thread.
  assert.ok(!scene.includes('var(') && !scene.includes('calc('));
  assert.ok(!/\d[eE][-+]?\d/.test(scene), `no exponential notation: ${scene}`);
});

test('sceneTransform still emits a translation when the die landed on the origin', () => {
  const geometry = dieGeometry(8);
  const die = trajectory([
    { t: 0, position: vec3(0, 2, 0) },
    { t: 400, position: vec3(0, 0, 0) },
  ]);
  const scene = sceneTransform(geometry, die, 2, 25);
  assert.ok(scene.startsWith('translate3d(0px,0px,0px)'), scene);
});

/** Rodrigues, restated here so the test does not lean on the code it checks. */
function rotate(v: Vec3, axis: Vec3, degrees: number): Vec3 {
  const radians = (degrees * Math.PI) / 180;
  const c = Math.cos(radians);
  const s = Math.sin(radians);
  const cross = vec3(
    axis[1] * v[2] - axis[2] * v[1],
    axis[2] * v[0] - axis[0] * v[2],
    axis[0] * v[1] - axis[1] * v[0],
  );
  const d = axis[0] * v[0] + axis[1] * v[1] + axis[2] * v[2];
  return vec3(
    v[0] * c + cross[0] * s + axis[0] * d * (1 - c),
    v[1] * c + cross[1] * s + axis[1] * d * (1 - c),
    v[2] * c + cross[2] * s + axis[2] * d * (1 - c),
  );
}
