import assert from 'node:assert/strict';
import test from 'node:test';
import { QUAT_IDENTITY, dieGeometry, vec3, type Quat, type Vec3 } from './die-geometry.ts';
import type { DieTrajectory } from './dice-sim.ts';
import {
  FRAME_MS,
  MAX_ENTRY_LEAN_DEGREES,
  MAX_ENTRY_OFFSET_DIE_WIDTHS,
  PERSPECTIVE_DEPTH_DIE_SIZES,
  TRAY_BOUNDS,
  dieRollSeed,
  dieRollTrajectory,
  rollScene,
  settledTrajectory,
} from './dice-roll.ts';
import { presentedFaceIndex } from './die-faces.ts';
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
test('a roll comes to rest on the origin of its own box', () => {
  const geometry = dieGeometry(6);
  const die = trajectory([
    { t: 0, position: vec3(0.4, 2.5, -0.3) },
    { t: 500, position: vec3(1.2, 0.5, -0.9) },
  ]);
  const radiusPx = 40;
  const scene = rollScene(geometry, die, 0, radiusPx, 500);
  // Composing the pose and then subtracting the POSED resting position (not the
  // raw one) is the part that is easy to get backwards, and getting it backwards
  // leaves the die resting somewhere off its own square.
  assert.ok(scene.resting.startsWith('translate3d(0px,0px,0px)'), scene.resting);
  assert.ok(scene.resting.includes('translate3d(0px,0px,0px)') , scene.resting);

  // ...and it got there from somewhere. The travel at the start of the throw is
  // minus the posed resting position, applied the way the browser will: the list
  // reads outermost first, so the last entry acts first.
  const offset = travelOf(scene.transform(0));
  const turns = readingPose(
    surfaceDirections(geometry, die.restingOrientation), 0, { uprightContent: false },
  ).filter((turn) => Math.abs(turn.degrees) > 1e-4);
  const posedAt = (point: Vec3) => turns.reduceRight(
    (value, turn) => rotate(value, turn.axis, turn.degrees), point);
  // toScreen (CSS y points down), then to px.
  const start = posedAt(vec3(0.4 * radiusPx, -2.5 * radiusPx, -0.3 * radiusPx));
  const rest = posedAt(vec3(1.2 * radiusPx, -0.5 * radiusPx, -0.9 * radiusPx));
  const posed = vec3(start[0] - rest[0], start[1] - rest[1], start[2] - rest[2]);
  // The entry similarity then rescales that offset and turns it in the SCREEN
  // PLANE, so the depth axis carries the scale on its own and the lateral
  // magnitude has to agree with it. See `entrySimilarity`.
  const scale = offset[2] / posed[2];
  assert.ok(scale > 0 && scale <= 1 + EMITTED_SLACK, `entry scale ${scale}`);
  assert.ok(
    Math.abs(Math.hypot(offset[0], offset[1]) - scale * Math.hypot(posed[0], posed[1]))
      < EMITTED_SLACK,
    `lateral travel ${Math.hypot(offset[0], offset[1])} is not ${scale} of ${Math.hypot(posed[0], posed[1])}`,
  );
  // And it is the POSED offset and not the raw one: this throw's pose really
  // does move the die somewhere else, so subtracting the raw resting position
  // would leave it resting off its own square rather than in the middle of it.
  const raw = vec3(
    (0.4 - 1.2) * radiusPx, -(2.5 - 0.5) * radiusPx, (-0.3 + 0.9) * radiusPx);
  assert.ok(Math.abs(Math.hypot(...raw) - Math.hypot(...posed)) > 1e-6
    || raw.some((value, axis) => Math.abs(value - posed[axis]) > 1e-6),
    'the pose does nothing to this throw, so nothing here distinguishes it');
  assert.ok(Math.abs(offset[2] - raw[2] * scale) > 1e-3, `travel is the RAW offset: ${offset}`);
});

/**
 * THE HOLE FIX, stated as the shape of the emitted list.
 *
 * `backface-visibility` culls a facet on the sign of its accumulated m33 and
 * never asks where the camera is, so the two only agree while the solid sits on
 * the camera's axis. `perspective()` projects everything to its RIGHT, so the
 * die's travel has to come BEFORE it -- the solid is projected about its own
 * centre and only then moved -- and its depth AFTER it, where it is on the axis
 * and changes nothing but the die's size.
 */
test('a roll travels outside its camera and falls toward it inside', () => {
  const geometry = dieGeometry(20);
  const die = trajectory([
    { t: 0, position: vec3(1.4, 1.9, 0.8) },
    { t: 300, position: vec3(0.2, 0.9, 0.1) },
    { t: 600, position: vec3(0, 0, 0) },
  ]);
  const radiusPx = 50;
  const frame = rollScene(geometry, die, 3, radiusPx, 600).transform(0);
  const order = /^translate3d\([^)]*\) perspective\([^)]*\) translate3d\([^)]*\)/.test(frame);
  assert.ok(order, `travel, then camera, then depth: ${frame}`);
  // The camera stands PERSPECTIVE_DEPTH_DIE_SIZES die sizes in front of the
  // solid, which is one die size across, i.e. two `radiusPx` (two NOMINAL
  // radii, half the die's box each -- not two bounding radii, which on a barrel
  // would be up to 2.63x more). The same constant
  // frames the die that has never rolled (boardgame-die.ts's #inner.solid), and
  // a roll framed at a different depth would resize the solid the moment its
  // animation was removed.
  const camera = /perspective\(([\d.]+)px\)/.exec(frame);
  assert.ok(camera, frame);
  assert.equal(Number(camera[1]), PERSPECTIVE_DEPTH_DIE_SIZES * 2 * radiusPx);
  // The lateral travel is outside the projection and the depth inside it: the
  // die really did start off-centre AND off the screen plane, so neither term
  // is vacuously zero here.
  const outside = /^translate3d\(([-\d.]+)px,([-\d.]+)px,([-\d.]+)px\)/.exec(frame)!;
  assert.equal(Number(outside[3]), 0, 'depth never rides outside the projection');
  assert.ok(Math.abs(Number(outside[1])) + Math.abs(Number(outside[2])) > 10, frame);
  const inside = /perspective\([^)]*\) translate3d\(0px,0px,([-\d.]+)px\)/.exec(frame)!;
  assert.ok(Math.abs(Number(inside[1])) > 1, `the depth term is doing work: ${frame}`);
});

/**
 * `fill: 'none'`, so the element renders its RESTING style the instant the
 * animation finishes -- or is finished early by the cycle sweep. A single
 * rounding digit of disagreement is the die twitching as it settles.
 */
test('the resting transform is the curve\'s own last frame, byte for byte', () => {
  const geometry = dieGeometry(12);
  const die = trajectory([
    { t: 0, position: vec3(1.1, 2.2, -0.4) },
    { t: 250, position: vec3(0.6, 0.9, -0.2) },
    { t: 700, position: vec3(-0.3, 0.5, 0.7) },
  ]);
  const scene = rollScene(geometry, die, 5, 37, 700);
  assert.equal(scene.resting, scene.transform(1));
  // ...and progress past the ends clamps rather than extrapolating.
  assert.equal(scene.transform(1.4), scene.resting);
  assert.equal(scene.transform(-2), scene.transform(0));
  // Literals only: a var() or calc() in a keyframe forfeits compositing and
  // drops a multi-second tumble onto the main thread.
  for (const value of [scene.transform(0), scene.transform(0.5), scene.resting]) {
    assert.ok(!value.includes('var(') && !value.includes('calc('), value);
    assert.ok(!/\d[eE][-+]?\d/.test(value), `no exponential notation: ${value}`);
  }
});

test('a roll that landed on the origin still emits its travel and its camera', () => {
  const geometry = dieGeometry(8);
  const die = trajectory([
    { t: 0, position: vec3(0, 2, 0) },
    { t: 400, position: vec3(0, 0, 0) },
  ]);
  const scene = rollScene(geometry, die, 2, 25, 400);
  assert.ok(scene.resting.startsWith('translate3d(0px,0px,0px) perspective('), scene.resting);
});

/** The travel of one emitted frame: the outer translate3d, plus the depth. */
function travelOf(transform: string): Vec3 {
  const outer = /^translate3d\(([-\d.]+)px,([-\d.]+)px,0px\)/.exec(transform);
  const depth = /perspective\([^)]*\) translate3d\(0px,0px,([-\d.]+)px\)/.exec(transform);
  assert.ok(outer && depth, `a frame must lead with its travel and then its depth: ${transform}`);
  return vec3(Number(outer[1]), Number(outer[2]), Number(depth[1]));
}

/**
 * Every roll a shape makes over a run of identities, framed the way
 * `boardgame-die.ts` frames it. Real throws, because the entry offset is a
 * property of where the SIMULATOR spawns a die and where the reading pose then
 * puts that spawn on screen, and no synthetic track states it.
 */
const ENTRY_SHAPES = [3, 4, 6, 7, 10, 12, 20] as const;
const ENTRY_SEEDS = 30;
const ENTRY_RADIUS_PX = 50; // A 100px die: `--die-size` on pig's board.
/**
 * Emitted lengths are rounded to five decimals (`cssNumber`), so a bound read
 * back off a transform string is only ever exact to about that. Far below a
 * pixel, and stated rather than hidden inside a fudged tolerance.
 */
const EMITTED_SLACK = 1e-4;

function* seededScenes(radiusPx: number) {
  for (const faceCount of ENTRY_SHAPES) {
    const geometry = dieGeometry(faceCount);
    for (let seed = 0; seed < ENTRY_SEEDS; seed++) {
      const die = settledTrajectory(
        dieRollTrajectory(geometry, `die-${faceCount}-${seed}`, 1).dice[0]);
      const durationMs = die.samples[die.samples.length - 1].t;
      const presented = presentedFaceIndex(geometry, die.restingOrientation);
      yield {
        faceCount,
        seed,
        scene: rollScene(geometry, die, presented, radiusPx, durationMs),
      };
    }
  }
}

/**
 * A roll enters mid-flight on purpose — a die that starts from rest and then
 * begins to tumble reads as a stutter, not a throw — but the SIZE of that jump
 * has to stay inside what a frame of the flight itself covers. Unbounded, it
 * was a median of 53px and up to 114px on a 100px die against a largest
 * in-flight step of 8px, i.e. the first frame carrying ten frames' worth of
 * travel on top of a complete change of orientation.
 *
 * Stated in DIE WIDTHS, never in pixels: `radiusPx` is the caller's and the
 * whole point of the bound is that it holds at any die size.
 */
test('a roll never enters further than the cap from where it lands', () => {
  let worst = 0;
  for (const { faceCount, seed, scene } of seededScenes(ENTRY_RADIUS_PX)) {
    const entry = travelOf(scene.transform(0));
    const offset = Math.hypot(entry[0], entry[1]) / (2 * ENTRY_RADIUS_PX);
    worst = Math.max(worst, offset);
    assert.ok(
      offset <= MAX_ENTRY_OFFSET_DIE_WIDTHS + EMITTED_SLACK,
      `d${faceCount} seed ${seed} enters ${offset} die widths away`,
    );
  }
  // ...and the cap really is doing work rather than sitting above every roll.
  assert.ok(worst > MAX_ENTRY_OFFSET_DIE_WIDTHS * 0.8, `the cap never binds: worst ${worst}`);
});

/**
 * A HARD clamp would satisfy the bound above and would pin four rolls in five
 * to exactly the cap — the median entry is 1.3 caps and the worst 2.9 — so
 * every throw would begin the same distance from where it lands, which is its
 * own kind of tell. The soft cap keeps the spread the throws have.
 */
test('the entry cap keeps the spread the throws have', () => {
  const offsets = [...seededScenes(ENTRY_RADIUS_PX)].map(({ scene }) => {
    const entry = travelOf(scene.transform(0));
    return Math.hypot(entry[0], entry[1]) / (2 * ENTRY_RADIUS_PX);
  });
  const sorted = [...offsets].sort((a, b) => a - b);
  const atTheCap = offsets.filter(
    (offset) => offset > MAX_ENTRY_OFFSET_DIE_WIDTHS - EMITTED_SLACK).length;
  assert.equal(atTheCap, 0, `${atTheCap} rolls sit exactly on the cap`);
  assert.ok(
    sorted[sorted.length - 1] / sorted[Math.floor(sorted.length / 2)] > 1.1,
    `entry offsets are flattened onto one value: ${sorted[0]} to ${sorted[sorted.length - 1]}`,
  );
});

/**
 * A die that enters BELOW where it lands floats upward onto the table, which is
 * the one thing a thrown die never does. It happened in 56% of rolls, by as
 * much as 74px on a 100px die, because the reading pose turns the whole world
 * to aim at the landed face and that turn is free to point the fall upward.
 */
test('a roll always enters from above where it lands', () => {
  for (const { faceCount, seed, scene } of seededScenes(ENTRY_RADIUS_PX)) {
    const entry = travelOf(scene.transform(0));
    // CSS y points down, so entering from above is a negative y offset.
    assert.ok(
      entry[1] <= EMITTED_SLACK,
      `d${faceCount} seed ${seed} enters ${entry[1]}px BELOW its resting centre`,
    );
    // Leaning is allowed — a die that always fell straight down the screen
    // would read as a lift, not a throw — but only this far off vertical.
    const lean = (Math.atan2(Math.abs(entry[0]), -entry[1]) * 180) / Math.PI;
    assert.ok(
      lean <= MAX_ENTRY_LEAN_DEGREES + EMITTED_SLACK,
      `d${faceCount} seed ${seed} enters ${lean} degrees off vertical`,
    );
  }
});

/**
 * WHAT THE CAP MAY NOT DO: bend the throw. It is applied as one similarity of
 * the whole travel path — a single scale and a single turn about the resting
 * point, chosen from the entry frame and then used for every frame — so the
 * path keeps its shape, keeps ending exactly at the origin, and gains no corner
 * for a player to see.
 */
test('the entry cap is one similarity of the whole path, not a per-frame nudge', () => {
  const geometry = dieGeometry(20);
  const die = trajectory([
    { t: 0, position: vec3(1.4, 1.9, 0.8) },
    { t: 200, position: vec3(0.9, 1.4, 0.5) },
    { t: 400, position: vec3(0.2, 0.9, 0.1) },
    { t: 600, position: vec3(0, 0, 0) },
  ]);
  const radiusPx = 50;
  const scene = rollScene(geometry, die, 3, radiusPx, 600);
  const raw = readingPose(
    surfaceDirections(geometry, die.restingOrientation), 3, { uprightContent: false },
  ).filter((turn) => Math.abs(turn.degrees) > 1e-4);
  const posedAt = (point: Vec3) => raw.reduceRight(
    (value, turn) => rotate(value, turn.axis, turn.degrees), point);
  const rest = posedAt(vec3(0, 0, 0));
  const untouched = (position: Vec3) =>
    posedAt(vec3(position[0] * radiusPx, -position[1] * radiusPx, position[2] * radiusPx))
      .map((value, axis) => value - rest[axis]);

  let ratio = NaN;
  let turn = NaN;
  for (const sample of die.samples) {
    const before = untouched(sample.position);
    const after = travelOf(scene.transform(sample.t / 600));
    const beforeLength = Math.hypot(before[0], before[1], before[2]);
    if (beforeLength < 1e-3) {
      // The resting frame: a similarity fixes the origin, whatever it does.
      assert.ok(Math.hypot(after[0], after[1], after[2]) < 1e-3, `${after}`);
      continue;
    }
    const afterLength = Math.hypot(after[0], after[1], after[2]);
    // One scale for every frame, all three axes.
    if (Number.isNaN(ratio)) ratio = afterLength / beforeLength;
    assert.ok(Math.abs(afterLength / beforeLength - ratio) < 1e-4,
      `t=${sample.t}: scaled by ${afterLength / beforeLength}, not ${ratio}`);
    assert.ok(Math.abs(after[2] - before[2] * ratio) < 1e-4,
      `t=${sample.t}: depth ${after[2]} is not ${before[2] * ratio}`);
    // ...and one turn, in the screen plane only.
    const angle = Math.atan2(after[1], after[0]) - Math.atan2(before[1], before[0]);
    const wrapped = Math.atan2(Math.sin(angle), Math.cos(angle));
    if (Number.isNaN(turn)) turn = wrapped;
    assert.ok(Math.abs(wrapped - turn) < 1e-4, `t=${sample.t}: turned ${wrapped}, not ${turn}`);
  }
  assert.ok(ratio > 0 && ratio <= 1, `a similarity may only shrink the throw: ${ratio}`);
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
